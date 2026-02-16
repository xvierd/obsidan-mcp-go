// Package obsidian provides a client for the Obsidian Local REST API.
package obsidian

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
