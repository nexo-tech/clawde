package clawde

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// MCPServerType identifies the MCP server connection type.
type MCPServerType string

const (
	// MCPServerTypeStdio connects via stdio subprocess.
	MCPServerTypeStdio MCPServerType = "stdio"

	// MCPServerTypeSSE connects via Server-Sent Events.
	MCPServerTypeSSE MCPServerType = "sse"

	// MCPServerTypeHTTP connects via HTTP.
	MCPServerTypeHTTP MCPServerType = "http"

	// MCPServerTypeSDK is an in-process SDK server.
	MCPServerTypeSDK MCPServerType = "sdk"
)

// MCPServerConfig configures an MCP server connection.
type MCPServerConfig struct {
	// Type is the connection type.
	Type MCPServerType `json:"type,omitempty"`

	// Command is the command to run (for stdio).
	Command string `json:"command,omitempty"`

	// Args are command arguments (for stdio).
	Args []string `json:"args,omitempty"`

	// Env are environment variables (for stdio).
	Env map[string]string `json:"env,omitempty"`

	// URL is the server URL (for SSE/HTTP).
	URL string `json:"url,omitempty"`

	// Headers are HTTP headers (for SSE/HTTP).
	Headers map[string]string `json:"headers,omitempty"`

	// SDKServer is the in-process server (for SDK type).
	SDKServer *MCPServer `json:"-"`
}

// ToJSON converts the config to JSON for the CLI.
func (c *MCPServerConfig) ToJSON() map[string]any {
	result := make(map[string]any)
	if c.Type != "" && c.Type != MCPServerTypeSDK {
		result["type"] = string(c.Type)
	}
	if c.Command != "" {
		result["command"] = c.Command
	}
	if len(c.Args) > 0 {
		result["args"] = c.Args
	}
	if len(c.Env) > 0 {
		result["env"] = c.Env
	}
	if c.URL != "" {
		result["url"] = c.URL
	}
	if len(c.Headers) > 0 {
		result["headers"] = c.Headers
	}
	return result
}

// MCPServer is an in-process MCP server.
type MCPServer struct {
	// Name is the server name.
	Name string

	// Version is the server version.
	Version string

	tools []*MCPTool
	mu    sync.RWMutex
}

// NewMCPServer creates a new in-process MCP server.
func NewMCPServer(name string) *MCPServer {
	return &MCPServer{
		Name:    name,
		Version: "1.0.0",
		tools:   make([]*MCPTool, 0),
	}
}

// AddTool adds a tool to the server.
func (s *MCPServer) AddTool(tool *MCPTool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, tool)
}

// Tools returns all registered tools.
func (s *MCPServer) Tools() []*MCPTool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*MCPTool, len(s.tools))
	copy(result, s.tools)
	return result
}

// HandleRequest handles an MCP JSON-RPC request.
func (s *MCPServer) HandleRequest(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	switch method {
	case "initialize":
		return s.handleInitialize(params)
	case "tools/list":
		return s.handleToolsList()
	case "tools/call":
		return s.handleToolsCall(ctx, params)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func (s *MCPServer) handleInitialize(params json.RawMessage) (json.RawMessage, error) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    s.Name,
			"version": s.Version,
		},
	}
	return json.Marshal(result)
}

func (s *MCPServer) handleToolsList() (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]map[string]any, len(s.tools))
	for i, tool := range s.tools {
		tools[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	result := map[string]any{
		"tools": tools,
	}
	return json.Marshal(result)
}

func (s *MCPServer) handleToolsCall(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var req struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	s.mu.RLock()
	var tool *MCPTool
	for _, t := range s.tools {
		if t.Name == req.Name {
			tool = t
			break
		}
	}
	s.mu.RUnlock()

	if tool == nil {
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}

	result, err := tool.Handler(ctx, req.Arguments)
	if err != nil {
		// Return error as tool result
		errResult := map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": err.Error(),
				},
			},
			"isError": true,
		}
		return json.Marshal(errResult)
	}

	return json.Marshal(result.ToJSON())
}

// MCPTool represents a tool in an MCP server.
type MCPTool struct {
	// Name is the tool name.
	Name string

	// Description describes what the tool does.
	Description string

	// InputSchema is the JSON schema for the tool input.
	InputSchema map[string]any

	// Handler processes tool invocations.
	Handler ToolHandler
}

// ToolHandler processes tool invocations.
type ToolHandler func(ctx context.Context, input json.RawMessage) (*ToolResult, error)

// ToolResult contains the result of a tool execution.
type ToolResult struct {
	// Content is the tool output content.
	Content []ToolContent

	// IsError indicates if this is an error result.
	IsError bool
}

// ToJSON converts the result to JSON.
func (r *ToolResult) ToJSON() map[string]any {
	content := make([]map[string]any, len(r.Content))
	for i, c := range r.Content {
		content[i] = c.ToJSON()
	}
	result := map[string]any{
		"content": content,
	}
	if r.IsError {
		result["isError"] = true
	}
	return result
}

// ToolContent represents content in a tool result.
type ToolContent struct {
	// Type is the content type ("text" or "image").
	Type string

	// Text is the text content (for type="text").
	Text string

	// Data is the binary data (for type="image").
	Data []byte

	// MimeType is the MIME type (for type="image").
	MimeType string
}

// ToJSON converts the content to JSON.
func (c *ToolContent) ToJSON() map[string]any {
	if c.Type == "image" {
		return map[string]any{
			"type":     "image",
			"data":     c.Data,
			"mimeType": c.MimeType,
		}
	}
	return map[string]any{
		"type": "text",
		"text": c.Text,
	}
}

// TextResult creates a text result.
func TextResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{Type: "text", Text: text}},
	}
}

// ErrorResult creates an error result.
func ErrorResult(message string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{Type: "text", Text: message}},
		IsError: true,
	}
}

// ImageResult creates an image result.
func ImageResult(data []byte, mimeType string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{Type: "image", Data: data, MimeType: mimeType}},
	}
}

// Tool creates a type-safe tool using Go generics.
// The input type T must be a struct with JSON tags.
func Tool[T any](name, description string, handler func(ctx context.Context, input T) (string, error)) *MCPTool {
	schema := generateSchema[T]()
	return &MCPTool{
		Name:        name,
		Description: description,
		InputSchema: schema,
		Handler: func(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
			var typedInput T
			if err := json.Unmarshal(input, &typedInput); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}
			result, err := handler(ctx, typedInput)
			if err != nil {
				return ErrorResult(err.Error()), nil
			}
			return TextResult(result), nil
		},
	}
}

// ToolWithResult creates a tool that returns a ToolResult directly.
func ToolWithResult[T any](name, description string, handler func(ctx context.Context, input T) (*ToolResult, error)) *MCPTool {
	schema := generateSchema[T]()
	return &MCPTool{
		Name:        name,
		Description: description,
		InputSchema: schema,
		Handler: func(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
			var typedInput T
			if err := json.Unmarshal(input, &typedInput); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}
			return handler(ctx, typedInput)
		},
	}
}

// generateSchema generates a JSON schema from a Go struct type.
func generateSchema[T any]() map[string]any {
	var zero T
	t := reflect.TypeOf(zero)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return map[string]any{
			"type": "object",
		}
	}

	properties := make(map[string]any)
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse json tag
		name := jsonTag
		omitempty := false
		if idx := len(jsonTag); idx > 0 {
			for j, c := range jsonTag {
				if c == ',' {
					name = jsonTag[:j]
					omitempty = true
					break
				}
			}
		}

		// Get description from tag
		desc := field.Tag.Get("description")

		prop := map[string]any{}

		// Determine JSON type from Go type
		switch field.Type.Kind() {
		case reflect.String:
			prop["type"] = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			prop["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			prop["type"] = "number"
		case reflect.Bool:
			prop["type"] = "boolean"
		case reflect.Slice, reflect.Array:
			prop["type"] = "array"
		case reflect.Map, reflect.Struct:
			prop["type"] = "object"
		default:
			prop["type"] = "string"
		}

		if desc != "" {
			prop["description"] = desc
		}

		properties[name] = prop

		if !omitempty {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}
