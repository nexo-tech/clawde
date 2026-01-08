// Package main demonstrates an interactive REPL session with Clawde SDK.
// Users can have multi-turn conversations with persistent context.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nexo-tech/clawde"
)

func main() {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nGoodbye!")
		cancel()
		os.Exit(0)
	}()

	// Create interactive client
	client, err := clawde.NewClient(
		clawde.WithSystemPrompt("You are a helpful coding assistant. Be concise but thorough."),
		clawde.WithAllowedTools("Read", "Bash", "Glob", "Grep"),
		clawde.WithContinueSession(), // Continue from previous session if available
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Connect
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Println("║     Clawde Interactive Session        ║")
	fmt.Println("╠═══════════════════════════════════════╣")
	fmt.Println("║  Commands:                            ║")
	fmt.Println("║    /quit   - Exit the session         ║")
	fmt.Println("║    /clear  - Clear conversation       ║")
	fmt.Println("║    /cost   - Show session cost        ║")
	fmt.Println("║    /help   - Show this help           ║")
	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	var totalCost float64
	var totalTurns int

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle commands
		switch strings.ToLower(input) {
		case "/quit", "/exit", "/q":
			fmt.Println("Goodbye!")
			return

		case "/clear":
			// In a real implementation, you'd create a new session
			fmt.Println("Note: Session clearing requires reconnecting.")
			fmt.Println("Use /quit and restart for a fresh session.")
			continue

		case "/cost":
			fmt.Printf("Session cost: $%.4f over %d turns\n", totalCost, totalTurns)
			continue

		case "/help":
			fmt.Println("Commands: /quit, /clear, /cost, /help")
			continue
		}

		// Skip commands starting with /
		if strings.HasPrefix(input, "/") {
			fmt.Println("Unknown command. Type /help for available commands.")
			continue
		}

		// Send message
		stream, err := client.Query(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Stream response
		fmt.Print("Claude: ")
		for stream.Next() {
			msg := stream.Current()
			switch m := msg.(type) {
			case *clawde.AssistantMessage:
				fmt.Print(m.Text())

			case *clawde.ResultMessage:
				totalCost += m.TotalCostUSD
				totalTurns += m.NumTurns
			}
		}

		if err := stream.Err(); err != nil {
			fmt.Printf("\nError: %v\n", err)
		}
		fmt.Println()
		fmt.Println()
	}
}
