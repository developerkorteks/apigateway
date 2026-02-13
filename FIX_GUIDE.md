# 🔧 PERMANENT FIX GUIDE - API Source Configuration

## 🎯 ROOT CAUSE IDENTIFIED

### Problem:
**`.env` file tidak memiliki API source URLs!**

Akibatnya, sistem menggunakan **hardcoded defaults** di `pkg/config/config.go` line 143-150:

```go
if len(sources) == 0 {
    sources = map[string]string{
        "winbutv": "http://localhost:8002",  // ❌ LOCAL DEV URL
        // ...
    }
}
```

### Why This Happens:
1. `.env` hanya punya PORT, DATABASE_PATH, REDIS_ADDR
2. **Tidak ada** `WINBUTV_URL` atau `API_SOURCES_JSON`
3. `loadAPISources()` tidak menemukan konfigurasi
4. Fallback ke hardcoded localhost URLs (untuk development)

---

## ✅ SOLUTION

### Option 1: Add to .env (RECOMMENDED)

```bash
# Add this to your .env file:
WINBUTV_URL=https://winbu-tv.humanmade.my.id
```

### Option 2: Use JSON Format (for multiple sources)

```bash
# Add this to .env:
API_SOURCES_JSON={"winbutv":"https://winbu-tv.humanmade.my.id","gomunime":"https://api.gomunime.com"}
```

### Option 3: Individual Environment Variables

```bash
# Add multiple sources:
WINBUTV_URL=https://winbu-tv.humanmade.my.id
GOMUNIME_URL=https://api.gomunime.com
SAMEHADAKU_URL=https://api.samehadaku.tv
```

---

## 📋 STEPS TO FIX

### Step 1: Update .env file

```bash
# Backup current .env
cp .env .env.backup

# Add API source URLs
echo "" >> .env
echo "# API Source URLs (Production)" >> .env
echo "WINBUTV_URL=https://winbu-tv.humanmade.my.id" >> .env
```

### Step 2: Delete old database (optional but recommended)

```bash
# Backup current database
cp data.db data.db.backup

# Delete to force recreation with correct URLs
rm data.db
```

### Step 3: Restart application

```bash
# Rebuild (if needed)
go build -o main cmd/main.go

# Restart
./main
```

### Step 4: Verify

```bash
# Check database has correct URLs
sqlite3 data.db "SELECT DISTINCT source_name, base_url FROM api_sources;"

# Test endpoint
curl "http://localhost:58080/api/v1/anime-detail?anime_slug=test&category=anime"
```

---

## 🔒 PREVENT FUTURE ISSUES

### 1. Update .env.example

```bash
cat >> .env.example << 'EXAMPLE'

# API Source Configuration
# IMPORTANT: Set production URLs, NOT localhost!
WINBUTV_URL=https://winbu-tv.humanmade.my.id
# GOMUNIME_URL=https://your-api.com
# Or use JSON format:
# API_SOURCES_JSON={"winbutv":"https://winbu-tv.humanmade.my.id"}
EXAMPLE
```

### 2. Add Validation in Code

Add to `pkg/config/config.go` after line 151:

```go
// Warn if using localhost URLs in production
for name, url := range sources {
    if strings.Contains(url, "localhost") && os.Getenv("ENV") == "production" {
        log.Printf("WARNING: Source '%s' using localhost URL in production: %s", name, url)
    }
}
```

### 3. Documentation

Update README.md:

```markdown
## Configuration

**IMPORTANT:** Before running, configure your API sources in `.env`:

```bash
# Required: Set production API URLs
WINBUTV_URL=https://winbu-tv.humanmade.my.id
```

Without this, the system will use localhost URLs (development mode).
```

---

## 📊 SUMMARY

| Issue | Cause | Fix |
|-------|-------|-----|
| Wrong base_url | Missing env vars | Add `WINBUTV_URL` to .env |
| Localhost URLs | Hardcoded defaults | Configure production URLs |
| Fresh DB issues | Auto-uses defaults | Always set env before first run |

---

## ✅ VERIFICATION CHECKLIST

- [ ] `.env` has `WINBUTV_URL=https://winbu-tv.humanmade.my.id`
- [ ] Delete old `data.db` 
- [ ] Restart application
- [ ] Check database: `SELECT DISTINCT base_url FROM api_sources;`
- [ ] Test endpoint: returns 200 OK
- [ ] No localhost URLs in production database

---

**File Created:** 2026-02-13  
**Issue:** Fresh database gets wrong URLs  
**Root Cause:** Missing environment variables  
**Solution:** Configure .env with production URLs
