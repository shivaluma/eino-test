.PHONY: run build test clean fmt vet tidy deps vendor dev air server docker-up docker-down docker-logs db-migrate db-reset help

# Variables
BINARY_NAME=food-agent-server
BUILD_DIR=build
SERVER_DIR=cmd/server
DOCKER_COMPOSE=docker-compose

# Default target
all: fmt vet test build

# Development commands
dev: air

# Run with live reload using Air
air:
	@echo "Starting development server with live reload..."
	@air

# Run the server directly
server:
	@echo "Starting server..."
	@go run $(SERVER_DIR)/main.go

# Run the application (legacy compatibility)
run:
	@go run .

# Build commands
build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SERVER_DIR)/main.go

build-prod:
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) $(SERVER_DIR)/main.go

# Testing commands
test:
	@echo "Running tests..."
	@go test ./...

test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-api:
	@echo "Running API integration tests..."
	@if pgrep -f "cmd/server/main.go" > /dev/null; then \
		./scripts/test-api.sh; \
	else \
		echo "Error: Server is not running. Please start it with 'make dev' or 'make server' first."; \
		exit 1; \
	fi

test-api-url:
	@echo "Running API tests against custom URL..."
	@read -p "Enter API base URL (e.g., http://localhost:8888): " url; \
	./scripts/test-api.sh "$$url"

# Code quality commands
fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

# Dependency management
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

deps:
	@echo "Downloading dependencies..."
	@go mod download

vendor:
	@echo "Creating vendor directory..."
	@go mod vendor

# Docker commands
docker-up:
	@echo "Starting Docker services..."
	@$(DOCKER_COMPOSE) up -d

docker-down:
	@echo "Stopping Docker services..."
	@$(DOCKER_COMPOSE) down

docker-logs:
	@echo "Showing Docker logs..."
	@$(DOCKER_COMPOSE) logs -f

docker-build:
	@echo "Building Docker image..."
	@docker build -t food-agent:latest .

docker-rebuild:
	@echo "Rebuilding Docker services..."
	@$(DOCKER_COMPOSE) down
	@$(DOCKER_COMPOSE) up -d --build

# Database commands
db-migrate:
	@echo "Running database migrations..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | grep -v '^$$' | xargs) && \
		if command -v psql >/dev/null 2>&1; then \
			psql postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME -f migrations/001_initial_schema.sql; \
		else \
			echo "psql not found locally, using Docker..."; \
			docker run --rm -i \
				-v "$(PWD)/migrations:/migrations" \
				postgres:15-alpine \
				psql postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME -f /migrations/001_initial_schema.sql; \
		fi; \
	else \
		echo "Error: .env file not found. Please copy .env.example to .env and configure it."; \
		exit 1; \
	fi

db-reset:
	@echo "Resetting database..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | grep -v '^$$' | xargs) && \
		if command -v psql >/dev/null 2>&1; then \
			psql postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" && \
			$(MAKE) db-migrate; \
		else \
			echo "psql not found locally, using Docker..."; \
			docker run --rm -i \
				postgres:15-alpine \
				psql postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" && \
			$(MAKE) db-migrate; \
		fi; \
	else \
		echo "Error: .env file not found. Please copy .env.example to .env and configure it."; \
		exit 1; \
	fi

db-connect:
	@echo "Connecting to database..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | grep -v '^$$' | xargs) && \
		if command -v psql >/dev/null 2>&1; then \
			psql postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME; \
		else \
			echo "psql not found locally, using Docker..."; \
			docker run --rm -it \
				postgres:15-alpine \
				psql postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME; \
		fi; \
	else \
		echo "Error: .env file not found. Please copy .env.example to .env and configure it."; \
		exit 1; \
	fi

# Cleanup commands
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf tmp

clean-all: clean
	@echo "Cleaning all artifacts including vendor and dependencies..."
	@rm -rf vendor
	@go clean -modcache

# Setup commands
setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env file created from .env.example. Please configure it."; \
	fi
	@$(MAKE) deps
	@echo "Installing Air for live reload..."
	@go install github.com/air-verse/air@latest
	@echo "Setup complete! Run 'make dev' to start development server."

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed!"

# Help command
help:
	@echo "Available commands:"
	@echo "  Development:"
	@echo "    dev              - Start development server with live reload (using Air)"
	@echo "    server           - Run server directly"
	@echo "    air              - Start Air live reload"
	@echo ""
	@echo "  Building:"
	@echo "    build            - Build application"
	@echo "    build-prod       - Build for production (optimized)"
	@echo ""
	@echo "  Testing:"
	@echo "    test             - Run unit tests"
	@echo "    test-verbose     - Run tests with verbose output"
	@echo "    test-coverage    - Run tests with coverage report"
	@echo "    test-api         - Run API integration tests (requires running server)"
	@echo "    test-api-url     - Run API tests against custom URL"
	@echo ""
	@echo "  Code Quality:"
	@echo "    fmt              - Format code"
	@echo "    vet              - Vet code"
	@echo "    lint             - Run golangci-lint"
	@echo ""
	@echo "  Dependencies:"
	@echo "    tidy             - Tidy dependencies"
	@echo "    deps             - Download dependencies"
	@echo "    vendor           - Create vendor directory"
	@echo ""
	@echo "  Docker:"
	@echo "    docker-up        - Start Docker services"
	@echo "    docker-down      - Stop Docker services"
	@echo "    docker-logs      - Show Docker logs"
	@echo "    docker-build     - Build Docker image"
	@echo "    docker-rebuild   - Rebuild Docker services"
	@echo ""
	@echo "  Database:"
	@echo "    db-migrate       - Run database migrations"
	@echo "    db-reset         - Reset database and run migrations"
	@echo "    db-connect       - Connect to database"
	@echo ""
	@echo "  Setup:"
	@echo "    setup            - Setup development environment"
	@echo "    install-tools    - Install development tools"
	@echo ""
	@echo "  Cleanup:"
	@echo "    clean            - Clean build artifacts"
	@echo "    clean-all        - Clean all artifacts"
	@echo ""
	@echo "  help             - Show this help message"