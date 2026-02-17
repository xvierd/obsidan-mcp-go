package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/application/services"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// Mock implementations for registry tests
type mockNoteRepo struct{}

func (m *mockNoteRepo) GetNote(ctx context.Context, path string) (*domain.Note, error) {
	return nil, nil
}
func (m *mockNoteRepo) ListNotes(ctx context.Context, directory string) ([]string, error) {
	return []string{}, nil
}
func (m *mockNoteRepo) CreateNote(ctx context.Context, path string, content string) error {
	return nil
}
func (m *mockNoteRepo) UpdateNote(ctx context.Context, path string, content string) error {
	return nil
}
func (m *mockNoteRepo) DeleteNote(ctx context.Context, path string) error {
	return nil
}
func (m *mockNoteRepo) AppendNote(ctx context.Context, path string, content string) error {
	return nil
}
func (m *mockNoteRepo) PatchNote(ctx context.Context, path string, patch domain.PatchRequest) error {
	return nil
}
func (m *mockNoteRepo) OpenNote(ctx context.Context, path string) error {
	return nil
}
func (m *mockNoteRepo) GetRecentChanges(ctx context.Context, limit int) ([]domain.RecentChange, error) {
	return []domain.RecentChange{}, nil
}
func (m *mockNoteRepo) GetPeriodicNote(ctx context.Context, period string, offset int) (*domain.Note, error) {
	return nil, nil
}

type mockActiveNoteRepo struct{}

func (m *mockActiveNoteRepo) GetActiveNote(ctx context.Context) (*domain.Note, error) {
	return nil, nil
}
func (m *mockActiveNoteRepo) UpdateActiveNote(ctx context.Context, content string) error {
	return nil
}
func (m *mockActiveNoteRepo) AppendActiveNote(ctx context.Context, content string) error {
	return nil
}
func (m *mockActiveNoteRepo) DeleteActiveNote(ctx context.Context) error {
	return nil
}
func (m *mockActiveNoteRepo) PatchActiveNote(ctx context.Context, patch domain.PatchRequest) error {
	return nil
}

type mockSearchRepo struct{}

func (m *mockSearchRepo) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	return []domain.SearchResult{}, nil
}
func (m *mockSearchRepo) ComplexSearch(ctx context.Context, query map[string]interface{}) ([]domain.SearchResult, error) {
	return []domain.SearchResult{}, nil
}
func (m *mockSearchRepo) DataviewQuery(ctx context.Context, query string) (*domain.DataviewResult, error) {
	return nil, nil
}

type mockStatusRepo struct{}

func (m *mockStatusRepo) ServerStatus(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "ok"}, nil
}

type mockCommandRepo struct{}

func (m *mockCommandRepo) ListCommands(ctx context.Context) ([]domain.Command, error) {
	return []domain.Command{}, nil
}
func (m *mockCommandRepo) ExecuteCommand(ctx context.Context, commandID string) error {
	return nil
}

func createTestServices(logger *slog.Logger) (*services.NoteService, *services.SearchService, *services.CommandService, *services.StatusService) {
	noteService := services.NewNoteService(
		&mockNoteRepo{},
		nil,
		&mockActiveNoteRepo{},
		logger,
	)
	searchService := services.NewSearchService(&mockSearchRepo{}, logger)
	cmdService := services.NewCommandService(&mockCommandRepo{}, logger)
	statusService := services.NewStatusService(&mockStatusRepo{}, logger)
	return noteService, searchService, cmdService, statusService
}

func TestNewRegistry(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)

	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	if registry == nil {
		t.Fatal("expected registry to be created")
	}

	if registry.logger != logger {
		t.Error("expected logger to be set")
	}

	if registry.noteService != noteService {
		t.Error("expected noteService to be set")
	}

	if len(registry.tools) != 0 {
		t.Errorf("expected empty tools map, got %d", len(registry.tools))
	}
}

func TestRegistryRegister(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: json.RawMessage(`{"type": "object"}`),
	}

	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "test result", nil
	}

	registry.Register(tool, handler)

	// Check tool is registered
	if len(registry.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(registry.tools))
	}

	// Check handler is registered
	if len(registry.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(registry.handlers))
	}

	// Get the tool
	retrievedTool, found := registry.GetTool("test_tool")
	if !found {
		t.Error("expected to find registered tool")
	}

	if retrievedTool.Name != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got %s", retrievedTool.Name)
	}

	// Get the handler
	retrievedHandler, found := registry.GetHandler("test_tool")
	if !found {
		t.Error("expected to find registered handler")
	}

	// Execute handler
	result, err := retrievedHandler(context.Background(), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != "test result" {
		t.Errorf("expected 'test result', got %v", result)
	}
}

func TestRegistryListTools(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	// Register multiple tools
	for i := 0; i < 3; i++ {
		tool := &Tool{
			Name:        fmt.Sprintf("tool_%d", i),
			Description: "Test tool",
			InputSchema: json.RawMessage(`{"type": "object"}`),
		}
		registry.Register(tool, func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			return nil, nil
		})
	}

	tools := registry.ListTools()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}
}

func TestRegistryExecute(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	tool := &Tool{
		Name:        "execute_test",
		Description: "Test execution",
		InputSchema: json.RawMessage(`{"type": "object"}`),
	}

	registry.Register(tool, func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		var args map[string]string
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, err
		}
		return args["value"], nil
	})

	params := json.RawMessage(`{"value": "hello"}`)
	result, err := registry.Execute(context.Background(), "execute_test", params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
}

func TestRegistryExecuteNotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	_, err := registry.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error for non-existent tool")
	}

	if err.Error() != "tool not found: nonexistent" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRegistryGetters(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	if registry.GetNoteService() != noteService {
		t.Error("GetNoteService returned wrong service")
	}

	if registry.GetSearchService() != searchService {
		t.Error("GetSearchService returned wrong service")
	}

	if registry.GetCommandService() != cmdService {
		t.Error("GetCommandService returned wrong service")
	}

	if registry.GetLogger() != logger {
		t.Error("GetLogger returned wrong logger")
	}
}

func TestGetToolNotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	_, found := registry.GetTool("nonexistent")
	if found {
		t.Error("expected not to find non-existent tool")
	}
}

func TestGetHandlerNotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	noteService, searchService, cmdService, statusService := createTestServices(logger)
	registry := NewRegistry(logger, noteService, searchService, cmdService, statusService)

	_, found := registry.GetHandler("nonexistent")
	if found {
		t.Error("expected not to find non-existent handler")
	}
}
