package handler

import (
	"net/http"

	"dbl-blog-backend/database"
	"dbl-blog-backend/routes"

	"github.com/gin-gonic/gin"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize database connection
	database.Connect()

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Setup routes (this returns a configured router)
	router := routes.SetupRoutes()

	// Handle the request
	router.ServeHTTP(w, r)
}
