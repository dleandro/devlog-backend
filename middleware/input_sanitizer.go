package middleware

import (
	"log"
	"regexp"
	"strings"

	"dbl-blog-backend/apierrors"

	"github.com/gin-gonic/gin"
)

// InputSanitizationMiddleware prevents potentially dangerous input
func InputSanitizationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check URL parameters for suspicious patterns
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSuspiciousPatterns(value) {
					log.Printf("[SECURITY] Suspicious query parameter detected: %s=%s from %s", key, value, c.ClientIP())
					apierrors.RespondWithCustomError(c, 400, "INVALID_INPUT", "Invalid characters in request", "Request contains potentially dangerous patterns")
					c.Abort()
					return
				}
			}
		}

		// Check path parameters
		for _, param := range c.Params {
			if containsSuspiciousPatterns(param.Value) {
				log.Printf("[SECURITY] Suspicious path parameter detected: %s=%s from %s", param.Key, param.Value, c.ClientIP())
				apierrors.RespondWithCustomError(c, 400, "INVALID_INPUT", "Invalid characters in request", "Request contains potentially dangerous patterns")
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// containsSuspiciousPatterns checks for common NoSQL injection patterns
func containsSuspiciousPatterns(input string) bool {
	// Convert to lowercase for case-insensitive matching
	lower := strings.ToLower(input)

	// Common NoSQL injection patterns
	suspiciousPatterns := []string{
		"$where",
		"$ne",
		"$gt",
		"$lt",
		"$regex",
		"$or",
		"$and",
		"$nor",
		"$not",
		"$exists",
		"$type",
		"$mod",
		"$text",
		"$search",
		"javascript:",
		"<script",
		"eval(",
		"function(",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Check for potential JavaScript injection in MongoDB
	jsRegex := regexp.MustCompile(`(?i)(function\s*\(|eval\s*\(|this\.|document\.|window\.)`)
	if jsRegex.MatchString(input) {
		return true
	}

	return false
}
