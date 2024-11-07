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


func (r *RealEstate) TransferPropertyOwnerShip(ctx contractapi.TransactionContextInterface,propertyId string,buyerId string,sellerId string,amount float64,data string) error{
	propertyBytes,err:=ctx.GetStub().GetState(propertyId)
	if err!=nil{
		return fmt.Errorf("failed to read property from world state: %v",err)
	}

	if(propertyBytes==nil){
		return fmt.Errorf("property not found")
	}

	var property Property
	err=json.Unmarshal(propertyBytes,&property)
	if(err!=nil){
		return fmt.Errorf("failed to unmarshal property: %v",err)
	}

	if !property.IsListed{
		return fmt.Errorf("property is not listed to sale")
	}

	if(property.CurrentOwnerId!=sellerId){
		return fmt.Errorf("seller is the not the current owner of the property")
	}

	if property.CurrentOwnerId==buyerId{
		return fmt.Errorf("buyer cannot be the current owner")
	}

	property.CurrentOwnerId=buyerId
	property.IsListed=false

	updatedPropertyJson,err:=json.Marshal(property)
	if(err!=nil){
		return fmt.Errorf("failed to  marshal updated property: %v",err)
	}
	err=ctx.GetStub().PutState(propertyId,updatedPropertyJson)
	if err!=nil{
		return fmt.Errorf("failed to update property in world state: %v",err)
	}
	transaction:=Transaction{Id:fmt.Sprintf("TXN-%s-%s-%s",sellerId,buyerId,propertyId),PropertyId: propertyId,BuyerId: buyerId,SellerId: sellerId,Amount: amount,Date: data,Status: "Complted"}

	transactionJson,err:=json.Marshal(transaction)
	if(err!=nil){
		return fmt.Errorf("failed to marshal transaction: %v",err)
	}
	err=ctx.GetStub().PutState(transaction.Id,transactionJson)
	if err!=nil{
		return fmt.Errorf("failed to save transaction to world state: %v",err)
	}
	return nil


}

func (r *RealEstate) GetAllUsers(ctx  contractapi.TransactionContextInterface)([]User,error){
	var users []User
	resultIterator,err:=ctx.GetStub().GetStateByRange("","")
	if err!=nil{
		return nil,fmt.Errorf("failed to get users: %v",err)
	}
	defer resultIterator.Close()
	for resultIterator.HasNext(){
		queryResponse,err:=resultIterator.Next()
		if err!=nil{
			return nil,fmt.Errorf("Failed to iterate over users: %v",err)
		}
		var user User
		err=json.Unmarshal(queryResponse.Value,&user)
        if(err!=nil){
			return nil,fmt.Errorf("failed to unmarshal user data: %v",err)
		}
		users=append(users, user)
	}
	return users,nil
}



