package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lewisjones2021/house-access-api/helpers"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lewisjones2021/house-access-api/database"
	"github.com/lewisjones2021/house-access-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// create a reference to the User collection in MongoDB.
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "User")

// create a validator instance for input validation.
var validate = validator.New()

// hashPassword is used to encrypt the password before it is stored in the DB.
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

// verifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprint("login or password is incorrect")
		check = false
	}
	return check, msg
}

// createUser is the api used to get a single user

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		// create a context with a timeout.
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		// bind the incoming JSON data to the user struct.
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// validate the user input.
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		}
		// check if the email already exists in the database.
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}
		// hash the user's password before storing it.
		password := HashPassword(user.Password)
		user.Password = password

		// if the email already exists, return an error.
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
			return
		}
		// generate user ID, access token, and refresh token.
		user.ID = primitive.NewObjectID()

		hexUserId := user.ID.Hex()
		user.User_id = hexUserId
		token, refreshToken, _ := helpers.GenerateAllTokens(user.Email, user.Password, user.User_id)
		user.Token = token
		user.Refresh_token = refreshToken

		// insert the user data into the database.
		resultInsertationNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("user item not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertationNumber)
	}

}

// login is the api used to get a single user.

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with a timeout of 100 seconds.
		ctx := c.Request.Context()

		// declare variables to hold user data.
		var user models.User
		var foundUser models.User

		// bind the JSON data from the request to the 'user' variable.
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error()})
			return
		}
		// attempt to find a user in the database based on their email.
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)

		// if an error occurs while finding the user, return an error response.
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}

		// check if the provided password matches the stored password for the user.
		passWordIsValid, msg := VerifyPassword(user.Password, foundUser.Password)

		// if the password is not valid, return a n error response.
		if !passWordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		fmt.Println(foundUser)
		// generate access and refresh tokens for the authenticated user.
		token, refreshToken, err := helpers.GenerateAllTokens(foundUser.Email, foundUser.Name, foundUser.User_id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		foundUser.Token = token

		// update the user's tokens in the database.
		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		// respond with a successful login and user data.
		c.JSON(http.StatusOK, foundUser)
	}

}
