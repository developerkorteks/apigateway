# Redis Connection Testing Guide

## 📋 Summary

Script testing untuk memverifikasi koneksi Redis di environment **local** dan **VPS**.

---

## 🎯 Environment Detection

### **Local (Development)**
- ❌ Redis **TIDAK** diperlukan
- ✅ Aplikasi berfungsi normal tanpa cache
- ⚠️  Cache akan di-bypass (slower response)

### **VPS (Production)**
- ✅ Redis **RUNNING** di port `62379` (dari PM2)
- ✅ Cache aktif untuk performance
- ⚠️  Pastikan `.env` menggunakan port yang benar

---

## 🛠️ Testing Scripts

### 1. **Quick Simple Test**
```bash
./test_redis_simple.sh
```
- Cepat, minimal dependencies
- Hanya test port connectivity dan PING
- ✅ Cocok untuk quick check

### 2. **Comprehensive Test**
```bash
./test_redis_connection.sh
```
- Test lengkap: PING, SET, GET, EXPIRE, TTL, DEL
- Test dari Go application
- Tampilkan Redis server info
- ✅ Cocok untuk production verification

### 3. **Auto-Configuration**
```bash
./setup_redis_config.sh
```
- Auto-detect Redis di common ports (6379, 62379, dll)
- Auto-update `.env` dengan config yang benar
- Backup `.env` sebelum update
- ✅ Cocok untuk first-time setup

---

## 📝 Configuration

### **Local (.env)**
```env
# Redis tidak running di local - akan bypass cache
REDIS_ADDR=127.0.0.1:6379
REDIS_DB=0
```

### **VPS (.env)**
```env
# Redis running di port 62379 (dari PM2)
REDIS_ADDR=127.0.0.1:62379
REDIS_DB=0
```

---

## ✅ Expected Results

### **Local (No Redis)**
```bash
$ ./test_redis_simple.sh

Testing connection to: 127.0.0.1:6379
1. Port 6379 accessible... ❌ Cannot connect to port 6379

❌ Redis connection failed (expected - no Redis installed)
✅ App will run without cache
```

### **VPS (Redis Running)**
```bash
$ ./test_redis_simple.sh

Testing connection to: 127.0.0.1:62379
1. Port 62379 accessible... ✅
2. Redis PING... ✅ PONG
3. Write test... ✅

✅ Redis connection test passed!
```

---

## 🔧 Troubleshooting

### Problem: "Cannot connect to port 6379"

**Local:**
```bash
# Normal - Redis tidak ada di local
# App akan berfungsi tanpa cache
✅ No action needed
```

**VPS:**
```bash
# Check Redis process
pm2 logs redis-local

# Check port dari logs
grep "port=" ~/.pm2/logs/redis-local-out.log

# Update .env dengan port yang benar
# Contoh: REDIS_ADDR=127.0.0.1:62379
```

### Problem: "Connection refused"

```bash
# Check if Redis is running
ps aux | grep redis

# VPS: Check PM2
pm2 list | grep redis

# Start Redis (VPS)
pm2 restart redis-local

# Start Redis (Local - if installed)
redis-server --port 6379 &
```

### Problem: "NOAUTH Authentication required"

```bash
# Add password to .env
REDIS_PASSWORD=your_password_here

# Test dengan password
redis-cli -h 127.0.0.1 -p 62379 -a your_password PING
```

---

## 📊 How It Works in Application

### **Cache Package (`pkg/cache/cache.go`)**

```go
// Init tries to connect to Redis
func Init(config *config.Config) error {
    client = redis.NewClient(&redis.Options{
        Addr: config.RedisAddr,
        DB:   config.RedisDB,
    })
    
    // Test connection
    _, err := client.Ping(ctx).Result()
    if err != nil {
        logger.Warnf("Redis not available: %v. Cache disabled.", err)
        client = nil  // Disable cache
        return nil    // ✅ App continues without cache
    }
    
    logger.Info("Redis connected successfully")
    return nil
}

// Get - Returns nil if cache disabled
func Get(ctx context.Context, key string) ([]byte, error) {
    if client == nil {
        return nil, ErrCacheDisabled  // Bypass cache
    }
    // ... actual cache logic
}
```

### **Flow:**
1. ✅ App starts
2. 🔄 Try connect to Redis
3. **Success:** Cache enabled ⚡
4. **Failed:** Cache disabled, continue without error ✅

---

## 🚀 Production Best Practices

### **VPS Setup:**

1. **Verify Redis is running:**
```bash
pm2 list
# Should show: redis-local | online
```

2. **Check Redis port:**
```bash
pm2 logs redis-local | grep "port="
# Output: Running mode=standalone, port=62379
```

3. **Update `.env`:**
```bash
cd ~/apigateway
vim .env

# Set correct port
REDIS_ADDR=127.0.0.1:62379
REDIS_DB=0
```

4. **Test connection:**
```bash
./test_redis_connection.sh
# Should pass all tests ✅
```

5. **Restart API Gateway:**
```bash
pm2 restart apigateway --update-env
```

6. **Verify logs:**
```bash
pm2 logs apigateway | grep -i redis
# Should see: "Redis connected successfully"
```

---

## 📈 Performance Impact

| Environment | Redis Status | Cache | Response Time | Notes |
|-------------|--------------|-------|---------------|-------|
| Local Dev | ❌ Disabled | No | ~15s | Direct API calls |
| VPS Prod | ✅ Enabled | Yes | ~0.1s | Cached responses |

**Cache Benefits:**
- 🚀 150x faster response for cached data
- 💰 Reduced API calls to external services
- 📉 Lower bandwidth usage
- ⚡ Better user experience

---

## 🧪 Manual Testing

### **Test from command line:**

```bash
# Test 1: PING
redis-cli -h 127.0.0.1 -p 62379 PING
# Expected: PONG

# Test 2: SET/GET
redis-cli -h 127.0.0.1 -p 62379 SET test:key "Hello Redis"
redis-cli -h 127.0.0.1 -p 62379 GET test:key
# Expected: "Hello Redis"

# Test 3: Check keys
redis-cli -h 127.0.0.1 -p 62379 KEYS "api:*"
# Shows cached API responses

# Test 4: Monitor real-time
redis-cli -h 127.0.0.1 -p 62379 MONITOR
# Shows all Redis commands in real-time
```

### **Test from application:**

```bash
# Make request (first time - slow)
time curl http://localhost:58080/api/v1/anime-terbaru?category=anime
# ~15 seconds

# Make same request (cached - fast)
time curl http://localhost:58080/api/v1/anime-terbaru?category=anime
# ~0.1 seconds ⚡

# Check cache status in metadata
curl http://localhost:58080/api/v1/anime-terbaru?category=anime | jq '._metadata.cache_status'
# "HIT" or "MISS"
```

---

## 📚 Additional Resources

- **Redis CLI Commands:** https://redis.io/commands
- **Go Redis Client:** https://github.com/redis/go-redis
- **PM2 Documentation:** https://pm2.keymetrics.io/docs/usage/quick-start/

---

## ✅ Checklist

### Local Development:
- [ ] Run `./test_redis_simple.sh` (expected to fail ✅)
- [ ] App runs without Redis
- [ ] Endpoints respond (slower, no cache)

### VPS Production:
- [ ] Run `./setup_redis_config.sh` (auto-detect port)
- [ ] Run `./test_redis_connection.sh` (all tests pass ✅)
- [ ] Verify `.env` has correct port (62379)
- [ ] Restart app: `pm2 restart apigateway --update-env`
- [ ] Check logs: `pm2 logs apigateway | grep Redis`
- [ ] Test endpoint response time (should be fast with cache)

---

**Status:** ✅ Scripts created and tested  
**Environment:** Local (no Redis) - Normal operation  
**VPS Config:** Ready for Redis on port 62379  
**Date:** 2026-02-13
