// Package cache provides an LRU cache with TTL support for Obsidian notes.
package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xvierd/mcp-obsidian-go/internal/obsidian"
)

// Cache provides an LRU cache with TTL support.
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
	value     *obsidian.Note
	timestamp time.Time
	element   *list.Element
}

// New creates a new cache with the specified size and TTL.
func New(maxSize int, ttl time.Duration) *Cache {
	return &Cache{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*cacheItem),
		order:   list.New(),
	}
}

// Get retrieves a note from the cache.
func (c *Cache) Get(key string) (*obsidian.Note, bool) {
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
func (c *Cache) Set(key string, value *obsidian.Note) {
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

// Stats returns cache statistics.
func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Size:      len(c.items),
		MaxSize:   c.maxSize,
		Hits:      c.hits,
		Misses:    c.misses,
		HitRate:   c.hitRate(),
		TTL:       c.ttl,
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

// Stats holds cache statistics.
type Stats struct {
	Size    int           `json:"size"`
	MaxSize int           `json:"max_size"`
	Hits    int64         `json:"hits"`
	Misses  int64         `json:"misses"`
	HitRate float64       `json:"hit_rate"`
	TTL     time.Duration `json:"ttl"`
}

// NoteClientWithCache wraps an Obsidian client with caching.
type NoteClientWithCache struct {
	client *obsidian.Client
	cache  *Cache
}

// NewNoteClientWithCache creates a new cached client wrapper.
func NewNoteClientWithCache(client *obsidian.Client, cache *Cache) *NoteClientWithCache {
	return &NoteClientWithCache{
		client: client,
		cache:  cache,
	}
}

// GetNote retrieves a note, using cache if available.
func (c *NoteClientWithCache) GetNote(ctx context.Context, path string) (*obsidian.Note, error) {
	// Try cache first
	if note, found := c.cache.Get(path); found {
		return note, nil
	}

	// Fetch from API
	note, err := c.client.GetNote(ctx, path)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(path, note)

	return note, nil
}

// Invalidate removes a note from the cache.
func (c *NoteClientWithCache) Invalidate(path string) {
	c.cache.Delete(path)
}

// InvalidateAll clears the entire cache.
func (c *NoteClientWithCache) InvalidateAll() {
	c.cache.Clear()
}

// GetCacheStats returns cache statistics.
func (c *NoteClientWithCache) GetCacheStats() Stats {
	return c.cache.Stats()
}

// CachedClient wraps the standard client interface with caching support.
type CachedClient struct {
	*obsidian.Client
	cache *Cache
}

// NewCachedClient creates a new client with caching enabled.
func NewCachedClient(client *obsidian.Client, maxSize int, ttl time.Duration) *CachedClient {
	return &CachedClient{
		Client: client,
		cache:  New(maxSize, ttl),
	}
}

// GetNote retrieves a note with caching.
func (c *CachedClient) GetNote(ctx context.Context, path string) (*obsidian.Note, error) {
	// Try cache first
	if note, found := c.cache.Get(path); found {
		return note, nil
	}

	// Fetch from API
	note, err := c.Client.GetNote(ctx, path)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(path, note)

	return note, nil
}

// UpdateNote updates a note and invalidates cache.
func (c *CachedClient) UpdateNote(ctx context.Context, path string, content string) error {
	err := c.Client.UpdateNote(ctx, path, content)
	if err != nil {
		return err
	}
	c.cache.Delete(path)
	return nil
}

// DeleteNote deletes a note and removes from cache.
func (c *CachedClient) DeleteNote(ctx context.Context, path string) error {
	err := c.Client.DeleteNote(ctx, path)
	if err != nil {
		return err
	}
	c.cache.Delete(path)
	return nil
}

// AppendNote appends to a note and invalidates cache.
func (c *CachedClient) AppendNote(ctx context.Context, path string, content string) error {
	err := c.Client.AppendNote(ctx, path, content)
	if err != nil {
		return err
	}
	c.cache.Delete(path)
	return nil
}

// PatchNote patches a note and invalidates cache.
func (c *CachedClient) PatchNote(ctx context.Context, path string, patch obsidian.PatchRequest) error {
	err := c.Client.PatchNote(ctx, path, patch)
	if err != nil {
		return err
	}
	c.cache.Delete(path)
	return nil
}

// CacheStats returns cache statistics.
func (c *CachedClient) CacheStats() Stats {
	return c.cache.Stats()
}

// InvalidateCache clears the entire cache.
func (c *CachedClient) InvalidateCache() {
	c.cache.Clear()
}

var _ = fmt.Sprintf // Placeholder to avoid import error if fmt not used
