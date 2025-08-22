#!/bin/bash

# Script to add standard endpoints for a new category
# Usage: ./add_endpoints_for_category.sh <category_name>

if [ -z "$1" ]; then
    echo "Usage: $0 <category_name>"
    echo "Example: $0 kdrama"
    exit 1
fi

CATEGORY_NAME="$1"
DB_PATH="./data.db"

# Check if category exists
CATEGORY_ID=$(sqlite3 "$DB_PATH" "SELECT id FROM categories WHERE name = '$CATEGORY_NAME' AND is_active = TRUE;")

if [ -z "$CATEGORY_ID" ]; then
    echo "Error: Category '$CATEGORY_NAME' not found or not active"
    echo "Please create the category first from the dashboard"
    exit 1
fi

echo "Adding standard endpoints for category: $CATEGORY_NAME (ID: $CATEGORY_ID)"

# Add standard endpoints
sqlite3 "$DB_PATH" "INSERT INTO endpoints (category_id, path) VALUES 
($CATEGORY_ID, '/api/v1/home'),
($CATEGORY_ID, '/api/v1/jadwal-rilis'),
($CATEGORY_ID, '/api/v1/anime-terbaru'),
($CATEGORY_ID, '/api/v1/movie'),
($CATEGORY_ID, '/api/v1/anime-detail'),
($CATEGORY_ID, '/api/v1/episode-detail'),
($CATEGORY_ID, '/api/v1/search');"

if [ $? -eq 0 ]; then
    echo "✅ Successfully added 7 endpoints for category '$CATEGORY_NAME'"
    
    # Show the added endpoints
    echo ""
    echo "Added endpoints:"
    sqlite3 "$DB_PATH" "SELECT e.id, e.path FROM endpoints e WHERE e.category_id = $CATEGORY_ID ORDER BY e.path;"
else
    echo "❌ Failed to add endpoints"
    exit 1
fi