package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"dbl-blog-backend/database"
	"dbl-blog-backend/models"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test constants for E2E testing against live API
const (
	postsEndpoint     = "/api/v1/posts"
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
	responseNotNil    = "Response should not be nil"
	apiKeyHeader      = "X-API-Key"
	updatedTitle      = "Updated E2E Test Post"
	updatedContent    = "Updated content via E2E test"
)

// getAPIBaseURL returns the API base URL from environment or default
func getAPIBaseURL() string {
	if url := os.Getenv("API_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

// getValidAPIKey retrieves the first valid API key from environment
func getValidAPIKey() string {
	adminAPIKeys := os.Getenv("ADMIN_API_KEYS")
	if adminAPIKeys == "" {
		// Fallback for testing
		return "test-api-key-123"
	}

	// Get the first key from the comma-separated list
	keys := strings.Split(adminAPIKeys, ",")
	if len(keys) > 0 {
		return strings.TrimSpace(keys[0])
	}

	return "test-api-key-123"
}

// setupE2ETestDB sets up database for testing against live API
func setupE2ETestDB() func() {
	// Load environment variables from .env file
	_ = godotenv.Load() // Ignore error as .env file may not exist in test environment

	// Store original database connection and environment
	originalDB := database.Database
	originalClient := database.Client
	originalEnvVars := map[string]string{
		"DB_NAME":             os.Getenv("DB_NAME"),
		"MONGODB_HOST":        os.Getenv("MONGODB_HOST"),
		"MONGODB_PORT":        os.Getenv("MONGODB_PORT"),
		"MONGODB_USERNAME":    os.Getenv("MONGODB_USERNAME"),
		"MONGODB_PASSWORD":    os.Getenv("MONGODB_PASSWORD"),
		"MONGODB_AUTH_SOURCE": os.Getenv("MONGODB_AUTH_SOURCE"),
	}

	// Set up test database environment - use existing env vars or Docker defaults
	if os.Getenv("DB_NAME") == "" {
		_ = os.Setenv("DB_NAME", "dbl_blog_e2e_test")
	}
	if os.Getenv("MONGODB_HOST") == "" {
		_ = os.Setenv("MONGODB_HOST", "localhost")
	}
	if os.Getenv("MONGODB_PORT") == "" {
		_ = os.Setenv("MONGODB_PORT", "27017")
	}
	if os.Getenv("MONGODB_USERNAME") == "" {
		_ = os.Setenv("MONGODB_USERNAME", "admin")
	}
	if os.Getenv("MONGODB_PASSWORD") == "" {
		_ = os.Setenv("MONGODB_PASSWORD", "password")
	}
	if os.Getenv("MONGODB_AUTH_SOURCE") == "" {
		_ = os.Setenv("MONGODB_AUTH_SOURCE", "admin")
	}

	// Use the ACTUAL database connection logic from the API
	database.Connect()

	// Clean up any existing test data
	if database.Database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = database.Database.Collection("posts").Drop(ctx) // Ignore error in test cleanup
	}

	return func() {
		// Clean up test data
		if database.Database != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = database.Database.Collection("posts").Drop(ctx) // Ignore error in test cleanup
		}

		// Disconnect using the real API's disconnect function
		database.Disconnect()

		// Restore original connections and environment variables
		database.Database = originalDB
		database.Client = originalClient

		// Restore all original environment variables
		for key, value := range originalEnvVars {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}
}

// TestLiveE2ECreatePostWithValidAPIKey tests against a RUNNING API server
// To run this test:
// 1. Start the API: go run main.go
// 2. In another terminal: go test -run TestLiveE2E -v
func TestE2ECreatePostWithValidAPIKey(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Test post data
	testPost := models.Post{
		Title:   "Live E2E Test Post",
		Content: "This is a test post content for LIVE E2E testing against running API",
		Slug:    "livee2etestpost",
		Tags:    []string{"live", "e2e", "test"},
	}

	postJSON, _ := json.Marshal(testPost)

	// Make REAL HTTP request to RUNNING API server
	req, _ := http.NewRequest("POST", getAPIBaseURL()+postsEndpoint, bytes.NewBuffer(postJSON))
	req.Header.Set(contentTypeHeader, applicationJSON)
	req.Header.Set(apiKeyHeader, getValidAPIKey())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	// Assertions
	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response, "id")
		assert.Equal(t, testPost.Title, response["title"])
	}
}

// TestLiveE2ECreatePostWithoutAPIKey tests real auth failure against live API
func TestE2ECreatePostWithoutAPIKey(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	testPost := models.Post{
		Title:   "Unauthorized Live Post",
		Content: "This should fail against LIVE API",
		Slug:    "unauthorizedlivepost",
		Tags:    []string{"unauthorized"},
	}

	postJSON, _ := json.Marshal(testPost)

	// Make request WITHOUT X-API-Key header to RUNNING API
	req, _ := http.NewRequest("POST", getAPIBaseURL()+postsEndpoint, bytes.NewBuffer(postJSON))
	req.Header.Set(contentTypeHeader, applicationJSON)
	// No X-API-Key header

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	// Should get 401 Unauthorized
	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

// TestE2EGetPosts tests the GET posts endpoint against live API
func TestE2EGetPosts(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create test posts directly in database
	collection := database.Database.Collection("posts")
	testPosts := []interface{}{
		models.Post{
			Title:     "E2E Test Post 1",
			Content:   "Content 1 for E2E testing",
			Slug:      "e2e-test-post-1",
			Published: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		models.Post{
			Title:     "E2E Test Post 2",
			Content:   "Content 2 for E2E testing",
			Slug:      "e2e-test-post-2",
			Published: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	_, err := collection.InsertMany(context.Background(), testPosts)
	assert.NoError(t, err)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test getting all posts
	req, _ := http.NewRequest("GET", getAPIBaseURL()+postsEndpoint, nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response, "posts")
		assert.Equal(t, float64(2), response["total"])
	}

	// Test filtering by published status
	req, _ = http.NewRequest("GET", getAPIBaseURL()+postsEndpoint+"?published=true", nil)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["total"])
	}
}

// TestE2ELikePost tests the like post endpoint against live API
func TestE2ELikePost(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post directly in database
	post := models.Post{
		Title:     "E2E Like Test Post",
		Content:   "Test Content for liking",
		Slug:      "e2e-like-test-post",
		Published: true,
		Likes:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), post)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test liking the post
	req, _ := http.NewRequest("PUT", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex()+"/like", nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response, "message")
	}

	// Verify the like was recorded in database
	var updatedPost models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&updatedPost)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), updatedPost.Likes)
}

// TestE2EDislikePost tests the dislike post endpoint against live API
func TestE2EDislikePost(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post with some likes
	post := models.Post{
		Title:     "Test Post for Dislike",
		Content:   "Content for testing dislike functionality",
		Slug:      "test-dislike-post",
		Published: true,
		Likes:     3, // Start with 3 likes
		Views:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), post)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test disliking the post
	req, _ := http.NewRequest("PUT", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex()+"/dislike", nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "likes")

		// Should return decremented like count
		likes, ok := response["likes"].(float64) // JSON numbers are float64
		assert.True(t, ok)
		assert.Equal(t, float64(2), likes) // Should be decremented from 3 to 2
	}

	// Verify the dislike was recorded in database
	var updatedPost models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&updatedPost)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), updatedPost.Likes)

	// Test disliking a post with 0 likes
	// First set likes to 0
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": postID},
		bson.M{"$set": bson.M{"likes": 0}},
	)
	assert.NoError(t, err)

	// Try to dislike again
	req, _ = http.NewRequest("PUT", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex()+"/dislike", nil)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "likes")

		// Should remain at 0
		likes, ok := response["likes"].(float64)
		assert.True(t, ok)
		assert.Equal(t, float64(0), likes)
	}

	// Verify likes count remains 0 in database
	err = collection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&updatedPost)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), updatedPost.Likes)
}

// TestE2EGetSinglePost tests fetching a single post by ID
func TestE2EGetSinglePost(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post in database
	testPost := models.Post{
		Title:     "E2E Single Post Test",
		Content:   "Content for single post E2E test",
		Slug:      "e2e-single-post-test",
		Published: true,
		Views:     5, // Start with some views
		Likes:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), testPost)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test GET single post by ID
	req, _ := http.NewRequest("GET", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex(), nil)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var post models.Post
		err = json.NewDecoder(resp.Body).Decode(&post)
		assert.NoError(t, err)
		assert.Equal(t, testPost.Title, post.Title)
		assert.Equal(t, testPost.Content, post.Content)
		assert.Equal(t, testPost.Slug, post.Slug)
		assert.Equal(t, int64(5), post.Views) // Should remain 5 (no auto-increment)
	}

	// Test GET single post by slug
	req, _ = http.NewRequest("GET", getAPIBaseURL()+postsEndpoint+"/"+testPost.Slug, nil)

	resp, err = client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var post models.Post
		err = json.NewDecoder(resp.Body).Decode(&post)
		assert.NoError(t, err)
		assert.Equal(t, testPost.Title, post.Title)
		assert.Equal(t, int64(5), post.Views) // Should still remain 5
	}
}

// TestE2ETrackPostView tests the track post view endpoint
func TestE2ETrackPostView(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post in database
	testPost := models.Post{
		Title:     "E2E View Tracking Test",
		Content:   "Content for view tracking E2E test",
		Slug:      "e2e-view-tracking-test",
		Published: true,
		Views:     10, // Start with 10 views
		Likes:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), testPost)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test PUT track view
	req, _ := http.NewRequest("PUT", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex()+"/view", nil)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "views")

		// Should return incremented view count
		views, ok := response["views"].(float64) // JSON numbers are float64
		assert.True(t, ok)
		assert.Equal(t, float64(11), views) // Should be incremented from 10 to 11
	}

	// Verify the view was recorded in database
	var updatedPost models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&updatedPost)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), updatedPost.Views)

	// Verify view record was created in post_views collection
	viewsCollection := database.Database.Collection("post_views")
	var viewRecord models.PostView
	err = viewsCollection.FindOne(context.Background(), bson.M{"post_id": postID}).Decode(&viewRecord)
	assert.NoError(t, err)
	assert.Equal(t, postID, viewRecord.PostID)
	assert.NotEmpty(t, viewRecord.ViewedAt)
}

// TestE2EUpdatePost tests the update post endpoint against live API
func TestE2EUpdatePost(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post directly in database
	originalPost := models.Post{
		Title:     "E2E Update Test Post",
		Content:   "Original content for updating",
		Slug:      "e2e-update-test-post",
		Summary:   "Original summary",
		Tags:      []string{"original", "test"},
		Published: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), originalPost)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	// Update data
	updateData := map[string]interface{}{
		"title":     updatedTitle,
		"content":   updatedContent,
		"slug":      "e2e-updated-test-post",
		"summary":   "Updated summary",
		"tags":      []string{"updated", "e2e", "test"},
		"published": true,
	}

	updateJSON, _ := json.Marshal(updateData)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test updating the post
	req, _ := http.NewRequest("PUT", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex(), bytes.NewBuffer(updateJSON))
	req.Header.Set(contentTypeHeader, applicationJSON)
	req.Header.Set(apiKeyHeader, getValidAPIKey())

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		// Debug: Print response body if status is not 200
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Error response body: %s", string(body))
			// Reset the body for further reading if needed
			resp.Body = io.NopCloser(bytes.NewReader(body))
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.Post
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, updatedTitle, response.Title)
		assert.Equal(t, updatedContent, response.Content)
		assert.Equal(t, true, response.Published)
	}

	// Verify the update was persisted in database
	var updatedPost models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&updatedPost)
	assert.NoError(t, err)
	assert.Equal(t, updatedTitle, updatedPost.Title)
	assert.Equal(t, updatedContent, updatedPost.Content)
	assert.Equal(t, true, updatedPost.Published)
}

// TestE2EDeletePost tests the delete post endpoint against live API
func TestE2EDeletePost(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post directly in database
	testPost := models.Post{
		Title:     "E2E Delete Test Post",
		Content:   "Content for deletion test",
		Slug:      "e2e-delete-test-post",
		Published: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), testPost)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test deleting the post
	req, _ := http.NewRequest("DELETE", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex(), nil)
	req.Header.Set(apiKeyHeader, getValidAPIKey())

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Contains(t, response, "message")
		assert.Equal(t, "Post deleted successfully", response["message"])
	}

	// Verify the post was deleted from database
	var deletedPost models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&deletedPost)
	assert.Error(t, err) // Should return "no documents in result" error
}

// TestE2EUpdatePostWithoutAuth tests update without authentication
func TestE2EUpdatePostWithoutAuth(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post
	testPost := models.Post{
		Title:     "Test Post for Unauthorized Update",
		Content:   "Content",
		Slug:      "test-unauthorized-update",
		Published: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), testPost)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	updateData := map[string]interface{}{
		"title": "Should Not Update",
	}
	updateJSON, _ := json.Marshal(updateData)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test updating without X-API-Key header
	req, _ := http.NewRequest("PUT", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex(), bytes.NewBuffer(updateJSON))
	req.Header.Set(contentTypeHeader, applicationJSON)
	// No X-API-Key header

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	// Should get 401 Unauthorized
	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

// TestE2EDeletePostWithoutAuth tests delete without authentication
func TestE2EDeletePostWithoutAuth(t *testing.T) {
	cleanup := setupE2ETestDB()
	defer cleanup()

	// Create a test post
	testPost := models.Post{
		Title:     "Test Post for Unauthorized Delete",
		Content:   "Content",
		Slug:      "test-unauthorized-delete",
		Published: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), testPost)
	assert.NoError(t, err)

	postID := result.InsertedID.(primitive.ObjectID)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test deleting without X-API-Key header
	req, _ := http.NewRequest("DELETE", getAPIBaseURL()+postsEndpoint+"/"+postID.Hex(), nil)
	// No X-API-Key header

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	// Should get 401 Unauthorized
	if assert.NotNil(t, resp, responseNotNil) {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

// Example of how to run these tests:
//
// Terminal 1: Start the API
// $ go run main.go
//
// Terminal 2: Run E2E tests
// $ go test -run TestE2E -v
//
// This approach tests the ACTUAL running API server, including:
// - Real HTTP server behavior
// - Real routing and middleware stack
// - Real database connections
// - Real environment configuration
// - Network timeouts and HTTP client behavior
