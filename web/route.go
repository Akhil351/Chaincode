package web

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
func Routers(contract *client.Contract) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Err loading .env file")
	}
	connStr := os.Getenv("DATA_PSQL_URL")
	if len(connStr) == 0 {
		log.Fatal("DATA_PSQL_URL environment variable is nit set")
	}
	DB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	DB.AutoMigrate(&User{})
	log.Println("Database Connected ")
	log.Println("Starting server...")
	router := mux.NewRouter()
	log.Println("Setting up routes")
	handler := &Handler{Contract: contract,DB: DB}
	apipath := "/api/v2"
	router.HandleFunc(apipath+"/createUser", handler.RegisterUser).Methods("POST")
	router.HandleFunc(apipath+"/getUsers", handler.GetAllUsers).Methods("GET")
	log.Println("Listening in port 8080")
	http.ListenAndServe("localhost:8080", router)
}
