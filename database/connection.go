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

	// Check if MONGODB_URI is provided directly (for Docker/production)
	mongoURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("DB_NAME")

	if mongoURI == "" {
		// Fallback: Build URI from individual components (for local development)
		username := os.Getenv("MONGODB_USERNAME")
		password := os.Getenv("MONGODB_PASSWORD")
		host := os.Getenv("MONGODB_HOST")
		port := os.Getenv("MONGODB_PORT")
		authSource := os.Getenv("MONGODB_AUTH_SOURCE")

		// Set defaults for local development
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "27017"
		}

		// Construct MongoDB URI dynamically
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
	}

	// Set default database name if not provided
	if dbName == "" {
		dbName = "dbl_blog"
	}

	// Debug logging for Vercel deployment
	log.Printf("=== MONGODB CONNECTION DEBUG ===")
	log.Printf("Environment: %s", os.Getenv("VERCEL_ENV"))
	log.Printf("DB_NAME: %s", dbName)

	if mongoURI != "" {
		// Don't log the full URI (contains credentials), just confirm it exists
		log.Printf("MONGODB_URI: [SET - %d characters]", len(mongoURI))
	} else {
		log.Printf("MONGODB_URI: [NOT SET]")
	}

	log.Printf("Attempting MongoDB connection...")

	// Create MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Printf("❌ MONGODB CONNECTION FAILED: %s", err)
		log.Printf("🔍 Check: 1) MONGODB_URI format 2) Atlas cluster status 3) Network access")
		log.Fatal("MongoDB connection failed")
	}

	log.Printf("✅ MongoDB client created, testing connection...")

	// Test the connection
	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Printf("❌ MONGODB PING FAILED: %s", err)
		log.Printf("🔍 Check: 1) Atlas cluster running 2) IP whitelist 3) Credentials")
		log.Fatal("MongoDB ping failed")
	}

	Database = Client.Database(dbName)
	log.Printf("🚀 Successfully connected to MongoDB database: %s", dbName)
	log.Printf("=== CONNECTION SUCCESS ===")
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
