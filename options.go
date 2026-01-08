package clawde

import (
	"time"
)

// Options configures a Client or Query.
type Options struct {
	// SystemPrompt is the system prompt for the conversation.
	SystemPrompt string

	// AppendSystemPrompt appends to the default system prompt instead of replacing.
	AppendSystemPrompt string

	// Model specifies which Claude model to use.
	Model string

	// MaxTurns limits the number of agentic turns.
	MaxTurns int

	// MaxBudgetUSD sets the maximum spend for the query.
	MaxBudgetUSD float64

	// AllowedTools restricts which tools Claude can use.
	AllowedTools []string

	// DisallowedTools prevents Claude from using specific tools.
	DisallowedTools []string

	// PermissionMode controls how permissions are handled.
	PermissionMode PermissionMode

	// PermissionCallback is called when Claude wants to use a tool.
	PermissionCallback PermissionCallback

	// MCPServers configures MCP tool servers.
	MCPServers map[string]MCPServerConfig

	// SDKServers are in-process MCP servers.
	SDKServers map[string]*MCPServer

	// Hooks configures event hooks.
	Hooks map[HookEvent][]HookMatcher

	// CLIPath is the path to the Claude CLI executable.
	CLIPath string

	// WorkingDir is the working directory for the subprocess.
	WorkingDir string

	// Env sets additional environment variables.
	Env map[string]string

	// Timeout is the maximum duration for a query.
	Timeout time.Duration

	// ResumeConversation continues an existing conversation.
	ResumeConversation string
}

// Option is a functional option for configuring Options.
type Option func(*Options)

// WithSystemPrompt sets a custom system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = prompt
	}
}

// WithAppendSystemPrompt appends to the default system prompt.
func WithAppendSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.AppendSystemPrompt = prompt
	}
}

// WithModel sets the Claude model to use.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithMaxTurns limits the number of agentic turns.
func WithMaxTurns(n int) Option {
	return func(o *Options) {
		o.MaxTurns = n
	}
}

// WithMaxBudget sets the maximum spend in USD.
func WithMaxBudget(usd float64) Option {
	return func(o *Options) {
		o.MaxBudgetUSD = usd
	}
}

// WithAllowedTools restricts which tools can be used.
func WithAllowedTools(tools ...string) Option {
	return func(o *Options) {
		o.AllowedTools = tools
	}
}

// WithDisallowedTools prevents specific tools from being used.
func WithDisallowedTools(tools ...string) Option {
	return func(o *Options) {
		o.DisallowedTools = tools
	}
}

// WithPermissionMode sets the permission mode.
func WithPermissionMode(mode PermissionMode) Option {
	return func(o *Options) {
		o.PermissionMode = mode
	}
}

// WithPermissionCallback sets a callback for tool permission decisions.
func WithPermissionCallback(cb PermissionCallback) Option {
	return func(o *Options) {
		o.PermissionCallback = cb
	}
}

// WithMCPServer adds an external MCP server configuration.
func WithMCPServer(name string, cfg MCPServerConfig) Option {
	return func(o *Options) {
		if o.MCPServers == nil {
			o.MCPServers = make(map[string]MCPServerConfig)
		}
		o.MCPServers[name] = cfg
	}
}

// WithSDKServer adds an in-process MCP server.
func WithSDKServer(name string, server *MCPServer) Option {
	return func(o *Options) {
		if o.SDKServers == nil {
			o.SDKServers = make(map[string]*MCPServer)
		}
		o.SDKServers[name] = server
	}
}

// WithHook adds an event hook.
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

// WithTimeout sets the maximum duration for a query.
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.Timeout = d
	}
}

// WithResumeConversation continues an existing conversation.
func WithResumeConversation(sessionID string) Option {
	return func(o *Options) {
		o.ResumeConversation = sessionID
	}
}

// applyOptions applies functional options to create an Options struct.
func applyOptions(opts []Option) *Options {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
