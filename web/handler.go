package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"gorm.io/gorm"
)

type Handler struct {
	Contract *client.Contract
	DB       *gorm.DB
}

func (handler *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user UserDto
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		CreateResponse(w, err, nil)
		return
	}
	err := ValidRequest(user)
	if err != nil {
		CreateResponse(w, err, nil)
		return
	}
	var existingUser User
	if err := handler.DB.Where("email=? OR contact=?", user.Email, user.Contact).First(&existingUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			CreateResponse(w, err, nil)
			return
		}
	}
	if existingUser.UserId != "" {
		CreateResponse(w, errors.New("email and contact number should be unique"), nil)
		return
	}
	userId := uuid.New().String()
	_, err = handler.Contract.SubmitTransaction("RegisterUser", userId, user.Name, user.Email, user.Address, user.Contact)
	if err != nil {
		log.Println("error in chaincode")
		CreateResponse(w, err, nil)
		return
	}
	savedUser := ConvertToDto(user, User{})
	savedUser.UserId = userId
	if err := handler.DB.Save(&savedUser).Error; err != nil {
		CreateResponse(w, err, nil)
		return
	}
	CreateResponse(w, nil, savedUser)
}

func (handler *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	data, err := handler.Contract.EvaluateTransaction("GetAllUsers")
	if err != nil {
		CreateResponse(w, err, nil)
		return
	}
	var users []UserDto
	err = json.Unmarshal(data, &users)
	if err != nil {
		CreateResponse(w, fmt.Errorf("failed to decode users data: %v", err), nil)
		return
	}
	CreateResponse(w, err, users)

}
