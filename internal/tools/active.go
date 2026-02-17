package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xvierd/mcp-obsidian-go/internal/obsidian"
)

// RegisterActiveNoteTools registers all active note operation tools.
func RegisterActiveNoteTools(r *Registry) {
	// get_active_note tool
	r.Register(
		&Tool{
			Name:        "get_active_note",
			Description: "Get the content of the currently active (open) note in Obsidian",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			note, err := r.GetClient().GetActiveNote(ctx)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"path":        note.Path,
				"content":     note.Content,
				"frontmatter": note.Frontmatter,
				"stat":        note.Stat,
			}, nil
		},
	)

	// update_active_note tool
	r.Register(
		&Tool{
			Name:        "update_active_note",
			Description: "Update the content of the currently active note by replacing all content",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"content": {
						"type": "string",
						"description": "New content to replace the active note's content"
					}
				},
				"required": ["content"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if err := r.GetClient().UpdateActiveNote(ctx, args.Content); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"message": "Active note updated successfully",
			}, nil
		},
	)

	// append_active_note tool
	r.Register(
		&Tool{
			Name:        "append_active_note",
			Description: "Append content to the end of the currently active note",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"content": {
						"type": "string",
						"description": "Content to append to the active note"
					}
				},
				"required": ["content"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if err := r.GetClient().AppendActiveNote(ctx, args.Content); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"message": "Content appended to active note",
			}, nil
		},
	)

	// patch_active_note tool
	r.Register(
		&Tool{
			Name:        "patch_active_note",
			Description: "Patch a specific section of the active note (heading, block, or frontmatter)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"operation": {
						"type": "string",
						"enum": ["append", "prepend", "replace"],
						"description": "Patch operation to perform"
					},
					"target": {
						"type": "string",
						"enum": ["content", "heading", "block", "frontmatter"],
						"description": "What to patch"
					},
					"targetValue": {
						"type": "string",
						"description": "For heading/block targets, the specific heading text or block ID"
					},
					"content": {
						"type": "string",
						"description": "New content to apply"
					}
				},
				"required": ["operation", "target", "content"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Operation   string `json:"operation"`
				Target      string `json:"target"`
				TargetValue string `json:"targetValue"`
				Content     string `json:"content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			patch := obsidian.PatchRequest{
				Operation:   args.Operation,
				Target:      args.Target,
				TargetValue: args.TargetValue,
				Content:     args.Content,
			}

			if err := r.GetClient().PatchActiveNote(ctx, patch); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"message": "Active note patched successfully",
			}, nil
		},
	)

	// delete_active_note tool
	r.Register(
		&Tool{
			Name:        "delete_active_note",
			Description: "Delete the currently active note. This action cannot be undone.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			if err := r.GetClient().DeleteActiveNote(ctx); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"message": "Active note deleted successfully",
			}, nil
		},
	)

	// open_note tool
	r.Register(
		&Tool{
			Name:        "open_note",
			Description: "Open a specific note in Obsidian (brings it to the foreground)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note to open"
					}
				},
				"required": ["path"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			if err := r.GetClient().OpenNote(ctx, args.Path); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"path":    args.Path,
				"message": fmt.Sprintf("Note opened: %s", args.Path),
			}, nil
		},
	)

	// get_recent_changes tool
	r.Register(
		&Tool{
			Name:        "get_recent_changes",
			Description: "Get a list of recently modified, created, or deleted files in the vault",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"limit": {
						"type": "integer",
						"default": 50,
						"description": "Maximum number of changes to return"
					}
				}
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Limit int `json:"limit"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Limit <= 0 {
				args.Limit = 50
			}
			if args.Limit > 1000 {
				args.Limit = 1000
			}

			changes, err := r.GetClient().GetRecentChanges(ctx, args.Limit)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"count":   len(changes),
				"changes": changes,
			}, nil
		},
	)

	// get_periodic_note tool
	r.Register(
		&Tool{
			Name:        "get_periodic_note",
			Description: "Get a periodic note (daily, weekly, monthly, etc.) by period type and offset",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"period": {
						"type": "string",
						"enum": ["daily", "weekly", "monthly", "yearly"],
						"description": "Type of periodic note"
					},
					"offset": {
						"type": "integer",
						"default": 0,
						"description": "Offset from current period (0 = today/this week, -1 = yesterday/last week, 1 = tomorrow/next week)"
					}
				},
				"required": ["period"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Period string `json:"period"`
				Offset int    `json:"offset"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Period == "" {
				return nil, fmt.Errorf("period is required")
			}

			note, err := r.GetClient().GetPeriodicNote(ctx, args.Period, args.Offset)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"path":        note.Path,
				"content":     note.Content,
				"frontmatter": note.Frontmatter,
			}, nil
		},
	)
}
