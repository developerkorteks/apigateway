# API Fallback System - Deployment Guide

## Overview
This guide covers deployment options for the API Fallback System, with special focus on CasaOS deployment.

## 🐳 Docker Deployment

### Prerequisites
- Docker Engine 20.10+
- Docker Compose 2.0+
- 256MB RAM minimum
- 100MB storage space

### Quick Start with Docker Compose

1. **Clone and prepare the project:**
```bash
git clone <repository-url>
cd apifallback
cp .env.example .env
```

2. **Build and run:**
```bash
make docker-build
make docker-run
```

3. **Access the application:**
- Web UI: http://localhost:8080/dashboard/
- API Documentation: http://localhost:8080/swagger/
- Health Check: http://localhost:8080/health

### Using Deployment Script (Recommended)

The project includes a deployment script that automates the entire process:

```bash
# Deploy for development
./deploy.sh dev

# Deploy for production
./deploy.sh prod

# Deploy with Nginx reverse proxy (uses ports 8081/8443)
./deploy.sh nginx

# Check application health
./deploy.sh health

# View logs
./deploy.sh logs

# Stop services
./deploy.sh stop

# Clean up everything
./deploy.sh clean
```

### Manual Docker Commands

```bash
# Build the image
docker build -t apifallback:latest .

# Run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f apifallback

# Stop services
docker-compose down
```

## 🏠 CasaOS Deployment

CasaOS is a simple, elegant, open-source home cloud system. This application is optimized for CasaOS deployment.

### Method 1: Using CasaOS App Store (Recommended)

1. **Import the app configuration:**
   - Copy the `casa-config.json` file to your CasaOS
   - In CasaOS, go to App Store → Custom Install
   - Import the configuration file

2. **Configure the application:**
   - Set environment variables as needed
   - Configure volume mappings
   - Set port mappings (default: 8080)

3. **Deploy:**
   - Click "Install" in CasaOS
   - Wait for the container to start
   - Access via the CasaOS dashboard

### Method 2: Manual Docker Installation in CasaOS

1. **Build the Docker image:**
```bash
make docker-build
```

2. **Create the container in CasaOS:**
   - Go to CasaOS → Docker → Add Container
   - Use image: `apifallback:latest`
   - Configure ports: `8080:8080`
   - Set environment variables (see Configuration section)
   - Add volume: `/DATA/AppData/apifallback/data:/app/data`

### CasaOS Configuration

| Setting | Value | Description |
|---------|-------|-------------|
| **Image** | `apifallback:latest` | Docker image name |
| **Port** | `8080:8080` | Web interface port |
| **Volume** | `/DATA/AppData/apifallback/data:/app/data` | Data persistence |
| **Network** | `bridge` | Network mode |
| **Restart** | `unless-stopped` | Restart policy |

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Application port |
| `DATABASE_PATH` | `/app/data/data.db` | SQLite database path |
| `REDIS_ADDR` | `redis:6379` | Redis server address |
| `REDIS_DB` | `0` | Redis database number |
| `API_TIMEOUT` | `20s` | External API timeout |
| `MAX_CONCURRENCY` | `10` | Max concurrent requests |
| `RATE_LIMIT` | `100` | Requests per minute |
| `RATE_LIMIT_WINDOW` | `1m` | Rate limit window |
| `HEALTH_CHECK_INTERVAL` | `10m` | Health check frequency |

### Volume Mounts

- `/app/data` - Database and persistent data storage

### Network Ports

- `8080/tcp` - Web interface and API endpoints (main application)
- `6379/tcp` - Redis/Valkey cache server
- `8081/tcp` - Nginx reverse proxy (when using nginx profile)
- `8443/tcp` - Nginx HTTPS (when SSL is configured)

**Note**: Ports 80 and 443 are intentionally avoided to prevent conflicts with system services.

## 🔧 Production Deployment

### Recommended Settings

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  apifallback:
    image: apifallback:latest
    restart: unless-stopped
    environment:
      - PORT=8080
      - API_TIMEOUT=30s
      - MAX_CONCURRENCY=20
      - RATE_LIMIT=200
    volumes:
      - /opt/apifallback/data:/app/data
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Security Considerations

1. **Reverse Proxy:** Use nginx or Traefik for SSL termination
2. **Firewall:** Restrict access to necessary ports only
3. **Updates:** Regularly update the container image
4. **Monitoring:** Set up health checks and logging

### Performance Tuning

1. **Memory:** Allocate at least 512MB for production
2. **CPU:** 1 CPU core minimum, 2+ recommended
3. **Storage:** SSD recommended for database performance
4. **Network:** Ensure stable internet connection for API calls

## 📊 Monitoring and Maintenance

### Health Checks

The application provides several health check endpoints:

- `/health` - Basic health status
- `/dashboard/health` - Detailed API health status
- `/dashboard/stats` - Performance statistics

### Logs

View application logs:
```bash
# Docker Compose
docker-compose logs -f apifallback

# Direct Docker
docker logs -f apifallback
```

### Backup

Backup the database:
```bash
# Copy database file
docker cp apifallback:/app/data/data.db ./backup-$(date +%Y%m%d).db
```

### Updates

Update the application:
```bash
# Pull latest changes
git pull

# Rebuild and restart
make docker-build
docker-compose up -d
```

## 🚨 Troubleshooting

### Common Issues

1. **Port already in use:**
   ```bash
   # Change port in docker-compose.yml or .env
   PORT=8081
   ```

2. **Database permission errors:**
   ```bash
   # Fix volume permissions
   sudo chown -R 1001:1001 /DATA/AppData/apifallback/data
   ```

3. **Redis connection failed:**
   ```bash
   # Check Redis container status
   docker-compose ps redis
   docker-compose logs redis
   ```

4. **API endpoints not responding:**
   - Check external API availability
   - Verify network connectivity
   - Review application logs

### Support

For issues and support:
1. Check the application logs
2. Verify configuration settings
3. Test external API connectivity
4. Review the troubleshooting section

## 📝 Notes

- The application uses SQLite for data storage and Redis/Valkey for caching
- All external API calls are cached to improve performance
- The system automatically falls back to alternative APIs when primary sources fail
- Web dashboard provides real-time monitoring and management capabilities