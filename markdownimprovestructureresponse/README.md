# API Fallback System - Anime Streaming APIs

Sistem API fallback yang terdiri dari 3 API berbeda untuk streaming anime dengan struktur response yang konsisten.

## 🚀 API Endpoints

### 1. FastAPI (Port 8000) - Samehadaku.how
- **Base URL**: `http://localhost:8000/api/v1`
- **Source**: `samehadaku.how`
- **Framework**: FastAPI (Python)

### 2. MultipleScrape (Port 8001) - Gomunime.co  
- **Base URL**: `http://localhost:8001/api/v1`
- **Source**: `gomunime.co`
- **Framework**: Gin (Go)

### 3. WinbuTV (Port 8002) - Winbu.tv
- **Base URL**: `http://localhost:8002/api/v1`
- **Source**: `winbu.tv`
- **Framework**: Gin (Go)

## 📋 Available Endpoints

Semua API memiliki endpoint yang sama dengan struktur response yang konsisten:

### 🏠 Home Data
```
GET /home
```

### 📺 Anime Terbaru
```
GET /anime-terbaru
```

### 🎬 Movie
```
GET /movie
```

### 📅 Jadwal Rilis
```
GET /jadwal-rilis
```

### 🔍 Search
```
GET /search?query={search_term}
```

### 📖 Anime Detail
```
GET /anime-detail?slug={anime_slug}
```

### 🎥 Episode Detail
```
GET /episode-detail?slug={episode_slug}
```

## 🔄 Consistent Response Structure

Semua API mengembalikan response dengan struktur yang konsisten:

### Standard Response Format
```json
{
  "confidence_score": 1.0,
  "message": "Data berhasil diambil",
  "source": "samehadaku.how|gomunime.co|winbu.tv",
  "data": [...] // atau struktur data spesifik
}
```

### Home Endpoint Response
```json
{
  "confidence_score": 1.0,
  "message": "Data berhasil diambil",
  "source": "samehadaku.how",
  "top10": [...],
  "new_eps": [...],
  "movies": [...],
  "jadwal_rilis": {...}
}
```

### List Endpoints Response (anime-terbaru, movie, search)
```json
{
  "confidence_score": 1.0,
  "message": "Data berhasil diambil",
  "source": "samehadaku.how",
  "data": [...]
}
```

### Detail Endpoints Response (anime-detail, episode-detail)
```json
{
  "confidence_score": 1.0,
  "message": "Data berhasil diambil",
  "source": "samehadaku.how",
  // ... detail fields
}
```

## 🎯 Confidence Score

- **1.0**: Data lengkap dan valid
- **0.8-0.99**: Data sebagian valid
- **0.0**: Data tidak valid atau error

## 🔧 Features

### ✅ Consistent Structure
- Semua API menggunakan struktur response yang sama
- Field `confidence_score`, `message`, dan `source` ada di semua response
- Format data yang seragam antar API

### ✅ Data Validation
- Validasi URL dan cover image
- Validasi slug format
- Validasi title format
- Auto-fill field opsional dengan data dummy

### ✅ Error Handling
- Graceful error handling
- Fallback ke API lain jika satu API down
- Logging yang komprehensif

### ✅ Caching
- Redis caching untuk performa optimal
- Cache invalidation dengan parameter `force_refresh`
- TTL yang dapat dikonfigurasi

## 🚦 Testing API Consistency

### Test All Home Endpoints
```bash
echo "=== Testing Home Endpoints ==="
echo "1. FastAPI:" && curl -s "http://localhost:8000/api/v1/home/" | jq '.confidence_score, .message, .source'
echo "2. MultipleScrape:" && curl -s "http://localhost:8001/api/v1/home/" | jq '.confidence_score, .message, .source'  
echo "3. WinbuTV:" && curl -s "http://localhost:8002/api/v1/home" | jq '.confidence_score, .message, .source'
```

### Test All Search Endpoints
```bash
echo "=== Testing Search Endpoints ==="
echo "1. FastAPI:" && curl -s "http://localhost:8000/api/v1/search/?query=naruto" | jq '.confidence_score, .message, .source, (.data | length)'
echo "2. MultipleScrape:" && curl -s "http://localhost:8001/api/v1/search/?query=naruto" | jq '.confidence_score, .message, .source, (.data | length)'
echo "3. WinbuTV:" && curl -s "http://localhost:8002/api/v1/search?query=naruto" | jq '.confidence_score, .message, .source, (.data | length)'
```

## 🏃‍♂️ Running the APIs

### FastAPI (Port 8000)
```bash
cd fastapi_app
source venv/bin/activate
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

### MultipleScrape (Port 8001)
```bash
cd multiplescrape
PORT=8001 go run .
```

### WinbuTV (Port 8002)
```bash
cd winbutv
PORT=8002 go run .
```

## 📊 API Status

| API | Status | Source | Framework | Port |
|-----|--------|--------|-----------|------|
| FastAPI | ✅ Active | samehadaku.how | FastAPI | 8000 |
| MultipleScrape | ✅ Active | gomunime.co | Gin | 8001 |
| WinbuTV | ✅ Active | winbu.tv | Gin | 8002 |

## 🔍 Validation Rules

### URL Validation
- Must start with `https://`
- Valid domain format
- No "N/A" or "-" values

### Image URL Validation
- Must be valid URL
- Must end with `.jpg`, `.jpeg`, `.png`, or `.webp`

### Slug Validation
- Lowercase letters, numbers, and hyphens only
- Format: `^[a-z0-9]+(-[a-z0-9]+)*$`

### Title Validation
- Minimum 2 characters, maximum 200 characters
- Must have at least one capitalized word
- No excessive punctuation (max 3 consecutive)
- No HTML tags

## 🎉 Success Metrics

✅ **All APIs return consistent JSON structure**  
✅ **All endpoints have `confidence_score`, `message`, and `source` fields**  
✅ **Data validation implemented across all APIs**  
✅ **Error handling and fallback mechanisms working**  
✅ **Caching system operational**  
✅ **Search functionality fixed and working**  

---

**Last Updated**: January 2025  
**Status**: All APIs Consistent ✅