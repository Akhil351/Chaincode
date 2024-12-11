package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Contract           *client.Contract
	UserCollection     *mongo.Collection
	PropertyCollection *mongo.Collection
	JwtKey             string
}

func (handler *Handler) generateToken(user User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		UserId:   user.UserId,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(handler.JwtKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (handler *Handler) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			CreateResponse(w, errors.New("no token provided"), nil, http.StatusUnauthorized)
			return
		}
		jwtToken := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(handler.JwtKey), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				CreateResponse(w, errors.New("invalid token signature"), nil, http.StatusUnauthorized)
				return
			}
			CreateResponse(w, errors.New("invalid token"), nil, http.StatusBadRequest)
			return
		}
		if !token.Valid {
			CreateResponse(w, errors.New("invalid token"), nil, http.StatusUnauthorized)
			return
		}
		filter := bson.M{"email": claims.Email}
		err = handler.UserCollection.FindOne(context.Background(), filter).Err()
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				CreateResponse(w, errors.New("user not found"), nil, http.StatusUnauthorized)
				return
			}
			CreateResponse(w, err, nil, http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (handler *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user UserDto
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	err := ValidRequest(user)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	filter := bson.M{"$or": []bson.M{
		{"email": user.Email},
		{"contact": user.Contact},
	}}
	var existingUser User
	if err = handler.UserCollection.FindOne(context.Background(), filter).Decode(&existingUser); err != nil {
		if err != mongo.ErrNoDocuments {
			CreateResponse(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	if existingUser.UserId != "" {
		CreateResponse(w, errors.New("email and contact number should be unique"), nil, http.StatusBadRequest)
		return
	}
	userId := "u" + uuid.New().String()
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	_, err = handler.Contract.SubmitTransaction("RegisterUser", userId, user.Name, user.Email, user.Address, user.Contact, string(bcryptPassword))
	if err != nil {
		log.Println("error in chaincode")
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	savedUser := ConvertToDto(user, User{})
	savedUser.UserId = userId
	savedUser.Password = string(bcryptPassword)
	_, err = handler.UserCollection.InsertOne(context.Background(), savedUser)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, nil, "User RegisterSuccessFully", http.StatusOK)
}

func (handler *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	data, err := handler.Contract.EvaluateTransaction("GetAllUsers")
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	if data == nil {
		CreateResponse(w, err, "no users are registred", http.StatusOK)
		return
	}
	var users []UserDto
	err = json.Unmarshal(data, &users)
	if err != nil {
		CreateResponse(w, fmt.Errorf("failed to decode users data: %v", err), nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, err, users, http.StatusOK)

}

func (handler *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		CreateResponse(w, fmt.Errorf("failed to decode request %v", err), nil, http.StatusBadRequest)
		return
	}
	var user User
	filter := bson.M{"email": request.Email}
	if err := handler.UserCollection.FindOne(context.Background(), filter).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			CreateResponse(w, errors.New("user doesn't exists"), nil, http.StatusUnauthorized)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		CreateResponse(w, errors.New("password is incorrect"), nil, http.StatusUnauthorized)
		return
	}
	jwtToken, err := handler.generateToken(user)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, nil, jwtToken, http.StatusOK)

}

func (handler *Handler) RegisterProperty(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	var property PropertyDto
	if err := json.NewDecoder(r.Body).Decode(&property); err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	err := ValidatePropertyDto(property)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	var existingProperty Property
	filter := bson.M{"$and": []bson.M{
		{"title": property.Title},
		{"location": property.Location},
	}}
	if err := handler.PropertyCollection.FindOne(context.Background(), filter).Decode(&existingProperty); err != nil {
		if err != mongo.ErrNoDocuments {
			CreateResponse(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	fmt.Println("property", existingProperty)
	if existingProperty.Id != "" {
		CreateResponse(w, errors.New("combination of  title and location should be unique"), nil, http.StatusBadRequest)
		return
	}
	propertyId := "p" + uuid.New().String()
	ownerEmail := claims.Email
	_, err = handler.Contract.SubmitTransaction("RegisterProperty", propertyId, property.Title, property.Location, strconv.FormatFloat(property.Size, 'f', 2, 64), ownerEmail, strconv.FormatFloat(property.Price, 'f', 2, 64), strconv.FormatBool(property.IsListed))
	if err != nil {
		log.Println("error in chaincode")
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	savedProperty := ConvertToDto(property, Property{})
	savedProperty.Id = propertyId
	savedProperty.OwnerEmail = ownerEmail
	_, err = handler.PropertyCollection.InsertOne(context.Background(), savedProperty)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, nil, "PropertySuccessFully", http.StatusOK)
}

func (handler *Handler) GetAllProperty(w http.ResponseWriter, r *http.Request) {
	data, err := handler.Contract.EvaluateTransaction("GetAllProperty")
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	if data == nil {
		CreateResponse(w, err, "no properties are registred", http.StatusOK)
		return
	}
	var property []PropertyDto
	err = json.Unmarshal(data, &property)
	if err != nil {
		CreateResponse(w, fmt.Errorf("failed to decode users data: %v", err), nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, err, property, http.StatusOK)

}

func (handler *Handler) BuyProperty(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	propertyId := r.URL.Query().Get("propertyId")
	buyerEmail := r.URL.Query().Get("buyerEmail")
	var property Property
	filter := bson.M{"_id": propertyId}
	if err := handler.PropertyCollection.FindOne(context.Background(), filter).Decode(&property); err != nil {
		if err == mongo.ErrNoDocuments {
			CreateResponse(w, errors.New("property not found"), nil, http.StatusNotFound)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	filter = bson.M{"email": buyerEmail}
	if err := handler.UserCollection.FindOne(context.Background(), filter).Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			CreateResponse(w, errors.New("buyer not registred"), nil, http.StatusNotFound)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	if !property.IsListed {
		CreateResponse(w, errors.New("property is not listed for sale"), nil, http.StatusBadRequest)
		return
	}
	if claims.Email != property.OwnerEmail {
		CreateResponse(w, errors.New("seller is not the current owner of the property"), nil, http.StatusBadRequest)
		return
	}
	if buyerEmail == property.OwnerEmail {
		CreateResponse(w, errors.New("buyer cannot be the current owner"), nil, http.StatusBadRequest)
		return
	}
	data, err := handler.Contract.SubmitTransaction("BuyProperty", propertyId, buyerEmail, claims.Email)
	if err != nil {
		log.Println("error in chaincode")
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	property.OwnerEmail = buyerEmail
	property.IsListed = false
	filter = bson.M{"_id": propertyId}
	update := bson.M{"$set": bson.M{"owner_email": buyerEmail, "is_listed": false}}
	_, err = handler.PropertyCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}

	CreateResponse(w, nil, string(data), http.StatusOK)

}

func (handler *Handler) GetAllTransaction(w http.ResponseWriter, r *http.Request) {
	transactionId := r.URL.Query().Get("transactionId")
	data, err := handler.Contract.EvaluateTransaction("GetAllTransaction", transactionId)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	if data == nil {
		CreateResponse(w, err, "no transactions", http.StatusOK)
		return
	}
	var transaction []TransactionDto
	err = json.Unmarshal(data, &transaction)
	if err != nil {
		CreateResponse(w, fmt.Errorf("failed to decode transaction data: %v", err), nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, err, transaction, http.StatusOK)

}

func (handler *Handler) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	propertyId := r.URL.Query().Get("propertyId")
	var property Property
	filter := bson.M{"_id": propertyId}
	if err := handler.PropertyCollection.FindOne(context.Background(), filter).Decode(&property); err != nil {
		if err == mongo.ErrNoDocuments {
			CreateResponse(w, errors.New("property not found"), nil, http.StatusNotFound)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}

	if property.IsListed {
		CreateResponse(w, errors.New("property already listed for sale"), nil, http.StatusBadRequest)
		return
	}

	if claims.Email != property.OwnerEmail {
		CreateResponse(w, errors.New("seller is not the current owner of the property"), nil, http.StatusBadRequest)
		return
	}
	_, err := handler.Contract.SubmitTransaction("UpdateFlag", propertyId, claims.Email)
	if err != nil {
		log.Println("error in chaincode")
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	property.IsListed = true
	update := bson.M{"$set": bson.M{"is_listed": property.IsListed}}
	//	opts := options.Update().SetUpsert(true)
	_, err = handler.PropertyCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, err, "Property Updated", http.StatusOK)

}
