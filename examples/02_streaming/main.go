// Package main demonstrates streaming patterns with Clawde SDK.
// Shows how to process messages as they arrive and collect results.
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

	// Example 1: Stream and print in real-time
	fmt.Println("=== Real-time Streaming ===")
	if err := realtimeStreaming(ctx); err != nil {
		log.Printf("Realtime streaming failed: %v", err)
	}

	fmt.Println()

	// Example 2: Collect all messages
	fmt.Println("=== Collect All Messages ===")
	if err := collectMessages(ctx); err != nil {
		log.Printf("Collect messages failed: %v", err)
	}

	fmt.Println()

	// Example 3: Process different message types
	fmt.Println("=== Process Message Types ===")
	if err := processMessageTypes(ctx); err != nil {
		log.Printf("Process message types failed: %v", err)
	}
}

// realtimeStreaming prints text as it arrives.
func realtimeStreaming(ctx context.Context) error {
	stream, err := clawde.Query(ctx, "Count from 1 to 5, with a brief pause between each number.")
	if err != nil {
		return err
	}
	defer stream.Close()

	fmt.Print("Response: ")
	for stream.Next() {
		if am, ok := stream.Current().(*clawde.AssistantMessage); ok {
			// Print text immediately as it arrives
			fmt.Print(am.Text())
		}
	}
	fmt.Println()

	return stream.Err()
}

// collectMessages collects all messages and processes at the end.
func collectMessages(ctx context.Context) error {
	stream, err := clawde.Query(ctx, "List three programming languages.")
	if err != nil {
		return err
	}
	defer stream.Close()

	// Collect all messages
	messages, err := stream.Collect()
	if err != nil {
		return err
	}

	fmt.Printf("Received %d messages:\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  %d. Type: %s\n", i+1, msg.MessageType())
	}

	// Get accumulated text
	if stream.Message() != nil {
		fmt.Printf("\nFull response:\n%s\n", stream.Message().Text())
	}

	return nil
}

// processMessageTypes shows how to handle different message types.
func processMessageTypes(ctx context.Context) error {
	stream, err := clawde.Query(ctx,
		"What's the weather like? (Just pretend and give a brief answer)",
		clawde.WithAllowedTools("Read"), // Enable some tools
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	var textParts []string
	var thinkingParts []string
	var toolUses []string

	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.UserMessage:
			fmt.Println("[User message received]")

		case *clawde.AssistantMessage:
			// Collect text
			textParts = append(textParts, m.Text())

			// Check for thinking
			if thinking := m.Thinking(); thinking != "" {
				thinkingParts = append(thinkingParts, thinking)
			}

			// Check for tool uses
			for _, tu := range m.ToolUses() {
				toolUses = append(toolUses, tu.Name)
			}

		case *clawde.SystemMessage:
			fmt.Printf("[System: %s]\n", m.Subtype)

		case *clawde.ResultMessage:
			fmt.Printf("\n--- Result ---\n")
			fmt.Printf("Duration: %dms\n", m.DurationMS)
			fmt.Printf("Cost: $%.4f\n", m.TotalCostUSD)
			fmt.Printf("Turns: %d\n", m.NumTurns)
			if m.IsError {
				fmt.Println("Status: Error")
			}
		}
	}

	if len(textParts) > 0 {
		fmt.Printf("\nText: %s\n", strings.Join(textParts, ""))
	}
	if len(thinkingParts) > 0 {
		fmt.Printf("Thinking: %s\n", strings.Join(thinkingParts, " "))
	}
	if len(toolUses) > 0 {
		fmt.Printf("Tools used: %v\n", toolUses)
	}

	return stream.Err()
}
