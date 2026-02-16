package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/xvierd/mcp-obsidian-go/internal/config"
	"github.com/xvierd/mcp-obsidian-go/internal/mcp"
	"github.com/xvierd/mcp-obsidian-go/internal/obsidian"
)

// Version is set at build time
var Version = "dev"

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting mcp-obsidian-go", "version", Version)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("configuration loaded",
		"host", cfg.Obsidian.Host,
		"port", cfg.Obsidian.Port,
	)

	// Create Obsidian client
	client := obsidian.NewClient(
		cfg.Obsidian.APIKey,
		cfg.Obsidian.Host,
		cfg.Obsidian.Port,
	)

	// Create MCP server
	server := mcp.NewServer(logger)

	// Register tools
	registerTools(server, client, logger)

	// Run server
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func registerTools(server *mcp.Server, client *obsidian.Client, logger *slog.Logger) {
	// Register server_status tool
	server.RegisterTool(
		&mcp.Tool{
			Name:        "server_status",
			Description: "Get the status of the Obsidian Local REST API server",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {}}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			status, err := client.ServerStatus(ctx)
			if err != nil {
				return nil, err
			}
			return status, nil
		},
	)

	// Register list_notes tool
	server.RegisterTool(
		&mcp.Tool{
			Name:        "list_notes",
			Description: "List all notes in the vault or a specific directory",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"directory": {
						"type": "string",
						"description": "Optional subdirectory to list (e.g., 'daily/2024')"
					}
				}
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Directory string `json:"directory"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			files, err := client.ListFiles(ctx, args.Directory)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"count": len(files),
				"files": files,
			}
			return result, nil
		},
	)

	// Register read_note tool
	server.RegisterTool(
		&mcp.Tool{
			Name:        "read_note",
			Description: "Read the content of a specific note",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note (e.g., 'daily/2024-01-01.md')"
					}
				},
				"required": ["path"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			note, err := client.GetNote(ctx, args.Path)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"path":    note.Path,
				"content": note.Content,
			}, nil
		},
	)

	logger.Info("tools registered", "count", 3)
}
