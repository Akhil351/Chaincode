package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/copier"
)

func CreateResponse(w http.ResponseWriter, err error, data interface{},code int) {
	w.Header().Set("Content-Type", "application/json")
	status := "Success"
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
		status = "Failed"
		w.WriteHeader(code)
	}
	response := Response{
		Status:    status,
		TimeStamp: time.Now(),
		Data:      data,
		Error:     errMsg,
	}
	json.NewEncoder(w).Encode(response)

}

func ValidRequest(user UserDto) error {
	if user.Name == "" {
		return fmt.Errorf("name field should not be empty")
	}
	if user.Contact == "" {
		return fmt.Errorf("contact field should not be empty")
	}
	if len(user.Contact) < 10 || len(user.Contact) > 15 {
		return errors.New("phone number should be between 10 and 15 digits")
	}
	if user.Address == "" {
		return fmt.Errorf("address field should not be empty")
	}
	if user.Email == "" {
		return errors.New("email field should not be empty")
	}
	if !strings.Contains(user.Email,"@gmail.com"){
		return errors.New("email must be @gmail.com address")
	}
	return nil
}

func ValidatePropertyDto(property PropertyDto) (error){
	 if property.Title==""{
		return errors.New("title field should not be empty")
	 }
	 if property.Location==""{
		return errors.New("location field should not be empty")
	 }
	 if property.Size==0{
		return errors.New("size field should not be empty")
	 }
	 if property.Price==0{
		return errors.New("price field should not be empty")
	 }
	 return nil
}

func ConvertToDto[S any, T any](source S, destination T) T {
	copier.Copy(&destination, source)
	return destination
}

