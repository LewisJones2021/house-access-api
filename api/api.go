package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lewisjones2021/house-access-api/middleware"
	"github.com/lewisjones2021/house-access-api/routes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// defining the House Model.
type House struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HouseName  string             `json:"houseName" bson:"houseName"`
	AccessCode string             `json:"accessCode" bson:"accessCode"`
	Notes      string             `json:"houseNotes" bson:"houseNotes"`
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
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.Status(200)
			return
		}

		c.Next()
	}
}

func ApiRoutes() error {

	router := gin.Default()
	router.Use(CORSMiddleware())
	housesRoutes := router.Group("/api/houses")
	housesRoutes.Use(CORSMiddleware())
	// register user routes defined in the application.
	housesRoutes.Use(middleware.Authentication())
	// api routes.
	housesRoutes.GET("", getHouses)
	// router.GET("//:id", getHouseByID)
	housesRoutes.POST("", addHouse)
	housesRoutes.PUT("/:id", updateHouse)
	housesRoutes.DELETE("/:id", deleteHouse)

	routes.UserRoutes(router)

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
	var data []House
	fmt.Println(data)
	pattern := primitive.Regex{Pattern: ".*" + regexp.QuoteMeta(houseName)}
	cursor, err := housesCollection.Find(context.Background(), bson.M{"houseName": pattern})
	if err != nil {
		fmt.Println("Failed to find a house: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch houses"})
		return
	}

	// ensure that the MongoDB cursor is properly closed.
	defer cursor.Close(context.Background())

	// Iterate through the cursor and decode documents into the data slice.
	for cursor.Next(context.Background()) {
		fmt.Println(data)
		var house House
		if err := cursor.Decode(&house); err != nil {
			fmt.Println("failed to decode house", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch the houses"})
			return
		}
		data = append(data, house)
	}

	// Check for cursor errors after iterating.

	if err := cursor.Err(); err != nil {
		fmt.Println("Cursor error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch the house data"})
		return
	}
	// respond with the array of matching houses.
	c.JSON(http.StatusOK, data)
}

// function to create a new house and add to the database.
func addHouse(c *gin.Context) {
	var newHouse House
	if err := c.ShouldBindJSON(&newHouse); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invaild data", "err": err})
		return
	}

	// insert the new house into the House collection
	result, err := housesCollection.InsertOne(context.Background(), newHouse)
	fmt.Println(result)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create a new house"})
		return
	}

	// get the inserted id.
	insertedID := result.InsertedID.(primitive.ObjectID).Hex()
	c.JSON(http.StatusCreated, gin.H{"message": "House successfully created.", "id": insertedID})
}

// function to delete the house from the database.

// handle DELETE request to delete a house by its ID
func deleteHouse(c *gin.Context) {
	// get the house id from the url param.
	houseID := c.Param("id")

	// convert the house id to a primitive.objectID.
	objID, err := primitive.ObjectIDFromHex(houseID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid house ID"})
		return
	}

	// construct a filter to find the house by its ID
	filter := bson.M{"_id": objID}

	// delete the house from the data collection.
	result, err := housesCollection.DeleteOne(context.Background(), filter)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the house from the system"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "House not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "House successfully deleted"})
}

// function to update a house in the data system.
func updateHouse(c *gin.Context) {

	// get the house id from the url.
	houseID := c.Param("id")

	// convert the house id to primitive object.
	objID, err := primitive.ObjectIDFromHex(houseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the house information."})
		return
	}

	// define a struct to hold the updated data.
	var updatedHouse House
	if err := c.ShouldBindJSON(&updatedHouse); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
	}

	// filter through the data to find the data.
	filter := bson.M{"_id": objID}

	// create an update document with the new data.
	updateData := bson.M{
		"$set": bson.M{
			"houseName":  updatedHouse.HouseName,
			"accessCode": updatedHouse.AccessCode,
			"houseNotes": updatedHouse.Notes,
		},
	}
	fmt.Println(updateData)
	// perform the updateData operation
	result, err := housesCollection.UpdateOne(context.Background(), filter, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update the house data."})
		return
	}

	// check if the houseData was found/ updated.
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "house not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "House data successfully udpated"})
}

// GET user api endpoint
func getUser(c *gin.Context) {

}
