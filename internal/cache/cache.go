package cache

import (
	"sync"
	"time"
)

// Cache is a thread-safe in-memory cache with TTL support
type Cache struct {
	mu      sync.RWMutex
	items   map[string]*cacheItem
	stats   *Statistics
	running bool
	stopCh  chan struct{}
}

// cacheItem represents a cached value with expiration
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// Statistics tracks cache performance metrics
type Statistics struct {
	mu              sync.RWMutex
	hits            int64
	misses          int64
	sets            int64
	evictions       int64
	currentSize     int
	totalSizeBytes  int64
}

// NewCache creates a new cache with automatic cleanup
func NewCache() *Cache {
	c := &Cache{
		items:   make(map[string]*cacheItem),
		stats:   &Statistics{},
		running: true,
		stopCh:  make(chan struct{}),
	}
	go c.cleanupExpired()
	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.recordMiss()
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		c.stats.recordMiss()
		return nil, false
	}

	c.stats.recordHit()
	return item.value, true
}

// Set stores a value in the cache with TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	c.stats.recordSet()
	c.stats.updateSize(len(c.items))
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		delete(c.items, key)
		c.stats.updateSize(len(c.items))
	}
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.stats.updateSize(0)
}

// GetStats returns cache statistics
func (c *Cache) GetStats() CacheStats {
	return c.stats.snapshot()
}

// Close stops the cache cleanup goroutine
func (c *Cache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		c.running = false
		close(c.stopCh)
	}
}

// cleanupExpired periodically removes expired items
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopCh:
			return
		}
	}
}

// removeExpired removes all expired items from the cache
func (c *Cache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	evicted := 0

	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
			evicted++
		}
	}

	if evicted > 0 {
		c.stats.recordEvictions(evicted)
		c.stats.updateSize(len(c.items))
	}
}

// CacheStats represents cache statistics snapshot
type CacheStats struct {
	Hits        int64
	Misses      int64
	Sets        int64
	Evictions   int64
	CurrentSize int
	HitRate     float64
}

// recordHit increments hit counter
func (s *Statistics) recordHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hits++
}

// recordMiss increments miss counter
func (s *Statistics) recordMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.misses++
}

// recordSet increments set counter
func (s *Statistics) recordSet() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sets++
}

// recordEvictions adds to eviction counter
func (s *Statistics) recordEvictions(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evictions += int64(count)
}

// updateSize updates current cache size
func (s *Statistics) updateSize(size int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentSize = size
}

// snapshot returns a statistics snapshot
func (s *Statistics) snapshot() CacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.hits + s.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(s.hits) / float64(total) * 100
	}

	return CacheStats{
		Hits:        s.hits,
		Misses:      s.misses,
		Sets:        s.sets,
		Evictions:   s.evictions,
		CurrentSize: s.currentSize,
		HitRate:     hitRate,
	}
}
