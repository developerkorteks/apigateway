#!/bin/bash

echo "🚀 Starting Complete API Fallback System with Bruteforce Implementation"
echo "=================================================================="

# Colors for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Kill any existing processes on the ports we'll use
echo -e "${YELLOW}🔧 Cleaning up existing processes...${NC}"
pkill -f "winbu.tv"
pkill -f "apicategorywithfallback"
pkill -f ":8080"
pkill -f ":8082"
sleep 2

# Function to check if a port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        return 1
    else
        return 0
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    local name=$1
    local url=$2
    local max_attempts=30
    local attempt=1
    
    echo -e "${BLUE}⏳ Waiting for $name to be ready...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $name is ready!${NC}"
            return 0
        fi
        
        printf "."
        sleep 1
        ((attempt++))
    done
    
    echo -e "${RED}❌ $name failed to start within $max_attempts seconds${NC}"
    return 1
}

# Build applications
echo -e "${YELLOW}🔨 Building applications...${NC}"

# Build main API fallback system
echo "Building main API fallback system..."
cd /home/korteks/Documents/project/apifallback
go mod tidy
if ! go build -o apifallback cmd/main.go; then
    echo -e "${RED}❌ Failed to build main API fallback system${NC}"
    exit 1
fi

# Build winbutv API
echo "Building WinbuTV API..."
cd winbutv
go mod tidy  
if ! go build -o winbutv main.go; then
    echo -e "${RED}❌ Failed to build WinbuTV API${NC}"
    exit 1
fi

# Build test program
echo "Building test program..."
cd ..
if ! go build -o test_bruteforce test_bruteforce.go; then
    echo -e "${RED}❌ Failed to build test program${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All builds completed successfully!${NC}"

# Start WinbuTV API server
echo -e "${YELLOW}🌐 Starting WinbuTV API Server (Port 8082)...${NC}"
cd winbutv
PORT=8082 ./winbutv &
WINBUTV_PID=$!
cd ..

# Wait for WinbuTV to be ready
if ! wait_for_service "WinbuTV API" "http://localhost:8082/health"; then
    kill $WINBUTV_PID 2>/dev/null
    exit 1
fi

# Start main API fallback system
echo -e "${YELLOW}🎯 Starting Main API Fallback System (Port 8080)...${NC}"
PORT=8080 ./apifallback &
MAIN_PID=$!

# Wait for main API to be ready
if ! wait_for_service "Main API Fallback System" "http://localhost:8080/health"; then
    kill $WINBUTV_PID $MAIN_PID 2>/dev/null
    exit 1
fi

echo -e "${GREEN}🎉 All services are now running!${NC}"
echo ""
echo -e "${BLUE}📊 Service Status:${NC}"
echo "├── Main API Fallback System: http://localhost:8080"
echo "├── WinbuTV API Server: http://localhost:8082"
echo "└── Swagger Documentation: http://localhost:8082/swagger/index.html"
echo ""

echo -e "${YELLOW}🧪 Running Bruteforce Tests...${NC}"
echo "This will test the new bruteforce implementation for anime detail endpoints"
echo ""

# Run tests
sleep 2
./test_bruteforce

echo ""
echo -e "${GREEN}🎊 Bruteforce Implementation Testing Complete!${NC}"
echo ""
echo -e "${BLUE}🔍 Key Features Implemented:${NC}"
echo "✅ Parallel bruteforce to ALL API sources (primary + fallback)"
echo "✅ Data validation before sending response"
echo "✅ Priority-based source selection"
echo "✅ Automatic fallback URL handling"
echo "✅ Comprehensive error handling and logging"
echo "✅ Response caching for valid results"
echo "✅ Dynamic timeout based on source count"
echo ""

echo -e "${YELLOW}📝 Manual Testing URLs:${NC}"
echo "Test anime detail bruteforce:"
echo "curl \"http://localhost:8080/api/v1/anime-detail?anime_slug=naruto&category=anime\""
echo ""
echo "curl \"http://localhost:8080/api/v1/anime-detail?anime_slug=one-piece&category=anime\""
echo ""
echo "Test episode detail bruteforce:"  
echo "curl \"http://localhost:8080/api/v1/episode-detail?episode_url=https://winbu.tv/anime/naruto-episode-1&category=anime\""
echo ""

echo -e "${BLUE}🛑 To stop all services, press Ctrl+C or run:${NC}"
echo "pkill -f winbutv && pkill -f apicategorywithfallback"

# Keep services running
echo ""
echo "Services are running... Press Ctrl+C to stop all services."
trap 'echo -e "\n${YELLOW}🛑 Stopping all services...${NC}"; kill $WINBUTV_PID $MAIN_PID 2>/dev/null; exit 0' INT

# Wait for user to stop
wait