package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// Mock search repository for testing
type mockSearchRepository struct {
	searchResults        []domain.SearchResult
	complexSearchResults []domain.SearchResult
	dataviewResult       *domain.DataviewResult
	searchErr            error
	complexSearchErr     error
	dataviewErr          error
}

func (m *mockSearchRepository) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.searchResults, nil
}

func (m *mockSearchRepository) ComplexSearch(ctx context.Context, query map[string]interface{}) ([]domain.SearchResult, error) {
	if m.complexSearchErr != nil {
		return nil, m.complexSearchErr
	}
	return m.complexSearchResults, nil
}

func (m *mockSearchRepository) DataviewQuery(ctx context.Context, query string) (*domain.DataviewResult, error) {
	if m.dataviewErr != nil {
		return nil, m.dataviewErr
	}
	return m.dataviewResult, nil
}

func TestNewSearchService(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{}

	service := NewSearchService(repo, logger)

	if service == nil {
		t.Fatal("expected SearchService to be created")
	}
	if service.repo != repo {
		t.Error("expected repository to be set")
	}
}

func TestSearchServiceSearch_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedResults := []domain.SearchResult{
		{Filename: "note1.md"},
		{Filename: "note2.md"},
	}
	repo := &mockSearchRepository{
		searchResults: expectedResults,
	}
	service := NewSearchService(repo, logger)

	results, err := service.Search(context.Background(), "test query")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != len(expectedResults) {
		t.Errorf("expected %d results, got %d", len(expectedResults), len(results))
	}
}

func TestSearchServiceSearch_EmptyQuery(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{}
	service := NewSearchService(repo, logger)

	_, err := service.Search(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty query")
	}
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Errorf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestSearchServiceSearch_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{
		searchErr: errors.New("search failed"),
	}
	service := NewSearchService(repo, logger)

	_, err := service.Search(context.Background(), "test")

	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "search failed" {
		t.Errorf("expected 'search failed' error, got %v", err)
	}
}

func TestSearchServiceComplexSearch_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedResults := []domain.SearchResult{
		{Filename: "result1.md"},
		{Filename: "result2.md"},
		{Filename: "result3.md"},
	}
	repo := &mockSearchRepository{
		complexSearchResults: expectedResults,
	}
	service := NewSearchService(repo, logger)

	query := map[string]interface{}{
		"and": []interface{}{
			map[string]interface{}{"==": []interface{}{"{{path}}", "*.md"}},
		},
	}
	results, err := service.ComplexSearch(context.Background(), query)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != len(expectedResults) {
		t.Errorf("expected %d results, got %d", len(expectedResults), len(results))
	}
}

func TestSearchServiceComplexSearch_NilQuery(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{}
	service := NewSearchService(repo, logger)

	_, err := service.ComplexSearch(context.Background(), nil)

	if err == nil {
		t.Error("expected error for nil query")
	}
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Errorf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestSearchServiceComplexSearch_EmptyQuery(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{}
	service := NewSearchService(repo, logger)

	// Empty map is still valid for complex search
	query := map[string]interface{}{}
	_, err := service.ComplexSearch(context.Background(), query)

	// Empty query should not error - it's passed to the repository
	if err != nil {
		t.Errorf("unexpected error for empty query: %v", err)
	}
}

func TestSearchServiceComplexSearch_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{
		complexSearchErr: errors.New("complex search failed"),
	}
	service := NewSearchService(repo, logger)

	query := map[string]interface{}{
		"==": []interface{}{"{{path}}", "*.md"},
	}
	_, err := service.ComplexSearch(context.Background(), query)

	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "complex search failed" {
		t.Errorf("expected 'complex search failed' error, got %v", err)
	}
}

func TestSearchServiceDataviewQuery_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedResult := &domain.DataviewResult{
		Headers: []string{"Name", "Due", "Completed"},
		Rows: [][]interface{}{
			{"Task 1", "2026-02-16", true},
			{"Task 2", "2026-02-17", false},
		},
		Count: 2,
	}
	repo := &mockSearchRepository{
		dataviewResult: expectedResult,
	}
	service := NewSearchService(repo, logger)

	result, err := service.DataviewQuery(context.Background(), "TABLE Name, Due, Completed FROM \"Tasks\"")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Count != expectedResult.Count {
		t.Errorf("expected count %d, got %d", expectedResult.Count, result.Count)
	}
	if len(result.Headers) != len(expectedResult.Headers) {
		t.Errorf("expected %d headers, got %d", len(expectedResult.Headers), len(result.Headers))
	}
}

func TestSearchServiceDataviewQuery_EmptyQuery(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{}
	service := NewSearchService(repo, logger)

	_, err := service.DataviewQuery(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty query")
	}
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Errorf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestSearchServiceDataviewQuery_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{
		dataviewErr: errors.New("dataview query failed"),
	}
	service := NewSearchService(repo, logger)

	_, err := service.DataviewQuery(context.Background(), "TABLE Name FROM \"Tasks\"")

	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "dataview query failed" {
		t.Errorf("expected 'dataview query failed' error, got %v", err)
	}
}

func TestSearchServiceSearch_NoResults(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockSearchRepository{
		searchResults: []domain.SearchResult{},
	}
	service := NewSearchService(repo, logger)

	results, err := service.Search(context.Background(), "nonexistent term")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected empty slice, not nil")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
