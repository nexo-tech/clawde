package clawde

import (
	"context"
	"encoding/json"
)

// Transport abstracts the communication layer with Claude Code CLI.
type Transport interface {
	// Start initializes the transport connection.
	Start(ctx context.Context) error

	// Write sends data to the transport.
	Write(data []byte) error

	// Messages returns a channel for receiving messages.
	Messages() <-chan json.RawMessage

	// Errors returns a channel for receiving errors.
	Errors() <-chan error

	// Done returns a channel that's closed when the transport is done.
	Done() <-chan struct{}

	// Close shuts down the transport.
	Close() error
}
