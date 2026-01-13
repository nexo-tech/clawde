package clawde

import (
	"encoding/json"
	"time"
)

// AgentDefinition defines a custom agent with specific tools and prompts.
type AgentDefinition struct {
	// Description describes what the agent does.
	Description string `json:"description"`

	// Prompt is the system prompt for the agent.
	Prompt string `json:"prompt"`

	// Tools lists the tools available to the agent.
	Tools []string `json:"tools,omitempty"`

	// Model specifies the model for the agent ("sonnet", "opus", "haiku", "inherit").
	Model string `json:"model,omitempty"`
}

// SettingSource specifies where to load settings from.
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// PluginConfig configures a plugin to load.
type PluginConfig struct {
	// Type is the plugin type ("local").
	Type string `json:"type"`

	// Path is the path to the plugin directory.
	Path string `json:"path"`
}

// SystemPromptConfig can be either a string or a preset configuration.
type SystemPromptConfig struct {
	// String is a simple string prompt.
	String string

	// Preset configures a preset system prompt.
	Preset *SystemPromptPreset
}

// SystemPromptPreset configures a preset system prompt.
type SystemPromptPreset struct {
	// Type must be "preset".
	Type string `json:"type"`

	// Preset is the preset name (e.g., "claude_code").
	Preset string `json:"preset"`

	// Append is additional text to append to the preset.
	Append string `json:"append,omitempty"`
}

// MarshalJSON implements json.Marshaler for SystemPromptConfig.
func (s SystemPromptConfig) MarshalJSON() ([]byte, error) {
	if s.Preset != nil {
		return json.Marshal(s.Preset)
	}
	return json.Marshal(s.String)
}

// ToolsConfig can be either a list of tool names or a preset configuration.
type ToolsConfig struct {
	// Tools is a list of tool names.
	Tools []string

	// Preset configures a preset tools configuration.
	Preset *ToolsPreset
}

// ToolsPreset configures a preset tools configuration.
type ToolsPreset struct {
	// Type must be "preset".
	Type string `json:"type"`

	// Preset is the preset name (e.g., "claude_code").
	Preset string `json:"preset"`
}

// MarshalJSON implements json.Marshaler for ToolsConfig.
func (t ToolsConfig) MarshalJSON() ([]byte, error) {
	if t.Preset != nil {
		return json.Marshal(t.Preset)
	}
	return json.Marshal(t.Tools)
}

// StderrCallback is called for each line of stderr output.
type StderrCallback func(line string)

// Options configures a Client or Query.
type Options struct {
	// SystemPrompt is the system prompt for the conversation.
	SystemPrompt string

	// SystemPromptConfig is an advanced system prompt configuration.
	SystemPromptConfig *SystemPromptConfig

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

	// Tools configures available tools (array or preset).
	Tools *ToolsConfig

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

	// Agents configures custom agents.
	Agents map[string]AgentDefinition

	// SettingSources controls which settings are loaded.
	SettingSources []SettingSource

	// Plugins configures plugins to load.
	Plugins []PluginConfig

	// CLIPath is the path to the Claude CLI executable.
	CLIPath string

	// WorkingDir is the working directory for the subprocess.
	WorkingDir string

	// Env sets additional environment variables.
	Env map[string]string

	// Timeout is the maximum duration for a query.
	Timeout time.Duration

	// MaxThinkingTokens enables extended thinking with the specified token budget.
	// Minimum is 1024 tokens. Set to 0 to disable.
	MaxThinkingTokens int

	// ResumeConversation continues an existing conversation.
	ResumeConversation string

	// StderrCallback receives stderr output from the CLI.
	StderrCallback StderrCallback

	// IncludePartialMessages enables streaming of partial messages.
	IncludePartialMessages bool

	// ExtraArgs allows passing arbitrary CLI arguments.
	ExtraArgs map[string]string
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

// WithMaxThinkingTokens enables extended thinking with the specified token budget.
// Minimum is 1024 tokens. Set to 0 to disable extended thinking.
func WithMaxThinkingTokens(tokens int) Option {
	return func(o *Options) {
		o.MaxThinkingTokens = tokens
	}
}

// WithResumeConversation continues an existing conversation.
func WithResumeConversation(sessionID string) Option {
	return func(o *Options) {
		o.ResumeConversation = sessionID
	}
}

// WithAgents configures custom agents.
func WithAgents(agents map[string]AgentDefinition) Option {
	return func(o *Options) {
		o.Agents = agents
	}
}

// WithAgent adds a single custom agent.
func WithAgent(name string, agent AgentDefinition) Option {
	return func(o *Options) {
		if o.Agents == nil {
			o.Agents = make(map[string]AgentDefinition)
		}
		o.Agents[name] = agent
	}
}

// WithSettingSources controls which settings are loaded.
func WithSettingSources(sources ...SettingSource) Option {
	return func(o *Options) {
		o.SettingSources = sources
	}
}

// WithPlugins configures plugins to load.
func WithPlugins(plugins ...PluginConfig) Option {
	return func(o *Options) {
		o.Plugins = plugins
	}
}

// WithPlugin adds a single plugin.
func WithPlugin(pluginType, path string) Option {
	return func(o *Options) {
		o.Plugins = append(o.Plugins, PluginConfig{
			Type: pluginType,
			Path: path,
		})
	}
}

// WithTools configures available tools (array or preset).
func WithTools(tools *ToolsConfig) Option {
	return func(o *Options) {
		o.Tools = tools
	}
}

// WithToolsList configures available tools as a list.
func WithToolsList(tools ...string) Option {
	return func(o *Options) {
		o.Tools = &ToolsConfig{Tools: tools}
	}
}

// WithToolsPreset configures tools using a preset.
func WithToolsPreset(preset string) Option {
	return func(o *Options) {
		o.Tools = &ToolsConfig{
			Preset: &ToolsPreset{
				Type:   "preset",
				Preset: preset,
			},
		}
	}
}

// WithSystemPromptPreset configures a preset system prompt.
func WithSystemPromptPreset(preset string, append string) Option {
	return func(o *Options) {
		o.SystemPromptConfig = &SystemPromptConfig{
			Preset: &SystemPromptPreset{
				Type:   "preset",
				Preset: preset,
				Append: append,
			},
		}
	}
}

// WithStderrCallback sets a callback for stderr output.
func WithStderrCallback(cb StderrCallback) Option {
	return func(o *Options) {
		o.StderrCallback = cb
	}
}

// WithIncludePartialMessages enables streaming of partial messages.
func WithIncludePartialMessages(include bool) Option {
	return func(o *Options) {
		o.IncludePartialMessages = include
	}
}

// WithExtraArgs sets arbitrary CLI arguments.
func WithExtraArgs(args map[string]string) Option {
	return func(o *Options) {
		o.ExtraArgs = args
	}
}

// WithExtraArg adds a single CLI argument.
func WithExtraArg(key, value string) Option {
	return func(o *Options) {
		if o.ExtraArgs == nil {
			o.ExtraArgs = make(map[string]string)
		}
		o.ExtraArgs[key] = value
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
