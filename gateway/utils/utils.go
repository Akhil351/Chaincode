package utils

import (
	"encoding/json"
	"net/http"
	"project/gateway/model"
	"time"
)
type Response=model.Response
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
