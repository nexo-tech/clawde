# Clawde SDK Examples

This directory contains runnable examples demonstrating the Clawde SDK.

## Running Examples

Each example is a standalone Go program:

```bash
# From repository root
go run ./examples/01_quickstart/

# Or from example directory
cd examples/01_quickstart && go run .
```

## Prerequisites

- Go 1.21 or later
- Claude Code CLI installed and accessible in PATH

## Example Index

### Basic Usage

| # | Name | Description | Key Concepts |
|---|------|-------------|--------------|
| 01 | [quickstart](01_quickstart/) | Simple one-shot queries | `Query()`, streaming, options |
| 02 | [streaming](02_streaming/) | Stream processing patterns | Message types, collecting results |

### Custom Tools

| # | Name | Description | Key Concepts |
|---|------|-------------|--------------|
| 03 | [mcp_tools](03_mcp_tools/) | Calculator with custom tools | MCP server, `Tool[T]()` generic |

### Control Flow

| # | Name | Description | Key Concepts |
|---|------|-------------|--------------|
| 04 | [hooks](04_hooks/) | Hook system for interception | `HookPreToolUse`, `HookPostToolUse` |
| 06 | [human_in_loop](06_human_in_loop/) | Interactive approval flow | `PermissionCallback`, user prompts |

### Sessions & Events

| # | Name | Description | Key Concepts |
|---|------|-------------|--------------|
| 05 | [interactive](05_interactive/) | REPL-style session | Multi-turn, conversation context |
| 07 | [events](07_events/) | Event tracking & metrics | Hooks for monitoring |

### Advanced

| # | Name | Description | Key Concepts |
|---|------|-------------|--------------|
| 08 | [budget](08_budget/) | Cost control & tracking | `WithMaxBudget`, cost monitoring |
| 09 | [structured_output](09_structured_output/) | JSON-formatted responses | `WithOutputFormat`, schemas |

## Code Patterns

### Error Handling

All examples demonstrate proper Go error handling:

```go
stream, err := clawde.Query(ctx, "Hello")
if err != nil {
    log.Fatalf("query failed: %v", err)
}

// Check for stream errors
if err := stream.Err(); err != nil {
    log.Printf("stream error: %v", err)
}
```

### Context Usage

Examples use `context.Context` for cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, _ := clawde.Query(ctx, "Hello")
```

### Resource Cleanup

Clients and streams are properly closed:

```go
client, _ := clawde.NewClient()
defer client.Close()

stream, _ := client.Query(ctx, "Hello")
defer stream.Close()
```

### Streaming Patterns

```go
// Pattern 1: Iterate and process
for stream.Next() {
    msg := stream.Current()
    // Handle message
}

// Pattern 2: Collect all
messages, err := stream.Collect()

// Pattern 3: Get text only
text, err := stream.CollectText()

// Pattern 4: Wait for result
result, err := stream.Wait()
```

## Common Options

```go
// System prompt
clawde.WithSystemPrompt("Be concise")

// Model selection
clawde.WithModel("claude-sonnet-4-20250514")

// Tool configuration
clawde.WithAllowedTools("Read", "Write", "Bash")

// Budget control
clawde.WithMaxBudget(1.00) // $1.00 max

// Hooks
clawde.WithHook(clawde.HookPreToolUse, matcher)

// Permissions
clawde.WithPermissionCallback(callback)
```
