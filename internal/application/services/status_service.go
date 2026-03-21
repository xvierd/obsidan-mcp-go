// Package services contains application services that implement use cases.
package services

import (
	"context"
	"log/slog"

	"github.com/xvierd/mcp-obsidian-go/internal/application/ports"
)

// StatusService provides business logic for server status operations.
type StatusService struct {
	repo   ports.ServerStatusRepository
	logger *slog.Logger
}

// NewStatusService creates a new StatusService.
func NewStatusService(repo ports.ServerStatusRepository, logger *slog.Logger) *StatusService {
	return &StatusService{
		repo:   repo,
		logger: logger,
	}
}

// GetStatus checks if the Obsidian server is running.
func (s *StatusService) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	status, err := s.repo.ServerStatus(ctx)
	if err != nil {
		s.logger.Error("failed to get server status", "error", err)
		return nil, err
	}
	s.logger.Info("server status check completed", "status", status)
	return status, nil
}
