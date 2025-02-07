package models

import "time"

// User structure represents the user model
type User struct {
	ID       string    `bson:"_id"`
	Email    string    `bson:"email"`
	Username string    `bson:"username"`
	Password string    `bson:"password"` // hashed password
	Created  time.Time `bson:"created"`
}

// Post structure represents the post model
type Post struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	Content   string    `bson:"content"`
	Anonymous bool      `bson:"anonymous"`
	Created   time.Time `bson:"created"`
}
