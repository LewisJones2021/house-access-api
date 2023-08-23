package main

import (
	"context"
	"fmt"

	"github.com/lewisjones2021/house-access-api/api"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// function to initialise the MongoDB client and return a reference to it:
func initMongoDB()(*mongo.Client, error){
	connectionString := "mongodb+srv://lewisjones:Adidass1122@cluster0.5wkgzkb.mongodb.net/"

	// setup MongoDB client options.
	clientOptions := options.Client().ApplyURI(connectionString)
	
	// connect to mongoDB.
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil{
		return nil, err
	}
	
	// check the connection.
	err = client.Ping(context.Background(),nil)
	if err != nil{
		return nil, err
	}
	
	fmt.Println("Connected to Mongo DataBase")
	return client, nil
}
func main (){
	_, err := initMongoDB()
	if err != nil{
		fmt.Println("Failed to connect to MonogoDB", err)
	}
	api.StartServer()
}