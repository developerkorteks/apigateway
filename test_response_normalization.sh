#!/bin/bash

# Test script for response normalization
echo "=== Testing Response Normalization ==="

# Function to test a specific endpoint with sample data
test_normalization() {
    local endpoint="$1"
    local description="$2"
    
    echo ""
    echo "Testing: $description"
    echo "Endpoint: $endpoint"
    echo "----------------------------------------"
    
    # Make request to the API gateway
    response=$(curl -s "http://localhost:8080${endpoint}" | jq '.' 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "✅ Request successful"
        
        # Check if response has consistent structure
        data_structure=$(echo "$response" | jq '.data | type' 2>/dev/null)
        
        if [ "$data_structure" = '"object"' ]; then
            echo "✅ Data field is object (consistent)"
            
            # Check for key fields that should be present
            anime_slug=$(echo "$response" | jq '.data.anime_slug // empty' 2>/dev/null)
            confidence_score=$(echo "$response" | jq '.data.confidence_score // empty' 2>/dev/null)
            source=$(echo "$response" | jq '.data.source // empty' 2>/dev/null)
            
            echo "  - anime_slug: $anime_slug"
            echo "  - confidence_score: $confidence_score"
            echo "  - source: $source"
            
            # Check for nested data.data (should not exist after normalization)
            nested_data=$(echo "$response" | jq '.data.data // empty' 2>/dev/null)
            if [ -n "$nested_data" ] && [ "$nested_data" != "null" ]; then
                echo "❌ WARNING: Found nested data.data structure - normalization may have failed"
            else
                echo "✅ No nested data.data found - normalization working"
            fi
        else
            echo "❌ Data field is not object or missing"
        fi
    else
        echo "❌ Request failed or invalid JSON response"
    fi
}

# Check if API is running
echo "Checking if API Gateway is running..."
if ! curl -s http://localhost:8080/health >/dev/null 2>&1; then
    echo "❌ API Gateway is not running on localhost:8080"
    echo "Please start the service with: make run"
    exit 1
fi

echo "✅ API Gateway is running"

# Test different endpoints that might have different response structures
test_normalization "/api/v1/home" "Home endpoint aggregation"
test_normalization "/api/v1/anime-terbaru" "Latest anime list"
test_normalization "/api/v1/anime-detail?anime_slug=example" "Anime detail (if available)"

# Test with different categories to see different sources
test_normalization "/api/v1/home?category=anime" "Home with anime category"

echo ""
echo "=== Test Summary ==="
echo "Check the above results to verify:"
echo "1. No nested data.data structures appear"
echo "2. All responses have consistent data object structure"  
echo "3. confidence_score and source fields are present"
echo "4. Response structures are normalized across all sources"
echo ""
echo "If you see any nested data.data structures, the normalization"
echo "may need adjustment for specific API sources."