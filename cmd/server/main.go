package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/xvierd/mcp-obsidian-go/internal/cache"
	"github.com/xvierd/mcp-obsidian-go/internal/config"
	"github.com/xvierd/mcp-obsidian-go/internal/mcp"
	"github.com/xvierd/mcp-obsidian-go/internal/obsidian"
	"github.com/xvierd/mcp-obsidian-go/internal/tools"
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
		"cache_enabled", cfg.Cache.Enabled,
	)

	// Create Obsidian client
	var client *obsidian.Client
	if cfg.Cache.Enabled {
		// Create cached client
		baseClient := obsidian.NewClient(
			cfg.Obsidian.APIKey,
			cfg.Obsidian.Host,
			cfg.Obsidian.Port,
		)
		cachedClient := cache.NewCachedClient(baseClient, cfg.Cache.Size, cfg.Cache.TTL)
		client = cachedClient.Client
		logger.Info("cache enabled", "size", cfg.Cache.Size, "ttl", cfg.Cache.TTL)
	} else {
		client = obsidian.NewClient(
			cfg.Obsidian.APIKey,
			cfg.Obsidian.Host,
			cfg.Obsidian.Port,
		)
	}

	// Create MCP server
	server := mcp.NewServer(logger)

	// Register tools
	registerTools(server, client, cfg, logger)

	// Run server
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func registerTools(server *mcp.Server, client *obsidian.Client, cfg *config.Config, logger *slog.Logger) {
	// Create tool registry
	registry := tools.NewRegistry(logger, client)

	// Register all tool categories
	tools.RegisterVaultTools(registry)
	tools.RegisterSearchTools(registry)
	tools.RegisterCommandTools(registry)
	tools.RegisterActiveNoteTools(registry)
	tools.RegisterBatchTools(registry)

	// Register server_status tool
	registry.Register(
		&tools.Tool{
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

	// Register vector tools if enabled
	// NOTE: Vector search disabled - requires 5GB model
	logger.Info("vector search disabled - using text search only")

	// Register all tools with MCP server
	toolCount := 0
	for _, tool := range registry.ListTools() {
		t := tool // capture range variable
		server.RegisterTool(
			&mcp.Tool{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			},
			func(ctx context.Context, params json.RawMessage) (interface{}, error) {
				return registry.Execute(ctx, t.Name, params)
			},
		)
		toolCount++
	}

	logger.Info("tools registered", "count", toolCount)
}
