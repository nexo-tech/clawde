package clawde

import (
	"context"
	"encoding/json"
	"time"
)

// HookEvent represents different hook trigger points.
type HookEvent string

const (
	// HookPreToolUse is called before a tool is executed.
	HookPreToolUse HookEvent = "PreToolUse"

	// HookPostToolUse is called after a tool execution completes.
	HookPostToolUse HookEvent = "PostToolUse"

	// HookUserPromptSubmit is called when a user prompt is submitted.
	HookUserPromptSubmit HookEvent = "UserPromptSubmit"

	// HookStop is called when the session is stopping.
	HookStop HookEvent = "Stop"

	// HookSubagentStop is called when a subagent stops.
	HookSubagentStop HookEvent = "SubagentStop"

	// HookPreCompact is called before conversation compaction.
	HookPreCompact HookEvent = "PreCompact"
)

// HookMatcher matches and handles hook events.
type HookMatcher struct {
	// ToolName is the tool to match, or "*" for all tools.
	// Only applicable for PreToolUse and PostToolUse events.
	ToolName string

	// Callback is the function called when the hook matches.
	Callback HookCallback

	// Timeout is the maximum time to wait for the callback.
	Timeout time.Duration
}

// HookCallback is the signature for hook handler functions.
type HookCallback func(ctx context.Context, input *HookInput) (*HookOutput, error)

// HookInput contains input data for hook callbacks.
type HookInput struct {
	// Event is the hook event type.
	Event HookEvent `json:"hook_event_name"`

	// SessionID is the current session ID.
	SessionID string `json:"session_id"`

	// TranscriptPath is the path to the transcript file.
	TranscriptPath string `json:"transcript_path"`

	// WorkingDir is the current working directory.
	WorkingDir string `json:"cwd"`

	// PermissionMode is the current permission mode.
	PermissionMode string `json:"permission_mode,omitempty"`

	// ToolName is the tool name (for PreToolUse/PostToolUse).
	ToolName string `json:"tool_name,omitempty"`

	// ToolInput is the tool input (for PreToolUse/PostToolUse).
	ToolInput json.RawMessage `json:"tool_input,omitempty"`

	// ToolOutput is the tool output (for PostToolUse).
	ToolOutput json.RawMessage `json:"tool_response,omitempty"`

	// ToolUseID is the unique ID for this tool use.
	ToolUseID string `json:"tool_use_id,omitempty"`

	// Prompt is the user prompt (for UserPromptSubmit).
	Prompt string `json:"prompt,omitempty"`

	// StopHookActive indicates if a stop hook is active.
	StopHookActive bool `json:"stop_hook_active,omitempty"`

	// Trigger is the compaction trigger (for PreCompact).
	Trigger string `json:"trigger,omitempty"` // "manual" or "auto"

	// CustomInstructions are custom compaction instructions.
	CustomInstructions string `json:"custom_instructions,omitempty"`
}

// ToolInputAs unmarshals the tool input into the provided type.
func (h *HookInput) ToolInputAs(v any) error {
	return json.Unmarshal(h.ToolInput, v)
}

// ToolOutputAs unmarshals the tool output into the provided type.
func (h *HookInput) ToolOutputAs(v any) error {
	return json.Unmarshal(h.ToolOutput, v)
}

// HookOutput contains the result of a hook callback.
type HookOutput struct {
	// Continue indicates whether to continue execution.
	// If false, the tool execution is blocked.
	Continue bool `json:"continue,omitempty"`

	// Decision is the hook decision ("block" to block).
	Decision string `json:"decision,omitempty"`

	// Reason explains the decision.
	Reason string `json:"reason,omitempty"`

	// StopReason is the reason for stopping.
	StopReason string `json:"stopReason,omitempty"`

	// SuppressOutput suppresses the tool output.
	SuppressOutput bool `json:"suppressOutput,omitempty"`

	// SystemMessage adds a system message.
	SystemMessage string `json:"systemMessage,omitempty"`

	// Async indicates async hook processing.
	Async bool `json:"async,omitempty"`

	// AsyncTimeout is the timeout for async processing.
	AsyncTimeout int `json:"asyncTimeout,omitempty"`
}

// ToJSON converts the output to JSON for the control protocol.
func (h *HookOutput) ToJSON() map[string]any {
	if h.Async {
		result := map[string]any{
			"async": true,
		}
		if h.AsyncTimeout > 0 {
			result["asyncTimeout"] = h.AsyncTimeout
		}
		return result
	}

	result := make(map[string]any)
	if !h.Continue {
		result["continue"] = false
	}
	if h.Decision != "" {
		result["decision"] = h.Decision
	}
	if h.Reason != "" {
		result["reason"] = h.Reason
	}
	if h.StopReason != "" {
		result["stopReason"] = h.StopReason
	}
	if h.SuppressOutput {
		result["suppressOutput"] = true
	}
	if h.SystemMessage != "" {
		result["systemMessage"] = h.SystemMessage
	}
	return result
}

// ContinueHook returns a hook output that allows execution to continue.
func ContinueHook() *HookOutput {
	return &HookOutput{Continue: true}
}

// BlockHook returns a hook output that blocks execution.
func BlockHook(reason string) *HookOutput {
	return &HookOutput{
		Continue: false,
		Decision: "block",
		Reason:   reason,
	}
}

// StopHook returns a hook output that stops the session.
func StopHook(reason string) *HookOutput {
	return &HookOutput{
		Continue:   false,
		StopReason: reason,
	}
}

// AsyncHook returns a hook output for async processing.
func AsyncHook(timeout time.Duration) *HookOutput {
	return &HookOutput{
		Async:        true,
		AsyncTimeout: int(timeout.Milliseconds()),
	}
}

// HookWithMessage returns a hook output with a system message.
func HookWithMessage(message string) *HookOutput {
	return &HookOutput{
		Continue:      true,
		SystemMessage: message,
	}
}

// MatchAllTools creates a HookMatcher that matches all tools.
func MatchAllTools(callback HookCallback) HookMatcher {
	return HookMatcher{
		ToolName: "*",
		Callback: callback,
	}
}

// MatchTool creates a HookMatcher for a specific tool.
func MatchTool(toolName string, callback HookCallback) HookMatcher {
	return HookMatcher{
		ToolName: toolName,
		Callback: callback,
	}
}

// MatchToolWithTimeout creates a HookMatcher with a timeout.
func MatchToolWithTimeout(toolName string, timeout time.Duration, callback HookCallback) HookMatcher {
	return HookMatcher{
		ToolName: toolName,
		Callback: callback,
		Timeout:  timeout,
	}
}

// matchesHook checks if a hook matcher matches the input.
func (m *HookMatcher) matchesHook(input *HookInput) bool {
	if m.ToolName == "*" || m.ToolName == "" {
		return true
	}
	return m.ToolName == input.ToolName
}
