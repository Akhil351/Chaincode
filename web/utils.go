package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/copier"
)

func CreateResponse(w http.ResponseWriter, err error, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	status := "Success"
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
		status = "Failed"
		w.WriteHeader(http.StatusBadRequest)
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
	return nil
}

func ConvertToDto[S any, T any](source S, destination T) T {
	copier.Copy(&destination, source)
	return destination
}
