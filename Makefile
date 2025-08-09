.PHONY: build run test clean docker-build docker-run docker-stop docker-clean deploy-casa

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=apifallback
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/main.go

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/main.go
	./$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v cmd/main.go

# Docker commands
docker-build:
	docker build -t apifallback:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-clean:
	docker-compose down -v
	docker rmi apifallback:latest

# Development
dev:
	air -c .air.toml

# Database setup
setup-db:
	sqlite3 data.db < setup_apis.sql

# CasaOS deployment
deploy-casa: docker-build
	@echo "Building Docker image for CasaOS..."
	@echo "Image built successfully!"
	@echo "To deploy in CasaOS:"
	@echo "1. Copy casa-config.json to your CasaOS app store"
	@echo "2. Import the app using the configuration"
	@echo "3. Configure volumes and environment variables as needed"

# Production build
build-prod:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BINARY_NAME) cmd/main.go

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Build and run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  build-linux  - Build for Linux"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker containers"
	@echo "  docker-clean - Clean Docker containers and images"
	@echo "  deploy-casa  - Prepare for CasaOS deployment"
	@echo "  build-prod   - Build optimized production binary"
	@echo "  help         - Show this help message"