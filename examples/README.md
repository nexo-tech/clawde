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
