// Example: Human in the Loop
// Demonstrates using permission callbacks for human approval.
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

func main() {
	ctx := context.Background()

	// Create a permission callback that asks for user approval
	approvalCallback := func(ctx context.Context, req *clawde.PermissionRequest) clawde.PermissionResult {
		fmt.Printf("\n[Permission Request]\n")
		fmt.Printf("Tool: %s\n", req.ToolName)

		// Pretty print the input
		var input map[string]any
		if err := json.Unmarshal(req.Input, &input); err == nil {
			inputJSON, _ := json.MarshalIndent(input, "  ", "  ")
			fmt.Printf("Input:\n  %s\n", inputJSON)
		}

		fmt.Print("Allow this action? [y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			return clawde.Allow()
		}

		return clawde.Deny("User denied the action")
	}

	client, err := clawde.NewClient(
		clawde.WithPermissionCallback(approvalCallback),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Human-in-the-Loop Demo")
	fmt.Println("You will be asked to approve each tool use.")
	fmt.Println("---")

	stream, err := client.Query(ctx, "What files are in the current directory? Then tell me the time.")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\nClaude: ")
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
