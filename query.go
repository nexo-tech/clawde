package clawde

import (
	"context"
	"encoding/json"
	"sync"
)

// Query handles the control protocol for a single query.
type Query struct {
	transport  Transport
	opts       *Options
	msgCh      chan Message
	errCh      chan error
	doneCh     chan struct{}
	mu         sync.Mutex
	started    bool
	closed     bool
}

// NewQuery creates a new query handler.
func NewQuery(transport Transport, opts *Options) *Query {
	return &Query{
		transport: transport,
		opts:      opts,
		msgCh:     make(chan Message, 100),
		errCh:     make(chan error, 10),
		doneCh:    make(chan struct{}),
	}
}

// Start begins processing messages from the transport.
func (q *Query) Start(ctx context.Context) error {
	q.mu.Lock()
	if q.started {
		q.mu.Unlock()
		return nil
	}
	q.started = true
	q.mu.Unlock()

	go q.processLoop(ctx)
	return nil
}

// processLoop reads messages and handles control requests.
func (q *Query) processLoop(ctx context.Context) {
	defer close(q.msgCh)

	for {
		select {
		case <-ctx.Done():
			q.errCh <- ctx.Err()
			return

		case <-q.doneCh:
			return

		case raw, ok := <-q.transport.Messages():
			if !ok {
				return
			}

			// Check if this is a control request
			var envelope struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(raw, &envelope); err != nil {
				q.errCh <- &ParseError{Line: string(raw), Err: err}
				continue
			}

			if envelope.Type == "control_request" {
				q.handleControlRequest(ctx, raw)
				continue
			}

			// Parse as regular message
			msg, err := ParseMessage(raw)
			if err != nil {
				q.errCh <- err
				continue
			}

			select {
			case q.msgCh <- msg:
			case <-ctx.Done():
				return
			case <-q.doneCh:
				return
			}

		case err, ok := <-q.transport.Errors():
			if ok && err != nil {
				q.errCh <- err
			}
		}
	}
}

// handleControlRequest processes a control request and sends a response.
func (q *Query) handleControlRequest(ctx context.Context, raw json.RawMessage) {
	var req ControlRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		q.errCh <- &ParseError{Line: string(raw), Err: err}
		return
	}

	innerReq, err := parseControlRequest(&req)
	if err != nil {
		q.errCh <- err
		return
	}

	var response any

	switch r := innerReq.(type) {
	case *InitializeRequest:
		response = q.handleInitialize(r)

	case *CanUseToolRequest:
		response = q.handleCanUseTool(ctx, r)

	case *HookCallbackRequest:
		response = q.handleHookCallback(ctx, r)

	case *MCPMessageRequest:
		response = q.handleMCPMessage(ctx, r)

	default:
		q.errCh <- &ProtocolError{Message: "unknown request type"}
		return
	}

	// Send response
	resp := ControlResponse{
		Type:      "control_response",
		RequestID: req.RequestID,
		Response:  response,
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		q.errCh <- err
		return
	}

	if err := q.transport.Write(respJSON); err != nil {
		q.errCh <- err
	}
}

// handleInitialize handles initialization requests.
func (q *Query) handleInitialize(req *InitializeRequest) *InitializeResponse {
	return &InitializeResponse{Success: true}
}

// handleCanUseTool handles permission requests.
func (q *Query) handleCanUseTool(ctx context.Context, req *CanUseToolRequest) *CanUseToolResponse {
	// If no callback, allow by default
	if q.opts.PermissionCallback == nil {
		return &CanUseToolResponse{Allowed: true}
	}

	permReq := &PermissionRequest{
		ToolName:    req.ToolName,
		Input:       req.Input,
		Suggestions: req.Suggestions,
	}

	result := q.opts.PermissionCallback(ctx, permReq)

	switch r := result.(type) {
	case PermissionAllow:
		return &CanUseToolResponse{
			Allowed:      true,
			UpdatedInput: r.UpdatedInput,
		}
	case PermissionDeny:
		return &CanUseToolResponse{
			Allowed:   false,
			Reason:    r.Message,
			Interrupt: r.Interrupt,
		}
	default:
		return &CanUseToolResponse{Allowed: true}
	}
}

// handleHookCallback handles hook callbacks.
func (q *Query) handleHookCallback(ctx context.Context, req *HookCallbackRequest) *HookCallbackResponse {
	event := HookEvent(req.Event)
	matchers, ok := q.opts.Hooks[event]
	if !ok || len(matchers) == 0 {
		return &HookCallbackResponse{Continue: true}
	}

	for _, matcher := range matchers {
		// Check if matcher applies to this tool
		if matcher.ToolName != "*" && matcher.ToolName != req.Input.ToolName {
			continue
		}

		// Apply timeout if specified
		callCtx := ctx
		if matcher.Timeout > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(ctx, matcher.Timeout)
			defer cancel()
		}

		output, err := matcher.Callback(callCtx, req.Input)
		if err != nil {
			return &HookCallbackResponse{
				Continue: false,
				Decision: "block",
				Reason:   err.Error(),
			}
		}

		if output != nil && !output.Continue {
			return &HookCallbackResponse{
				Continue:      output.Continue,
				Decision:      output.Decision,
				Reason:        output.Reason,
				StopReason:    output.StopReason,
				ModifiedInput: output.ModifiedInput,
			}
		}
	}

	return &HookCallbackResponse{Continue: true}
}

// handleMCPMessage handles MCP server messages.
func (q *Query) handleMCPMessage(ctx context.Context, req *MCPMessageRequest) *MCPMessageResponse {
	server, ok := q.opts.SDKServers[req.ServerName]
	if !ok {
		return &MCPMessageResponse{
			Error: &MCPError{Code: -32601, Message: "server not found: " + req.ServerName},
		}
	}

	result, err := server.HandleMCPRequest(ctx, req.Method, req.Params)
	if err != nil {
		return &MCPMessageResponse{
			Error: &MCPError{Code: -32603, Message: err.Error()},
		}
	}

	return &MCPMessageResponse{Result: result}
}

// SendPrompt sends a user prompt.
func (q *Query) SendPrompt(prompt string) error {
	msg := map[string]any{
		"type":   "user_message",
		"prompt": prompt,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.transport.Write(data)
}

// Messages returns the message channel.
func (q *Query) Messages() <-chan Message {
	return q.msgCh
}

// Errors returns the error channel.
func (q *Query) Errors() <-chan error {
	return q.errCh
}

// Close shuts down the query.
func (q *Query) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}
	q.closed = true
	close(q.doneCh)
	return nil
}
