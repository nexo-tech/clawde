// Package clawde provides a Go SDK for building AI agents with Claude Code capabilities.
//
// Clawde offers two main APIs:
//
// Simple Query API - For one-shot queries:
//
//	stream, err := clawde.Query(ctx, "What is 2 + 2?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for stream.Next() {
//	    fmt.Print(stream.Current())
//	}
//
// Client API - For multi-turn conversations:
//
//	client, err := clawde.NewClient(
//	    clawde.WithSystemPrompt("You are helpful."),
//	    clawde.WithAllowedTools("Read", "Write"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	stream, _ := client.Query(ctx, "Hello!")
//	for stream.Next() {
//	    fmt.Print(stream.Current())
//	}
package clawde

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Message is the interface implemented by all message types.
type Message interface {
	isMessage()
	// MessageType returns the type identifier for this message.
	MessageType() string
}

// UserMessage represents a message from the user.
type UserMessage struct {
	Content         []ContentBlock `json:"content"`
	UUID            string         `json:"uuid,omitempty"`
	ParentToolUseID string         `json:"parent_tool_use_id,omitempty"`
}

func (*UserMessage) isMessage()            {}
func (*UserMessage) MessageType() string   { return "user" }

// Text returns all text content concatenated.
func (m *UserMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if tb, ok := block.(*TextBlock); ok {
			parts = append(parts, tb.Text)
		}
	}
	return strings.Join(parts, "")
}

// AssistantMessage represents a response from Claude.
type AssistantMessage struct {
	Content         []ContentBlock        `json:"content"`
	Model           string                `json:"model,omitempty"`
	ParentToolUseID string                `json:"parent_tool_use_id,omitempty"`
	Error           AssistantMessageError `json:"error,omitempty"`
}

func (*AssistantMessage) isMessage()            {}
func (*AssistantMessage) MessageType() string   { return "assistant" }

// Text returns all text content concatenated.
func (m *AssistantMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if tb, ok := block.(*TextBlock); ok {
			parts = append(parts, tb.Text)
		}
	}
	return strings.Join(parts, "")
}

// Thinking returns all thinking content concatenated.
func (m *AssistantMessage) Thinking() string {
	var parts []string
	for _, block := range m.Content {
		if tb, ok := block.(*ThinkingBlock); ok {
			parts = append(parts, tb.Thinking)
		}
	}
	return strings.Join(parts, "")
}

// ToolUses returns all tool use blocks.
func (m *AssistantMessage) ToolUses() []*ToolUseBlock {
	var uses []*ToolUseBlock
	for _, block := range m.Content {
		if tb, ok := block.(*ToolUseBlock); ok {
			uses = append(uses, tb)
		}
	}
	return uses
}

// AssistantMessageError represents error types from the assistant.
type AssistantMessageError string

const (
	ErrorAuthFailed     AssistantMessageError = "authentication_failed"
	ErrorBilling        AssistantMessageError = "billing_error"
	ErrorRateLimit      AssistantMessageError = "rate_limit"
	ErrorInvalidRequest AssistantMessageError = "invalid_request"
	ErrorServer         AssistantMessageError = "server_error"
	ErrorUnknown        AssistantMessageError = "unknown"
)

// SystemMessage represents a system message with metadata.
type SystemMessage struct {
	Subtype string         `json:"subtype"`
	Data    map[string]any `json:"data,omitempty"`
}

func (*SystemMessage) isMessage()            {}
func (*SystemMessage) MessageType() string   { return "system" }

// ResultMessage contains the final result of a query.
type ResultMessage struct {
	Subtype          string         `json:"subtype"`
	DurationMS       int            `json:"duration_ms"`
	DurationAPIMS    int            `json:"duration_api_ms"`
	IsError          bool           `json:"is_error"`
	NumTurns         int            `json:"num_turns"`
	SessionID        string         `json:"session_id"`
	TotalCostUSD     float64        `json:"total_cost_usd,omitempty"`
	Usage            map[string]any `json:"usage,omitempty"`
	Result           string         `json:"result,omitempty"`
	StructuredOutput json.RawMessage `json:"structured_output,omitempty"`
}

func (*ResultMessage) isMessage()            {}
func (*ResultMessage) MessageType() string   { return "result" }

// StreamEvent represents a streaming event from the API.
type StreamEvent struct {
	UUID            string         `json:"uuid"`
	SessionID       string         `json:"session_id"`
	Event           map[string]any `json:"event"`
	ParentToolUseID string         `json:"parent_tool_use_id,omitempty"`
}

func (*StreamEvent) isMessage()            {}
func (*StreamEvent) MessageType() string   { return "stream_event" }

// ContentBlock is the interface for message content blocks.
type ContentBlock interface {
	isContentBlock()
	// BlockType returns the type identifier for this block.
	BlockType() string
}

// TextBlock contains text content.
type TextBlock struct {
	Text string `json:"text"`
}

func (*TextBlock) isContentBlock()        {}
func (*TextBlock) BlockType() string      { return "text" }

func (b *TextBlock) String() string {
	return b.Text
}

// ThinkingBlock contains Claude's thinking process.
type ThinkingBlock struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature,omitempty"`
}

func (*ThinkingBlock) isContentBlock()        {}
func (*ThinkingBlock) BlockType() string      { return "thinking" }

func (b *ThinkingBlock) String() string {
	return fmt.Sprintf("[thinking: %s]", b.Thinking)
}

// ToolUseBlock represents a tool invocation request.
type ToolUseBlock struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

func (*ToolUseBlock) isContentBlock()        {}
func (*ToolUseBlock) BlockType() string      { return "tool_use" }

func (b *ToolUseBlock) String() string {
	return fmt.Sprintf("[tool_use: %s]", b.Name)
}

// InputAs unmarshals the tool input into the provided type.
func (b *ToolUseBlock) InputAs(v any) error {
	return json.Unmarshal(b.Input, v)
}

// ToolResultBlock contains the result of a tool execution.
type ToolResultBlock struct {
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

func (*ToolResultBlock) isContentBlock()        {}
func (*ToolResultBlock) BlockType() string      { return "tool_result" }

func (b *ToolResultBlock) String() string {
	if b.IsError {
		return fmt.Sprintf("[tool_result: error for %s]", b.ToolUseID)
	}
	return fmt.Sprintf("[tool_result: %s]", b.ToolUseID)
}

// ContentString returns the content as a string if it's a simple string value.
func (b *ToolResultBlock) ContentString() string {
	var s string
	if err := json.Unmarshal(b.Content, &s); err == nil {
		return s
	}
	return string(b.Content)
}

// ImageBlock contains image content.
type ImageBlock struct {
	Source ImageSource `json:"source"`
}

func (*ImageBlock) isContentBlock()        {}
func (*ImageBlock) BlockType() string      { return "image" }

func (b *ImageBlock) String() string {
	return fmt.Sprintf("[image: %s]", b.Source.MediaType)
}

// ImageSource contains image data.
type ImageSource struct {
	Type      string `json:"type"` // "base64"
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// rawMessage is used for JSON unmarshaling of messages.
type rawMessage struct {
	Type            string          `json:"type"`
	Content         json.RawMessage `json:"content,omitempty"`
	Model           string          `json:"model,omitempty"`
	UUID            string          `json:"uuid,omitempty"`
	SessionID       string          `json:"session_id,omitempty"`
	ParentToolUseID string          `json:"parent_tool_use_id,omitempty"`
	Error           string          `json:"error,omitempty"`
	Subtype         string          `json:"subtype,omitempty"`
	Data            json.RawMessage `json:"data,omitempty"`
	DurationMS      int             `json:"duration_ms,omitempty"`
	DurationAPIMS   int             `json:"duration_api_ms,omitempty"`
	IsError         bool            `json:"is_error,omitempty"`
	NumTurns        int             `json:"num_turns,omitempty"`
	TotalCostUSD    float64         `json:"total_cost_usd,omitempty"`
	Usage           json.RawMessage `json:"usage,omitempty"`
	Result          string          `json:"result,omitempty"`
	StructuredOutput json.RawMessage `json:"structured_output,omitempty"`
	Event           json.RawMessage `json:"event,omitempty"`
}

// rawContentBlock is used for JSON unmarshaling of content blocks.
type rawContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Signature string          `json:"signature,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	Source    json.RawMessage `json:"source,omitempty"`
}
