package services

import (
	"context"
	"log/slog"

	"github.com/xvierd/mcp-obsidian-go/internal/application/ports"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// CommandService provides business logic for command operations.
type CommandService struct {
	repo   ports.CommandRepository
	logger *slog.Logger
}

// NewCommandService creates a new CommandService.
func NewCommandService(repo ports.CommandRepository, logger *slog.Logger) *CommandService {
	return &CommandService{
		repo:   repo,
		logger: logger,
	}
}

// ListCommands lists all available Obsidian commands.
func (s *CommandService) ListCommands(ctx context.Context) ([]domain.Command, error) {
	commands, err := s.repo.ListCommands(ctx)
	if err != nil {
		s.logger.Error("failed to list commands", "error", err)
		return nil, err
	}

	s.logger.Info("commands listed", "count", len(commands))
	return commands, nil
}

// ExecuteCommand executes an Obsidian command by ID.
func (s *CommandService) ExecuteCommand(ctx context.Context, commandID string) error {
	if commandID == "" {
		return domain.ErrInvalidRequest
	}

	if err := s.repo.ExecuteCommand(ctx, commandID); err != nil {
		s.logger.Error("failed to execute command", "command_id", commandID, "error", err)
		return err
	}

	s.logger.Info("command executed", "command_id", commandID)
	return nil
}
