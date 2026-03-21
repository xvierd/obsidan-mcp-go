package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/application/ports"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// Mock implementations for testing
type mockNoteRepository struct {
	notes         map[string]*domain.Note
	listResult    []string
	getErr        error
	listErr       error
	createErr     error
	updateErr     error
	deleteErr     error
	appendErr     error
	patchErr      error
	openErr       error
	recentErr     error
	periodicErr   error
	recentChanges []domain.RecentChange
	periodicNote  *domain.Note
}

func (m *mockNoteRepository) GetNote(ctx context.Context, path string) (*domain.Note, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if note, ok := m.notes[path]; ok {
		return note, nil
	}
	return nil, domain.ErrNoteNotFound
}

func (m *mockNoteRepository) ListNotes(ctx context.Context, directory string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.listResult, nil
}

func (m *mockNoteRepository) CreateNote(ctx context.Context, path string, content string) error {
	return m.createErr
}

func (m *mockNoteRepository) UpdateNote(ctx context.Context, path string, content string) error {
	return m.updateErr
}

func (m *mockNoteRepository) DeleteNote(ctx context.Context, path string) error {
	return m.deleteErr
}

func (m *mockNoteRepository) AppendNote(ctx context.Context, path string, content string) error {
	return m.appendErr
}

func (m *mockNoteRepository) PatchNote(ctx context.Context, path string, patch domain.PatchRequest) error {
	return m.patchErr
}

func (m *mockNoteRepository) OpenNote(ctx context.Context, path string) error {
	return m.openErr
}

func (m *mockNoteRepository) GetRecentChanges(ctx context.Context, limit int) ([]domain.RecentChange, error) {
	if m.recentErr != nil {
		return nil, m.recentErr
	}
	return m.recentChanges, nil
}

func (m *mockNoteRepository) GetPeriodicNote(ctx context.Context, period string, offset int) (*domain.Note, error) {
	if m.periodicErr != nil {
		return nil, m.periodicErr
	}
	return m.periodicNote, nil
}

type mockActiveNoteRepository struct {
	activeNote *domain.Note
	getErr     error
	updateErr  error
	appendErr  error
	deleteErr  error
	patchErr   error
}

func (m *mockActiveNoteRepository) GetActiveNote(ctx context.Context) (*domain.Note, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.activeNote, nil
}

func (m *mockActiveNoteRepository) UpdateActiveNote(ctx context.Context, content string) error {
	return m.updateErr
}

func (m *mockActiveNoteRepository) AppendActiveNote(ctx context.Context, content string) error {
	return m.appendErr
}

func (m *mockActiveNoteRepository) DeleteActiveNote(ctx context.Context) error {
	return m.deleteErr
}

func (m *mockActiveNoteRepository) PatchActiveNote(ctx context.Context, patch domain.PatchRequest) error {
	return m.patchErr
}

type mockCacheRepository struct {
	data  map[string]*domain.Note
	stats ports.CacheStats
}

func (m *mockCacheRepository) Get(key string) (*domain.Note, bool) {
	note, ok := m.data[key]
	return note, ok
}

func (m *mockCacheRepository) Set(key string, note *domain.Note) {
	if m.data == nil {
		m.data = make(map[string]*domain.Note)
	}
	m.data[key] = note
}

func (m *mockCacheRepository) Delete(key string) {
	delete(m.data, key)
}

func (m *mockCacheRepository) Clear() {
	m.data = make(map[string]*domain.Note)
}

func (m *mockCacheRepository) Stats() ports.CacheStats {
	return m.stats
}

func TestNewNoteService(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{}
	cache := &mockCacheRepository{}
	activeRepo := &mockActiveNoteRepository{}

	service := NewNoteService(repo, cache, activeRepo, logger)

	if service == nil {
		t.Fatal("expected NoteService to be created")
	}
	if service.repo != repo {
		t.Error("expected repository to be set")
	}
	if service.cache != cache {
		t.Error("expected cache to be set")
	}
	if service.activeRepo != activeRepo {
		t.Error("expected activeRepo to be set")
	}
}

func TestNoteServiceGetNote_CacheMiss(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedNote := &domain.Note{
		Path:    "test.md",
		Content: "# Test Content",
	}
	repo := &mockNoteRepository{
		notes: map[string]*domain.Note{
			"test.md": expectedNote,
		},
	}
	cache := &mockCacheRepository{data: make(map[string]*domain.Note)}
	service := NewNoteService(repo, cache, nil, logger)

	note, err := service.GetNote(context.Background(), "test.md")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if note.Path != expectedNote.Path {
		t.Errorf("expected path %s, got %s", expectedNote.Path, note.Path)
	}
	// Verify note was added to cache
	if _, found := cache.data["test.md"]; !found {
		t.Error("expected note to be added to cache")
	}
}

func TestNoteServiceGetNote_CacheHit(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cachedNote := &domain.Note{
		Path:    "cached.md",
		Content: "Cached Content",
	}
	repo := &mockNoteRepository{
		getErr: errors.New("should not be called"),
	}
	cache := &mockCacheRepository{
		data: map[string]*domain.Note{
			"cached.md": cachedNote,
		},
	}
	service := NewNoteService(repo, cache, nil, logger)

	note, err := service.GetNote(context.Background(), "cached.md")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if note.Content != cachedNote.Content {
		t.Errorf("expected content %s, got %s", cachedNote.Content, note.Content)
	}
}

func TestNoteServiceGetNote_NotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{
		notes: map[string]*domain.Note{},
	}
	service := NewNoteService(repo, nil, nil, logger)

	_, err := service.GetNote(context.Background(), "missing.md")

	if err == nil {
		t.Error("expected error for missing note")
	}
	if !domain.IsNotFound(err) {
		t.Error("expected IsNotFound to be true")
	}
}

func TestNoteServiceGetNote_NoCache(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedNote := &domain.Note{
		Path:    "test.md",
		Content: "Content",
	}
	repo := &mockNoteRepository{
		notes: map[string]*domain.Note{
			"test.md": expectedNote,
		},
	}
	service := NewNoteService(repo, nil, nil, logger)

	note, err := service.GetNote(context.Background(), "test.md")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if note.Path != expectedNote.Path {
		t.Errorf("expected path %s, got %s", expectedNote.Path, note.Path)
	}
}

func TestNoteServiceCreateNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{}
	service := NewNoteService(repo, nil, nil, logger)

	err := service.CreateNote(context.Background(), "new.md", "# New Note")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteServiceCreateNote_Error(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{
		createErr: errors.New("create failed"),
	}
	service := NewNoteService(repo, nil, nil, logger)

	err := service.CreateNote(context.Background(), "new.md", "content")

	if err == nil {
		t.Error("expected error")
	}
}

func TestNoteServiceUpdateNote_CacheInvalidation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{}
	cache := &mockCacheRepository{
		data: map[string]*domain.Note{
			"update.md": {Path: "update.md", Content: "Old"},
		},
	}
	service := NewNoteService(repo, cache, nil, logger)

	err := service.UpdateNote(context.Background(), "update.md", "# New Content")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Verify cache was invalidated
	if _, found := cache.data["update.md"]; found {
		t.Error("expected cache entry to be deleted")
	}
}

func TestNoteServiceDeleteNote_CacheInvalidation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{}
	cache := &mockCacheRepository{
		data: map[string]*domain.Note{
			"delete.md": {Path: "delete.md", Content: "Content"},
		},
	}
	service := NewNoteService(repo, cache, nil, logger)

	err := service.DeleteNote(context.Background(), "delete.md")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Verify cache was invalidated
	if _, found := cache.data["delete.md"]; found {
		t.Error("expected cache entry to be deleted")
	}
}

func TestNoteServiceAppendNote_CacheInvalidation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{}
	cache := &mockCacheRepository{
		data: map[string]*domain.Note{
			"append.md": {Path: "append.md", Content: "Original"},
		},
	}
	service := NewNoteService(repo, cache, nil, logger)

	err := service.AppendNote(context.Background(), "append.md", "\nAppended")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Verify cache was invalidated
	if _, found := cache.data["append.md"]; found {
		t.Error("expected cache entry to be deleted")
	}
}

func TestNoteServicePatchNote_CacheInvalidation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{}
	cache := &mockCacheRepository{
		data: map[string]*domain.Note{
			"patch.md": {Path: "patch.md", Content: "Original"},
		},
	}
	service := NewNoteService(repo, cache, nil, logger)

	patch := domain.PatchRequest{
		Operation:   "append",
		Target:      "content",
		TargetValue: "",
		Content:     "Patched content",
	}
	err := service.PatchNote(context.Background(), "patch.md", patch)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Verify cache was invalidated
	if _, found := cache.data["patch.md"]; found {
		t.Error("expected cache entry to be deleted")
	}
}

func TestNoteServiceListNotes(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedFiles := []string{"note1.md", "note2.md"}
	repo := &mockNoteRepository{
		listResult: expectedFiles,
	}
	service := NewNoteService(repo, nil, nil, logger)

	files, err := service.ListNotes(context.Background(), "")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(files) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d", len(expectedFiles), len(files))
	}
}

func TestNoteServiceListNotes_Error(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{
		listErr: errors.New("list failed"),
	}
	service := NewNoteService(repo, nil, nil, logger)

	_, err := service.ListNotes(context.Background(), "")

	if err == nil {
		t.Error("expected error")
	}
}

func TestNoteServiceGetCacheStats(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedStats := ports.CacheStats{
		Size:    10,
		Hits:    100,
		Misses:  20,
		HitRate: 0.83,
	}
	cache := &mockCacheRepository{stats: expectedStats}
	service := NewNoteService(nil, cache, nil, logger)

	stats := service.GetCacheStats()

	if stats.Size != expectedStats.Size {
		t.Errorf("expected size %d, got %d", expectedStats.Size, stats.Size)
	}
	if stats.Hits != expectedStats.Hits {
		t.Errorf("expected hits %d, got %d", expectedStats.Hits, stats.Hits)
	}
}

func TestNoteServiceGetCacheStats_NoCache(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	service := NewNoteService(nil, nil, nil, logger)

	stats := service.GetCacheStats()

	if stats.Size != 0 {
		t.Errorf("expected size 0, got %d", stats.Size)
	}
}

func TestNoteServiceInvalidateCache(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cache := &mockCacheRepository{
		data: map[string]*domain.Note{
			"note1.md": {Path: "note1.md"},
			"note2.md": {Path: "note2.md"},
		},
	}
	service := NewNoteService(nil, cache, nil, logger)

	service.InvalidateCache()

	if len(cache.data) != 0 {
		t.Error("expected cache to be cleared")
	}
}

func TestNoteServiceInvalidateCache_NoCache(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	service := NewNoteService(nil, nil, nil, logger)

	// Should not panic
	service.InvalidateCache()
}

func TestNoteServiceOpenNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{}
	service := NewNoteService(repo, nil, nil, logger)

	err := service.OpenNote(context.Background(), "test.md")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteServiceOpenNote_Error(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{
		openErr: errors.New("open failed"),
	}
	service := NewNoteService(repo, nil, nil, logger)

	err := service.OpenNote(context.Background(), "test.md")

	if err == nil {
		t.Error("expected error")
	}
}

func TestNoteServiceGetRecentChanges(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedChanges := []domain.RecentChange{
		{Path: "note1.md", Operation: "created"},
		{Path: "note2.md", Operation: "modified"},
	}
	repo := &mockNoteRepository{
		recentChanges: expectedChanges,
	}
	service := NewNoteService(repo, nil, nil, logger)

	changes, err := service.GetRecentChanges(context.Background(), 10)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(changes) != len(expectedChanges) {
		t.Errorf("expected %d changes, got %d", len(expectedChanges), len(changes))
	}
}

func TestNoteServiceGetRecentChanges_Error(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{
		recentErr: errors.New("recent changes failed"),
	}
	service := NewNoteService(repo, nil, nil, logger)

	_, err := service.GetRecentChanges(context.Background(), 10)

	if err == nil {
		t.Error("expected error")
	}
}

func TestNoteServiceGetPeriodicNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedNote := &domain.Note{
		Path:    "Daily/2026-02-16.md",
		Content: "# Daily Note",
	}
	repo := &mockNoteRepository{
		periodicNote: expectedNote,
	}
	service := NewNoteService(repo, nil, nil, logger)

	note, err := service.GetPeriodicNote(context.Background(), "daily", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if note.Path != expectedNote.Path {
		t.Errorf("expected path %s, got %s", expectedNote.Path, note.Path)
	}
}

func TestNoteServiceGetPeriodicNote_Error(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	repo := &mockNoteRepository{
		periodicErr: errors.New("periodic note failed"),
	}
	service := NewNoteService(repo, nil, nil, logger)

	_, err := service.GetPeriodicNote(context.Background(), "daily", 0)

	if err == nil {
		t.Error("expected error")
	}
}

func TestNoteServiceGetActiveNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	expectedNote := &domain.Note{
		Path:    "active.md",
		Content: "Active content",
	}
	activeRepo := &mockActiveNoteRepository{
		activeNote: expectedNote,
	}
	service := NewNoteService(nil, nil, activeRepo, logger)

	note, err := service.GetActiveNote(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if note.Path != expectedNote.Path {
		t.Errorf("expected path %s, got %s", expectedNote.Path, note.Path)
	}
}

func TestNoteServiceGetActiveNote_Error(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	activeRepo := &mockActiveNoteRepository{
		getErr: errors.New("no active note"),
	}
	service := NewNoteService(nil, nil, activeRepo, logger)

	_, err := service.GetActiveNote(context.Background())

	if err == nil {
		t.Error("expected error")
	}
}

func TestNoteServiceUpdateActiveNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	activeRepo := &mockActiveNoteRepository{}
	service := NewNoteService(nil, nil, activeRepo, logger)

	err := service.UpdateActiveNote(context.Background(), "New content")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteServiceAppendActiveNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	activeRepo := &mockActiveNoteRepository{}
	service := NewNoteService(nil, nil, activeRepo, logger)

	err := service.AppendActiveNote(context.Background(), "\nAppended")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteServiceDeleteActiveNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	activeRepo := &mockActiveNoteRepository{}
	service := NewNoteService(nil, nil, activeRepo, logger)

	err := service.DeleteActiveNote(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoteServicePatchActiveNote(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	activeRepo := &mockActiveNoteRepository{}
	service := NewNoteService(nil, nil, activeRepo, logger)

	patch := domain.PatchRequest{
		Operation: "append",
		Target:    "content",
		Content:   "Patched",
	}
	err := service.PatchActiveNote(context.Background(), patch)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
