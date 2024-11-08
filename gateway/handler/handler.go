package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"project/gateway/model"
	"project/gateway/utils"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type UserDto = model.UserDto

type Handler struct {
	Contract *client.Contract
}

func (handler *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user UserDto
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.CreateResponse(w, err, nil)
		return
	}
	_, err := handler.Contract.SubmitTransaction("RegisterUser", user.UserId, user.Name, user.Address, user.Contact)
	if err != nil {
		utils.CreateResponse(w, err, nil)
		return
	}
	utils.CreateResponse(w, err, "UserRegister Successfully")
}

func (handler *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	data, err := handler.Contract.EvaluateTransaction("GetAllUsers")
	if err != nil {
		utils.CreateResponse(w, err, nil)
		return
	}
	var users []UserDto
	err = json.Unmarshal(data, &users)
	if err != nil {
		utils.CreateResponse(w, fmt.Errorf("failed to decode users data: %v", err), nil)
		return
	}
	utils.CreateResponse(w, err, users)

}
