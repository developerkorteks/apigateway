# Winbu.TV Web Scraping API

API web scraping untuk mengambil data dari situs Winbu.TV menggunakan Go dan Gin framework. API ini menyediakan endpoint untuk mengakses data anime, film, dan jadwal rilis dengan format JSON yang konsisten.

## 🚀 Fitur Utama

- **Homepage Data**: Top 10 anime, episode terbaru, film terbaru
- **Anime Terbaru**: Daftar anime terbaru dengan pagination
- **Film**: Daftar film dengan pagination
- **Jadwal Rilis**: Jadwal rilis anime per hari
- **Confidence Score**: Setiap response memiliki confidence score (0.0-1.0)
- **Error Handling**: Robust error handling dengan HTTP status codes
- **Rate Limiting**: Menghormati situs target dengan delay antar request

## 📋 Struktur Proyek

```
winbutv/
├── main.go                 # Entry point aplikasi
├── config/
│   └── config.go          # Konfigurasi aplikasi
├── scrapers/              # Logic scraping per kategori
│   ├── home_scraper.go    # Homepage scraper
│   ├── anime_scraper.go   # Anime terbaru scraper
│   ├── movie_scraper.go   # Movie scraper
│   └── schedule_scraper.go # Jadwal rilis scraper
├── models/                # Data models
│   └── response_models.go # Response structures
├── api/                   # API handlers
│   └── v1/
│       └── endpoints.go   # Main endpoints
├── utils/                 # Utilities
│   ├── http_client.go     # HTTP client wrapper
│   └── helpers.go         # Helper functions
├── scrape/                # Test files untuk scraping
└── go.mod
```

## 🛠️ Instalasi dan Menjalankan

### Prerequisites
- Go 1.24.4 atau lebih baru
- Internet connection untuk scraping

### Instalasi
```bash
# Clone repository
git clone <repository-url>
cd winbutv

# Download dependencies
go mod tidy

# Build aplikasi
go build -o api-server main.go

# Jalankan server
./api-server
```

### Konfigurasi Environment Variables
```bash
export PORT=8080                    # Port server (default: 8080)
export ENVIRONMENT=development      # Environment mode
export BASE_URL=https://winbu.tv   # Base URL target
export TIMEOUT=30s                 # Request timeout
export RATE_LIMIT=1s               # Delay antar request
export MAX_RETRIES=3               # Maximum retry attempts
export CACHE_ENABLED=true          # Enable caching
export CACHE_TTL=5m                # Cache TTL
```

## 📚 API Endpoints

### Health Check
```
GET /health
```
**Response:**
```json
{
  "status": "ok",
  "message": "API is running"
}
```

### Homepage Data
```
GET /api/v1/home
```
**Response:**
```json
{
  "confidence_score": 1.0,
  "message": "Data berhasil diambil",
  "source": "winbu.tv",
  "top10": [
    {
      "judul": "One Piece",
      "url": "https://winbu.tv/anime/one-piece/",
      "anime_slug": "one-piece",
      "rating": "8.71",
      "cover": "https://winbu.tv/wp-content/uploads/2020/04/E5RxYkWX0AAwdGH.png.jpg",
      "genres": ["Anime"]
    }
  ],
  "new_eps": [
    {
      "judul": "Dandadan Season 2",
      "url": "https://winbu.tv/anime/dandadan-season-2/",
      "anime_slug": "dandadan-season-2",
      "episode": "Episode 6",
      "rilis": "10 jam",
      "cover": "https://winbu.tv/wp-content/uploads/2025/07/dandan.jpg"
    }
  ],
  "movies": [
    {
      "judul": "Overlord Movie 3",
      "url": "https://winbu.tv/anime/overlord-movie-3/",
      "anime_slug": "overlord-movie-3",
      "tanggal": "Jun 4, 2021",
      "cover": "https://winbu.tv/wp-content/uploads/2025/04/144101.jpg",
      "genres": ["Action", "Adventure"]
    }
  ],
  "jadwal_rilis": {
    "Monday": [],
    "Tuesday": [],
    "Wednesday": [],
    "Thursday": [],
    "Friday": [],
    "Saturday": [],
    "Sunday": []
  }
}
```

### Anime Terbaru
```
GET /api/v1/anime-terbaru?page=1
```
**Parameters:**
- `page` (optional): Nomor halaman (default: 1)

**Response:**
```json
{
  "confidence_score": 1.0,
  "message": "Data berhasil diambil",
  "source": "winbu.tv",
  "data": [
    {
      "judul": "Dandadan Season 2",
      "url": "https://winbu.tv/anime/dandadan-season-2/",
      "anime_slug": "dandadan-season-2",
      "episode": "Episode 6",
      "uploader": "Unknown",
      "rilis": "10 jam",
      "cover": "https://winbu.tv/wp-content/uploads/2025/07/dandan.jpg"
    }
  ]
}
```

### Film
```
GET /api/v1/movie?page=1
```
**Parameters:**
- `page` (optional): Nomor halaman (default: 1)

**Response:**
```json
{
  "confidence_score": 0.9,
  "message": "Data berhasil diambil",
  "source": "winbu.tv",
  "data": [
    {
      "judul": "Overlord Movie 3",
      "url": "https://winbu.tv/anime/overlord-movie-3/",
      "anime_slug": "overlord-movie-3",
      "status": "Completed",
      "skor": "7.80",
      "sinopsis": "Kerajaan Suci telah damai bertahun-tahun...",
      "views": "427583 Views",
      "cover": "https://winbu.tv/wp-content/uploads/2025/04/144101.jpg",
      "genres": ["Action", "Adventure", "Fantasy"],
      "tanggal": "N/A"
    }
  ]
}
```

### Jadwal Rilis
```
GET /api/v1/jadwal-rilis
```
**Response:**
```json
{
  "confidence_score": 0.1,
  "message": "No schedule data found",
  "source": "winbu.tv",
  "Monday": [],
  "Tuesday": [],
  "Wednesday": [],
  "Thursday": [],
  "Friday": [],
  "Saturday": [],
  "Sunday": []
}
```

### Search
```
GET /api/v1/search?q=<string>&page=<int>
```
**Parameters:**
- `q` (required): Query string untuk pencarian
- `page` (optional): Nomor halaman (default: 1)

**Response:**
```json
{
  "confidence_score": 0.8,
  "message": "Data berhasil diambil",
  "source": "winbu.tv",
  "query": "one piece",
  "page": 1,
  "data": [
    {
      "judul": "One Piece",
      "url": "https://winbu.tv/anime/one-piece/",
      "anime_slug": "one-piece",
      "cover": "https://winbu.tv/wp-content/uploads/2020/04/E5RxYkWX0AAwdGH.png.jpg",
      "episode": "Episode 1120",
      "rating": "8.71",
      "type": "Anime"
    }
  ]
}
```

## 🔧 Confidence Score

Setiap response API menyertakan `confidence_score` (0.0-1.0):
- **1.0**: Data lengkap dan akurat
- **0.5-0.9**: Data mungkin tidak lengkap atau ada anomali kecil
- **< 0.5**: Data sangat tidak lengkap atau ada masalah signifikan

## 🧪 Testing

Menjalankan test scraping yang ada:
```bash
# Test homepage scraping
go test -v ./scrape -run TestHome

# Test anime terbaru scraping
go test -v ./scrape -run TestScrapeAnimeTerbaruAnimasuLimited

# Test film scraping
go test -v ./scrape -run TestScrapeFilmLimited
```

## 📝 Error Handling

API menggunakan HTTP status codes standar:
- `200 OK`: Request berhasil
- `404 Not Found`: Data tidak ditemukan
- `500 Internal Server Error`: Error server
- `503 Service Unavailable`: Situs target tidak dapat dijangkau

Format error response:
```json
{
  "error": true,
  "message": "Failed to scrape data: Timeout",
  "confidence_score": 0.0
}
```

## 🚦 Rate Limiting

API menerapkan rate limiting untuk menghormati situs target:
- Default delay: 1 detik antar request
- Maximum retries: 3 kali
- Timeout: 30 detik per request

## 🔄 Caching

API mendukung caching untuk mengurangi beban pada situs target:
- Cache enabled by default
- Default TTL: 5 menit
- Dapat dikonfigurasi via environment variables

## 🤝 Contributing

1. Fork repository
2. Buat feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buat Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

## 📞 Contact

Project Link: [https://github.com/nabilulilalbab/winbu.tv](https://github.com/nabilulilalbab/winbu.tv)