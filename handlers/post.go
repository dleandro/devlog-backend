package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dbl-blog-backend/apierrors"
	"dbl-blog-backend/database"
	"dbl-blog-backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreatePost creates a new blog post
func CreatePost(c *gin.Context) {
	log.Printf("[INFO] CreatePost: Received request from %s", c.ClientIP())

	var post models.Post

	if err := c.ShouldBindJSON(&post); err != nil {
		log.Printf("[ERROR] CreatePost: Validation failed - %s", err.Error())
		apierrors.RespondWithValidationError(c, err.Error())
		return
	}

	// Generate slug from title if not provided
	if post.Slug == "" {
		post.Slug = generateSlug(post.Title)
	}

	// Set timestamps
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	// Insert into MongoDB
	collection := database.Database.Collection("posts")
	result, err := collection.InsertOne(context.Background(), post)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Printf("[ERROR] CreatePost: Duplicate key error for slug '%s'", post.Slug)
			apierrors.RespondPostAlreadyExists(c)
			return
		}
		log.Printf("[ERROR] CreatePost: Failed to insert post - %s", err.Error())
		apierrors.RespondFailedToCreatePost(c)
		return
	}

	post.ID = result.InsertedID.(primitive.ObjectID)
	log.Printf("[SUCCESS] CreatePost: Created post with ID %s, title: '%s'", post.ID.Hex(), post.Title)
	c.JSON(http.StatusCreated, post)
}

// GetPosts retrieves all blog posts with pagination
func GetPosts(c *gin.Context) {
	log.Printf("[INFO] GetPosts: Received request from %s", c.ClientIP())

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	published := c.Query("published")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	skip := (page - 1) * limit

	// Build filter
	filter := bson.M{}
	switch published {
	case "true":
		filter["published"] = true
	case "false":
		filter["published"] = false
	}

	collection := database.Database.Collection("posts")

	// Get total count
	total, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		apierrors.RespondFailedToCountPosts(c)
		return
	}

	// Find posts with pagination
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by newest first

	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		log.Printf("[ERROR] GetPosts: Failed to find posts - %s", err.Error())
		apierrors.RespondFailedToFetchPosts(c)
		return
	}
	defer func() { _ = cursor.Close(context.Background()) }()

	var posts []models.Post
	if err = cursor.All(context.Background(), &posts); err != nil {
		log.Printf("[ERROR] GetPosts: Failed to decode posts - %s", err.Error())
		apierrors.RespondFailedToDecodePosts(c)
		return
	}

	log.Printf("[SUCCESS] GetPosts: Retrieved %d posts (page %d, limit %d, total %d)", len(posts), page, limit, total)
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// GetPost retrieves a single blog post by ID or slug
func GetPost(c *gin.Context) {
	identifier := c.Param("id")
	log.Printf("[INFO] GetPost: Received request for identifier '%s' from %s", identifier, c.ClientIP())

	var post models.Post
	collection := database.Database.Collection("posts")

	// Try to parse as ObjectID first, then as slug
	var err error
	if objectID, parseErr := primitive.ObjectIDFromHex(identifier); parseErr == nil {
		err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&post)
	} else {
		err = collection.FindOne(context.Background(), bson.M{"slug": identifier}).Decode(&post)
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[ERROR] GetPost: Post not found for identifier '%s'", identifier)
			apierrors.RespondPostNotFound(c)
			return
		}
		log.Printf("[ERROR] GetPost: Failed to fetch post for identifier '%s' - %s", identifier, err.Error())
		apierrors.RespondFailedToFetchPost(c)
		return
	}

	// Increment view count asynchronously
	go incrementPostViews(post.ID, c.ClientIP(), c.GetHeader("User-Agent"))

	log.Printf("[SUCCESS] GetPost: Retrieved post '%s' (ID: %s)", post.Title, post.ID.Hex())
	c.JSON(http.StatusOK, post)
}

// UpdatePost updates an existing blog post
func UpdatePost(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[INFO] UpdatePost: Received request for post ID '%s' from %s", id, c.ClientIP())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		apierrors.RespondInvalidPostID(c)
		return
	}

	var updates models.Post
	if err := c.ShouldBindJSON(&updates); err != nil {
		log.Printf("[ERROR] UpdatePost: Validation failed for post ID '%s' - %s", id, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set updated timestamp
	updates.UpdatedAt = time.Now()

	// Create update document (exclude ID and created_at)
	updateDoc := bson.M{
		"$set": bson.M{
			"title":      updates.Title,
			"content":    updates.Content,
			"slug":       updates.Slug,
			"summary":    updates.Summary,
			"tags":       updates.Tags,
			"published":  updates.Published,
			"updated_at": updates.UpdatedAt,
		},
	}

	collection := database.Database.Collection("posts")
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		updateDoc,
	)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Printf("[ERROR] UpdatePost: Duplicate key error for post ID '%s'", id)
			apierrors.RespondPostAlreadyExists(c)
			return
		}
		log.Printf("[ERROR] UpdatePost: Failed to update post ID '%s' - %s", id, err.Error())
		apierrors.RespondFailedToUpdatePost(c)
		return
	}

	if result.MatchedCount == 0 {
		log.Printf("[ERROR] UpdatePost: Post not found for ID '%s'", id)
		apierrors.RespondPostNotFound(c)
		return
	}

	// Fetch and return updated post
	var updatedPost models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&updatedPost)
	if err != nil {
		log.Printf("[ERROR] UpdatePost: Failed to fetch updated post ID '%s' - %s", id, err.Error())
		apierrors.RespondFailedToFetchUpdatedPost(c)
		return
	}

	log.Printf("[SUCCESS] UpdatePost: Updated post ID '%s', title: '%s'", id, updatedPost.Title)
	c.JSON(http.StatusOK, updatedPost)
}

// DeletePost deletes a blog post by ID
func DeletePost(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[INFO] DeletePost: Received request to delete post ID '%s' from %s", id, c.ClientIP())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("[ERROR] DeletePost: Invalid post ID format '%s'", id)
		apierrors.RespondInvalidPostID(c)
		return
	}

	collection := database.Database.Collection("posts")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		log.Printf("[ERROR] DeletePost: Failed to delete post ID '%s' - %s", id, err.Error())
		apierrors.RespondFailedToDeletePost(c)
		return
	}

	if result.DeletedCount == 0 {
		log.Printf("[ERROR] DeletePost: Post not found for ID '%s'", id)
		apierrors.RespondPostNotFound(c)
		return
	}

	log.Printf("[SUCCESS] DeletePost: Successfully deleted post ID '%s'", id)
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// LikePost increments the like count for a blog post
func LikePost(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[INFO] LikePost: Received request to like post ID '%s' from %s", id, c.ClientIP())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("[ERROR] LikePost: Invalid post ID format '%s'", id)
		apierrors.RespondInvalidPostID(c)
		return
	}

	clientIP := c.ClientIP()

	// Check if post exists
	postsCollection := database.Database.Collection("posts")
	var post models.Post
	err = postsCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			apierrors.RespondPostNotFound(c)
			return
		}
		apierrors.RespondFailedToFetchPost(c)
		return
	}

	// Check if already liked by this IP
	likesCollection := database.Database.Collection("post_likes")
	existingLike := likesCollection.FindOne(context.Background(), bson.M{
		"post_id":    objectID,
		"ip_address": clientIP,
	})

	if existingLike.Err() == nil {
		apierrors.RespondPostAlreadyLiked(c)
		return
	}

	// Create like record
	like := models.PostLike{
		PostID:    objectID,
		IPAddress: clientIP,
		LikedAt:   time.Now(),
	}

	_, err = likesCollection.InsertOne(context.Background(), like)
	if err != nil {
		log.Printf("[ERROR] LikePost: Failed to record like for post ID '%s' - %s", id, err.Error())
		apierrors.RespondFailedToRecordLike(c)
		return
	}

	// Increment likes count in post
	_, err = postsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$inc": bson.M{"likes": 1}},
	)
	if err != nil {
		log.Printf("[ERROR] LikePost: Failed to update like count for post ID '%s' - %s", id, err.Error())
		apierrors.RespondFailedToUpdateLikeCount(c)
		return
	}

	log.Printf("[SUCCESS] LikePost: Successfully liked post ID '%s' from IP %s", id, clientIP)
	c.JSON(http.StatusOK, gin.H{"message": "Post liked successfully"})
}

// ViewPost increments the view count for a blog post
func ViewPost(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[INFO] ViewPost: Received request to view post ID '%s' from %s", id, c.ClientIP())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("[ERROR] ViewPost: Invalid post ID format '%s'", id)
		apierrors.RespondInvalidPostID(c)
		return
	}

	// Check if post exists
	collection := database.Database.Collection("posts")
	var post models.Post
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[ERROR] ViewPost: Post not found for ID '%s'", id)
			apierrors.RespondPostNotFound(c)
			return
		}
		log.Printf("[ERROR] ViewPost: Failed to fetch post ID '%s' - %s", id, err.Error())
		apierrors.RespondFailedToFetchPost(c)
		return
	}

	// Increment view count
	go incrementPostViews(objectID, c.ClientIP(), c.GetHeader("User-Agent"))

	log.Printf("[SUCCESS] ViewPost: Successfully recorded view for post ID '%s' from IP %s", id, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Post view recorded successfully"})
}

// incrementPostViews tracks post views (called asynchronously)
func incrementPostViews(postID primitive.ObjectID, ipAddress, userAgent string) {
	// Create view record
	view := models.PostView{
		PostID:    postID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ViewedAt:  time.Now(),
	}

	viewsCollection := database.Database.Collection("post_views")
	_, err := viewsCollection.InsertOne(context.Background(), view)
	if err != nil {
		log.Printf("[ERROR] incrementPostViews: Failed to record view for post %s - %s", postID.Hex(), err.Error())
		return
	}

	// Increment view count in post
	postsCollection := database.Database.Collection("posts")
	_, err = postsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": postID},
		bson.M{"$inc": bson.M{"views": 1}},
	)
	if err != nil {
		log.Printf("[ERROR] incrementPostViews: Failed to increment view count for post %s - %s", postID.Hex(), err.Error())
		return
	}
}

// generateSlug creates a URL-friendly slug from a title
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, "!", "")
	slug = strings.ReplaceAll(slug, "?", "")
	return slug
}
