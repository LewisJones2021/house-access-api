package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/lewisjones2021/house-access-api/api"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// // create a new Gin router.
	// router := gin.New()
	// router.Use(gin.Logger())

	// // client, err :=

	client, err := InitMongoDB()

	if err != nil {
		fmt.Println("Failed to connect to MonogoDB", err)
	}
	defer client.Disconnect(context.Background())

	api.SetMongoClient(client) // Pass the client to the API package
	if err := api.ApiRoutes(); err != nil {
		log.Fatal("Failed to start the server:8080", err)
	}

}

// function to initialise the MongoDB client and return a reference to it:
func InitMongoDB() (*mongo.Client, error) {

	connectionString := "mongodb+srv://lewisjones:Adidass1122@cluster0.5wkgzkb.mongodb.net"

	// setup MongoDB client options.
	clientOptions := options.Client().ApplyURI(connectionString)

	// connect to mongoDB.
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// check the connection.
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to Mongo DataBase")

	return client, nil
}
