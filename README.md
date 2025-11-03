# DBL Blog Backend

A RESTful API backend for a blog application built with Go, Gin, and MongoDB.

## Features

- **CRUD Operations** for blog posts
- **Analytics Tracking** (views, likes)
- **MongoDB Integration** with official Go driver
- **Structured Error Handling** with consistent API responses
- **RESTful API Design** with proper HTTP status codes and idempotent operations
- **CORS Support** with configurable origins
- **Environment Configuration** for development and production
- **Request/Response Logging** for debugging and monitoring
- **Docker Support** for easy deployment
- **API Key Authentication** with Bearer token support
- **Rate Limiting** protection against abuse
- **Input Sanitization** preventing NoSQL injection attacks
- **Security Middleware** with timing attack resistance
- **Comprehensive Testing** with E2E and unit tests
- **Simple & Fast** - No complex migrations needed

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
- `POST /api/v1/posts/:id/view` - Track post view

### Protected Endpoints (Admin API Key Required)

- `POST /api/v1/posts` - Create a new post
- `PUT /api/v1/posts/:id` - Update a post
- `DELETE /api/v1/posts/:id` - Delete a post

**Authentication Header Required:**

```bash
Authorization: Bearer <your-admin-api-key>
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

Include the API key in the Authorization header:

```bash
# Create a new post
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer your-admin-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Blog Post",
    "content": "Post content...",
    "slug": "new-blog-post",
    "published": true
  }'

# Update a post
curl -X PUT http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer your-admin-api-key" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title"}'

# Delete a post
curl -X DELETE http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer your-admin-api-key"
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

**What AIR does:**

- 🔄 **Automatically rebuilds** your app when you save files
- 🚀 **Restarts the server** instantly with changes
- 📁 **Watches** Go files, templates, and config files
- ⚡ **Fast development** cycle - no manual restarts needed

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

### Security Configuration

The application automatically configures security features with flexible rate limiting:

**Rate Limiting Configuration:**

- **Admin Operations**: Configurable per-minute limits (default: 30/minute)
- **Public GET Requests**: Configurable browsing limits (default: 120/minute)
- **Social Interactions**: Configurable like/view limits (default: 60/minute)
- **General Public**: Configurable default limits (default: 100/minute)

**Security Features:**

- **Multiple API Key Support**: Comma-separated keys for rotation
- **Input Sanitization**: Enabled by default for all requests
- **Timing Attack Resistance**: Constant-time API key comparison
- **CORS**: Configurable origins for cross-origin requests

### Public Endpoint Protection

While hosting providers like Vercel provide infrastructure-level DDoS protection, you can enable additional application-level rate limiting:

```bash
# In your .env file
ENABLE_PUBLIC_RATE_LIMIT=true
```

**When to Enable Public Rate Limiting:**

- ✅ High-traffic production deployments
- ✅ Cost-sensitive MongoDB usage
- ✅ Extra protection against application-level abuse
- ❌ Not needed for most small-medium blogs (Vercel's protection is sufficient)

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

This layered approach ensures security checks happen in the right order while maintaining performance.

### Security Features

**NoSQL Injection Protection:**

```go
// Automatically blocks requests containing:
- MongoDB operators: $where, $ne, $gt, etc.
- JavaScript injection attempts
- Malicious regex patterns
- Invalid BSON structures
```

**Rate Limiting:**

```go
// Configurable per-IP rate limiting with:
- Admin Operations: Environment configurable (default: 30/minute)
- Public GET: Environment configurable (default: 120/minute)
- Social Actions: Environment configurable (default: 60/minute)
- Sliding window algorithm with automatic cleanup
- Thread-safe implementation with mutex protection
- Optional public endpoint protection
```

**API Key Security:**

```go
// Secure API key handling with:
- Constant-time comparison (timing attack resistance)
- Bearer token format validation
- Environment-based configuration
- Comprehensive error logging
```

## Usage Examples

### Create a Post

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer your-admin-api-key" \
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

### Track a Post View

```bash
curl -X POST http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011/view
```

## Development

### Available Make Commands

```bash
make dev          # Run with AIR hot reload (recommended)
make run          # Run the application normally
make build        # Build the application
make test         # Run all tests (unit + E2E)
make test-docker  # Run tests with Docker MongoDB
make test-coverage # Run tests with coverage report
make fmt          # Format code
make lint         # Lint code with golangci-lint
make mongo-shell  # Open MongoDB shell
make mongo-ping   # Check MongoDB connection
make clean        # Clean build artifacts
make setup        # Setup development environment
make docker-build # Build Docker image
make docker-run   # Run with docker-compose
make install-air  # Install AIR for hot reload
```

### Testing

The project includes comprehensive test coverage:

**End-to-End (E2E) Tests:**

- Authentication security tests
- Protected endpoint access control
- Public endpoint accessibility
- Malformed request handling
- Timing attack resistance validation

**Unit Tests:**

- Rate limiting functionality
- Middleware components
- Handler logic
- Helper functions

**Run Tests:**

```bash
# Run all tests
make test

# Run with verbose output
go test ./... -v

# Run only E2E tests
go test ./handlers/ -run TestE2E -v

# Run only rate limiting tests
go test ./middleware/ -run TestRateLimiting -v

# Run tests with Docker MongoDB
make test-docker

# Run with coverage
go test ./... -cover
```

**Test Coverage:**

- **End-to-end testing**: ✅ 10 comprehensive E2E test cases (`e2e_test.go`)
- **Rate limiting**: ✅ 11 unit test cases (`middleware/rate_limit_test.go`)
- **Complete CRUD operations**: ✅ E2E tests for all endpoints with authentication (`e2e_test.go`)
- **Input validation**: ✅ Multiple edge cases across all tests
- **Environment configuration**: ✅ Helper function tests

### Development Workflow

1. **Start development server with hot reload:**

   ```bash
   make dev
   ```

   This uses [AIR](https://github.com/cosmtrek/air) to automatically rebuild and restart the server when you make changes to Go files.

2. **Make changes to your code** - AIR will detect changes and reload automatically

3. **Test your changes** using curl or your frontend application

4. **View logs** - All requests and responses are logged with timestamps

**AIR Benefits:**

- ✅ **Instant feedback** - See changes immediately
- ✅ **Automatic rebuilds** - No manual `ctrl+c` and restart
- ✅ **Environment variables** - Loads from `.env` automatically
- ✅ **Error recovery** - Continues watching even if build fails

### Docker Support

```bash
# Run with Docker Compose (includes MongoDB)
docker-compose up

# Build Docker image
make docker-build
```

## Deployment

### Hosting Provider Protection

**Vercel:**

- ✅ Global CDN with DDoS protection
- ✅ Automatic scaling and load balancing
- ✅ Built-in rate limiting (plan-dependent)
- ✅ Bot detection and malicious traffic filtering
- 💡 **Recommendation**: Vercel's protection is sufficient for most blogs

**Railway/Render:**

- ✅ Infrastructure-level DDoS protection
- ✅ Automatic scaling capabilities
- 💡 **Recommendation**: Consider enabling public rate limiting for extra protection

**When to Enable Additional Rate Limiting:**

```bash
# Enable for high-traffic or cost-sensitive deployments
ENABLE_PUBLIC_RATE_LIMIT=true
```

### Option 1: Vercel Deployment (Recommended)

1. Deploy to Vercel with automatic scaling
2. Set environment variables in Vercel dashboard
3. Vercel handles DDoS protection automatically
4. No additional rate limiting needed for most use cases

### Option 2: Traditional Deployment

1. Set environment variables for production (especially `MONGODB_URI`)
2. Build the application: `make build`
3. Start the server: `./dbl-blog-backend`
4. Consider enabling public rate limiting: `ENABLE_PUBLIC_RATE_LIMIT=true`

### Option 3: Docker Deployment

1. Set environment variables in docker-compose.yml
2. Deploy: `docker-compose up -d`

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
curl -H "Authorization: Bearer your-api-key" http://localhost:8080/api/v1/posts
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

**Tests Failing?**

```bash
# Run specific test suites
go test ./handlers/ -v -run TestE2E
go test ./middleware/ -v -run TestRateLimiting

# Check MongoDB test database access
# E2E tests use separate test database
```

**Build Issues?**

```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

## Frontend Integration

For frontend developers wanting to integrate with this API, see the comprehensive integration guide:

📖 **[Frontend Integration Guide](./FRONTEND_INTEGRATION_PROMPT.md)**

This guide includes:

- Complete API documentation with TypeScript interfaces
- Ready-to-use code examples for popular frameworks
- Authentication setup instructions
- Error handling patterns
- Real-world integration examples

**Quick Start for Frontend:**

```javascript
// Example API call
const response = await fetch("http://localhost:8080/api/v1/posts");
const data = await response.json();
console.log(data.posts); // Array of blog posts
```

**Framework Support:**

- Svelte/SvelteKit
- React/Next.js
- Vue/Nuxt.js
- Vanilla JavaScript
- Any framework that supports HTTP requests

## Architecture Summary

This blog backend implements a modern, secure, and scalable REST API with the following key architectural decisions:

### 🏗️ **Modular Design**

- **Separation of Concerns**: Clear separation between handlers, middleware, models, and routes
- **Consolidated Rate Limiting**: Single module handling both admin and public rate limiting
- **Structured Error Handling**: Consistent API error responses across all endpoints

### 🔒 **Security-First Approach**

- **Layered Security**: Multiple middleware layers (CORS → Sanitization → Rate Limiting → Auth)
- **Configurable Protection**: Environment-driven rate limits and security settings
- **Production-Ready**: Timing attack resistance, input validation, and comprehensive logging

### ⚡ **Performance & Scalability**

- **Efficient Rate Limiting**: Thread-safe implementation with automatic cleanup
- **Minimal Dependencies**: Lean dependency tree for better security and performance
- **Cloud-Ready**: Docker support and hosting provider compatibility

### 🧪 **Quality Assurance**

- **Comprehensive Testing**: E2E security tests, unit tests, and integration tests
- **Development Experience**: Hot reload, linting, and automated testing
- **Documentation**: Complete API documentation and frontend integration guides

This architecture provides a solid foundation for a production blog API while remaining maintainable and extensible.
