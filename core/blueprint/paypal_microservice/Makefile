.PHONY: build run test clean docker-build docker-run

# Go parameters
BINARY_NAME=main
GO_FILES=$(shell find . -name '*.go' -type f)

# Build the application
build:
	@echo "Building..."
	@go build -o $(BINARY_NAME) cmd/api/main.go

# Run the application
run: build
	@echo "Running..."
	@./$(BINARY_NAME)

# Run with air for hot reload
dev:
	@echo "Starting development server with hot reload..."
	@air

# Run tests
test:
	@echo "Testing..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf tmp/

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t paypal-microservice:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker-compose up -d

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	@docker-compose down

# View Docker logs
docker-logs:
	@echo "Viewing Docker logs..."
	@docker-compose logs -f

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Database migrations (if needed)
migrate-up:
	@echo "Running database migrations..."
	@psql -h localhost -U paypal_user -d paypal_db -f schema.sql

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run with hot reload (air)"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with docker-compose"
	@echo "  docker-stop    - Stop docker-compose services"
	@echo "  docker-logs    - View docker logs"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  deps           - Install dependencies"
	@echo "  migrate-up     - Run database migrations"
	@echo "  help           - Show this help message"
