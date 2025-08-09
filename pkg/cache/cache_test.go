package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache()

	// Test Set and Get
	key := "test_key"
	value := []byte("test_value")
	ttl := 1 * time.Minute

	err := cache.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Failed to set cache: %v", err)
	}

	retrieved, err := cache.Get(key)
	if err != nil {
		t.Errorf("Failed to get cache: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}

	// Test Delete
	err = cache.Delete(key)
	if err != nil {
		t.Errorf("Failed to delete cache: %v", err)
	}

	retrieved, err = cache.Get(key)
	if err != nil {
		t.Errorf("Failed to get cache after delete: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected nil after delete, got %s", string(retrieved))
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	cache := NewMemoryCache()

	key := "expire_test"
	value := []byte("expire_value")
	ttl := 100 * time.Millisecond

	err := cache.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Failed to set cache: %v", err)
	}

	// Should exist immediately
	retrieved, err := cache.Get(key)
	if err != nil {
		t.Errorf("Failed to get cache: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	retrieved, err = cache.Get(key)
	if err != nil {
		t.Errorf("Failed to get cache after expiration: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected nil after expiration, got %s", string(retrieved))
	}
}

func TestGenerateCacheKey(t *testing.T) {
	category := "anime"
	endpoint := "/api/v1/home"
	params := map[string]string{
		"page": "1",
		"sort": "latest",
	}

	key1 := generateCacheKey(category, endpoint, params)
	key2 := generateCacheKey(category, endpoint, params)

	// Same parameters should generate same key
	if key1 != key2 {
		t.Errorf("Expected same cache key for same parameters")
	}

	// Different parameters should generate different key
	params2 := map[string]string{
		"page": "2",
		"sort": "latest",
	}

	key3 := generateCacheKey(category, endpoint, params2)
	if key1 == key3 {
		t.Errorf("Expected different cache key for different parameters")
	}
}

func TestNewCache(t *testing.T) {
	// Test with invalid Redis address (should fallback to memory cache)
	cache := NewCache("invalid:6379", 0)

	// Should be memory cache
	if _, ok := cache.(*MemoryCache); !ok {
		t.Errorf("Expected MemoryCache when Redis is unavailable")
	}

	// Test basic functionality
	key := "fallback_test"
	value := []byte("fallback_value")
	ttl := 1 * time.Minute

	err := cache.Set(key, value, ttl)
	if err != nil {
		t.Errorf("Failed to set cache: %v", err)
	}

	retrieved, err := cache.Get(key)
	if err != nil {
		t.Errorf("Failed to get cache: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}
}
