// Example: Partial Message Streaming
// Demonstrates streaming partial messages for real-time UI updates.
// Equivalent to Python SDK's include_partial_messages.py
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Partial Message Streaming Example")
	fmt.Println("==================================================")

	client, err := clawde.NewClient(
		clawde.WithIncludePartialMessages(true),
		clawde.WithModel("claude-sonnet-4-5"),
		clawde.WithMaxTurns(2),
		clawde.WithEnv(map[string]string{
			"MAX_THINKING_TOKENS": "8000",
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Send a prompt that will generate a streaming response
	prompt := "Think of three jokes, then tell one"
	fmt.Printf("Prompt: %s\n", prompt)
	fmt.Println("==================================================")

	stream, err := client.Query(ctx, prompt)
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			fmt.Printf("[Assistant] %s\n", truncate(m.Text(), 100))
			if thinking := m.Thinking(); thinking != "" {
				fmt.Printf("[Thinking] %s...\n", truncate(thinking, 50))
			}
		case *clawde.StreamEvent:
			fmt.Printf("[StreamEvent] type=%s subtype=%s\n", m.Type, m.Subtype)
		case *clawde.SystemMessage:
			fmt.Printf("[System] type=%s subtype=%s\n", m.Type, m.Subtype)
		case *clawde.ResultMessage:
			fmt.Printf("[Result] cost=$%.4f turns=%d\n", m.CostUSD, m.NumTurns)
		default:
			fmt.Printf("[Message] %T\n", msg)
		}
	}

	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
