# DBL Blog Backend

A RESTful API backend for a blog application built with Go, Gin, and MongoDB.

## Project Structure

```
dbl-blog-backend/
├── apierrors/           # Structured error handling and API responses
│   └── errors.go        # Error definitions and response helpers
├── database/            # MongoDB connection and configuration
│   └── connection.go    # Database connection setup
├── handlers/            # HTTP handlers for API endpoints
│   └── post.go          # Post CRUD handlers
├── middleware/          # Custom middleware (CORS, auth, security, rate limiting)
│   ├── auth.go          # Admin API key authentication with security features
│   ├── cors.go          # CORS configuration
│   ├── input_sanitizer.go # NoSQL injection protection
│   ├── rate_limit.go    # Consolidated rate limiting for admin and public endpoints
│   └── rate_limit_test.go # Rate limiting unit tests
├── models/              # MongoDB models and data structures
│   └── post.go          # Post model with validation constraints
├── routes/              # Route definitions and setup
│   └── routes.go        # API route configuration with middleware stack
├── scripts/             # Utility scripts (API key generation, etc.)
│   └── generate-api-key.sh # Secure API key generation script
├── main.go              # Application entry point
├── go.mod               # Go module dependencies
├── go.sum               # Go module checksums
├── .env.example         # Environment variables template with rate limiting config
├── .gitignore           # Git ignore patterns
├── docker-compose.yml   # Docker development setup
├── Dockerfile          # Container configuration
├── Makefile            # Build and development commands
├── FRONTEND_INTEGRATION_PROMPT.md # Frontend integration guide
└── README.md           # This documentation file
```

## API Endpoints

### Public Endpoints (No Authentication Required)

- `GET /api/v1/posts` - Get all posts (with pagination and filtering)
- `GET /api/v1/posts/:id` - Get a specific post by ID or slug
- `PUT /api/v1/posts/:id/like` - Like a post
- `PUT /api/v1/posts/:id/dislike` - Dislike a post (decrement likes)
- `PUT /api/v1/posts/:id/view` - Track post view

### Protected Endpoints (Admin API Key Required)

- `POST /api/v1/posts` - Create a new post
- `PUT /api/v1/posts/:id` - Update a post
- `DELETE /api/v1/posts/:id` - Delete a post

**Authentication Header Required:**

```bash
X-API-Key: <your-admin-api-key>
```

### Health Check

- `GET /health` - Health check endpoint

## Authentication

The API uses API key-based authentication to protect admin operations (create, update, delete posts).

### Setup Admin API Key

1. **Generate a secure API key:**

   ```bash
   # Option A: Use the provided script
   ./scripts/generate-api-key.sh

   # Option B: Generate manually
   openssl rand -hex 32
   ```

2. **Add to environment variables:**

   ```bash
   # In your .env file
   # Multiple API keys support key rotation and multiple admin users
   ADMIN_API_KEYS=key1-here,key2-here,key3-here
   ```

### Using Protected Endpoints

Include the API key in the X-API-Key header:

```bash
# Create a new post
curl -X POST http://localhost:8080/api/v1/posts \
  -H "X-API-Key: your-admin-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Blog Post",
    "content": "Post content...",
    "slug": "new-blog-post",
    "published": true
  }'

# Update a post
curl -X PUT http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011 \
  -H "X-API-Key: your-admin-api-key" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title"}'

# Delete a post
curl -X DELETE http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011 \
  -H "X-API-Key: your-admin-api-key"
```

### Authentication Errors

| HTTP Status | Error Code                | Description                                   |
| ----------- | ------------------------- | --------------------------------------------- |
| 401         | `UNAUTHORIZED`            | Missing, invalid format, or incorrect API key |
| 429         | `RATE_LIMITED`            | Too many requests - rate limit exceeded       |
| 500         | `SERVER_MISCONFIGURATION` | Admin API key not configured on server        |

### Security Features

The API implements multiple layers of security:

- **API Key Authentication**: Bearer token-based authentication for admin operations with multiple key support
- **Configurable Rate Limiting**: Separate rate limits for admin and public endpoints (configurable via environment variables)
- **Input Sanitization**: Automatic detection and blocking of NoSQL injection attempts
- **Timing Attack Resistance**: Constant-time string comparison for API keys
- **Request Validation**: Strict input validation with comprehensive error messages

## Error Handling

The API uses a structured error handling system with consistent responses:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE"
}
```

**Common Error Codes:**

**Authentication Errors:**

- `UNAUTHORIZED` - Missing, invalid format, or incorrect API key
- `SERVER_MISCONFIGURATION` - Admin API key not configured on server

**Post Errors:**

- `VALIDATION_ERROR` - Invalid input data
- `POST_NOT_FOUND` - Requested post doesn't exist
- `POST_ALREADY_EXISTS` - Duplicate slug detected
- `INVALID_POST_ID` - Invalid MongoDB ObjectID format
- `FAILED_TO_CREATE_POST` - Database insertion failed
- `FAILED_TO_UPDATE_POST` - Database update failed
- `FAILED_TO_DELETE_POST` - Database deletion failed

All endpoints return appropriate HTTP status codes (200, 201, 400, 404, 500) along with structured error messages.

## Setup Instructions

### Prerequisites

- Go 1.25.1 or higher
- MongoDB 7.0 or higher (or MongoDB Atlas free tier)
- Git
- [AIR](https://github.com/cosmtrek/air) for hot reload (optional but recommended for development)

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd dbl-blog-backend
```

### 2. Environment Configuration

```bash
cp .env.example .env
# Edit .env with your database credentials and security settings
```

**Key Configuration Items:**

- **Database**: MongoDB connection settings
- **Authentication**: Admin API keys for protected endpoints
- **Rate Limiting**: Configurable limits for admin and public endpoints (optional)
- **CORS**: Allowed origins for cross-origin requests

### 3. Database Setup

**Option A: Local MongoDB**

```bash
# Install MongoDB locally
brew install mongodb/brew/mongodb-community
# Start MongoDB
brew services start mongodb/brew/mongodb-community
```

**Option B: MongoDB Atlas (Free Cloud)**

1. Create account at [MongoDB Atlas](https://cloud.mongodb.com)
2. Create a free cluster
3. Get connection string and update `.env`

### 4. Install Dependencies

```bash
go mod tidy
```

### 5. Run the Application

**Option A: Standard Go Run**

```bash
# Development mode
go run main.go

# Or using the Makefile
make run
```

**Option B: Hot Reload with AIR (Recommended for Development)**

[AIR](https://github.com/cosmtrek/air) provides live reload functionality for Go applications.

Install AIR:

```bash
# Install globally
go install github.com/cosmtrek/air@latest

# Or install to project (already included in go.mod)
go mod tidy
```

Run with hot reload:

```bash
# Using AIR directly
air

# Or using the Makefile
make dev
```

**AIR Configuration:**
The project includes `.air.toml` configuration file that:

- Watches `*.go`, `*.html`, `*.yaml`, `*.yml`, `*.toml` files
- Excludes vendor, tmp, and test files
- Builds to `./tmp/main` (automatically cleaned up)
- Runs with environment variables from `.env`

The server will start on `http://localhost:8080`

## Environment Variables

| Variable                               | Description                                     | Default          | Required |
| -------------------------------------- | ----------------------------------------------- | ---------------- | -------- |
| `MONGODB_USERNAME`                     | MongoDB username                                | admin            | Yes      |
| `MONGODB_PASSWORD`                     | MongoDB password                                | password         | Yes      |
| `MONGODB_HOST`                         | MongoDB host                                    | localhost        | Yes      |
| `MONGODB_PORT`                         | MongoDB port                                    | 27017            | Yes      |
| `MONGODB_AUTH_SOURCE`                  | MongoDB auth source                             | admin            | Yes      |
| `DB_NAME`                              | Database name                                   | dbl_blog         | Yes      |
| `PORT`                                 | Server port                                     | 8080             | No       |
| `GIN_MODE`                             | Gin mode (debug/release)                        | debug            | No       |
| `ENVIRONMENT`                          | Application environment                         | development      | No       |
| `ADMIN_API_KEYS`                       | API keys for admin operations (comma-separated) | (none)           | **Yes**  |
| `ALLOWED_ORIGINS`                      | CORS allowed origins                            | \* (development) | No       |
| `TEST_MONGODB_URI`                     | MongoDB URI for E2E tests                       | (auto-generated) | No       |
| `ENABLE_PUBLIC_RATE_LIMIT`             | Enable public endpoint rate limiting            | false            | No       |
| `ADMIN_RATE_LIMIT_PER_MINUTE`          | Admin operations rate limit                     | 30               | No       |
| `PUBLIC_GET_RATE_LIMIT_PER_MINUTE`     | Public GET requests rate limit                  | 120              | No       |
| `PUBLIC_SOCIAL_RATE_LIMIT_PER_MINUTE`  | Public social interactions rate limit           | 60               | No       |
| `PUBLIC_DEFAULT_RATE_LIMIT_PER_MINUTE` | Public default rate limit                       | 100              | No       |

## MongoDB Collections

### Posts Collection

```json
{
  "_id": "ObjectId",
  "title": "Post title",
  "content": "Post content (HTML/Markdown)",
  "slug": "url-friendly-slug",
  "summary": "Short description",
  "tags": ["array", "of", "tags"],
  "published": true,
  "views": 0,
  "likes": 0,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Post Model Validation

The Post model includes comprehensive validation constraints:

```go
type Post struct {
    Title     string   `binding:"required,min=1,max=200"`
    Content   string   `binding:"required,min=1,max=50000"`
    Slug      string   `binding:"required,min=1,max=100"`
    Summary   string   `binding:"max=500"`
    Tags      []string `binding:"max=10,dive,min=1,max=50,alphanum"`
    Published bool
    Views     int64
    Likes     int64
}
```

**Validation Rules:**

- **Title**: Required, 1-200 characters
- **Content**: Required, 1-50,000 characters
- **Slug**: Required, 1-100 characters (allows letters, numbers, dashes, and underscores)
- **Summary**: Optional, max 500 characters
- **Tags**: Max 10 tags, each 1-50 alphanumeric characters
- **Analytics**: Auto-managed (views, likes, timestamps)

### Analytics Collections

**post_views** - Track individual page views

```json
{
  "_id": "ObjectId",
  "post_id": "ObjectId",
  "ip_address": "192.168.1.1",
  "user_agent": "Browser info",
  "viewed_at": "2024-01-01T00:00:00Z"
}
```

**post_likes** - Track individual likes (prevent duplicates)

```json
{
  "_id": "ObjectId",
  "post_id": "ObjectId",
  "ip_address": "192.168.1.1",
  "liked_at": "2024-01-01T00:00:00Z"
}
```

## Security Architecture

### Middleware Stack

The application implements a comprehensive security middleware stack:

1. **CORS Middleware** (`middleware/cors.go`)

   - Configurable allowed origins
   - Proper headers for cross-origin requests
   - Preflight request handling

2. **Input Sanitization** (`middleware/input_sanitizer.go`)

   - NoSQL injection detection and prevention
   - Malicious pattern recognition
   - Request body and parameter validation

3. **Rate Limiting Middleware** (`middleware/rate_limit.go`)

   - **Admin Rate Limiting**: Configurable limits for admin operations (default: 30/minute)
   - **Public Rate Limiting**: Optional tiered limits for different endpoint types
   - Thread-safe implementation with automatic cleanup
   - Environment variable configuration for all limits

4. **Authentication Middleware** (`middleware/auth.go`)
   - API key validation with Bearer token format
   - Timing attack resistance with constant-time comparison
   - Multiple API key support (comma-separated)
   - Comprehensive error responses

### Middleware Execution Order

The middleware stack executes in the following order for different endpoint types:

**Global Middleware (All Endpoints):**

```go
1. CORS Middleware           // Handle cross-origin requests
2. Gin Logger               // Request/response logging
3. Gin Recovery             // Panic recovery
4. Input Sanitization       // NoSQL injection protection
```

**Public Endpoints** (`GET /posts`, `PUT /posts/:id/like`, etc.):

```go
5. [Optional] Public Rate Limiting  // If ENABLE_PUBLIC_RATE_LIMIT=true
6. Handler                          // Execute endpoint logic
```

**Protected Admin Endpoints** (`POST /posts`, `PUT /posts/:id`, `DELETE /posts/:id`):

```go
5. Admin Rate Limiting      // Rate limit admin operations
6. Admin Authentication     // Validate API key
7. Handler                  // Execute endpoint logic
```

## Usage Examples

### Create a Post

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "X-API-Key: your-admin-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Engineering Post",
    "content": "This is the content of my first post about engineering...",
    "summary": "A post about engineering concepts",
    "tags": ["engineering", "backend", "golang"],
    "published": true
  }'
```

### Get All Posts

```bash
curl http://localhost:8080/api/v1/posts?page=1&limit=10&published=true
```

### Like a Post

```bash
# Use actual MongoDB ObjectID from the post creation response
curl -X PUT http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011/like
```

**Note:** Like functionality uses PUT method because it's idempotent - multiple requests from the same IP have the same effect (prevents duplicate likes).

### Dislike a Post

```bash
curl -X PUT http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011/dislike
```

**Note:** Dislike functionality decrements the like count (minimum 0). If the requesting IP has previously liked the post, their like record will be removed.

### Track a Post View

```bash
curl -X PUT http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011/view
```

### Testing

````bash
# Run unit and integration tests
make test

# Run e2e tests with the blog-api and mongodb running on docker
make test-e2e

### Docker Support

```bash
# Run with Docker Compose (includes MongoDB)
make docker-run

# Build Docker image
make docker-build
````

## Troubleshooting

### Common Issues

**1. MongoDB Connection Issues**

```bash
# Check MongoDB is running
brew services list | grep mongodb

# Restart MongoDB
brew services restart mongodb/brew/mongodb-community

# Check connection string format
# mongodb://username:password@host:port/database?authSource=admin
```

**2. Authentication Issues**

```bash
# Verify API keys are set
echo $ADMIN_API_KEYS

# Generate new API key
./scripts/generate-api-key.sh

# Test authentication
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/posts
```

**3. Validation Errors**

```bash
# Slug allows letters, numbers, dashes, and underscores
# Example: "my-post" ❌ should be "mypost" ✅

# Check field length limits:
# Title: max 200 characters
# Content: max 50,000 characters
# Summary: max 500 characters
# Tags: max 10 tags, each max 50 characters
```

**4. Rate Limiting**

```bash
# Admin endpoints limited to 30 requests/minute (one every 2 seconds)
# Public endpoints (GET, like, view) are NOT rate limited
# If you hit rate limits on admin operations:
# - Use different IP addresses for testing
# - Restart server to reset counters (development only)
```

**5. CORS Issues**

```bash
# Update ALLOWED_ORIGINS in .env
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# For development, use:
ALLOWED_ORIGINS=*
```

### Development Tips

**Hot Reload Not Working?**

```bash
# Ensure AIR is installed
go install github.com/cosmtrek/air@latest

# Check .air.toml configuration
# Verify file extensions in cmd/include_ext
```

**Build Issues?**

```bash
# Clean and rebuild
make clean
go mod tidy
make build
```
