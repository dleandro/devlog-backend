package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents a blog post in the database
type Post struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title" binding:"required,min=1,max=200"`
	Content   string             `json:"content" bson:"content" binding:"required,min=1,max=50000"`
	Slug      string             `json:"slug" bson:"slug" binding:"required,min=1,max=100"`
	Summary   string             `json:"summary" bson:"summary,omitempty" binding:"max=500"`
	Tags      []string           `json:"tags" bson:"tags,omitempty" binding:"max=10,dive,min=1,max=50,alphanum"` // Max 10 tags, each alphanumeric
	Published bool               `json:"published" bson:"published"`
	Views     int64              `json:"views" bson:"views"`
	Likes     int64              `json:"likes" bson:"likes"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// PostView represents a view record for analytics
type PostView struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PostID    primitive.ObjectID `json:"post_id" bson:"post_id"`
	IPAddress string             `json:"ip_address" bson:"ip_address,omitempty"`
	UserAgent string             `json:"user_agent" bson:"user_agent,omitempty"`
	ViewedAt  time.Time          `json:"viewed_at" bson:"viewed_at"`
}

// PostLike represents a like on a blog post
type PostLike struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PostID    primitive.ObjectID `json:"post_id" bson:"post_id"`
	IPAddress string             `json:"ip_address" bson:"ip_address,omitempty"`
	LikedAt   time.Time          `json:"liked_at" bson:"liked_at"`
}
