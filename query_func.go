package clawde

import (
	"context"
	"fmt"
)

// Query executes a one-shot query without maintaining a client.
// This is the simplest way to interact with Claude Code.
//
// Example:
//
//	stream, err := clawde.Query(ctx, "What is 2 + 2?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for stream.Next() {
//	    fmt.Print(stream.Current())
//	}
func Query(ctx context.Context, prompt string, opts ...Option) (*Stream, error) {
	client, err := NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	stream, err := client.Query(ctx, prompt)
	if err != nil {
		client.Close()
		return nil, err
	}

	// Wrap stream to close client when done
	return &clientOwningStream{
		Stream: stream,
		client: client,
	}, nil
}

// QueryText executes a query and returns only the text response.
// This is a convenience function for simple text responses.
//
// Example:
//
//	text, err := clawde.QueryText(ctx, "What is the capital of France?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(text)
func QueryText(ctx context.Context, prompt string, opts ...Option) (string, error) {
	stream, err := Query(ctx, prompt, opts...)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	return stream.CollectText()
}

// QueryResult executes a query and waits for the result.
// Returns the final ResultMessage with cost and usage information.
//
// Example:
//
//	result, err := clawde.QueryResult(ctx, "Do some work")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
func QueryResult(ctx context.Context, prompt string, opts ...Option) (*ResultMessage, error) {
	stream, err := Query(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	return stream.Wait()
}

// clientOwningStream wraps a Stream and closes the client when done.
type clientOwningStream struct {
	*Stream
	client *Client
}

// Close closes both the stream and the client.
func (s *clientOwningStream) Close() error {
	streamErr := s.Stream.Close()
	clientErr := s.client.Close()
	if streamErr != nil {
		return streamErr
	}
	return clientErr
}

// Next advances the stream and closes the client when done.
func (s *clientOwningStream) Next() bool {
	if !s.Stream.Next() {
		// Stream is done, close client
		s.client.Close()
		return false
	}
	return true
}
