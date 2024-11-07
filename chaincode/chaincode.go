package chaincode

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

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
	Size           float64  `json:"size"`
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

type RealEstate struct {
	contractapi.Contract
}
