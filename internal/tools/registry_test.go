package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/obsidian"
)

func TestNewRegistry(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}

	registry := NewRegistry(logger, client)

	if registry == nil {
		t.Fatal("expected registry to be created")
	}

	if registry.logger != logger {
		t.Error("expected logger to be set")
	}

	if registry.client != client {
		t.Error("expected client to be set")
	}

	if len(registry.tools) != 0 {
		t.Errorf("expected empty tools map, got %d", len(registry.tools))
	}
}

func TestRegistryRegister(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: json.RawMessage(`{"type": "object"}`),
	}

	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "test result", nil
	}

	registry.Register(tool, handler)

	// Check tool is registered
	if len(registry.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(registry.tools))
	}

	// Check handler is registered
	if len(registry.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(registry.handlers))
	}

	// Get the tool
	retrievedTool, found := registry.GetTool("test_tool")
	if !found {
		t.Error("expected to find registered tool")
	}

	if retrievedTool.Name != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got %s", retrievedTool.Name)
	}

	// Get the handler
	retrievedHandler, found := registry.GetHandler("test_tool")
	if !found {
		t.Error("expected to find registered handler")
	}

	// Execute handler
	result, err := retrievedHandler(context.Background(), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != "test result" {
		t.Errorf("expected 'test result', got %v", result)
	}
}

func TestRegistryListTools(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	// Register multiple tools
	for i := 0; i < 3; i++ {
		tool := &Tool{
			Name:        fmt.Sprintf("tool_%d", i),
			Description: "Test tool",
			InputSchema: json.RawMessage(`{"type": "object"}`),
		}
		registry.Register(tool, func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			return nil, nil
		})
	}

	tools := registry.ListTools()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}
}

func TestRegistryExecute(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	tool := &Tool{
		Name:        "execute_test",
		Description: "Test execution",
		InputSchema: json.RawMessage(`{"type": "object"}`),
	}

	registry.Register(tool, func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		var args map[string]string
		if err := json.Unmarshal(params, &args); err != nil {
			return nil, err
		}
		return args["value"], nil
	})

	params := json.RawMessage(`{"value": "hello"}`)
	result, err := registry.Execute(context.Background(), "execute_test", params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
}

func TestRegistryExecuteNotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	_, err := registry.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error for non-existent tool")
	}

	if err.Error() != "tool not found: nonexistent" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRegistryGetters(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	if registry.GetClient() != client {
		t.Error("GetClient returned wrong client")
	}

	if registry.GetLogger() != logger {
		t.Error("GetLogger returned wrong logger")
	}
}

func TestGetToolNotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	_, found := registry.GetTool("nonexistent")
	if found {
		t.Error("expected not to find non-existent tool")
	}
}

func TestGetHandlerNotFound(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	_, found := registry.GetHandler("nonexistent")
	if found {
		t.Error("expected not to find non-existent handler")
	}
}
