# Deployment Guide - Winbu.TV Web Scraping API

## 🚀 Quick Start

### Local Development
```bash
# Clone dan setup
git clone <repository-url>
cd winbutv

# Install dependencies
make deps

# Run in development mode
make dev

# Or run with custom port
make run-port PORT=8081
```

### Using Docker
```bash
# Build dan run dengan Docker
make docker-build
make docker-run

# Atau menggunakan Docker Compose
make docker-compose-up
```

## 🔧 Environment Configuration

Buat file `.env` untuk konfigurasi:
```bash
# Server Configuration
PORT=8080
ENVIRONMENT=production

# Target Site Configuration
BASE_URL=https://winbu.tv

# Scraping Configuration
TIMEOUT=30s
RATE_LIMIT=1s
MAX_RETRIES=3

# Cache Configuration
CACHE_ENABLED=true
CACHE_TTL=5m
```

## 📋 Production Deployment

### 1. Docker Deployment (Recommended)

**Dockerfile sudah siap untuk production:**
```dockerfile
# Multi-stage build untuk optimized image
FROM golang:1.24.4-alpine AS builder
# ... build process
FROM alpine:latest
# ... final image dengan security best practices
```

**Deploy dengan Docker Compose:**
```yaml
version: '3.8'
services:
  winbutv-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - CACHE_ENABLED=true
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
```

### 2. Binary Deployment

```bash
# Build untuk production
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-server main.go

# Deploy binary
scp api-server user@server:/opt/winbutv-api/
ssh user@server "cd /opt/winbutv-api && ./api-server"
```

### 3. Systemd Service

Buat file `/etc/systemd/system/winbutv-api.service`:
```ini
[Unit]
Description=Winbu.TV Web Scraping API
After=network.target

[Service]
Type=simple
User=winbutv
WorkingDirectory=/opt/winbutv-api
ExecStart=/opt/winbutv-api/api-server
Restart=always
RestartSec=5

Environment=PORT=8080
Environment=ENVIRONMENT=production
Environment=BASE_URL=https://winbu.tv
Environment=CACHE_ENABLED=true

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable winbutv-api
sudo systemctl start winbutv-api
sudo systemctl status winbutv-api
```

## 🔍 Monitoring & Health Checks

### Health Check Endpoint
```bash
curl http://localhost:8080/health
# Response: {"status": "ok", "message": "API is running"}
```

### Monitoring dengan Docker
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### Log Monitoring
```bash
# Docker logs
docker logs -f winbutv-api-container

# Systemd logs
journalctl -u winbutv-api -f
```

## 🌐 Reverse Proxy Setup

### Nginx Configuration
```nginx
server {
    listen 80;
    server_name api.winbutv.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization";
    }

    location /health {
        proxy_pass http://localhost:8080/health;
        access_log off;
    }
}
```

### Apache Configuration
```apache
<VirtualHost *:80>
    ServerName api.winbutv.example.com
    
    ProxyPreserveHost On
    ProxyRequests Off
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/
    
    # CORS headers
    Header always set Access-Control-Allow-Origin "*"
    Header always set Access-Control-Allow-Methods "GET, POST, OPTIONS"
    Header always set Access-Control-Allow-Headers "Content-Type, Authorization"
</VirtualHost>
```

## 📊 Performance Tuning

### 1. Caching Strategy
```bash
# Enable caching untuk production
CACHE_ENABLED=true
CACHE_TTL=5m  # Adjust based on data freshness needs
```

### 2. Rate Limiting
```bash
# Adjust rate limiting based on target site capacity
RATE_LIMIT=1s    # Conservative default
RATE_LIMIT=500ms # More aggressive (use with caution)
```

### 3. Concurrent Requests
```bash
# Increase timeout for slow responses
TIMEOUT=60s

# Increase retries for reliability
MAX_RETRIES=5
```

## 🔒 Security Considerations

### 1. Network Security
- Deploy behind reverse proxy
- Use HTTPS in production
- Implement rate limiting at proxy level
- Whitelist allowed origins for CORS

### 2. Application Security
- Run as non-root user
- Use minimal Docker image (alpine)
- Regular security updates
- Monitor for unusual traffic patterns

### 3. API Security
```bash
# Optional: Add API key authentication
# (Not implemented in current version)
API_KEY=your-secret-key
```

## 📈 Scaling Considerations

### 1. Horizontal Scaling
```yaml
# Docker Compose with multiple instances
version: '3.8'
services:
  winbutv-api-1:
    build: .
    ports:
      - "8081:8080"
  winbutv-api-2:
    build: .
    ports:
      - "8082:8080"
  
  nginx:
    image: nginx
    ports:
      - "80:80"
    # Load balancer configuration
```

### 2. Database Integration (Future Enhancement)
```bash
# For persistent caching and analytics
DATABASE_URL=postgresql://user:pass@localhost/winbutv
REDIS_URL=redis://localhost:6379
```

## 🧪 Testing in Production

### 1. Smoke Tests
```bash
# Test all endpoints
curl http://your-domain/health
curl http://your-domain/api/v1/home
curl "http://your-domain/api/v1/anime-terbaru?page=1"
curl "http://your-domain/api/v1/movie?page=1"
curl "http://your-domain/api/v1/search?q=naruto"
```

### 2. Load Testing
```bash
# Using Apache Bench
ab -n 1000 -c 10 http://your-domain/api/v1/home

# Using curl for continuous monitoring
while true; do
  curl -s http://your-domain/health | jq .
  sleep 30
done
```

## 🚨 Troubleshooting

### Common Issues

1. **Port Already in Use**
   ```bash
   # Find process using port
   lsof -i :8080
   # Kill process or use different port
   PORT=8081 ./api-server
   ```

2. **Permission Denied**
   ```bash
   # Make binary executable
   chmod +x api-server
   # Or run with proper user
   sudo -u winbutv ./api-server
   ```

3. **Memory Issues**
   ```bash
   # Monitor memory usage
   docker stats winbutv-api-container
   # Adjust cache TTL if needed
   CACHE_TTL=1m
   ```

4. **Target Site Blocking**
   ```bash
   # Increase rate limiting
   RATE_LIMIT=2s
   # Reduce concurrent requests
   MAX_RETRIES=2
   ```

## 📞 Support & Maintenance

### Regular Maintenance Tasks
1. Monitor logs for errors
2. Check confidence scores trends
3. Update dependencies regularly
4. Monitor target site changes
5. Backup configuration files

### Performance Monitoring
```bash
# Check API response times
curl -w "@curl-format.txt" -s -o /dev/null http://your-domain/api/v1/home

# Monitor confidence scores
curl -s http://your-domain/api/v1/home | jq .confidence_score
```

## 🎯 Integration with Django KortekStream

### Django Settings
```python
# settings.py
WINBUTV_API_BASE_URL = "http://your-domain"
WINBUTV_API_TIMEOUT = 30
WINBUTV_CONFIDENCE_THRESHOLD = 0.7  # Minimum confidence score
```

### Django Usage Example
```python
import requests

def get_home_data():
    response = requests.get(f"{settings.WINBUTV_API_BASE_URL}/api/v1/home")
    data = response.json()
    
    if data.get('confidence_score', 0) >= settings.WINBUTV_CONFIDENCE_THRESHOLD:
        return data
    else:
        # Fallback to other API or cached data
        return get_fallback_data()
```

Deployment guide ini memberikan semua informasi yang diperlukan untuk deploy API ke production dengan aman dan efisien.