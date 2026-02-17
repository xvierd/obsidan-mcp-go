package obsidian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key", "127.0.0.1", 27124)

	// Access unexported fields through testing
	if client.apiKey != "test-key" {
		t.Errorf("expected API key 'test-key', got %s", client.apiKey)
	}

	if client.baseURL != "https://127.0.0.1:27124" {
		t.Errorf("expected baseURL 'https://127.0.0.1:27124', got %s", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("expected HTTP client to be initialized")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 5 * time.Second}
	client := NewClient("test-key", "127.0.0.1", 27124,
		WithHTTPClient(customClient),
		WithTimeout(10*time.Second),
	)

	if client.httpClient != customClient {
		t.Error("expected custom HTTP client to be set")
	}

	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", client.httpClient.Timeout)
	}
}

func TestClientGetNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vault/test-note.md" {
			t.Errorf("expected path '/vault/test-note.md', got %s", r.URL.Path)
		}

		note := domain.Note{
			Path:    "test-note.md",
			Content: "# Test\nThis is a test note.",
		}
		json.NewEncoder(w).Encode(note)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	note, err := client.GetNote(context.Background(), "test-note.md")
	if err != nil {
		t.Errorf("GetNote() error = %v", err)
	}

	if note.Path != "test-note.md" {
		t.Errorf("expected path 'test-note.md', got %s", note.Path)
	}

	if note.Content != "# Test\nThis is a test note." {
		t.Errorf("unexpected content: %s", note.Content)
	}
}

func TestClientGetNoteNotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Note not found"))
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	_, err := client.GetNote(context.Background(), "missing-note.md")
	if err == nil {
		t.Error("expected error for not found note")
	}

	if !domain.IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true, got false")
	}
}

func TestClientListNotes(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vault/" {
			t.Errorf("expected path '/vault/', got %s", r.URL.Path)
		}

		files := []string{"note1.md", "note2.md", "folder/note3.md"}
		json.NewEncoder(w).Encode(files)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	files, err := client.ListNotes(context.Background(), "")
	if err != nil {
		t.Errorf("ListNotes() error = %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

func TestClientCreateNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected method PUT, got %s", r.Method)
		}

		if r.URL.Path != "/vault/new-note.md" {
			t.Errorf("expected path '/vault/new-note.md', got %s", r.URL.Path)
		}

		// Verify content type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type header to be application/json")
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.CreateNote(context.Background(), "new-note.md", "# New Note")
	if err != nil {
		t.Errorf("CreateNote() error = %v", err)
	}
}

func TestClientUpdateNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected method PUT, got %s", r.Method)
		}

		if r.URL.Path != "/vault/existing-note.md" {
			t.Errorf("expected path '/vault/existing-note.md', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.UpdateNote(context.Background(), "existing-note.md", "# Updated Content")
	if err != nil {
		t.Errorf("UpdateNote() error = %v", err)
	}
}

func TestClientDeleteNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected method DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/vault/delete-me.md" {
			t.Errorf("expected path '/vault/delete-me.md', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.DeleteNote(context.Background(), "delete-me.md")
	if err != nil {
		t.Errorf("DeleteNote() error = %v", err)
	}
}

func TestClientSearch(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/simple/" {
			t.Errorf("expected path '/search/simple/', got %s", r.URL.Path)
		}

		// Check query parameter
		query := r.URL.Query().Get("query")
		if query != "test query" {
			t.Errorf("expected query 'test query', got %s", query)
		}

		results := []domain.SearchResult{
			{Filename: "note1.md"},
			{Filename: "note2.md"},
		}
		json.NewEncoder(w).Encode(results)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	results, err := client.Search(context.Background(), "test query")
	if err != nil {
		t.Errorf("Search() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestClientListCommands(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/commands/" {
			t.Errorf("expected path '/commands/', got %s", r.URL.Path)
		}

		commands := []domain.Command{
			{ID: "cmd1", Name: "Command 1"},
			{ID: "cmd2", Name: "Command 2"},
		}
		json.NewEncoder(w).Encode(commands)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	commands, err := client.ListCommands(context.Background())
	if err != nil {
		t.Errorf("ListCommands() error = %v", err)
	}

	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}
}

func TestClientUnauthorized(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewClient("wrong-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	_, err := client.GetNote(context.Background(), "test.md")
	if err == nil {
		t.Error("expected error for unauthorized request")
	}

	if !domain.IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized to be true, got false")
	}
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetNote(ctx, "test.md")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
