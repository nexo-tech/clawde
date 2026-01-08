// Package main demonstrates creating custom MCP tools with Clawde SDK.
// Shows how to define, register, and use in-process tools.
package main

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/nexo-tech/clawde"
)

// CalculatorInput defines the input for basic calculator operations.
type CalculatorInput struct {
	A float64 `json:"a" description:"First number"`
	B float64 `json:"b" description:"Second number"`
}

// PowerInput defines the input for power operation.
type PowerInput struct {
	Base     float64 `json:"base" description:"Base number"`
	Exponent float64 `json:"exponent" description:"Exponent"`
}

// SqrtInput defines the input for square root operation.
type SqrtInput struct {
	N float64 `json:"n" description:"Number to find square root of"`
}

func main() {
	ctx := context.Background()

	// Create MCP server with calculator tools
	server := clawde.NewMCPServer("calculator")

	// Add tools using the type-safe Tool helper
	server.AddTool(clawde.Tool("add", "Add two numbers", func(ctx context.Context, input CalculatorInput) (string, error) {
		result := input.A + input.B
		return fmt.Sprintf("%.2f + %.2f = %.2f", input.A, input.B, result), nil
	}))

	server.AddTool(clawde.Tool("subtract", "Subtract two numbers", func(ctx context.Context, input CalculatorInput) (string, error) {
		result := input.A - input.B
		return fmt.Sprintf("%.2f - %.2f = %.2f", input.A, input.B, result), nil
	}))

	server.AddTool(clawde.Tool("multiply", "Multiply two numbers", func(ctx context.Context, input CalculatorInput) (string, error) {
		result := input.A * input.B
		return fmt.Sprintf("%.2f * %.2f = %.2f", input.A, input.B, result), nil
	}))

	server.AddTool(clawde.Tool("divide", "Divide two numbers", func(ctx context.Context, input CalculatorInput) (string, error) {
		if input.B == 0 {
			return "", fmt.Errorf("division by zero")
		}
		result := input.A / input.B
		return fmt.Sprintf("%.2f / %.2f = %.2f", input.A, input.B, result), nil
	}))

	server.AddTool(clawde.Tool("power", "Raise base to exponent", func(ctx context.Context, input PowerInput) (string, error) {
		result := math.Pow(input.Base, input.Exponent)
		return fmt.Sprintf("%.2f ^ %.2f = %.2f", input.Base, input.Exponent, result), nil
	}))

	server.AddTool(clawde.Tool("sqrt", "Square root of a number", func(ctx context.Context, input SqrtInput) (string, error) {
		if input.N < 0 {
			return "", fmt.Errorf("cannot take square root of negative number")
		}
		result := math.Sqrt(input.N)
		return fmt.Sprintf("sqrt(%.2f) = %.2f", input.N, result), nil
	}))

	// Create client with the MCP server
	client, err := clawde.NewClient(
		clawde.WithSDKMCPServer(server),
		clawde.WithAllowedTools(
			"mcp__calculator__add",
			"mcp__calculator__subtract",
			"mcp__calculator__multiply",
			"mcp__calculator__divide",
			"mcp__calculator__power",
			"mcp__calculator__sqrt",
		),
		clawde.WithSystemPrompt("You are a calculator assistant. Use the calculator tools to perform calculations."),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Connect
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Test prompts
	prompts := []string{
		"What calculator tools do you have available?",
		"Calculate 15 + 27",
		"What is the square root of 144?",
		"Calculate 2 to the power of 10",
		"What is 100 divided by 4?",
	}

	for _, prompt := range prompts {
		fmt.Printf("\n>>> %s\n", prompt)

		stream, err := client.Query(ctx, prompt)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}

		for stream.Next() {
			msg := stream.Current()
			switch m := msg.(type) {
			case *clawde.AssistantMessage:
				fmt.Print(m.Text())
				// Show tool uses
				for _, tu := range m.ToolUses() {
					fmt.Printf("\n[Using tool: %s]\n", tu.Name)
				}
			case *clawde.ResultMessage:
				fmt.Printf("\n--- Cost: $%.4f ---\n", m.TotalCostUSD)
			}
		}

		if err := stream.Err(); err != nil {
			log.Printf("Stream error: %v", err)
		}
	}
}
