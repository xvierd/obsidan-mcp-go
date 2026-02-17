package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// Mock command repository for testing
type mockCommandRepository struct {
	commands      []domain.Command
	executeErr    error
	listErr       error
	executedCmdID string
}

func (m *mockCommandRepository) ListCommands(ctx context.Context) ([]domain.Command, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.commands, nil
}

func (m *mockCommandRepository) ExecuteCommand(ctx context.Context, commandID string) error {
	m.executedCmdID = commandID
	return m.executeErr
}

func TestNewCommandService(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{}

	service := NewCommandService(repo, logger)

	if service == nil {
		t.Fatal("expected CommandService to be created")
	}
	if service.repo != repo {
		t.Error("expected repository to be set")
	}
}

func TestCommandServiceListCommands_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedCommands := []domain.Command{
		{ID: "app:open-settings", Name: "Open Settings"},
		{ID: "editor:save-file", Name: "Save Current File"},
		{ID: "workspace:close", Name: "Close Workspace"},
	}
	repo := &mockCommandRepository{
		commands: expectedCommands,
	}
	service := NewCommandService(repo, logger)

	commands, err := service.ListCommands(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(commands) != len(expectedCommands) {
		t.Errorf("expected %d commands, got %d", len(expectedCommands), len(commands))
	}
	for i, cmd := range commands {
		if cmd.ID != expectedCommands[i].ID {
			t.Errorf("expected command ID %s, got %s", expectedCommands[i].ID, cmd.ID)
		}
		if cmd.Name != expectedCommands[i].Name {
			t.Errorf("expected command name %s, got %s", expectedCommands[i].Name, cmd.Name)
		}
	}
}

func TestCommandServiceListCommands_Empty(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{
		commands: []domain.Command{},
	}
	service := NewCommandService(repo, logger)

	commands, err := service.ListCommands(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if commands == nil {
		t.Error("expected empty slice, not nil")
	}
	if len(commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(commands))
	}
}

func TestCommandServiceListCommands_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{
		listErr: errors.New("failed to list commands"),
	}
	service := NewCommandService(repo, logger)

	_, err := service.ListCommands(context.Background())

	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "failed to list commands" {
		t.Errorf("expected 'failed to list commands' error, got %v", err)
	}
}

func TestCommandServiceExecuteCommand_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{}
	service := NewCommandService(repo, logger)

	commandID := "editor:save-file"
	err := service.ExecuteCommand(context.Background(), commandID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if repo.executedCmdID != commandID {
		t.Errorf("expected command ID %s to be executed, got %s", commandID, repo.executedCmdID)
	}
}

func TestCommandServiceExecuteCommand_EmptyID(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{}
	service := NewCommandService(repo, logger)

	err := service.ExecuteCommand(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty command ID")
	}
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Errorf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestCommandServiceExecuteCommand_WhitespaceID(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{}
	service := NewCommandService(repo, logger)

	// Service doesn't trim whitespace, so this would be passed to repository
	// The validation currently only checks for empty string
	err := service.ExecuteCommand(context.Background(), "   ")

	// Current implementation only checks for empty string
	// This test documents the current behavior
	if err != nil {
		t.Logf("Note: whitespace-only command ID handling: %v", err)
	}
}

func TestCommandServiceExecuteCommand_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{
		executeErr: errors.New("command not found"),
	}
	service := NewCommandService(repo, logger)

	commandID := "nonexistent:command"
	err := service.ExecuteCommand(context.Background(), commandID)

	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "command not found" {
		t.Errorf("expected 'command not found' error, got %v", err)
	}
	if repo.executedCmdID != commandID {
		t.Errorf("expected command ID %s to be attempted, got %s", commandID, repo.executedCmdID)
	}
}

func TestCommandServiceExecuteCommand_ContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{
		executeErr: context.Canceled,
	}
	service := NewCommandService(repo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := service.ExecuteCommand(ctx, "editor:save-file")

	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestCommandServiceListCommands_ContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockCommandRepository{
		listErr: context.Canceled,
	}
	service := NewCommandService(repo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.ListCommands(ctx)

	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}
