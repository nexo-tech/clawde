package clawde

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Session provides a V2-style API for multi-turn conversations with Claude.
// It wraps a Client and provides a cleaner send/stream pattern.
type Session struct {
	client    *Client
	sessionID string
	mu        sync.RWMutex
	closed    bool
}

// CreateSession creates a new session with the given options.
func CreateSession(ctx context.Context, opts ...Option) (*Session, error) {
	client, err := NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return &Session{
		client: client,
	}, nil
}

// ResumeSession resumes an existing session by ID.
func ResumeSession(ctx context.Context, sessionID string, opts ...Option) (*Session, error) {
	// Add resume option
	opts = append(opts, WithResumeConversation(sessionID))

	client, err := NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return &Session{
		client:    client,
		sessionID: sessionID,
	}, nil
}

// Send sends a message to the session and waits for it to be delivered.
func (s *Session) Send(ctx context.Context, message string) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return ErrSessionClosed
	}
	s.mu.RUnlock()

	return s.client.Send(ctx, message)
}

// Stream returns a channel of messages from the session.
// Call this after Send() to receive Claude's response.
func (s *Session) Stream(ctx context.Context) (*Stream, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, ErrSessionClosed
	}
	s.mu.RUnlock()

	// Get the message channel from the query handler
	msgCh := s.client.Receive(ctx)
	errCh := make(chan error, 1)

	return NewStream(ctx, msgCh, errCh), nil
}

// SessionID returns the current session ID.
// This may be empty until the first response is received.
func (s *Session) SessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID
}

// Close closes the session and releases resources.
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Prompt is a convenience function for one-shot queries.
// It creates a session, sends the prompt, collects the response, and closes the session.
func Prompt(ctx context.Context, prompt string, opts ...Option) (*PromptResult, error) {
	session, err := CreateSession(ctx, opts...)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	if err := session.Send(ctx, prompt); err != nil {
		return nil, err
	}

	stream, err := session.Stream(ctx)
	if err != nil {
		return nil, err
	}

	result := &PromptResult{}

	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *AssistantMessage:
			result.Text += m.Text()
		case *ResultMessage:
			result.Subtype = m.Subtype
			result.TotalCostUSD = m.TotalCostUSD
			result.DurationMS = m.DurationMS
			result.NumTurns = m.NumTurns
			result.SessionID = m.SessionID
		}
	}

	if stream.Err() != nil {
		return nil, stream.Err()
	}

	return result, nil
}

// PromptResult contains the result of a one-shot prompt.
type PromptResult struct {
	Text         string  // The response text
	Subtype      string  // "success" or "error"
	TotalCostUSD float64 // Total cost in USD
	DurationMS   int64   // Duration in milliseconds
	NumTurns     int     // Number of conversation turns
	SessionID    string  // Session ID for resuming
}

// ErrSessionClosed is returned when trying to use a closed session.
var ErrSessionClosed = errors.New("clawde: session closed")
