package utils

import (
	"encoding/json"
	"time"

	"github.com/nabilulilalbab/winbu.tv/config"
	"github.com/patrickmn/go-cache"
)

type CacheManager struct {
	cache   *cache.Cache
	enabled bool
	ttl     time.Duration
}

func NewCacheManager(cfg *config.Config) *CacheManager {
	return &CacheManager{
		cache:   cache.New(cfg.CacheTTL, cfg.CacheTTL*2),
		enabled: cfg.CacheEnabled,
		ttl:     cfg.CacheTTL,
	}
}

func (c *CacheManager) Get(key string, result interface{}) bool {
	if !c.enabled {
		return false
	}

	data, found := c.cache.Get(key)
	if !found {
		return false
	}

	// Try to unmarshal the cached data
	if jsonData, ok := data.([]byte); ok {
		err := json.Unmarshal(jsonData, result)
		return err == nil
	}

	return false
}

func (c *CacheManager) Set(key string, data interface{}) {
	if !c.enabled {
		return
	}

	// Marshal data to JSON for storage
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	c.cache.Set(key, jsonData, c.ttl)
}

func (c *CacheManager) Delete(key string) {
	if !c.enabled {
		return
	}
	c.cache.Delete(key)
}

func (c *CacheManager) Clear() {
	if !c.enabled {
		return
	}
	c.cache.Flush()
}

// Cache is an alias for CacheManager for backward compatibility
type Cache = CacheManager

// NewCache creates a new cache with default configuration
func NewCache() *Cache {
	return &Cache{
		cache:   cache.New(5*time.Minute, 10*time.Minute),
		enabled: true,
		ttl:     5 * time.Minute,
	}
}

// SetWithTTL method with TTL parameter for Cache
func (c *Cache) SetWithTTL(key string, data interface{}, ttlSeconds int) {
	if !c.enabled {
		return
	}

	// Marshal data to JSON for storage
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	c.cache.Set(key, jsonData, ttl)
}
