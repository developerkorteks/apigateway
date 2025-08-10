#!/bin/bash

# Update Deployment Configuration Script
# This script updates the deployment configuration to be more dynamic and production-ready

set -e

echo "🔧 Updating deployment configuration..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if .env file exists, if not create from example
if [ ! -f .env ]; then
    print_info "Creating .env file from .env.example..."
    cp .env.example .env
    print_success ".env file created"
else
    print_info ".env file already exists"
fi

# Update database configuration
print_info "Updating database configuration..."
if [ -f "fix_database_config.sql" ]; then
    sqlite3 data.db < fix_database_config.sql
    print_success "Database configuration updated"
else
    print_warning "fix_database_config.sql not found, skipping database update"
fi

# Build the application
print_info "Building application..."
go build -o main cmd/main.go
print_success "Application built successfully"

# Create production docker-compose if it doesn't exist
if [ ! -f "docker-compose.prod.yml" ]; then
    print_info "Creating production docker-compose configuration..."
    cp docker-compose.yml docker-compose.prod.yml
    
    # Update production configuration
    sed -i 's/GIN_MODE: ${GIN_MODE:-release}/GIN_MODE: release/' docker-compose.prod.yml
    sed -i 's/localhost:8001/gomunime:8001/' docker-compose.prod.yml
    sed -i 's/localhost:8002/winbutv:8002/' docker-compose.prod.yml
    sed -i 's/localhost:8081/multiplescrape:8081/' docker-compose.prod.yml
    
    print_success "Production docker-compose.prod.yml created"
fi

# Validate configuration
print_info "Validating configuration..."

# Check if required environment variables are set
required_vars=("GOMUNIME_URL" "WINBUTV_URL" "SAMEHADAKU_URL")
missing_vars=()

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        missing_vars+=("$var")
    fi
done

if [ ${#missing_vars[@]} -gt 0 ]; then
    print_warning "The following environment variables are not set:"
    for var in "${missing_vars[@]}"; do
        echo "  - $var"
    done
    print_info "Please check your .env file or set these variables"
fi

# Test database connection
print_info "Testing database connection..."
if sqlite3 data.db "SELECT COUNT(*) FROM api_sources;" > /dev/null 2>&1; then
    source_count=$(sqlite3 data.db "SELECT COUNT(*) FROM api_sources;")
    print_success "Database connection OK - Found $source_count API sources"
else
    print_error "Database connection failed"
    exit 1
fi

# Check if ports are available
print_info "Checking port availability..."
if command -v netstat > /dev/null 2>&1; then
    if netstat -tuln | grep -q ":8080 "; then
        print_warning "Port 8080 is already in use"
    else
        print_success "Port 8080 is available"
    fi
fi

print_success "Deployment configuration updated successfully!"
print_info ""
print_info "Next steps:"
print_info "1. Review and update .env file with your actual API URLs"
print_info "2. For local development: ./run.sh"
print_info "3. For Docker deployment: docker-compose up -d"
print_info "4. For production deployment: docker-compose -f docker-compose.prod.yml up -d"
print_info ""
print_info "Access points after deployment:"
print_info "- Web Dashboard: http://localhost:8080/dashboard/"
print_info "- API Documentation: http://localhost:8080/swagger/"
print_info "- Health Check: http://localhost:8080/health"