package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
)

// Mock server status repository for testing
type mockServerStatusRepository struct {
	status map[string]interface{}
	err    error
}

func (m *mockServerStatusRepository) ServerStatus(ctx context.Context) (map[string]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.status, nil
}

func TestNewStatusService(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockServerStatusRepository{}

	service := NewStatusService(repo, logger)

	if service == nil {
		t.Fatal("expected StatusService to be created")
	}
	if service.repo != repo {
		t.Error("expected repository to be set")
	}
}

func TestStatusServiceGetStatus_Success(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedStatus := map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
	}
	repo := &mockServerStatusRepository{
		status: expectedStatus,
	}
	service := NewStatusService(repo, logger)

	status, err := service.GetStatus(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status["status"] != expectedStatus["status"] {
		t.Errorf("expected status %v, got %v", expectedStatus["status"], status["status"])
	}
	if status["version"] != expectedStatus["version"] {
		t.Errorf("expected version %v, got %v", expectedStatus["version"], status["version"])
	}
}

func TestStatusServiceGetStatus_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockServerStatusRepository{
		err: errors.New("connection refused"),
	}
	service := NewStatusService(repo, logger)

	_, err := service.GetStatus(context.Background())

	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused' error, got %v", err)
	}
}

func TestStatusServiceGetStatus_ContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockServerStatusRepository{
		err: context.Canceled,
	}
	service := NewStatusService(repo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetStatus(ctx)

	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}
