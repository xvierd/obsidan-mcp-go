// Package ports defines the interfaces (ports) that the application layer uses.
// These interfaces are implemented by infrastructure adapters.
package ports

import (
	"context"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// NoteRepository defines the interface for note storage operations.
type NoteRepository interface {
	// GetNote retrieves a note by path.
	GetNote(ctx context.Context, path string) (*domain.Note, error)

	// ListNotes lists all notes in the vault or a specific directory.
	ListNotes(ctx context.Context, directory string) ([]string, error)

	// CreateNote creates a new note with the given content.
	CreateNote(ctx context.Context, path string, content string) error

	// UpdateNote updates an existing note (replaces content).
	UpdateNote(ctx context.Context, path string, content string) error

	// DeleteNote deletes a note.
	DeleteNote(ctx context.Context, path string) error

	// AppendNote appends content to an existing note.
	AppendNote(ctx context.Context, path string, content string) error

	// PatchNote patches a specific section of a note.
	PatchNote(ctx context.Context, path string, patch domain.PatchRequest) error

	// OpenNote opens a note in Obsidian.
	OpenNote(ctx context.Context, path string) error

	// GetRecentChanges gets recent file changes in the vault.
	GetRecentChanges(ctx context.Context, limit int) ([]domain.RecentChange, error)

	// GetPeriodicNote gets a periodic note (daily, weekly, monthly, etc.).
	GetPeriodicNote(ctx context.Context, period string, offset int) (*domain.Note, error)
}

// ActiveNoteRepository defines operations for the currently active note.
type ActiveNoteRepository interface {
	// GetActiveNote gets the currently active note.
	GetActiveNote(ctx context.Context) (*domain.Note, error)

	// UpdateActiveNote updates the currently active note.
	UpdateActiveNote(ctx context.Context, content string) error

	// AppendActiveNote appends content to the currently active note.
	AppendActiveNote(ctx context.Context, content string) error

	// DeleteActiveNote deletes the currently active note.
	DeleteActiveNote(ctx context.Context) error

	// PatchActiveNote patches a specific section of the active note.
	PatchActiveNote(ctx context.Context, patch domain.PatchRequest) error
}
