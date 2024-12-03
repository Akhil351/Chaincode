package chaincode

import (
	"errors"
	"fmt"
	"log"

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
