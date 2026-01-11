// Example: Custom Agents
// Demonstrates how to define and use custom agents with specific tools, prompts, and models.
// Equivalent to Python SDK's agents.py
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexo-tech/clawde"
)

func codeReviewerExample(ctx context.Context) {
	fmt.Println("=== Code Reviewer Agent Example ===")

	stream, err := clawde.Query(ctx, "Use the code-reviewer agent to review this code: func add(a, b int) int { return a + b }",
		clawde.WithAgents(map[string]clawde.AgentDefinition{
			"code-reviewer": {
				Description: "Reviews code for best practices and potential issues",
				Prompt: "You are a code reviewer. Analyze code for bugs, performance issues, " +
					"security vulnerabilities, and adherence to best practices. " +
					"Provide constructive feedback.",
				Tools: []string{"Read", "Grep"},
				Model: "sonnet",
			},
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		switch msg := stream.Current().(type) {
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

func documentationWriterExample(ctx context.Context) {
	fmt.Println("=== Documentation Writer Agent Example ===")

	stream, err := clawde.Query(ctx, "Use the doc-writer agent to explain what AgentDefinition is used for",
		clawde.WithAgents(map[string]clawde.AgentDefinition{
			"doc-writer": {
				Description: "Writes comprehensive documentation",
				Prompt: "You are a technical documentation expert. Write clear, comprehensive " +
					"documentation with examples. Focus on clarity and completeness.",
				Tools: []string{"Read", "Write", "Edit"},
				Model: "sonnet",
			},
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		switch msg := stream.Current().(type) {
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

func multipleAgentsExample(ctx context.Context) {
	fmt.Println("=== Multiple Agents Example ===")

	stream, err := clawde.Query(ctx, "Use the analyzer agent to find all Go files in the examples/ directory",
		clawde.WithAgents(map[string]clawde.AgentDefinition{
			"analyzer": {
				Description: "Analyzes code structure and patterns",
				Prompt:      "You are a code analyzer. Examine code structure, patterns, and architecture.",
				Tools:       []string{"Read", "Grep", "Glob"},
			},
			"tester": {
				Description: "Creates and runs tests",
				Prompt:      "You are a testing expert. Write comprehensive tests and ensure code quality.",
				Tools:       []string{"Read", "Write", "Bash"},
				Model:       "sonnet",
			},
		}),
		clawde.WithSettingSources(clawde.SettingSourceUser, clawde.SettingSourceProject),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		switch msg := stream.Current().(type) {
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

	codeReviewerExample(ctx)
	documentationWriterExample(ctx)
	multipleAgentsExample(ctx)
}
