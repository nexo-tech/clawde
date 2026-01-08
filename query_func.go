package clawde

import (
	"context"
)

// Query performs a one-shot query and returns a stream.
// This is a convenience function that creates a client, connects, queries, and handles cleanup.
func Query(ctx context.Context, prompt string, opts ...Option) (*Stream, error) {
	client, err := NewClient(opts...)
	if err != nil {
		return nil, err
	}

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	stream, err := client.Query(ctx, prompt)
	if err != nil {
		client.Close()
		return nil, err
	}

	// Wrap stream to close client when done
	return &managedStream{
		Stream: stream,
		client: client,
	}, nil
}

// QueryText performs a query and returns just the text response.
func QueryText(ctx context.Context, prompt string, opts ...Option) (string, error) {
	stream, err := Query(ctx, prompt, opts...)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	return stream.CollectText()
}

// QueryResult performs a query and returns all messages.
func QueryResult(ctx context.Context, prompt string, opts ...Option) ([]Message, error) {
	stream, err := Query(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	return stream.Collect()
}

// managedStream wraps a Stream and closes the client when done.
type managedStream struct {
	*Stream
	client *Client
}

// Close closes both the stream and the underlying client.
func (s *managedStream) Close() error {
	s.Stream.Close()
	return s.client.Close()
}

// Next wraps the underlying Next and closes client when done.
func (s *managedStream) Next() bool {
	if !s.Stream.Next() {
		s.client.Close()
		return false
	}
	return true
}
