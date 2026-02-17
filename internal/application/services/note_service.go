// Package services contains application services that implement use cases.
// Services depend ONLY on domain and ports (interfaces).
package services

import (
	"context"
	"log/slog"

	"github.com/xvierd/mcp-obsidian-go/internal/application/ports"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// NoteService provides business logic for note operations.
type NoteService struct {
	repo       ports.NoteRepository
	cache      ports.CacheRepository
	activeRepo ports.ActiveNoteRepository
	logger     *slog.Logger
}

// NewNoteService creates a new NoteService.
func NewNoteService(
	repo ports.NoteRepository,
	cache ports.CacheRepository,
	activeRepo ports.ActiveNoteRepository,
	logger *slog.Logger,
) *NoteService {
	return &NoteService{
		repo:       repo,
		cache:      cache,
		activeRepo: activeRepo,
		logger:     logger,
	}
}

// GetNote retrieves a note, using cache if available.
func (s *NoteService) GetNote(ctx context.Context, path string) (*domain.Note, error) {
	// Try cache first if available
	if s.cache != nil {
		if note, found := s.cache.Get(path); found {
			s.logger.Debug("cache hit", "path", path)
			return note, nil
		}
	}

	// Fetch from repository
	note, err := s.repo.GetNote(ctx, path)
	if err != nil {
		s.logger.Error("failed to get note", "path", path, "error", err)
		return nil, err
	}

	// Store in cache
	if s.cache != nil {
		s.cache.Set(path, note)
	}

	return note, nil
}

// ListNotes lists all notes in the vault or a specific directory.
func (s *NoteService) ListNotes(ctx context.Context, directory string) ([]string, error) {
	return s.repo.ListNotes(ctx, directory)
}

// CreateNote creates a new note.
func (s *NoteService) CreateNote(ctx context.Context, path string, content string) error {
	if err := s.repo.CreateNote(ctx, path, content); err != nil {
		s.logger.Error("failed to create note", "path", path, "error", err)
		return err
	}
	return nil
}

// UpdateNote updates an existing note and invalidates cache.
func (s *NoteService) UpdateNote(ctx context.Context, path string, content string) error {
	if err := s.repo.UpdateNote(ctx, path, content); err != nil {
		s.logger.Error("failed to update note", "path", path, "error", err)
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(path)
	}

	return nil
}

// DeleteNote deletes a note and removes from cache.
func (s *NoteService) DeleteNote(ctx context.Context, path string) error {
	if err := s.repo.DeleteNote(ctx, path); err != nil {
		s.logger.Error("failed to delete note", "path", path, "error", err)
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(path)
	}

	return nil
}

// AppendNote appends content to a note and invalidates cache.
func (s *NoteService) AppendNote(ctx context.Context, path string, content string) error {
	if err := s.repo.AppendNote(ctx, path, content); err != nil {
		s.logger.Error("failed to append to note", "path", path, "error", err)
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(path)
	}

	return nil
}

// PatchNote patches a specific section of a note and invalidates cache.
func (s *NoteService) PatchNote(ctx context.Context, path string, patch domain.PatchRequest) error {
	if err := s.repo.PatchNote(ctx, path, patch); err != nil {
		s.logger.Error("failed to patch note", "path", path, "error", err)
		return err
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.Delete(path)
	}

	return nil
}

// OpenNote opens a note in Obsidian.
func (s *NoteService) OpenNote(ctx context.Context, path string) error {
	return s.repo.OpenNote(ctx, path)
}

// GetRecentChanges gets recent file changes in the vault.
func (s *NoteService) GetRecentChanges(ctx context.Context, limit int) ([]domain.RecentChange, error) {
	return s.repo.GetRecentChanges(ctx, limit)
}

// GetPeriodicNote gets a periodic note.
func (s *NoteService) GetPeriodicNote(ctx context.Context, period string, offset int) (*domain.Note, error) {
	return s.repo.GetPeriodicNote(ctx, period, offset)
}

// GetActiveNote gets the currently active note.
func (s *NoteService) GetActiveNote(ctx context.Context) (*domain.Note, error) {
	return s.activeRepo.GetActiveNote(ctx)
}

// UpdateActiveNote updates the currently active note.
func (s *NoteService) UpdateActiveNote(ctx context.Context, content string) error {
	return s.activeRepo.UpdateActiveNote(ctx, content)
}

// AppendActiveNote appends content to the currently active note.
func (s *NoteService) AppendActiveNote(ctx context.Context, content string) error {
	return s.activeRepo.AppendActiveNote(ctx, content)
}

// DeleteActiveNote deletes the currently active note.
func (s *NoteService) DeleteActiveNote(ctx context.Context) error {
	return s.activeRepo.DeleteActiveNote(ctx)
}

// PatchActiveNote patches a specific section of the active note.
func (s *NoteService) PatchActiveNote(ctx context.Context, patch domain.PatchRequest) error {
	return s.activeRepo.PatchActiveNote(ctx, patch)
}

// GetCacheStats returns cache statistics.
func (s *NoteService) GetCacheStats() ports.CacheStats {
	if s.cache == nil {
		return ports.CacheStats{}
	}
	return s.cache.Stats()
}

// InvalidateCache clears the entire cache.
func (s *NoteService) InvalidateCache() {
	if s.cache != nil {
		s.cache.Clear()
	}
}
