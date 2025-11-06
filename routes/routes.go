package routes

import (
	"os"

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
		_ = router.SetTrustedProxies([]string{
			"10.0.0.0/8",     // Private networks
			"172.16.0.0/12",  // Private networks
			"192.168.0.0/16", // Private networks
			"127.0.0.1",      // Localhost
		})
	} else {
		// For local development
		_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	}

	// Add CORS middleware
	router.Use(middleware.CorsMiddleware())

	// Add logging middleware
	router.Use(gin.Logger())

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add input sanitization middleware
	router.Use(middleware.InputSanitizationMiddleware())

	// API v1 group
	v1 := router.Group("/api/v1")

	// Optional: Add public rate limiting (controlled by environment variable)
	// This provides an extra layer of protection beyond Vercel's built-in limits
	if os.Getenv("ENABLE_PUBLIC_RATE_LIMIT") == "true" {
		v1.Use(middleware.PublicRateLimitMiddleware())
	}
	{
		// Blog posts routes
		posts := v1.Group("/posts")
		{
			// Public endpoints (no authentication required)
			posts.GET("", handlers.GetPosts)                // Get all posts
			posts.GET("/:id", handlers.GetPost)             // Get single post
			posts.PUT("/:id/like", handlers.LikePost)       // Like a post
			posts.PUT("/:id/dislike", handlers.DislikePost) // Dislike a post
			posts.PUT("/:id/view", handlers.ViewPost)       // Track post view

			// Protected endpoints (admin only)
			adminPosts := posts.Group("", middleware.AdminRateLimitMiddleware(), middleware.AdminAuthMiddleware())
			{
				adminPosts.POST("", handlers.CreatePost)       // Create post
				adminPosts.PUT("/:id", handlers.UpdatePost)    // Update post
				adminPosts.DELETE("/:id", handlers.DeletePost) // Delete post
			}
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
