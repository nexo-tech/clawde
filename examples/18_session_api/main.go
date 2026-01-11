// Example: V2 Session API
// Demonstrates the session-based interface with send/stream pattern.
// Equivalent to TypeScript SDK's hello-world-v2/v2-examples.ts
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nexo-tech/clawde"
)

func main() {
	example := "basic"
	if len(os.Args) > 1 {
		example = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	switch example {
	case "basic":
		basicSession(ctx)
	case "multi-turn":
		multiTurn(ctx)
	case "one-shot":
		oneShot(ctx)
	case "resume":
		sessionResume(ctx)
	default:
		fmt.Println("Usage: go run main.go [basic|multi-turn|one-shot|resume]")
	}
}

// basicSession demonstrates basic session with send/stream pattern
func basicSession(ctx context.Context) {
	fmt.Println("=== Basic Session ===")
	fmt.Println()

	session, err := clawde.CreateSession(ctx, clawde.WithModel("sonnet"))
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	if err := session.Send(ctx, "Hello! Introduce yourself in one sentence."); err != nil {
		log.Fatal(err)
	}

	stream, err := session.Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			text := msg.Text()
			if text != "" {
				fmt.Printf("Claude: %s\n", text)
			}
		}
	}

	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
}

// multiTurn demonstrates multi-turn conversation - V2's key advantage
func multiTurn(ctx context.Context) {
	fmt.Println("=== Multi-Turn Conversation ===")
	fmt.Println()

	session, err := clawde.CreateSession(ctx, clawde.WithModel("sonnet"))
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Turn 1
	if err := session.Send(ctx, "What is 5 + 3? Just the number."); err != nil {
		log.Fatal(err)
	}

	stream1, err := session.Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for stream1.Next() {
		if msg, ok := stream1.Current().(*clawde.AssistantMessage); ok {
			text := msg.Text()
			if text != "" {
				fmt.Printf("Turn 1: %s\n", text)
			}
		}
	}

	// Turn 2 - Claude remembers context
	if err := session.Send(ctx, "Multiply that by 2. Just the number."); err != nil {
		log.Fatal(err)
	}

	stream2, err := session.Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for stream2.Next() {
		if msg, ok := stream2.Current().(*clawde.AssistantMessage); ok {
			text := msg.Text()
			if text != "" {
				fmt.Printf("Turn 2: %s\n", text)
			}
		}
	}
}

// oneShot demonstrates the one-shot convenience function
func oneShot(ctx context.Context) {
	fmt.Println("=== One-Shot Prompt ===")
	fmt.Println()

	result, err := clawde.Prompt(ctx, "What is the capital of France? One word.",
		clawde.WithModel("sonnet"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Answer: %s\n", result.Text)
	fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
}

// sessionResume demonstrates session persistence across sessions
func sessionResume(ctx context.Context) {
	fmt.Println("=== Session Resume ===")
	fmt.Println()

	var sessionID string

	// First session - establish a memory
	{
		session, err := clawde.CreateSession(ctx, clawde.WithModel("sonnet"))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("[Session 1] Telling Claude my favorite color...")

		if err := session.Send(ctx, "My favorite color is blue. Remember this!"); err != nil {
			session.Close()
			log.Fatal(err)
		}

		stream, err := session.Stream(ctx)
		if err != nil {
			session.Close()
			log.Fatal(err)
		}

		for stream.Next() {
			msg := stream.Current()
			switch m := msg.(type) {
			case *clawde.SystemMessage:
				if m.Subtype == "init" {
					fmt.Printf("[Session 1] Initialized\n")
				}
			case *clawde.AssistantMessage:
				text := m.Text()
				if text != "" {
					fmt.Printf("[Session 1] Claude: %s\n", text)
				}
			case *clawde.ResultMessage:
				if m.SessionID != "" {
					sessionID = m.SessionID
					fmt.Printf("[Session 1] ID: %s\n", sessionID)
				}
			}
		}

		session.Close()
	}

	fmt.Println("\n--- Session closed. Time passes... ---")

	// Resume and verify Claude remembers
	if sessionID == "" {
		fmt.Println("No session ID captured, skipping resume test")
		return
	}

	{
		session, err := clawde.ResumeSession(ctx, sessionID, clawde.WithModel("sonnet"))
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()

		fmt.Println("[Session 2] Resuming and asking Claude...")

		if err := session.Send(ctx, "What is my favorite color?"); err != nil {
			log.Fatal(err)
		}

		stream, err := session.Stream(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for stream.Next() {
			if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
				text := msg.Text()
				if text != "" {
					fmt.Printf("[Session 2] Claude: %s\n", text)
				}
			}
		}
	}
}
