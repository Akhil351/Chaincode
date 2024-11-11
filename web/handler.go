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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	Contract *client.Contract
	DB       *gorm.DB
	JwtKey   string
}

func (handler *Handler) generateToken(user User) (string, error) {
	expirationTime := time.Now().Add(time.Hour)
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
		ctx := context.WithValue(r.Context(), "claims", claims)
		if err := handler.DB.Where("email=?", claims.Email).First(&User{}).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				CreateResponse(w, errors.New("user not found"), nil, http.StatusUnauthorized)
				return
			}
			CreateResponse(w, err, nil, http.StatusBadRequest)
			return
		}
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
	var existingUser User
	if err := handler.DB.Where("email=? OR contact=?", user.Email, user.Contact).First(&existingUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
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
	if err := handler.DB.Save(&savedUser).Error; err != nil {
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
	if err := handler.DB.Where("email=?", request.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			CreateResponse(w, errors.New("user doesn't exists"), nil, http.StatusUnauthorized)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		log.Println("error")
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
	if err := handler.DB.Where("title=? AND location=?", property.Title, property.Location).First(&existingProperty).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			CreateResponse(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	if existingProperty.Id != "" {
		CreateResponse(w, errors.New("combination of  title and location should be unique"), nil, http.StatusBadRequest)
		return
	}
	propertyId := "p" + uuid.New().String()
	ownerEmail := claims.Email
	_, err = handler.Contract.SubmitTransaction("RegisterProperty", propertyId, property.Title, property.Location, fmt.Sprintf("%f", property.Size), ownerEmail, fmt.Sprintf("%f", property.Price), strconv.FormatBool(property.IsListed))
	if err != nil {
		log.Println("error in chaincode")
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	savedProperty := ConvertToDto(property, Property{})
	savedProperty.Id = propertyId
	savedProperty.OwnerEmail = ownerEmail
	if err := handler.DB.Save(&savedProperty).Error; err != nil {
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
	if err := handler.DB.Where("id=?", propertyId).First(&property).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			CreateResponse(w, errors.New("property not found"), nil, http.StatusNotFound)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	if err := handler.DB.Where("email=?", buyerEmail).First(&User{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			CreateResponse(w, errors.New("buyer not registred"), nil, http.StatusNotFound)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
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

	if !property.IsListed {
		CreateResponse(w, errors.New("property is not listed for sale"), nil, http.StatusBadRequest)
		return
	}
	transactionId := "t" + uuid.New().String()
	_, err := handler.Contract.SubmitTransaction("BuyProperty", transactionId, propertyId, buyerEmail, claims.Email)
	if err != nil {
		log.Println("error in chaincode")
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	property.OwnerEmail = buyerEmail
	property.IsListed = false
	if err := handler.DB.Save(&property).Error; err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	currentTime := time.Now()
	transaction := Transaction{
		Id:          transactionId,
		PropertyId:  propertyId,
		BuyerEmail:  buyerEmail,
		SellerEmail: claims.Email,
		Amount:      property.Price,
		Date:        currentTime.Format("02-Jan-2006 03:04:05 PM"),
		Status:      "Completed",
	}
	if err := handler.DB.Save(&transaction).Error; err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, nil, "TransactionSuccessfully", http.StatusOK)

}

func (handler *Handler) GetAllTransaction(w http.ResponseWriter, r *http.Request) {
	data, err := handler.Contract.EvaluateTransaction("GetAllTransaction")
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
		CreateResponse(w, fmt.Errorf("failed to decode users data: %v", err), nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, err, transaction, http.StatusOK)

}

func (handler *Handler) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	propertyId := r.URL.Query().Get("propertyId")
	var property Property
	if err := handler.DB.Where("id=?", propertyId).First(&property).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			CreateResponse(w, errors.New("property not found"), nil, http.StatusNotFound)
			return
		}
		CreateResponse(w, err, nil, http.StatusBadRequest)
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
	if err := handler.DB.Save(&property).Error; err != nil {
		CreateResponse(w, err, nil, http.StatusBadRequest)
		return
	}
	CreateResponse(w, err, "Property Updated", http.StatusOK)

}
