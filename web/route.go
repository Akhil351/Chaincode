package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func Routers(contract *client.Contract) {
	r := mux.NewRouter()
	handler := Handler{Contract: contract}
	apipath := "/api/v2"
	r.HandleFunc(apipath+"/createUser", handler.RegisterUser).Methods("POST")
	r.HandleFunc(apipath+"/getUsers", handler.GetAllUsers).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe("localhost:8080", r)
	fmt.Println("Running")
}
