package ports

import (
	"context"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// CommandRepository defines the interface for command operations.
type CommandRepository interface {
	// ListCommands lists all available Obsidian commands.
	ListCommands(ctx context.Context) ([]domain.Command, error)

	// ExecuteCommand executes an Obsidian command by ID.
	ExecuteCommand(ctx context.Context, commandID string) error
}
