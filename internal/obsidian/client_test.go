package obsidian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestClientHTTPMethods(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		callFunc   func(*Client, context.Context) error
		wantMethod string
	}{
		{
			name:   "GET request",
			method: "GET",
			callFunc: func(c *Client, ctx context.Context) error {
				resp, err := c.Get(ctx, "/test")
				if err == nil {
					resp.Body.Close()
				}
				return err
			},
			wantMethod: "GET",
		},
		{
			name:   "POST request",
			method: "POST",
			callFunc: func(c *Client, ctx context.Context) error {
				resp, err := c.Post(ctx, "/test", []byte(`{"test":"data"}`))
				if err == nil {
					resp.Body.Close()
				}
				return err
			},
			wantMethod: "POST",
		},
		{
			name:   "PUT request",
			method: "PUT",
			callFunc: func(c *Client, ctx context.Context) error {
				resp, err := c.Put(ctx, "/test", []byte(`{"test":"data"}`))
				if err == nil {
					resp.Body.Close()
				}
				return err
			},
			wantMethod: "PUT",
		},
		{
			name:   "PATCH request",
			method: "PATCH",
			callFunc: func(c *Client, ctx context.Context) error {
				resp, err := c.Patch(ctx, "/test", []byte(`{"test":"data"}`))
				if err == nil {
					resp.Body.Close()
				}
				return err
			},
			wantMethod: "PATCH",
		},
		{
			name:   "DELETE request",
			method: "DELETE",
			callFunc: func(c *Client, ctx context.Context) error {
				resp, err := c.Delete(ctx, "/test")
				if err == nil {
					resp.Body.Close()
				}
				return err
			},
			wantMethod: "DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedMethod string
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient("test-key", "127.0.0.1", 27124)
			client.httpClient = server.Client()
			client.baseURL = server.URL

			tt.callFunc(client, context.Background())

			if receivedMethod != tt.wantMethod {
				t.Errorf("expected method %s, got %s", tt.wantMethod, receivedMethod)
			}
		})
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

func TestClientListFilesInDirectory(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vault/daily" {
			t.Errorf("expected path '/vault/daily', got %s", r.URL.Path)
		}

		files := []string{"daily/2024-01-01.md", "daily/2024-01-02.md"}
		json.NewEncoder(w).Encode(files)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	files, err := client.ListFiles(context.Background(), "daily")
	if err != nil {
		t.Errorf("ListFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
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

func TestClientAppendNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
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

	err := client.AppendNote(context.Background(), "existing-note.md", "\nAppended content")
	if err != nil {
		t.Errorf("AppendNote() error = %v", err)
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

func TestClientPatchNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected method PATCH, got %s", r.Method)
		}

		if r.URL.Path != "/vault/patch-note.md" {
			t.Errorf("expected path '/vault/patch-note.md', got %s", r.URL.Path)
		}

		// Verify body content
		var patch PatchRequest
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			t.Errorf("failed to decode patch request: %v", err)
		}

		if patch.Operation != "append" {
			t.Errorf("expected operation 'append', got %s", patch.Operation)
		}

		if patch.Target != "heading" {
			t.Errorf("expected target 'heading', got %s", patch.Target)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	patch := PatchRequest{
		Operation:   "append",
		Target:      "heading",
		TargetValue: "## My Heading",
		Content:     "New content",
	}

	err := client.PatchNote(context.Background(), "patch-note.md", patch)
	if err != nil {
		t.Errorf("PatchNote() error = %v", err)
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

		results := []SearchResult{
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

func TestClientComplexSearch(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/search/jsonlogic/" {
			t.Errorf("expected path '/search/jsonlogic/', got %s", r.URL.Path)
		}

		results := []SearchResult{
			{Filename: "found-note.md"},
		}
		json.NewEncoder(w).Encode(results)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	query := map[string]interface{}{
		"and": []interface{}{
			map[string]interface{}{"contains": map[string]interface{}{"var": "content", "search": "test"}},
		},
	}

	results, err := client.ComplexSearch(context.Background(), query)
	if err != nil {
		t.Errorf("ComplexSearch() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestClientGetActiveNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/active/" {
			t.Errorf("expected path '/active/', got %s", r.URL.Path)
		}

		note := Note{
			Path:    "active-note.md",
			Content: "# Active Note",
		}
		json.NewEncoder(w).Encode(note)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	note, err := client.GetActiveNote(context.Background())
	if err != nil {
		t.Errorf("GetActiveNote() error = %v", err)
	}

	if note.Path != "active-note.md" {
		t.Errorf("expected path 'active-note.md', got %s", note.Path)
	}
}

func TestClientUpdateActiveNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected method PUT, got %s", r.Method)
		}

		if r.URL.Path != "/active/" {
			t.Errorf("expected path '/active/', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.UpdateActiveNote(context.Background(), "# Updated Active Note")
	if err != nil {
		t.Errorf("UpdateActiveNote() error = %v", err)
	}
}

func TestClientAppendActiveNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/active/" {
			t.Errorf("expected path '/active/', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.AppendActiveNote(context.Background(), "\nAppended content")
	if err != nil {
		t.Errorf("AppendActiveNote() error = %v", err)
	}
}

func TestClientDeleteActiveNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected method DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/active/" {
			t.Errorf("expected path '/active/', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.DeleteActiveNote(context.Background())
	if err != nil {
		t.Errorf("DeleteActiveNote() error = %v", err)
	}
}

func TestClientPatchActiveNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected method PATCH, got %s", r.Method)
		}

		if r.URL.Path != "/active/" {
			t.Errorf("expected path '/active/', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	patch := PatchRequest{
		Operation: "replace",
		Target:    "content",
		Content:   "Replaced content",
	}

	err := client.PatchActiveNote(context.Background(), patch)
	if err != nil {
		t.Errorf("PatchActiveNote() error = %v", err)
	}
}

func TestClientOpenNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/open/test-note.md" {
			t.Errorf("expected path '/open/test-note.md', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.OpenNote(context.Background(), "test-note.md")
	if err != nil {
		t.Errorf("OpenNote() error = %v", err)
	}
}

func TestClientListCommands(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/commands/" {
			t.Errorf("expected path '/commands/', got %s", r.URL.Path)
		}

		commands := []Command{
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

func TestClientExecuteCommand(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/commands/app:toggle-sidebar/" {
			t.Errorf("expected path '/commands/app:toggle-sidebar/', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.ExecuteCommand(context.Background(), "app:toggle-sidebar")
	if err != nil {
		t.Errorf("ExecuteCommand() error = %v", err)
	}
}

func TestClientGetRecentChanges(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vault/recent/" {
			t.Errorf("expected path '/vault/recent/', got %s", r.URL.Path)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit '10', got %s", limit)
		}

		changes := []RecentChange{
			{Path: "note1.md", Operation: "modified"},
			{Path: "note2.md", Operation: "created"},
		}
		json.NewEncoder(w).Encode(changes)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	changes, err := client.GetRecentChanges(context.Background(), 10)
	if err != nil {
		t.Errorf("GetRecentChanges() error = %v", err)
	}

	if len(changes) != 2 {
		t.Errorf("expected 2 changes, got %d", len(changes))
	}
}

func TestClientGetPeriodicNote(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/periodic/daily/" {
			t.Errorf("expected path '/periodic/daily/', got %s", r.URL.Path)
		}

		offset := r.URL.Query().Get("offset")
		if offset != "0" {
			t.Errorf("expected offset '0', got %s", offset)
		}

		note := Note{
			Path:    "daily/2024-01-01.md",
			Content: "# Daily Note",
		}
		json.NewEncoder(w).Encode(note)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	note, err := client.GetPeriodicNote(context.Background(), "daily", 0)
	if err != nil {
		t.Errorf("GetPeriodicNote() error = %v", err)
	}

	if note.Path != "daily/2024-01-01.md" {
		t.Errorf("expected path 'daily/2024-01-01.md', got %s", note.Path)
	}
}

func TestClientDataviewQuery(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/dataview/" {
			t.Errorf("expected path '/dataview/', got %s", r.URL.Path)
		}

		result := DataviewResult{
			Headers: []string{"file", "tags"},
			Rows:    [][]interface{}{{"note1.md", []string{"#tag1"}}},
			Count:   1,
		}
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	result, err := client.DataviewQuery(context.Background(), "TABLE file.tags FROM #tag1")
	if err != nil {
		t.Errorf("DataviewQuery() error = %v", err)
	}

	if result.Count != 1 {
		t.Errorf("expected count 1, got %d", result.Count)
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

	if !IsUnauthorized(err) {
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

func TestClientRateLimit(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limited"))
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	_, err := client.GetNote(context.Background(), "test.md")
	if err == nil {
		t.Error("expected error for rate limited request")
	}
}

func TestClientServerError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient("test-key", "127.0.0.1", 27124)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	_, err := client.GetNote(context.Background(), "test.md")
	if err == nil {
		t.Error("expected error for server error")
	}
}
