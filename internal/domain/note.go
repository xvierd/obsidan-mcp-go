// Package domain contains core domain entities and value objects.
// This package has NO external dependencies.
package domain

import (
	"time"
)

// Note represents an Obsidian note with its metadata and content.
type Note struct {
	Path        string                 `json:"path"`
	Content     string                 `json:"content"`
	Frontmatter map[string]interface{} `json:"frontmatter,omitempty"`
	Stat        FileStat               `json:"stat,omitempty"`
}

// FileStat contains file system metadata for a note.
type FileStat struct {
	Size    int64     `json:"size"`
	Mtime   int64     `json:"mtime"` // Modification time in milliseconds
	Ctime   int64     `json:"ctime"` // Creation time in milliseconds
	ModTime time.Time `json:"-"`     // Parsed modification time
}

// Command represents an Obsidian command that can be executed.
type Command struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SearchResult represents a search result from the Obsidian API.
type SearchResult struct {
	Filename string  `json:"filename"`
	Matches  []Match `json:"matches,omitempty"`
}

// Match represents a single match within a search result.
type Match struct {
	Context string `json:"context"`
}

// Link represents a link between notes.
type Link struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "wikilink" or "markdown"
}

// Tag represents a tag found in a note.
type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Heading represents a heading within a note.
type Heading struct {
	Level int    `json:"level"`
	Title string `json:"title"`
	Line  int    `json:"line"`
}

// Task represents a task/checkbox found in a note.
type Task struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
	Line      int    `json:"line"`
}

// PatchRequest represents a request to patch a note.
type PatchRequest struct {
	// Operation is the type of patch operation: "append", "prepend", "replace"
	Operation string `json:"operation"`

	// Target specifies what to patch: "content", "heading", "block", "frontmatter"
	Target string `json:"target"`

	// TargetValue specifies the target identifier (e.g., heading text, block ID)
	TargetValue string `json:"targetValue,omitempty"`

	// Content is the new content to apply
	Content string `json:"content"`
}

// RecentChange represents a recent file change in the vault.
type RecentChange struct {
	Path      string    `json:"path"`
	Operation string    `json:"operation"` // "created", "modified", "deleted"
	Timestamp time.Time `json:"timestamp"`
}

// DataviewResult represents the result of a Dataview query.
type DataviewResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
}

// PeriodicNoteType represents the type of periodic note.
type PeriodicNoteType string

const (
	PeriodicDaily   PeriodicNoteType = "daily"
	PeriodicWeekly  PeriodicNoteType = "weekly"
	PeriodicMonthly PeriodicNoteType = "monthly"
	PeriodicYearly  PeriodicNoteType = "yearly"
)

// ServerStatus represents the status of the Obsidian server.
type ServerStatus struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}
