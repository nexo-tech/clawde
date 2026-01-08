package clawde

import (
	"encoding/json"
	"fmt"
)

// ParseMessage parses a raw JSON message into a typed Message.
func ParseMessage(data json.RawMessage) (Message, error) {
	var raw rawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &ParseError{Message: "failed to parse message", Err: err}
	}

	switch raw.Type {
	case "user":
		return parseUserMessage(&raw)
	case "assistant":
		return parseAssistantMessage(&raw)
	case "system":
		return parseSystemMessage(&raw)
	case "result":
		return parseResultMessage(&raw)
	case "stream_event":
		return parseStreamEvent(&raw)
	default:
		return nil, &ParseError{Message: fmt.Sprintf("unknown message type: %s", raw.Type)}
	}
}

// parseUserMessage parses a user message.
func parseUserMessage(raw *rawMessage) (*UserMessage, error) {
	msg := &UserMessage{
		UUID:            raw.UUID,
		ParentToolUseID: raw.ParentToolUseID,
	}

	if len(raw.Content) > 0 {
		content, err := parseContent(raw.Content)
		if err != nil {
			return nil, err
		}
		msg.Content = content
	}

	return msg, nil
}

// parseAssistantMessage parses an assistant message.
func parseAssistantMessage(raw *rawMessage) (*AssistantMessage, error) {
	msg := &AssistantMessage{
		Model:           raw.Model,
		ParentToolUseID: raw.ParentToolUseID,
		Error:           AssistantMessageError(raw.Error),
	}

	if len(raw.Content) > 0 {
		content, err := parseContent(raw.Content)
		if err != nil {
			return nil, err
		}
		msg.Content = content
	}

	return msg, nil
}

// parseSystemMessage parses a system message.
func parseSystemMessage(raw *rawMessage) (*SystemMessage, error) {
	msg := &SystemMessage{
		Subtype: raw.Subtype,
	}

	if len(raw.Data) > 0 {
		if err := json.Unmarshal(raw.Data, &msg.Data); err != nil {
			return nil, &ParseError{Message: "failed to parse system message data", Err: err}
		}
	}

	return msg, nil
}

// parseResultMessage parses a result message.
func parseResultMessage(raw *rawMessage) (*ResultMessage, error) {
	msg := &ResultMessage{
		Subtype:          raw.Subtype,
		DurationMS:       raw.DurationMS,
		DurationAPIMS:    raw.DurationAPIMS,
		IsError:          raw.IsError,
		NumTurns:         raw.NumTurns,
		TotalCostUSD:     raw.TotalCostUSD,
		Result:           raw.Result,
		StructuredOutput: raw.StructuredOutput,
	}

	// Extract session ID from data if present
	if len(raw.Data) > 0 {
		var data struct {
			SessionID string `json:"session_id"`
		}
		if err := json.Unmarshal(raw.Data, &data); err == nil {
			msg.SessionID = data.SessionID
		}
	}

	if len(raw.Usage) > 0 {
		if err := json.Unmarshal(raw.Usage, &msg.Usage); err != nil {
			return nil, &ParseError{Message: "failed to parse usage data", Err: err}
		}
	}

	return msg, nil
}

// parseStreamEvent parses a stream event.
func parseStreamEvent(raw *rawMessage) (*StreamEvent, error) {
	msg := &StreamEvent{
		UUID:            raw.UUID,
		SessionID:       raw.SessionID,
		ParentToolUseID: raw.ParentToolUseID,
	}

	if len(raw.Event) > 0 {
		if err := json.Unmarshal(raw.Event, &msg.Event); err != nil {
			return nil, &ParseError{Message: "failed to parse stream event", Err: err}
		}
	}

	return msg, nil
}

// parseContent parses content blocks.
func parseContent(data json.RawMessage) ([]ContentBlock, error) {
	// First try to parse as an array
	var rawBlocks []rawContentBlock
	if err := json.Unmarshal(data, &rawBlocks); err != nil {
		// Try as a single string (simple content)
		var text string
		if err := json.Unmarshal(data, &text); err == nil {
			return []ContentBlock{&TextBlock{Text: text}}, nil
		}
		return nil, &ParseError{Message: "failed to parse content", Err: err}
	}

	blocks := make([]ContentBlock, 0, len(rawBlocks))
	for _, raw := range rawBlocks {
		block, err := parseContentBlock(&raw)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// parseContentBlock parses a single content block.
func parseContentBlock(raw *rawContentBlock) (ContentBlock, error) {
	switch raw.Type {
	case "text":
		return &TextBlock{Text: raw.Text}, nil

	case "thinking":
		return &ThinkingBlock{
			Thinking:  raw.Thinking,
			Signature: raw.Signature,
		}, nil

	case "tool_use":
		return &ToolUseBlock{
			ID:    raw.ID,
			Name:  raw.Name,
			Input: raw.Input,
		}, nil

	case "tool_result":
		return &ToolResultBlock{
			ToolUseID: raw.ToolUseID,
			Content:   raw.Content,
			IsError:   raw.IsError,
		}, nil

	case "image":
		block := &ImageBlock{}
		if len(raw.Source) > 0 {
			if err := json.Unmarshal(raw.Source, &block.Source); err != nil {
				return nil, &ParseError{Message: "failed to parse image source", Err: err}
			}
		}
		return block, nil

	default:
		// Return as text block with the raw type info
		return &TextBlock{Text: fmt.Sprintf("[unknown block type: %s]", raw.Type)}, nil
	}
}

// IsControlRequest checks if a message is a control request.
func IsControlRequest(data json.RawMessage) bool {
	var check struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &check); err != nil {
		return false
	}
	return check.Type == "control_request"
}

// IsControlResponse checks if a message is a control response.
func IsControlResponse(data json.RawMessage) bool {
	var check struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &check); err != nil {
		return false
	}
	return check.Type == "control_response"
}
