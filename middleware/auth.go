package middleware

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"

	"dbl-blog-backend/apierrors"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware validates admin API key with enhanced security features
func AdminAuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		log.Printf("[INFO] AdminAuth: Checking authorization for %s %s from %s", c.Request.Method, c.Request.URL.Path, clientIP)

		// Get API keys from environment (comma-separated for multiple keys)
		adminAPIKeys := os.Getenv("ADMIN_API_KEYS")

		if adminAPIKeys == "" {
			log.Printf("[ERROR] AdminAuth: No admin API keys configured")
			apierrors.RespondWithCustomError(c, http.StatusInternalServerError, "SERVER_MISCONFIGURATION", "Server configuration error", "Admin API keys not configured")
			c.Abort()
			return
		}

		// Get X-API-Key header
		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			log.Printf("[SECURITY] AdminAuth: Missing X-API-Key header from %s", clientIP)
			apierrors.RespondMissingAuthorization(c)
			c.Abort()
			return
		}
		// Validate against all configured API keys using constant-time comparison
		validKeys := strings.Split(adminAPIKeys, ",")
		isValid := false

		for _, validKey := range validKeys {
			validKey = strings.TrimSpace(validKey)
			if validKey != "" && subtle.ConstantTimeCompare([]byte(providedKey), []byte(validKey)) == 1 {
				isValid = true
				break
			}
		}

		if !isValid {
			log.Printf("[SECURITY] AdminAuth: Invalid API key attempt from %s (key: %s...)", clientIP, providedKey[:min(8, len(providedKey))])
			apierrors.RespondInvalidAPIKey(c)
			c.Abort()
			return
		}

		log.Printf("[SUCCESS] AdminAuth: Valid API key for %s %s from %s", c.Request.Method, c.Request.URL.Path, clientIP)
		c.Next()
	})
}
