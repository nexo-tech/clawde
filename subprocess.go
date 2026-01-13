package clawde

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// SubprocessTransport implements Transport using a subprocess.
type SubprocessTransport struct {
	opts   *Options
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	msgCh  chan json.RawMessage
	errCh  chan error
	doneCh chan struct{}
	mu     sync.Mutex
	closed bool
}

// NewSubprocessTransport creates a new subprocess transport.
func NewSubprocessTransport(opts *Options) *SubprocessTransport {
	return &SubprocessTransport{
		opts:   opts,
		msgCh:  make(chan json.RawMessage, 100),
		errCh:  make(chan error, 10),
		doneCh: make(chan struct{}),
	}
}

// Start initializes the subprocess and begins reading messages.
func (t *SubprocessTransport) Start(ctx context.Context) error {
	cliPath, err := t.findCLI()
	if err != nil {
		return err
	}

	args := t.buildArgs()
	t.cmd = exec.CommandContext(ctx, cliPath, args...)

	// Set working directory
	if t.opts.WorkingDir != "" {
		t.cmd.Dir = t.opts.WorkingDir
	}

	// Set environment
	t.cmd.Env = os.Environ()
	for k, v := range t.opts.Env {
		t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set up pipes
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("clawde: failed to create stdin pipe: %w", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("clawde: failed to create stdout pipe: %w", err)
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("clawde: failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("clawde: failed to start process: %w", err)
	}

	// Start reading goroutines
	go t.readLoop()
	go t.readStderr()
	go t.waitLoop()

	return nil
}

// findCLI locates the Claude CLI executable.
func (t *SubprocessTransport) findCLI() (string, error) {
	// Check explicit path first
	if t.opts.CLIPath != "" {
		if _, err := os.Stat(t.opts.CLIPath); err == nil {
			return t.opts.CLIPath, nil
		}
		return "", fmt.Errorf("%w: %s", ErrCLINotFound, t.opts.CLIPath)
	}

	// Check PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// Check common locations
	commonPaths := []string{
		"/usr/local/bin/claude",
		"/usr/bin/claude",
		filepath.Join(os.Getenv("HOME"), ".local/bin/claude"),
		filepath.Join(os.Getenv("HOME"), ".claude/local/claude"),
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", ErrCLINotFound
}

// buildArgs constructs command line arguments.
func (t *SubprocessTransport) buildArgs() []string {
	args := []string{"--output-format", "stream-json", "--verbose", "--input-format", "stream-json"}

	// System prompt - handle both simple string and config
	if t.opts.SystemPromptConfig != nil {
		if t.opts.SystemPromptConfig.Preset != nil {
			// JSON format for preset
			jsonBytes, _ := json.Marshal(t.opts.SystemPromptConfig.Preset)
			args = append(args, "--system-prompt", string(jsonBytes))
		} else if t.opts.SystemPromptConfig.String != "" {
			args = append(args, "--system-prompt", t.opts.SystemPromptConfig.String)
		}
	} else if t.opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", t.opts.SystemPrompt)
	}

	if t.opts.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", t.opts.AppendSystemPrompt)
	}

	if t.opts.Model != "" {
		args = append(args, "--model", t.opts.Model)
	}

	if t.opts.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", t.opts.MaxTurns))
	}

	if t.opts.MaxBudgetUSD > 0 {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", t.opts.MaxBudgetUSD))
	}

	if t.opts.MaxThinkingTokens > 0 {
		args = append(args, "--max-thinking-tokens", fmt.Sprintf("%d", t.opts.MaxThinkingTokens))
	}

	if len(t.opts.AllowedTools) > 0 {
		args = append(args, "--allowed-tools", strings.Join(t.opts.AllowedTools, ","))
	}

	if len(t.opts.DisallowedTools) > 0 {
		args = append(args, "--disallowed-tools", strings.Join(t.opts.DisallowedTools, ","))
	}

	// Tools option (array or preset)
	if t.opts.Tools != nil {
		if t.opts.Tools.Preset != nil {
			jsonBytes, _ := json.Marshal(t.opts.Tools.Preset)
			args = append(args, "--tools", string(jsonBytes))
		} else {
			// Array of tools (or empty array to disable all tools)
			jsonBytes, _ := json.Marshal(t.opts.Tools.Tools)
			args = append(args, "--tools", string(jsonBytes))
		}
	}

	if t.opts.PermissionMode != "" {
		args = append(args, "--permission-mode", string(t.opts.PermissionMode))
	}

	if t.opts.ResumeConversation != "" {
		args = append(args, "--resume", t.opts.ResumeConversation)
	}

	// Add MCP server configurations
	for name, cfg := range t.opts.MCPServers {
		args = append(args, "--mcp", formatMCPConfig(name, cfg))
	}

	// Add agents
	if len(t.opts.Agents) > 0 {
		agentsJSON, _ := json.Marshal(t.opts.Agents)
		args = append(args, "--agents", string(agentsJSON))
	}

	// Add setting sources - only pass if explicitly configured
	// If not configured, let Claude CLI use its defaults (which include project settings)
	if len(t.opts.SettingSources) > 0 {
		sources := make([]string, len(t.opts.SettingSources))
		for i, s := range t.opts.SettingSources {
			sources[i] = string(s)
		}
		args = append(args, "--setting-sources", strings.Join(sources, ","))
	}

	// Add plugins
	for _, plugin := range t.opts.Plugins {
		pluginJSON, _ := json.Marshal(plugin)
		args = append(args, "--plugin", string(pluginJSON))
	}

	// Include partial messages
	if t.opts.IncludePartialMessages {
		args = append(args, "--include-partial-messages")
	}

	// Extra args
	for key, value := range t.opts.ExtraArgs {
		if value == "" {
			args = append(args, "--"+key)
		} else {
			args = append(args, "--"+key, value)
		}
	}

	return args
}

// formatMCPConfig formats an MCP server configuration.
func formatMCPConfig(name string, cfg MCPServerConfig) string {
	parts := []string{name}

	switch cfg.Type {
	case "stdio":
		cmdParts := []string{cfg.Command}
		cmdParts = append(cmdParts, cfg.Args...)
		parts = append(parts, strings.Join(cmdParts, " "))
	case "sse":
		parts = append(parts, cfg.URL)
	}

	return strings.Join(parts, ":")
}

// readLoop reads JSON messages from stdout using a streaming reader.
// Uses bufio.Reader instead of Scanner for lower latency - ReadSlice returns
// data immediately when a newline is found, rather than waiting for more input.
func (t *SubprocessTransport) readLoop() {
	reader := bufio.NewReaderSize(t.stdout, 64*1024) // 64KB buffer for faster reads
	var accumulator []byte                           // For lines longer than buffer

	for {
		line, err := reader.ReadSlice('\n')
		if err != nil {
			if err == bufio.ErrBufferFull {
				// Line is longer than buffer, accumulate it
				accumulator = append(accumulator, line...)
				continue
			}
			if err == io.EOF {
				// Process any remaining accumulated data
				if len(accumulator) > 0 {
					t.sendLine(accumulator)
				}
				return
			}
			// Other error
			select {
			case t.errCh <- &ParseError{Err: err}:
			case <-t.doneCh:
			}
			return
		}

		// Got a complete line (ends with \n)
		if len(accumulator) > 0 {
			// Had partial data, combine with this line
			accumulator = append(accumulator, line...)
			line = accumulator
			accumulator = nil
		}

		// Trim the newline
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		// Also trim carriage return if present (Windows line endings)
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		if len(line) == 0 {
			continue
		}

		// Send the line immediately
		if !t.sendLine(line) {
			return
		}
	}
}

// sendLine sends a line to the message channel, returning false if done.
func (t *SubprocessTransport) sendLine(line []byte) bool {
	// Make a copy since the buffer may be reused
	msg := make(json.RawMessage, len(line))
	copy(msg, line)

	select {
	case t.msgCh <- msg:
		return true
	case <-t.doneCh:
		return false
	}
}

// readStderr reads error output.
func (t *SubprocessTransport) readStderr() {
	scanner := bufio.NewScanner(t.stderr)
	var stderr strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Call stderr callback if provided
		if t.opts.StderrCallback != nil {
			t.opts.StderrCallback(line)
		}

		stderr.WriteString(line)
		stderr.WriteString("\n")
	}

	if stderr.Len() > 0 {
		select {
		case t.errCh <- &ProcessError{Stderr: stderr.String()}:
		case <-t.doneCh:
		}
	}
}

// waitLoop waits for the process to exit.
func (t *SubprocessTransport) waitLoop() {
	err := t.cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			select {
			case t.errCh <- &ProcessError{ExitCode: exitErr.ExitCode()}:
			case <-t.doneCh:
			}
		}
	}
	close(t.msgCh)
}

// Write sends data to the subprocess.
func (t *SubprocessTransport) Write(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrStreamClosed
	}

	// Append newline if not present
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	_, err := t.stdin.Write(data)
	return err
}

// Messages returns the message channel.
func (t *SubprocessTransport) Messages() <-chan json.RawMessage {
	return t.msgCh
}

// Errors returns the error channel.
func (t *SubprocessTransport) Errors() <-chan error {
	return t.errCh
}

// Close shuts down the transport.
func (t *SubprocessTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}
	t.closed = true

	close(t.doneCh)

	if t.stdin != nil {
		t.stdin.Close()
	}

	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}

	return nil
}
