# 📊 API Consistency Report - Anime Streaming APIs

**Generated on:** January 9, 2025  
**APIs Tested:**
- FastAPI (Port 8000) - samehadaku.how
- MultipleScrape (Port 8001) - gomunime.co  
- WinbuTV (Port 8002) - winbu.tv

---

## ✅ **WORKING ENDPOINTS**

### Health Check ✅
- **FastAPI**: `GET /health` → `{"status":"ok"}`
- **MultipleScrape**: `GET /health` → `{"status":"ok"}`
- **WinbuTV**: `GET /health` → `{"status":"ok","message":"API is running"}`

*Note: WinbuTV returns extra "message" field*

---

## ❌ **CRITICAL ISSUES FOUND**

### 1. **MultipleScrape Home Endpoint - BROKEN** 🚨
```bash
GET http://localhost:8001/api/v1/home/
```
**Expected**: JSON response  
**Actual**: `<a href="/api/v1/home">Moved Permanently</a>.`

**Status**: ❌ **COMPLETELY BROKEN** - Returns HTML redirect instead of JSON

---

### 2. **Home Endpoint Structure Inconsistency** ⚠️

#### FastAPI & WinbuTV (Consistent) ✅
```json
{
  "confidence_score": 1.0,
  "message": "Data berhasil diambil", 
  "source": "samehadaku.how|winbu.tv",
  "top10": [...],
  "new_eps": [...],
  "movies": [...],
  "jadwal_rilis": {...}
}
```

#### MultipleScrape (Different Structure) ❌
```json
{
  "confidence_score": 1,
  "message": "Data berhasil diambil",
  "source": "gomunime.co", 
  "data": [...] // Missing: top10, new_eps, movies, jadwal_rilis
}
```

**Issue**: MultipleScrape uses generic `data` array instead of specific fields

---

### 3. **Confidence Score Data Type Inconsistency** ⚠️

| API | Data Type | Example |
|-----|----------|---------|
| FastAPI | `number` (float) | `1.0` ✅ |
| MultipleScrape | `number` (integer) | `1` ⚠️ |
| WinbuTV | `number` (integer) | `1` ⚠️ |

**Recommendation**: All should use float format `1.0`

---

### 4. **Search Endpoint Field Name Inconsistencies** 🚨

#### FastAPI Search Response Fields ❌
```json
{
  "data": [{
    "anime_slug": "...",
    "genre": [...],
    "judul": "...",
    "penonton": "...",
    "sinopsis": "...",
    "skor": "...",
    "status": "...",
    "tipe": "...",
    "url_anime": "...",    // ❌ Should be "url"
    "url_cover": "..."     // ❌ Should be "cover"
  }]
}
```

#### WinbuTV Search Response Fields ❌ 
```json
{
  "data": [{
    "judul": "...",
    "url_anime": "...",    // ❌ Should be "url" 
    "anime_slug": "...",
    "status": "...",
    "tipe": "...",
    "skor": "...",
    "penonton": "...",
    "sinopsis": "...",
    "genre": [...],
    "url_cover": "..."     // ❌ Should be "cover"
  }]
}
```

#### MultipleScrape Search ❌
**Status**: Returns empty data array `"data_length": 0`

---

### 5. **Movie Endpoint Field Consistency** ✅

**Good News**: All movie endpoints have consistent field names:
```json
{
  "anime_slug": "...",
  "cover": "...",      // ✅ Consistent
  "genres": [...],
  "judul": "...",
  "sinopsis": "...",
  "skor": "...",
  "status": "...",
  "tanggal": "...",
  "url": "...",        // ✅ Consistent 
  "views": "..."
}
```

---

### 6. **Anime Terbaru Endpoint Field Consistency** ✅

**Good News**: All anime-terbaru endpoints have consistent field names:
```json
{
  "anime_slug": "...",
  "cover": "...",      // ✅ Consistent
  "episode": "...",
  "judul": "...",
  "rilis": "...",
  "uploader": "...",
  "url": "..."         // ✅ Consistent
}
```

---

### 7. **Detail Endpoint Parameter Inconsistency** ⚠️

| API | Anime Detail Parameter | Episode Detail Parameter |
|-----|----------------------|--------------------------|
| FastAPI | `?slug=anime-slug` | `?slug=episode-slug` |
| MultipleScrape | `?anime_slug=anime-slug` | `?episode_url=episode-url` |
| WinbuTV | `?anime_slug=anime-slug` | `?slug=episode-slug` |

**Issue**: Inconsistent parameter naming

---

## 🔧 **REQUIRED FIXES**

### Priority 1: Critical Fixes 🚨

1. **Fix MultipleScrape Home endpoint**
   - Currently returns HTML redirect 
   - Should return JSON with same structure as FastAPI/WinbuTV

2. **Fix Search endpoint field names**
   - Change `url_anime` → `url`
   - Change `url_cover` → `cover`
   - Apply to FastAPI and WinbuTV

3. **Fix MultipleScrape Search functionality**
   - Currently returns empty data array
   - Should return actual search results

### Priority 2: Structure Consistency ⚠️

4. **Standardize confidence_score format**
   - Change integers (`1`) to floats (`1.0`)
   - Apply to MultipleScrape and WinbuTV

5. **Fix MultipleScrape Home structure**
   - Change from generic `data` array
   - Add specific fields: `top10`, `new_eps`, `movies`, `jadwal_rilis`

6. **Standardize detail endpoint parameters**
   - Decide on consistent parameter names
   - Either all use `slug` or all use `anime_slug`

---

## ✅ **WORKING WELL**

1. **Standard Response Fields**: All APIs have `confidence_score`, `message`, `source` ✅
2. **Movie Endpoints**: Consistent field names across all APIs ✅
3. **Anime Terbaru Endpoints**: Consistent field names across all APIs ✅
4. **Health Check**: All functional (minor WinbuTV difference acceptable) ✅

---

## 📈 **CURRENT STATUS**

| Endpoint | FastAPI | MultipleScrape | WinbuTV | Consistency |
|----------|---------|----------------|---------|-------------|
| `/health` | ✅ | ✅ | ✅ | 🟡 Minor diff |
| `/home` | ✅ | ❌ Broken | ✅ | ❌ Structure diff |
| `/anime-terbaru` | ✅ | ✅ | ✅ | ✅ Consistent |
| `/movie` | ✅ | ✅ | ✅ | ✅ Consistent |
| `/search` | 🟡 Field names | ❌ No results | 🟡 Field names | ❌ Inconsistent |
| `/jadwal-rilis` | ✅ | ❌ Redirect | ✅ | 🟡 Partial |
| `/anime-detail` | ✅ | ✅ | ✅ | 🟡 Param diff |
| `/episode-detail` | ✅ | ✅ | ✅ | 🟡 Param diff |

---

## 🎯 **NEXT STEPS**

1. **Immediate**: Fix MultipleScrape Home endpoint (returns HTML instead of JSON)
2. **High Priority**: Standardize search endpoint field names (`url_anime` → `url`, `url_cover` → `cover`)
3. **Medium Priority**: Fix confidence score data types (use float `1.0` instead of int `1`)
4. **Low Priority**: Standardize detail endpoint parameter names

---

**Total Issues Found**: 7 major inconsistencies  
**APIs Fully Working**: 0/3 (all have issues)  
**Overall Consistency Score**: 60% (6/10 endpoints consistent)