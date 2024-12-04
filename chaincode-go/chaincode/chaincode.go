package chaincode

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
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

const userCompositeKey = "user~userId~name~email~address~contact~password"
const propertCompositeKey = "property~propertyId~title~location~size~ownerEmail~price~isListed"
const transactionCompositeKey = "transaction~transactionId~propertyId~buyerEmail~sellerEmail~amount~date~status"

func (r *RealEstate) RegisterUser(ctx contractapi.TransactionContextInterface, userId string, name string, email string, address string, contact string, password string) error {
	userKey, err := ctx.GetStub().CreateCompositeKey(userCompositeKey, []string{"user", userId, name, email, address, contact, password})
	if err != nil {
		log.Println("failed to create composite key for user ")
		return errors.New("failed to create composite key for user ")
	}
	err = ctx.GetStub().PutState(userKey, []byte{0x00})
	if err != nil {
		log.Println("failed to put user in world state")
		return errors.New("failed to put user in world state")
	}
	return nil
}

func (r *RealEstate) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]User, error) {
	var users []User
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(userCompositeKey, []string{"user"})
	if err != nil {
		return nil, errors.New("failed to get users")
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, errors.New("failed to iterate over users ")
		}
		_, keyParts, splitKeyErr := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if splitKeyErr != nil {
			return nil, fmt.Errorf("error splitting key: %s", splitKeyErr.Error())
		}
		user := User{
			UserId:   keyParts[1],
			Email:    keyParts[3],
			Name:     keyParts[2],
			Address:  keyParts[4],
			Contact:  keyParts[5],
			Password: keyParts[6],
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *RealEstate) RegisterProperty(ctx contractapi.TransactionContextInterface, propertyId string, title string, location string, size string, ownerEmail string, price string, isListed string) error {
	propertyKey, err := ctx.GetStub().CreateCompositeKey(propertCompositeKey, []string{"property", propertyId, title, location, size, ownerEmail, price, isListed})
	if err != nil {
		return errors.New("failed to create composite key for property")
	}
	err = ctx.GetStub().PutState(propertyKey, []byte{0x00})
	if err != nil {
		return errors.New("failed to put property in world state")
	}
	return nil

}

func (r *RealEstate) GetAllProperty(ctx contractapi.TransactionContextInterface) ([]Property, error) {
	var properties []Property
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(propertCompositeKey, []string{"property"})
	if err != nil {
		return nil, errors.New("failed to get properties")
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, errors.New("failed to iterate over properties")
		}
		_, keyParts, splitKeyErr := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if splitKeyErr != nil {
			return nil, fmt.Errorf("error splitting key: %s", splitKeyErr.Error())
		}
		size, err := strconv.ParseFloat(keyParts[4], 64)
		if err != nil {
			return nil, err
		}
		price, err := strconv.ParseFloat(keyParts[6], 64)
		if err != nil {
			return nil, err
		}
		isListed, err := strconv.ParseBool(keyParts[7])
		if err != nil {
			return nil, err
		}
		property := Property{
			Id:         keyParts[1],
			Title:      keyParts[2],
			Location:   keyParts[3],
			Size:       size,
			OwnerEmail: keyParts[5],
			Price:      price,
			IsListed:   isListed,
		}

		properties = append(properties, property)
	}
	return properties, nil
}

func (r *RealEstate) UpdateFlag(ctx contractapi.TransactionContextInterface, propertyId string, OwnerEmail string) error {

	propertyBytes, err := ctx.GetStub().GetStateByPartialCompositeKey(propertCompositeKey, []string{"property", propertyId})
	if err != nil {
		log.Println("failed to read property from world state")
		return errors.New("failed to read property from world state")
	}
	if !propertyBytes.HasNext() {
		log.Println("property not found")
		return errors.New("property not found")
	}
	queryResponse, err := propertyBytes.Next()
	if err != nil {
		return errors.New("failed to iterate over properties")
	}
	_, keyParts, splitKeyErr := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
	if splitKeyErr != nil {
		return fmt.Errorf("error splitting key: %s", splitKeyErr.Error())
	}
	keyParts[7] = "true"

	err = ctx.GetStub().DelState(queryResponse.Key)
	if err != nil {
		return errors.New("failed to delete old property state")
	}

	err = r.RegisterProperty(ctx, propertyId, keyParts[2], keyParts[3], keyParts[4], keyParts[5], keyParts[6], keyParts[7])
	if err != nil {
		return err
	}
	return nil

}

func (r *RealEstate) BuyProperty(ctx contractapi.TransactionContextInterface, propertyId string, buyerEmail string, sellerEmail string) (string, error) {
	propertyBytes, err := ctx.GetStub().GetStateByPartialCompositeKey(propertCompositeKey, []string{"property", propertyId})
	if err != nil {
		log.Println("failed to read property from world state")
		return "", errors.New("failed to read property from world state")
	}
	if !propertyBytes.HasNext() {
		log.Println("property not found")
		return "", errors.New("property not found")
	}
	queryResponse, err := propertyBytes.Next()
	if err != nil {
		return "", errors.New("failed to iterate over properties")
	}
	_, keyParts, splitKeyErr := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
	if splitKeyErr != nil {
		return "", fmt.Errorf("error splitting key: %s", splitKeyErr.Error())
	}
	
	transactionId := ctx.GetStub().GetTxID()
	currentTime := time.Now()
	transactionKey, err := ctx.GetStub().CreateCompositeKey(transactionCompositeKey, []string{"transaction", transactionId, propertyId, buyerEmail, sellerEmail, keyParts[6], currentTime.Format("2006-01-02 15:04:05"), "Completed"})
	if err != nil {
		log.Println("failed to create composite key for transaction")
		return "", errors.New("failed to create composite key for transaction")
	}

	err = ctx.GetStub().PutState(transactionKey, []byte{0x00})
	if err != nil {
		log.Println("failed to save transaction to world state")
		return "", errors.New("failed to save transaction to world state")
	}
	
	keyParts[5] = buyerEmail
	keyParts[7] = "false"

	err = ctx.GetStub().DelState(queryResponse.Key)
	if err != nil {
		return "", errors.New("failed to delete old property state")
	}

	err = r.RegisterProperty(ctx, propertyId, keyParts[2], keyParts[3], keyParts[4], keyParts[5], keyParts[6], keyParts[7])
	if err != nil {
		return "", err
	}
	return transactionId, nil
}

func (r *RealEstate) GetAllTransaction(ctx contractapi.TransactionContextInterface, transactionId string) ([]Transaction, error) {
	var resultIterator shim.StateQueryIteratorInterface
	var err error
	var transactions []Transaction
	if transactionId != "" {
		resultIterator, err = ctx.GetStub().GetStateByPartialCompositeKey(transactionCompositeKey, []string{"transaction", transactionId})
	} else {
		resultIterator, err = ctx.GetStub().GetStateByPartialCompositeKey(transactionCompositeKey, []string{"transaction"})
	}

	if err != nil {
		return nil, errors.New("failed to get transactions")
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, errors.New("failed to iterate over transactions")
		}
		_, keyParts, splitKeyErr := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if splitKeyErr != nil {
			return nil, fmt.Errorf("error splitting key: %s", splitKeyErr.Error())
		}
		amount, err := strconv.ParseFloat(keyParts[5], 64)
		if err != nil {
			return nil, err
		}
		transaction := Transaction{
			Id:          keyParts[1],
			PropertyId:  keyParts[2],
			BuyerEmail:  keyParts[3],
			SellerEmail: keyParts[4],
			Amount:      amount,
			Date:        keyParts[6],
			Status:      keyParts[7],
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
