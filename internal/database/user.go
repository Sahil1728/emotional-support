package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Sahil1728/emotional-support/internal/models"
)

func CreateUser(collection *mongo.Collection, user models.User) (*mongo.InsertOneResult, error) {
	user.Created = time.Now()
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Println("Error creating user:", err)
		return nil, err
	}
	return result, nil
}

func FindUserByEmail(collection *mongo.Collection, email string) (*models.User, error) {
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Println("Error finding user by email:", err)
		return nil, err
	}
	return &user, nil
}
