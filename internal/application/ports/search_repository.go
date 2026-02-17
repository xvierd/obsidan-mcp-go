package ports

import (
	"context"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// SearchRepository defines the interface for search operations.
type SearchRepository interface {
	// Search performs a simple text search in the vault.
	Search(ctx context.Context, query string) ([]domain.SearchResult, error)

	// ComplexSearch performs a complex search using JsonLogic query.
	ComplexSearch(ctx context.Context, query map[string]interface{}) ([]domain.SearchResult, error)

	// DataviewQuery executes a Dataview query.
	DataviewQuery(ctx context.Context, query string) (*domain.DataviewResult, error)
}
