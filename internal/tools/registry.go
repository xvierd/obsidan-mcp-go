// Package tools provides MCP tool implementations for Obsidian operations.
// Tools call application services, not infrastructure directly.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/xvierd/mcp-obsidian-go/internal/application/services"
)

// Tool represents an MCP tool.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Handler is a function that handles a tool call.
type Handler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Registry manages MCP tools and their handlers.
type Registry struct {
	tools         map[string]*Tool
	handlers      map[string]Handler
	logger        *slog.Logger
	noteService   *services.NoteService
	searchService *services.SearchService
	cmdService    *services.CommandService
}

// NewRegistry creates a new tool registry.
func NewRegistry(
	logger *slog.Logger,
	noteService *services.NoteService,
	searchService *services.SearchService,
	cmdService *services.CommandService,
) *Registry {
	return &Registry{
		tools:         make(map[string]*Tool),
		handlers:      make(map[string]Handler),
		logger:        logger,
		noteService:   noteService,
		searchService: searchService,
		cmdService:    cmdService,
	}
}

// Register registers a tool with its handler.
func (r *Registry) Register(tool *Tool, handler Handler) {
	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
	r.logger.Debug("registered tool", "name", tool.Name)
}

// GetTool returns a tool by name.
func (r *Registry) GetTool(name string) (*Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetHandler returns a handler by name.
func (r *Registry) GetHandler(name string) (Handler, bool) {
	handler, ok := r.handlers[name]
	return handler, ok
}

// ListTools returns all registered tools.
func (r *Registry) ListTools() []*Tool {
	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute executes a tool by name with the given parameters.
func (r *Registry) Execute(ctx context.Context, name string, params json.RawMessage) (interface{}, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	start := time.Now()
	result, err := handler(ctx, params)
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("tool execution failed",
			"tool", name,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)
		return nil, err
	}

	r.logger.Info("tool executed",
		"tool", name,
		"duration_ms", duration.Milliseconds(),
	)

	return result, nil
}

// GetNoteService returns the note service.
func (r *Registry) GetNoteService() *services.NoteService {
	return r.noteService
}

// GetSearchService returns the search service.
func (r *Registry) GetSearchService() *services.SearchService {
	return r.searchService
}

// GetCommandService returns the command service.
func (r *Registry) GetCommandService() *services.CommandService {
	return r.cmdService
}

// GetLogger returns the logger.
func (r *Registry) GetLogger() *slog.Logger {
	return r.logger
}
