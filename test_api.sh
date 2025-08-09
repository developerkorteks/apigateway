#!/bin/bash

# API Testing Script
BASE_URL="http://localhost:8080"

echo "=== API Fallback System Testing ==="
echo "Base URL: $BASE_URL"
echo ""

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq '.' || echo "Health check failed"
echo ""

# Test dashboard
echo "2. Testing dashboard..."
curl -s -I "$BASE_URL/dashboard/" | head -1
echo ""

# Test API endpoints
echo "3. Testing API endpoints..."

endpoints=(
    "/api/v1/home"
    "/api/v1/jadwal-rilis"
    "/api/v1/anime-terbaru"
    "/api/v1/movie"
    "/api/v1/search?q=naruto"
)

for endpoint in "${endpoints[@]}"; do
    echo "Testing: $endpoint"
    response=$(curl -s -w "HTTP_CODE:%{http_code};TIME:%{time_total}" "$BASE_URL$endpoint")
    
    # Extract HTTP code and time
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    time_total=$(echo "$response" | grep -o "TIME:[0-9.]*" | cut -d: -f2)
    
    # Remove the status info from response
    json_response=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*;TIME:[0-9.]*//')
    
    echo "  Status: $http_code"
    echo "  Time: ${time_total}s"
    
    if [ "$http_code" = "200" ]; then
        echo "  ✓ Success"
        # Try to parse JSON and show basic info
        if command -v jq &> /dev/null; then
            source=$(echo "$json_response" | jq -r '.source // "unknown"' 2>/dev/null)
            confidence=$(echo "$json_response" | jq -r '.confidence_score // "unknown"' 2>/dev/null)
            echo "  Source: $source"
            echo "  Confidence: $confidence"
        fi
    else
        echo "  ✗ Failed"
        echo "  Response: $json_response"
    fi
    echo ""
done

# Test dashboard API endpoints
echo "4. Testing dashboard API endpoints..."

dashboard_endpoints=(
    "/dashboard/health"
    "/dashboard/logs"
    "/dashboard/stats"
    "/dashboard/categories"
)

for endpoint in "${dashboard_endpoints[@]}"; do
    echo "Testing: $endpoint"
    response=$(curl -s -w "HTTP_CODE:%{http_code}" "$BASE_URL$endpoint")
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    json_response=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*//')
    
    echo "  Status: $http_code"
    if [ "$http_code" = "200" ]; then
        echo "  ✓ Success"
    else
        echo "  ✗ Failed"
    fi
    echo ""
done

echo "=== Testing Complete ==="