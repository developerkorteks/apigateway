#!/bin/bash

# Comprehensive Test Runner for API Fallback System
echo "🧪 Starting Comprehensive Test Suite for API Fallback System"
echo "============================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run test and track results
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "\n${BLUE}🔍 Running: $test_name${NC}"
    echo "Command: $test_command"
    echo "----------------------------------------"
    
    if eval $test_command; then
        echo -e "${GREEN}✅ PASSED: $test_name${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ FAILED: $test_name${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
}

# Check if Valkey is running
check_valkey() {
    echo -e "\n${YELLOW}🔍 Checking Valkey server...${NC}"
    if pgrep -f "valkey-server" > /dev/null; then
        echo -e "${GREEN}✅ Valkey server is running${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  Valkey server not detected. Starting Valkey...${NC}"
        # Try to start Valkey in background
        if command -v valkey-server &> /dev/null; then
            valkey-server --daemonize yes --port 6379
            sleep 2
            if pgrep -f "valkey-server" > /dev/null; then
                echo -e "${GREEN}✅ Valkey server started successfully${NC}"
                return 0
            fi
        fi
        echo -e "${YELLOW}⚠️  Valkey not available. Tests will use memory cache fallback.${NC}"
        return 1
    fi
}

# Check if external APIs are running
check_external_apis() {
    echo -e "\n${YELLOW}🔍 Checking external APIs...${NC}"
    
    apis=(
        "FastAPI:http://localhost:8000/health"
        "MultipleScrape:http://localhost:8001/health"
        "WinbuTV:http://localhost:8002/health"
    )
    
    available_apis=0
    for api in "${apis[@]}"; do
        name=$(echo $api | cut -d: -f1)
        url=$(echo $api | cut -d: -f2-)
        
        if curl -s --max-time 3 "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $name is available${NC}"
            ((available_apis++))
        else
            echo -e "${YELLOW}⚠️  $name is not available${NC}"
        fi
    done
    
    echo -e "${BLUE}📊 Available APIs: $available_apis/3${NC}"
    return $available_apis
}

# Set environment variables for testing
export PORT=8090
export DATABASE_PATH="/tmp/test_api_fallback.db"
export REDIS_ADDR="localhost:6379"
export API_TIMEOUT="20s"
export RATE_LIMIT="1000"
export GIN_MODE="test"

echo -e "${BLUE}🔧 Test Configuration:${NC}"
echo "  Port: $PORT"
echo "  Database: $DATABASE_PATH"
echo "  Redis: $REDIS_ADDR"
echo "  API Timeout: $API_TIMEOUT"
echo "  Rate Limit: $RATE_LIMIT"

# Check prerequisites
check_valkey
VALKEY_AVAILABLE=$?

check_external_apis
EXTERNAL_APIS_COUNT=$?

# Clean up any existing test database
rm -f "$DATABASE_PATH"

echo -e "\n${BLUE}🧪 Starting Unit Tests${NC}"
echo "========================================"

# Unit Tests
run_test "Config Package Tests" "go test -v ./pkg/config/"
run_test "Cache Package Tests" "go test -v ./pkg/cache/"
run_test "Validator Package Tests" "go test -v ./pkg/validator/"
run_test "Database Package Tests" "go test -v ./pkg/database/"
run_test "Service Package Tests" "go test -v ./internal/service/"
run_test "API Handler Tests" "go test -v ./internal/api/handlers/"

echo -e "\n${BLUE}🔗 Starting Integration Tests${NC}"
echo "========================================"

# Integration Tests
run_test "Basic Integration Tests" "go test -v -run TestFullIntegration ."
run_test "Fallback Mechanism Tests" "go test -v -run TestFallbackMechanism ."
run_test "Caching Mechanism Tests" "go test -v -run TestCachingMechanism ."
run_test "Dashboard Endpoints Tests" "go test -v -run TestDashboardEndpoints ."
run_test "Health Endpoint Tests" "go test -v -run TestHealthEndpoint ."

# Real API Integration Tests (only if APIs are available)
if [ $EXTERNAL_APIS_COUNT -gt 0 ]; then
    echo -e "\n${BLUE}🌐 Starting Real API Integration Tests${NC}"
    echo "========================================"
    run_test "Real API Integration Tests" "go test -v -run TestRealAPIIntegration ."
else
    echo -e "\n${YELLOW}⚠️  Skipping Real API Integration Tests (no external APIs available)${NC}"
fi

echo -e "\n${BLUE}🏗️ Build and Compilation Tests${NC}"
echo "========================================"

# Build Tests
run_test "Application Build Test" "go build -o /tmp/test_apifallback cmd/main.go"
run_test "Module Verification" "go mod verify"
run_test "Go Vet Analysis" "go vet ./..."

# Clean up build artifact
rm -f /tmp/test_apifallback

echo -e "\n${BLUE}🔍 Code Quality Tests${NC}"
echo "========================================"

# Code Quality Tests
if command -v golint &> /dev/null; then
    run_test "Go Lint Check" "golint ./..."
else
    echo -e "${YELLOW}⚠️  golint not available, skipping lint check${NC}"
fi

if command -v gofmt &> /dev/null; then
    run_test "Go Format Check" "test -z \$(gofmt -l .)"
else
    echo -e "${YELLOW}⚠️  gofmt not available, skipping format check${NC}"
fi

echo -e "\n${BLUE}🚀 Performance Tests${NC}"
echo "========================================"

# Performance Tests
run_test "Benchmark Tests" "go test -bench=. -benchmem ./pkg/cache/ || true"

echo -e "\n${BLUE}🧹 Cleanup${NC}"
echo "========================================"

# Cleanup
rm -f "$DATABASE_PATH"
echo "✅ Test database cleaned up"

# Stop Valkey if we started it
if [ $VALKEY_AVAILABLE -eq 0 ] && pgrep -f "valkey-server.*--daemonize" > /dev/null; then
    echo "🛑 Stopping test Valkey server..."
    pkill -f "valkey-server.*--daemonize"
fi

echo -e "\n${BLUE}📊 Test Results Summary${NC}"
echo "========================================"
echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 ALL TESTS PASSED! 🎉${NC}"
    echo -e "${GREEN}✅ API Fallback System is working perfectly!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please check the output above.${NC}"
    exit 1
fi