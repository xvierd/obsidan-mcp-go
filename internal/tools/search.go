package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// RegisterSearchTools registers all search-related tools.
func RegisterSearchTools(r *Registry) {
	// search_notes tool
	r.Register(
		&Tool{
			Name:        "search_notes",
			Description: "Search for notes containing specific text using simple text search",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "Text to search for in notes"
					}
				},
				"required": ["query"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Query == "" {
				return nil, fmt.Errorf("query is required")
			}

			results, err := r.GetClient().Search(ctx, args.Query)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"count":   len(results),
				"results": results,
			}, nil
		},
	)

	// complex_search tool
	r.Register(
		&Tool{
			Name:        "complex_search",
			Description: "Perform complex searches using JsonLogic queries. Supports AND, OR, NOT operators and various conditions.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "object",
						"description": "JsonLogic query object defining search conditions"
					}
				},
				"required": ["query"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Query map[string]interface{} `json:"query"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Query == nil {
				return nil, fmt.Errorf("query is required")
			}

			results, err := r.GetClient().ComplexSearch(ctx, args.Query)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"count":   len(results),
				"results": results,
			}, nil
		},
	)

	// dataview_query tool
	r.Register(
		&Tool{
			Name:        "dataview_query",
			Description: "Execute a Dataview query against the vault. Dataview is a powerful query language for Obsidian.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "Dataview query string (e.g., 'TABLE file.tags FROM #tag')"
					}
				},
				"required": ["query"]
			}`),
		},
		func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			var args struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if args.Query == "" {
				return nil, fmt.Errorf("query is required")
			}

			result, err := r.GetClient().DataviewQuery(ctx, args.Query)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"headers": result.Headers,
				"rows":    result.Rows,
				"count":   result.Count,
			}, nil
		},
	)
}
