package tools

import (
	"log/slog"
	"os"
	"testing"

	"github.com/xvierd/mcp-obsidian-go/internal/obsidian"
)

func TestRegisterVaultTools(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	client := &obsidian.Client{}
	registry := NewRegistry(logger, client)

	RegisterVaultTools(registry)

	// Check that all vault tools are registered
	expectedTools := []string{
		"list_notes",
		"read_note",
		"create_note",
		"update_note",
		"append_note",
		"delete_note",
		"patch_note",
	}

	for _, toolName := range expectedTools {
		_, found := registry.GetTool(toolName)
		if !found {
			t.Errorf("expected tool %s to be registered", toolName)
		}
	}
}
