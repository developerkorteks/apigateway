# API Fallback System - Usage Guide

## Quick Start

### 1. Build dan Jalankan Aplikasi

```bash
# Build aplikasi
go build -o apifallback cmd/main.go

# Atau gunakan script runner
./run.sh
```

### 2. Akses Dashboard

Buka browser dan akses: `http://localhost:8080/dashboard/`

Dashboard menyediakan:
- Real-time monitoring
- Health status semua API sources
- Request logs
- Statistics
- API source management

### 3. Test API Endpoints

```bash
# Test semua endpoints
./test_api.sh

# Test manual
curl "http://localhost:8080/api/v1/home?category=anime"
curl "http://localhost:8080/api/v1/search?q=naruto&category=anime"
```

## API Endpoints

### Core API Endpoints

| Endpoint | Description | Parameters |
|----------|-------------|------------|
| `GET /api/v1/home` | Homepage data | `category` (default: anime) |
| `GET /api/v1/jadwal-rilis` | Release schedule | `category` |
| `GET /api/v1/jadwal-rilis/{day}` | Daily schedule | `category` |
| `GET /api/v1/anime-terbaru` | Latest anime | `category` |
| `GET /api/v1/movie` | Movies | `category` |
| `GET /api/v1/anime-detail` | Anime details | `category`, `slug` |
| `GET /api/v1/episode-detail` | Episode details | `category`, `slug` |
| `GET /api/v1/search` | Search | `category`, `q` |

### Dashboard API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /dashboard/health` | Health status |
| `GET /dashboard/logs` | Request logs |
| `GET /dashboard/stats` | Statistics |
| `GET /dashboard/categories` | Categories list |

### System Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | System health check |

## Configuration

### Environment Variables

```bash
# Server
PORT=8080

# Database
DATABASE_PATH=./data.db

# Redis (optional)
REDIS_ADDR=localhost:6379
REDIS_DB=0

# API Settings
API_TIMEOUT=20s
MAX_CONCURRENCY=10

# Rate Limiting
RATE_LIMIT=100
RATE_LIMIT_WINDOW=1m

# Health Check
HEALTH_CHECK_INTERVAL=10m
```

### Database Setup

Database SQLite akan dibuat otomatis dengan struktur default:
- 1 kategori: "anime"
- 7 endpoints standar
- 2 API sources: multiplescrape (port 8081), winbutv (port 8082)

Untuk kustomisasi, edit file `setup_apis.sql` dan jalankan:

```bash
sqlite3 data.db < setup_apis.sql
```

## Fallback Mechanism

### Cara Kerja

1. **Request masuk** → Sistem identifikasi kategori dan endpoint
2. **Concurrent requests** → Kirim request ke semua primary APIs secara paralel
3. **Validation** → Validasi confidence_score (≥0.5) dan schema JSON
4. **Priority response** → Return response pertama yang valid berdasarkan priority
5. **Fallback** → Jika semua primary gagal, coba fallback APIs
6. **Caching** → Cache response yang valid
7. **Error handling** → Return 503 jika semua gagal

### Response Headers

```
X-Source: multiplescrape          # Source yang digunakan
X-Response-Time: 1.234s          # Response time
X-Cache: HIT|MISS                # Cache status
```

### Error Responses

```json
{
  "error": true,
  "message": "All API sources failed for endpoint /api/v1/home",
  "source": "apicategorywithfallback"
}
```

## Monitoring

### Health Check

Background worker melakukan health check setiap 10 menit:
- Test koneksi ke semua API sources
- Update status: OK, TIMEOUT, ERROR
- Log response time dan error messages

### Logging

Structured JSON logging dengan informasi:
- Request details
- Response times
- Source used
- Fallback usage
- Error messages

### Metrics

Dashboard menampilkan:
- Total requests
- Success/failure rates
- Fallback usage statistics
- Average response times
- API source health status

## API Source Management

### Menambah API Source Baru

1. **Via Database**:
```sql
INSERT INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary) 
VALUES (1, 'new_source', 'http://localhost:8083', 3, 1);
```

2. **Via Dashboard**: Gunakan interface web (coming soon)

### Menambah Fallback API

```sql
INSERT INTO fallback_apis (api_source_id, fallback_url, priority) 
VALUES (1, 'http://backup.example.com/api/v1', 1);
```

### Menonaktifkan API Source

```sql
UPDATE api_sources SET is_active = 0 WHERE source_name = 'problematic_source';
```

## Troubleshooting

### Common Issues

1. **Database locked**
   - Pastikan tidak ada proses lain yang menggunakan database
   - Restart aplikasi

2. **Redis connection failed**
   - Sistem otomatis fallback ke memory cache
   - Check Redis server status

3. **All APIs failing**
   - Check API source URLs
   - Verify network connectivity
   - Check health status di dashboard

4. **High response times**
   - Check API_TIMEOUT setting
   - Monitor API source performance
   - Consider adding more fallback APIs

### Debug Mode

Set environment variable untuk debug:
```bash
export GIN_MODE=debug
./apifallback
```

### Logs Location

Logs ditulis ke stdout dalam format JSON. Untuk file logging:
```bash
./apifallback > app.log 2>&1
```

## Performance Tuning

### Recommended Settings

**Production**:
```bash
API_TIMEOUT=15s
MAX_CONCURRENCY=20
RATE_LIMIT=1000
HEALTH_CHECK_INTERVAL=5m
```

**Development**:
```bash
API_TIMEOUT=30s
MAX_CONCURRENCY=5
RATE_LIMIT=100
HEALTH_CHECK_INTERVAL=10m
```

### Cache TTL Optimization

Default TTL per endpoint:
- `/api/v1/home`: 15 minutes
- `/api/v1/jadwal-rilis`: 30 minutes
- `/api/v1/anime-terbaru`: 15 minutes
- `/api/v1/movie`: 1 hour
- `/api/v1/anime-detail`: 1 hour
- `/api/v1/episode-detail`: 30 minutes
- `/api/v1/search`: 10 minutes

## Integration

### Dengan Existing APIs

Sistem ini dirancang untuk berintegrasi dengan:
- `multiplescrape` (port 8081)
- `winbutv` (port 8082)
- API lainnya yang mengikuti schema yang sama

### Client Integration

```javascript
// JavaScript example
const response = await fetch('http://localhost:8080/api/v1/home?category=anime');
const data = await response.json();

console.log('Source:', response.headers.get('X-Source'));
console.log('Cache:', response.headers.get('X-Cache'));
console.log('Data:', data);
```

```python
# Python example
import requests

response = requests.get('http://localhost:8080/api/v1/home', 
                       params={'category': 'anime'})

print(f"Source: {response.headers.get('X-Source')}")
print(f"Cache: {response.headers.get('X-Cache')}")
print(f"Data: {response.json()}")
```

## Security

### Rate Limiting

- Default: 100 requests per minute per IP
- Configurable via `RATE_LIMIT` environment variable
- Returns 429 status code when exceeded

### Input Validation

- Parameter sanitization
- URL validation
- Schema validation untuk responses

### Recommended Security Headers

Tambahkan reverse proxy (nginx) dengan headers:
```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
```