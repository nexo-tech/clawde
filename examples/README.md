# Clawde Examples

This directory contains examples demonstrating various features of the Clawde SDK.

## Running Examples

```bash
# From the repository root
go run ./examples/01_quickstart

# Or from the example directory
cd examples/01_quickstart
go run main.go
```

## Examples

### 01_quickstart
The simplest usage of Clawde - one-shot queries and streaming responses.

### 02_streaming
Real-time streaming of Claude's responses with progress information.

### 03_mcp_tools
Creating custom tools using the MCP (Model Context Protocol) system. Demonstrates a calculator with add, multiply, and square root operations.

### 04_hooks
Using hooks to intercept and control tool usage. Shows how to block dangerous commands and log file operations.

### 05_interactive
A REPL-style interactive session with Claude, demonstrating continuous conversation.

### 06_human_in_loop
Permission callbacks that require human approval before tool execution.

### 07_events
Handling different types of events and messages from Claude, including tool uses, system messages, and results.

### 08_budget
Setting and enforcing budget limits for queries, with cost tracking.

### 09_structured_output
Parsing structured JSON data from Claude's responses.

### 10_agents
Custom agent definitions with specific tools, prompts, and models.

### 11_system_prompt
Different system prompt configurations - string, preset, and append patterns.

### 12_tools_option
Configuring available tools - array, empty array, and preset configurations.

### 13_setting_sources
Controlling which settings are loaded (user, project, local).

### 14_plugins
Loading plugins from the filesystem.

### 15_stderr_callback
Capturing CLI debug output via a callback function.

### 16_partial_messages
Streaming partial messages for real-time UI updates.

---

## TypeScript SDK Parity Examples

The following examples are ported from the TypeScript Claude Agent SDK demos to ensure feature parity.

### 17_hello_world_ts
**Parity with:** `claude-agent-sdk-demos/hello-world`

Basic query with hooks for path validation. Demonstrates:
- AllowedTools configuration
- PreToolUse hooks for file path validation
- Message content extraction

### 18_session_api
**Parity with:** `claude-agent-sdk-demos/hello-world-v2`

V2-style session-based interface with send/stream pattern. Includes four sub-examples:
- `basic` - Basic session with send/stream
- `multi-turn` - Multi-turn conversations with context
- `one-shot` - Convenience function for single queries
- `resume` - Session persistence and resumption

```bash
go run main.go basic
go run main.go multi-turn
go run main.go one-shot
go run main.go resume
```

### 20_resume_generator
**Parity with:** `claude-agent-sdk-demos/resume-generator`

Web search and document generation with custom system prompts. Demonstrates:
- CLI argument parsing
- Custom system prompts
- WebSearch tool for research
- Skill integration for document generation
- Progress output with tool tracking

```bash
go run main.go "Person Name"
```

### 21_research_agent
**Parity with:** `claude-agent-sdk-demos/research-agent`

Multi-agent research coordination system. Demonstrates:
- Lead agent that orchestrates subagents
- Specialized subagents (researcher, data-analyst, report-writer)
- SubagentTracker with tool call logging
- TranscriptWriter for session logs
- Hooks for pre/post tool use tracking

Subagent types:
- `researcher` - Gathers information via WebSearch
- `data-analyst` - Generates charts with matplotlib
- `report-writer` - Creates PDF reports with reportlab

---

## Prerequisites

- Claude CLI installed and authenticated
- Go 1.21 or later

## Common Patterns

### One-shot Query
```go
text, err := clawde.QueryText(ctx, "Your question here")
```

### Streaming Response
```go
stream, err := clawde.Query(ctx, "Your prompt")
for stream.Next() {
    if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
        fmt.Print(msg.Text())
    }
}
```

### Custom Tools
```go
server := clawde.NewMCPServer("mytools")
server.Tools = append(server.Tools, clawde.Tool("mytool", "description", handler))
client, _ := clawde.NewClient(clawde.WithSDKServer("mytools", server))
```

### Hooks
```go
hook := clawde.MatchTool("Bash", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
    // Check command and allow/block
    return clawde.ContinueHook(), nil
})
client, _ := clawde.NewClient(clawde.WithHook(clawde.HookPreToolUse, hook))
```

### V2 Session API
```go
// Create session
session, _ := clawde.CreateSession(ctx, clawde.WithModel("sonnet"))
defer session.Close()

// Send message
session.Send(ctx, "Hello!")

// Stream response
stream, _ := session.Stream(ctx)
for stream.Next() {
    if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
        fmt.Print(msg.Text())
    }
}

// Resume session later
session2, _ := clawde.ResumeSession(ctx, session.SessionID(), clawde.WithModel("sonnet"))
```

### One-Shot Prompt
```go
result, _ := clawde.Prompt(ctx, "What is 2+2?", clawde.WithModel("sonnet"))
fmt.Println(result.Text)
fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
```

### Subagent Tracking
```go
tracker, _ := clawde.NewSubagentTracker("./logs/session")
defer tracker.Close()

preHook := clawde.MatchTool(".*", tracker.PreToolUseHook)
postHook := clawde.MatchTool(".*", tracker.PostToolUseHook)

client, _ := clawde.NewClient(
    clawde.WithHook(clawde.HookPreToolUse, preHook),
    clawde.WithHook(clawde.HookPostToolUse, postHook),
)
```
