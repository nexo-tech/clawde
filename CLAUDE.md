# CLAUDE.md - Development Guidelines

This document provides guidelines for developing and contributing to the Clawde SDK.

## Project Overview

Clawde is a Go SDK for building AI agents with Claude Code capabilities. It provides a Go-idiomatic interface to Claude via the CLI subprocess.

## Architecture

```
clawde/
├── types.go        # Message and content block types
├── options.go      # Options struct and functional options
├── errors.go       # Error types
├── permission.go   # Permission system
├── hooks.go        # Hook system
├── mcp.go          # MCP tool system
├── transport.go    # Transport interface
├── subprocess.go   # CLI subprocess transport
├── parser.go       # Message parsing
├── protocol.go     # Control protocol types
├── query.go        # Protocol handler
├── client.go       # High-level client
├── stream.go       # Response streaming
├── query_func.go   # Convenience functions
└── examples/       # Usage examples
```

## Design Principles

### 1. Idiomatic Go
- Use `context.Context` for cancellation and timeouts
- Return errors explicitly (no panic)
- Use channels for streaming
- Use interfaces for abstraction

### 2. Functional Options
```go
client, _ := clawde.NewClient(
    clawde.WithSystemPrompt("..."),
    clawde.WithMaxTurns(10),
)
```

### 3. Small Interfaces
```go
type Transport interface {
    Start(ctx context.Context) error
    Write(data []byte) error
    Messages() <-chan json.RawMessage
    Errors() <-chan error
    Close() error
}
```

### 4. Composition Over Inheritance
- Use embedding for shared behavior
- Prefer interfaces over base structs

## Code Style

### Naming
- Exported types: `PascalCase`
- Unexported types: `camelCase`
- Acronyms: `MCP`, `JSON`, `URL` (all caps)
- Error variables: `ErrSomething`

### Documentation
- Every exported type/function has a doc comment
- Start with the name: `// Client provides...`
- Include examples for complex APIs

### Error Handling
```go
// Define sentinel errors
var ErrNotConnected = errors.New("clawde: client not connected")

// Wrap errors with context
return fmt.Errorf("clawde: failed to start: %w", err)
```

### Concurrency
- Protect shared state with `sync.Mutex`
- Use channels for communication
- Always respect context cancellation

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific test
go test -run TestClient ./...
```

## Common Tasks

### Adding a New Option
1. Add field to `Options` struct in `options.go`
2. Create `WithXxx` function
3. Handle in relevant code (subprocess args, etc.)

### Adding a New Hook Event
1. Add constant to `HookEvent` in `hooks.go`
2. Handle in `handleHookCallback` in `query.go`

### Adding a New Message Type
1. Add type to `types.go` implementing `Message`
2. Add parsing logic to `parser.go`

## Dependencies

The SDK uses only the Go standard library:
- `context` - Cancellation and timeouts
- `encoding/json` - JSON marshaling
- `os/exec` - Subprocess management
- `sync` - Concurrency primitives
- `bufio` - Buffered I/O

## Release Process

1. Update version in code if applicable
2. Run full test suite
3. Update CHANGELOG
4. Tag release: `git tag v1.0.0`
5. Push: `git push origin v1.0.0`

## Troubleshooting

### Claude CLI not found
- Ensure `claude` is in PATH
- Or use `WithCLIPath("/path/to/claude")`

### Permission denied
- Check Claude CLI authentication
- Verify API key is set

### Stream hangs
- Check context timeout
- Ensure Claude CLI is responding
