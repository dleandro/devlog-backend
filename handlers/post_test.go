package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"dbl-blog-backend/database"
	"dbl-blog-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB creates a test MongoDB connection
func setupTestDB() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get test MongoDB URI from environment or use defaults
	testURI := os.Getenv("TEST_MONGODB_URI")
	if testURI == "" {
		testURI = "mongodb://admin:password@localhost:27017/dbl_blog_test?authSource=admin"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testURI))
	if err != nil {
		// Fallback to non-auth MongoDB for local testing
		client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			panic("Failed to connect to test database: " + err.Error())
		}
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		panic("Failed to ping test database: " + err.Error())
	}

	// Use a test database
	testDB := client.Database("dbl_blog_test")

	// Set global database variables
	database.Client = client
	database.Database = testDB

	// Return cleanup function
	return func() {
		// Clean up test data
		testDB.Drop(context.Background())
		client.Disconnect(context.Background())
	}
}

func TestCreatePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	cleanup := setupTestDB()
	defer cleanup()

	router := gin.New()
	router.POST("/posts", CreatePost)

	post := models.Post{
		Title:   "Test Post",
		Content: "This is a test post content",
		Slug:    "test-post",
		Summary: "Test summary",
		Tags:    []string{"test", "golang"},
	}

	postJSON, _ := json.Marshal(post)

	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(postJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Post
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test Post", response.Title)
	assert.Equal(t, "test-post", response.Slug)
}

func TestGetPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	cleanup := setupTestDB()
	defer cleanup()

	// Create test posts
	collection := database.Database.Collection("posts")
	testPosts := []interface{}{
		models.Post{
			Title:     "Post 1",
			Content:   "Content 1",
			Slug:      "post-1",
			Published: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		models.Post{
			Title:     "Post 2",
			Content:   "Content 2",
			Slug:      "post-2",
			Published: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	_, err := collection.InsertMany(context.Background(), testPosts)
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/posts", GetPosts)

	// Test getting all posts
	req, _ := http.NewRequest("GET", "/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["total"])

	// Test filtering by published status
	req, _ = http.NewRequest("GET", "/posts?published=true", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["total"])
}

func TestLikePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	cleanup := setupTestDB()
	defer cleanup()

	// Create a test post
	post := models.Post{
		Title:     "Test Post",
		Content:   "Test Content",
		Slug:      "test-post",
		Published: true,
		Likes:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), post)
	assert.NoError(t, err)

	postID := result.InsertedID

	router := gin.New()
	router.POST("/posts/:id/like", LikePost)

	// Test liking the post
	req, _ := http.NewRequest("POST", "/posts/"+postID.(primitive.ObjectID).Hex()+"/like", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the like was recorded
	var updatedPost models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": postID}).Decode(&updatedPost)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), updatedPost.Likes)
}
