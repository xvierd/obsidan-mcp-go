package obsidian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key", "127.0.0.1", 27124)

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

func TestClientServerStatus(t *testing.T) {
	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or invalid authorization header")
		}

		if r.URL.Path != "/" {
			t.Errorf("expected path '/', got %s", r.URL.Path)
		}

		response := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server
	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	status, err := client.ServerStatus(context.Background())
	if err != nil {
		t.Errorf("ServerStatus() error = %v", err)
	}

	if status["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", status["status"])
	}
}

func TestClientGetNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vault/test-note.md" {
			t.Errorf("expected path '/vault/test-note.md', got %s", r.URL.Path)
		}

		note := Note{
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

	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true, got false")
	}
}

func TestClientListFiles(t *testing.T) {
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

	files, err := client.ListFiles(context.Background(), "")
	if err != nil {
		t.Errorf("ListFiles() error = %v", err)
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
