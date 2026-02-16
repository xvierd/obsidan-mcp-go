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
)

// Client provides an HTTP client for the Obsidian Local REST API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// ClientOption allows customization of the Client.
type ClientOption func(*Client)

// NewClient creates a new Obsidian API client with connection pooling.
func NewClient(apiKey, host string, port int, opts ...ClientOption) *Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Obsidian uses self-signed cert
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

// GetNote retrieves a note by path.
func (c *Client) GetNote(ctx context.Context, notePath string) (*Note, error) {
	endpoint := path.Join("/vault/", notePath)
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, MapHTTPStatus(resp.StatusCode, string(body))
	}

	var note Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, fmt.Errorf("failed to decode note: %w", err)
	}

	return &note, nil
}

// ListFiles lists all files in the vault or a specific directory.
func (c *Client) ListFiles(ctx context.Context, directory string) ([]string, error) {
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
		return nil, MapHTTPStatus(resp.StatusCode, string(body))
	}

	var files []string
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode files list: %w", err)
	}

	return files, nil
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
		return MapHTTPStatus(resp.StatusCode, string(body))
	}

	return nil
}

// UpdateNote updates an existing note.
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
		return MapHTTPStatus(resp.StatusCode, string(body))
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
		return MapHTTPStatus(resp.StatusCode, string(body))
	}

	return nil
}

// Search performs a simple text search in the vault.
func (c *Client) Search(ctx context.Context, query string) ([]SearchResult, error) {
	endpoint := fmt.Sprintf("/search/simple/?query=%s", url.QueryEscape(query))

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, MapHTTPStatus(resp.StatusCode, string(body))
	}

	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return results, nil
}

// GetActiveNote gets the currently active note.
func (c *Client) GetActiveNote(ctx context.Context) (*Note, error) {
	resp, err := c.doRequest(ctx, "GET", "/active/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, MapHTTPStatus(resp.StatusCode, string(body))
	}

	var note Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, fmt.Errorf("failed to decode active note: %w", err)
	}

	return &note, nil
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
