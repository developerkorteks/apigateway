#!/bin/bash

# API Fallback System Deployment Script
# This script helps deploy the application in different environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

# Build the application
build_app() {
    print_info "Building API Fallback System..."
    
    if make docker-build; then
        print_success "Application built successfully"
    else
        print_error "Failed to build application"
        exit 1
    fi
}

# Deploy for development
deploy_dev() {
    print_info "Deploying for development..."
    
    # Copy environment file
    if [ ! -f .env ]; then
        cp .env.example .env
        print_warning "Created .env file from .env.example. Please review and update as needed."
    fi
    
    # Start services
    docker-compose up -d
    
    print_success "Development deployment completed"
    print_info "Access the application at:"
    print_info "  - Web Dashboard: http://localhost:8080/dashboard/"
    print_info "  - API Documentation: http://localhost:8080/swagger/"
    print_info "  - Health Check: http://localhost:8080/health"
}

# Deploy for production
deploy_prod() {
    print_info "Deploying for production..."
    
    # Check if production environment file exists
    if [ ! -f .env.production ]; then
        print_warning "No .env.production file found. Creating from .env.example..."
        cp .env.example .env.production
        print_warning "Please review and update .env.production before continuing."
        read -p "Press Enter to continue after updating .env.production..."
    fi
    
    # Start production services
    docker-compose -f docker-compose.prod.yml up -d
    
    print_success "Production deployment completed"
    print_info "Access the application at:"
    print_info "  - Web Dashboard: http://localhost:8080/dashboard/"
    print_info "  - API Documentation: http://localhost:8080/swagger/"
    print_info "  - Health Check: http://localhost:8080/health"
    
    if docker-compose -f docker-compose.prod.yml ps nginx &> /dev/null; then
        print_info "  - Nginx Proxy: http://localhost:8081/"
    fi
}

# Deploy with Nginx
deploy_with_nginx() {
    print_info "Deploying with Nginx reverse proxy..."
    
    # Create SSL directory if it doesn't exist
    mkdir -p ssl
    
    # Start services with nginx profile
    docker-compose -f docker-compose.prod.yml --profile with-nginx up -d
    
    print_success "Deployment with Nginx completed"
    print_info "Access the application at:"
    print_info "  - Direct Access: http://localhost:8080/dashboard/"
    print_info "  - Via Nginx: http://localhost:8081/"
    print_info "  - API Documentation: http://localhost:8081/swagger/"
}

# Stop services
stop_services() {
    print_info "Stopping services..."
    
    if [ -f docker-compose.prod.yml ]; then
        docker-compose -f docker-compose.prod.yml down
    fi
    
    docker-compose down
    
    print_success "Services stopped"
}

# Clean up
cleanup() {
    print_info "Cleaning up..."
    
    # Stop and remove containers
    stop_services
    
    # Remove images
    docker rmi apifallback:latest 2>/dev/null || true
    
    # Remove volumes (optional)
    read -p "Do you want to remove data volumes? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume rm apifallback_apifallback_data apifallback_redis_data 2>/dev/null || true
        print_warning "Data volumes removed"
    fi
    
    print_success "Cleanup completed"
}

# Show logs
show_logs() {
    print_info "Showing application logs..."
    docker-compose logs -f apifallback
}

# Health check
health_check() {
    print_info "Performing health check..."
    
    # Wait for service to be ready
    sleep 5
    
    if curl -f http://localhost:8080/health &> /dev/null; then
        print_success "Application is healthy"
        
        # Show detailed status
        echo
        print_info "Detailed status:"
        curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health
    else
        print_error "Application health check failed"
        print_info "Check logs with: $0 logs"
        exit 1
    fi
}

# Show help
show_help() {
    echo "API Fallback System Deployment Script"
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  dev         Deploy for development (default)"
    echo "  prod        Deploy for production"
    echo "  nginx       Deploy with Nginx reverse proxy"
    echo "  build       Build the application only"
    echo "  stop        Stop all services"
    echo "  clean       Clean up containers and images"
    echo "  logs        Show application logs"
    echo "  health      Perform health check"
    echo "  help        Show this help message"
    echo
    echo "Examples:"
    echo "  $0 dev      # Deploy for development"
    echo "  $0 prod     # Deploy for production"
    echo "  $0 nginx    # Deploy with Nginx"
    echo "  $0 health   # Check application health"
}

# Main script
main() {
    case "${1:-dev}" in
        "dev"|"development")
            check_docker
            build_app
            deploy_dev
            health_check
            ;;
        "prod"|"production")
            check_docker
            build_app
            deploy_prod
            health_check
            ;;
        "nginx")
            check_docker
            build_app
            deploy_with_nginx
            health_check
            ;;
        "build")
            check_docker
            build_app
            ;;
        "stop")
            stop_services
            ;;
        "clean"|"cleanup")
            cleanup
            ;;
        "logs")
            show_logs
            ;;
        "health"|"check")
            health_check
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"