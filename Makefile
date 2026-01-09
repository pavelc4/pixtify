.PHONY: dev build test install clean help

dev:
	@echo "Starting development server..."
	go run cmd/api/main.go

build:
	@echo "Building binary..."
	go build -o bin/pixtify cmd/api/main.go

test:
	@echo "Running tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

lint:
	@echo "Running linter..."
	golangci-lint run

help:
	@echo "Available commands:"
	@echo "  make dev            - Run development server"
	@echo "  make build          - Build binary"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make install        - Install dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run linter"
	@echo "  make help           - Show this help message"

.DEFAULT_GOAL := help
