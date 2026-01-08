package clawde

import (
	"context"
	"sync"
)

// Stream represents a streaming response from Claude.
type Stream struct {
	ctx    context.Context
	cancel context.CancelFunc
	query  *Query

	current Message
	err     error

	// Accumulated assistant message
	assistant *AssistantMessage
	result    *ResultMessage

	mu     sync.RWMutex
	closed bool
	done   bool
}

// newStream creates a new stream.
func newStream(ctx context.Context, query *Query) *Stream {
	ctx, cancel := context.WithCancel(ctx)
	return &Stream{
		ctx:    ctx,
		cancel: cancel,
		query:  query,
	}
}

// Next advances to the next message in the stream.
// Returns false when the stream is complete or an error occurred.
func (s *Stream) Next() bool {
	s.mu.Lock()
	if s.closed || s.done {
		s.mu.Unlock()
		return false
	}
	s.mu.Unlock()

	select {
	case <-s.ctx.Done():
		s.mu.Lock()
		s.err = s.ctx.Err()
		s.done = true
		s.mu.Unlock()
		return false

	case msg, ok := <-s.query.Messages():
		if !ok {
			s.mu.Lock()
			s.done = true
			s.mu.Unlock()
			return false
		}

		s.mu.Lock()
		s.current = msg
		s.accumulate(msg)

		// Check if this is the final message
		if result, isResult := msg.(*ResultMessage); isResult {
			s.result = result
			s.done = true
		}
		s.mu.Unlock()

		return true

	case err := <-s.query.Errors():
		s.mu.Lock()
		s.err = err
		s.done = true
		s.mu.Unlock()
		return false
	}
}

// Current returns the current message.
// Must call Next() first.
func (s *Stream) Current() Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Err returns any error that occurred during streaming.
func (s *Stream) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}

// Message returns the accumulated assistant message.
func (s *Stream) Message() *AssistantMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.assistant
}

// Result returns the result message, if available.
func (s *Stream) Result() *ResultMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result
}

// Close terminates the stream.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	s.cancel()
	return nil
}

// accumulate accumulates content from messages.
func (s *Stream) accumulate(msg Message) {
	switch m := msg.(type) {
	case *AssistantMessage:
		if s.assistant == nil {
			s.assistant = &AssistantMessage{
				Model: m.Model,
			}
		}
		// Append content blocks
		s.assistant.Content = append(s.assistant.Content, m.Content...)
		s.assistant.Model = m.Model

	case *StreamEvent:
		// Handle stream events for partial content
		if s.assistant == nil {
			s.assistant = &AssistantMessage{}
		}
		s.handleStreamEvent(m)
	}
}

// handleStreamEvent processes a stream event for accumulation.
func (s *Stream) handleStreamEvent(event *StreamEvent) {
	eventType, _ := event.Event["type"].(string)

	switch eventType {
	case "message_start":
		// Initialize message if not already
		if msg, ok := event.Event["message"].(map[string]any); ok {
			if model, ok := msg["model"].(string); ok {
				s.assistant.Model = model
			}
		}

	case "content_block_start":
		// Start a new content block
		if block, ok := event.Event["content_block"].(map[string]any); ok {
			blockType, _ := block["type"].(string)
			switch blockType {
			case "text":
				s.assistant.Content = append(s.assistant.Content, &TextBlock{})
			case "thinking":
				s.assistant.Content = append(s.assistant.Content, &ThinkingBlock{})
			case "tool_use":
				id, _ := block["id"].(string)
				name, _ := block["name"].(string)
				s.assistant.Content = append(s.assistant.Content, &ToolUseBlock{
					ID:   id,
					Name: name,
				})
			}
		}

	case "content_block_delta":
		// Update existing content block
		index, ok := event.Event["index"].(float64)
		if !ok || int(index) >= len(s.assistant.Content) {
			return
		}

		delta, ok := event.Event["delta"].(map[string]any)
		if !ok {
			return
		}

		idx := int(index)
		block := s.assistant.Content[idx]

		deltaType, _ := delta["type"].(string)
		switch deltaType {
		case "text_delta":
			if tb, ok := block.(*TextBlock); ok {
				if text, ok := delta["text"].(string); ok {
					tb.Text += text
				}
			}
		case "thinking_delta":
			if tb, ok := block.(*ThinkingBlock); ok {
				if thinking, ok := delta["thinking"].(string); ok {
					tb.Thinking += thinking
				}
			}
		case "input_json_delta":
			// Tool use input accumulation
			// This is complex as it involves partial JSON
			// For now, we'll skip this and rely on final tool_use blocks
		}
	}
}

// Collect collects all messages from the stream into a slice.
func (s *Stream) Collect() ([]Message, error) {
	var messages []Message
	for s.Next() {
		messages = append(messages, s.Current())
	}
	return messages, s.Err()
}

// CollectText collects all text content from the stream.
func (s *Stream) CollectText() (string, error) {
	var text string
	for s.Next() {
		if am, ok := s.Current().(*AssistantMessage); ok {
			text += am.Text()
		}
	}
	return text, s.Err()
}

// Wait waits for the stream to complete and returns the result.
func (s *Stream) Wait() (*ResultMessage, error) {
	for s.Next() {
		// Consume all messages
	}
	return s.Result(), s.Err()
}
