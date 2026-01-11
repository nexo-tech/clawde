package clawde

import (
	"context"
	"sync"
)

// Client provides a high-level interface for interacting with Claude.
type Client struct {
	opts      *Options
	transport Transport
	query     *QueryHandler
	mu        sync.RWMutex
	connected bool
}

// NewClient creates a new Claude client.
func NewClient(opts ...Option) (*Client, error) {
	options := applyOptions(opts)

	return &Client{
		opts: options,
	}, nil
}

// Connect establishes a connection to Claude.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return ErrAlreadyConnected
	}

	// Create transport
	c.transport = NewSubprocessTransport(c.opts)

	// Start transport
	if err := c.transport.Start(ctx); err != nil {
		return err
	}

	// Create and start query handler
	c.query = NewQueryHandler(c.transport, c.opts)
	if err := c.query.Start(ctx); err != nil {
		c.transport.Close()
		return err
	}

	// Send initialization request
	if err := c.query.Initialize(ctx); err != nil {
		c.transport.Close()
		return err
	}

	c.connected = true
	return nil
}

// Query sends a prompt and returns a stream of responses.
func (c *Client) Query(ctx context.Context, prompt string) (*Stream, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, ErrNotConnected
	}
	c.mu.RUnlock()

	// Send the prompt
	if err := c.query.SendPrompt(prompt); err != nil {
		return nil, err
	}

	// Create stream
	return NewStream(ctx, c.query.Messages(), c.query.Errors()), nil
}

// Send sends a prompt without waiting for the response.
func (c *Client) Send(ctx context.Context, prompt string) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	return c.query.SendPrompt(prompt)
}

// Receive returns a channel of messages from the current query.
func (c *Client) Receive(ctx context.Context) <-chan Message {
	c.mu.RLock()
	if !c.connected || c.query == nil {
		c.mu.RUnlock()
		ch := make(chan Message)
		close(ch)
		return ch
	}
	c.mu.RUnlock()

	return c.query.Messages()
}

// Interrupt sends an interrupt signal to Claude.
func (c *Client) Interrupt() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return ErrNotConnected
	}

	// Send interrupt message
	return c.transport.Write([]byte(`{"type":"interrupt"}`))
}

// Close shuts down the client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false

	if c.query != nil {
		c.query.Close()
	}

	if c.transport != nil {
		return c.transport.Close()
	}

	return nil
}

// IsConnected returns whether the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Options returns the client options.
func (c *Client) Options() *Options {
	return c.opts
}
