// Example: Hello World (TypeScript SDK Parity)
// Demonstrates basic query with hooks for path validation.
// Equivalent to TypeScript SDK's hello-world/hello-world.ts
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Hello World Example (TypeScript SDK Parity)")
	fmt.Println("============================================")

	// Get working directory for path validation
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	customScriptsPath := filepath.Join(cwd, "agent", "custom_scripts")

	// Ensure output directory exists
	if err := os.MkdirAll(customScriptsPath, 0755); err != nil {
		log.Fatal(err)
	}

	// Create hook for path validation (same as TypeScript version)
	pathValidationHook := clawde.MatchTool("Write|Edit|MultiEdit", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		toolName := input.ToolName

		// Only validate Write, Edit, MultiEdit
		if toolName != "Write" && toolName != "Edit" && toolName != "MultiEdit" {
			return clawde.ContinueHook(), nil
		}

		// Extract file_path from tool input
		filePath := input.FilePath()

		// Check if it's a script file
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == ".js" || ext == ".ts" {
			// Script files must be in custom_scripts directory
			if !strings.HasPrefix(filePath, customScriptsPath) {
				return &clawde.HookOutput{
					Decision:   "block",
					StopReason: fmt.Sprintf("Script files (.js and .ts) must be written to the custom_scripts directory. Please use the path: %s/%s", customScriptsPath, filepath.Base(filePath)),
					Continue:   false,
				}, nil
			}
		}

		return clawde.ContinueHook(), nil
	})

	// Create client with allowed tools and hook
	client, err := clawde.NewClient(
		clawde.WithModel("opus"),
		clawde.WithMaxTurns(100),
		clawde.WithWorkingDir(filepath.Join(cwd, "agent")),
		clawde.WithAllowedTools(
			"Task", "Bash", "Glob", "Grep", "LS", "ExitPlanMode", "Read", "Edit", "MultiEdit", "Write", "NotebookEdit",
			"WebFetch", "TodoWrite", "WebSearch", "BashOutput", "KillBash",
		),
		clawde.WithHook(clawde.HookPreToolUse, pathValidationHook),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Send the prompt
	prompt := "Hello, Claude! Please introduce yourself in one sentence."
	fmt.Printf("\nPrompt: %s\n\n", prompt)

	stream, err := client.Query(ctx, prompt)
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		msg := stream.Current()
		if assistant, ok := msg.(*clawde.AssistantMessage); ok {
			text := assistant.Text()
			if text != "" {
				fmt.Printf("Claude says: %s\n", text)
			}
		}
	}

	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
}
