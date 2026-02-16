// Package config provides configuration management for the MCP Obsidian server.
// It supports environment variables and YAML configuration files.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	Obsidian ObsidianConfig `yaml:"obsidian"`
	Server   ServerConfig   `yaml:"server"`
	Cache    CacheConfig    `yaml:"cache"`
	Vector   VectorConfig   `yaml:"vector"`
}

// ObsidianConfig holds configuration for the Obsidian REST API connection.
type ObsidianConfig struct {
	APIKey string `yaml:"api_key"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
}

// ServerConfig holds configuration for the MCP server.
type ServerConfig struct {
	Transport string        `yaml:"transport"` // stdio, sse, http
	LogLevel  string        `yaml:"log_level"`
	Timeout   time.Duration `yaml:"timeout"`
}

// CacheConfig holds configuration for the note cache.
type CacheConfig struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"`
	Size    int           `yaml:"size"`
}

// VectorConfig holds configuration for vector search.
type VectorConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Provider     string `yaml:"provider"` // ollama, openai, google, cohere
	DBPath       string `yaml:"db_path"`
	BatchSize    int    `yaml:"batch_size"`
	ChunkSize    int    `yaml:"chunk_size"`
	ChunkOverlap int    `yaml:"chunk_overlap"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Obsidian: ObsidianConfig{
			Host: "127.0.0.1",
			Port: 27124,
		},
		Server: ServerConfig{
			Transport: "stdio",
			LogLevel:  "info",
			Timeout:   30 * time.Second,
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     30 * time.Second,
			Size:    1000,
		},
		Vector: VectorConfig{
			Enabled:      true,
			Provider:     "ollama",
			DBPath:       "~/.mcp-obsidian/vector.db",
			BatchSize:    10,
			ChunkSize:    500,
			ChunkOverlap: 50,
		},
	}
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load Obsidian config from env
	if apiKey := os.Getenv("OBSIDIAN_API_KEY"); apiKey != "" {
		cfg.Obsidian.APIKey = apiKey
	}
	if host := os.Getenv("OBSIDIAN_HOST"); host != "" {
		cfg.Obsidian.Host = host
	}
	if port := os.Getenv("OBSIDIAN_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid OBSIDIAN_PORT: %w", err)
		}
		cfg.Obsidian.Port = p
	}

	// Load server config from env
	if transport := os.Getenv("MCP_TRANSPORT"); transport != "" {
		cfg.Server.Transport = transport
	}
	if logLevel := os.Getenv("MCP_LOG_LEVEL"); logLevel != "" {
		cfg.Server.LogLevel = logLevel
	}

	// Load vector config from env
	if provider := os.Getenv("VECTOR_PROVIDER"); provider != "" {
		cfg.Vector.Provider = provider
	}
	if dbPath := os.Getenv("VECTOR_DB_PATH"); dbPath != "" {
		cfg.Vector.DBPath = dbPath
	}

	return cfg, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Obsidian.APIKey == "" {
		return fmt.Errorf("OBSIDIAN_API_KEY is required")
	}
	if c.Obsidian.Host == "" {
		return fmt.Errorf("obsidian host is required")
	}
	if c.Obsidian.Port <= 0 || c.Obsidian.Port > 65535 {
		return fmt.Errorf("invalid obsidian port: %d", c.Obsidian.Port)
	}
	return nil
}

// BaseURL returns the base URL for the Obsidian REST API.
func (c *ObsidianConfig) BaseURL() string {
	return fmt.Sprintf("https://%s:%d", c.Host, c.Port)
}
