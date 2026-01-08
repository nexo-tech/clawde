package clawde

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// SubprocessTransport implements Transport using Claude Code CLI subprocess.
type SubprocessTransport struct {
	opts *Options

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	msgCh  chan json.RawMessage
	errCh  chan error
	doneCh chan struct{}

	mu      sync.Mutex
	started bool
	closed  bool
}

// NewSubprocessTransport creates a new subprocess-based transport.
func NewSubprocessTransport(opts *Options) *SubprocessTransport {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &SubprocessTransport{
		opts:   opts,
		msgCh:  make(chan json.RawMessage, 100),
		errCh:  make(chan error, 10),
		doneCh: make(chan struct{}),
	}
}

// Start initializes and starts the subprocess.
func (t *SubprocessTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.started {
		return ErrAlreadyConnected
	}

	// Find CLI path
	cliPath := t.opts.CLIPath
	if cliPath == "" {
		cliPath = findCLI()
	}
	if cliPath == "" {
		return ErrCLINotFound
	}

	// Build command arguments
	args := t.buildArgs()

	// Create command
	t.cmd = exec.CommandContext(ctx, cliPath, args...)

	// Set working directory
	if t.opts.WorkingDir != "" {
		t.cmd.Dir = t.opts.WorkingDir
	}

	// Set environment
	t.cmd.Env = os.Environ()
	for k, v := range t.opts.Env {
		t.cmd.Env = append(t.cmd.Env, k+"="+v)
	}

	// Get pipes
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return &TransportError{Message: "failed to get stdin pipe", Err: err}
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return &TransportError{Message: "failed to get stdout pipe", Err: err}
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return &TransportError{Message: "failed to get stderr pipe", Err: err}
	}

	// Start process
	if err := t.cmd.Start(); err != nil {
		return &TransportError{Message: "failed to start subprocess", Err: err}
	}

	t.started = true

	// Start reader goroutines
	go t.readLoop()
	go t.readStderr()
	go t.waitLoop()

	return nil
}

// buildArgs builds the CLI command arguments.
func (t *SubprocessTransport) buildArgs() []string {
	args := []string{
		"--output-format", "stream-json",
		"--verbose",
	}

	// Add options
	if t.opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", t.opts.SystemPrompt)
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
		for _, tool := range t.opts.AllowedTools {
			args = append(args, "--allowedTools", tool)
		}
	}

	if len(t.opts.DisallowedTools) > 0 {
		for _, tool := range t.opts.DisallowedTools {
			args = append(args, "--disallowedTools", tool)
		}
	}

	if t.opts.PermissionMode != "" {
		args = append(args, "--permission-mode", string(t.opts.PermissionMode))
	}

	if t.opts.ContinueSession {
		args = append(args, "--continue")
	}

	if t.opts.ResumeSessionID != "" {
		args = append(args, "--resume", t.opts.ResumeSessionID)
	}

	if t.opts.IncludePartialMessages {
		args = append(args, "--include-partial-messages")
	}

	if t.opts.ForkSession {
		args = append(args, "--fork-session")
	}

	// Add MCP servers config
	if len(t.opts.MCPServers) > 0 {
		mcpConfig := t.buildMCPConfig()
		if mcpConfig != "" {
			args = append(args, "--mcp-config", mcpConfig)
		}
	}

	// Add agents config
	if len(t.opts.Agents) > 0 {
		agentsJSON, _ := json.Marshal(t.opts.Agents)
		args = append(args, "--agents", string(agentsJSON))
	}

	// Output format schema
	if len(t.opts.OutputFormat) > 0 {
		args = append(args, "--output-format-schema", string(t.opts.OutputFormat))
	}

	return args
}

// buildMCPConfig builds the MCP configuration JSON.
func (t *SubprocessTransport) buildMCPConfig() string {
	servers := make(map[string]any)
	hasSDK := false

	for name, cfg := range t.opts.MCPServers {
		if cfg.Type == MCPServerTypeSDK {
			hasSDK = true
			// SDK servers are handled internally, mark for control protocol
			servers[name] = map[string]any{
				"type": "sdk",
			}
		} else {
			servers[name] = cfg.ToJSON()
		}
	}

	if len(servers) == 0 && !hasSDK {
		return ""
	}

	config := map[string]any{
		"mcpServers": servers,
	}
	data, _ := json.Marshal(config)
	return string(data)
}

// readLoop reads messages from stdout.
func (t *SubprocessTransport) readLoop() {
	scanner := bufio.NewScanner(t.stdout)
	// Handle potentially large JSON messages (up to 10MB)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Make a copy of the data
		data := make([]byte, len(line))
		copy(data, line)

		select {
		case t.msgCh <- json.RawMessage(data):
		case <-t.doneCh:
			return
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errCh <- &TransportError{Message: "read error", Err: err}:
		case <-t.doneCh:
		}
	}
}

// readStderr reads and discards stderr (or logs it).
func (t *SubprocessTransport) readStderr() {
	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		// Stderr is typically for debugging
		// You could log this if needed
		_ = scanner.Text()
	}
}

// waitLoop waits for the process to exit.
func (t *SubprocessTransport) waitLoop() {
	err := t.cmd.Wait()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			select {
			case t.errCh <- &ProcessError{
				ExitCode: exitErr.ExitCode(),
				Message:  "process exited",
			}:
			default:
			}
		} else {
			select {
			case t.errCh <- &TransportError{Message: "process error", Err: err}:
			default:
			}
		}
	}

	t.mu.Lock()
	if !t.closed {
		close(t.doneCh)
		t.closed = true
	}
	t.mu.Unlock()
}

// Write sends data to the subprocess stdin.
func (t *SubprocessTransport) Write(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrStreamClosed
	}

	if !t.started {
		return ErrNotConnected
	}

	// Append newline for line-delimited JSON
	data = append(data, '\n')
	_, err := t.stdin.Write(data)
	if err != nil {
		return &TransportError{Message: "write error", Err: err}
	}

	return nil
}

// Messages returns the message channel.
func (t *SubprocessTransport) Messages() <-chan json.RawMessage {
	return t.msgCh
}

// Errors returns the error channel.
func (t *SubprocessTransport) Errors() <-chan error {
	return t.errCh
}

// Done returns the done channel.
func (t *SubprocessTransport) Done() <-chan struct{} {
	return t.doneCh
}

// Close shuts down the subprocess.
func (t *SubprocessTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}
	t.closed = true

	// Close stdin to signal EOF to subprocess
	if t.stdin != nil {
		t.stdin.Close()
	}

	// Give the process a moment to exit gracefully
	// then force kill if needed
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}

	select {
	case <-t.doneCh:
	default:
		close(t.doneCh)
	}

	return nil
}
