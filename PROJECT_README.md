# API Category with Fallback

Sistem API Gateway berbasis Golang yang tangguh, scalable, dan dinamis dengan mekanisme fallback otomatis, validasi data yang ketat, caching cerdas, dan dashboard manajemen.

## Fitur Utama

- **Fallback Mechanism**: Otomatis beralih ke API fallback jika API utama gagal
- **Schema Validation**: Validasi ketat terhadap struktur JSON response
- **Caching**: Redis/Memory cache untuk mengurangi latensi
- **Health Monitoring**: Background health checker untuk semua API sources
- **Rate Limiting**: Perlindungan terhadap penyalahgunaan
- **Dashboard**: Web interface untuk monitoring dan manajemen
- **Concurrent Processing**: Goroutines untuk request paralel

## Arsitektur

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client        │───▶│  API Gateway     │───▶│  Primary APIs   │
└─────────────────┘    │  (This System)   │    └─────────────────┘
                       │                  │           │
                       │  ┌─────────────┐ │           ▼
                       │  │   Cache     │ │    ┌─────────────────┐
                       │  │ (Redis/Mem) │ │    │  Fallback APIs  │
                       │  └─────────────┘ │    └─────────────────┘
                       │                  │
                       │  ┌─────────────┐ │
                       │  │  Database   │ │
                       │  │  (SQLite)   │ │
                       │  └─────────────┘ │
                       └──────────────────┘
```

## Endpoints

### API Endpoints
- `GET /api/v1/home` - Homepage data
- `GET /api/v1/jadwal-rilis` - Release schedule
- `GET /api/v1/jadwal-rilis/{day}` - Daily release schedule
- `GET /api/v1/anime-terbaru` - Latest anime
- `GET /api/v1/movie` - Movies
- `GET /api/v1/anime-detail` - Anime details
- `GET /api/v1/episode-detail` - Episode details
- `GET /api/v1/search` - Search functionality

### Dashboard Endpoints
- `GET /dashboard/` - Main dashboard
- `GET /dashboard/health` - Health status API
- `GET /dashboard/logs` - Request logs API
- `GET /dashboard/stats` - Statistics API

## Installation

1. Clone repository:
```bash
git clone <repository-url>
cd apicategorywithfallback
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run cmd/main.go
```

## Configuration

Environment variables:
- `PORT` - Server port (default: 8080)
- `DATABASE_PATH` - SQLite database path (default: ./data.db)
- `REDIS_ADDR` - Redis address (default: localhost:6379)
- `REDIS_DB` - Redis database number (default: 0)
- `API_TIMEOUT` - API request timeout (default: 20s)
- `RATE_LIMIT` - Rate limit per minute (default: 100)
- `HEALTH_CHECK_INTERVAL` - Health check interval (default: 10m)

## Database Schema

### Categories
- `id` - Primary key
- `name` - Category name (e.g., "anime")
- `is_active` - Active status

### Endpoints
- `id` - Primary key
- `category_id` - Foreign key to categories
- `path` - Endpoint path

### API Sources
- `id` - Primary key
- `endpoint_id` - Foreign key to endpoints
- `source_name` - Source identifier
- `base_url` - Base URL of the API
- `priority` - Priority order
- `is_primary` - Primary/fallback flag
- `is_active` - Active status

### Fallback APIs
- `id` - Primary key
- `api_source_id` - Foreign key to api_sources
- `fallback_url` - Fallback URL
- `priority` - Priority order

## Validation

Sistem melakukan validasi ketat terhadap response API:

1. **Confidence Score**: Harus >= 0.5
2. **Required Fields**: Validasi field wajib seperti `url`, `judul`, `cover`
3. **URL Validation**: Memastikan URL valid
4. **Placeholder Detection**: Mendeteksi nilai placeholder error

## Caching

- **Cache Key Format**: `category:endpoint:parameter_hash`
- **TTL Configuration**: Berbeda untuk setiap endpoint
- **Fallback**: Memory cache jika Redis tidak tersedia

## Health Monitoring

Background worker melakukan health check setiap 10 menit:
- Status: OK, TIMEOUT, ERROR
- Response time tracking
- Error message logging

## Dashboard

Web dashboard tersedia di `/dashboard/` dengan fitur:
- Real-time statistics
- Health status monitoring
- Request logs viewer
- API source management

## Development

### Project Structure
```
apicategorywithfallback/
├── cmd/main.go                 # Application entry point
├── pkg/                        # Shared packages
│   ├── config/                 # Configuration
│   ├── database/               # Database operations
│   ├── cache/                  # Cache implementation
│   ├── logger/                 # Logging
│   └── validator/              # Response validation
├── internal/                   # Internal packages
│   ├── api/                    # API handlers and routes
│   ├── domain/                 # Domain models
│   └── service/                # Business logic
├── web/                        # Web assets
│   └── templates/              # HTML templates
└── README.md
```

### Adding New API Sources

1. Insert into `api_sources` table
2. Optionally add fallback URLs to `fallback_apis` table
3. System will automatically discover and use new sources

### Adding New Endpoints

1. Add endpoint to `endpoints` table
2. Configure API sources for the endpoint
3. Add validation schema in `pkg/validator/validator.go`
4. Add handler in `internal/api/handlers/`

## Monitoring

### Metrics Available
- Total requests
- Success/failure rates
- Fallback usage statistics
- Response times
- API source health status

### Logging
- Structured JSON logging
- Request/response logging
- Error tracking
- Performance metrics

## Production Deployment

1. Set environment variables
2. Configure Redis (recommended)
3. Set up reverse proxy (nginx)
4. Configure monitoring/alerting
5. Set up log aggregation

## License

[Add your license here]