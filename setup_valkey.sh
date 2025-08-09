#!/bin/bash

# Valkey Setup Script for API Fallback System
echo "🔧 Setting up Valkey for API Fallback System"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
VALKEY_PORT=6379
VALKEY_CONFIG_FILE="/tmp/valkey-apifallback.conf"
VALKEY_LOG_FILE="/tmp/valkey-apifallback.log"
VALKEY_PID_FILE="/tmp/valkey-apifallback.pid"

# Check if Valkey is installed
check_valkey_installation() {
    echo -e "${BLUE}🔍 Checking Valkey installation...${NC}"
    
    if command -v valkey-server &> /dev/null; then
        echo -e "${GREEN}✅ Valkey server found: $(which valkey-server)${NC}"
        valkey-server --version
        return 0
    elif command -v redis-server &> /dev/null; then
        echo -e "${YELLOW}⚠️  Valkey not found, but Redis is available: $(which redis-server)${NC}"
        echo -e "${YELLOW}    Redis can be used as a compatible alternative${NC}"
        redis-server --version
        return 1
    else
        echo -e "${RED}❌ Neither Valkey nor Redis found${NC}"
        echo -e "${YELLOW}💡 Installation suggestions:${NC}"
        echo "   - Arch/CachyOS: sudo pacman -S valkey"
        echo "   - Ubuntu/Debian: sudo apt install valkey-server"
        echo "   - Or install Redis: sudo pacman -S redis"
        return 2
    fi
}

# Create Valkey configuration
create_valkey_config() {
    echo -e "\n${BLUE}📝 Creating Valkey configuration...${NC}"
    
    cat > "$VALKEY_CONFIG_FILE" << EOF
# Valkey Configuration for API Fallback System
port $VALKEY_PORT
bind 127.0.0.1
daemonize yes
pidfile $VALKEY_PID_FILE
logfile $VALKEY_LOG_FILE
loglevel notice

# Memory and persistence settings
maxmemory 256mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000

# Security
protected-mode yes
# requirepass your_password_here

# Performance
tcp-keepalive 300
timeout 0
tcp-backlog 511
databases 16

# Disable dangerous commands in production
# rename-command FLUSHDB ""
# rename-command FLUSHALL ""
# rename-command DEBUG ""
EOF

    echo -e "${GREEN}✅ Configuration created: $VALKEY_CONFIG_FILE${NC}"
}

# Start Valkey server
start_valkey() {
    echo -e "\n${BLUE}🚀 Starting Valkey server...${NC}"
    
    # Check if already running
    if pgrep -f "valkey-server.*$VALKEY_PORT" > /dev/null; then
        echo -e "${YELLOW}⚠️  Valkey server already running on port $VALKEY_PORT${NC}"
        return 0
    fi
    
    # Start Valkey with our configuration
    if command -v valkey-server &> /dev/null; then
        valkey-server "$VALKEY_CONFIG_FILE"
        sleep 2
        
        if pgrep -f "valkey-server.*$VALKEY_PORT" > /dev/null; then
            echo -e "${GREEN}✅ Valkey server started successfully${NC}"
            echo -e "${BLUE}📊 Server info:${NC}"
            echo "   Port: $VALKEY_PORT"
            echo "   Config: $VALKEY_CONFIG_FILE"
            echo "   Log: $VALKEY_LOG_FILE"
            echo "   PID: $VALKEY_PID_FILE"
            return 0
        else
            echo -e "${RED}❌ Failed to start Valkey server${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Valkey server not found${NC}"
        return 1
    fi
}

# Test Valkey connection
test_valkey_connection() {
    echo -e "\n${BLUE}🔍 Testing Valkey connection...${NC}"
    
    if command -v valkey-cli &> /dev/null; then
        CLI_CMD="valkey-cli"
    elif command -v redis-cli &> /dev/null; then
        CLI_CMD="redis-cli"
    else
        echo -e "${YELLOW}⚠️  No CLI client found, skipping connection test${NC}"
        return 1
    fi
    
    # Test basic connection
    if $CLI_CMD -p $VALKEY_PORT ping > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Connection successful${NC}"
        
        # Get server info
        echo -e "${BLUE}📊 Server information:${NC}"
        $CLI_CMD -p $VALKEY_PORT info server | grep -E "(redis_version|valkey_version|os|arch|process_id|uptime_in_seconds)"
        
        # Test basic operations
        echo -e "\n${BLUE}🧪 Testing basic operations:${NC}"
        $CLI_CMD -p $VALKEY_PORT set test_key "API Fallback Test" > /dev/null
        TEST_VALUE=$($CLI_CMD -p $VALKEY_PORT get test_key)
        
        if [ "$TEST_VALUE" = "API Fallback Test" ]; then
            echo -e "${GREEN}✅ SET/GET operations working${NC}"
            $CLI_CMD -p $VALKEY_PORT del test_key > /dev/null
        else
            echo -e "${RED}❌ SET/GET operations failed${NC}"
        fi
        
        return 0
    else
        echo -e "${RED}❌ Connection failed${NC}"
        return 1
    fi
}

# Stop Valkey server
stop_valkey() {
    echo -e "\n${BLUE}🛑 Stopping Valkey server...${NC}"
    
    if [ -f "$VALKEY_PID_FILE" ]; then
        PID=$(cat "$VALKEY_PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID"
            echo -e "${GREEN}✅ Valkey server stopped (PID: $PID)${NC}"
        else
            echo -e "${YELLOW}⚠️  PID file exists but process not running${NC}"
        fi
        rm -f "$VALKEY_PID_FILE"
    else
        # Try to find and kill by process name
        if pgrep -f "valkey-server.*$VALKEY_PORT" > /dev/null; then
            pkill -f "valkey-server.*$VALKEY_PORT"
            echo -e "${GREEN}✅ Valkey server stopped${NC}"
        else
            echo -e "${YELLOW}⚠️  Valkey server not running${NC}"
        fi
    fi
}

# Cleanup function
cleanup() {
    echo -e "\n${BLUE}🧹 Cleaning up temporary files...${NC}"
    rm -f "$VALKEY_CONFIG_FILE" "$VALKEY_LOG_FILE"
    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

# Show usage
show_usage() {
    echo "Usage: $0 [start|stop|restart|test|status|cleanup]"
    echo ""
    echo "Commands:"
    echo "  start    - Start Valkey server"
    echo "  stop     - Stop Valkey server"
    echo "  restart  - Restart Valkey server"
    echo "  test     - Test Valkey connection"
    echo "  status   - Show Valkey status"
    echo "  cleanup  - Clean up temporary files"
    echo ""
}

# Show status
show_status() {
    echo -e "${BLUE}📊 Valkey Status${NC}"
    echo "=================="
    
    if pgrep -f "valkey-server.*$VALKEY_PORT" > /dev/null; then
        PID=$(pgrep -f "valkey-server.*$VALKEY_PORT")
        echo -e "${GREEN}✅ Status: Running (PID: $PID)${NC}"
        echo "   Port: $VALKEY_PORT"
        echo "   Config: $VALKEY_CONFIG_FILE"
        echo "   Log: $VALKEY_LOG_FILE"
        
        if [ -f "$VALKEY_LOG_FILE" ]; then
            echo -e "\n${BLUE}📝 Recent log entries:${NC}"
            tail -5 "$VALKEY_LOG_FILE"
        fi
    else
        echo -e "${RED}❌ Status: Not running${NC}"
    fi
}

# Main script logic
case "${1:-start}" in
    "start")
        check_valkey_installation
        VALKEY_STATUS=$?
        
        if [ $VALKEY_STATUS -eq 0 ]; then
            create_valkey_config
            start_valkey
            test_valkey_connection
        elif [ $VALKEY_STATUS -eq 1 ]; then
            echo -e "${YELLOW}💡 You can use Redis instead of Valkey${NC}"
            echo -e "${YELLOW}   Just update REDIS_ADDR to point to your Redis instance${NC}"
        else
            echo -e "${RED}❌ Cannot start - Valkey/Redis not installed${NC}"
            exit 1
        fi
        ;;
    "stop")
        stop_valkey
        ;;
    "restart")
        stop_valkey
        sleep 1
        create_valkey_config
        start_valkey
        test_valkey_connection
        ;;
    "test")
        test_valkey_connection
        ;;
    "status")
        show_status
        ;;
    "cleanup")
        stop_valkey
        cleanup
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        echo -e "${RED}❌ Unknown command: $1${NC}"
        show_usage
        exit 1
        ;;
esac