package ports

import (
	"context"
)

// ServerStatusRepository defines the interface for server status operations.
type ServerStatusRepository interface {
	// ServerStatus checks if the Obsidian server is running.
	ServerStatus(ctx context.Context) (map[string]interface{}, error)
}
