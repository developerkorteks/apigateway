# Variables
BINARY_NAME=api-server
DOCKER_IMAGE=winbutv-api
PORT=8080

# Build the application
build:
	go build -o $(BINARY_NAME) main.go

# Run the application
run: build
	./$(BINARY_NAME)

# Run with custom port
run-port: build
	PORT=$(PORT) ./$(BINARY_NAME)

# Run in development mode
dev:
	ENVIRONMENT=development go run main.go

# Test the application
test:
	go test -v ./...

# Test specific scraper
test-home:
	go test -v ./scrape -run TestHome

test-anime:
	go test -v ./scrape -run TestScrapeAnimeTerbaruAnimasuLimited

test-film:
	go test -v ./scrape -run TestScrapeFilmLimited

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run:
	docker run -p $(PORT):8080 --name winbutv-api-container $(DOCKER_IMAGE)

# Run with Docker Compose
docker-compose-up:
	docker-compose up --build

# Stop Docker Compose
docker-compose-down:
	docker-compose down

# API Health Check
health:
	curl -s http://localhost:$(PORT)/health | jq .

# Test API endpoints
test-api: health
	@echo "Testing /api/v1/home endpoint..."
	curl -s http://localhost:$(PORT)/api/v1/home | jq . | head -20
	@echo "\nTesting /api/v1/anime-terbaru endpoint..."
	curl -s "http://localhost:$(PORT)/api/v1/anime-terbaru?page=1" | jq . | head -20
	@echo "\nTesting /api/v1/movie endpoint..."
	curl -s "http://localhost:$(PORT)/api/v1/movie?page=1" | jq . | head -20

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Show help
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Build and run the application"
	@echo "  run-port        - Run with custom port (make run-port PORT=8081)"
	@echo "  dev             - Run in development mode"
	@echo "  test            - Run all tests"
	@echo "  test-home       - Test home scraper"
	@echo "  test-anime      - Test anime scraper"
	@echo "  test-film       - Test film scraper"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-compose-up   - Run with Docker Compose"
	@echo "  docker-compose-down - Stop Docker Compose"
	@echo "  health          - Check API health"
	@echo "  test-api        - Test all API endpoints"
	@echo "  install-tools   - Install development tools"
	@echo "  help            - Show this help message"

.PHONY: build run run-port dev test test-home test-anime test-film clean deps fmt lint docker-build docker-run docker-compose-up docker-compose-down health test-api install-tools help