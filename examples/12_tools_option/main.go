// Example: Tools Configuration
// Demonstrates different ways to configure available tools.
// Equivalent to Python SDK's tools_option.py
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexo-tech/clawde"
)

func toolsArrayExample(ctx context.Context) {
	fmt.Println("=== Tools Array Example ===")
	fmt.Println("Setting tools=['Read', 'Glob', 'Grep']")
	fmt.Println()

	stream, err := clawde.Query(ctx, "What tools do you have available? Just list them briefly.",
		clawde.WithToolsList("Read", "Glob", "Grep"),
		clawde.WithMaxTurns(1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.SystemMessage:
			if msg.Subtype == "init" {
				fmt.Printf("System init message received\n")
			}
		case *clawde.AssistantMessage:
			fmt.Printf("Claude: %s\n", msg.Text())
		case *clawde.ResultMessage:
			if msg.CostUSD > 0 {
				fmt.Printf("\nCost: $%.4f\n", msg.CostUSD)
			}
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func toolsEmptyArrayExample(ctx context.Context) {
	fmt.Println("=== Tools Empty Array Example ===")
	fmt.Println("Setting tools=[] (disables all built-in tools)")
	fmt.Println()

	stream, err := clawde.Query(ctx, "What tools do you have available? Just list them briefly.",
		clawde.WithTools(&clawde.ToolsConfig{Tools: []string{}}), // Empty array
		clawde.WithMaxTurns(1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.SystemMessage:
			if msg.Subtype == "init" {
				fmt.Printf("System init message received\n")
			}
		case *clawde.AssistantMessage:
			fmt.Printf("Claude: %s\n", msg.Text())
		case *clawde.ResultMessage:
			if msg.CostUSD > 0 {
				fmt.Printf("\nCost: $%.4f\n", msg.CostUSD)
			}
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func toolsPresetExample(ctx context.Context) {
	fmt.Println("=== Tools Preset Example ===")
	fmt.Println("Setting tools={'type': 'preset', 'preset': 'claude_code'}")
	fmt.Println()

	stream, err := clawde.Query(ctx, "What tools do you have available? Just list them briefly.",
		clawde.WithToolsPreset("claude_code"),
		clawde.WithMaxTurns(1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.SystemMessage:
			if msg.Subtype == "init" {
				fmt.Printf("System init message received\n")
			}
		case *clawde.AssistantMessage:
			fmt.Printf("Claude: %s\n", msg.Text())
		case *clawde.ResultMessage:
			if msg.CostUSD > 0 {
				fmt.Printf("\nCost: $%.4f\n", msg.CostUSD)
			}
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	toolsArrayExample(ctx)
	toolsEmptyArrayExample(ctx)
	toolsPresetExample(ctx)
}
