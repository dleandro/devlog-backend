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

# Run all tests locally (requires DB and API server to be running)
test-all-local:
	@echo "Running all tests locally (unit, integration, and E2E)..."
	@echo "Prerequisites: MongoDB and API server must be running"
	@echo ""
	@echo "=== Running Unit and Integration Tests ==="
	@UNIT_DIRS=$$(find . -name "*_test.go" -not -name "*e2e*" -exec dirname {} \; | sort -u | grep -E '\./.+'); \
	if [ -n "$$UNIT_DIRS" ]; then \
		go test -v $$UNIT_DIRS; \
	else \
		echo "No unit/integration tests found"; \
	fi
	@echo ""
	@echo "=== Running E2E Tests ==="
	@E2E_DIRS=$$(find . -name "*e2e*test.go" -exec dirname {} \; | sort -u); \
	if [ -n "$$E2E_DIRS" ]; then \
		go test -v $$E2E_DIRS; \
	else \
		echo "No E2E tests found"; \
	fi
	@echo ""
	@echo "All tests completed!"

# Run E2E tests with Docker (starts all services and runs tests inside Docker)
test-e2e-docker:
	@echo "Running E2E tests with Docker..."
	@echo "This will start all services and run E2E tests inside Docker containers"
	@docker compose --profile test up e2e-tests --build --abort-on-container-exit
	
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

# Check code formatting (for CI)
fmt-check:
	@echo "Checking code formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ The following files are not formatted:"; \
		gofmt -s -l .; \
		echo "Please run 'make fmt' to fix formatting issues."; \
		exit 1; \
	else \
		echo "✅ Code formatting is correct"; \
	fi

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Run CI pipeline locally (simulate GitHub Actions)
ci-local:
	@echo "Running CI pipeline locally..."
	@echo "=== Step 1: Code Formatting Check ==="
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ Code formatting issues found:"; \
		gofmt -s -l .; \
		echo "Run 'make fmt' to fix formatting."; \
		exit 1; \
	else \
		echo "✅ Code formatting is correct"; \
	fi
	@echo ""
	@echo "=== Step 2: Linting ==="
	@golangci-lint run --timeout=5m && echo "✅ Linting passed" || (echo "❌ Linting failed"; exit 1)
	@echo ""
	@echo "=== Step 3: Build ==="
	@go build -v -o dbl-blog-backend ./main.go && echo "✅ Build successful" || (echo "❌ Build failed"; exit 1)
	@echo ""
	@echo "=== Step 4: Unit/Integration Tests ==="
	@$(MAKE) test && echo "✅ Unit/Integration tests passed" || (echo "❌ Unit/Integration tests failed"; exit 1)
	@echo ""
	@echo "🎉 Local CI pipeline completed successfully!"
	@echo "Note: E2E tests require running 'make test-all-local' or 'make test-e2e-docker'"

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
	@echo "  test-all-local   - Run all tests locally (requires DB and server running)"
	@echo "  test-e2e-docker  - Run E2E tests with Docker (full isolation)"
	@echo "  test-e2e-docker-dev - Run E2E tests with Docker (development mode)"
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
	@echo "  fmt-check        - Check code formatting (for CI)"
	@echo "  lint             - Lint code"
	@echo "  ci-local         - Run CI pipeline locally (format, lint, build, test)"
	@echo "  env              - Create .env file from example"
	@echo "  setup            - Setup development environment"
	@echo "  dev              - Start development server with hot reload"
	@echo "  install-air      - Install air for hot reload"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker containers (foreground)"
	@echo "  docker-dev       - Start Docker services in background"
	@echo "  docker-debug     - Debug Docker services (show logs and status)"
	@echo "  help             - Show this help"

.PHONY: build run test test-e2e test-all-local test-e2e-docker test-e2e-docker-dev test-all-with-docker test-coverage test-list deps clean mongo-shell mongo-ping mongo-status mongo-collections mongo-drop-db fmt fmt-check lint ci-local env setup dev install-air docker-build docker-run docker-dev docker-debug help
