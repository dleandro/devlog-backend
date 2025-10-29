# DBL Blog Backend

A RESTful API backend for a blog application built with Go, Gin, and MongoDB.

## Features

- **CRUD Operations** for blog posts
- **Analytics Tracking** (views, likes)
- **MongoDB Integration** with official Go driver
- **Structured Error Handling** with consistent API responses
- **RESTful API Design** with proper HTTP status codes
- **CORS Support** with configurable origins
- **Environment Configuration** for development and production
- **Request/Response Logging** for debugging and monitoring
- **Docker Support** for easy deployment
- **Simple & Fast** - No complex migrations needed

## Project Structure

```
dbl-blog-backend/
├── apierrors/           # Structured error handling and API responses
├── database/            # MongoDB connection and configuration
├── handlers/            # HTTP handlers for API endpoints
├── middleware/          # Custom middleware (CORS, etc.)
├── models/              # MongoDB models and data structures
├── routes/              # Route definitions and setup
├── main.go              # Application entry point
├── go.mod               # Go module dependencies
├── .env.example         # Environment variables template
├── docker-compose.yml   # Docker development setup
├── Dockerfile          # Container configuration
├── Makefile            # Build and development commands
└── README.md           # This file
```

## API Endpoints

### Blog Posts

- `POST /api/v1/posts` - Create a new post
- `GET /api/v1/posts` - Get all posts (with pagination and filtering)
- `GET /api/v1/posts/:id` - Get a specific post by ID or slug
- `PUT /api/v1/posts/:id` - Update a post
- `DELETE /api/v1/posts/:id` - Delete a post
- `POST /api/v1/posts/:id/like` - Like a post
- `POST /api/v1/posts/:id/view` - Track post view (manual)

### Health Check

- `GET /health` - Health check endpoint

## Error Handling

The API uses a structured error handling system with consistent responses:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE"
}
```

**Common Error Codes:**

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

- Go 1.21 or higher
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
# Edit .env with your database credentials
```

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

| Variable      | Description               | Default                   |
| ------------- | ------------------------- | ------------------------- |
| `MONGODB_URI` | MongoDB connection string | mongodb://localhost:27017 |
| `DB_NAME`     | Database name             | dbl_blog                  |
| `PORT`        | Server port               | 8080                      |
| `GIN_MODE`    | Gin mode (debug/release)  | debug                     |
| `ENVIRONMENT` | Application environment   | development               |

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

## Usage Examples

### Create a Post

```bash
curl -X POST http://localhost:8080/api/v1/posts \
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
curl -X POST http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439011/like
```

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
make test         # Run tests
make mongo-shell  # Open MongoDB shell
make clean        # Clean build artifacts
make setup        # Setup development environment
```

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

### Option 1: Traditional Deployment

1. Set environment variables for production (especially `MONGODB_URI`)
2. Build the application: `make build`
3. Start the server: `./dbl-blog-backend`

### Option 2: Docker Deployment

1. Set environment variables in docker-compose.yml
2. Deploy: `docker-compose up -d`
