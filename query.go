package clawde

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Query handles the control protocol and message routing.
type Query struct {
	transport Transport
	opts      *Options

	msgCh chan Message
	errCh chan error

	pendingRequests map[string]chan json.RawMessage
	pendingMu       sync.RWMutex

	hooks      map[HookEvent][]HookMatcher
	mcpServers map[string]*MCPServer
	canUseTool PermissionCallback

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu         sync.RWMutex
	started    bool
	sessionID  string
	initialized bool
}

// newQuery creates a new Query handler.
func newQuery(transport Transport, opts *Options) *Query {
	if opts == nil {
		opts = DefaultOptions()
	}

	ctx, cancel := context.WithCancel(context.Background())

	q := &Query{
		transport:       transport,
		opts:            opts,
		msgCh:           make(chan Message, 100),
		errCh:           make(chan error, 10),
		pendingRequests: make(map[string]chan json.RawMessage),
		hooks:           opts.Hooks,
		mcpServers:      make(map[string]*MCPServer),
		canUseTool:      opts.PermissionCallback,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Extract SDK MCP servers
	for name, cfg := range opts.MCPServers {
		if cfg.Type == MCPServerTypeSDK && cfg.SDKServer != nil {
			q.mcpServers[name] = cfg.SDKServer
		}
	}

	return q
}

// Start starts the query handler.
func (q *Query) Start(ctx context.Context) error {
	q.mu.Lock()
	if q.started {
		q.mu.Unlock()
		return ErrAlreadyConnected
	}
	q.started = true
	q.mu.Unlock()

	// Start the message router
	q.wg.Add(1)
	go q.routeMessages(ctx)

	return nil
}

// Messages returns the message channel.
func (q *Query) Messages() <-chan Message {
	return q.msgCh
}

// Errors returns the error channel.
func (q *Query) Errors() <-chan error {
	return q.errCh
}

// SendPrompt sends a prompt to the CLI.
func (q *Query) SendPrompt(prompt string) error {
	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role": "user",
			"content": []map[string]any{
				{
					"type": "text",
					"text": prompt,
				},
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt: %w", err)
	}

	return q.transport.Write(data)
}

// Initialize sends the initialize request and waits for response.
func (q *Query) Initialize(ctx context.Context) error {
	q.mu.Lock()
	if q.initialized {
		q.mu.Unlock()
		return nil
	}
	q.mu.Unlock()

	// Build hooks config for initialization
	hooksConfig := make(map[string]any)
	for event, matchers := range q.hooks {
		eventHooks := make([]map[string]any, 0)
		for _, m := range matchers {
			hookCfg := map[string]any{}
			if m.ToolName != "" && m.ToolName != "*" {
				hookCfg["matcher"] = m.ToolName
			}
			if m.Timeout > 0 {
				hookCfg["timeout"] = m.Timeout.Milliseconds()
			}
			eventHooks = append(eventHooks, hookCfg)
		}
		if len(eventHooks) > 0 {
			hooksConfig[string(event)] = eventHooks
		}
	}

	initReq := map[string]any{
		"type": "control_request",
		"request": map[string]any{
			"subtype": "initialize",
			"hooks":   hooksConfig,
		},
	}

	data, err := json.Marshal(initReq)
	if err != nil {
		return fmt.Errorf("failed to marshal initialize request: %w", err)
	}

	if err := q.transport.Write(data); err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	q.mu.Lock()
	q.initialized = true
	q.mu.Unlock()

	return nil
}

// Interrupt sends an interrupt signal.
func (q *Query) Interrupt() error {
	req := map[string]any{
		"type": "control_request",
		"request": map[string]any{
			"subtype": "interrupt",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal interrupt request: %w", err)
	}

	return q.transport.Write(data)
}

// SetPermissionMode changes the permission mode.
func (q *Query) SetPermissionMode(mode PermissionMode) error {
	req := map[string]any{
		"type": "control_request",
		"request": map[string]any{
			"subtype": "set_permission_mode",
			"mode":    string(mode),
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal set permission mode request: %w", err)
	}

	return q.transport.Write(data)
}

// SetModel changes the model.
func (q *Query) SetModel(model string) error {
	req := map[string]any{
		"type": "control_request",
		"request": map[string]any{
			"subtype": "set_model",
			"model":   model,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal set model request: %w", err)
	}

	return q.transport.Write(data)
}

// Close stops the query handler.
func (q *Query) Close() error {
	q.cancel()
	q.wg.Wait()
	close(q.msgCh)
	close(q.errCh)
	return nil
}

// routeMessages routes incoming messages to appropriate handlers.
func (q *Query) routeMessages(ctx context.Context) {
	defer q.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-q.ctx.Done():
			return
		case <-q.transport.Done():
			return
		case err := <-q.transport.Errors():
			select {
			case q.errCh <- err:
			default:
			}
		case data := <-q.transport.Messages():
			q.handleMessage(ctx, data)
		}
	}
}

// handleMessage processes a single message.
func (q *Query) handleMessage(ctx context.Context, data json.RawMessage) {
	// Check if it's a control request
	if IsControlRequest(data) {
		q.handleControlRequest(ctx, data)
		return
	}

	// Check if it's a control response
	if IsControlResponse(data) {
		q.handleControlResponse(data)
		return
	}

	// Parse as regular message
	msg, err := ParseMessage(data)
	if err != nil {
		select {
		case q.errCh <- err:
		default:
		}
		return
	}

	// Track session ID
	if result, ok := msg.(*ResultMessage); ok {
		q.mu.Lock()
		q.sessionID = result.SessionID
		q.mu.Unlock()
	}

	// Send to message channel
	select {
	case q.msgCh <- msg:
	case <-ctx.Done():
	case <-q.ctx.Done():
	}
}

// handleControlRequest handles an incoming control request.
func (q *Query) handleControlRequest(ctx context.Context, data json.RawMessage) {
	req, err := parseControlRequest(data)
	if err != nil {
		q.sendErrorResponse("", err.Error())
		return
	}

	subtype, err := parseControlRequestSubtype(req.Request)
	if err != nil {
		q.sendErrorResponse(req.RequestID, err.Error())
		return
	}

	var response any
	var respErr error

	switch subtype {
	case ControlCanUseTool:
		response, respErr = q.handleCanUseTool(ctx, req.Request)
	case ControlHookCallback:
		response, respErr = q.handleHookCallback(ctx, req.Request)
	case ControlMCPMessage:
		response, respErr = q.handleMCPMessage(ctx, req.Request)
	case ControlInitialize:
		response = map[string]any{
			"capabilities": map[string]any{
				"hooks":       len(q.hooks) > 0,
				"permissions": q.canUseTool != nil,
				"mcp":         len(q.mcpServers) > 0,
			},
		}
	default:
		respErr = fmt.Errorf("unknown control request subtype: %s", subtype)
	}

	if respErr != nil {
		q.sendErrorResponse(req.RequestID, respErr.Error())
		return
	}

	q.sendSuccessResponse(req.RequestID, response)
}

// handleCanUseTool handles a permission request.
func (q *Query) handleCanUseTool(ctx context.Context, request json.RawMessage) (any, error) {
	var req CanUseToolRequest
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("failed to parse can_use_tool request: %w", err)
	}

	// If no callback, allow by default
	if q.canUseTool == nil {
		return map[string]any{"behavior": "allow"}, nil
	}

	permReq := &PermissionRequest{
		ToolName:    req.ToolName,
		Input:       req.Input,
		Suggestions: req.Suggestions,
		BlockedPath: req.BlockedPath,
	}

	result := q.canUseTool(ctx, permReq)
	return result.ToJSON(), nil
}

// handleHookCallback handles a hook callback.
func (q *Query) handleHookCallback(ctx context.Context, request json.RawMessage) (any, error) {
	var req HookCallbackRequest
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("failed to parse hook_callback request: %w", err)
	}

	// Parse the hook input
	var hookInput HookInput
	if err := json.Unmarshal(req.Input, &hookInput); err != nil {
		return nil, fmt.Errorf("failed to parse hook input: %w", err)
	}

	// Find matching hooks
	matchers, ok := q.hooks[hookInput.Event]
	if !ok || len(matchers) == 0 {
		return ContinueHook().ToJSON(), nil
	}

	// Run matching hooks
	for _, matcher := range matchers {
		if !matcher.matchesHook(&hookInput) {
			continue
		}

		// Apply timeout if specified
		hookCtx := ctx
		if matcher.Timeout > 0 {
			var cancel context.CancelFunc
			hookCtx, cancel = context.WithTimeout(ctx, matcher.Timeout)
			defer cancel()
		}

		output, err := matcher.Callback(hookCtx, &hookInput)
		if err != nil {
			return nil, fmt.Errorf("hook callback error: %w", err)
		}

		if output != nil && !output.Continue {
			// Hook blocked execution
			return output.ToJSON(), nil
		}
	}

	return ContinueHook().ToJSON(), nil
}

// handleMCPMessage handles an MCP message for an SDK server.
func (q *Query) handleMCPMessage(ctx context.Context, request json.RawMessage) (any, error) {
	var req MCPMessageRequest
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("failed to parse mcp_message request: %w", err)
	}

	server, ok := q.mcpServers[req.ServerName]
	if !ok {
		return nil, fmt.Errorf("unknown MCP server: %s", req.ServerName)
	}

	// Parse the JSON-RPC message
	var rpcReq struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(req.Message, &rpcReq); err != nil {
		return nil, fmt.Errorf("failed to parse MCP JSON-RPC request: %w", err)
	}

	// Handle the request
	result, err := server.HandleRequest(ctx, rpcReq.Method, rpcReq.Params)
	if err != nil {
		return map[string]any{
			"error": map[string]any{
				"code":    -32603,
				"message": err.Error(),
			},
		}, nil
	}

	var resultAny any
	if err := json.Unmarshal(result, &resultAny); err != nil {
		resultAny = string(result)
	}

	return map[string]any{"result": resultAny}, nil
}

// handleControlResponse handles an incoming control response.
func (q *Query) handleControlResponse(data json.RawMessage) {
	var resp struct {
		Response struct {
			RequestID string          `json:"request_id"`
			Response  json.RawMessage `json:"response"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}

	q.pendingMu.RLock()
	ch, ok := q.pendingRequests[resp.Response.RequestID]
	q.pendingMu.RUnlock()

	if ok {
		select {
		case ch <- resp.Response.Response:
		default:
		}
	}
}

// sendSuccessResponse sends a success control response.
func (q *Query) sendSuccessResponse(requestID string, response any) {
	data, err := buildControlResponse(requestID, true, response)
	if err != nil {
		return
	}
	q.transport.Write(data)
}

// sendErrorResponse sends an error control response.
func (q *Query) sendErrorResponse(requestID string, errMsg string) {
	data, err := buildErrorResponse(requestID, errMsg)
	if err != nil {
		return
	}
	q.transport.Write(data)
}

// SessionID returns the current session ID.
func (q *Query) SessionID() string {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.sessionID
}

// sendControlRequest sends a control request and waits for response.
func (q *Query) sendControlRequest(ctx context.Context, requestID string, request any) (json.RawMessage, error) {
	// Create response channel
	respCh := make(chan json.RawMessage, 1)
	q.pendingMu.Lock()
	q.pendingRequests[requestID] = respCh
	q.pendingMu.Unlock()

	defer func() {
		q.pendingMu.Lock()
		delete(q.pendingRequests, requestID)
		q.pendingMu.Unlock()
	}()

	// Send request
	req := map[string]any{
		"type":       "control_request",
		"request_id": requestID,
		"request":    request,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control request: %w", err)
	}

	if err := q.transport.Write(data); err != nil {
		return nil, fmt.Errorf("failed to send control request: %w", err)
	}

	// Wait for response with timeout
	timeout := 60 * time.Second
	if q.opts.Timeout > 0 {
		timeout = q.opts.Timeout
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-q.ctx.Done():
		return nil, ErrInterrupted
	}
}
