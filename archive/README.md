# API Gateway with Fallback System

A robust API gateway service with fallback capabilities for handling API failures gracefully. This system provides a reliable way to manage API dependencies and ensure service continuity even when upstream services experience issues.

## Features

- **API Fallback Mechanism**: Automatically serves cached responses when upstream APIs fail
- **Rate Limiting**: Protects your services from overload
- **Caching**: Improves performance and reduces load on backend services
- **Health Monitoring**: Continuously checks the health of connected APIs
- **Dashboard**: Web interface for monitoring and managing API endpoints
- **Swagger Documentation**: Interactive API documentation
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **CasaOS Integration**: Optimized for deployment on CasaOS

## Deployment Guide

### Prerequisites
- Docker Engine 20.10+
- Docker Compose 2.0+
- 256MB RAM minimum
- 100MB storage space

### Deployment Options

#### 1. Using the Deployment Script (Recommended)

The repository includes a comprehensive deployment script (`deploy.sh`) that simplifies the deployment process:

```bash
# Development deployment
./deploy.sh dev

# Production deployment
./deploy.sh prod

# Deployment with Nginx reverse proxy
./deploy.sh nginx

# Show all available commands
./deploy.sh help
```

#### 2. Manual Docker Deployment

```bash
# Build the Docker image
make docker-build

# Start the services
docker-compose up -d

# Check application health
curl http://localhost:8080/health

# View logs
docker-compose logs -f apifallback
```

#### 3. CasaOS Deployment

This application is optimized for CasaOS with proper labels and configuration:

1. Build the Docker image:
   ```bash
   make docker-build
   ```

2. Save the image to transfer to your CasaOS server:
   ```bash
   docker save apifallback:latest > apifallback.tar
   ```

3. Transfer the image and docker-compose.yml to your CasaOS server

4. On the CasaOS server:
   ```bash
   docker load < apifallback.tar
   docker-compose up -d
   ```

## Docker Compose Configuration

The application uses the following docker-compose.yml configuration:

```yaml
services:
  apifallback:
    image: apifallback:latest
    container_name: apifallback
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      DATABASE_PATH: /app/data/data.db
      REDIS_ADDR: redis:6379
      REDIS_DB: 0
      API_TIMEOUT: 20s
      MAX_CONCURRENCY: 10
      RATE_LIMIT: 100
      RATE_LIMIT_WINDOW: 1m
      HEALTH_CHECK_INTERVAL: 10m
      GIN_MODE: release
    volumes:
      - apifallback_data:/app/data
    depends_on:
      redis:
        condition: service_healthy
    networks:
      - apifallback_network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    labels:
      - "casaos.name=API Fallback"
      - "casaos.description=API Category with Fallback Service"
      - "casaos.icon=https://cdn-icons-png.flaticon.com/512/2103/2103633.png"
      - "casaos.port=8080"
      - "casaos.scheme=http"

  redis:
    image: valkey/valkey:7-alpine
    container_name: apifallback_redis
    restart: unless-stopped
    expose:
      - "6379"
    volumes:
      - redis_data:/data
    networks:
      - apifallback_network
    command: valkey-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
    healthcheck:
      test: ["CMD", "valkey-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 5s

volumes:
  apifallback_data:
    driver: local
  redis_data:
    driver: local

networks:
  apifallback_network:
    driver: bridge
```

## Access the Application

After deployment, you can access the application at:

- Web Dashboard: http://localhost:8080/dashboard/
- API Documentation: http://localhost:8080/swagger/
- Health Check: http://localhost:8080/health

## For More Information

For detailed deployment instructions and configuration options, see [DEPLOYMENT.md](DEPLOYMENT.md).
