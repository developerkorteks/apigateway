# 🚀 Dynamic API Sources Configuration Guide

This guide explains how to configure unlimited API sources dynamically without modifying code. Perfect for scaling to anime, drakor, film, donghua, manhwa, and any other content types!

## 🎯 Why Dynamic Configuration?

- **Unlimited Sources**: Add as many API sources as needed
- **No Code Changes**: Configure everything via environment variables
- **Multiple Content Types**: Support anime, drakor, film, donghua, etc.
- **Easy Scaling**: Add new sources instantly
- **Backward Compatible**: Existing configurations still work

## 🔧 Configuration Methods

### Method 1: JSON Configuration (Recommended for Many Sources)

Perfect when you have many sources or complex configurations:

```bash
# Single JSON string with all sources
API_SOURCES_JSON='{"gomunime":"http://localhost:8001","winbutv":"http://localhost:8002","samehadaku":"http://128.199.109.211:8182","drakorindo":"http://localhost:8003","filmapik":"http://localhost:8004","donghua_world":"http://localhost:8005","manhwa_club":"http://localhost:8006"}'
```

### Method 2: Individual Environment Variables (Easy to Manage)

Perfect for Docker and clear configuration management:

```bash
# Anime Sources
API_SOURCE_GOMUNIME_URL=http://localhost:8001
API_SOURCE_WINBUTV_URL=http://localhost:8002
API_SOURCE_SAMEHADAKU_URL=http://128.199.109.211:8182
API_SOURCE_OTAKUDESU_URL=https://otakudesu.quest
API_SOURCE_KUSONIME_URL=https://kusonime.com

# Drakor Sources
API_SOURCE_DRAKORINDO_URL=http://localhost:8003
API_SOURCE_KDRAMA_LAND_URL=http://localhost:8007
API_SOURCE_VIKI_SCRAPER_URL=http://localhost:8008

# Film Sources
API_SOURCE_FILMAPIK_URL=http://localhost:8004
API_SOURCE_LAYARKACA21_URL=http://localhost:8009
API_SOURCE_CINEMA21_URL=http://localhost:8010

# Donghua Sources
API_SOURCE_DONGHUA_WORLD_URL=http://localhost:8005
API_SOURCE_DONGHUA_STREAM_URL=http://localhost:8011

# Manhwa Sources
API_SOURCE_MANHWA_CLUB_URL=http://localhost:8006
API_SOURCE_WEBTOON_SCRAPER_URL=http://localhost:8012
```

### Method 3: Legacy Support (Backward Compatibility)

Existing configurations continue to work:

```bash
GOMUNIME_URL=http://localhost:8001
WINBUTV_URL=http://localhost:8002
SAMEHADAKU_URL=http://128.199.109.211:8182
# ... other legacy variables
```

## 🎬 Content Type Examples

### Anime Sources
```bash
API_SOURCE_GOMUNIME_URL=http://localhost:8001
API_SOURCE_WINBUTV_URL=http://localhost:8002
API_SOURCE_SAMEHADAKU_URL=http://128.199.109.211:8182
API_SOURCE_OTAKUDESU_URL=https://otakudesu.quest
API_SOURCE_KUSONIME_URL=https://kusonime.com
API_SOURCE_ANIMEINDO_URL=http://localhost:8013
API_SOURCE_OPLOVERZ_URL=http://localhost:8014
```

### Korean Drama (Drakor) Sources
```bash
API_SOURCE_DRAKORINDO_URL=http://localhost:8003
API_SOURCE_KDRAMA_LAND_URL=http://localhost:8007
API_SOURCE_VIKI_SCRAPER_URL=http://localhost:8008
API_SOURCE_DRAMACOOL_URL=http://localhost:8015
API_SOURCE_KISSASIAN_URL=http://localhost:8016
```

### Film/Movie Sources
```bash
API_SOURCE_FILMAPIK_URL=http://localhost:8004
API_SOURCE_LAYARKACA21_URL=http://localhost:8009
API_SOURCE_CINEMA21_URL=http://localhost:8010
API_SOURCE_INDOXXI_URL=http://localhost:8017
API_SOURCE_BIOSKOPKEREN_URL=http://localhost:8018
```

### Donghua (Chinese Animation) Sources
```bash
API_SOURCE_DONGHUA_WORLD_URL=http://localhost:8005
API_SOURCE_DONGHUA_STREAM_URL=http://localhost:8011
API_SOURCE_BILIBILI_SCRAPER_URL=http://localhost:8019
API_SOURCE_IQIYI_SCRAPER_URL=http://localhost:8020
```

### Manhwa/Webtoon Sources
```bash
API_SOURCE_MANHWA_CLUB_URL=http://localhost:8006
API_SOURCE_WEBTOON_SCRAPER_URL=http://localhost:8012
API_SOURCE_NAVER_WEBTOON_URL=http://localhost:8021
API_SOURCE_LEZHIN_SCRAPER_URL=http://localhost:8022
```

## 🐳 Docker Configuration

### docker-compose.yml
```yaml
services:
  apifallback:
    image: apifallback:latest
    environment:
      # Method 1: JSON (for many sources)
      API_SOURCES_JSON: '${API_SOURCES_JSON:-}'
      
      # Method 2: Individual variables
      API_SOURCE_GOMUNIME_URL: ${API_SOURCE_GOMUNIME_URL:-http://localhost:8001}
      API_SOURCE_WINBUTV_URL: ${API_SOURCE_WINBUTV_URL:-http://localhost:8002}
      API_SOURCE_DRAKORINDO_URL: ${API_SOURCE_DRAKORINDO_URL:-}
      API_SOURCE_FILMAPIK_URL: ${API_SOURCE_FILMAPIK_URL:-}
      # ... add more as needed
```

### .env file
```bash
# Choose your preferred method:

# Method 1: JSON Configuration
API_SOURCES_JSON={"gomunime":"http://localhost:8001","winbutv":"http://localhost:8002","drakorindo":"http://localhost:8003","filmapik":"http://localhost:8004"}

# OR Method 2: Individual Variables
API_SOURCE_GOMUNIME_URL=http://localhost:8001
API_SOURCE_WINBUTV_URL=http://localhost:8002
API_SOURCE_DRAKORINDO_URL=http://localhost:8003
API_SOURCE_FILMAPIK_URL=http://localhost:8004
```

## 🔄 How It Works

### 1. Configuration Loading Priority
1. **JSON Configuration** (`API_SOURCES_JSON`) - Highest priority
2. **Individual Variables** (`API_SOURCE_<NAME>_URL`) - Medium priority  
3. **Legacy Variables** (`GOMUNIME_URL`, etc.) - Lowest priority

### 2. Automatic Priority Assignment
The system automatically assigns priorities based on:
- **Source Type**: Known sources get optimized priorities
- **Endpoint Type**: Different endpoints prioritize different sources
- **Alphabetical Order**: Consistent fallback ordering

### 3. Smart Endpoint Mapping
- **Detail Endpoints** (`/anime-detail`, `/episode-detail`): Prioritize detailed data sources
- **List Endpoints** (`/home`, `/anime-terbaru`): Prioritize aggregation sources  
- **Search Endpoints**: Prioritize sources with good search capabilities

## 📊 Monitoring & Management

### Check Configured Sources
```bash
# View all configured sources
curl http://localhost:8080/health/sources

# Check specific endpoint sources
curl http://localhost:8080/api/v1/anime-terbaru?debug=true
```

### Add New Source Runtime
```bash
# Add new source via environment variable
export API_SOURCE_NEW_ANIME_SITE_URL=http://localhost:8025

# Restart application to pick up new source
docker-compose restart apifallback
```

## 🚀 Scaling Examples

### Small Setup (3-5 Sources)
```bash
API_SOURCE_GOMUNIME_URL=http://localhost:8001
API_SOURCE_WINBUTV_URL=http://localhost:8002
API_SOURCE_SAMEHADAKU_URL=http://128.199.109.211:8182
```

### Medium Setup (10-15 Sources)
```bash
# Use individual variables for clarity
API_SOURCE_GOMUNIME_URL=http://localhost:8001
API_SOURCE_WINBUTV_URL=http://localhost:8002
API_SOURCE_SAMEHADAKU_URL=http://128.199.109.211:8182
API_SOURCE_DRAKORINDO_URL=http://localhost:8003
API_SOURCE_FILMAPIK_URL=http://localhost:8004
API_SOURCE_DONGHUA_WORLD_URL=http://localhost:8005
# ... continue adding
```

### Large Setup (20+ Sources)
```bash
# Use JSON for easier management
API_SOURCES_JSON='{
  "gomunime":"http://localhost:8001",
  "winbutv":"http://localhost:8002",
  "samehadaku":"http://128.199.109.211:8182",
  "drakorindo":"http://localhost:8003",
  "filmapik":"http://localhost:8004",
  "donghua_world":"http://localhost:8005",
  "manhwa_club":"http://localhost:8006",
  "kdrama_land":"http://localhost:8007",
  "viki_scraper":"http://localhost:8008",
  "layarkaca21":"http://localhost:8009",
  "cinema21":"http://localhost:8010",
  "donghua_stream":"http://localhost:8011",
  "webtoon_scraper":"http://localhost:8012",
  "animeindo":"http://localhost:8013",
  "oploverz":"http://localhost:8014",
  "dramacool":"http://localhost:8015",
  "kissasian":"http://localhost:8016",
  "indoxxi":"http://localhost:8017",
  "bioskopkeren":"http://localhost:8018",
  "bilibili_scraper":"http://localhost:8019",
  "iqiyi_scraper":"http://localhost:8020"
}'
```

## 🔧 Best Practices

### 1. Naming Convention
- Use lowercase names: `gomunime`, `drakorindo`
- Use underscores for spaces: `donghua_world`, `manhwa_club`
- Be descriptive: `kdrama_land` not just `kdrama`

### 2. URL Management
- Use consistent ports: 8001-8099 for internal services
- Group by content type: 8001-8010 anime, 8011-8020 drakor, etc.
- Document your port assignments

### 3. Environment Management
```bash
# Development
API_SOURCE_GOMUNIME_URL=http://localhost:8001

# Staging  
API_SOURCE_GOMUNIME_URL=http://staging-gomunime:8001

# Production
API_SOURCE_GOMUNIME_URL=https://api.gomunime.com
```

## 🎯 Migration Guide

### From Hardcoded to Dynamic

**Before (Hardcoded):**
```go
// Had to modify code for each new source
WinbuTVURL: "http://localhost:8002"
```

**After (Dynamic):**
```bash
# Just add environment variable
API_SOURCE_NEW_SOURCE_URL=http://localhost:8025
```

### Backward Compatibility
All existing configurations continue to work:
```bash
# These still work
GOMUNIME_URL=http://localhost:8001
WINBUTV_URL=http://localhost:8002

# But these are preferred for new sources
API_SOURCE_NEW_ANIME_SITE_URL=http://localhost:8025
```

## 🎉 Benefits Achieved

✅ **Unlimited Scalability**: Add any number of sources  
✅ **Zero Code Changes**: Pure configuration-based  
✅ **Multi-Content Support**: Anime, drakor, film, donghua, manhwa  
✅ **Easy Management**: Clear environment variable structure  
✅ **Backward Compatible**: Existing setups continue working  
✅ **Smart Prioritization**: Automatic optimization per endpoint  
✅ **Production Ready**: Supports all deployment scenarios  

---

**Ready to scale infinitely!** 🚀 Add as many sources as you need without ever touching the code again!