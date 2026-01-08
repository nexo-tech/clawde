// Package main demonstrates the hook system for intercepting and
// controlling agent behavior at various execution points.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx := context.Background()

	// Example 1: PreToolUse hook to block dangerous commands
	fmt.Println("=== Bash Guard Hook ===")
	if err := bashGuardExample(ctx); err != nil {
		log.Printf("Bash guard example failed: %v", err)
	}

	fmt.Println()

	// Example 2: PostToolUse hook to monitor output
	fmt.Println("=== Output Monitor Hook ===")
	if err := outputMonitorExample(ctx); err != nil {
		log.Printf("Output monitor example failed: %v", err)
	}

	fmt.Println()

	// Example 3: Logging all tool uses
	fmt.Println("=== Tool Use Logger ===")
	if err := toolUseLoggerExample(ctx); err != nil {
		log.Printf("Tool use logger failed: %v", err)
	}
}

// bashGuardExample blocks execution of dangerous bash commands.
func bashGuardExample(ctx context.Context) error {
	// Define forbidden patterns
	forbidden := []string{"rm -rf", "sudo", "chmod 777", "> /dev/"}

	// Create PreToolUse hook for Bash
	bashHook := clawde.MatchTool("Bash", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		// Extract command from input
		var toolInput struct {
			Command string `json:"command"`
		}
		if err := input.ToolInputAs(&toolInput); err != nil {
			return clawde.ContinueHook(), nil
		}

		command := toolInput.Command
		fmt.Printf("  [Hook] Checking command: %s\n", command)

		for _, pattern := range forbidden {
			if strings.Contains(command, pattern) {
				fmt.Printf("  [Hook] BLOCKED: contains forbidden pattern '%s'\n", pattern)
				return clawde.BlockHook(fmt.Sprintf("Blocked: command contains forbidden pattern '%s'", pattern)), nil
			}
		}

		fmt.Println("  [Hook] Command allowed")
		return clawde.ContinueHook(), nil
	})

	client, err := clawde.NewClient(
		clawde.WithHook(clawde.HookPreToolUse, bashHook),
		clawde.WithAllowedTools("Bash"),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	// Test with safe command
	fmt.Println("Testing safe command...")
	stream, err := client.Query(ctx, "Run 'echo hello world' in bash")
	if err != nil {
		return err
	}
	drainStream(stream)

	return nil
}

// outputMonitorExample monitors all tool outputs.
func outputMonitorExample(ctx context.Context) error {
	// Create PostToolUse hook for all tools
	monitorHook := clawde.MatchAllTools(func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		fmt.Printf("  [Monitor] Tool '%s' completed\n", input.ToolName)

		// Check for errors in output
		output := string(input.ToolOutput)
		if strings.Contains(strings.ToLower(output), "error") {
			fmt.Println("  [Monitor] Warning: Tool output contains 'error'")
		}

		// Log output size
		fmt.Printf("  [Monitor] Output size: %d bytes\n", len(output))

		return clawde.ContinueHook(), nil
	})

	client, err := clawde.NewClient(
		clawde.WithHook(clawde.HookPostToolUse, monitorHook),
		clawde.WithAllowedTools("Bash", "Read"),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	stream, err := client.Query(ctx, "List files in current directory using ls")
	if err != nil {
		return err
	}
	drainStream(stream)

	return nil
}

// toolUseLoggerExample logs all tool uses for auditing.
func toolUseLoggerExample(ctx context.Context) error {
	var toolLog []string

	// Pre-tool hook to log start
	preHook := clawde.MatchAllTools(func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		entry := fmt.Sprintf("START: %s", input.ToolName)
		toolLog = append(toolLog, entry)
		fmt.Printf("  [Log] %s\n", entry)
		return clawde.ContinueHook(), nil
	})

	// Post-tool hook to log end
	postHook := clawde.MatchAllTools(func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		entry := fmt.Sprintf("END: %s", input.ToolName)
		toolLog = append(toolLog, entry)
		fmt.Printf("  [Log] %s\n", entry)
		return clawde.ContinueHook(), nil
	})

	client, err := clawde.NewClient(
		clawde.WithHook(clawde.HookPreToolUse, preHook),
		clawde.WithHook(clawde.HookPostToolUse, postHook),
		clawde.WithAllowedTools("Bash"),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	stream, err := client.Query(ctx, "Run 'pwd' in bash")
	if err != nil {
		return err
	}
	drainStream(stream)

	fmt.Printf("\n  Tool Log Summary: %d entries\n", len(toolLog))
	for _, entry := range toolLog {
		fmt.Printf("    - %s\n", entry)
	}

	return nil
}

// drainStream consumes all messages from a stream.
func drainStream(stream *clawde.Stream) {
	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			fmt.Print(m.Text())
		case *clawde.ResultMessage:
			fmt.Printf("\n--- Done (cost: $%.4f) ---\n", m.TotalCostUSD)
		}
	}
	if err := stream.Err(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
