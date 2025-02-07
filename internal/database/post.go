package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Sahil1728/emotional-support/internal/models"
)

// Creating post
func CreatePost(collection *mongo.Collection, post models.Post) (*mongo.InsertOneResult, error) {
	post.Created = time.Now()
	result, err := collection.InsertOne(context.TODO(), post)
	if err != nil {
		log.Println("Error creating post:", err)
		return nil, err
	}
	return result, nil
}

// Retreiving all posts
func GetPosts(collection *mongo.Collection) ([]models.Post, error) {
	var posts []models.Post
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("Error getting posts:", err)
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var post models.Post
		cursor.Decode(&post)
		posts = append(posts, post)
	}
	return posts, nil
}
