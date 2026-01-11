package clawde

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseMessage parses a raw JSON message into a typed Message.
func ParseMessage(data json.RawMessage) (Message, error) {
	// First, determine the message type
	// CLI sends: {"type": "user", "message": {...}} or {"type": "assistant", "message": {...}}
	var envelope struct {
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}

	// Route based on type
	switch envelope.Type {
	case "user":
		return parseUserMessage(data)
	case "assistant":
		return parseAssistantMessage(data)
	case "system":
		return parseSystemMessage(data)
	case "result":
		return parseResultMessage(data)
	case "stream_event":
		return parseStreamEvent(data)
	case "content_block_start",
		"content_block_delta",
		"content_block_stop",
		"message_start",
		"message_delta",
		"message_stop":
		return parseStreamEvent(data)
	default:
		// Return as system message for unknown types
		return &SystemMessage{Type: envelope.Type, Subtype: envelope.Subtype}, nil
	}
}

// parseUserMessage parses a user message.
// CLI sends: {"type": "user", "message": {"role": "user", "content": "..."}, "uuid": "...", "parent_tool_use_id": ...}
func parseUserMessage(data json.RawMessage) (*UserMessage, error) {
	var raw struct {
		Type             string `json:"type"`
		UUID             string `json:"uuid"`
		ParentToolUseID  *string `json:"parent_tool_use_id"`
		Message          struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}

	msg := &UserMessage{
		Role: raw.Message.Role,
		UUID: raw.UUID,
	}

	// Content can be a string or array of blocks
	var contentStr string
	if err := json.Unmarshal(raw.Message.Content, &contentStr); err == nil {
		msg.Content = []ContentBlock{&TextBlock{Text: contentStr}}
		return msg, nil
	}

	// Try as array
	var blocks []json.RawMessage
	if err := json.Unmarshal(raw.Message.Content, &blocks); err != nil {
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
// CLI sends: {"type": "assistant", "message": {"role": "assistant", "content": [...], "model": "..."}, "parent_tool_use_id": ...}
func parseAssistantMessage(data json.RawMessage) (*AssistantMessage, error) {
	var raw struct {
		Type            string  `json:"type"`
		ParentToolUseID *string `json:"parent_tool_use_id"`
		Message         struct {
			Role    string            `json:"role"`
			Content []json.RawMessage `json:"content"`
			Model   string            `json:"model"`
			Error   *string           `json:"error"`
		} `json:"message"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &ParseError{Line: string(data), Err: err}
	}

	msg := &AssistantMessage{
		Role:  raw.Message.Role,
		Model: raw.Message.Model,
	}

	for _, block := range raw.Message.Content {
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
		// ToolResultBlock content can be string or array of content blocks
		var raw struct {
			ToolUseID string          `json:"tool_use_id"`
			Content   json.RawMessage `json:"content"`
			IsError   bool            `json:"is_error,omitempty"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, &ParseError{Line: string(data), Err: err}
		}

		block := &ToolResultBlock{
			ToolUseID: raw.ToolUseID,
			Content:   raw.Content,
			IsError:   raw.IsError,
		}

		// Try to extract content as string
		var contentStr string
		if err := json.Unmarshal(raw.Content, &contentStr); err == nil {
			block.ContentString = contentStr
		} else {
			// Try as array of content blocks and extract text
			var contentBlocks []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(raw.Content, &contentBlocks); err == nil {
				var texts []string
				for _, cb := range contentBlocks {
					if cb.Type == "text" && cb.Text != "" {
						texts = append(texts, cb.Text)
					}
				}
				block.ContentString = strings.Join(texts, "\n")
			}
		}

		return block, nil

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
