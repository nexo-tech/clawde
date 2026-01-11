// Example: Events
// Demonstrates handling different types of events and messages.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx := context.Background()

	client, err := clawde.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Event Types Demo")
	fmt.Println("---")

	stream, err := client.Query(ctx, "What is 2+2? Use the calculator if needed.")
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		msg := stream.Current()

		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			fmt.Printf("[Assistant] Text: %s\n", m.Text())
			if thinking := m.Thinking(); thinking != "" {
				fmt.Printf("[Assistant] Thinking: %s...\n", truncate(thinking, 50))
			}
			for _, tu := range m.ToolUses() {
				fmt.Printf("[Assistant] Tool Use: %s (id=%s)\n", tu.Name, tu.ID)
			}

		case *clawde.UserMessage:
			fmt.Printf("[User] %s\n", m.Text())

		case *clawde.SystemMessage:
			fmt.Printf("[System] Type=%s Subtype=%s\n", m.Type, m.Subtype)

		case *clawde.ResultMessage:
			fmt.Println("---")
			fmt.Printf("[Result] Duration: %dms\n", m.DurationMS)
			fmt.Printf("[Result] Cost: $%.4f\n", m.CostUSD)
			fmt.Printf("[Result] Turns: %d\n", m.NumTurns)
			fmt.Printf("[Result] Session: %s\n", m.SessionID)

		case *clawde.StreamEvent:
			fmt.Printf("[Stream] Type=%s\n", m.Type)

		default:
			fmt.Printf("[Unknown] %T\n", msg)
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
