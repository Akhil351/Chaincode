package web

import "time"

type UserDto struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Contact string `json:"contact"`
}
type PropertyDto struct {
	Id             string  `json:"id"`
	Title          string  `json:"title"`
	Location       string  `json:"location"`
	Size           float64 `json:"size"`
	CurrentOwnerId string  `json:"current_owner_id"`
	Price          float64 `json:"price"`
	IsListed       bool    `json:"is_listed"`
}

type TransactionDto struct {
	Id         string  `json:"id"`
	PropertyId string  `json:"property_id"`
	BuyerId    string  `json:"buyer_id"`
	SellerId   string  `json:"seller_id"`
	Amount     float64 `json:"amount"`
	Date       string  `json:"date"`
	Status     string  `json:"status"`
}
type Response struct {
	Status    string      `json:"status"`
	TimeStamp time.Time   `json:"timeStamp"`
	Data      interface{} `json:"data"`
	Error     interface{} `json:"error"`
}

type User struct {
	UserId  string `json:"userId"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Contact string `json:"contact"`
}

type Property struct {
	Id             string  `json:"id"`
	Title          string  `json:"title"`
	Location       string  `json:"location"`
	Size           float64 `json:"size"`
	CurrentOwnerId string  `json:"current_owner_id"`
	Price          float64 `json:"price"`
	IsListed       bool    `json:"is_listed"`
}

type Transaction struct {
	Id         string  `json:"id"`
	PropertyId string  `json:"property_id"`
	BuyerId    string  `json:"buyer_id"`
	SellerId   string  `json:"seller_id"`
	Amount     float64 `json:"amount"`
	Date       string  `json:"date"`
	Status     string  `json:"status"`
}
