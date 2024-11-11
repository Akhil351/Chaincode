package chaincode

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type User struct {
	UserId   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Contact  string `json:"contact"`
	Password string `json:"password"`
}
type Property struct {
	Id         string  `json:"id"`
	Title      string  `json:"title"`
	Location   string  `json:"location"`
	Size       float64 `json:"size"`
	OwnerEmail string  `json:"current_owner_email"`
	Price      float64 `json:"price"`
	IsListed   bool    `json:"is_listed"`
}

type Transaction struct {
	Id          string  `json:"id"`
	PropertyId  string  `json:"property_id"`
	BuyerEmail  string  `json:"buyer_email"`
	SellerEmail string  `json:"seller_email"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Status      string  `json:"status"`
}

type RealEstate struct {
	contractapi.Contract
}

func (r *RealEstate) RegisterUser(ctx contractapi.TransactionContextInterface, userId string, name string, email string, address string, contact string, password string) error {
	compositeIndexName := "userType~userEmail"
	userKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"USER", email})
	if err != nil {
		log.Println("failed to create composite key for user ")
		return errors.New("failed to create composite key for user ")
	}
	userExists, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		log.Println("failed to read user from world state")
		return errors.New("failed to read user from world state")
	}
	if userExists != nil {
		log.Println("User already exists")
		return errors.New("User already exists")
	}
	user := User{UserId: userId, Name: name, Email: email, Address: address, Contact: contact, Password: password}
	userJson, err := json.Marshal(user)
	if err != nil {
		log.Println("failed to convert user struct to json")
		return errors.New("failed to convert user struct to json")
	}
	err = ctx.GetStub().PutState(userKey, userJson)
	if err != nil {
		log.Println("failed to put user in world state")
		return errors.New("failed to put user in world state")
	}
	return nil

}

func (r *RealEstate) RegisterProperty(ctx contractapi.TransactionContextInterface, propertyId string, title string, location string, size float64, ownerEmail string, price float64, isListed bool) error {
	compositeIndexName := "propertyType~propertyId"
	propertyKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"PROPERTY", propertyId})
	if err != nil {
		return errors.New("failed to create composite key for property")
	}
	propertyExists, err := ctx.GetStub().GetState(propertyKey)
	if err != nil {
		return errors.New("failed to read property from world state")
	}
	if propertyExists != nil {
		return errors.New("Property already exists")
	}
	property := Property{Id: propertyId, Title: title, Location: location, Size: size, OwnerEmail: ownerEmail, Price: price, IsListed: isListed}
	propertyJson, err := json.Marshal(property)
	if err != nil {
		return errors.New("failed to convert property struct to JSON")
	}
	err = ctx.GetStub().PutState(propertyKey, propertyJson)
	if err != nil {
		return errors.New("failed to put property in world state")
	}
	return nil

}
func (r *RealEstate) UpdateFlag(ctx contractapi.TransactionContextInterface, propertyId string, OwnerEmail string) error {
	compositeIndexName := "propertyType~propertyId"
	propertyKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"PROPERTY", propertyId})
	if err != nil {
		log.Print("failed to create composite key for property")
		return errors.New("failed to create composite key for property")
	}

	propertyBytes, err := ctx.GetStub().GetState(propertyKey)
	if err != nil {
		log.Println("failed to read property from world state")
		return errors.New("failed to read property from world state")
	}
	if propertyBytes == nil {
		log.Println("property not found")
		return errors.New("property not found")
	}
	var property Property

	err = json.Unmarshal(propertyBytes, &property)
	if err != nil {
		log.Println("failed to unmarshal property")
		return errors.New("failed to unmarshal property")
	}
	if property.IsListed {
		return errors.New("property already listed for sale")
	}
	if property.OwnerEmail != OwnerEmail {
		log.Println("seller is not the current owner of the property")
		return errors.New("seller is not the current owner of the property")
	}

	property.IsListed = true

	updatedPropertyJson, err := json.Marshal(property)
	if err != nil {
		log.Println("failed to marshal updated property")
		return errors.New("failed to marshal updated property")
	}
	err = ctx.GetStub().PutState(propertyKey, updatedPropertyJson)
	if err != nil {
		log.Println("failed to update property in world state")
		return errors.New("failed to update property in world state")
	}
	return nil

}

func (r *RealEstate) BuyProperty(ctx contractapi.TransactionContextInterface, propertyId string, buyerEmail string, sellerEmail string) (Transaction, error) {
	var transaction Transaction
	compositeIndexName := "propertyType~propertyId"
	propertyKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"PROPERTY", propertyId})
	if err != nil {
		log.Print("failed to create composite key for property")
		return transaction, errors.New("failed to create composite key for property")
	}

	propertyBytes, err := ctx.GetStub().GetState(propertyKey)
	if err != nil {
		log.Println("failed to read property from world state")
		return transaction, errors.New("failed to read property from world state")
	}
	if propertyBytes == nil {
		log.Println("property not found")
		return transaction, errors.New("property not found")
	}

	var property Property
	err = json.Unmarshal(propertyBytes, &property)
	if err != nil {
		log.Println("failed to unmarshal property")
		return transaction, errors.New("failed to unmarshal property")
	}

	if !property.IsListed {
		log.Println("property is not listed for sale")
		return transaction, errors.New("property is not listed for sale")
	}
	if property.OwnerEmail != sellerEmail {
		log.Println("seller is not the current owner of the property")
		return transaction, errors.New("seller is not the current owner of the property")
	}
	if property.OwnerEmail == buyerEmail {
		log.Println("buyer cannot be the current owner")
		return transaction, errors.New("buyer cannot be the current owner")
	}

	property.OwnerEmail = buyerEmail
	property.IsListed = false

	updatedPropertyJson, err := json.Marshal(property)
	if err != nil {
		log.Println("failed to marshal updated property")
		return transaction, errors.New("failed to marshal updated property")
	}
	err = ctx.GetStub().PutState(propertyKey, updatedPropertyJson)
	if err != nil {
		log.Println("failed to update property in world state")
		return transaction, errors.New("failed to update property in world state")
	}

	transactionCompositeIndexName := "transactionType~transactionId"
	transactionId := ctx.GetStub().GetTxID()
	transactionKey, err := ctx.GetStub().CreateCompositeKey(transactionCompositeIndexName, []string{"TRANSACTION", transactionId})
	if err != nil {
		log.Println("failed to create composite key for transaction")
		return transaction, errors.New("failed to create composite key for transaction")
	}
	currentTime := time.Now()
	transaction = Transaction{
		Id:          transactionId,
		PropertyId:  propertyId,
		BuyerEmail:  buyerEmail,
		SellerEmail: sellerEmail,
		Amount:      property.Price,
		Date:        currentTime.Format("02-Jan-2006 03:04:05 PM"),
		Status:      "Completed",
	}

	transactionJson, err := json.Marshal(transaction)
	if err != nil {
		log.Println("failed to marshal transaction")
		return transaction, errors.New("failed to marshal transaction")
	}

	err = ctx.GetStub().PutState(transactionKey, transactionJson)
	if err != nil {
		log.Println("failed to save transaction to world state")
		return transaction, errors.New("failed to save transaction to world state")
	}
	return transaction, nil
}

func (r *RealEstate) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]User, error) {
	var users []User
	compositeIndexName := "userType~userEmail"
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(compositeIndexName, []string{"USER"})
	if err != nil {
		return nil, errors.New("failed to get users")
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, errors.New("failed to iterate over users ")
		}
		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, errors.New("failed to unmarshal user data")
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
		return nil, errors.New("failed to get properties")
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, errors.New("failed to iterate over properties")
		}
		var property Property
		err = json.Unmarshal(queryResponse.Value, &property)
		if err != nil {
			return nil, errors.New("failed to unmarshal property data")
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
		return nil, errors.New("failed to get transactions")
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, errors.New("failed to iterate over transactions")
		}
		var transaction Transaction
		err = json.Unmarshal(queryResponse.Value, &transaction)
		if err != nil {
			return nil, errors.New("failed to unmarshal transaction data")
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *RealEstate) GetTransactionById(ctx contractapi.TransactionContextInterface, transactionId string) (Transaction, error) {
	var transaction Transaction
	compositeIndexName := "transactionType~transactionId"
	transactionKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{"TRANSACTION", transactionId})
	if err != nil {
		log.Print("failed to create composite key for property")
		return transaction, errors.New("failed to create composite key for property")
	}
	transactionBytes, err := ctx.GetStub().GetState(transactionKey)
	if err != nil {
		log.Println("failed to read transaction from world state")
		return transaction, errors.New("failed to read transaction from world state")
	}
	if transactionBytes == nil {
		log.Println("transaction not found")
		return transaction, errors.New("transaction not found")
	}

	err = json.Unmarshal(transactionBytes, &transaction)
	if err != nil {
		log.Println("failed to unmarshal transaction")
		return transaction, errors.New("failed to unmarshal transaction")
	}
	return transaction, nil
}
