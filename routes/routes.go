package routes

import (
	"dbl-blog-backend/handlers"
	"dbl-blog-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the API routes
func SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Configure trusted proxies based on environment
	if gin.Mode() == gin.ReleaseMode {
		// For production (Vercel, Railway, etc.), trust common proxy networks
		router.SetTrustedProxies([]string{
			"10.0.0.0/8",     // Private networks
			"172.16.0.0/12",  // Private networks
			"192.168.0.0/16", // Private networks
			"127.0.0.1",      // Localhost
		})
	} else {
		// For local development
		router.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	}

	// Add CORS middleware
	router.Use(middleware.CorsMiddleware())

	// Add logging middleware
	router.Use(gin.Logger())

	// Add recovery middleware
	router.Use(gin.Recovery())

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Blog posts routes
		posts := v1.Group("/posts")
		{
			posts.POST("", handlers.CreatePost)
			posts.GET("", handlers.GetPosts)
			posts.GET("/:id", handlers.GetPost)
			posts.PUT("/:id", handlers.UpdatePost)
			posts.DELETE("/:id", handlers.DeletePost)
			posts.POST("/:id/like", handlers.LikePost)
			posts.POST("/:id/view", handlers.ViewPost)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Blog API is running",
		})
	})

	return router
}
