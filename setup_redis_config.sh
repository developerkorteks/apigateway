#!/bin/bash

# Redis Configuration Setup Script
# Automatically detects available Redis and updates .env

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         Redis Auto-Configuration Script                      ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Common Redis ports to check
PORTS=(6379 6380 62379 7000)

echo "🔍 Scanning for Redis servers..."
echo ""

FOUND_REDIS=""
FOUND_PORT=""

for port in "${PORTS[@]}"; do
    echo -n "Checking port $port... "
    
    # Check if port is open
    if timeout 1 bash -c "cat < /dev/null > /dev/tcp/127.0.0.1/$port" 2>/dev/null; then
        echo "✅ Port open"
        
        # Try PING if redis-cli available
        if command -v redis-cli &> /dev/null; then
            PING_RESULT=$(redis-cli -h 127.0.0.1 -p $port PING 2>&1)
            if [ "$PING_RESULT" == "PONG" ]; then
                echo "   ✅ Redis responding on port $port!"
                FOUND_REDIS="yes"
                FOUND_PORT=$port
                
                # Get Redis info
                VERSION=$(redis-cli -h 127.0.0.1 -p $port INFO server 2>&1 | grep "redis_version:" | cut -d':' -f2 | tr -d '\r')
                echo "   📋 Redis version: $VERSION"
                break
            fi
        else
            echo "   ⚠️  Port open but cannot verify (redis-cli not found)"
            FOUND_REDIS="maybe"
            FOUND_PORT=$port
        fi
    else
        echo "❌ Closed"
    fi
done

echo ""

if [ "$FOUND_REDIS" == "yes" ]; then
    echo "════════════════════════════════════════════════════════════════"
    echo "✅ Redis found on port $FOUND_PORT"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    
    # Update .env
    if [ -f ".env" ]; then
        echo "Updating .env file..."
        
        # Backup .env
        cp .env .env.backup_redis_$(date +%Y%m%d_%H%M%S)
        
        # Update or add REDIS_ADDR
        if grep -q "^REDIS_ADDR=" .env; then
            sed -i "s/^REDIS_ADDR=.*/REDIS_ADDR=127.0.0.1:$FOUND_PORT/" .env
            echo "✅ Updated REDIS_ADDR=127.0.0.1:$FOUND_PORT"
        else
            echo "" >> .env
            echo "# Redis Configuration (Auto-detected)" >> .env
            echo "REDIS_ADDR=127.0.0.1:$FOUND_PORT" >> .env
            echo "✅ Added REDIS_ADDR=127.0.0.1:$FOUND_PORT"
        fi
        
        # Ensure REDIS_DB exists
        if ! grep -q "^REDIS_DB=" .env; then
            echo "REDIS_DB=0" >> .env
            echo "✅ Added REDIS_DB=0"
        fi
        
        echo ""
        echo "Current Redis configuration:"
        grep "REDIS" .env | grep -v "^#"
    else
        echo "⚠️  No .env file found. Create one with:"
        echo ""
        echo "REDIS_ADDR=127.0.0.1:$FOUND_PORT"
        echo "REDIS_DB=0"
    fi
    
elif [ "$FOUND_REDIS" == "maybe" ]; then
    echo "════════════════════════════════════════════════════════════════"
    echo "⚠️  Port $FOUND_PORT is open but cannot verify it's Redis"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    echo "To configure manually, add to .env:"
    echo "  REDIS_ADDR=127.0.0.1:$FOUND_PORT"
    echo "  REDIS_DB=0"
    
else
    echo "════════════════════════════════════════════════════════════════"
    echo "❌ No Redis server found on common ports"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    echo "Options:"
    echo ""
    echo "1. For LOCAL development (no Redis needed):"
    echo "   - App will work without cache"
    echo "   - Set in .env: REDIS_ADDR=127.0.0.1:6379 (it will fail gracefully)"
    echo ""
    echo "2. For VPS with Redis on custom port:"
    echo "   - Check: pm2 logs redis-local"
    echo "   - Find port in logs or pm2 config"
    echo "   - Update .env: REDIS_ADDR=127.0.0.1:<port>"
    echo ""
    echo "3. Install Redis (optional for local):"
    echo "   - Ubuntu/Debian: sudo apt install redis-server"
    echo "   - macOS: brew install redis"
    echo "   - Start: redis-server --port 6379 &"
fi

echo ""
echo "Next step: Run ./test_redis_connection.sh to verify"
