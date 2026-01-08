// Package main demonstrates human-in-the-loop patterns with Clawde SDK.
// Shows how to implement interactive approval for tool executions.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nexo-tech/clawde"
)

var scanner = bufio.NewScanner(os.Stdin)

func main() {
	ctx := context.Background()

	// Example 1: Interactive permission callback
	fmt.Println("=== Human-in-the-Loop Permission ===")
	if err := humanApprovalExample(ctx); err != nil {
		log.Printf("Human approval example failed: %v", err)
	}
}

// humanApprovalExample asks the user before executing sensitive tools.
func humanApprovalExample(ctx context.Context) error {
	// Create permission callback that asks user
	askUser := func(ctx context.Context, req *clawde.PermissionRequest) clawde.PermissionResult {
		fmt.Println()
		fmt.Println("┌────────────────────────────────────────┐")
		fmt.Println("│       Permission Request               │")
		fmt.Println("├────────────────────────────────────────┤")
		fmt.Printf("│ Tool: %-32s │\n", req.ToolName)
		fmt.Println("├────────────────────────────────────────┤")

		// Pretty print the input
		var inputMap map[string]any
		if err := json.Unmarshal(req.Input, &inputMap); err == nil {
			fmt.Println("│ Input:                                 │")
			for k, v := range inputMap {
				str := fmt.Sprintf("%v", v)
				if len(str) > 30 {
					str = str[:27] + "..."
				}
				fmt.Printf("│   %s: %-28s │\n", k, str)
			}
		} else {
			fmt.Printf("│ Input: %s\n", string(req.Input))
		}

		fmt.Println("└────────────────────────────────────────┘")
		fmt.Print("Allow this operation? (y/n/v): ")

		if !scanner.Scan() {
			return clawde.Deny("User input not available")
		}

		response := strings.ToLower(strings.TrimSpace(scanner.Text()))

		switch response {
		case "y", "yes":
			fmt.Println("✓ Approved")
			return clawde.Allow()

		case "v", "view":
			// Show full input
			prettyInput, _ := json.MarshalIndent(inputMap, "", "  ")
			fmt.Printf("\nFull input:\n%s\n\n", string(prettyInput))
			fmt.Print("Allow this operation? (y/n): ")

			if scanner.Scan() {
				if strings.ToLower(strings.TrimSpace(scanner.Text())) == "y" {
					fmt.Println("✓ Approved")
					return clawde.Allow()
				}
			}
			fmt.Println("✗ Denied")
			return clawde.Deny("User denied after viewing details")

		default:
			fmt.Println("✗ Denied")
			return clawde.Deny("User denied")
		}
	}

	client, err := clawde.NewClient(
		clawde.WithPermissionCallback(askUser),
		clawde.WithAllowedTools("Write", "Bash", "Read"),
		clawde.WithSystemPrompt("You are a helpful assistant. When asked to create files or run commands, do so."),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("This example will ask for your approval before executing tools.")
	fmt.Println("Options: y=yes, n=no, v=view full input")
	fmt.Println()

	// Test prompt that will trigger tool use
	prompts := []string{
		"Create a file called test_hello.txt with the content 'Hello, World!'",
		"Read the contents of go.mod if it exists",
	}

	for _, prompt := range prompts {
		fmt.Printf("\n>>> Sending: %s\n\n", prompt)

		stream, err := client.Query(ctx, prompt)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}

		for stream.Next() {
			msg := stream.Current()
			switch m := msg.(type) {
			case *clawde.AssistantMessage:
				text := m.Text()
				if text != "" {
					fmt.Print(text)
				}
			case *clawde.ResultMessage:
				fmt.Printf("\n\n[Cost: $%.4f]\n", m.TotalCostUSD)
			}
		}

		if err := stream.Err(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Println()
	}

	return nil
}
