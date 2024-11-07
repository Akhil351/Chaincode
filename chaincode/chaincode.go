package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

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

type RealEstate struct {
	contractapi.Contract
}

func (r *RealEstate) RegisterUser(ctx contractapi.TransactionContextInterface, userId string, name string, address string, contact string) error {
	userExists, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return fmt.Errorf("failed to read user from world state: %v", err)
	}
	if userExists != nil {
		return fmt.Errorf("User Already exist")
	}
	user := User{UserId: userId, Name: name, Address: address, Contact: contact}
	userJson, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to convert user struct to json : %v", err)
	}
	err = ctx.GetStub().PutState(userId, userJson)
	if err != nil {
		return fmt.Errorf("failed to put user in world state: %v", err)
	}
	return nil

}

func(r *RealEstate) RegisterProperty(ctx contractapi.TransactionContextInterface,propertyId string,title string,location string,size float64,currentOwnerId string,price float64,isListed bool) error{
	propertyExists,err:=ctx.GetStub().GetState(propertyId)
	if(err!=nil){
		return fmt.Errorf("failed to read property from world state: %v",err)
	}
	if propertyExists!=nil{
		return fmt.Errorf("property Already exist")
	}
	property:=Property{Id: propertyId,Title: title,Location: location,Size: size,CurrentOwnerId: currentOwnerId,Price: price,IsListed: isListed}
	propertyJson,err:=json.Marshal(property)
	if err != nil {
		return fmt.Errorf("failed to convert property struct to json : %v", err)
	}
	err=ctx.GetStub().PutState(propertyId,propertyJson)

	if err != nil {
		return fmt.Errorf("failed to put property in world state: %v", err)
	}
	return nil
	
}



