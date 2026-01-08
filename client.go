package clawde

import (
	"context"
	"fmt"
	"sync"
)

// Client is the main Clawde SDK client.
type Client struct {
	opts      *Options
	transport Transport
	query     *Query

	mu        sync.RWMutex
	connected bool
}

// NewClient creates a new Clawde client.
func NewClient(opts ...Option) (*Client, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &Client{
		opts: options,
	}, nil
}

// Connect establishes the connection to Claude Code CLI.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return ErrAlreadyConnected
	}

	// Create transport
	c.transport = NewSubprocessTransport(c.opts)
	if err := c.transport.Start(ctx); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	// Create query handler
	c.query = newQuery(c.transport, c.opts)
	if err := c.query.Start(ctx); err != nil {
		c.transport.Close()
		return fmt.Errorf("failed to start query handler: %w", err)
	}

	// Initialize
	if err := c.query.Initialize(ctx); err != nil {
		c.query.Close()
		c.transport.Close()
		return fmt.Errorf("failed to initialize: %w", err)
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
		return nil, fmt.Errorf("failed to send prompt: %w", err)
	}

	// Create stream
	return newStream(ctx, c.query), nil
}

// Send sends a prompt without waiting for completion.
// Use Receive() to get responses.
func (c *Client) Send(ctx context.Context, prompt string) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	return c.query.SendPrompt(prompt)
}

// Receive returns a channel of messages.
// Messages are received until a ResultMessage is received.
func (c *Client) Receive(ctx context.Context) <-chan Message {
	ch := make(chan Message, 100)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-c.query.Messages():
				if !ok {
					return
				}

				select {
				case ch <- msg:
				case <-ctx.Done():
					return
				}

				// Stop after result message
				if _, isResult := msg.(*ResultMessage); isResult {
					return
				}
			case err := <-c.query.Errors():
				// Could wrap error in a message type
				_ = err
				return
			}
		}
	}()

	return ch
}

// Interrupt sends an interrupt signal to stop the current operation.
func (c *Client) Interrupt() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return ErrNotConnected
	}

	return c.query.Interrupt()
}

// SetPermissionMode changes the permission mode mid-conversation.
func (c *Client) SetPermissionMode(mode PermissionMode) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return ErrNotConnected
	}

	return c.query.SetPermissionMode(mode)
}

// SetModel changes the model mid-conversation.
func (c *Client) SetModel(model string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return ErrNotConnected
	}

	return c.query.SetModel(model)
}

// SessionID returns the current session ID.
func (c *Client) SessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.query != nil {
		return c.query.SessionID()
	}
	return ""
}

// Close shuts down the client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false

	var errs []error

	if c.query != nil {
		if err := c.query.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.transport != nil {
		if err := c.transport.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
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
