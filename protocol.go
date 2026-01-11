package clawde

import (
	"encoding/json"
)

// ControlRequest represents a request from Claude requiring a response.
type ControlRequest struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Request   json.RawMessage `json:"request"`
}

// ControlResponse is sent back in response to a ControlRequest.
type ControlResponse struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	Response  any    `json:"response"`
}

// InitializeRequest is sent to initialize the session.
type InitializeRequest struct {
	ProtocolVersion string         `json:"protocol_version"`
	Hooks           map[string]any `json:"hooks,omitempty"`
	MCPServers      map[string]any `json:"mcp_servers,omitempty"`
}

// InitializeResponse acknowledges initialization.
type InitializeResponse struct {
	Success bool `json:"success"`
}

// CanUseToolRequest asks if a tool can be used.
type CanUseToolRequest struct {
	ToolName    string             `json:"tool_name"`
	Input       json.RawMessage    `json:"input"`
	Suggestions []PermissionUpdate `json:"suggestions,omitempty"`
}

// CanUseToolResponse contains the permission decision.
type CanUseToolResponse struct {
	Allowed      bool            `json:"allowed"`
	Reason       string          `json:"reason,omitempty"`
	UpdatedInput json.RawMessage `json:"updated_input,omitempty"`
	Interrupt    bool            `json:"interrupt,omitempty"`
}

// HookCallbackRequest triggers a hook callback.
type HookCallbackRequest struct {
	CallbackID string     `json:"callback_id"`
	Event      string     `json:"event"`
	Input      *HookInput `json:"input"`
}

// HookCallbackResponse contains the hook result.
type HookCallbackResponse struct {
	Continue      bool            `json:"continue"`
	Decision      string          `json:"decision,omitempty"`
	Reason        string          `json:"reason,omitempty"`
	StopReason    string          `json:"stop_reason,omitempty"`
	ModifiedInput json.RawMessage `json:"modified_input,omitempty"`
}

// MCPMessageRequest routes a message to an MCP server.
type MCPMessageRequest struct {
	ServerName string          `json:"server_name"`
	Method     string          `json:"method"`
	Params     json.RawMessage `json:"params"`
}

// MCPMessageResponse contains the MCP server response.
type MCPMessageResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  *MCPError       `json:"error,omitempty"`
}

// MCPError represents an MCP error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// parseControlRequest parses the inner request of a ControlRequest.
func parseControlRequest(req *ControlRequest) (any, error) {
	// The CLI sends control requests with "subtype" field to identify the request type
	var envelope struct {
		Subtype string `json:"subtype"`
	}

	if err := json.Unmarshal(req.Request, &envelope); err != nil {
		return nil, err
	}

	switch envelope.Subtype {
	case "initialize":
		var init InitializeRequest
		if err := json.Unmarshal(req.Request, &init); err != nil {
			return nil, err
		}
		return &init, nil

	case "can_use_tool":
		var cut CanUseToolRequest
		if err := json.Unmarshal(req.Request, &cut); err != nil {
			return nil, err
		}
		return &cut, nil

	case "hook_callback":
		var hc HookCallbackRequest
		if err := json.Unmarshal(req.Request, &hc); err != nil {
			return nil, err
		}
		return &hc, nil

	case "mcp_message":
		var mcp MCPMessageRequest
		if err := json.Unmarshal(req.Request, &mcp); err != nil {
			return nil, err
		}
		return &mcp, nil

	default:
		return nil, &ProtocolError{Message: "unknown request subtype: " + envelope.Subtype}
	}
}
