package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	if user.Address == "" {
		return fmt.Errorf("user field should not be empty")
	}
	return nil
}
