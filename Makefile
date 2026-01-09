.PHONY: dev build test install clean help

dev:
	@echo "Starting development server..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | xargs) && go run cmd/api/main.go; \
	else \
		go run cmd/api/main.go; \
	fi


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

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f postgres

docker-clean:
	@echo "Cleaning Docker volumes..."
	docker-compose down -v

db-shell:
	@echo "Opening PostgreSQL shell..."
	docker exec -it pixtify_db psql -U pixtify -d pixtify_db

help:
	@echo "Available commands:"
	@echo "  make dev            - Run development server"
	@echo "  make build          - Build binary"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make install        - Install dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run linter"
	@echo "  make docker-up      - Start Docker services (PostgreSQL)"
	@echo "  make docker-down    - Stop Docker services"
	@echo "  make docker-logs    - Show Docker logs"
	@echo "  make docker-clean   - Stop and remove Docker volumes"
	@echo "  make db-shell       - Open PostgreSQL shell"
	@echo "  make help           - Show this help message"

.DEFAULT_GOAL := help
