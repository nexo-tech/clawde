package clawde

import (
	"context"
	"encoding/json"
)

// Transport defines the interface for communicating with Claude.
type Transport interface {
	// Start initializes the transport and begins reading messages.
	Start(ctx context.Context) error

	// Write sends data to the transport.
	Write(data []byte) error

	// Messages returns a channel of incoming JSON messages.
	Messages() <-chan json.RawMessage

	// Errors returns a channel of transport errors.
	Errors() <-chan error

	// Close shuts down the transport.
	Close() error
}
