package clawde

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Common errors returned by the SDK.
var (
	// ErrCLINotFound is returned when the Claude CLI cannot be located.
	ErrCLINotFound = errors.New("clawde: claude CLI not found")

	// ErrNotConnected is returned when operations are attempted before connecting.
	ErrNotConnected = errors.New("clawde: client not connected")

	// ErrAlreadyConnected is returned when Connect is called on a connected client.
	ErrAlreadyConnected = errors.New("clawde: client already connected")

	// ErrStreamClosed is returned when reading from a closed stream.
	ErrStreamClosed = errors.New("clawde: stream closed")

	// ErrBudgetExceeded is returned when the query exceeds the budget limit.
	ErrBudgetExceeded = errors.New("clawde: budget exceeded")

	// ErrMaxTurnsExceeded is returned when the query exceeds the turn limit.
	ErrMaxTurnsExceeded = errors.New("clawde: max turns exceeded")

	// ErrTimeout is returned when a query times out.
	ErrTimeout = errors.New("clawde: timeout")

	// ErrInterrupted is returned when a query is interrupted.
	ErrInterrupted = errors.New("clawde: interrupted")
)

// ProcessError represents an error from the subprocess.
type ProcessError struct {
	ExitCode int
	Stderr   string
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("clawde: process exited with code %d: %s", e.ExitCode, e.Stderr)
}

// ProtocolError represents an error in the communication protocol.
type ProtocolError struct {
	Message string
	Data    json.RawMessage
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("clawde: protocol error: %s", e.Message)
}

// ParseError represents an error parsing a message.
type ParseError struct {
	Line string
	Err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("clawde: parse error: %v (line: %s)", e.Err, truncate(e.Line, 100))
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// ToolError represents an error from a tool execution.
type ToolError struct {
	ToolName string
	Message  string
}

func (e *ToolError) Error() string {
	return fmt.Sprintf("clawde: tool %q error: %s", e.ToolName, e.Message)
}

// truncate truncates a string to the specified length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
