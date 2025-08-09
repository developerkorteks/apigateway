package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) error
	GenerateKey(category, endpoint string, params map[string]string) string
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// MemoryCache implements Cache interface using in-memory storage
type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	return &RedisCache{
		client: rdb,
		ctx:    context.Background(),
	}
}

// NewMemoryCache creates a new in-memory cache instance
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Redis Cache Implementation
func (r *RedisCache) Get(key string) ([]byte, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key not found
		}
		return nil, err
	}
	return []byte(val), nil
}

func (r *RedisCache) Set(key string, value []byte, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisCache) GenerateKey(category, endpoint string, params map[string]string) string {
	return generateCacheKey(category, endpoint, params)
}

// Memory Cache Implementation
func (m *MemoryCache) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, nil // Key not found
	}

	if time.Now().After(item.expiresAt) {
		delete(m.data, key)
		return nil, nil // Expired
	}

	return item.value, nil
}

func (m *MemoryCache) Set(key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (m *MemoryCache) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *MemoryCache) GenerateKey(category, endpoint string, params map[string]string) string {
	return generateCacheKey(category, endpoint, params)
}

// cleanup removes expired items from memory cache
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, item := range m.data {
				if now.After(item.expiresAt) {
					delete(m.data, key)
				}
			}
			m.mu.Unlock()
		}
	}
}

// generateCacheKey creates a cache key from category, endpoint, and parameters
func generateCacheKey(category, endpoint string, params map[string]string) string {
	// Create a consistent hash of parameters
	paramBytes, _ := json.Marshal(params)
	paramHash := fmt.Sprintf("%x", md5.Sum(paramBytes))

	return fmt.Sprintf("%s:%s:%s", category, endpoint, paramHash)
}

// NewCache creates a cache instance, trying Redis/Valkey first, falling back to memory cache
func NewCache(redisAddr string, redisDB int) Cache {
	// Try to create Redis/Valkey cache
	redisCache := NewRedisCache(redisAddr, redisDB)

	// Test Redis/Valkey connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisCache.client.Ping(ctx).Err(); err != nil {
		// Redis/Valkey not available, use memory cache
		fmt.Printf("Redis/Valkey not available (%v), falling back to memory cache\n", err)
		return NewMemoryCache()
	}

	fmt.Printf("Successfully connected to Redis/Valkey at %s\n", redisAddr)
	return redisCache
}
