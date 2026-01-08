// Package clawde provides a Go SDK for building AI agents with Claude Code capabilities.
package clawde

import (
	"encoding/json"
	"strings"
)

// Message represents any message in the conversation.
type Message interface {
	isMessage()
}

// UserMessage represents a message from the user.
type UserMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

func (UserMessage) isMessage() {}

// Text returns the concatenated text content of the message.
func (m *UserMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if tb, ok := block.(*TextBlock); ok {
			parts = append(parts, tb.Text)
		}
	}
	return strings.Join(parts, "")
}

// AssistantMessage represents a message from Claude.
type AssistantMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

func (AssistantMessage) isMessage() {}

// Text returns the concatenated text content of the message.
func (m *AssistantMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if tb, ok := block.(*TextBlock); ok {
			parts = append(parts, tb.Text)
		}
	}
	return strings.Join(parts, "")
}

// Thinking returns the concatenated thinking content of the message.
func (m *AssistantMessage) Thinking() string {
	var parts []string
	for _, block := range m.Content {
		if tb, ok := block.(*ThinkingBlock); ok {
			parts = append(parts, tb.Thinking)
		}
	}
	return strings.Join(parts, "")
}

// ToolUses returns all tool use blocks in the message.
func (m *AssistantMessage) ToolUses() []*ToolUseBlock {
	var uses []*ToolUseBlock
	for _, block := range m.Content {
		if tu, ok := block.(*ToolUseBlock); ok {
			uses = append(uses, tu)
		}
	}
	return uses
}

// SystemMessage represents a system message.
type SystemMessage struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
	Message string `json:"message,omitempty"`
}

func (SystemMessage) isMessage() {}

// ResultMessage represents the final result of a query.
type ResultMessage struct {
	Type         string  `json:"type"`
	Subtype      string  `json:"subtype,omitempty"`
	DurationMS   int64   `json:"duration_ms,omitempty"`
	DurationAPI  int64   `json:"duration_api_ms,omitempty"`
	NumTurns     int     `json:"num_turns,omitempty"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
	IsError      bool    `json:"is_error,omitempty"`
	SessionID    string  `json:"session_id,omitempty"`
	TotalCostUSD float64 `json:"total_cost_usd,omitempty"`
}

func (ResultMessage) isMessage() {}

// StreamEvent represents a streaming event during message generation.
type StreamEvent struct {
	Type      string          `json:"type"`
	Subtype   string          `json:"subtype,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Index     int             `json:"index,omitempty"`
	Delta     json.RawMessage `json:"delta,omitempty"`
}

func (StreamEvent) isMessage() {}

// ContentBlock represents a block of content in a message.
type ContentBlock interface {
	Type() string
}

// TextBlock represents a text content block.
type TextBlock struct {
	Text string `json:"text"`
}

func (TextBlock) Type() string { return "text" }

// ThinkingBlock represents Claude's thinking/reasoning.
type ThinkingBlock struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature,omitempty"`
}

func (ThinkingBlock) Type() string { return "thinking" }

// ToolUseBlock represents a tool invocation.
type ToolUseBlock struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

func (ToolUseBlock) Type() string { return "tool_use" }

// ToolResultBlock represents the result of a tool invocation.
type ToolResultBlock struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

func (ToolResultBlock) Type() string { return "tool_result" }

// ImageBlock represents an image content block.
type ImageBlock struct {
	Source ImageSource `json:"source"`
}

func (ImageBlock) Type() string { return "image" }

// ImageSource contains image data.
type ImageSource struct {
	Type      string `json:"type"` // "base64"
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}
