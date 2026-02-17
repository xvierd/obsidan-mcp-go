package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// BatchReadResult represents the result of reading a single note in a batch operation
type BatchReadResult struct {
	Path        string                 `json:"path"`
	Content     string                 `json:"content,omitempty"`
	Frontmatter map[string]interface{} `json:"frontmatter,omitempty"`
	Found       bool                   `json:"found"`
	Error       string                 `json:"error,omitempty"`
}

// BatchReadError represents an error for a specific path
type BatchReadError struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// BatchNoteService defines the interface needed for batch operations
type BatchNoteService interface {
	GetNote(ctx context.Context, path string) (*domain.Note, error)
}

// batchReadNotes reads multiple notes in parallel using goroutines
func batchReadNotes(ctx context.Context, svc BatchNoteService, paths []string) ([]BatchReadResult, []BatchReadError) {
	var wg sync.WaitGroup
	results := make([]BatchReadResult, 0, len(paths))
	errors := make([]BatchReadError, 0)

	// Channel for collecting results
	resultChan := make(chan BatchReadResult, len(paths))
	errorChan := make(chan BatchReadError, len(paths))

	// Process each path in parallel
	for _, path := range paths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			// Check context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			note, err := svc.GetNote(ctx, p)
			if err != nil {
				select {
				case errorChan <- BatchReadError{
					Path:  p,
					Error: err.Error(),
				}:
				case <-ctx.Done():
				}
				return
			}

			select {
			case resultChan <- BatchReadResult{
				Path:        note.Path,
				Content:     note.Content,
				Frontmatter: note.Frontmatter,
				Found:       true,
			}:
			case <-ctx.Done():
			}
		}(path)
	}

	// Close channels when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	for batchErr := range errorChan {
		errors = append(errors, batchErr)
	}

	return results, errors
}

// RegisterBatchTools registers all batch operation tools.
func RegisterBatchTools(r *Registry) {
	// batch_read_notes tool
	r.Register(
		&Tool{
			Name:        "batch_read_notes",
			Description: "Read multiple notes in parallel using goroutines for improved performance",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"paths": {
						"type": "array",
						"items": {
							"type": "string"
						},
						"description": "Array of note paths to read"
					}
				},
				"required": ["paths"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Paths []string `json:"paths"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if len(args.Paths) == 0 {
				return nil, fmt.Errorf("paths array is required")
			}

			// Limit batch size to prevent overwhelming the server
			if len(args.Paths) > 100 {
				return nil, fmt.Errorf("batch size exceeds maximum of 100")
			}

			results, errors := batchReadNotes(ctx, r.GetNoteService(), args.Paths)

			return map[string]interface{}{
				"total":   len(args.Paths),
				"success": len(results),
				"failed":  len(errors),
				"results": results,
				"errors":  errors,
			}, nil
		},
	)
}
