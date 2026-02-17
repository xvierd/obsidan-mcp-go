package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xvierd/mcp-obsidian-go/internal/domain"
)

// RegisterVaultTools registers all vault operation tools.
func RegisterVaultTools(r *Registry) {
	// list_notes tool
	r.Register(
		&Tool{
			Name:        "list_notes",
			Description: "List all notes in the vault or a specific directory",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"directory": {
						"type": "string",
						"description": "Optional subdirectory to list (e.g., 'daily/2024')"
					}
				}
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Directory string `json:"directory"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			files, err := r.GetNoteService().ListNotes(ctx, args.Directory)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"count": len(files),
				"files": files,
			}, nil
		},
	)

	// read_note tool
	r.Register(
		&Tool{
			Name:        "read_note",
			Description: "Read the content of a specific note",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note (e.g., 'daily/2024-01-01.md')"
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

			note, err := r.GetNoteService().GetNote(ctx, args.Path)
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

	// create_note tool
	r.Register(
		&Tool{
			Name:        "create_note",
			Description: "Create a new note with the given content. The note will be created at the specified path.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path where the note should be created (e.g., 'projects/new-project.md')"
					},
					"content": {
						"type": "string",
						"description": "Content of the note in Markdown format"
					}
				},
				"required": ["path", "content"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			if err := r.GetNoteService().CreateNote(ctx, args.Path, args.Content); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"path":    args.Path,
				"message": fmt.Sprintf("Note created successfully: %s", args.Path),
			}, nil
		},
	)

	// update_note tool
	r.Register(
		&Tool{
			Name:        "update_note",
			Description: "Update an existing note by replacing its entire content. Use this to overwrite a note completely.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note to update"
					},
					"content": {
						"type": "string",
						"description": "New content to replace the existing note content"
					}
				},
				"required": ["path", "content"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			if err := r.GetNoteService().UpdateNote(ctx, args.Path, args.Content); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"path":    args.Path,
				"message": fmt.Sprintf("Note updated successfully: %s", args.Path),
			}, nil
		},
	)

	// append_note tool
	r.Register(
		&Tool{
			Name:        "append_note",
			Description: "Append content to the end of an existing note without modifying the existing content",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note to append to"
					},
					"content": {
						"type": "string",
						"description": "Content to append to the note"
					}
				},
				"required": ["path", "content"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			if err := r.GetNoteService().AppendNote(ctx, args.Path, args.Content); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"path":    args.Path,
				"message": fmt.Sprintf("Content appended to note: %s", args.Path),
			}, nil
		},
	)

	// delete_note tool
	r.Register(
		&Tool{
			Name:        "delete_note",
			Description: "Delete a note from the vault. This action cannot be undone.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note to delete"
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

			if err := r.GetNoteService().DeleteNote(ctx, args.Path); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"path":    args.Path,
				"message": fmt.Sprintf("Note deleted successfully: %s", args.Path),
			}, nil
		},
	)

	// patch_note tool
	r.Register(
		&Tool{
			Name:        "patch_note",
			Description: "Patch a specific section of a note. Can target headings, blocks, or frontmatter. Supports append, prepend, and replace operations.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the note to patch"
					},
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
				"required": ["path", "operation", "target", "content"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Path        string `json:"path"`
				Operation   string `json:"operation"`
				Target      string `json:"target"`
				TargetValue string `json:"targetValue"`
				Content     string `json:"content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			patch := domain.PatchRequest{
				Operation:   args.Operation,
				Target:      args.Target,
				TargetValue: args.TargetValue,
				Content:     args.Content,
			}

			if err := r.GetNoteService().PatchNote(ctx, args.Path, patch); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"path":    args.Path,
				"message": fmt.Sprintf("Note patched successfully: %s", args.Path),
			}, nil
		},
	)
}
