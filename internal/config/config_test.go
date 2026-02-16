package config

import (
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
