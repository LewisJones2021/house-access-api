package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lewisjones2021/house-access-api/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// sign-in details struct.
type SignInDetails struct {
	Email string
	Name  string
	Uid   string
	jwt.StandardClaims
}

// define a MongoDB collection for user data.
var userColllection *mongo.Collection = database.OpenCollection(database.Client, "user")

// retrieve the SECRET_KEY from environment variables.
var SECRET_KEY string = os.Getenv("SECRET_KEY")

// generateAllTokens generates both the detailed token and refresh token.
func GenerateAllTokens(Email string, Name string, Uid string) (signedToken string, signedRefreshToken string, err error) {
	// create claims for the access token.
	claims := &SignInDetails{
		Email: Email,
		Name:  Name,
		Uid:   Uid,
		StandardClaims: jwt.StandardClaims{
			// set the expiration time for the access token to 24 hours from now.
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	// create claims for the refresh token.
	refreshClaims := &SignInDetails{
		StandardClaims: jwt.StandardClaims{
			// set the expiration time for the refresh token to 7 days from now.
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(7*24)).Unix(),
		},
	}
	fmt.Println("SECRET_KEY:", SECRET_KEY)

	// generate the access token and check for any errors.
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("SECRET_KEY"))

	if err != nil {
		return "", "", err
	}

	// generate the refresh token and check for any errors.
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte("SECRET_KEY"))

	if err != nil {
		log.Panic(err)
		return
	}
	// return both tokens.
	return token, refreshToken, nil

}

// validateToken validates the jwt token.
func ValidateToken(signedToken string) (claims *SignInDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignInDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte("SECRET_KEY"), nil
		},
	)
	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignInDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}
	return claims, msg

}

// updateAllTokens renews the user tokens when they login.

func UpdateAllTokens(signedToken string, refreshToken string, userId string) {
	// create a context with a timeout of 100 seconds.
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// initialize an empty primitive.D (document) for updates.
	var updateObj primitive.D

	// append the signedToken and refreshToken to the updateObj.
	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refreshToken", Value: refreshToken})

	// set the upsert option to true, which means it will insert a new document if no matching document is found.
	// define a filter to specify which document to update based on the userId.
	// create options for the update operation, including the upsert option.
	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	// perform the update operation on the userCollection
	_, err := userColllection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)
	// defer the cancellation of the context to ensure it's cleaned up.
	defer cancel()

	// handle any errors that occurred during the update operation.
	if err != nil {
		log.Panic(err)
	}
	return
}
