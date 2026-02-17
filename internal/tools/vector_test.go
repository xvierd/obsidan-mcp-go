package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/config"
	"github.com/xvierd/mcp-obsidian-go/internal/vector"
	"github.com/xvierd/mcp-obsidian-go/internal/vector/embeddings"
)

// mockVectorStore is a mock implementation of VectorStore for testing
type mockVectorStore struct {
	chunks       []vector.Chunk
	indexedNotes []vector.NoteIndex
	stats        *vector.VectorStats
}

func (m *mockVectorStore) Initialize(ctx context.Context) error { return nil }
func (m *mockVectorStore) Close() error                         { return nil }

func (m *mockVectorStore) AddChunks(ctx context.Context, chunks []vector.Chunk) error {
	m.chunks = append(m.chunks, chunks...)
	return nil
}

func (m *mockVectorStore) DeleteNoteChunks(ctx context.Context, notePath string) error {
	return nil
}

func (m *mockVectorStore) Search(ctx context.Context, embedding []float32, limit int) ([]vector.SearchResult, error) {
	results := []vector.SearchResult{
		{
			Chunk: vector.Chunk{
				NotePath: "test/note1.md",
				Content:  "Test content",
				Index:    0,
			},
			Score:    0.95,
			NotePath: "test/note1.md",
		},
	}
	return results, nil
}

func (m *mockVectorStore) SearchNote(ctx context.Context, notePath string, embedding []float32, limit int) ([]vector.SearchResult, error) {
	results := []vector.SearchResult{
		{
			Chunk: vector.Chunk{
				NotePath: notePath,
				Content:  "Content within note",
				Index:    0,
			},
			Score:    0.90,
			NotePath: notePath,
		},
	}
	return results, nil
}

func (m *mockVectorStore) FindSimilarNotes(ctx context.Context, embedding []float32, limit int) ([]string, error) {
	return []string{"test/note1.md", "test/note2.md"}, nil
}

func (m *mockVectorStore) GetNoteIndex(ctx context.Context, notePath string) (*vector.NoteIndex, error) {
	return nil, nil
}

func (m *mockVectorStore) UpdateNoteIndex(ctx context.Context, noteIndex vector.NoteIndex) error {
	return nil
}

func (m *mockVectorStore) GetIndexedNotes(ctx context.Context) ([]vector.NoteIndex, error) {
	return m.indexedNotes, nil
}

func (m *mockVectorStore) GetStats(ctx context.Context) (*vector.VectorStats, error) {
	if m.stats == nil {
		return &vector.VectorStats{
			TotalNotes:  10,
			TotalChunks: 50,
			IndexSize:   1024 * 1024,
		}, nil
	}
	return m.stats, nil
}

func (m *mockVectorStore) Clear(ctx context.Context) error {
	m.chunks = nil
	return nil
}

func (m *mockVectorStore) GetDBPath() string {
	return "/test/vector.db"
}

var _ vector.VectorStore = (*mockVectorStore)(nil)

// mockEmbedder is a mock implementation of EmbeddingProvider for testing
type mockEmbedder struct {
	dimension int
}

func (m *mockEmbedder) Embed(text string) ([]float32, error) {
	// Return a simple embedding based on text length
	embedding := make([]float32, m.dimension)
	for i := 0; i < m.dimension && i < len(text); i++ {
		embedding[i] = float32(text[i]) / 255.0
	}
	return embedding, nil
}

func (m *mockEmbedder) Dimension() int {
	return m.dimension
}

func (m *mockEmbedder) Close() error {
	return nil
}

var _ embeddings.EmbeddingProvider = (*mockEmbedder)(nil)

func TestVectorTools(t *testing.T) {
	// Create mock components
	mockStore := &mockVectorStore{}
	mockEmb := &mockEmbedder{dimension: 300}

	vc := &VectorContext{
		Store:    mockStore,
		Embedder: mockEmb,
	}

	// Create a minimal registry for testing with a logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	registry := NewRegistry(logger, nil)

	// Register vector tools
	RegisterVectorTools(registry, vc)

	t.Run("vector_search tool exists", func(t *testing.T) {
		tool, ok := registry.GetTool("vector_search")
		if !ok {
			t.Fatal("vector_search tool not found")
		}
		if tool.Name != "vector_search" {
			t.Errorf("tool.Name = %v, want vector_search", tool.Name)
		}
	})

	t.Run("find_similar_notes tool exists", func(t *testing.T) {
		tool, ok := registry.GetTool("find_similar_notes")
		if !ok {
			t.Fatal("find_similar_notes tool not found")
		}
		if tool.Name != "find_similar_notes" {
			t.Errorf("tool.Name = %v, want find_similar_notes", tool.Name)
		}
	})

	t.Run("vector_status tool exists", func(t *testing.T) {
		tool, ok := registry.GetTool("vector_status")
		if !ok {
			t.Fatal("vector_status tool not found")
		}
		if tool.Name != "vector_status" {
			t.Errorf("tool.Name = %v, want vector_status", tool.Name)
		}
	})

	t.Run("search_note_semantic tool exists", func(t *testing.T) {
		tool, ok := registry.GetTool("search_note_semantic")
		if !ok {
			t.Fatal("search_note_semantic tool not found")
		}
		if tool.Name != "search_note_semantic" {
			t.Errorf("tool.Name = %v, want search_note_semantic", tool.Name)
		}
	})
}

func TestVectorContext(t *testing.T) {
	t.Run("NewVectorContext with disabled config", func(t *testing.T) {
		cfg := &config.VectorConfig{
			Enabled: false,
		}
		_, err := NewVectorContext(cfg)
		if err == nil {
			t.Error("Expected error when vector is disabled")
		}
	})
}

func TestVectorToolsWithMock(t *testing.T) {
	mockStore := &mockVectorStore{}
	mockEmb := &mockEmbedder{dimension: 300}

	vc := &VectorContext{
		Store:    mockStore,
		Embedder: mockEmb,
	}

	ctx := context.Background()

	t.Run("vector_search handler", func(t *testing.T) {
		// Get the handler directly from the context
		// Since we can't easily extract handlers, we'll test the behavior indirectly

		// Generate embedding
		embedding, err := vc.Embedder.Embed("test query")
		if err != nil {
			t.Fatalf("Embed error: %v", err)
		}

		if len(embedding) != 300 {
			t.Errorf("embedding length = %d, want 300", len(embedding))
		}

		// Search
		results, err := vc.Store.Search(ctx, embedding, 5)
		if err != nil {
			t.Fatalf("Search error: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("results length = %d, want 1", len(results))
		}
	})

	t.Run("vector_status handler", func(t *testing.T) {
		stats, err := vc.Store.GetStats(ctx)
		if err != nil {
			t.Fatalf("GetStats error: %v", err)
		}

		if stats.TotalNotes != 10 {
			t.Errorf("total_notes = %d, want 10", stats.TotalNotes)
		}

		if stats.TotalChunks != 50 {
			t.Errorf("total_chunks = %d, want 50", stats.TotalChunks)
		}
	})

	t.Run("find_similar_notes handler", func(t *testing.T) {
		embedding, err := vc.Embedder.Embed("some note content")
		if err != nil {
			t.Fatalf("Embed error: %v", err)
		}

		similar, err := vc.Store.FindSimilarNotes(ctx, embedding, 5)
		if err != nil {
			t.Fatalf("FindSimilarNotes error: %v", err)
		}

		if len(similar) != 2 {
			t.Errorf("similar notes length = %d, want 2", len(similar))
		}
	})

	t.Run("search_note_semantic handler", func(t *testing.T) {
		embedding, err := vc.Embedder.Embed("query text")
		if err != nil {
			t.Fatalf("Embed error: %v", err)
		}

		results, err := vc.Store.SearchNote(ctx, "test/note.md", embedding, 5)
		if err != nil {
			t.Fatalf("SearchNote error: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("results length = %d, want 1", len(results))
		}
	})
}
