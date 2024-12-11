package web

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserDto struct {
	UserId   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Contact  string `json:"contact"`
	Password string `json:"password,omitempty"`
}

type User struct {
	UserId   string `json:"user_id" bson:"_id,omitempty"`
	Email    string `json:"email" bson:"email,omitempty"`
	Name     string `json:"name" bson:"name,omitempty"`
	Address  string `json:"address" bson:"address,omitempty"`
	Contact  string `json:"contact" bson:"contact,omitempty"`
	Password string `json:"password" bson:"password,omitempty"`
}
type PropertyDto struct {
	Id         string  `json:"id"`
	Title      string  `json:"title"`
	Location   string  `json:"location"`
	Size       float64 `json:"size"`
	OwnerEmail string  `json:"current_owner_email"`
	Price      float64 `json:"price"`
	IsListed   bool    `json:"is_listed"`
}
type Property struct {
    Id         string  `bson:"_id,omitempty" json:"property_id"`       
    Title      string  `bson:"title,omitempty" json:"title"`        
    Location   string  `bson:"location,omitempty" json:"location"`    
    Size       float64 `bson:"size,omitempty" json:"size"`            
    OwnerEmail string  `bson:"owner_email,omitempty" json:"current_owner_email"` 
    Price      float64 `bson:"price,omitempty" json:"price"`          
    IsListed   bool    `bson:"is_listed,omitempty" json:"is_listed"`  
}

type TransactionDto struct {
	Id          string  `json:"id"`
	PropertyId  string  `json:"property_id"`
	BuyerEmail  string  `json:"buyer_email"`
	SellerEmail string  `json:"seller_email"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Status      string  `json:"status"`
}

type Response struct {
	Status    string      `json:"status"`
	TimeStamp time.Time   `json:"timeStamp"`
	Data      interface{} `json:"data"`
	Error     interface{} `json:"error"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	UserId   string
	Name     string
	Email    string
	Password string
	jwt.RegisteredClaims
}
