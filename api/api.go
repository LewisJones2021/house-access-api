package api

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// defining the House Model.
type House struct{
ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
HouseName string `json:"houseName" bson:"houseName"`
AccessCode string `json:"accessCode" bson:"accessCode"`
Notes string `json:"notes" bson:"notes"`
}

// establish and maintain a connection to the MongoDB server.
var client *mongo.Client

// reference to the "houses" collection in the connected MongoDB database.
var housesCollection *mongo.Collection


func SetMongoClient(c *mongo.Client){
	// Get a reference to the "houses" collection
housesCollection = c.Database("test").Collection("houses")

}
func ApiRoutes(){
	
	router := gin.Default()

	// api routes.
	 router.GET("/api/houses", getHouses)
	// router.GET("/api/houses/:id", getHouseByID)
	// router.POST("/api/houses", createHouse)
	// router.PUT("/api/houses/:id", updateHouse)
	// router.DELETE("/api/houses/:id", deleteHouse)

// start the server.
if err := router.Run(":8080"); err != nil{
	log.Fatal("Failed to start the server:8080", err)
}
}

// func to GET the houses.
func getHouses(c *gin.Context){
	// fetch all houses from the collection.
	cursor, err := housesCollection.Find(context.Background(), bson.M{})
	 if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch houses"})
		return
}
// ensure that the MongoDB cursor is properly closed.
defer cursor.Close(context.Background())

}