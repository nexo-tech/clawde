package clawde

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrCLINotFound is returned when the Claude CLI cannot be found.
	ErrCLINotFound = errors.New("clawde: claude CLI not found")

	// ErrNotConnected is returned when operations are attempted on a disconnected client.
	ErrNotConnected = errors.New("clawde: client not connected")

	// ErrAlreadyConnected is returned when Connect is called on an already connected client.
	ErrAlreadyConnected = errors.New("clawde: client already connected")

	// ErrStreamClosed is returned when operations are attempted on a closed stream.
	ErrStreamClosed = errors.New("clawde: stream closed")

	// ErrBudgetExceeded is returned when the budget limit has been exceeded.
	ErrBudgetExceeded = errors.New("clawde: budget exceeded")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("clawde: operation timed out")

	// ErrPermissionDenied is returned when a permission is denied.
	ErrPermissionDenied = errors.New("clawde: permission denied")

	// ErrInterrupted is returned when an operation is interrupted.
	ErrInterrupted = errors.New("clawde: operation interrupted")
)

// ProcessError is returned when the CLI subprocess exits with an error.
type ProcessError struct {
	ExitCode int
	Stderr   string
	Message  string
}

func (e *ProcessError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("clawde: process error: %s (exit code %d)", e.Message, e.ExitCode)
	}
	if e.Stderr != "" {
		return fmt.Sprintf("clawde: process exited with code %d: %s", e.ExitCode, e.Stderr)
	}
	return fmt.Sprintf("clawde: process exited with code %d", e.ExitCode)
}

// Is implements errors.Is.
func (e *ProcessError) Is(target error) bool {
	_, ok := target.(*ProcessError)
	return ok
}

// ProtocolError is returned when there's a protocol-level error.
type ProtocolError struct {
	Message   string
	RequestID string
	Data      json.RawMessage
}

func (e *ProtocolError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("clawde: protocol error [%s]: %s", e.RequestID, e.Message)
	}
	return fmt.Sprintf("clawde: protocol error: %s", e.Message)
}

// Is implements errors.Is.
func (e *ProtocolError) Is(target error) bool {
	_, ok := target.(*ProtocolError)
	return ok
}

// ParseError is returned when message parsing fails.
type ParseError struct {
	Message string
	Line    string
	Err     error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("clawde: parse error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("clawde: parse error: %s", e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is.
func (e *ParseError) Is(target error) bool {
	_, ok := target.(*ParseError)
	return ok
}

// TransportError is returned when there's a transport-level error.
type TransportError struct {
	Message string
	Err     error
}

func (e *TransportError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("clawde: transport error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("clawde: transport error: %s", e.Message)
}

func (e *TransportError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is.
func (e *TransportError) Is(target error) bool {
	_, ok := target.(*TransportError)
	return ok
}

// HookError is returned when a hook fails.
type HookError struct {
	Event   HookEvent
	Message string
	Err     error
}

func (e *HookError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("clawde: hook error [%s]: %s: %v", e.Event, e.Message, e.Err)
	}
	return fmt.Sprintf("clawde: hook error [%s]: %s", e.Event, e.Message)
}

func (e *HookError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is.
func (e *HookError) Is(target error) bool {
	_, ok := target.(*HookError)
	return ok
}

// IsCLINotFound returns true if the error indicates the CLI was not found.
func IsCLINotFound(err error) bool {
	return errors.Is(err, ErrCLINotFound)
}

// IsNotConnected returns true if the error indicates the client is not connected.
func IsNotConnected(err error) bool {
	return errors.Is(err, ErrNotConnected)
}

// IsBudgetExceeded returns true if the error indicates the budget was exceeded.
func IsBudgetExceeded(err error) bool {
	return errors.Is(err, ErrBudgetExceeded)
}

// IsTimeout returns true if the error indicates a timeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsPermissionDenied returns true if the error indicates permission was denied.
func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

// IsProcessError returns true if the error is a ProcessError.
func IsProcessError(err error) bool {
	var pe *ProcessError
	return errors.As(err, &pe)
}

// IsProtocolError returns true if the error is a ProtocolError.
func IsProtocolError(err error) bool {
	var pe *ProtocolError
	return errors.As(err, &pe)
}

// IsParseError returns true if the error is a ParseError.
func IsParseError(err error) bool {
	var pe *ParseError
	return errors.As(err, &pe)
}
