package clawde

import (
	"context"
	"encoding/json"
)

// PermissionMode controls how tool permissions are handled.
type PermissionMode string

const (
	// PermissionDefault uses the default Claude permission behavior.
	PermissionDefault PermissionMode = "default"

	// PermissionAcceptEdits automatically accepts file edits.
	PermissionAcceptEdits PermissionMode = "acceptEdits"

	// PermissionPlan runs in plan-only mode (no tool execution).
	PermissionPlan PermissionMode = "plan"

	// PermissionBypassAll bypasses all permission checks.
	PermissionBypassAll PermissionMode = "bypassPermissions"
)

// PermissionCallback is called when Claude wants to use a tool.
// Return PermissionAllow to allow, PermissionDeny to deny.
type PermissionCallback func(ctx context.Context, req *PermissionRequest) PermissionResult

// PermissionRequest contains information about a tool permission request.
type PermissionRequest struct {
	// ToolName is the name of the tool being requested.
	ToolName string

	// Input is the raw JSON input for the tool.
	Input json.RawMessage

	// Suggestions contains suggested permission updates.
	Suggestions []PermissionUpdate
}

// PermissionUpdate represents a suggested permission change.
type PermissionUpdate struct {
	Type  string `json:"type"`
	Path  string `json:"path,omitempty"`
	Regex string `json:"regex,omitempty"`
}

// PermissionResult is the result of a permission callback.
type PermissionResult interface {
	isPermissionResult()
}

// PermissionAllow allows the tool to execute.
type PermissionAllow struct {
	// UpdatedInput optionally modifies the tool input.
	UpdatedInput json.RawMessage
}

func (PermissionAllow) isPermissionResult() {}

// PermissionDeny denies the tool execution.
type PermissionDeny struct {
	// Message explains why the tool was denied.
	Message string

	// Interrupt stops the entire query if true.
	Interrupt bool
}

func (PermissionDeny) isPermissionResult() {}

// Allow creates a PermissionAllow result.
func Allow() PermissionResult {
	return PermissionAllow{}
}

// AllowWithInput creates a PermissionAllow with modified input.
func AllowWithInput(input json.RawMessage) PermissionResult {
	return PermissionAllow{UpdatedInput: input}
}

// Deny creates a PermissionDeny result.
func Deny(message string) PermissionResult {
	return PermissionDeny{Message: message}
}

// DenyAndInterrupt creates a PermissionDeny that stops the query.
func DenyAndInterrupt(message string) PermissionResult {
	return PermissionDeny{Message: message, Interrupt: true}
}

// AlwaysAllow returns a callback that always allows tool use.
func AlwaysAllow() PermissionCallback {
	return func(ctx context.Context, req *PermissionRequest) PermissionResult {
		return Allow()
	}
}

// AlwaysDeny returns a callback that always denies tool use.
func AlwaysDeny(message string) PermissionCallback {
	return func(ctx context.Context, req *PermissionRequest) PermissionResult {
		return Deny(message)
	}
}
