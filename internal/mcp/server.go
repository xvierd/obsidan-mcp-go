// Package mcp provides a minimal MCP (Model Context Protocol) server implementation.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Request represents an MCP JSON-RPC request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents an MCP JSON-RPC response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC error.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Tool represents an MCP tool.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Handler is a function that handles a tool call.
type Handler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Server represents an MCP server.
type Server struct {
	tools    map[string]*Tool
	handlers map[string]Handler
	reader   io.Reader
	writer   io.Writer
	logger   *slog.Logger
}

// NewServer creates a new MCP server.
func NewServer(logger *slog.Logger) *Server {
	return &Server{
		tools:    make(map[string]*Tool),
		handlers: make(map[string]Handler),
		reader:   io.Reader(nil),
		writer:   io.Writer(nil),
		logger:   logger,
	}
}

// SetIO sets the input/output for the server (for testing).
func (s *Server) SetIO(reader io.Reader, writer io.Writer) {
	s.reader = reader
	s.writer = writer
}

// RegisterTool registers a tool with the server.
func (s *Server) RegisterTool(tool *Tool, handler Handler) {
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
	s.logger.Debug("registered tool", "name", tool.Name)
}

// Run starts the MCP server and listens for requests.
func (s *Server) Run(ctx context.Context) error {
	reader := s.reader
	if reader == nil {
		reader = io.Reader(nil) // Will use stdin
	}

	writer := s.writer
	if writer == nil {
		writer = io.Writer(nil) // Will use stdout
	}

	// Use stdin/stdout if not set
	var scanner *bufio.Scanner
	if reader != nil {
		scanner = bufio.NewScanner(reader)
	} else {
		// Default to stdin
		scanner = bufio.NewScanner(os.Stdin)
	}

	out := writer
	if out == nil {
		out = io.Writer(nil) // Will use stdout
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		s.logger.Debug("received request", "line", line)

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, ParseError, "Parse error", err.Error())
			continue
		}

		resp := s.handleRequest(ctx, &req)
		if resp != nil {
			if err := s.sendResponse(out, resp); err != nil {
				s.logger.Error("failed to send response", "error", err)
			}
		}
	}

	return scanner.Err()
}

// handleRequest processes a single request.
func (s *Server) handleRequest(ctx context.Context, req *Request) *Response {
	if req.JSONRPC != "2.0" {
		return s.makeError(req.ID, InvalidRequest, "Invalid JSON-RPC version", nil)
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(ctx, req)
	default:
		return s.makeError(req.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}
}

// handleInitialize handles the initialize method.
func (s *Server) handleInitialize(req *Request) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "mcp-obsidian-go",
				"version": "0.1.0",
			},
		},
	}
}

// handleToolsList handles the tools/list method.
func (s *Server) handleToolsList(req *Request) *Response {
	tools := make([]*Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// handleToolCall handles the tools/call method.
func (s *Server) handleToolCall(ctx context.Context, req *Request) *Response {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.makeError(req.ID, InvalidParams, "Invalid params", err.Error())
	}

	handler, ok := s.handlers[params.Name]
	if !ok {
		return s.makeError(req.ID, MethodNotFound, fmt.Sprintf("Tool not found: %s", params.Name), nil)
	}

	result, err := handler(ctx, params.Arguments)
	if err != nil {
		s.logger.Error("tool execution failed", "tool", params.Name, "error", err)
		return s.makeError(req.ID, InternalError, err.Error(), nil)
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": resultToString(result),
				},
			},
		},
	}
}

// makeError creates an error response.
func (s *Server) makeError(id interface{}, code int, message string, data interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// sendError sends an error response.
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) {
	resp := s.makeError(id, code, message, data)
	if err := s.sendResponse(s.writer, resp); err != nil {
		s.logger.Error("failed to send error response", "error", err)
	}
}

// sendResponse sends a response.
func (s *Server) sendResponse(w io.Writer, resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	// Use stdout if writer not set
	if w == nil {
		fmt.Println(string(data))
	} else {
		fmt.Fprintln(w, string(data))
	}

	return nil
}

// resultToString converts a result to a string.
func resultToString(result interface{}) string {
	switch v := result.(type) {
	case string:
		return v
	default:
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(data)
	}
}
