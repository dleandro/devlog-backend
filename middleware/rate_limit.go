package middleware

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"dbl-blog-backend/apierrors"

	"github.com/gin-gonic/gin"
)

// RateLimiter stores rate limiting data
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
}

// Global rate limiters for different endpoint types
var (
	adminRateLimiter  = &RateLimiter{requests: make(map[string][]time.Time)}
	publicRateLimiter = &RateLimiter{requests: make(map[string][]time.Time)}
)

// getEnvInt gets an environment variable as integer with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// AdminRateLimitMiddleware provides rate limiting for admin operations
func AdminRateLimitMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Get rate limit from environment (default: 30 requests per minute)
		maxRequests := getEnvInt("ADMIN_RATE_LIMIT_PER_MINUTE", 30)

		if !checkRateLimit(adminRateLimiter, clientIP, maxRequests, time.Minute) {
			log.Printf("[SECURITY] AdminRateLimit: Rate limit exceeded for IP %s (%d requests/minute)", clientIP, maxRequests)
			apierrors.RespondWithCustomError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", "Please wait before trying again")
			c.Abort()
			return
		}

		c.Next()
	})
}

// PublicRateLimitMiddleware provides gentle rate limiting for public endpoints
func PublicRateLimitMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Different limits based on endpoint type with environment configuration
		var maxRequests int
		var window time.Duration

		switch {
		case c.Request.Method == "GET":
			// GET requests - configurable (default: 120 per minute)
			maxRequests = getEnvInt("PUBLIC_GET_RATE_LIMIT_PER_MINUTE", 120)
			window = time.Minute
		case c.Request.URL.Path == "/api/v1/posts/:id/like" || c.Request.URL.Path == "/api/v1/posts/:id/view":
			// Social interactions - configurable (default: 60 per minute)
			maxRequests = getEnvInt("PUBLIC_SOCIAL_RATE_LIMIT_PER_MINUTE", 60)
			window = time.Minute
		default:
			// Default for other public endpoints
			maxRequests = getEnvInt("PUBLIC_DEFAULT_RATE_LIMIT_PER_MINUTE", 100)
			window = time.Minute
		}

		if !checkRateLimit(publicRateLimiter, clientIP, maxRequests, window) {
			log.Printf("[INFO] PublicRateLimit: Rate limit exceeded for IP %s on %s %s (%d requests/minute)",
				clientIP, c.Request.Method, c.Request.URL.Path, maxRequests)

			apierrors.RespondWithCustomError(c, http.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"Too many requests",
				"Please slow down and try again in a moment")
			c.Abort()
			return
		}

		c.Next()
	})
}

// checkRateLimit implements rate limiting logic
func checkRateLimit(limiter *RateLimiter, clientIP string, maxRequests int, window time.Duration) bool {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-window)

	// Get existing requests for this IP
	requests := limiter.requests[clientIP]

	// Filter out requests outside the current window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if under the limit
	if len(validRequests) >= maxRequests {
		return false
	}

	// Add current request and update
	validRequests = append(validRequests, now)
	limiter.requests[clientIP] = validRequests

	// Cleanup old entries periodically to prevent memory leaks
	if len(limiter.requests) > 1000 {
		cleanupRateLimit(limiter, windowStart)
	}

	return true
}

// cleanupRateLimit removes old entries to prevent memory leaks
func cleanupRateLimit(limiter *RateLimiter, cutoff time.Time) {
	for ip, requests := range limiter.requests {
		validRequests := make([]time.Time, 0)
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(limiter.requests, ip)
		} else {
			limiter.requests[ip] = validRequests
		}
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
