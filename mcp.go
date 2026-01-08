package clawde

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// MCPServerConfig configures an external MCP server.
type MCPServerConfig struct {
	// Type is the server type: "stdio", "sse", or "sdk".
	Type string

	// Command is the command to run (stdio servers).
	Command string

	// Args are command line arguments (stdio servers).
	Args []string

	// Env are environment variables for the server.
	Env map[string]string

	// URL is the server URL (sse servers).
	URL string

	// Headers are HTTP headers (sse servers).
	Headers map[string]string
}

// MCPServer represents an in-process MCP server.
type MCPServer struct {
	Name  string
	Tools []*MCPTool
}

// MCPTool represents a tool provided by an MCP server.
type MCPTool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
	Handler     ToolHandler
}

// ToolHandler handles tool invocations.
type ToolHandler func(ctx context.Context, input json.RawMessage) (*ToolResult, error)

// ToolResult is the result of a tool invocation.
type ToolResult struct {
	Content []ToolContent
	IsError bool
}

// ToolContent represents content in a tool result.
type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     []byte `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// NewMCPServer creates a new in-process MCP server.
func NewMCPServer(name string) *MCPServer {
	return &MCPServer{
		Name:  name,
		Tools: make([]*MCPTool, 0),
	}
}

// AddTool adds a tool to the MCP server.
func (s *MCPServer) AddTool(name, description string, schema any, handler ToolHandler) {
	schemaJSON, _ := json.Marshal(generateSchema(schema))
	s.Tools = append(s.Tools, &MCPTool{
		Name:        name,
		Description: description,
		InputSchema: schemaJSON,
		Handler:     handler,
	})
}

// Tool creates a typed tool with automatic schema generation.
func Tool[T any](name, description string, handler func(ctx context.Context, input T) (string, error)) *MCPTool {
	var zero T
	schemaJSON, _ := json.Marshal(generateSchema(zero))

	return &MCPTool{
		Name:        name,
		Description: description,
		InputSchema: schemaJSON,
		Handler: func(ctx context.Context, raw json.RawMessage) (*ToolResult, error) {
			var input T
			if err := json.Unmarshal(raw, &input); err != nil {
				return ErrorResult(fmt.Sprintf("invalid input: %v", err)), nil
			}
			result, err := handler(ctx, input)
			if err != nil {
				return ErrorResult(err.Error()), nil
			}
			return TextResult(result), nil
		},
	}
}

// TextResult creates a text tool result.
func TextResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{Type: "text", Text: text}},
	}
}

// ErrorResult creates an error tool result.
func ErrorResult(msg string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{Type: "text", Text: msg}},
		IsError: true,
	}
}

// ImageResult creates an image tool result.
func ImageResult(data []byte, mimeType string) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{{
			Type:     "image",
			Data:     data,
			MimeType: mimeType,
		}},
	}
}

// generateSchema generates a JSON schema from a Go type.
func generateSchema(v any) map[string]any {
	t := reflect.TypeOf(v)
	if t == nil {
		return map[string]any{"type": "object"}
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := map[string]any{
		"type": "object",
	}

	if t.Kind() == reflect.Struct {
		properties := make(map[string]any)
		required := make([]string, 0)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			name := field.Name
			isRequired := true

			if jsonTag != "" {
				parts := splitTag(jsonTag)
				if parts[0] != "" {
					name = parts[0]
				}
				for _, part := range parts[1:] {
					if part == "omitempty" {
						isRequired = false
					}
				}
			}

			propSchema := typeToSchema(field.Type)

			// Add description from tag
			if desc := field.Tag.Get("description"); desc != "" {
				propSchema["description"] = desc
			}

			properties[name] = propSchema

			if isRequired {
				required = append(required, name)
			}
		}

		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}
	}

	return schema
}

// typeToSchema converts a Go type to a JSON schema type.
func typeToSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice:
		return map[string]any{
			"type":  "array",
			"items": typeToSchema(t.Elem()),
		}
	case reflect.Map:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": typeToSchema(t.Elem()),
		}
	case reflect.Ptr:
		return typeToSchema(t.Elem())
	case reflect.Struct:
		return generateSchema(reflect.New(t).Elem().Interface())
	default:
		return map[string]any{"type": "string"}
	}
}

// splitTag splits a struct tag value.
func splitTag(tag string) []string {
	var parts []string
	current := ""
	for _, c := range tag {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

// HandleMCPRequest handles an MCP request for SDK servers.
func (s *MCPServer) HandleMCPRequest(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	switch method {
	case "tools/list":
		tools := make([]map[string]any, len(s.Tools))
		for i, tool := range s.Tools {
			var schema map[string]any
			json.Unmarshal(tool.InputSchema, &schema)
			tools[i] = map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": schema,
			}
		}
		return json.Marshal(map[string]any{"tools": tools})

	case "tools/call":
		var req struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}

		for _, tool := range s.Tools {
			if tool.Name == req.Name {
				result, err := tool.Handler(ctx, req.Arguments)
				if err != nil {
					return json.Marshal(map[string]any{
						"content": []map[string]any{{"type": "text", "text": err.Error()}},
						"isError": true,
					})
				}
				return json.Marshal(map[string]any{
					"content": result.Content,
					"isError": result.IsError,
				})
			}
		}
		return nil, fmt.Errorf("tool not found: %s", req.Name)

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}
