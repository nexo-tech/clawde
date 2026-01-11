package clawde

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SubagentTracker tracks tool calls across subagents for logging and debugging.
type SubagentTracker struct {
	// TranscriptWriter receives human-readable transcript output
	TranscriptWriter io.Writer

	// SessionDir is the directory where logs are saved
	SessionDir string

	mu            sync.Mutex
	toolCallsFile *os.File
	activeAgents  map[string]*AgentInfo
	toolCalls     map[string]*ToolCallInfo
}

// AgentInfo contains information about an active subagent
type AgentInfo struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	StartTime time.Time `json:"start_time"`
	ToolUseID string    `json:"tool_use_id"` // Parent tool use ID
}

// ToolCallInfo contains information about a tool call
type ToolCallInfo struct {
	ID              string          `json:"id"`
	ToolName        string          `json:"tool_name"`
	Input           json.RawMessage `json:"input,omitempty"`
	Output          json.RawMessage `json:"output,omitempty"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time,omitempty"`
	DurationMS      int64           `json:"duration_ms,omitempty"`
	ParentToolUseID string          `json:"parent_tool_use_id,omitempty"`
	SubagentID      string          `json:"subagent_id,omitempty"`
	Success         bool            `json:"success"`
	Error           string          `json:"error,omitempty"`
}

// NewSubagentTracker creates a new tracker with the given session directory.
func NewSubagentTracker(sessionDir string) (*SubagentTracker, error) {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}

	toolCallsPath := filepath.Join(sessionDir, "tool_calls.jsonl")
	f, err := os.OpenFile(toolCallsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open tool calls file: %w", err)
	}

	return &SubagentTracker{
		SessionDir:       sessionDir,
		TranscriptWriter: os.Stdout,
		toolCallsFile:    f,
		activeAgents:     make(map[string]*AgentInfo),
		toolCalls:        make(map[string]*ToolCallInfo),
	}, nil
}

// PreToolUseHook is a hook that tracks tool calls before they execute.
func (t *SubagentTracker) PreToolUseHook(ctx context.Context, input *HookInput) (*HookOutput, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Generate ID if not provided
	toolUseID := input.ToolUseID
	if toolUseID == "" {
		toolUseID = fmt.Sprintf("tool_%d", time.Now().UnixNano())
	}

	toolCall := &ToolCallInfo{
		ID:        toolUseID,
		ToolName:  input.ToolName,
		StartTime: time.Now(),
	}

	// Store tool input
	if input.ToolInput != nil {
		toolCall.Input = input.ToolInput
	}

	// Track parent tool use ID for subagent association
	if input.ParentToolUseID != nil {
		toolCall.ParentToolUseID = *input.ParentToolUseID

		// Check if this is a subagent tool call
		for _, agent := range t.activeAgents {
			if agent.ToolUseID == *input.ParentToolUseID {
				toolCall.SubagentID = agent.ID
				break
			}
		}
	}

	// Check if this is a Task tool (spawning a subagent)
	if input.ToolName == "Task" && input.ToolInputMap != nil {
		if subagentType, ok := input.ToolInputMap["subagent_type"].(string); ok {
			agentInfo := &AgentInfo{
				ID:        toolUseID,
				Type:      subagentType,
				StartTime: time.Now(),
				ToolUseID: toolUseID,
			}
			t.activeAgents[toolUseID] = agentInfo

			shortID := toolUseID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}
			if t.TranscriptWriter != nil {
				fmt.Fprintf(t.TranscriptWriter, "\n[Subagent] Starting %s (%s)\n", subagentType, shortID)
			}
		}
	}

	t.toolCalls[toolUseID] = toolCall

	if t.TranscriptWriter != nil {
		fmt.Fprintf(t.TranscriptWriter, "[Tool] %s\n", input.ToolName)
	}

	return ContinueHook(), nil
}

// PostToolUseHook is a hook that tracks tool calls after they complete.
func (t *SubagentTracker) PostToolUseHook(ctx context.Context, input *HookInput) (*HookOutput, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	toolUseID := input.ToolUseID
	if toolUseID == "" {
		return ContinueHook(), nil
	}

	toolCall, ok := t.toolCalls[toolUseID]
	if !ok {
		return ContinueHook(), nil
	}

	toolCall.EndTime = time.Now()
	toolCall.DurationMS = toolCall.EndTime.Sub(toolCall.StartTime).Milliseconds()
	toolCall.Success = true

	// Check for error in result
	if input.ToolResult != nil {
		if errMsg, ok := input.ToolResult["error"].(string); ok && errMsg != "" {
			toolCall.Success = false
			toolCall.Error = errMsg
		}

		// Store truncated output
		outputJSON, _ := json.Marshal(input.ToolResult)
		if len(outputJSON) > 1000 {
			outputJSON = append(outputJSON[:1000], []byte("...")...)
		}
		toolCall.Output = outputJSON
	}

	// Remove from active agents if this was a Task tool
	if input.ToolName == "Task" {
		if agent, ok := t.activeAgents[toolUseID]; ok {
			if t.TranscriptWriter != nil {
				fmt.Fprintf(t.TranscriptWriter, "[Subagent] Completed %s (%dms)\n", agent.Type, toolCall.DurationMS)
			}
			delete(t.activeAgents, toolUseID)
		}
	}

	// Write to JSONL file
	if t.toolCallsFile != nil {
		line, _ := json.Marshal(toolCall)
		t.toolCallsFile.Write(append(line, '\n'))
	}

	delete(t.toolCalls, toolUseID)

	return ContinueHook(), nil
}

// PostToolUseFailureHook handles tool failures.
func (t *SubagentTracker) PostToolUseFailureHook(ctx context.Context, input *HookInput) (*HookOutput, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	toolUseID := input.ToolUseID
	if toolUseID == "" {
		return ContinueHook(), nil
	}

	toolCall, ok := t.toolCalls[toolUseID]
	if !ok {
		return ContinueHook(), nil
	}

	toolCall.EndTime = time.Now()
	toolCall.DurationMS = toolCall.EndTime.Sub(toolCall.StartTime).Milliseconds()
	toolCall.Success = false

	if input.ToolResult != nil {
		if errMsg, ok := input.ToolResult["error"].(string); ok {
			toolCall.Error = errMsg
		}
	}

	if t.TranscriptWriter != nil {
		fmt.Fprintf(t.TranscriptWriter, "[Tool Error] %s: %s\n", input.ToolName, toolCall.Error)
	}

	// Write to JSONL file
	if t.toolCallsFile != nil {
		line, _ := json.Marshal(toolCall)
		t.toolCallsFile.Write(append(line, '\n'))
	}

	delete(t.toolCalls, toolUseID)

	return ContinueHook(), nil
}

// Close closes the tracker and flushes any pending writes.
func (t *SubagentTracker) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.toolCallsFile != nil {
		return t.toolCallsFile.Close()
	}
	return nil
}

// GetActiveAgents returns a list of currently active subagents.
func (t *SubagentTracker) GetActiveAgents() []*AgentInfo {
	t.mu.Lock()
	defer t.mu.Unlock()

	agents := make([]*AgentInfo, 0, len(t.activeAgents))
	for _, agent := range t.activeAgents {
		agents = append(agents, agent)
	}
	return agents
}

// TranscriptWriter writes human-readable transcripts to a file.
type TranscriptWriter struct {
	file *os.File
	mu   sync.Mutex
}

// NewTranscriptWriter creates a new transcript writer.
func NewTranscriptWriter(path string) (*TranscriptWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &TranscriptWriter{file: f}, nil
}

// Write writes to the transcript file.
func (w *TranscriptWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Write(p)
}

// WriteString writes a string to the transcript.
func (w *TranscriptWriter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

// Close closes the transcript file.
func (w *TranscriptWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
