#!/bin/bash

# Simple Redis Connection Test (No dependencies)

echo "=== Simple Redis Connection Test ==="
echo ""

# Check .env
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | grep REDIS | xargs)
    echo "Configuration from .env:"
    echo "  REDIS_ADDR: ${REDIS_ADDR:-<not set>}"
    echo "  REDIS_DB: ${REDIS_DB:-<not set>}"
else
    echo "⚠️  No .env file found"
    REDIS_ADDR="127.0.0.1:6379"
fi

REDIS_HOST=$(echo ${REDIS_ADDR:-127.0.0.1:6379} | cut -d':' -f1)
REDIS_PORT=$(echo ${REDIS_ADDR:-127.0.0.1:6379} | cut -d':' -f2)

echo ""
echo "Testing connection to: $REDIS_HOST:$REDIS_PORT"
echo ""

# Test 1: Port connectivity
echo -n "1. Port $REDIS_PORT accessible... "
if timeout 2 bash -c "cat < /dev/null > /dev/tcp/$REDIS_HOST/$REDIS_PORT" 2>/dev/null; then
    echo "✅"
else
    echo "❌ Cannot connect to port $REDIS_PORT"
    exit 1
fi

# Test 2: Redis PING (if redis-cli available)
if command -v redis-cli &> /dev/null; then
    echo -n "2. Redis PING... "
    RESULT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT PING 2>&1)
    if [ "$RESULT" == "PONG" ]; then
        echo "✅ $RESULT"
    else
        echo "❌ $RESULT"
        exit 1
    fi
    
    # Test 3: Write test
    echo -n "3. Write test... "
    redis-cli -h $REDIS_HOST -p $REDIS_PORT SET test:simple "OK" > /dev/null 2>&1
    RESULT=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT GET test:simple 2>&1)
    if [ "$RESULT" == "OK" ]; then
        echo "✅"
        redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL test:simple > /dev/null 2>&1
    else
        echo "❌"
        exit 1
    fi
else
    echo "2. redis-cli not found - skipping PING test"
fi

echo ""
echo "✅ Redis connection test passed!"
echo ""
echo "Next step: Update your .env file with:"
echo "  REDIS_ADDR=$REDIS_ADDR"
echo "  REDIS_DB=${REDIS_DB:-0}"
