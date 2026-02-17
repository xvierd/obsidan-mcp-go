// Package memory provides an in-memory cache adapter implementing the CacheRepository port.
package memory

import (
	"container/list"
	"sync"
	"time"

	"github.com/xvierd/mcp-obsidian-go/internal/application/ports"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// Cache provides an LRU cache with TTL support.
// It implements the CacheRepository port.
type Cache struct {
	maxSize int
	ttl     time.Duration
	items   map[string]*cacheItem
	order   *list.List
	mu      sync.RWMutex
	hits    int64
	misses  int64
}

// cacheItem represents a cached item with metadata.
type cacheItem struct {
	key       string
	value     *domain.Note
	timestamp time.Time
	element   *list.Element
}

// NewCache creates a new cache with the specified size and TTL.
func NewCache(maxSize int, ttl time.Duration) *Cache {
	return &Cache{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*cacheItem),
		order:   list.New(),
	}
}

// Ensure Cache implements the CacheRepository port.
var _ ports.CacheRepository = (*Cache)(nil)

// Get retrieves a note from the cache.
func (c *Cache) Get(key string) (*domain.Note, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// Check if item has expired
	if time.Since(item.timestamp) > c.ttl {
		c.mu.Lock()
		c.remove(key)
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.order.MoveToFront(item.element)
	c.hits++
	c.mu.Unlock()

	return item.value, true
}

// Set adds or updates a note in the cache.
func (c *Cache) Set(key string, value *domain.Note) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, update it
	if item, exists := c.items[key]; exists {
		item.value = value
		item.timestamp = time.Now()
		c.order.MoveToFront(item.element)
		return
	}

	// Evict oldest items if at capacity
	for len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	// Add new item
	element := c.order.PushFront(key)
	c.items[key] = &cacheItem{
		key:       key,
		value:     value,
		timestamp: time.Now(),
		element:   element,
	}
}

// Delete removes a note from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remove(key)
}

// Clear removes all items from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheItem)
	c.order = list.New()
	c.hits = 0
	c.misses = 0
}

// Stats returns cache statistics.
func (c *Cache) Stats() ports.CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ports.CacheStats{
		Size:    len(c.items),
		MaxSize: c.maxSize,
		Hits:    c.hits,
		Misses:  c.misses,
		HitRate: c.hitRate(),
		TTL:     c.ttl,
	}
}

// remove removes an item from the cache (must be called with lock held).
func (c *Cache) remove(key string) {
	if item, exists := c.items[key]; exists {
		c.order.Remove(item.element)
		delete(c.items, key)
	}
}

// evictOldest removes the oldest item from the cache (must be called with lock held).
func (c *Cache) evictOldest() {
	element := c.order.Back()
	if element != nil {
		key := element.Value.(string)
		c.remove(key)
	}
}

// hitRate calculates the cache hit rate.
func (c *Cache) hitRate() float64 {
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}
