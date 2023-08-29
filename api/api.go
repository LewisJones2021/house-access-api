package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// defining the House Model.
type House struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HouseName  string             `json:"houseName" bson:"houseName"`
	AccessCode string             `json:"accessCode" bson:"accessCode"`
	Notes      string             `json:"houseNotes" bson:"notes"`
}

// establish and maintain a connection to the MongoDB server.
// var client *mongo.Client

// reference to the "houses" collection in the connected MongoDB database.
var housesCollection *mongo.Collection

func SetMongoClient(c *mongo.Client) {
	// Get a reference to the "houses" collection
	housesCollection = c.Database("test").Collection("houses")
}

// Setups up rules on this server known as a CORS policy, this allows JavaScript's server to
// read our data.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func ApiRoutes() error {

	router := gin.Default()

	router.Use(CORSMiddleware())

	// api routes.
	router.GET("/api/houses", getHouses)
	// router.GET("/api/houses/:id", getHouseByID)
	router.POST("/api/houses", addHouse)
	// router.PUT("/api/houses/:id", updateHouse)
	// router.DELETE("/api/houses/:id", deleteHouse)

	// start the server.
	if err := router.Run(":8080"); err != nil {
		return err
	}

	return nil
}

// func to GET the houses.
func getHouses(c *gin.Context) {
	houseName := strings.TrimSpace(strings.ToLower(c.Query("houseName")))
	if houseName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "houseName must have a value to search"})
		return
	}

	fmt.Println("searching for:", houseName)

	// fetch all houses from the collection.
	var data  []House
	fmt.Println(data)
	pattern := primitive.Regex{Pattern: ".*" + regexp.QuoteMeta(houseName)}
	cursor, err  := housesCollection.Find(context.Background(), bson.M{"houseName": pattern})
	if err != nil {
		fmt.Println("Failed to find a house: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch houses"})
		return
	}

	// ensure that the MongoDB cursor is properly closed.
	 defer cursor.Close(context.Background())

// Iterate through the cursor and decode documents into the data slice.
for cursor.Next(context.Background()){
	var house House 
	if err := cursor.Decode(&house); err != nil{
		fmt.Println("failed to decode house", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to fetch the houses"})
		return
	}
	data = append(data,house)
}

	// Check for cursor errors after iterating.

	if err := cursor.Err(); err != nil{
		fmt.Println("Cursor error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to fetch the house data"})
		return
	}
	// respond with the array of matching houses.
	c.JSON(http.StatusOK, data)
}

// function to create a new house and add to the da	tabase.
func addHouse(c* gin.Context) { 
	var newHouse House
if err := c.ShouldBindJSON(&newHouse); err != nil {
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invaild data", "err": err})
	return
}

// insert the new house into the House collection
result, err := housesCollection.InsertOne(context.Background(), newHouse)
fmt.Println(result)
if err != nil{
	c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create a new house"})
	return
}

// get the inserted id.
insertedID := result.InsertedID.(primitive.ObjectID).Hex()
c.JSON(http.StatusCreated, gin.H{"message":"House successfully created.", "id": insertedID})
}
