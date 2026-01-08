package clawde

import (
	"encoding/json"
	"fmt"
)

// ParseMessage parses a raw JSON message into a typed Message.
func ParseMessage(data json.RawMessage) (Message, error) {
	// First, determine the message type
	var envelope struct {
		Type    string `json:"type"`
		Role    string `json:"role"`
		Subtype string `json:"subtype"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}

	// Route based on type/role
	switch {
	case envelope.Role == "user":
		return parseUserMessage(data)
	case envelope.Role == "assistant":
		return parseAssistantMessage(data)
	case envelope.Type == "system":
		return parseSystemMessage(data)
	case envelope.Type == "result":
		return parseResultMessage(data)
	case envelope.Type == "content_block_start",
		envelope.Type == "content_block_delta",
		envelope.Type == "content_block_stop",
		envelope.Type == "message_start",
		envelope.Type == "message_delta",
		envelope.Type == "message_stop":
		return parseStreamEvent(data)
	default:
		// Return as system message for unknown types
		return &SystemMessage{Type: envelope.Type, Subtype: envelope.Subtype}, nil
	}
}

// parseUserMessage parses a user message.
func parseUserMessage(data json.RawMessage) (*UserMessage, error) {
	var raw struct {
		Role    string            `json:"role"`
		Content json.RawMessage   `json:"content"`
		RawContent []json.RawMessage `json:"-"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}

	msg := &UserMessage{Role: raw.Role}

	// Content can be a string or array of blocks
	var contentStr string
	if err := json.Unmarshal(raw.Content, &contentStr); err == nil {
		msg.Content = []ContentBlock{&TextBlock{Text: contentStr}}
		return msg, nil
	}

	// Try as array
	var blocks []json.RawMessage
	if err := json.Unmarshal(raw.Content, &blocks); err != nil {
		return nil, &ParseError{Line: string(data), Err: fmt.Errorf("invalid content format")}
	}

	for _, block := range blocks {
		cb, err := parseContentBlock(block)
		if err != nil {
			return nil, err
		}
		msg.Content = append(msg.Content, cb)
	}

	return msg, nil
}

// parseAssistantMessage parses an assistant message.
func parseAssistantMessage(data json.RawMessage) (*AssistantMessage, error) {
	var raw struct {
		Role    string            `json:"role"`
		Content []json.RawMessage `json:"content"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}

	msg := &AssistantMessage{Role: raw.Role}

	for _, block := range raw.Content {
		cb, err := parseContentBlock(block)
		if err != nil {
			return nil, err
		}
		msg.Content = append(msg.Content, cb)
	}

	return msg, nil
}

// parseSystemMessage parses a system message.
func parseSystemMessage(data json.RawMessage) (*SystemMessage, error) {
	var msg SystemMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}
	return &msg, nil
}

// parseResultMessage parses a result message.
func parseResultMessage(data json.RawMessage) (*ResultMessage, error) {
	var msg ResultMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}
	return &msg, nil
}

// parseStreamEvent parses a streaming event.
func parseStreamEvent(data json.RawMessage) (*StreamEvent, error) {
	var event StreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}
	return &event, nil
}

// parseContentBlock parses a content block.
func parseContentBlock(data json.RawMessage) (ContentBlock, error) {
	var envelope struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}

	switch envelope.Type {
	case "text":
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, &ParseError{Line: string(data), Err: err}
		}
		return &block, nil

	case "thinking":
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, &ParseError{Line: string(data), Err: err}
		}
		return &block, nil

	case "tool_use":
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, &ParseError{Line: string(data), Err: err}
		}
		return &block, nil

	case "tool_result":
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, &ParseError{Line: string(data), Err: err}
		}
		return &block, nil

	case "image":
		var block ImageBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, &ParseError{Line: string(data), Err: err}
		}
		return &block, nil

	default:
		// Return text block for unknown types
		return &TextBlock{}, nil
	}
}
