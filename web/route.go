package web

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/joho/godotenv"
	"github.com/justinas/alice"
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
	DB.AutoMigrate(&User{},&Property{},&Transaction{})
	log.Println("Database Connected ")
	log.Println("Starting server...")
	router := mux.NewRouter()
	log.Println("Setting up routes")
	handler := &Handler{Contract: contract, DB: DB, JwtKey: os.Getenv("JWT_KEY")}
	apipath := "/api/v2"
	router.HandleFunc(apipath+"/createUser", handler.RegisterUser).Methods("POST")
	router.HandleFunc(apipath+"/login", handler.Login).Methods("POST")
	// chain
	chain := alice.New(handler.jwtMiddleware)
	router.Handle(apipath+"/getUsers", chain.ThenFunc(handler.GetAllUsers)).Methods("GET")
	router.Handle(apipath+"/registerProperty", chain.ThenFunc(handler.RegisterProperty)).Methods("POST")
	router.Handle(apipath+"/getProperties", chain.ThenFunc(handler.GetAllProperty)).Methods("GET")
	router.Handle(apipath+"/sellProperty", chain.ThenFunc(handler.BuyProperty)).Methods("GET")
	router.Handle(apipath+"/getTransactions", chain.ThenFunc(handler.GetAllTransaction)).Methods("GET")
	router.Handle(apipath+"/getTransaction", chain.ThenFunc(handler.GetTransactionById)).Methods("GET")
	router.Handle(apipath+"/updateProperty", chain.ThenFunc(handler.UpdateFlag)).Methods("PUT")
	log.Println("Listening in port 8080")
	http.ListenAndServe("localhost:8080", router)
}
