package clawde

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// QueryHandler handles the control protocol for a single query.
type QueryHandler struct {
	transport        Transport
	opts             *Options
	msgCh            chan Message
	errCh            chan error
	doneCh           chan struct{}
	mu               sync.Mutex
	started          bool
	closed           bool
	initRequestID    string
	initResponseCh   chan json.RawMessage
	pendingResponses map[string]chan json.RawMessage
}

// NewQueryHandler creates a new query handler.
func NewQueryHandler(transport Transport, opts *Options) *QueryHandler {
	return &QueryHandler{
		transport:        transport,
		opts:             opts,
		msgCh:            make(chan Message, 100),
		errCh:            make(chan error, 10),
		doneCh:           make(chan struct{}),
		initResponseCh:   make(chan json.RawMessage, 1),
		pendingResponses: make(map[string]chan json.RawMessage),
	}
}

// Start begins processing messages from the transport.
func (q *QueryHandler) Start(ctx context.Context) error {
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

// Initialize sends the initialization request to the CLI.
// This must be called before sending prompts.
func (q *QueryHandler) Initialize(ctx context.Context) error {
	// Build hooks configuration
	var hooksConfig map[string]any
	if len(q.opts.Hooks) > 0 {
		hooksConfig = make(map[string]any)
		for event, matchers := range q.opts.Hooks {
			var matcherConfigs []map[string]any
			for _, matcher := range matchers {
				matcherConfig := map[string]any{
					"matcher":         matcher.ToolName,
					"hookCallbackIds": []string{string(event) + "_callback"},
				}
				if matcher.Timeout > 0 {
					matcherConfig["timeout"] = matcher.Timeout.Milliseconds()
				}
				matcherConfigs = append(matcherConfigs, matcherConfig)
			}
			hooksConfig[string(event)] = matcherConfigs
		}
	}

	// Build MCP servers configuration
	var mcpServersConfig map[string]any
	if len(q.opts.SDKServers) > 0 {
		mcpServersConfig = make(map[string]any)
		for name, server := range q.opts.SDKServers {
			// Register SDK MCP server with tools
			var tools []map[string]any
			for _, tool := range server.Tools {
				tools = append(tools, map[string]any{
					"name":        tool.Name,
					"description": tool.Description,
					"inputSchema": tool.InputSchema,
				})
			}
			mcpServersConfig[name] = map[string]any{
				"type":  "sdk",
				"tools": tools,
			}
		}
	}

	// Build inner request
	innerRequest := map[string]any{
		"subtype": "initialize",
	}
	if hooksConfig != nil {
		innerRequest["hooks"] = hooksConfig
	}
	if mcpServersConfig != nil {
		innerRequest["mcp_servers"] = mcpServersConfig
	}

	// Generate request ID
	q.mu.Lock()
	q.initRequestID = fmt.Sprintf("init_%d", time.Now().UnixNano())
	requestID := q.initRequestID
	q.mu.Unlock()

	// Send control request with correct format
	controlRequest := map[string]any{
		"type":       "control_request",
		"request_id": requestID,
		"request":    innerRequest,
	}

	data, err := json.Marshal(controlRequest)
	if err != nil {
		return err
	}

	if err := q.transport.Write(data); err != nil {
		return err
	}

	// Wait for initialization response
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-q.initResponseCh:
		return nil
	case <-time.After(30 * time.Second):
		return &ProtocolError{Message: "initialization timeout"}
	}
}

// processLoop reads messages and handles control requests.
func (q *QueryHandler) processLoop(ctx context.Context) {
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

			// Handle control response (response to our requests)
			if envelope.Type == "control_response" {
				q.handleControlResponse(raw)
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
func (q *QueryHandler) handleControlRequest(ctx context.Context, raw json.RawMessage) {
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

// handleControlResponse handles responses to control requests we sent.
func (q *QueryHandler) handleControlResponse(raw json.RawMessage) {
	// The control_response structure from CLI is:
	// {"type": "control_response", "response": {"request_id": "...", "subtype": "success", "response": {...}}}
	var envelope struct {
		Type     string `json:"type"`
		Response struct {
			RequestID string          `json:"request_id"`
			Subtype   string          `json:"subtype"`
			Response  json.RawMessage `json:"response"`
			Error     string          `json:"error,omitempty"`
		} `json:"response"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		q.errCh <- &ParseError{Line: string(raw), Err: err}
		return
	}

	requestID := envelope.Response.RequestID

	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if this is the init response
	if requestID == q.initRequestID {
		select {
		case q.initResponseCh <- envelope.Response.Response:
		default:
		}
		return
	}

	// Check for other pending responses
	if ch, ok := q.pendingResponses[requestID]; ok {
		select {
		case ch <- envelope.Response.Response:
		default:
		}
		delete(q.pendingResponses, requestID)
	}
}

// handleInitialize handles initialization requests.
func (q *QueryHandler) handleInitialize(req *InitializeRequest) *InitializeResponse {
	return &InitializeResponse{Success: true}
}

// handleCanUseTool handles permission requests.
func (q *QueryHandler) handleCanUseTool(ctx context.Context, req *CanUseToolRequest) *CanUseToolResponse {
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
func (q *QueryHandler) handleHookCallback(ctx context.Context, req *HookCallbackRequest) *HookCallbackResponse {
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
		var cancel context.CancelFunc
		if matcher.Timeout > 0 {
			callCtx, cancel = context.WithTimeout(ctx, matcher.Timeout)
		}

		output, err := matcher.Callback(callCtx, req.Input)
		if cancel != nil {
			cancel()
		}
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
func (q *QueryHandler) handleMCPMessage(ctx context.Context, req *MCPMessageRequest) *MCPMessageResponse {
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
func (q *QueryHandler) SendPrompt(prompt string) error {
	// Format must match what the CLI expects:
	// {"type": "user", "message": {"role": "user", "content": "..."}, ...}
	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": prompt,
		},
		"parent_tool_use_id": nil,
		"session_id":         nil,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.transport.Write(data)
}

// Messages returns the message channel.
func (q *QueryHandler) Messages() <-chan Message {
	return q.msgCh
}

// Errors returns the error channel.
func (q *QueryHandler) Errors() <-chan error {
	return q.errCh
}

// Close shuts down the query.
func (q *QueryHandler) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}
	q.closed = true
	close(q.doneCh)
	return nil
}
