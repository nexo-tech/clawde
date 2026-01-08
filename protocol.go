package clawde

import (
	"encoding/json"
)

// ControlRequest represents a request from the CLI to the SDK.
type ControlRequest struct {
	Type      string          `json:"type"` // "control_request"
	RequestID string          `json:"request_id"`
	Request   json.RawMessage `json:"request"`
}

// ControlResponse represents a response from the SDK to the CLI.
type ControlResponse struct {
	Type     string `json:"type"` // "control_response"
	Response any    `json:"response"`
}

// ControlRequestSubtype identifies the type of control request.
type ControlRequestSubtype string

const (
	ControlInitialize       ControlRequestSubtype = "initialize"
	ControlCanUseTool       ControlRequestSubtype = "can_use_tool"
	ControlHookCallback     ControlRequestSubtype = "hook_callback"
	ControlMCPMessage       ControlRequestSubtype = "mcp_message"
	ControlInterrupt        ControlRequestSubtype = "interrupt"
	ControlSetPermissionMode ControlRequestSubtype = "set_permission_mode"
	ControlSetModel         ControlRequestSubtype = "set_model"
	ControlRewindFiles      ControlRequestSubtype = "rewind_files"
)

// InitializeRequest is sent to initialize the SDK.
type InitializeRequest struct {
	Subtype string         `json:"subtype"` // "initialize"
	Hooks   map[string]any `json:"hooks,omitempty"`
}

// InitializeResponse is the response to an initialize request.
type InitializeResponse struct {
	Subtype   string `json:"subtype"` // "success"
	RequestID string `json:"request_id"`
	Response  any    `json:"response,omitempty"`
}

// CanUseToolRequest is sent when a tool requests permission.
type CanUseToolRequest struct {
	Subtype     string                `json:"subtype"` // "can_use_tool"
	ToolName    string                `json:"tool_name"`
	Input       json.RawMessage       `json:"input"`
	Suggestions []PermissionUpdate    `json:"permission_suggestions,omitempty"`
	BlockedPath string                `json:"blocked_path,omitempty"`
}

// HookCallbackRequest is sent when a hook should be invoked.
type HookCallbackRequest struct {
	Subtype    string          `json:"subtype"` // "hook_callback"
	CallbackID string          `json:"callback_id"`
	Input      json.RawMessage `json:"input"`
	ToolUseID  string          `json:"tool_use_id,omitempty"`
}

// MCPMessageRequest is sent for MCP server communication.
type MCPMessageRequest struct {
	Subtype    string          `json:"subtype"` // "mcp_message"
	ServerName string          `json:"server_name"`
	Message    json.RawMessage `json:"message"`
}

// InterruptRequest is sent to interrupt the current operation.
type InterruptRequest struct {
	Subtype string `json:"subtype"` // "interrupt"
}

// SetPermissionModeRequest changes the permission mode.
type SetPermissionModeRequest struct {
	Subtype string `json:"subtype"` // "set_permission_mode"
	Mode    string `json:"mode"`
}

// SetModelRequest changes the model.
type SetModelRequest struct {
	Subtype string `json:"subtype"` // "set_model"
	Model   string `json:"model"`
}

// RewindFilesRequest rewinds files to a checkpoint.
type RewindFilesRequest struct {
	Subtype       string `json:"subtype"` // "rewind_files"
	UserMessageID string `json:"user_message_id"`
}

// parseControlRequest parses a control request.
func parseControlRequest(data json.RawMessage) (*ControlRequest, error) {
	var req ControlRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, &ParseError{Message: "failed to parse control request", Err: err}
	}
	return &req, nil
}

// parseControlRequestSubtype parses the subtype of a control request.
func parseControlRequestSubtype(request json.RawMessage) (ControlRequestSubtype, error) {
	var subtype struct {
		Subtype string `json:"subtype"`
	}
	if err := json.Unmarshal(request, &subtype); err != nil {
		return "", &ParseError{Message: "failed to parse control request subtype", Err: err}
	}
	return ControlRequestSubtype(subtype.Subtype), nil
}

// buildControlResponse builds a control response.
func buildControlResponse(requestID string, success bool, response any) ([]byte, error) {
	resp := map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"subtype":    "success",
			"request_id": requestID,
			"response":   response,
		},
	}
	if !success {
		resp["response"].(map[string]any)["subtype"] = "error"
	}
	return json.Marshal(resp)
}

// buildErrorResponse builds an error control response.
func buildErrorResponse(requestID string, errMsg string) ([]byte, error) {
	resp := map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"subtype":    "error",
			"request_id": requestID,
			"error":      errMsg,
		},
	}
	return json.Marshal(resp)
}
