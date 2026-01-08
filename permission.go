package clawde

import (
	"context"
	"encoding/json"
)

// PermissionMode controls how permissions are handled.
type PermissionMode string

const (
	// PermissionDefault uses Claude Code's default permission handling.
	PermissionDefault PermissionMode = "default"

	// PermissionAcceptEdits automatically accepts file edits.
	PermissionAcceptEdits PermissionMode = "acceptEdits"

	// PermissionPlan enables plan mode for reviewing changes.
	PermissionPlan PermissionMode = "plan"

	// PermissionBypassAll bypasses all permission checks.
	PermissionBypassAll PermissionMode = "bypassPermissions"
)

// PermissionCallback is called when a tool requests permission.
type PermissionCallback func(ctx context.Context, req *PermissionRequest) PermissionResult

// PermissionRequest contains information about a permission request.
type PermissionRequest struct {
	// ToolName is the name of the tool requesting permission.
	ToolName string `json:"tool_name"`

	// Input is the tool's input parameters.
	Input json.RawMessage `json:"input"`

	// Suggestions contains suggested permission updates.
	Suggestions []PermissionUpdate `json:"permission_suggestions,omitempty"`

	// BlockedPath is the path that was blocked, if applicable.
	BlockedPath string `json:"blocked_path,omitempty"`
}

// InputAs unmarshals the input into the provided type.
func (r *PermissionRequest) InputAs(v any) error {
	return json.Unmarshal(r.Input, v)
}

// PermissionResult is the result of a permission decision.
type PermissionResult interface {
	isPermissionResult()
	// ToJSON converts the result to JSON for the control protocol.
	ToJSON() map[string]any
}

// PermissionAllow allows the tool to proceed.
type PermissionAllow struct {
	// UpdatedInput optionally modifies the tool input.
	UpdatedInput json.RawMessage

	// UpdatedPermissions optionally updates permissions.
	UpdatedPermissions []PermissionUpdate
}

func (*PermissionAllow) isPermissionResult() {}

func (p *PermissionAllow) ToJSON() map[string]any {
	result := map[string]any{
		"behavior": "allow",
	}
	if p.UpdatedInput != nil {
		result["updatedInput"] = p.UpdatedInput
	}
	if len(p.UpdatedPermissions) > 0 {
		updates := make([]map[string]any, len(p.UpdatedPermissions))
		for i, u := range p.UpdatedPermissions {
			updates[i] = u.ToJSON()
		}
		result["updatedPermissions"] = updates
	}
	return result
}

// PermissionDeny denies the tool from proceeding.
type PermissionDeny struct {
	// Message explains why permission was denied.
	Message string

	// Interrupt stops the entire conversation if true.
	Interrupt bool
}

func (*PermissionDeny) isPermissionResult() {}

func (p *PermissionDeny) ToJSON() map[string]any {
	return map[string]any{
		"behavior":  "deny",
		"message":   p.Message,
		"interrupt": p.Interrupt,
	}
}

// Allow returns a PermissionAllow result.
func Allow() *PermissionAllow {
	return &PermissionAllow{}
}

// AllowWithInput returns a PermissionAllow result with modified input.
func AllowWithInput(input json.RawMessage) *PermissionAllow {
	return &PermissionAllow{UpdatedInput: input}
}

// Deny returns a PermissionDeny result.
func Deny(message string) *PermissionDeny {
	return &PermissionDeny{Message: message}
}

// DenyAndInterrupt returns a PermissionDeny result that interrupts the conversation.
func DenyAndInterrupt(message string) *PermissionDeny {
	return &PermissionDeny{Message: message, Interrupt: true}
}

// PermissionUpdate describes a permission update.
type PermissionUpdate struct {
	// Type is the update type.
	Type PermissionUpdateType `json:"type"`

	// Rules contains permission rules.
	Rules []PermissionRule `json:"rules,omitempty"`

	// Behavior is the permission behavior.
	Behavior PermissionBehavior `json:"behavior,omitempty"`

	// Mode is the permission mode.
	Mode PermissionMode `json:"mode,omitempty"`

	// Directories contains directory paths.
	Directories []string `json:"directories,omitempty"`

	// Destination specifies where to store the update.
	Destination PermissionDestination `json:"destination,omitempty"`
}

func (u *PermissionUpdate) ToJSON() map[string]any {
	result := map[string]any{
		"type": string(u.Type),
	}
	if len(u.Rules) > 0 {
		rules := make([]map[string]any, len(u.Rules))
		for i, r := range u.Rules {
			rules[i] = map[string]any{
				"tool_name":    r.ToolName,
				"rule_content": r.RuleContent,
			}
		}
		result["rules"] = rules
	}
	if u.Behavior != "" {
		result["behavior"] = string(u.Behavior)
	}
	if u.Mode != "" {
		result["mode"] = string(u.Mode)
	}
	if len(u.Directories) > 0 {
		result["directories"] = u.Directories
	}
	if u.Destination != "" {
		result["destination"] = string(u.Destination)
	}
	return result
}

// PermissionUpdateType is the type of permission update.
type PermissionUpdateType string

const (
	PermissionAddRules       PermissionUpdateType = "addRules"
	PermissionReplaceRules   PermissionUpdateType = "replaceRules"
	PermissionRemoveRules    PermissionUpdateType = "removeRules"
	PermissionSetMode        PermissionUpdateType = "setMode"
	PermissionAddDirectories PermissionUpdateType = "addDirectories"
	PermissionRemoveDirectories PermissionUpdateType = "removeDirectories"
)

// PermissionRule defines a permission rule.
type PermissionRule struct {
	ToolName    string `json:"tool_name"`
	RuleContent string `json:"rule_content,omitempty"`
}

// PermissionBehavior is the behavior for a permission.
type PermissionBehavior string

const (
	BehaviorAllow PermissionBehavior = "allow"
	BehaviorDeny  PermissionBehavior = "deny"
	BehaviorAsk   PermissionBehavior = "ask"
)

// PermissionDestination is where to store permission updates.
type PermissionDestination string

const (
	DestinationUserSettings    PermissionDestination = "userSettings"
	DestinationProjectSettings PermissionDestination = "projectSettings"
	DestinationLocalSettings   PermissionDestination = "localSettings"
	DestinationSession         PermissionDestination = "session"
)

// PermissionContext provides context for permission decisions.
type PermissionContext struct {
	// SessionID is the current session ID.
	SessionID string

	// TranscriptPath is the path to the transcript file.
	TranscriptPath string

	// WorkingDir is the current working directory.
	WorkingDir string

	// PermissionMode is the current permission mode.
	PermissionMode string
}
