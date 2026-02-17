package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// RegisterCommandTools registers all command-related tools.
func RegisterCommandTools(r *Registry) {
	// list_commands tool
	r.Register(
		&Tool{
			Name:        "list_commands",
			Description: "List all available Obsidian commands that can be executed",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			commands, err := r.GetClient().ListCommands(ctx)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"count":    len(commands),
				"commands": commands,
			}, nil
		},
	)

	// execute_command tool
	r.Register(
		&Tool{
			Name:        "execute_command",
			Description: "Execute an Obsidian command by its ID. Use list_commands to see available commands.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"command_id": {
						"type": "string",
						"description": "ID of the command to execute (e.g., 'app:toggle-sidebar')"
					}
				},
				"required": ["command_id"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				CommandID string `json:"command_id"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.CommandID == "" {
				return nil, fmt.Errorf("command_id is required")
			}

			if err := r.GetClient().ExecuteCommand(ctx, args.CommandID); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Command executed: %s", args.CommandID),
			}, nil
		},
	)
}
