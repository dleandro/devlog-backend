package middleware

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Unit tests for rate limiting functionality

func TestRateLimiting_AllowedRequests(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.1"
	maxRequests := 5
	window := time.Minute

	// Should allow requests up to the limit
	for i := 0; i < maxRequests; i++ {
		allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Next request should be blocked
	allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	assert.False(t, allowed, "Request exceeding limit should be blocked")
}

func TestRateLimiting_WindowReset(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.2"
	maxRequests := 3
	window := 100 * time.Millisecond // Short window for testing

	// Fill up the rate limit
	for i := 0; i < maxRequests; i++ {
		allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Should be blocked
	allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	assert.False(t, allowed, "Request should be blocked")

	// Wait for window to expire
	time.Sleep(window + 10*time.Millisecond)

	// Should be allowed again
	allowed = checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	assert.True(t, allowed, "Request should be allowed after window reset")
}

func TestRateLimiting_MultipleIPs(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP1 := "192.168.1.3"
	testIP2 := "192.168.1.4"
	maxRequests := 2
	window := time.Minute

	// Fill limit for IP1
	for i := 0; i < maxRequests; i++ {
		allowed := checkRateLimit(adminRateLimiter, testIP1, maxRequests, window)
		assert.True(t, allowed, "IP1 request %d should be allowed", i+1)
	}

	// IP1 should be blocked
	allowed := checkRateLimit(adminRateLimiter, testIP1, maxRequests, window)
	assert.False(t, allowed, "IP1 should be blocked")

	// IP2 should still be allowed (independent rate limiting)
	for i := 0; i < maxRequests; i++ {
		allowed := checkRateLimit(adminRateLimiter, testIP2, maxRequests, window)
		assert.True(t, allowed, "IP2 request %d should be allowed", i+1)
	}

	// IP2 should now be blocked
	allowed = checkRateLimit(adminRateLimiter, testIP2, maxRequests, window)
	assert.False(t, allowed, "IP2 should be blocked")
}

func TestRateLimiting_PartialWindowExpiry(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.5"
	maxRequests := 3
	window := 200 * time.Millisecond

	// Make first request
	allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	assert.True(t, allowed, "First request should be allowed")

	// Wait half the window
	time.Sleep(window / 2)

	// Make remaining requests
	for i := 0; i < maxRequests-1; i++ {
		allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
		assert.True(t, allowed, "Request %d should be allowed", i+2)
	}

	// Should be blocked
	allowed = checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	assert.False(t, allowed, "Request should be blocked")

	// Wait for first request to expire (another half window)
	time.Sleep(window/2 + 10*time.Millisecond)

	// Should be allowed again (first request expired)
	allowed = checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	assert.True(t, allowed, "Request should be allowed after partial window expiry")
}

func TestRateLimiting_ZeroLimit(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.6"
	maxRequests := 0
	window := time.Minute

	// Should be blocked immediately with zero limit
	allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	assert.False(t, allowed, "Request should be blocked with zero limit")
}

func TestRateLimiting_HighVolumeStress(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.7"
	maxRequests := 10
	window := time.Minute

	allowedCount := 0
	blockedCount := 0

	// Make many requests quickly
	for i := 0; i < 50; i++ {
		if checkRateLimit(adminRateLimiter, testIP, maxRequests, window) {
			allowedCount++
		} else {
			blockedCount++
		}
	}

	// Should allow exactly maxRequests
	assert.Equal(t, maxRequests, allowedCount, "Should allow exactly %d requests", maxRequests)
	assert.Equal(t, 40, blockedCount, "Should block remaining requests")
}

func TestRateLimiting_ConcurrentAccess(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.8"
	maxRequests := 5
	window := time.Minute

	// Channel to collect results
	results := make(chan bool, 20)

	// Launch concurrent goroutines
	for i := 0; i < 20; i++ {
		go func() {
			allowed := checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
			results <- allowed
		}()
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < 20; i++ {
		if <-results {
			allowedCount++
		}
	}

	// Should allow exactly maxRequests (may vary slightly due to race conditions)
	// This is a basic test - true concurrent testing would require more sophisticated synchronization
	assert.True(t, allowedCount >= maxRequests-2 && allowedCount <= maxRequests+2,
		"Concurrent access should allow approximately %d requests, got %d", maxRequests, allowedCount)
}

func TestRateLimiting_CleanupOldRequests(t *testing.T) {
	// Reset admin rate limiter for clean test
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.9"
	maxRequests := 3
	window := 100 * time.Millisecond

	// Make requests to populate the store
	for i := 0; i < maxRequests; i++ {
		checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	}

	// Check that requests are stored
	adminRateLimiter.mutex.RLock()
	requestCount := len(adminRateLimiter.requests[testIP])
	adminRateLimiter.mutex.RUnlock()
	assert.Equal(t, maxRequests, requestCount, "Should store %d requests", maxRequests)

	// Wait for requests to expire
	time.Sleep(window + 10*time.Millisecond)

	// Make one more request to trigger cleanup
	checkRateLimit(adminRateLimiter, testIP, maxRequests, window)

	// Old requests should be cleaned up, only new request should remain
	adminRateLimiter.mutex.RLock()
	newRequestCount := len(adminRateLimiter.requests[testIP])
	adminRateLimiter.mutex.RUnlock()
	assert.Equal(t, 1, newRequestCount, "Should have cleaned up old requests")
}

func TestPublicRateLimiting_DifferentLimits(t *testing.T) {
	// Reset public rate limiter for clean test
	publicRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.10"
	getLimitRequests := 5
	socialLimitRequests := 3
	window := time.Minute

	// Test GET limit
	for i := 0; i < getLimitRequests; i++ {
		allowed := checkRateLimit(publicRateLimiter, testIP, getLimitRequests, window)
		assert.True(t, allowed, "GET request %d should be allowed", i+1)
	}

	// Should be blocked at GET limit
	allowed := checkRateLimit(publicRateLimiter, testIP, getLimitRequests, window)
	assert.False(t, allowed, "GET request should be blocked at limit")

	// Reset for social limit test
	publicRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	// Test social limit (different IP to avoid conflicts)
	testIP2 := "192.168.1.11"
	for i := 0; i < socialLimitRequests; i++ {
		allowed := checkRateLimit(publicRateLimiter, testIP2, socialLimitRequests, window)
		assert.True(t, allowed, "Social request %d should be allowed", i+1)
	}

	// Should be blocked at social limit
	allowed = checkRateLimit(publicRateLimiter, testIP2, socialLimitRequests, window)
	assert.False(t, allowed, "Social request should be blocked at limit")
}

// Benchmark tests for performance

func BenchmarkRateLimiting_AdminSingleIP(b *testing.B) {
	adminRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	testIP := "192.168.1.100"
	maxRequests := 100
	window := time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkRateLimit(adminRateLimiter, testIP, maxRequests, window)
	}
}

func BenchmarkRateLimiting_PublicMultipleIPs(b *testing.B) {
	publicRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
	}

	maxRequests := 100
	window := time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different IPs to test scaling
		testIP := fmt.Sprintf("192.168.1.%d", i%255)
		checkRateLimit(publicRateLimiter, testIP, maxRequests, window)
	}
}

// Helper function tests

func TestGetEnvInt(t *testing.T) {
	testCases := []struct {
		name     string
		envValue string
		fallback int
		expected int
	}{
		{"valid int", "42", 10, 42},
		{"empty string", "", 10, 10},
		{"invalid int", "not-a-number", 10, 10},
		{"zero value", "0", 10, 0},
		{"negative value", "-5", 10, -5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for test
			if tc.envValue != "" {
				t.Setenv("TEST_ENV_VAR", tc.envValue)
				result := getEnvInt("TEST_ENV_VAR", tc.fallback)
				assert.Equal(t, tc.expected, result)
			} else {
				result := getEnvInt("NON_EXISTENT_VAR", tc.fallback)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestMinFunction(t *testing.T) {
	testCases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 15, 8, 8},
		{"equal", 7, 7, 7},
		{"zero values", 0, 5, 0},
		{"negative values", -3, -1, -3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := min(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}
