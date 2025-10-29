package main

import (
	"log"
	"os"

	"dbl-blog-backend/database"
	"dbl-blog-backend/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	
	log.Printf("Setting up database connection")

	// Connect to database
	database.Connect()

	// Create indexes for better performance
	database.CreateIndexes()
	
	// Setup routes
	router := routes.SetupRoutes()
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
