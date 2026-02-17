package cache

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/xvierd/mcp-obsidian-go/internal/obsidian"
)

func TestNew(t *testing.T) {
	c := New(100, 30*time.Second)

	if c == nil {
		t.Fatal("expected cache to be created")
	}

	if c.maxSize != 100 {
		t.Errorf("expected maxSize 100, got %d", c.maxSize)
	}

	if c.ttl != 30*time.Second {
		t.Errorf("expected ttl 30s, got %v", c.ttl)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := New(10, time.Hour)

	note := &obsidian.Note{
		Path:    "test.md",
		Content: "# Test",
	}

	// Set the note
	c.Set("test.md", note)

	// Get the note
	result, found := c.Get("test.md")

	if !found {
		t.Error("expected to find note in cache")
	}

	if result.Path != "test.md" {
		t.Errorf("expected path 'test.md', got %s", result.Path)
	}
}

func TestCacheGetNotFound(t *testing.T) {
	c := New(10, time.Hour)

	_, found := c.Get("nonexistent.md")

	if found {
		t.Error("expected not to find note in cache")
	}
}

func TestCacheTTLExpiration(t *testing.T) {
	c := New(10, 50*time.Millisecond)

	note := &obsidian.Note{
		Path:    "test.md",
		Content: "# Test",
	}

	c.Set("test.md", note)

	// Should find immediately
	_, found := c.Get("test.md")
	if !found {
		t.Error("expected to find note before TTL expiration")
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should not find after TTL
	_, found = c.Get("test.md")
	if found {
		t.Error("expected note to be expired after TTL")
	}
}

func TestCacheEviction(t *testing.T) {
	c := New(3, time.Hour)

	// Add 3 items (at capacity)
	c.Set("note1.md", &obsidian.Note{Path: "note1.md"})
	c.Set("note2.md", &obsidian.Note{Path: "note2.md"})
	c.Set("note3.md", &obsidian.Note{Path: "note3.md"})

	// Add one more (should evict oldest)
	c.Set("note4.md", &obsidian.Note{Path: "note4.md"})

	// note1 should be evicted (LRU)
	_, found := c.Get("note1.md")
	if found {
		t.Error("expected note1.md to be evicted")
	}

	// note2, note3, note4 should still be there
	for _, path := range []string{"note2.md", "note3.md", "note4.md"} {
		_, found := c.Get(path)
		if !found {
			t.Errorf("expected %s to be in cache", path)
		}
	}
}

func TestCacheUpdate(t *testing.T) {
	c := New(10, time.Hour)

	note1 := &obsidian.Note{
		Path:    "test.md",
		Content: "Original content",
	}
	c.Set("test.md", note1)

	note2 := &obsidian.Note{
		Path:    "test.md",
		Content: "Updated content",
	}
	c.Set("test.md", note2)

	result, found := c.Get("test.md")
	if !found {
		t.Fatal("expected to find note")
	}

	if result.Content != "Updated content" {
		t.Errorf("expected updated content, got %s", result.Content)
	}
}

func TestCacheDelete(t *testing.T) {
	c := New(10, time.Hour)

	c.Set("test.md", &obsidian.Note{Path: "test.md"})
	c.Delete("test.md")

	_, found := c.Get("test.md")
	if found {
		t.Error("expected note to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	c := New(10, time.Hour)

	c.Set("note1.md", &obsidian.Note{Path: "note1.md"})
	c.Set("note2.md", &obsidian.Note{Path: "note2.md"})

	c.Clear()

	_, found := c.Get("note1.md")
	if found {
		t.Error("expected note1.md to be cleared")
	}

	_, found = c.Get("note2.md")
	if found {
		t.Error("expected note2.md to be cleared")
	}

	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("expected size 0 after clear, got %d", stats.Size)
	}
}

func TestCacheStats(t *testing.T) {
	c := New(100, time.Hour)

	// Initial stats
	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("expected initial size 0, got %d", stats.Size)
	}
	if stats.Hits != 0 {
		t.Errorf("expected initial hits 0, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("expected initial misses 0, got %d", stats.Misses)
	}

	// Add items and access them
	c.Set("test1.md", &obsidian.Note{Path: "test1.md"})
	c.Set("test2.md", &obsidian.Note{Path: "test2.md"})

	c.Get("test1.md")   // hit
	c.Get("test2.md")   // hit
	c.Get("missing.md") // miss

	stats = c.Stats()
	if stats.Size != 2 {
		t.Errorf("expected size 2, got %d", stats.Size)
	}
	if stats.Hits != 2 {
		t.Errorf("expected hits 2, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected misses 1, got %d", stats.Misses)
	}

	// Check hit rate with tolerance for floating point
	expectedHitRate := 2.0 / 3.0
	if math.Abs(stats.HitRate-expectedHitRate) > 0.0001 {
		t.Errorf("expected hit rate %f, got %f", expectedHitRate, stats.HitRate)
	}
}

func TestCacheHitRate(t *testing.T) {
	c := New(10, time.Hour)

	// No accesses yet
	rate := c.hitRate()
	if rate != 0 {
		t.Errorf("expected hit rate 0 with no accesses, got %f", rate)
	}

	// All misses
	c.Get("missing1.md")
	c.Get("missing2.md")

	rate = c.hitRate()
	if rate != 0 {
		t.Errorf("expected hit rate 0 with all misses, got %f", rate)
	}

	// Add and access
	c.Set("test.md", &obsidian.Note{Path: "test.md"})
	c.Get("test.md")

	rate = c.hitRate()
	expectedRate := 1.0 / 3.0
	if math.Abs(rate-expectedRate) > 0.0001 {
		t.Errorf("expected hit rate %f, got %f", expectedRate, rate)
	}
}

func TestCacheConcurrency(t *testing.T) {
	c := New(100, time.Hour)

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			path := fmt.Sprintf("note%d.md", n)
			c.Set(path, &obsidian.Note{Path: path})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all items are present
	stats := c.Stats()
	if stats.Size != 10 {
		t.Errorf("expected size 10 after concurrent writes, got %d", stats.Size)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(n int) {
			path := fmt.Sprintf("note%d.md", n)
			c.Get(path)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
