# 🚀 Deployment Guide - API Gateway

This guide provides comprehensive instructions for deploying the API Gateway in various environments.

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- SQLite3
- Redis/Valkey

## 🔧 Quick Setup

### 1. Update Configuration

```bash
# Run the update script to configure everything
./update_deployment.sh
```

### 2. Configure Environment Variables

Edit `.env` file with your actual API endpoints:

```bash
# Example configuration
GOMUNIME_URL=http://your-gomunime-api:8001
WINBUTV_URL=http://your-winbutv-api:8002
SAMEHADAKU_URL=http://128.199.109.211:8182
```

## 🏠 Local Development

### Option 1: Direct Run
```bash
# Set environment variables
export GOMUNIME_URL=http://localhost:8001
export WINBUTV_URL=http://localhost:8002
export SAMEHADAKU_URL=http://128.199.109.211:8182

# Run the application
./run.sh
```

### Option 2: Docker Compose
```bash
# Development with Docker
docker-compose up -d
```

## 🐳 Docker Deployment

### Development Environment
```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f apifallback

# Stop
docker-compose down
```

### Production Environment
```bash
# Use production configuration
docker-compose -f docker-compose.prod.yml up -d

# With custom environment file
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

## 🌐 Production Deployment

### 1. Server Setup

```bash
# Clone repository
git clone <your-repo-url>
cd apigateway

# Run setup script
./update_deployment.sh

# Configure production URLs in .env
nano .env
```

### 2. Environment Configuration

Create `.env.prod` for production:

```bash
# Production Configuration
PORT=8080
GIN_MODE=release
DATABASE_PATH=/app/data/data.db
REDIS_ADDR=redis:6379

# Production API URLs
GOMUNIME_URL=https://api.gomunime.com
WINBUTV_URL=https://api.winbu.tv
SAMEHADAKU_URL=https://api.samehadaku.com
OTAKUDESU_URL=https://api.otakudesu.com
KUSONIME_URL=https://api.kusonime.com

# Performance Settings
API_TIMEOUT=30s
MAX_CONCURRENCY=20
RATE_LIMIT=1000
HEALTH_CHECK_INTERVAL=5m

# Cache Settings
CACHE_TTL_HOME=15m
CACHE_TTL_ANIME_TERBARU=10m
CACHE_TTL_SEARCH=5m
CACHE_TTL_DETAIL=30m
```

### 3. Deploy with Nginx (Recommended)

```bash
# Deploy with Nginx proxy
docker-compose -f docker-compose.prod.yml up -d

# Or use the deploy script
./deploy.sh --production --with-nginx
```

## 🔍 Health Checks & Monitoring

### Health Check Endpoints
- **Application Health**: `http://localhost:8080/health`
- **API Sources Status**: `http://localhost:8080/health/sources`
- **Database Status**: `http://localhost:8080/health/database`

### Monitoring Commands
```bash
# Check application status
curl http://localhost:8080/health

# View detailed health information
curl http://localhost:8080/health/sources | jq

# Check logs
docker-compose logs -f apifallback

# Monitor resource usage
docker stats apifallback
```

## 🛠️ Troubleshooting

### Common Issues

#### 1. Port Already in Use
```bash
# Check what's using port 8080
sudo netstat -tulpn | grep :8080

# Kill the process or change port in .env
PORT=8081
```

#### 2. Database Connection Issues
```bash
# Check database file permissions
ls -la data.db

# Reset database
rm data.db
./main  # Will recreate database
```

#### 3. API Sources Not Responding
```bash
# Check API source health
curl http://localhost:8080/health/sources

# Test individual API
curl http://your-api-url/api/v1/home
```

#### 4. Redis Connection Issues
```bash
# Check Redis container
docker-compose ps redis

# Test Redis connection
docker-compose exec redis valkey-cli ping
```

### Debug Mode

Enable debug logging:
```bash
# Set debug mode
export GIN_MODE=debug
export LOG_LEVEL=debug

# Run application
./main
```

## 📊 Performance Optimization

### 1. Cache Configuration
```bash
# Optimize cache TTL based on your needs
CACHE_TTL_HOME=30m      # Less frequent updates
CACHE_TTL_SEARCH=2m     # More frequent updates
```

### 2. Rate Limiting
```bash
# Adjust based on your traffic
RATE_LIMIT=500          # Requests per window
RATE_LIMIT_WINDOW=1m    # Time window
```

### 3. Concurrency
```bash
# Increase for high-traffic scenarios
MAX_CONCURRENCY=50      # Concurrent API requests
```

## 🔒 Security Considerations

### 1. Environment Variables
- Never commit `.env` files to version control
- Use secrets management in production
- Rotate API keys regularly

### 2. Network Security
```bash
# Use internal networks for API communication
# Configure firewall rules
# Enable HTTPS in production
```

### 3. Access Control
```bash
# Configure CORS origins
CORS_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Use API keys for authentication
API_KEY=your-secure-api-key
```

## 📈 Scaling

### Horizontal Scaling
```bash
# Scale application instances
docker-compose up -d --scale apifallback=3

# Use load balancer (Nginx/HAProxy)
# Configure session affinity if needed
```

### Database Scaling
```bash
# For high-traffic scenarios, consider:
# - Read replicas
# - Connection pooling
# - Database clustering
```

## 🔄 Updates & Maintenance

### Application Updates
```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
docker-compose build --no-cache
docker-compose up -d
```

### Database Migrations
```bash
# Backup database before updates
cp data.db data.db.backup

# Run migrations if any
./update_deployment.sh
```

### Log Rotation
```bash
# Configure log rotation in docker-compose
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## 📞 Support

For issues and questions:
1. Check the troubleshooting section above
2. Review application logs
3. Test individual components
4. Check API source availability

## 🎯 Access Points

After successful deployment:

- **Web Dashboard**: http://localhost:8080/dashboard/
- **API Documentation**: http://localhost:8080/swagger/
- **Health Check**: http://localhost:8080/health
- **API Endpoints**: http://localhost:8080/api/v1/

---

**Note**: Replace `localhost` with your actual domain/IP in production deployments.