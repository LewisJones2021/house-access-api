package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// define the user struct
type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	User_id       *string            `json:"user_id" `
	Name          *string            `json:"name"  `
	Email         *string            `json:"email" `
	Password      *string            `json:"password" `
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	User_type     *string            `json:"user_type"`
}
