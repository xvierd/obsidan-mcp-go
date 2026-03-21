package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/application/services"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// Mock implementations for testing
type mockNoteRepository struct{}

func (m *mockNoteRepository) GetNote(ctx context.Context, path string) (*domain.Note, error) {
	return nil, nil
}
func (m *mockNoteRepository) ListNotes(ctx context.Context, directory string) ([]string, error) {
	return []string{}, nil
}
func (m *mockNoteRepository) CreateNote(ctx context.Context, path string, content string) error {
	return nil
}
func (m *mockNoteRepository) UpdateNote(ctx context.Context, path string, content string) error {
	return nil
}
func (m *mockNoteRepository) DeleteNote(ctx context.Context, path string) error {
	return nil
}
func (m *mockNoteRepository) AppendNote(ctx context.Context, path string, content string) error {
	return nil
}
func (m *mockNoteRepository) PatchNote(ctx context.Context, path string, patch domain.PatchRequest) error {
	return nil
}
func (m *mockNoteRepository) OpenNote(ctx context.Context, path string) error {
	return nil
}
func (m *mockNoteRepository) GetRecentChanges(ctx context.Context, limit int) ([]domain.RecentChange, error) {
	return []domain.RecentChange{}, nil
}
func (m *mockNoteRepository) GetPeriodicNote(ctx context.Context, period string, offset int) (*domain.Note, error) {
	return nil, nil
}

type mockActiveNoteRepository struct{}

func (m *mockActiveNoteRepository) GetActiveNote(ctx context.Context) (*domain.Note, error) {
	return nil, nil
}
func (m *mockActiveNoteRepository) UpdateActiveNote(ctx context.Context, content string) error {
	return nil
}
func (m *mockActiveNoteRepository) AppendActiveNote(ctx context.Context, content string) error {
	return nil
}
func (m *mockActiveNoteRepository) DeleteActiveNote(ctx context.Context) error {
	return nil
}
func (m *mockActiveNoteRepository) PatchActiveNote(ctx context.Context, patch domain.PatchRequest) error {
	return nil
}

type mockSearchRepository struct{}

func (m *mockSearchRepository) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	return []domain.SearchResult{}, nil
}
func (m *mockSearchRepository) ComplexSearch(ctx context.Context, query map[string]interface{}) ([]domain.SearchResult, error) {
	return []domain.SearchResult{}, nil
}
func (m *mockSearchRepository) DataviewQuery(ctx context.Context, query string) (*domain.DataviewResult, error) {
	return nil, nil
}

type mockCommandRepository struct{}

func (m *mockCommandRepository) ListCommands(ctx context.Context) ([]domain.Command, error) {
	return []domain.Command{}, nil
}
func (m *mockCommandRepository) ExecuteCommand(ctx context.Context, commandID string) error {
	return nil
}

func TestRegisterVaultTools(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create mock services
	noteService := services.NewNoteService(
		&mockNoteRepository{},
		nil,
		&mockActiveNoteRepository{},
		logger,
	)
	searchService := services.NewSearchService(&mockSearchRepository{}, logger)
	cmdService := services.NewCommandService(&mockCommandRepository{}, logger)

	statusService := services.NewStatusService(&mockStatusRepo{}, logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	RegisterVaultTools(registry)

	// Check that all vault tools are registered
	expectedTools := []string{
		"list_notes",
		"read_note",
		"create_note",
		"update_note",
		"append_note",
		"delete_note",
		"patch_note",
	}

	for _, toolName := range expectedTools {
		_, found := registry.GetTool(toolName)
		if !found {
			t.Errorf("expected tool %s to be registered", toolName)
		}
	}
}
