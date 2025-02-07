package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"

	"github.com/Sahil1728/emotional-support/internal/database"
	"github.com/Sahil1728/emotional-support/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FIREBASE stuff
var authClient *auth.Client

func initializeFirebase() {
	opt := option.WithCredentialsFile("firebase_config.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	authClient, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
	fmt.Println("Firebase initialized")
}

func HandleFirebaseSignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := (&auth.UserToCreate{}).
			Email(req.Email).
			Password(req.Password)

		user, err := authClient.CreateUser(context.Background(), params)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "uid": user.UID})
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := authHeader[len("Bearer "):]
		token, err := authClient.VerifyIDToken(context.Background(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", token.UID)
		c.Next()
	}
}

var client *mongo.Client

// POST RELATED METHODS
func HandleCreatePost(collection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var post models.Post
		// bind json input
		if err := ctx.BindJSON(&post); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// insert post
		_, err := database.CreatePost(collection, post)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"message": "Post created successfully"})
	}
}

func HandleGetPosts(collection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		posts, err := database.GetPosts(collection)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, posts)
	}
}

// USER RELATED METHODS
func HandleRegisterUser(collection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user models.User
		// bind json input
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// if already exists
		existingUser, _ := database.FindUserByEmail(collection, user.Email)
		if existingUser != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		// insert user
		_, err := database.CreateUser(collection, user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
	}
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}
	fmt.Println("Connected to MongoDB")
	defer client.Disconnect(context.TODO())

	// Initialize Gin router
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")

	// select database and collection
	// db := client.Database("test")
	// user_collection := db.Collection("users")
	// post_collection := db.Collection("posts")
	// initializeFirebase()
	initializeFirebase()

	// registering routes
	registerRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on port " + port)
	router.Run(":" + port)
}

// func registerRoutes(router *gin.Engine) {
// 	// select database and collection
// 	db := client.Database("test")
// 	user_collection := db.Collection("users")
// 	post_collection := db.Collection("posts")

// 	// POST RELATED ROUTES
// 	router.POST("/posts", HandleCreatePost(post_collection))
// 	router.GET("/posts", HandleGetPosts(post_collection))

// 	// USER RELATED ROUTES
// 	router.POST("/users", HandleRegisterUser(user_collection))

// 	// AUTHENTICATION MIDDLEWARE
// 	router.POST("/posts", AuthMiddleware(), HandleCreatePost(post_collection))

// 	router.POST("/signup", HandleFirebaseSignUp())

// 	// Routes
// 	router.GET("/", func(c *gin.Context) {
// 		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Welcome"})
// 	})
// 	router.GET("/message", func(c *gin.Context) {
// 		c.String(http.StatusOK, "<p>Hello! You are not alone. 💙</p>")
// 	})
// }

func registerRoutes(router *gin.Engine) {
	db := client.Database("test")
	user_collection := db.Collection("users")
	post_collection := db.Collection("posts")

	// Create a group with authentication middleware
	authRoutes := router.Group("/")
	authRoutes.Use(AuthMiddleware())
	{
		authRoutes.POST("/posts", HandleCreatePost(post_collection))
	}

	// Public routes
	router.GET("/posts", HandleGetPosts(post_collection))
	router.POST("/users", HandleRegisterUser(user_collection))
	router.POST("/signup", HandleFirebaseSignUp())

	// Routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Welcome"})
	})
	router.GET("/message", func(c *gin.Context) {
		c.String(http.StatusOK, "<p>Hello! You are not alone. 💙</p>")
	})
}
