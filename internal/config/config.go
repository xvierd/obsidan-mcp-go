package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
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
			Provider:     "fasttext",
			DBPath:       "~/.mcp-obsidian/vector.db",
			BatchSize:    10,
			ChunkSize:    500,
			ChunkOverlap: 50,
		},
	}
}

// Load loads configuration from environment variables and optionally from a YAML file.
// Environment variables take precedence over YAML config.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// First, try to load from YAML config file
	if err := cfg.loadFromFile(); err != nil {
		// Log but don't fail - env vars might provide all needed config
		// We'll validate later
	}

	// Environment variables override YAML config
	cfg.loadFromEnv()

	return cfg, nil
}

// loadFromFile loads configuration from YAML files.
// It looks for config in the following locations (in order of precedence):
// 1. $MCP_OBSIDIAN_CONFIG environment variable (path to config file)
// 2. ./.mcp-obsidian.yaml (current directory)
// 3. ~/.config/mcp-obsidian/config.yaml
func (c *Config) loadFromFile() error {
	var configPath string

	// Check environment variable first
	if envPath := os.Getenv("MCP_OBSIDIAN_CONFIG"); envPath != "" {
		configPath = envPath
	} else {
		// Look for config in standard locations
		possiblePaths := []string{
			".mcp-obsidian.yaml",
			".mcp-obsidian.yml",
		}

		// Add home directory config paths
		if home, err := os.UserHomeDir(); err == nil {
			possiblePaths = append(possiblePaths,
				filepath.Join(home, ".config", "mcp-obsidian", "config.yaml"),
				filepath.Join(home, ".config", "mcp-obsidian", "config.yml"),
				filepath.Join(home, ".mcp-obsidian.yaml"),
				filepath.Join(home, ".mcp-obsidian.yml"),
			)
		}

		// Find first existing config file
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	// If no config file found, return without error
	if configPath == "" {
		return nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables.
func (c *Config) loadFromEnv() {
	// Load Obsidian config from env
	if apiKey := os.Getenv("OBSIDIAN_API_KEY"); apiKey != "" {
		c.Obsidian.APIKey = apiKey
	}
	if host := os.Getenv("OBSIDIAN_HOST"); host != "" {
		c.Obsidian.Host = host
	}
	if port := os.Getenv("OBSIDIAN_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Obsidian.Port = p
		}
	}

	// Load server config from env
	if transport := os.Getenv("MCP_TRANSPORT"); transport != "" {
		c.Server.Transport = transport
	}
	if logLevel := os.Getenv("MCP_LOG_LEVEL"); logLevel != "" {
		c.Server.LogLevel = logLevel
	}
	if timeout := os.Getenv("MCP_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Server.Timeout = d
		}
	}

	// Load cache config from env
	if enabled := os.Getenv("CACHE_ENABLED"); enabled != "" {
		c.Cache.Enabled = enabled == "true" || enabled == "1"
	}
	if ttl := os.Getenv("CACHE_TTL"); ttl != "" {
		if d, err := time.ParseDuration(ttl); err == nil {
			c.Cache.TTL = d
		}
	}
	if size := os.Getenv("CACHE_SIZE"); size != "" {
		if s, err := strconv.Atoi(size); err == nil {
			c.Cache.Size = s
		}
	}

	// Load vector config from env
	if enabled := os.Getenv("VECTOR_ENABLED"); enabled != "" {
		c.Vector.Enabled = enabled == "true" || enabled == "1"
	}
	if provider := os.Getenv("VECTOR_PROVIDER"); provider != "" {
		c.Vector.Provider = provider
	}
	if dbPath := os.Getenv("VECTOR_DB_PATH"); dbPath != "" {
		c.Vector.DBPath = dbPath
	}
}

// SaveToFile saves the configuration to a YAML file.
func (c *Config) SaveToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
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
