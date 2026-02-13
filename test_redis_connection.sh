#!/bin/bash

# Test Redis Connection Script
# Works for both local and VPS environments

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         Redis Connection Test Script                         ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Load environment variables
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
    echo "✅ Loaded .env file"
else
    echo "⚠️  No .env file found, using default values"
fi

# Get Redis config from .env or use defaults
REDIS_ADDR=${REDIS_ADDR:-"127.0.0.1:6379"}
REDIS_DB=${REDIS_DB:-0}
REDIS_PASSWORD=${REDIS_PASSWORD:-""}

# Extract host and port
REDIS_HOST=$(echo $REDIS_ADDR | cut -d':' -f1)
REDIS_PORT=$(echo $REDIS_ADDR | cut -d':' -f2)

echo ""
echo "📋 Configuration:"
echo "   Host: $REDIS_HOST"
echo "   Port: $REDIS_PORT"
echo "   DB: $REDIS_DB"
echo "   Password: ${REDIS_PASSWORD:-<none>}"
echo ""

# Test 1: Check if Redis server is running
echo "════════════════════════════════════════════════════════════════"
echo "Test 1: Checking if Redis is running on $REDIS_HOST:$REDIS_PORT"
echo "════════════════════════════════════════════════════════════════"

if command -v redis-cli &> /dev/null; then
    echo "✅ redis-cli found"
    
    # Test connection
    if [ -z "$REDIS_PASSWORD" ]; then
        PING_RESULT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT PING 2>&1)
    else
        PING_RESULT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD PING 2>&1)
    fi
    
    if [ "$PING_RESULT" == "PONG" ]; then
        echo "✅ Redis is responding: $PING_RESULT"
    else
        echo "❌ Redis not responding: $PING_RESULT"
        echo ""
        echo "💡 Troubleshooting:"
        echo "   1. Check if Redis is running: ps aux | grep redis"
        echo "   2. Check port: netstat -tlnp | grep $REDIS_PORT"
        echo "   3. Check .env REDIS_ADDR configuration"
        exit 1
    fi
else
    echo "⚠️  redis-cli not found, trying netcat..."
    
    if command -v nc &> /dev/null; then
        if nc -z $REDIS_HOST $REDIS_PORT 2>/dev/null; then
            echo "✅ Port $REDIS_PORT is open on $REDIS_HOST"
        else
            echo "❌ Cannot connect to $REDIS_HOST:$REDIS_PORT"
            exit 1
        fi
    else
        echo "⚠️  Neither redis-cli nor nc found, skipping connectivity test"
    fi
fi

echo ""

# Test 2: Test basic Redis operations (if redis-cli available)
if command -v redis-cli &> /dev/null; then
    echo "════════════════════════════════════════════════════════════════"
    echo "Test 2: Testing Redis Operations"
    echo "════════════════════════════════════════════════════════════════"
    
    TEST_KEY="test:connection:$(date +%s)"
    TEST_VALUE="Redis connection test successful!"
    
    if [ -z "$REDIS_PASSWORD" ]; then
        REDIS_CMD="redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB"
    else
        REDIS_CMD="redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $REDIS_DB"
    fi
    
    # SET operation
    echo -n "   Testing SET... "
    SET_RESULT=$($REDIS_CMD SET "$TEST_KEY" "$TEST_VALUE" 2>&1)
    if [ "$SET_RESULT" == "OK" ]; then
        echo "✅"
    else
        echo "❌ Failed: $SET_RESULT"
    fi
    
    # GET operation
    echo -n "   Testing GET... "
    GET_RESULT=$($REDIS_CMD GET "$TEST_KEY" 2>&1)
    if [ "$GET_RESULT" == "$TEST_VALUE" ]; then
        echo "✅"
    else
        echo "❌ Failed: Expected '$TEST_VALUE', got '$GET_RESULT'"
    fi
    
    # EXPIRE operation
    echo -n "   Testing EXPIRE... "
    EXPIRE_RESULT=$($REDIS_CMD EXPIRE "$TEST_KEY" 10 2>&1)
    if [ "$EXPIRE_RESULT" == "1" ]; then
        echo "✅"
    else
        echo "❌ Failed: $EXPIRE_RESULT"
    fi
    
    # TTL operation
    echo -n "   Testing TTL... "
    TTL_RESULT=$($REDIS_CMD TTL "$TEST_KEY" 2>&1)
    if [ "$TTL_RESULT" -gt 0 ] && [ "$TTL_RESULT" -le 10 ]; then
        echo "✅ (TTL: ${TTL_RESULT}s)"
    else
        echo "⚠️  Unexpected TTL: $TTL_RESULT"
    fi
    
    # DEL operation
    echo -n "   Testing DEL... "
    DEL_RESULT=$($REDIS_CMD DEL "$TEST_KEY" 2>&1)
    if [ "$DEL_RESULT" == "1" ]; then
        echo "✅"
    else
        echo "❌ Failed: $DEL_RESULT"
    fi
    
    echo ""
fi

# Test 3: Get Redis Info
if command -v redis-cli &> /dev/null; then
    echo "════════════════════════════════════════════════════════════════"
    echo "Test 3: Redis Server Information"
    echo "════════════════════════════════════════════════════════════════"
    
    if [ -z "$REDIS_PASSWORD" ]; then
        INFO_RESULT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT INFO server 2>&1 | grep -E "redis_version|os|tcp_port|uptime_in_seconds")
    else
        INFO_RESULT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD INFO server 2>&1 | grep -E "redis_version|os|tcp_port|uptime_in_seconds")
    fi
    
    echo "$INFO_RESULT"
    echo ""
    
    # Get memory info
    if [ -z "$REDIS_PASSWORD" ]; then
        MEMORY_INFO=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT INFO memory 2>&1 | grep -E "used_memory_human|used_memory_peak_human")
    else
        MEMORY_INFO=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD INFO memory 2>&1 | grep -E "used_memory_human|used_memory_peak_human")
    fi
    
    echo "$MEMORY_INFO"
    echo ""
fi

# Test 4: Test from Go application
echo "════════════════════════════════════════════════════════════════"
echo "Test 4: Testing Go Application Redis Connection"
echo "════════════════════════════════════════════════════════════════"

cat > tmp_test_redis.go << 'GOTEST'
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Get config from environment
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	
	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	
	fmt.Printf("Connecting to Redis at %s...\n", addr)
	
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Test PING
	fmt.Print("   Testing PING... ")
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ %s\n", pong)
	
	// Test SET
	testKey := fmt.Sprintf("go:test:connection:%d", time.Now().Unix())
	testValue := "Go Redis connection test successful!"
	
	fmt.Print("   Testing SET... ")
	err = client.Set(ctx, testKey, testValue, 10*time.Second).Err()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅")
	
	// Test GET
	fmt.Print("   Testing GET... ")
	val, err := client.Get(ctx, testKey).Result()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	if val != testValue {
		fmt.Printf("❌ Value mismatch: got '%s', expected '%s'\n", val, testValue)
		os.Exit(1)
	}
	fmt.Println("✅")
	
	// Test DEL
	fmt.Print("   Testing DEL... ")
	err = client.Del(ctx, testKey).Err()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅")
	
	fmt.Println("\n✅ All Go Redis tests passed!")
	
	client.Close()
}
GOTEST

echo "Running Go test..."
go run tmp_test_redis.go 2>&1

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Go application can connect to Redis successfully!"
else
    echo ""
    echo "❌ Go application failed to connect to Redis"
    echo ""
    echo "💡 Troubleshooting:"
    echo "   1. Check .env file: REDIS_ADDR should be $REDIS_ADDR"
    echo "   2. Verify Redis is running: redis-cli -h $REDIS_HOST -p $REDIS_PORT PING"
    echo "   3. Check firewall: sudo firewall-cmd --list-ports"
fi

# Cleanup
rm -f tmp_test_redis.go

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✅ Redis Connection Test Complete"
echo "════════════════════════════════════════════════════════════════"
