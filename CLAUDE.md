# CLAUDE.md - Clawde Development Guidelines

## Project Overview

Clawde is a Go SDK for building AI agents with Claude Code capabilities. It provides a Go-idiomatic interface to the Claude Code CLI, supporting streaming responses, custom tools, hooks, and permission callbacks.

## Architecture

```
clawde/
├── client.go       # High-level Client API
├── stream.go       # Streaming response handler
├── query_func.go   # One-shot Query functions
├── query.go        # Control protocol handler
├── protocol.go     # Control protocol types
├── transport.go    # Transport interface
├── subprocess.go   # CLI subprocess transport
├── parser.go       # Message parsing
├── types.go        # Message and content types
├── options.go      # Options and functional options
├── errors.go       # Error types
├── permission.go   # Permission system
├── hooks.go        # Hook system
├── mcp.go          # MCP tool system
└── examples/       # Usage examples
```

## Key Design Principles

### 1. Idiomatic Go

- Use `context.Context` for cancellation and timeouts
- Return errors explicitly; no panics for recoverable errors
- Use channels for streaming data
- Prefer composition over inheritance
- Keep interfaces small (1-3 methods)

### 2. Functional Options Pattern

Configuration uses functional options:

```go
// Good
client, _ := clawde.NewClient(
    clawde.WithSystemPrompt("Be helpful"),
    clawde.WithMaxBudget(1.0),
)

// Not idiomatic Go
client := clawde.Client{
    SystemPrompt: "Be helpful",
    MaxBudget: 1.0,
}
```

### 3. Channel-Based Streaming

Use channels for message streaming:

```go
// Stream returns a channel
for stream.Next() {
    msg := stream.Current()
    // Process message
}

// Check for errors after loop
if err := stream.Err(); err != nil {
    // Handle error
}
```

### 4. Type-Safe Generics

Use generics for type-safe tool definitions:

```go
// Good - type-safe input
clawde.Tool("add", "Add numbers", func(ctx context.Context, input AddInput) (string, error) {
    return fmt.Sprintf("%f", input.A + input.B), nil
})

// Less safe - manual parsing
clawde.Tool("add", "Add numbers", func(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
    var in AddInput
    json.Unmarshal(input, &in)
    // ...
})
```

## Code Style

### Naming Conventions

- Package names: lowercase, single word (`clawde`)
- Exported types: PascalCase (`Client`, `HookMatcher`)
- Unexported types: camelCase (`rawMessage`, `queryHandler`)
- Acronyms: consistent casing (`HTTPClient` or `httpClient`, not `HttpClient`)

### Error Handling

```go
// Good: wrap errors with context
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}

// Good: sentinel errors for matching
var ErrNotConnected = errors.New("clawde: client not connected")

// Good: custom error types for details
type ProcessError struct {
    ExitCode int
    Stderr   string
}
```

### Comments

- Document all exported symbols
- Use complete sentences
- Start with the symbol name

```go
// Client is the main Clawde SDK client for interacting with Claude Code.
type Client struct { ... }

// Query sends a prompt and returns a stream of responses.
func (c *Client) Query(ctx context.Context, prompt string) (*Stream, error) { ... }
```

### JSON Tags

Use json tags for all serializable fields:

```go
type Message struct {
    Type    string `json:"type"`
    Content string `json:"content,omitempty"`
}
```

## Testing

Run tests:

```bash
go test ./...                    # Run all tests
go test -race ./...              # With race detector
go test -cover ./...             # With coverage
go test -v ./examples/...        # Test examples compile
```

### Test Patterns

```go
func TestClient_Query(t *testing.T) {
    // Arrange
    client, _ := NewClient()
    defer client.Close()

    // Act
    stream, err := client.Query(ctx, "test")

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

// Table-driven tests
func TestParseMessage(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Message
        wantErr bool
    }{
        {"user message", `{"type":"user"}`, &UserMessage{}, false},
        {"invalid json", `{invalid}`, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

## Dependencies

- Standard library preferred
- Minimize external dependencies
- Use `go mod tidy` to clean up

## Common Tasks

### Adding a New Option

1. Add field to `Options` struct in `options.go`
2. Create `With*` function
3. Handle in `SubprocessTransport.buildArgs()` if CLI flag needed
4. Add to README documentation

### Adding a New Hook Event

1. Add constant to `HookEvent` in `hooks.go`
2. Add input type if different from existing
3. Handle in `Query.handleHookCallback()`
4. Add example usage
5. Document in README

### Adding a New Message Type

1. Add struct in `types.go`
2. Implement `Message` interface (`isMessage()`, `MessageType()`)
3. Add case in `parser.go` `ParseMessage()`
4. Add to documentation

### Adding a New Example

1. Create directory `examples/NN_name/`
2. Add `main.go` with package documentation
3. Update `examples/README.md`
4. Test with `go run`

## Release Checklist

1. Update version in documentation
2. Run full test suite
3. Update CHANGELOG.md
4. Tag release: `git tag v1.x.x`
5. Push: `git push origin v1.x.x`

## Git Workflow

- Branch naming: `feat/description`, `fix/description`
- Commit messages: imperative mood ("Add feature", not "Added feature")
- Keep commits atomic and focused
- Squash fixup commits before merging

## Performance Considerations

- Use buffered channels for high-throughput scenarios
- Avoid unnecessary allocations in hot paths
- Pool large buffers (JSON parsing)
- Use `sync.Pool` for frequently allocated objects

## Security

- Never log sensitive data (API keys, tokens)
- Validate all input from subprocess
- Use timeouts for all operations
- Sanitize paths before use
