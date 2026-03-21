package ports

import (
	"time"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// CacheRepository defines the interface for caching operations.
type CacheRepository interface {
	// Get retrieves a note from the cache.
	Get(key string) (*domain.Note, bool)

	// Set adds or updates a note in the cache.
	Set(key string, value *domain.Note)

	// Delete removes a note from the cache.
	Delete(key string)

	// Clear removes all items from the cache.
	Clear()

	// Stats returns cache statistics.
	Stats() CacheStats
}

// CacheStats holds cache statistics.
type CacheStats struct {
	Size    int           `json:"size"`
	MaxSize int           `json:"max_size"`
	Hits    int64         `json:"hits"`
	Misses  int64         `json:"misses"`
	HitRate float64       `json:"hit_rate"`
	TTL     time.Duration `json:"ttl"`
}
