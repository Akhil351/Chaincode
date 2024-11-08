package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type Handler struct {
	Contract *client.Contract
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
	_, err = handler.Contract.SubmitTransaction("RegisterUser", uuid.New().String(), user.Name, user.Address, user.Contact)
	if err != nil {
		fmt.Println("RegisterUser")
		CreateResponse(w, err, nil)
		return
	}
	CreateResponse(w, nil, "UserRegister Successfully")
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
