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
	UserId   string `gorm:"primary_key" json:"user_id"`
	Email    string `gorm:"unique" json:"email"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Contact  string `json:"contact"`
	Password string `json:"password"`
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
	Id         string  `gorm:"primary_key" json:"property_id"`
	Title      string  `json:"title"`
	Location   string  `json:"location"`
	Size       float64 `json:"size"`
	OwnerEmail string  `json:"current_owner_email"`
	Price      float64 `json:"price"`
	IsListed   bool    `json:"is_listed"`
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

type Transaction struct {
	Id          string  `gorm:"primary_key" json:"id"`
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
