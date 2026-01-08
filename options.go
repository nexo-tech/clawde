package clawde

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Options configures the Clawde client.
type Options struct {
	// SystemPrompt sets the system prompt for the conversation.
	SystemPrompt string

	// SystemPromptFile reads the system prompt from a file.
	SystemPromptFile string

	// Model specifies the Claude model to use.
	Model string

	// FallbackModel specifies a fallback model if the primary is unavailable.
	FallbackModel string

	// MaxTurns limits the number of conversation turns.
	MaxTurns int

	// MaxBudgetUSD sets the maximum cost in USD.
	MaxBudgetUSD float64

	// MaxThinkingTokens limits thinking tokens for extended thinking.
	MaxThinkingTokens int

	// AllowedTools specifies which tools are allowed.
	AllowedTools []string

	// DisallowedTools specifies which tools are disallowed.
	DisallowedTools []string

	// PermissionMode controls how permissions are handled.
	PermissionMode PermissionMode

	// PermissionCallback is called for permission decisions.
	PermissionCallback PermissionCallback

	// MCPServers configures MCP server connections.
	MCPServers map[string]*MCPServerConfig

	// Hooks configures event hooks.
	Hooks map[HookEvent][]HookMatcher

	// CLIPath specifies the path to the Claude CLI.
	CLIPath string

	// WorkingDir sets the working directory for the CLI.
	WorkingDir string

	// Env sets additional environment variables.
	Env map[string]string

	// ContinueSession resumes a previous session.
	ContinueSession bool

	// ResumeSessionID specifies a session ID to resume.
	ResumeSessionID string

	// ForkSession creates a fork of the session.
	ForkSession bool

	// IncludePartialMessages includes partial streaming messages.
	IncludePartialMessages bool

	// OutputFormat specifies a JSON schema for structured output.
	OutputFormat json.RawMessage

	// Timeout sets the default timeout for operations.
	Timeout time.Duration

	// Agents configures custom agent definitions.
	Agents map[string]*AgentDefinition

	// Sandbox configures sandbox settings.
	Sandbox *SandboxSettings
}

// DefaultOptions returns the default options.
func DefaultOptions() *Options {
	return &Options{
		MCPServers: make(map[string]*MCPServerConfig),
		Hooks:      make(map[HookEvent][]HookMatcher),
		Env:        make(map[string]string),
		Agents:     make(map[string]*AgentDefinition),
		Timeout:    5 * time.Minute,
	}
}

// Option is a functional option for configuring Options.
type Option func(*Options)

// WithSystemPrompt sets the system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = prompt
	}
}

// WithSystemPromptFile reads the system prompt from a file.
func WithSystemPromptFile(path string) Option {
	return func(o *Options) {
		o.SystemPromptFile = path
	}
}

// WithModel sets the model to use.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithFallbackModel sets the fallback model.
func WithFallbackModel(model string) Option {
	return func(o *Options) {
		o.FallbackModel = model
	}
}

// WithMaxTurns sets the maximum number of turns.
func WithMaxTurns(n int) Option {
	return func(o *Options) {
		o.MaxTurns = n
	}
}

// WithMaxBudget sets the maximum budget in USD.
func WithMaxBudget(usd float64) Option {
	return func(o *Options) {
		o.MaxBudgetUSD = usd
	}
}

// WithMaxThinkingTokens sets the maximum thinking tokens.
func WithMaxThinkingTokens(n int) Option {
	return func(o *Options) {
		o.MaxThinkingTokens = n
	}
}

// WithAllowedTools sets the allowed tools.
func WithAllowedTools(tools ...string) Option {
	return func(o *Options) {
		o.AllowedTools = append(o.AllowedTools, tools...)
	}
}

// WithDisallowedTools sets the disallowed tools.
func WithDisallowedTools(tools ...string) Option {
	return func(o *Options) {
		o.DisallowedTools = append(o.DisallowedTools, tools...)
	}
}

// WithPermissionMode sets the permission mode.
func WithPermissionMode(mode PermissionMode) Option {
	return func(o *Options) {
		o.PermissionMode = mode
	}
}

// WithPermissionCallback sets the permission callback.
func WithPermissionCallback(cb PermissionCallback) Option {
	return func(o *Options) {
		o.PermissionCallback = cb
	}
}

// WithMCPServer adds an MCP server configuration.
func WithMCPServer(name string, cfg *MCPServerConfig) Option {
	return func(o *Options) {
		if o.MCPServers == nil {
			o.MCPServers = make(map[string]*MCPServerConfig)
		}
		o.MCPServers[name] = cfg
	}
}

// WithSDKMCPServer adds an in-process SDK MCP server.
func WithSDKMCPServer(server *MCPServer) Option {
	return func(o *Options) {
		if o.MCPServers == nil {
			o.MCPServers = make(map[string]*MCPServerConfig)
		}
		o.MCPServers[server.Name] = &MCPServerConfig{
			Type:      MCPServerTypeSDK,
			SDKServer: server,
		}
	}
}

// WithHook adds a hook for the specified event.
func WithHook(event HookEvent, matcher HookMatcher) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[HookEvent][]HookMatcher)
		}
		o.Hooks[event] = append(o.Hooks[event], matcher)
	}
}

// WithCLIPath sets the path to the Claude CLI.
func WithCLIPath(path string) Option {
	return func(o *Options) {
		o.CLIPath = path
	}
}

// WithWorkingDir sets the working directory.
func WithWorkingDir(dir string) Option {
	return func(o *Options) {
		o.WorkingDir = dir
	}
}

// WithEnv adds environment variables.
func WithEnv(env map[string]string) Option {
	return func(o *Options) {
		if o.Env == nil {
			o.Env = make(map[string]string)
		}
		for k, v := range env {
			o.Env[k] = v
		}
	}
}

// WithContinueSession enables session continuation.
func WithContinueSession() Option {
	return func(o *Options) {
		o.ContinueSession = true
	}
}

// WithResumeSession resumes a specific session.
func WithResumeSession(sessionID string) Option {
	return func(o *Options) {
		o.ResumeSessionID = sessionID
	}
}

// WithForkSession enables session forking.
func WithForkSession() Option {
	return func(o *Options) {
		o.ForkSession = true
	}
}

// WithPartialMessages enables partial message streaming.
func WithPartialMessages() Option {
	return func(o *Options) {
		o.IncludePartialMessages = true
	}
}

// WithOutputFormat sets a JSON schema for structured output.
func WithOutputFormat(schema json.RawMessage) Option {
	return func(o *Options) {
		o.OutputFormat = schema
	}
}

// WithTimeout sets the default timeout.
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.Timeout = d
	}
}

// WithAgent adds a custom agent definition.
func WithAgent(name string, def *AgentDefinition) Option {
	return func(o *Options) {
		if o.Agents == nil {
			o.Agents = make(map[string]*AgentDefinition)
		}
		o.Agents[name] = def
	}
}

// WithSandbox sets sandbox settings.
func WithSandbox(settings *SandboxSettings) Option {
	return func(o *Options) {
		o.Sandbox = settings
	}
}

// AgentDefinition defines a custom agent.
type AgentDefinition struct {
	// Description describes what the agent does.
	Description string `json:"description"`

	// Prompt is the system prompt for the agent.
	Prompt string `json:"prompt"`

	// Tools specifies which tools the agent can use.
	Tools []string `json:"tools,omitempty"`

	// Model specifies the model for the agent.
	Model string `json:"model,omitempty"` // "sonnet", "opus", "haiku", "inherit"
}

// SandboxSettings configures sandbox behavior.
type SandboxSettings struct {
	// Enabled enables sandboxing.
	Enabled bool `json:"enabled,omitempty"`

	// AutoAllowBash auto-allows bash if sandboxed.
	AutoAllowBash bool `json:"autoAllowBashIfSandboxed,omitempty"`

	// ExcludedCommands lists commands excluded from sandbox.
	ExcludedCommands []string `json:"excludedCommands,omitempty"`

	// AllowUnsandboxed allows unsandboxed commands.
	AllowUnsandboxed bool `json:"allowUnsandboxedCommands,omitempty"`

	// Network configures network sandbox settings.
	Network *SandboxNetworkConfig `json:"network,omitempty"`
}

// SandboxNetworkConfig configures network sandboxing.
type SandboxNetworkConfig struct {
	// AllowUnixSockets lists allowed unix sockets.
	AllowUnixSockets []string `json:"allowUnixSockets,omitempty"`

	// AllowAllUnixSockets allows all unix sockets.
	AllowAllUnixSockets bool `json:"allowAllUnixSockets,omitempty"`

	// AllowLocalBinding allows binding to local ports.
	AllowLocalBinding bool `json:"allowLocalBinding,omitempty"`

	// HTTPProxyPort sets the HTTP proxy port.
	HTTPProxyPort int `json:"httpProxyPort,omitempty"`

	// SOCKSProxyPort sets the SOCKS proxy port.
	SOCKSProxyPort int `json:"socksProxyPort,omitempty"`
}

// findCLI searches for the Claude CLI in common locations.
func findCLI() string {
	// Check explicit environment variable
	if path := os.Getenv("CLAUDE_CLI_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Common CLI names
	names := []string{"claude"}

	// Check PATH
	for _, name := range names {
		if path, err := findInPath(name); err == nil {
			return path
		}
	}

	// Common installation locations
	home, _ := os.UserHomeDir()
	locations := []string{
		"/usr/local/bin/claude",
		"/usr/bin/claude",
		filepath.Join(home, ".local", "bin", "claude"),
		filepath.Join(home, ".claude", "bin", "claude"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// findInPath searches for an executable in PATH.
func findInPath(name string) (string, error) {
	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		full := filepath.Join(dir, name)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return full, nil
		}
	}
	return "", os.ErrNotExist
}
