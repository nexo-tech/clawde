package clawde

import (
	"context"
	"encoding/json"
	"time"
)

// HookEvent represents the type of hook event.
type HookEvent string

const (
	// HookPreToolUse is called before a tool is executed.
	HookPreToolUse HookEvent = "PreToolUse"

	// HookPostToolUse is called after a tool is executed.
	HookPostToolUse HookEvent = "PostToolUse"

	// HookUserPromptSubmit is called when a user prompt is submitted.
	HookUserPromptSubmit HookEvent = "UserPromptSubmit"

	// HookStop is called when the agent stops.
	HookStop HookEvent = "Stop"

	// HookSubagentStop is called when a subagent stops.
	HookSubagentStop HookEvent = "SubagentStop"

	// HookPreCompact is called before context compaction.
	HookPreCompact HookEvent = "PreCompact"
)

// HookMatcher defines which tools a hook applies to and its callback.
type HookMatcher struct {
	// ToolName is the tool to match ("*" for all tools).
	ToolName string

	// Callback is the function to call when the hook matches.
	Callback HookCallback

	// Timeout is the maximum duration for the hook callback.
	Timeout time.Duration
}

// HookCallback is called when a hook event occurs.
type HookCallback func(ctx context.Context, input *HookInput) (*HookOutput, error)

// HookInput contains information about the hook event.
type HookInput struct {
	// SessionID is the current session ID.
	SessionID string `json:"session_id"`

	// ToolName is the name of the tool (for tool hooks).
	ToolName string `json:"tool_name"`

	// ToolUseID is the unique ID of the tool use.
	ToolUseID string `json:"tool_use_id"`

	// ParentToolUseID is the ID of the parent tool use (for subagent calls).
	ParentToolUseID *string `json:"parent_tool_use_id,omitempty"`

	// ToolInput is the raw JSON input for the tool.
	ToolInput json.RawMessage `json:"tool_input"`

	// ToolInputMap is the parsed tool input as a map for convenience.
	ToolInputMap map[string]interface{} `json:"-"`

	// ToolOutput is the raw JSON output from the tool (PostToolUse only).
	// Note: CLI sends this as "tool_response" in the JSON.
	ToolOutput json.RawMessage `json:"tool_response"`

	// ToolResult is the parsed tool result as a map for convenience.
	ToolResult map[string]interface{} `json:"-"`

	// Prompt is the user prompt (UserPromptSubmit only).
	Prompt string `json:"prompt,omitempty"`

	// StopReason is the reason for stopping (Stop/SubagentStop only).
	StopReason string `json:"stop_reason,omitempty"`
}

// Command extracts the command from a Bash tool input.
func (h *HookInput) Command() string {
	if h.ToolName != "Bash" {
		return ""
	}
	var input struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(h.ToolInput, &input); err != nil {
		return ""
	}
	return input.Command
}

// FilePath extracts the file path from tool inputs (Read, Write, Edit).
func (h *HookInput) FilePath() string {
	var input struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	}
	if err := json.Unmarshal(h.ToolInput, &input); err != nil {
		return ""
	}
	if input.FilePath != "" {
		return input.FilePath
	}
	return input.Path
}

// HookOutput is the result of a hook callback.
type HookOutput struct {
	// Continue indicates whether to continue execution.
	Continue bool

	// Decision is "block" to block the action.
	Decision string

	// Reason explains why the action was blocked.
	Reason string

	// StopReason is the reason for stopping (Stop hooks only).
	StopReason string

	// ModifiedInput allows modifying the tool input.
	ModifiedInput json.RawMessage
}

// ContinueHook returns a HookOutput that allows execution to continue.
func ContinueHook() *HookOutput {
	return &HookOutput{Continue: true}
}

// BlockHook returns a HookOutput that blocks the action.
func BlockHook(reason string) *HookOutput {
	return &HookOutput{
		Continue: false,
		Decision: "block",
		Reason:   reason,
	}
}

// ModifyHook returns a HookOutput that modifies the tool input.
func ModifyHook(input json.RawMessage) *HookOutput {
	return &HookOutput{
		Continue:      true,
		ModifiedInput: input,
	}
}

// StopHook returns a HookOutput that stops the agent.
func StopHook(reason string) *HookOutput {
	return &HookOutput{
		Continue:   false,
		StopReason: reason,
	}
}

// MatchAll creates a HookMatcher that matches all tools.
func MatchAll(callback HookCallback) HookMatcher {
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
