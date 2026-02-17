// Package tools provides MCP tool implementations for Obsidian operations.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xvierd/mcp-obsidian-go/internal/config"
	"github.com/xvierd/mcp-obsidian-go/internal/vector"
	"github.com/xvierd/mcp-obsidian-go/internal/vector/embeddings"
)

// VectorContext holds the vector store and embedding provider for tool handlers.
type VectorContext struct {
	Store    vector.VectorStore
	Embedder embeddings.EmbeddingProvider
}

// NewVectorContext creates a new vector context from configuration.
func NewVectorContext(cfg *config.VectorConfig) (*VectorContext, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("vector search is disabled")
	}

	// Create store
	store, err := vector.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}

	// Initialize store
	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		store.Close()
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	// Create embedder
	embedder, err := createEmbedderForTools(cfg)
	if err != nil {
		store.Close()
		return nil, err
	}

	return &VectorContext{
		Store:    store,
		Embedder: embedder,
	}, nil
}

// Close releases resources held by the vector context.
func (vc *VectorContext) Close() error {
	if vc.Embedder != nil {
		vc.Embedder.Close()
	}
	if vc.Store != nil {
		return vc.Store.Close()
	}
	return nil
}

// createEmbedderForTools creates an embedding provider for tools.
func createEmbedderForTools(cfg *config.VectorConfig) (embeddings.EmbeddingProvider, error) {
	// Get model path from environment or use default
	modelPath := os.Getenv("FASTTEXT_MODEL_PATH")
	if modelPath == "" {
		// Try to find in standard locations
		home, err := os.UserHomeDir()
		if err == nil {
			possiblePaths := []string{
				filepath.Join(home, ".mcp-obsidian", "models", "cc.en.300.vec"),
				filepath.Join(home, ".mcp-obsidian", "models", "fasttext.vec"),
				"models/cc.en.300.vec",
				"models/fasttext.vec",
			}
			for _, path := range possiblePaths {
				if _, err := os.Stat(path); err == nil {
					modelPath = path
					break
				}
			}
		}
	}

	if modelPath == "" {
		return nil, fmt.Errorf("fastText model not found. Set FASTTEXT_MODEL_PATH or run indexer download-model")
	}

	provider := embeddings.NewFastTextProvider(modelPath)
	if err := provider.Load(); err != nil {
		return nil, fmt.Errorf("failed to load fastText model: %w", err)
	}

	return provider, nil
}

// RegisterVectorTools registers all vector-related tools.
func RegisterVectorTools(r *Registry, vc *VectorContext) {
	if vc == nil {
		r.logger.Warn("vector context is nil, skipping vector tool registration")
		return
	}

	// vector_search tool
	r.Register(
		&Tool{
			Name:        "vector_search",
			Description: "Perform semantic search using vector embeddings. Finds notes similar in meaning to the query text.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "The search query text"
					},
					"limit": {
						"type": "integer",
						"description": "Maximum number of results to return (default: 10)",
						"default": 10
					}
				},
				"required": ["query"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Query string `json:"query"`
				Limit int    `json:"limit"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Query == "" {
				return nil, fmt.Errorf("query is required")
			}

			if args.Limit <= 0 {
				args.Limit = 10
			}

			// Generate embedding for query
			embedding, err := vc.Embedder.Embed(args.Query)
			if err != nil {
				return nil, fmt.Errorf("failed to embed query: %w", err)
			}

			// Search
			results, err := vc.Store.Search(ctx, embedding, args.Limit)
			if err != nil {
				return nil, fmt.Errorf("search failed: %w", err)
			}

			// Format results
			formattedResults := make([]map[string]interface{}, len(results))
			for i, r := range results {
				formattedResults[i] = map[string]interface{}{
					"note_path": r.NotePath,
					"content":   r.Chunk.Content,
					"score":     r.Score,
				}
			}

			return map[string]interface{}{
				"count":   len(results),
				"results": formattedResults,
			}, nil
		},
	)

	// find_similar_notes tool
	r.Register(
		&Tool{
			Name:        "find_similar_notes",
			Description: "Find notes that are semantically similar to a given note. Returns paths of similar notes.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note to find similar notes for"
					},
					"limit": {
						"type": "integer",
						"description": "Maximum number of similar notes to return (default: 5)",
						"default": 5
					}
				},
				"required": ["path"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path  string `json:"path"`
				Limit int    `json:"limit"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			if args.Limit <= 0 {
				args.Limit = 5
			}

			// Get the note content
			note, err := r.client.GetNote(ctx, args.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to get note: %w", err)
			}

			if note.Content == "" {
				return nil, fmt.Errorf("note has no content")
			}

			// Generate embedding for the note
			embedding, err := vc.Embedder.Embed(note.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to embed note: %w", err)
			}

			// Find similar notes
			similarPaths, err := vc.Store.FindSimilarNotes(ctx, embedding, args.Limit+1) // +1 to filter out the note itself
			if err != nil {
				return nil, fmt.Errorf("failed to find similar notes: %w", err)
			}

			// Filter out the source note
			var filteredPaths []string
			for _, path := range similarPaths {
				if path != args.Path {
					filteredPaths = append(filteredPaths, path)
				}
				if len(filteredPaths) >= args.Limit {
					break
				}
			}

			return map[string]interface{}{
				"source_note":    args.Path,
				"similar_notes":  filteredPaths,
				"count":          len(filteredPaths),
			}, nil
		},
	)

	// vector_status tool
	r.Register(
		&Tool{
			Name:        "vector_status",
			Description: "Get statistics about the vector search index including total notes, chunks, and last index time.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			stats, err := vc.Store.GetStats(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get stats: %w", err)
			}

			return map[string]interface{}{
				"total_notes":     stats.TotalNotes,
				"total_chunks":    stats.TotalChunks,
				"index_size_bytes": stats.IndexSize,
				"last_index_time": stats.LastIndexTime,
				"db_path":         vc.Store.GetDBPath(),
				"embedding_dim":   vc.Embedder.Dimension(),
			}, nil
		},
	)

	// search_note_semantic tool - search within a specific note
	r.Register(
		&Tool{
			Name:        "search_note_semantic",
			Description: "Perform semantic search within a specific note. Finds sections similar in meaning to the query.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note to search within"
					},
					"query": {
						"type": "string",
						"description": "The search query text"
					},
					"limit": {
						"type": "integer",
						"description": "Maximum number of results to return (default: 5)",
						"default": 5
					}
				},
				"required": ["path", "query"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path  string `json:"path"`
				Query string `json:"query"`
				Limit int    `json:"limit"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Path == "" {
				return nil, fmt.Errorf("path is required")
			}
			if args.Query == "" {
				return nil, fmt.Errorf("query is required")
			}

			if args.Limit <= 0 {
				args.Limit = 5
			}

			// Generate embedding for query
			embedding, err := vc.Embedder.Embed(args.Query)
			if err != nil {
				return nil, fmt.Errorf("failed to embed query: %w", err)
			}

			// Search within note
			results, err := vc.Store.SearchNote(ctx, args.Path, embedding, args.Limit)
			if err != nil {
				return nil, fmt.Errorf("search failed: %w", err)
			}

			// Format results
			formattedResults := make([]map[string]interface{}, len(results))
			for i, r := range results {
				formattedResults[i] = map[string]interface{}{
					"content": r.Chunk.Content,
					"score":   r.Score,
					"index":   r.Chunk.Index,
				}
			}

			return map[string]interface{}{
				"note_path": args.Path,
				"count":     len(results),
				"results":   formattedResults,
			}, nil
		},
	)
}
