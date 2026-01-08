# Clawde 🦫

A Go-idiomatic SDK for building AI agents with Claude Code capabilities.

> **Clawde** = Claw + Claude — The Go Gopher's SDK for Claude!

## Installation

```bash
go get github.com/nexo-tech/clawde
```

## Quick Start

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

    // One-shot query
    text, err := clawde.QueryText(ctx, "What is 2+2?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(text)
}
```

## Features

- **Simple Query API** - One-liner queries with `QueryText()` and `Query()`
- **Streaming Responses** - Real-time streaming with channel-based iteration
- **Custom MCP Tools** - Create tools using Go generics for type-safe handlers
- **Hook System** - Intercept and control tool execution with pre/post hooks
- **Permission Callbacks** - Human-in-the-loop approval for tool usage
- **Budget Control** - Set spending limits and turn counts
- **Idiomatic Go** - Context for cancellation, functional options, interfaces

## Examples

### Streaming Response

```go
stream, _ := clawde.Query(ctx, "Tell me a joke")
for stream.Next() {
    if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
        fmt.Print(msg.Text())
    }
}
```

### Custom Tools

```go
type AddInput struct {
    A float64 `json:"a"`
    B float64 `json:"b"`
}

server := clawde.NewMCPServer("calculator")
server.Tools = append(server.Tools,
    clawde.Tool("add", "Add two numbers", func(ctx context.Context, in AddInput) (string, error) {
        return fmt.Sprintf("%f", in.A + in.B), nil
    }),
)

client, _ := clawde.NewClient(clawde.WithSDKServer("calculator", server))
```

### Hooks

```go
bashGuard := clawde.MatchTool("Bash", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
    if strings.Contains(input.Command(), "rm -rf") {
        return clawde.BlockHook("Dangerous command blocked"), nil
    }
    return clawde.ContinueHook(), nil
})

client, _ := clawde.NewClient(clawde.WithHook(clawde.HookPreToolUse, bashGuard))
```

### Permission Callback

```go
client, _ := clawde.NewClient(
    clawde.WithPermissionCallback(func(ctx context.Context, req *clawde.PermissionRequest) clawde.PermissionResult {
        fmt.Printf("Allow %s? [y/n]: ", req.ToolName)
        // Get user input...
        return clawde.Allow() // or clawde.Deny("reason")
    }),
)
```

### Configuration Options

```go
client, _ := clawde.NewClient(
    clawde.WithSystemPrompt("You are a helpful assistant"),
    clawde.WithModel("claude-sonnet-4-20250514"),
    clawde.WithMaxTurns(10),
    clawde.WithMaxBudget(1.00),
    clawde.WithAllowedTools("Read", "Bash"),
    clawde.WithPermissionMode(clawde.PermissionAcceptEdits),
    clawde.WithTimeout(5 * time.Minute),
)
```

## API Reference

### Client Functions

| Function | Description |
|----------|-------------|
| `NewClient(opts...)` | Create a new client |
| `Query(ctx, prompt, opts...)` | One-shot query returning a stream |
| `QueryText(ctx, prompt, opts...)` | One-shot query returning text |
| `QueryResult(ctx, prompt, opts...)` | One-shot query returning all messages |

### Client Methods

| Method | Description |
|--------|-------------|
| `Connect(ctx)` | Connect to Claude |
| `Query(ctx, prompt)` | Send a query and get a stream |
| `Send(ctx, prompt)` | Send without waiting |
| `Receive(ctx)` | Get message channel |
| `Interrupt()` | Interrupt current query |
| `Close()` | Close the client |

### Stream Methods

| Method | Description |
|--------|-------------|
| `Next()` | Advance to next message |
| `Current()` | Get current message |
| `Text()` | Get accumulated text |
| `Message()` | Get accumulated AssistantMessage |
| `Result()` | Get final ResultMessage |
| `Collect()` | Collect all messages |
| `CollectText()` | Collect all text |
| `Wait()` | Wait until done |
| `Err()` | Get any error |
| `Close()` | Cancel the stream |

### Message Types

- `AssistantMessage` - Claude's response with text, thinking, and tool uses
- `UserMessage` - User input
- `SystemMessage` - System notifications
- `ResultMessage` - Final query result with stats
- `StreamEvent` - Streaming events

### Content Blocks

- `TextBlock` - Text content
- `ThinkingBlock` - Claude's reasoning
- `ToolUseBlock` - Tool invocation
- `ToolResultBlock` - Tool result

## Requirements

- Go 1.21+
- Claude CLI installed and authenticated (`claude` in PATH)

## License

MIT License - see [LICENSE](LICENSE)
