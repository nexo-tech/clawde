# Clawde 🦫

A Go-idiomatic SDK for building AI agents with Claude Code capabilities.

> **Clawde** = Claw + Claude — The Go Gopher's SDK for Claude!

## Features

- **Simple Query API** — One-liner queries with streaming responses
- **Full Client API** — Multi-turn conversations with persistent context
- **Custom MCP Tools** — Type-safe tool creation with Go generics
- **Hook System** — Intercept and control agent behavior
- **Permission Callbacks** — Human-in-the-loop approval flows
- **Budget Control** — Cost limits and usage tracking
- **Structured Output** — JSON Schema response validation

## Installation

```bash
go get github.com/nexo-tech/clawde
```

### Prerequisites

- Go 1.21 or later
- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed

## Quick Start

### Simple Query

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/nexo-tech/clawde"
)

func main() {
    ctx := context.Background()

    stream, err := clawde.Query(ctx, "What is 2 + 2?")
    if err != nil {
        log.Fatal(err)
    }
    defer stream.Close()

    for stream.Next() {
        if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
            fmt.Print(msg.Text())
        }
    }
}
```

### Client with Options

```go
client, err := clawde.NewClient(
    clawde.WithSystemPrompt("You are a helpful coding assistant."),
    clawde.WithAllowedTools("Read", "Write", "Bash"),
    clawde.WithMaxBudget(1.00), // $1.00 limit
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}

stream, _ := client.Query(ctx, "List files in current directory")
for stream.Next() {
    fmt.Print(stream.Current())
}
```

### Custom MCP Tools

```go
// Create an MCP server with type-safe tools
server := clawde.NewMCPServer("calculator")

type AddInput struct {
    A float64 `json:"a" description:"First number"`
    B float64 `json:"b" description:"Second number"`
}

server.AddTool(clawde.Tool("add", "Add two numbers",
    func(ctx context.Context, input AddInput) (string, error) {
        return fmt.Sprintf("%.2f", input.A + input.B), nil
    },
))

client, _ := clawde.NewClient(
    clawde.WithSDKMCPServer(server),
    clawde.WithAllowedTools("mcp__calculator__add"),
)
```

### Hooks for Interception

```go
// Block dangerous commands
bashGuard := clawde.MatchTool("Bash", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
    var toolInput struct { Command string `json:"command"` }
    input.ToolInputAs(&toolInput)

    if strings.Contains(toolInput.Command, "rm -rf") {
        return clawde.BlockHook("Dangerous command blocked"), nil
    }
    return clawde.ContinueHook(), nil
})

client, _ := clawde.NewClient(
    clawde.WithHook(clawde.HookPreToolUse, bashGuard),
)
```

### Permission Callbacks

```go
// Ask user before executing tools
askUser := func(ctx context.Context, req *clawde.PermissionRequest) clawde.PermissionResult {
    fmt.Printf("Allow %s? (y/n): ", req.ToolName)
    var response string
    fmt.Scanln(&response)

    if response == "y" {
        return clawde.Allow()
    }
    return clawde.Deny("User denied")
}

client, _ := clawde.NewClient(
    clawde.WithPermissionCallback(askUser),
)
```

## Examples

| Example | Description |
|---------|-------------|
| [01_quickstart](examples/01_quickstart/) | Basic query usage |
| [02_streaming](examples/02_streaming/) | Stream processing patterns |
| [03_mcp_tools](examples/03_mcp_tools/) | Custom calculator tools |
| [04_hooks](examples/04_hooks/) | Hook system for control flow |
| [05_interactive](examples/05_interactive/) | REPL-style sessions |
| [06_human_in_loop](examples/06_human_in_loop/) | Interactive approvals |
| [07_events](examples/07_events/) | Event tracking & metrics |
| [08_budget](examples/08_budget/) | Cost control |
| [09_structured_output](examples/09_structured_output/) | JSON Schema responses |

Run examples:

```bash
go run ./examples/01_quickstart/
```

## API Reference

### Query Functions

```go
// One-shot query with streaming
stream, err := clawde.Query(ctx, "prompt", opts...)

// Convenience: get text only
text, err := clawde.QueryText(ctx, "prompt")

// Convenience: wait for result
result, err := clawde.QueryResult(ctx, "prompt")
```

### Client Methods

```go
client, _ := clawde.NewClient(opts...)
defer client.Close()

client.Connect(ctx)           // Connect to CLI
client.Query(ctx, "prompt")   // Send query, get stream
client.Send(ctx, "prompt")    // Send without stream
client.Receive(ctx)           // Get message channel
client.Interrupt()            // Stop current operation
client.SetPermissionMode(m)   // Change permission mode
client.SetModel(model)        // Change model
client.SessionID()            // Get session ID
```

### Stream Methods

```go
stream.Next()           // Advance to next message
stream.Current()        // Get current message
stream.Message()        // Get accumulated assistant message
stream.Result()         // Get result message
stream.Err()            // Get any error
stream.Close()          // Close stream
stream.Collect()        // Collect all messages
stream.CollectText()    // Get all text
stream.Wait()           // Wait for result
```

### Options

```go
clawde.WithSystemPrompt(prompt)      // Set system prompt
clawde.WithModel(model)              // Set model
clawde.WithMaxTurns(n)               // Limit turns
clawde.WithMaxBudget(usd)            // Set budget limit
clawde.WithAllowedTools(tools...)    // Allow specific tools
clawde.WithDisallowedTools(tools...) // Block specific tools
clawde.WithPermissionMode(mode)      // Set permission mode
clawde.WithPermissionCallback(cb)    // Set permission callback
clawde.WithHook(event, matcher)      // Add hook
clawde.WithSDKMCPServer(server)      // Add MCP server
clawde.WithCLIPath(path)             // Set CLI path
clawde.WithWorkingDir(dir)           // Set working directory
clawde.WithTimeout(duration)         // Set timeout
clawde.WithOutputFormat(schema)      // Set JSON schema output
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      User Code                          │
├─────────────────────────────────────────────────────────┤
│  Query() / Client                                       │
│  └── Stream (channel-based message delivery)            │
├─────────────────────────────────────────────────────────┤
│  Query Handler                                          │
│  ├── Control Protocol (permissions, hooks, MCP)         │
│  └── Message Parser (JSON → typed messages)             │
├─────────────────────────────────────────────────────────┤
│  Transport                                              │
│  └── SubprocessTransport (Claude CLI stdin/stdout)      │
├─────────────────────────────────────────────────────────┤
│                    Claude Code CLI                      │
└─────────────────────────────────────────────────────────┘
```

## License

MIT License — see [LICENSE](LICENSE)

## Contributing

Contributions welcome! Please read the [CLAUDE.md](CLAUDE.md) for development guidelines.
