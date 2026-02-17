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
		"tls_insecure", cfg.Obsidian.Insecure,
	)

	// Create infrastructure adapters
	obsidianClient := obsidian.NewClient(
		cfg.Obsidian.APIKey,
		cfg.Obsidian.Host,
		cfg.Obsidian.Port,
		cfg.Obsidian.Insecure,
	)

	// Create cache adapter (if enabled)
	var cacheAdapter *memory.Cache
	if cfg.Cache.Enabled {
		cacheAdapter = memory.NewCache(cfg.Cache.Size, cfg.Cache.TTL)
		logger.Info("cache enabled", "size", cfg.Cache.Size, "ttl", cfg.Cache.TTL)
	}

	// Create application services (depend only on ports, not concrete implementations)
	noteService := services.NewNoteService(obsidianClient, cacheAdapter, obsidianClient, logger)
	searchService := services.NewSearchService(obsidianClient, logger)
	commandService := services.NewCommandService(obsidianClient, logger)
	statusService := services.NewStatusService(obsidianClient, logger)

	// Create MCP server
	server := mcp.NewServer(logger)

	// Create tool registry and register all tool categories
	registry := tools.NewRegistry(logger, noteService, searchService, commandService, statusService)
	tools.RegisterVaultTools(registry)
	tools.RegisterSearchTools(registry)
	tools.RegisterCommandTools(registry)
	tools.RegisterActiveNoteTools(registry)
	tools.RegisterBatchTools(registry)
	tools.RegisterStatusTools(registry)

	// Register all tools with MCP server
	toolCount := 0
	for _, tool := range registry.ListTools() {
		t := tool
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

	// Run server
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
