package tools

import (
	"context"
	"encoding/json"
)

// RegisterStatusTools registers server status tools.
func RegisterStatusTools(r *Registry) {
	r.Register(
		&Tool{
			Name:        "server_status",
			Description: "Get the status of the Obsidian Local REST API server",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {}}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			status, err := r.GetStatusService().GetStatus(ctx)
			if err != nil {
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
}
