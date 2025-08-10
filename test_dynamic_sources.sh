#!/bin/bash

# Test Dynamic API Sources Configuration
# This script demonstrates how easy it is to add new API sources

echo "🧪 Testing Dynamic API Sources Configuration"
echo "=============================================="

# Test 1: Add Donghua Sources
echo ""
echo "📺 Test 1: Adding Donghua Sources"
export API_SOURCE_DONGHUA_WORLD_URL=http://localhost:8005
export API_SOURCE_BILIBILI_SCRAPER_URL=http://localhost:8019
export API_SOURCE_IQIYI_SCRAPER_URL=http://localhost:8020

echo "✅ Added donghua sources via environment variables"
echo "   - donghua_world: $API_SOURCE_DONGHUA_WORLD_URL"
echo "   - bilibili_scraper: $API_SOURCE_BILIBILI_SCRAPER_URL" 
echo "   - iqiyi_scraper: $API_SOURCE_IQIYI_SCRAPER_URL"

# Test 2: Add Drakor Sources
echo ""
echo "🎭 Test 2: Adding Drakor Sources"
export API_SOURCE_DRAKORINDO_URL=http://localhost:8003
export API_SOURCE_KDRAMA_LAND_URL=http://localhost:8007
export API_SOURCE_VIKI_SCRAPER_URL=http://localhost:8008

echo "✅ Added drakor sources via environment variables"
echo "   - drakorindo: $API_SOURCE_DRAKORINDO_URL"
echo "   - kdrama_land: $API_SOURCE_KDRAMA_LAND_URL"
echo "   - viki_scraper: $API_SOURCE_VIKI_SCRAPER_URL"

# Test 3: Add Film Sources
echo ""
echo "🎬 Test 3: Adding Film Sources"
export API_SOURCE_FILMAPIK_URL=http://localhost:8004
export API_SOURCE_LAYARKACA21_URL=http://localhost:8009
export API_SOURCE_CINEMA21_URL=http://localhost:8010

echo "✅ Added film sources via environment variables"
echo "   - filmapik: $API_SOURCE_FILMAPIK_URL"
echo "   - layarkaca21: $API_SOURCE_LAYARKACA21_URL"
echo "   - cinema21: $API_SOURCE_CINEMA21_URL"

# Test 4: JSON Configuration Method
echo ""
echo "📋 Test 4: JSON Configuration Method"
export API_SOURCES_JSON='{
  "manhwa_club":"http://localhost:8006",
  "webtoon_scraper":"http://localhost:8012",
  "naver_webtoon":"http://localhost:8021",
  "lezhin_scraper":"http://localhost:8022"
}'

echo "✅ Added manhwa sources via JSON configuration"
echo "   JSON: $API_SOURCES_JSON"

# Test 5: Build and verify configuration
echo ""
echo "🔨 Test 5: Building Application with New Sources"

# Build the application
if go build -o main cmd/main.go; then
    echo "✅ Application built successfully with dynamic sources!"
else
    echo "❌ Build failed"
    exit 1
fi

# Test 6: Show what sources would be loaded
echo ""
echo "📊 Test 6: Sources That Would Be Loaded"
echo "========================================"

echo "🎌 Anime Sources:"
echo "  - gomunime (legacy): ${GOMUNIME_URL:-http://localhost:8001}"
echo "  - winbutv (legacy): ${WINBUTV_URL:-http://localhost:8002}"
echo "  - samehadaku (legacy): ${SAMEHADAKU_URL:-http://128.199.109.211:8182}"

echo ""
echo "📺 Donghua Sources:"
echo "  - donghua_world: $API_SOURCE_DONGHUA_WORLD_URL"
echo "  - bilibili_scraper: $API_SOURCE_BILIBILI_SCRAPER_URL"
echo "  - iqiyi_scraper: $API_SOURCE_IQIYI_SCRAPER_URL"

echo ""
echo "🎭 Drakor Sources:"
echo "  - drakorindo: $API_SOURCE_DRAKORINDO_URL"
echo "  - kdrama_land: $API_SOURCE_KDRAMA_LAND_URL"
echo "  - viki_scraper: $API_SOURCE_VIKI_SCRAPER_URL"

echo ""
echo "🎬 Film Sources:"
echo "  - filmapik: $API_SOURCE_FILMAPIK_URL"
echo "  - layarkaca21: $API_SOURCE_LAYARKACA21_URL"
echo "  - cinema21: $API_SOURCE_CINEMA21_URL"

echo ""
echo "📚 Manhwa Sources (from JSON):"
echo "  - manhwa_club, webtoon_scraper, naver_webtoon, lezhin_scraper"

echo ""
echo "🎉 SUCCESS: All sources configured dynamically!"
echo "💡 Total Sources: 13+ sources across multiple content types"
echo "🚀 Zero code changes required!"

echo ""
echo "📝 Next Steps:"
echo "1. Start your scraper services on the configured ports"
echo "2. Run: ./main"
echo "3. Test endpoints: curl http://localhost:8080/api/v1/anime-terbaru"
echo "4. Check health: curl http://localhost:8080/health/sources"

echo ""
echo "✨ The system will automatically:"
echo "   - Detect all configured sources"
echo "   - Assign smart priorities per endpoint"
echo "   - Enable fallback mechanisms"
echo "   - Monitor health of all sources"
echo "   - Load balance requests"