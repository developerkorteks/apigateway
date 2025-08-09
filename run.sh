#!/bin/bash

# API Fallback System Runner
echo "Starting API Fallback System..."

# Check if binary exists
if [ ! -f "./apifallback" ]; then
    echo "Building application..."
    go build -o apifallback cmd/main.go
    if [ $? -ne 0 ]; then
        echo "Build failed!"
        exit 1
    fi
fi

# Set default environment variables if not set
export PORT=${PORT:-8080}
export DATABASE_PATH=${DATABASE_PATH:-./data.db}
export REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
export REDIS_DB=${REDIS_DB:-0}
export API_TIMEOUT=${API_TIMEOUT:-20s}
export RATE_LIMIT=${RATE_LIMIT:-100}
export HEALTH_CHECK_INTERVAL=${HEALTH_CHECK_INTERVAL:-10m}

echo "Configuration:"
echo "  Port: $PORT"
echo "  Database: $DATABASE_PATH"
echo "  Redis: $REDIS_ADDR"
echo "  API Timeout: $API_TIMEOUT"
echo "  Rate Limit: $RATE_LIMIT requests/minute"
echo ""

echo "Starting server on port $PORT..."
echo "Dashboard available at: http://localhost:$PORT/dashboard/"
echo "API endpoints available at: http://localhost:$PORT/api/v1/"
echo ""

# Run the application
./apifallback