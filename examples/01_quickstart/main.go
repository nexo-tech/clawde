// Package main demonstrates the simplest way to query Claude using Clawde SDK.
// This example shows one-shot queries with streaming responses.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nexo-tech/clawde"
)

func main() {
	// Create cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	// Example 1: Simple query
	fmt.Println("=== Simple Query ===")
	if err := simpleQuery(ctx); err != nil {
		log.Printf("Simple query failed: %v", err)
	}

	fmt.Println()

	// Example 2: Query with options
	fmt.Println("=== Query With Options ===")
	if err := queryWithOptions(ctx); err != nil {
		log.Printf("Query with options failed: %v", err)
	}

	fmt.Println()

	// Example 3: Convenience function
	fmt.Println("=== QueryText Convenience Function ===")
	if err := queryTextExample(ctx); err != nil {
		log.Printf("QueryText failed: %v", err)
	}
}

// simpleQuery demonstrates the most basic usage pattern.
func simpleQuery(ctx context.Context) error {
	stream, err := clawde.Query(ctx, "What is 2 + 2? Answer in one word.")
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer stream.Close()

	// Stream responses
	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			fmt.Print(m.Text())
		case *clawde.ResultMessage:
			fmt.Printf("\n\nCost: $%.4f, Turns: %d\n", m.TotalCostUSD, m.NumTurns)
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("stream error: %w", err)
	}

	return nil
}

// queryWithOptions demonstrates using options to configure the query.
func queryWithOptions(ctx context.Context) error {
	stream, err := clawde.Query(ctx, "Explain what Go is.",
		clawde.WithSystemPrompt("You are a concise assistant. Answer in one sentence."),
		clawde.WithMaxTurns(1),
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	for stream.Next() {
		if am, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Print(am.Text())
		}
	}
	fmt.Println()

	return stream.Err()
}

// queryTextExample shows the QueryText convenience function.
func queryTextExample(ctx context.Context) error {
	text, err := clawde.QueryText(ctx, "What is the capital of France? Answer in one word.")
	if err != nil {
		return err
	}

	fmt.Printf("Answer: %s\n", text)
	return nil
}
