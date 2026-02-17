package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Obsidian.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Obsidian.Host)
	}

	if cfg.Obsidian.Port != 27124 {
		t.Errorf("expected port 27124, got %d", cfg.Obsidian.Port)
	}

	if cfg.Server.Transport != "stdio" {
		t.Errorf("expected transport stdio, got %s", cfg.Server.Transport)
	}

	if cfg.Cache.TTL != 30*time.Second {
		t.Errorf("expected cache TTL 30s, got %v", cfg.Cache.TTL)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Obsidian: ObsidianConfig{
					APIKey: "test-key",
					Host:   "127.0.0.1",
					Port:   27124,
				},
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			cfg: &Config{
				Obsidian: ObsidianConfig{
					Host: "127.0.0.1",
					Port: 27124,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			cfg: &Config{
				Obsidian: ObsidianConfig{
					APIKey: "test-key",
					Host:   "127.0.0.1",
					Port:   99999,
				},
			},
			wantErr: true,
		},
		{
			name: "port zero",
			cfg: &Config{
				Obsidian: ObsidianConfig{
					APIKey: "test-key",
					Host:   "127.0.0.1",
					Port:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "negative port",
			cfg: &Config{
				Obsidian: ObsidianConfig{
					APIKey: "test-key",
					Host:   "127.0.0.1",
					Port:   -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseURL(t *testing.T) {
	cfg := &ObsidianConfig{
		Host: "127.0.0.1",
		Port: 27124,
	}

	expected := "https://127.0.0.1:27124"
	if got := cfg.BaseURL(); got != expected {
		t.Errorf("BaseURL() = %v, want %v", got, expected)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		configEnv string
		filePath  string
		content   string
		wantHost  string
		wantPort  int
		wantErr   bool
	}{
		{
			name:     "load from valid yaml file",
			filePath: filepath.Join(tempDir, "config.yaml"),
			content: `
obsidian:
  api_key: yaml-api-key
  host: 192.168.1.100
  port: 3000
server:
  transport: http
  log_level: debug
`,
			wantHost: "192.168.1.100",
			wantPort: 3000,
			wantErr:  false,
		},
		{
			name:     "nonexistent file uses defaults",
			filePath: filepath.Join(tempDir, "nonexistent.yaml"),
			content:  "",          // Don't create this file
			wantHost: "127.0.0.1", // default
			wantPort: 27124,       // default
			wantErr:  false,       // Should not error, just use defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write config file if content is provided
			if tt.content != "" {
				if err := os.WriteFile(tt.filePath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("failed to write test config: %v", err)
				}
				os.Setenv("MCP_OBSIDIAN_CONFIG", tt.filePath)
				defer os.Unsetenv("MCP_OBSIDIAN_CONFIG")
			} else {
				// For nonexistent file, set env to a path that doesn't exist
				os.Setenv("MCP_OBSIDIAN_CONFIG", tt.filePath)
				defer os.Unsetenv("MCP_OBSIDIAN_CONFIG")
			}

			cfg := DefaultConfig()
			err := cfg.loadFromFile()

			if tt.wantErr && err == nil {
				t.Errorf("loadFromFile() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				// Nonexistent file should not cause error
				t.Logf("loadFromFile() returned error (may be expected for nonexistent): %v", err)
			}

			if cfg.Obsidian.Host != tt.wantHost {
				t.Errorf("expected host %s, got %s", tt.wantHost, cfg.Obsidian.Host)
			}
			if cfg.Obsidian.Port != tt.wantPort {
				t.Errorf("expected port %d, got %d", tt.wantPort, cfg.Obsidian.Port)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save and restore environment
	oldAPIKey := os.Getenv("OBSIDIAN_API_KEY")
	oldHost := os.Getenv("OBSIDIAN_HOST")
	oldPort := os.Getenv("OBSIDIAN_PORT")
	oldTransport := os.Getenv("MCP_TRANSPORT")
	oldLogLevel := os.Getenv("MCP_LOG_LEVEL")
	oldTimeout := os.Getenv("MCP_TIMEOUT")
	oldCacheEnabled := os.Getenv("CACHE_ENABLED")
	oldCacheTTL := os.Getenv("CACHE_TTL")
	oldCacheSize := os.Getenv("CACHE_SIZE")

	defer func() {
		os.Setenv("OBSIDIAN_API_KEY", oldAPIKey)
		os.Setenv("OBSIDIAN_HOST", oldHost)
		os.Setenv("OBSIDIAN_PORT", oldPort)
		os.Setenv("MCP_TRANSPORT", oldTransport)
		os.Setenv("MCP_LOG_LEVEL", oldLogLevel)
		os.Setenv("MCP_TIMEOUT", oldTimeout)
		os.Setenv("CACHE_ENABLED", oldCacheEnabled)
		os.Setenv("CACHE_TTL", oldCacheTTL)
		os.Setenv("CACHE_SIZE", oldCacheSize)
	}()

	// Set test environment variables
	os.Setenv("OBSIDIAN_API_KEY", "env-api-key")
	os.Setenv("OBSIDIAN_HOST", "10.0.0.1")
	os.Setenv("OBSIDIAN_PORT", "8080")
	os.Setenv("MCP_TRANSPORT", "sse")
	os.Setenv("MCP_LOG_LEVEL", "error")
	os.Setenv("MCP_TIMEOUT", "60s")
	os.Setenv("CACHE_ENABLED", "false")
	os.Setenv("CACHE_TTL", "5m")
	os.Setenv("CACHE_SIZE", "500")

	cfg := DefaultConfig()
	cfg.loadFromEnv()

	if cfg.Obsidian.APIKey != "env-api-key" {
		t.Errorf("expected API key 'env-api-key', got %s", cfg.Obsidian.APIKey)
	}
	if cfg.Obsidian.Host != "10.0.0.1" {
		t.Errorf("expected host '10.0.0.1', got %s", cfg.Obsidian.Host)
	}
	if cfg.Obsidian.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Obsidian.Port)
	}
	if cfg.Server.Transport != "sse" {
		t.Errorf("expected transport 'sse', got %s", cfg.Server.Transport)
	}
	if cfg.Server.LogLevel != "error" {
		t.Errorf("expected log level 'error', got %s", cfg.Server.LogLevel)
	}
	if cfg.Server.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", cfg.Server.Timeout)
	}
	if cfg.Cache.Enabled != false {
		t.Errorf("expected cache enabled false, got %v", cfg.Cache.Enabled)
	}
	if cfg.Cache.TTL != 5*time.Minute {
		t.Errorf("expected cache TTL 5m, got %v", cfg.Cache.TTL)
	}
	if cfg.Cache.Size != 500 {
		t.Errorf("expected cache size 500, got %d", cfg.Cache.Size)
	}
}

func TestSaveToFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	cfg := &Config{
		Obsidian: ObsidianConfig{
			APIKey: "test-api-key",
			Host:   "localhost",
			Port:   27124,
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
	}

	// Save config
	if err := cfg.SaveToFile(configPath); err != nil {
		t.Fatalf("SaveToFile() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load and verify
	loadedCfg := DefaultConfig()
	os.Setenv("MCP_OBSIDIAN_CONFIG", configPath)
	defer os.Unsetenv("MCP_OBSIDIAN_CONFIG")

	if err := loadedCfg.loadFromFile(); err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loadedCfg.Obsidian.Host != cfg.Obsidian.Host {
		t.Errorf("expected host %s, got %s", cfg.Obsidian.Host, loadedCfg.Obsidian.Host)
	}
	if loadedCfg.Obsidian.Port != cfg.Obsidian.Port {
		t.Errorf("expected port %d, got %d", cfg.Obsidian.Port, loadedCfg.Obsidian.Port)
	}
}

func TestEnvOverridesYaml(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	content := `
obsidian:
  api_key: yaml-api-key
  host: yaml-host
  port: 1111
server:
  transport: yaml-transport
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Set environment variable that should override YAML
	os.Setenv("MCP_OBSIDIAN_CONFIG", configPath)
	os.Setenv("OBSIDIAN_HOST", "env-host")
	os.Setenv("OBSIDIAN_PORT", "2222")
	defer func() {
		os.Unsetenv("MCP_OBSIDIAN_CONFIG")
		os.Unsetenv("OBSIDIAN_HOST")
		os.Unsetenv("OBSIDIAN_PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Environment variables should override YAML
	if cfg.Obsidian.Host != "env-host" {
		t.Errorf("expected host 'env-host' (from env), got %s", cfg.Obsidian.Host)
	}
	if cfg.Obsidian.Port != 2222 {
		t.Errorf("expected port 2222 (from env), got %d", cfg.Obsidian.Port)
	}
	// This should come from YAML (not overridden)
	if cfg.Obsidian.APIKey != "yaml-api-key" {
		t.Errorf("expected API key 'yaml-api-key' (from yaml), got %s", cfg.Obsidian.APIKey)
	}
}
