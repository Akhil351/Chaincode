package web

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/joho/godotenv"
	"github.com/justinas/alice"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Routers(contract *client.Contract) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Err loading .env file")
	}
	mongoUri := os.Getenv("MONGO_URI")
	if len(mongoUri) == 0 {
		log.Fatal("MONGO_URI environment variable is no set")
	}
	clientOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
	db := client.Database("Chaincode")
	userCollection := db.Collection(os.Getenv("USER_COLLECTION"))
	propertyCollection := db.Collection(os.Getenv("PROPERTY_COLLECTION"))
	log.Println("Database Connected ")
	log.Println("Starting server...")
	router := mux.NewRouter()
	log.Println("Setting up routes")
	handler := &Handler{Contract: contract, UserCollection: userCollection, PropertyCollection: propertyCollection, JwtKey: os.Getenv("JWT_KEY")}
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
	router.Handle(apipath+"/updateProperty", chain.ThenFunc(handler.UpdateFlag)).Methods("PUT")
	log.Println("Listening in port 8080")
	http.ListenAndServe("localhost:8080", router)
}
