// Example: MCP Tools
// Demonstrates creating custom tools using the MCP system.
package main

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/nexo-tech/clawde"
)

// Calculator tool input types
type AddInput struct {
	A float64 `json:"a" description:"First number"`
	B float64 `json:"b" description:"Second number"`
}

type MultiplyInput struct {
	A float64 `json:"a" description:"First number"`
	B float64 `json:"b" description:"Second number"`
}

type SqrtInput struct {
	N float64 `json:"n" description:"Number to take square root of"`
}

func main() {
	ctx := context.Background()

	// Create an MCP server with calculator tools
	server := clawde.NewMCPServer("calculator")

	// Add tools using the typed helper
	server.Tools = append(server.Tools,
		clawde.Tool("add", "Add two numbers", func(ctx context.Context, in AddInput) (string, error) {
			return fmt.Sprintf("%.2f", in.A+in.B), nil
		}),
		clawde.Tool("multiply", "Multiply two numbers", func(ctx context.Context, in MultiplyInput) (string, error) {
			return fmt.Sprintf("%.2f", in.A*in.B), nil
		}),
		clawde.Tool("sqrt", "Square root of a number", func(ctx context.Context, in SqrtInput) (string, error) {
			if in.N < 0 {
				return "", fmt.Errorf("cannot take square root of negative number")
			}
			return fmt.Sprintf("%.4f", math.Sqrt(in.N)), nil
		}),
	)

	// Create client with the MCP server
	client, err := clawde.NewClient(
		clawde.WithSDKServer("calculator", server),
		clawde.WithSystemPrompt("You have access to a calculator. Use the tools to help with math."),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	// Ask a math question
	stream, err := client.Query(ctx, "What is the square root of (25 + 75)?")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Question: What is the square root of (25 + 75)?")
	fmt.Print("Answer: ")

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
