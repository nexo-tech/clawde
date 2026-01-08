package clawde

import (
	"context"
	"strings"
)

// Stream provides iteration over query responses.
type Stream struct {
	ctx     context.Context
	cancel  context.CancelFunc
	msgCh   <-chan Message
	errCh   <-chan error
	current Message
	err     error
	done    bool
	message *AssistantMessage // accumulated message
	result  *ResultMessage    // final result
}

// NewStream creates a new response stream.
func NewStream(ctx context.Context, msgCh <-chan Message, errCh <-chan error) *Stream {
	ctx, cancel := context.WithCancel(ctx)
	return &Stream{
		ctx:     ctx,
		cancel:  cancel,
		msgCh:   msgCh,
		errCh:   errCh,
		message: &AssistantMessage{Role: "assistant"},
	}
}

// Next advances to the next message. Returns false when done.
func (s *Stream) Next() bool {
	if s.done {
		return false
	}

	select {
	case <-s.ctx.Done():
		s.err = s.ctx.Err()
		s.done = true
		return false

	case err := <-s.errCh:
		if err != nil {
			s.err = err
			s.done = true
			return false
		}

	case msg, ok := <-s.msgCh:
		if !ok {
			s.done = true
			return false
		}

		s.current = msg
		s.accumulate(msg)

		// Check for result message (end of stream)
		if result, ok := msg.(*ResultMessage); ok {
			s.result = result
			s.done = true
			return true
		}

		return true
	}

	return false
}

// accumulate adds content to the accumulated message.
func (s *Stream) accumulate(msg Message) {
	switch m := msg.(type) {
	case *AssistantMessage:
		s.message.Content = append(s.message.Content, m.Content...)
	}
}

// Current returns the current message.
func (s *Stream) Current() Message {
	return s.current
}

// Message returns the accumulated assistant message.
func (s *Stream) Message() *AssistantMessage {
	return s.message
}

// Result returns the final result message, if available.
func (s *Stream) Result() *ResultMessage {
	return s.result
}

// Text returns the accumulated text content.
func (s *Stream) Text() string {
	return s.message.Text()
}

// Err returns any error that occurred.
func (s *Stream) Err() error {
	return s.err
}

// Close cancels the stream.
func (s *Stream) Close() error {
	s.cancel()
	s.done = true
	return nil
}

// Done returns whether the stream is complete.
func (s *Stream) Done() bool {
	return s.done
}

// Collect reads all messages and returns them as a slice.
func (s *Stream) Collect() ([]Message, error) {
	var messages []Message

	for s.Next() {
		messages = append(messages, s.current)
	}

	return messages, s.err
}

// CollectText reads all messages and returns the concatenated text.
func (s *Stream) CollectText() (string, error) {
	var parts []string

	for s.Next() {
		if msg, ok := s.current.(*AssistantMessage); ok {
			parts = append(parts, msg.Text())
		}
	}

	return strings.Join(parts, ""), s.err
}

// Wait blocks until the stream is complete.
func (s *Stream) Wait() error {
	for s.Next() {
		// Consume all messages
	}
	return s.err
}
