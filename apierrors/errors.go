package apierrors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError represents a structured API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ErrorResponse represents the full error response structure
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// Common error codes
const (
	// Client errors (4xx)
	CodeBadRequest       = "BAD_REQUEST"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeValidationFailed = "VALIDATION_FAILED"
	
	// Server errors (5xx)
	CodeInternalError    = "INTERNAL_ERROR"
	CodeDatabaseError    = "DATABASE_ERROR"
)

// Predefined API errors for posts
var (
	// Post-related errors
	ErrInvalidPostID = APIError{
		Code:    CodeBadRequest,
		Message: "Invalid post ID format",
		Details: "The provided post ID is not a valid MongoDB ObjectID",
	}
	
	ErrPostNotFound = APIError{
		Code:    CodeNotFound,
		Message: "Post not found",
		Details: "The requested post does not exist or has been deleted",
	}
	
	ErrPostAlreadyExists = APIError{
		Code:    CodeConflict,
		Message: "Post with this slug already exists",
		Details: "Please choose a different slug for your post",
	}
	
	ErrPostAlreadyLiked = APIError{
		Code:    CodeConflict,
		Message: "Post already liked",
		Details: "You have already liked this post from this IP address",
	}
	
	// Database operation errors
	ErrFailedToCreatePost = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to create post",
		Details: "An error occurred while saving the post to the database",
	}
	
	ErrFailedToFetchPost = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to fetch post",
		Details: "An error occurred while retrieving the post from the database",
	}
	
	ErrFailedToFetchPosts = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to fetch posts",
		Details: "An error occurred while retrieving posts from the database",
	}
	
	ErrFailedToUpdatePost = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to update post",
		Details: "An error occurred while updating the post in the database",
	}
	
	ErrFailedToDeletePost = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to delete post",
		Details: "An error occurred while deleting the post from the database",
	}
	
	ErrFailedToCountPosts = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to count posts",
		Details: "An error occurred while counting posts in the database",
	}
	
	ErrFailedToDecodePosts = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to decode posts",
		Details: "An error occurred while processing posts data from the database",
	}
	
	ErrFailedToRecordLike = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to record like",
		Details: "An error occurred while saving the like to the database",
	}
	
	ErrFailedToUpdateLikeCount = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to update like count",
		Details: "An error occurred while updating the post's like count",
	}
	
	ErrFailedToFetchUpdatedPost = APIError{
		Code:    CodeDatabaseError,
		Message: "Failed to fetch updated post",
		Details: "The post was updated but could not be retrieved for response",
	}
)

// Helper functions to send structured error responses

// RespondWithError sends a structured error response
func RespondWithError(c *gin.Context, statusCode int, apiError APIError) {
	response := ErrorResponse{
		Error: apiError,
	}
	c.JSON(statusCode, response)
}

// RespondWithCustomError sends a custom error response
func RespondWithCustomError(c *gin.Context, statusCode int, code, message, details string) {
	apiError := APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
	RespondWithError(c, statusCode, apiError)
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(c *gin.Context, details string) {
	apiError := APIError{
		Code:    CodeValidationFailed,
		Message: "Request validation failed",
		Details: details,
	}
	RespondWithError(c, http.StatusBadRequest, apiError)
}

// Common error response helpers
func RespondInvalidPostID(c *gin.Context) {
	RespondWithError(c, http.StatusBadRequest, ErrInvalidPostID)
}

func RespondPostNotFound(c *gin.Context) {
	RespondWithError(c, http.StatusNotFound, ErrPostNotFound)
}

func RespondPostAlreadyExists(c *gin.Context) {
	RespondWithError(c, http.StatusConflict, ErrPostAlreadyExists)
}

func RespondPostAlreadyLiked(c *gin.Context) {
	RespondWithError(c, http.StatusConflict, ErrPostAlreadyLiked)
}

func RespondFailedToCreatePost(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToCreatePost)
}

func RespondFailedToFetchPost(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToFetchPost)
}

func RespondFailedToFetchPosts(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToFetchPosts)
}

func RespondFailedToUpdatePost(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToUpdatePost)
}

func RespondFailedToDeletePost(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToDeletePost)
}

func RespondFailedToCountPosts(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToCountPosts)
}

func RespondFailedToDecodePosts(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToDecodePosts)
}

func RespondFailedToRecordLike(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToRecordLike)
}

func RespondFailedToUpdateLikeCount(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToUpdateLikeCount)
}

func RespondFailedToFetchUpdatedPost(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, ErrFailedToFetchUpdatedPost)
}