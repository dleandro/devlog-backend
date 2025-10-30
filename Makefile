# Go Blog Backend Makefile

# Variables
BINARY_NAME=dbl-blog-backend
MAIN_PATH=./main.go
DB_NAME=dbl_blog

# Default target
.DEFAULT_GOAL := help

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	@echo "Running the application..."
	@go run $(MAIN_PATH)

# Run unit and integration tests (excludes E2E tests)
test:
	@echo "Running unit and integration tests..."
	@go test -v $$(find . -name "*_test.go" -not -name "*e2e*" -exec dirname {} \; | sort -u | grep -E '\./.+')

# Run E2E tests only (requires running API server)
test-e2e:
	@echo "Running E2E tests (requires running API server)..."
	@go test -v $$(find . -name "*e2e*test.go" -exec dirname {} \; | sort -u)

# Run tests with coverage (excludes E2E tests)
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out $$(find . -name "*_test.go" -not -name "*e2e*" -exec dirname {} \; | sort -u | grep -E '\./.+')
	@go tool cover -html=coverage.out -o coverage.html

# need to test the coverage command to see if it works correctly
# need to review the code that the was added
# need to add test pipelines
# need to test on postman or curl quickly
# need to apply this to the frontend and add an admin page where posts can be added and deleted and updated
# need to test that posts are seen as expected
# need to deploy the backend and then the frontend
# need to add the remaining 4 posts

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html

# MongoDB shell access
mongo-shell:
	@echo "Opening MongoDB shell..."
	@mongosh

# Check MongoDB connection
mongo-ping:
	@echo "Pinging MongoDB..."
	@mongosh --eval "db.adminCommand('ping')"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Create .env file from example
env:
	@echo "Creating .env file from .env.example..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"

# Setup development environment
setup: deps env
	@echo "Development environment setup complete!"
	@echo "1. Edit .env file with your MongoDB URI"
	@echo "2. Make sure MongoDB is running (locally or Atlas)"
	@echo "3. Start the server: make run"

# Development mode with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@air

# Install air for hot reload development
install-air:
	@echo "Installing air for hot reload..."
	@go install github.com/cosmtrek/air@latest

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME) .

docker-run:
	@echo "Running Docker containers with docker-compose..."
	@docker-compose up

# Start Docker services in background
docker-dev:
	@echo "Starting Docker services in background..."
	@docker-compose up -d

# MongoDB commands
mongo-status:
	@echo "Checking MongoDB status..."
	@mongosh --eval "db.adminCommand('serverStatus').ok"

mongo-collections:
	@echo "Listing MongoDB collections..."
	@mongosh $(DB_NAME) --eval "db.getCollectionNames()"

mongo-drop-db:
	@echo "Dropping MongoDB database (WARNING: This deletes all data!)..."
	@mongosh $(DB_NAME) --eval "db.dropDatabase()"

# Show which test files will be executed
test-list:
	@echo "Unit/Integration test directories:"
	@find . -name "*_test.go" -not -name "*e2e*" -exec dirname {} \; | sort -u | grep -E '\./.+' || echo "  (none found)"
	@echo ""
	@echo "E2E test directories:"
	@find . -name "*e2e*test.go" -exec dirname {} \; | sort -u || echo "  (none found)"

# Help
help:
	@echo "Available commands:"
	@echo "  build            - Build the application"
	@echo "  run              - Run the application"
	@echo "  test             - Run unit/integration tests (auto-discovers subdirectories)"
	@echo "  test-e2e         - Run E2E tests only (auto-discovers e2e test files)"
	@echo "  test-all-with-docker - Start Docker Compose and run all tests"
	@echo "  test-coverage    - Run tests with coverage (auto-discovers subdirectories)"
	@echo "  test-list        - Show which test directories will be executed"
	@echo "  deps             - Install dependencies"
	@echo "  clean            - Clean build artifacts"
	@echo "  mongo-shell      - Open MongoDB shell"
	@echo "  mongo-ping       - Check MongoDB connection"
	@echo "  mongo-status     - Check MongoDB server status"
	@echo "  mongo-collections- List MongoDB collections"
	@echo "  mongo-drop-db    - Drop MongoDB database (WARNING!)"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint code"
	@echo "  env              - Create .env file from example"
	@echo "  setup            - Setup development environment"
	@echo "  dev              - Start development server with hot reload"
	@echo "  install-air      - Install air for hot reload"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker containers (foreground)"
	@echo "  docker-dev       - Start Docker services in background"
	@echo "  help             - Show this help"

.PHONY: build run test test-e2e test-all-with-docker test-coverage test-list deps clean mongo-shell mongo-ping mongo-status mongo-collections mongo-drop-db fmt lint env setup dev install-air docker-build docker-run docker-dev help
