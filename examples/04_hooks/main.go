// Example: Hooks
// Demonstrates using hooks to intercept and control tool usage.
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

	// Create a hook that blocks dangerous bash commands
	bashGuard := clawde.MatchTool("Bash", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		cmd := input.Command()
		fmt.Printf("[Hook] Checking command: %s\n", cmd)

		// Block dangerous commands
		dangerous := []string{"rm -rf", "mkfs", "dd if=", "> /dev/"}
		for _, d := range dangerous {
			if strings.Contains(cmd, d) {
				fmt.Printf("[Hook] BLOCKED: %s\n", d)
				return clawde.BlockHook("Dangerous command blocked: " + d), nil
			}
		}

		fmt.Println("[Hook] Command allowed")
		return clawde.ContinueHook(), nil
	})

	// Create a hook that logs all file operations
	fileLogger := clawde.MatchTool("Write", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		path := input.FilePath()
		fmt.Printf("[Hook] File write requested: %s\n", path)
		return clawde.ContinueHook(), nil
	})

	// Create a hook for all tools (logging)
	allToolsLogger := clawde.MatchAll(func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		fmt.Printf("[Hook] Tool called: %s\n", input.ToolName)
		return clawde.ContinueHook(), nil
	})

	client, err := clawde.NewClient(
		clawde.WithHook(clawde.HookPreToolUse, bashGuard),
		clawde.WithHook(clawde.HookPreToolUse, fileLogger),
		clawde.WithHook(clawde.HookPreToolUse, allToolsLogger),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Hooks configured. Querying Claude...")
	fmt.Println("---")

	stream, err := client.Query(ctx, "List the files in the current directory.")
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Print(msg.Text())
		}
	}
	fmt.Println()

	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}
}
