package core

import (
	"sync"
	"time"
)

// TTLCache is a small thread-safe in-memory cache with per-entry TTL.
// Suitable for short-lived caching of provider responses (60s in v1).
type TTLCache struct {
	mu      sync.Mutex
	entries map[string]ttlEntry
	ttl     time.Duration
	max     int
}

type ttlEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewTTLCache creates a TTL cache. If max > 0 and capacity is reached,
// the oldest-expiring entries are evicted on Set.
func NewTTLCache(ttl time.Duration, max int) *TTLCache {
	return &TTLCache{
		entries: make(map[string]ttlEntry),
		ttl:     ttl,
		max:     max,
	}
}

// Get returns the cached value for key if present and unexpired.
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(c.entries, key)
		return nil, false
	}
	return entry.value, true
}

// Set stores value under key with the cache's TTL.
func (c *TTLCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.max > 0 && len(c.entries) >= c.max {
		// Evict the entry expiring soonest.
		var evictKey string
		var evictAt time.Time
		first := true
		for k, e := range c.entries {
			if first || e.expiresAt.Before(evictAt) {
				evictKey = k
				evictAt = e.expiresAt
				first = false
			}
		}
		if evictKey != "" {
			delete(c.entries, evictKey)
		}
	}
	c.entries[key] = ttlEntry{value: value, expiresAt: time.Now().Add(c.ttl)}
}
