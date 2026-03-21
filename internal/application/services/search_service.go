package services

import (
	"context"
	"log/slog"

	"github.com/xvierd/mcp-obsidian-go/internal/application/ports"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// SearchService provides business logic for search operations.
type SearchService struct {
	repo   ports.SearchRepository
	logger *slog.Logger
}

// NewSearchService creates a new SearchService.
func NewSearchService(repo ports.SearchRepository, logger *slog.Logger) *SearchService {
	return &SearchService{
		repo:   repo,
		logger: logger,
	}
}

// Search performs a simple text search in the vault.
func (s *SearchService) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	if query == "" {
		return nil, domain.ErrInvalidRequest
	}

	results, err := s.repo.Search(ctx, query)
	if err != nil {
		s.logger.Error("search failed", "query", query, "error", err)
		return nil, err
	}

	s.logger.Info("search completed", "query", query, "results", len(results))
	return results, nil
}

// ComplexSearch performs a complex search using JsonLogic query.
func (s *SearchService) ComplexSearch(ctx context.Context, query map[string]interface{}) ([]domain.SearchResult, error) {
	if query == nil {
		return nil, domain.ErrInvalidRequest
	}

	results, err := s.repo.ComplexSearch(ctx, query)
	if err != nil {
		s.logger.Error("complex search failed", "error", err)
		return nil, err
	}

	s.logger.Info("complex search completed", "results", len(results))
	return results, nil
}

// DataviewQuery executes a Dataview query.
func (s *SearchService) DataviewQuery(ctx context.Context, query string) (*domain.DataviewResult, error) {
	if query == "" {
		return nil, domain.ErrInvalidRequest
	}

	result, err := s.repo.DataviewQuery(ctx, query)
	if err != nil {
		s.logger.Error("dataview query failed", "query", query, "error", err)
		return nil, err
	}

	s.logger.Info("dataview query completed", "rows", result.Count)
	return result, nil
}
