// Package obsidian provides an HTTP adapter for the Obsidian Local REST API.
// This adapter implements the application ports.
package obsidian

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/xvierd/mcp-obsidian-go/internal/application/ports"
	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// Client provides an HTTP client for the Obsidian Local REST API.
// It implements multiple repository ports.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// ClientOption allows customization of the Client.
type ClientOption func(*Client)

// NewClient creates a new Obsidian API client with connection pooling.
// Set insecure to true to skip TLS verification (required for Obsidian's self-signed certs).
func NewClient(apiKey, host string, port int, insecure bool, opts ...ClientOption) *Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	client := &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		baseURL: fmt.Sprintf("https://%s:%d", host, port),
		apiKey:  apiKey,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// WithHTTPClient allows setting a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// Ensure Client implements all required ports.
var (
	_ ports.NoteRepository         = (*Client)(nil)
	_ ports.ActiveNoteRepository   = (*Client)(nil)
	_ ports.CommandRepository      = (*Client)(nil)
	_ ports.SearchRepository       = (*Client)(nil)
	_ ports.ServerStatusRepository = (*Client)(nil)
)

// GetNote retrieves a note by path.
func (c *Client) GetNote(ctx context.Context, notePath string) (*domain.Note, error) {
	endpoint := path.Join("/vault/", notePath)
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var note domain.Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, fmt.Errorf("failed to decode note: %w", err)
	}

	return &note, nil
}

// ListNotes lists all files in the vault or a specific directory.
func (c *Client) ListNotes(ctx context.Context, directory string) ([]string, error) {
	endpoint := "/vault/"
	if directory != "" {
		endpoint = path.Join("/vault/", directory)
	}

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var result struct {
		Files []string `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode files list: %w", err)
	}

	return result.Files, nil
}

// CreateNote creates a new note with the given content.
func (c *Client) CreateNote(ctx context.Context, notePath string, content string) error {
	endpoint := path.Join("/vault/", notePath)

	body := map[string]string{"content": content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.doRequest(ctx, "PUT", endpoint, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// UpdateNote updates an existing note (replaces content).
func (c *Client) UpdateNote(ctx context.Context, notePath string, content string) error {
	endpoint := path.Join("/vault/", notePath)

	body := map[string]string{"content": content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.doRequest(ctx, "PUT", endpoint, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// DeleteNote deletes a note.
func (c *Client) DeleteNote(ctx context.Context, notePath string) error {
	endpoint := path.Join("/vault/", notePath)

	resp, err := c.doRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// AppendNote appends content to an existing note.
func (c *Client) AppendNote(ctx context.Context, notePath string, content string) error {
	endpoint := path.Join("/vault/", notePath)

	body := map[string]string{"content": content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", endpoint, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// PatchNote patches a specific section of a note.
func (c *Client) PatchNote(ctx context.Context, notePath string, patch domain.PatchRequest) error {
	endpoint := path.Join("/vault/", notePath)

	jsonBody, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch request: %w", err)
	}

	resp, err := c.doRequest(ctx, "PATCH", endpoint, jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// OpenNote opens a note in Obsidian.
func (c *Client) OpenNote(ctx context.Context, notePath string) error {
	endpoint := path.Join("/open/", notePath)

	resp, err := c.doRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// GetRecentChanges gets recent file changes in the vault.
func (c *Client) GetRecentChanges(ctx context.Context, limit int) ([]domain.RecentChange, error) {
	endpoint := fmt.Sprintf("/vault/recent/?limit=%d", limit)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var changes []domain.RecentChange
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return nil, fmt.Errorf("failed to decode recent changes: %w", err)
	}

	return changes, nil
}

// GetPeriodicNote gets a periodic note.
func (c *Client) GetPeriodicNote(ctx context.Context, period string, offset int) (*domain.Note, error) {
	endpoint := fmt.Sprintf("/periodic/%s/?offset=%d", url.PathEscape(period), offset)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var note domain.Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, fmt.Errorf("failed to decode periodic note: %w", err)
	}

	return &note, nil
}

// GetActiveNote gets the currently active note.
func (c *Client) GetActiveNote(ctx context.Context) (*domain.Note, error) {
	resp, err := c.doRequest(ctx, "GET", "/active/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var note domain.Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, fmt.Errorf("failed to decode active note: %w", err)
	}

	return &note, nil
}

// UpdateActiveNote updates the currently active note.
func (c *Client) UpdateActiveNote(ctx context.Context, content string) error {
	body := map[string]string{"content": content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.doRequest(ctx, "PUT", "/active/", jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// AppendActiveNote appends content to the currently active note.
func (c *Client) AppendActiveNote(ctx context.Context, content string) error {
	body := map[string]string{"content": content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/active/", jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// DeleteActiveNote deletes the currently active note.
func (c *Client) DeleteActiveNote(ctx context.Context) error {
	resp, err := c.doRequest(ctx, "DELETE", "/active/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// PatchActiveNote patches a specific section of the active note.
func (c *Client) PatchActiveNote(ctx context.Context, patch domain.PatchRequest) error {
	jsonBody, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch request: %w", err)
	}

	resp, err := c.doRequest(ctx, "PATCH", "/active/", jsonBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// Search performs a simple text search in the vault.
func (c *Client) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	endpoint := fmt.Sprintf("/search/simple/?query=%s", url.QueryEscape(query))

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var results []domain.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return results, nil
}

// ComplexSearch performs a complex search using JsonLogic query.
func (c *Client) ComplexSearch(ctx context.Context, query map[string]interface{}) ([]domain.SearchResult, error) {
	endpoint := "/search/jsonlogic/"

	jsonBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", endpoint, jsonBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var results []domain.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return results, nil
}

// DataviewQuery executes a Dataview query.
func (c *Client) DataviewQuery(ctx context.Context, query string) (*domain.DataviewResult, error) {
	endpoint := "/dataview/?query=" + url.QueryEscape(query)

	resp, err := c.doRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var result domain.DataviewResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode dataview result: %w", err)
	}

	return &result, nil
}

// ListCommands lists all available Obsidian commands.
func (c *Client) ListCommands(ctx context.Context) ([]domain.Command, error) {
	resp, err := c.doRequest(ctx, "GET", "/commands/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	var result struct {
		Commands []domain.Command `json:"commands"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode commands: %w", err)
	}

	return result.Commands, nil
}

// ExecuteCommand executes an Obsidian command by ID.
func (c *Client) ExecuteCommand(ctx context.Context, commandID string) error {
	endpoint := fmt.Sprintf("/commands/%s/", url.PathEscape(commandID))

	resp, err := c.doRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return mapHTTPStatusToDomain(resp.StatusCode, string(body))
	}

	return nil
}

// ServerStatus checks if the Obsidian server is running.
func (c *Client) ServerStatus(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", "/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status: %w", err)
	}

	return status, nil
}

// doRequest performs an HTTP request with authentication.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body []byte) (*http.Response, error) {
	url := c.baseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/vnd.olrapi.note+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// mapHTTPStatusToDomain maps HTTP status codes to domain errors.
func mapHTTPStatusToDomain(statusCode int, message string) error {
	switch statusCode {
	case 200, 201, 204:
		return nil
	case 400:
		return domain.NewDomainErrorWrap("INVALID_REQUEST", domain.ErrInvalidRequest)
	case 401:
		return domain.NewDomainErrorWrap("UNAUTHORIZED", domain.ErrUnauthorized)
	case 403:
		return domain.NewDomainErrorWrap("FORBIDDEN", domain.ErrForbidden)
	case 404:
		return domain.NewDomainErrorWrap("NOT_FOUND", domain.ErrNoteNotFound)
	case 429:
		return domain.NewDomainErrorWrap("RATE_LIMITED", domain.ErrRateLimited)
	case 500, 502, 503, 504:
		return domain.NewDomainErrorWrap("SERVER_ERROR", domain.ErrServerError)
	default:
		return domain.NewDomainError("UNKNOWN", message)
	}
}
