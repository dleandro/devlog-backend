package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

// Connect initializes the MongoDB connection
func Connect() {
	var err error

	// Get MongoDB connection parameters from environment
	username := os.Getenv("MONGODB_USERNAME")
	password := os.Getenv("MONGODB_PASSWORD")
	host := os.Getenv("MONGODB_HOST")
	port := os.Getenv("MONGODB_PORT")
	authSource := os.Getenv("MONGODB_AUTH_SOURCE")
	dbName := os.Getenv("DB_NAME")

	// Set defaults
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "27017"
	}
	if dbName == "" {
		dbName = "dbl_blog"
	}

	// Construct MongoDB URI dynamically
	var mongoURI string
	if username != "" && password != "" {
		// With authentication
		if authSource != "" {
			mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
				username, password, host, port, dbName, authSource)
		} else {
			mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
				username, password, host, port, dbName)
		}
	} else {
		// Without authentication (local development)
		mongoURI = fmt.Sprintf("mongodb://%s:%s", host, port)
	}

	// Create MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Test the connection
	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	Database = Client.Database(dbName)
	log.Printf("Connected to MongoDB database: %s", dbName)
}

// CreateIndexes creates necessary indexes for better performance
func CreateIndexes() {
	ctx := context.Background()

	// Create unique index on post slug
	postsCollection := Database.Collection("posts")
	_, err := postsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]int{"slug": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: Failed to create slug index: %v", err)
	}

	log.Println("Database indexes created successfully")
}

// Disconnect closes the MongoDB connection
func Disconnect() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}
