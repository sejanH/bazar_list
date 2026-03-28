# Bazar List Makefile

# Load environment variables from .env file (if it exists)
-include .env

.PHONY: help build build-web build-cli clean test test-coverage fmt lint run run-web install deps

# Variables
APP_NAME=bazarlist
WEB_APP_NAME=bazarlist-web
BUILD_DIR=build
CLI_CMD_DIR=cmd/cli
WEB_CMD_DIR=cmd/web
CLI_MAIN_FILE=$(CLI_CMD_DIR)/main.go
WEB_MAIN_FILE=$(WEB_CMD_DIR)/main.go
CLI_BINARY=$(BUILD_DIR)/$(APP_NAME)
WEB_BINARY=$(BUILD_DIR)/$(WEB_APP_NAME)
GO_FILES=$(shell find . -type f -name '*.go' -not -path './vendor/*' -not -path './build/*')

# Database configuration
# These are default values. Override them by setting environment variables
# or creating a .env file in the project root.
DB_USER?=bazarlist
DB_PASSWORD?=bazarlist123
DB_HOST?=localhost
DB_PORT?=3306
DB_NAME?=bazarlist

# Default target
help:
	@echo "Available targets:"
	@echo "  make build          - Build CLI and web applications"
	@echo "  make build-cli      - Build CLI application only"
	@echo "  make build-web      - Build web application only"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run            - Run CLI application"
	@echo "  make run-web        - Run web application"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make fmt            - Format Go code"
	@echo "  make lint           - Run linter"
	@echo "  make deps           - Download dependencies"
	@echo "  make install        - Install the applications"

# Build all applications
build: build-cli build-web

# Build CLI application
build-cli:
	@echo "Building CLI application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(CLI_BINARY) $(CLI_MAIN_FILE)
	@echo "CLI build complete: $(CLI_BINARY)"

# Build web application
build-web:
	@echo "Building web application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(WEB_BINARY) $(WEB_MAIN_FILE)
	@echo "Web build complete: $(WEB_BINARY)"

# Run CLI application
run-cli:
	@echo "Running CLI application..."
	@go run $(CLI_MAIN_FILE)

# Run web application
run-web:
	@echo "Running web application..."
	@DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_NAME=$(DB_NAME) PORT=$(PORT) go run $(WEB_MAIN_FILE)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f *.out
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/"; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Install the application
install: build
	@echo "Installing $(APP_NAME)..."
	@go install $(MAIN_FILE)
	@echo "Installation complete"
