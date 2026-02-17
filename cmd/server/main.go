package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/xvierd/mcp-obsidian-go/internal/application/services"
	"github.com/xvierd/mcp-obsidian-go/internal/config"
	"github.com/xvierd/mcp-obsidian-go/internal/infrastructure/adapters/memory"
	"github.com/xvierd/mcp-obsidian-go/internal/infrastructure/adapters/obsidian"
	"github.com/xvierd/mcp-obsidian-go/internal/mcp"
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

	// Create infrastructure adapters
	// 1. Create Obsidian HTTP client adapter
	obsidianClient := obsidian.NewClient(
		cfg.Obsidian.APIKey,
		cfg.Obsidian.Host,
		cfg.Obsidian.Port,
	)

	// 2. Create cache adapter (if enabled)
	var cacheAdapter *memory.Cache
	if cfg.Cache.Enabled {
		cacheAdapter = memory.NewCache(cfg.Cache.Size, cfg.Cache.TTL)
		logger.Info("cache enabled", "size", cfg.Cache.Size, "ttl", cfg.Cache.TTL)
	}

	// Create application services
	// Services depend only on ports (interfaces), not concrete implementations
	noteService := services.NewNoteService(
		obsidianClient, // implements ports.NoteRepository
		cacheAdapter,   // implements ports.CacheRepository (can be nil)
		obsidianClient, // implements ports.ActiveNoteRepository
		logger,
	)

	searchService := services.NewSearchService(
		obsidianClient, // implements ports.SearchRepository
		logger,
	)

	commandService := services.NewCommandService(
		obsidianClient, // implements ports.CommandRepository
		logger,
	)

	statusService := services.NewStatusService(
		obsidianClient, // implements ports.ServerStatusRepository
		logger,
	)

	// Create MCP server
	server := mcp.NewServer(logger)

	// Register tools
	registerTools(server, noteService, searchService, commandService, statusService, cfg, logger)

	// Run server
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func registerTools(
	server *mcp.Server,
	noteService *services.NoteService,
	searchService *services.SearchService,
	commandService *services.CommandService,
	statusService *services.StatusService,
	cfg *config.Config,
	logger *slog.Logger,
) {
	// Create tool registry
	// Registry now depends on services, not the HTTP client directly
	registry := tools.NewRegistry(logger, noteService, searchService, commandService)

	// Register all tool categories
	tools.RegisterVaultTools(registry)
	tools.RegisterSearchTools(registry)
	tools.RegisterCommandTools(registry)
	tools.RegisterActiveNoteTools(registry)
	tools.RegisterBatchTools(registry)

	// Register server_status tool
	// Uses the StatusService which properly implements the ServerStatusRepository port
	registry.Register(
		&tools.Tool{
			Name:        "server_status",
			Description: "Get the status of the Obsidian Local REST API server",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {}}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			status, err := statusService.GetStatus(ctx)
			if err != nil {
				// If we can't get status, server might be down
				return map[string]interface{}{
					"status":  "unreachable",
					"message": "Could not connect to Obsidian server",
				}, nil
			}
			return map[string]interface{}{
				"status":  "ok",
				"message": "Obsidian server is reachable",
				"details": status,
			}, nil
		},
	)

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
