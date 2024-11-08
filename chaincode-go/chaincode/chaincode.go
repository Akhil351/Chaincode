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
	compositeIndexName := "userType~userId"
	userKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"USER", userId})
	if err != nil {
		return fmt.Errorf("failed to create composite key for user : %v", err)
	}
	userExists, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return fmt.Errorf("failed to read user from world state: %v", err)
	}
	if userExists != nil {
		return fmt.Errorf("User already exists")
	}
	user := User{UserId: userId, Name: name, Address: address, Contact: contact}
	userJson, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to convert user struct to json: %v", err)
	}
	err = ctx.GetStub().PutState(userKey, userJson)
	if err != nil {
		return fmt.Errorf("failed to put user in world state: %v", err)
	}
	return nil

}

func (r *RealEstate) RegisterProperty(ctx contractapi.TransactionContextInterface, propertyId string, title string, location string, size float64, currentOwnerId string, price float64, isListed bool) error {
	compositeIndexName := "propertyType~propertyId"
	propertyKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"PROPERTY", propertyId})
	if err != nil {
		return fmt.Errorf("failed to create composite key for property : %v", err)
	}
	propertyExists, err := ctx.GetStub().GetState(propertyKey)
	if err != nil {
		return fmt.Errorf("failed to read property from world state: %v", err)
	}
	if propertyExists != nil {
		return fmt.Errorf("Property already exists")
	}
	property := Property{Id: propertyId, Title: title, Location: location, Size: size, CurrentOwnerId: currentOwnerId, Price: price, IsListed: isListed}
	propertyJson, err := json.Marshal(property)
	if err != nil {
		return fmt.Errorf("failed to convert property struct to JSON: %v", err)
	}
	err = ctx.GetStub().PutState(propertyKey, propertyJson)
	if err != nil {
		return fmt.Errorf("failed to put property in world state: %v", err)
	}
	return nil

}

func (r *RealEstate) BuyProperty(ctx contractapi.TransactionContextInterface, propertyId string, buyerId string, sellerId string, price float64, date string) error {
	compositeIndexName := "propertyType~propertyId"
	propertyKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"PROPERTY", propertyId})
	if err != nil {
		return fmt.Errorf("failed to create composite key for property: %v", err)
	}

	propertyBytes, err := ctx.GetStub().GetState(propertyKey)
	if err != nil {
		return fmt.Errorf("failed to read property from world state: %v", err)
	}
	if propertyBytes == nil {
		return fmt.Errorf("property not found")
	}

	var property Property
	err = json.Unmarshal(propertyBytes, &property)
	if err != nil {
		return fmt.Errorf("failed to unmarshal property: %v", err)
	}

	if !property.IsListed {
		return fmt.Errorf("property is not listed for sale")
	}
	if property.CurrentOwnerId != sellerId {
		return fmt.Errorf("seller is not the current owner of the property")
	}
	if property.CurrentOwnerId == buyerId {
		return fmt.Errorf("buyer cannot be the current owner")
	}

	property.CurrentOwnerId = buyerId
	property.IsListed = false

	updatedPropertyJson, err := json.Marshal(property)
	if err != nil {
		return fmt.Errorf("failed to marshal updated property: %v", err)
	}
	err = ctx.GetStub().PutState(propertyKey, updatedPropertyJson)
	if err != nil {
		return fmt.Errorf("failed to update property in world state: %v", err)
	}

	transactionCompositeIndexName := "transactionType~transactionId"
	transactionId := fmt.Sprintf("TXN-%s-%s-%s", sellerId, buyerId, propertyId)
	transactionKey, err := ctx.GetStub().CreateCompositeKey(transactionCompositeIndexName, []string{"TRANSACTION", transactionId})
	if err != nil {
		return fmt.Errorf("failed to create composite key for transaction: %v", err)
	}

	transaction := Transaction{
		Id:         transactionId,
		PropertyId: propertyId,
		BuyerId:    buyerId,
		SellerId:   sellerId,
		Amount:     price,
		Date:       date,
		Status:     "Completed",
	}

	transactionJson, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}

	err = ctx.GetStub().PutState(transactionKey, transactionJson)
	if err != nil {
		return fmt.Errorf("failed to save transaction to world state: %v", err)
	}

	return nil
}

func (r *RealEstate) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]User, error) {
	var users []User
	compositeIndexName := "userType~userId"
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(compositeIndexName, []string{"USER"})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v ", err)
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over users : %v", err)
		}
		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *RealEstate) GetAllProperty(ctx contractapi.TransactionContextInterface) ([]Property, error) {
	var properties []Property
	compositeIndexName := "propertyType~propertyId"
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(compositeIndexName, []string{"PROPERTY"})
	if err != nil {
		return nil, fmt.Errorf("failed to get properties: %v ", err)
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over properties: %v", err)
		}
		var property Property
		err = json.Unmarshal(queryResponse.Value, &property)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal property data: %v", err)
		}
		properties = append(properties, property)
	}
	return properties, nil
}

func (r *RealEstate) GetAllTransaction(ctx contractapi.TransactionContextInterface) ([]Transaction, error) {
	var transactions []Transaction
	compositeIndexName := "transactionType~transactionId"
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(compositeIndexName, []string{"TRANSACTION"})
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %v ", err)
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over transactions: %v", err)
		}
		var transaction Transaction
		err = json.Unmarshal(queryResponse.Value, &transaction)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction data: %v", err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
