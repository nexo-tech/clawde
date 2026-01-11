// Example: Streaming
// Demonstrates streaming responses with real-time output.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println("Starting streaming query...")
	fmt.Println("---")

	stream, err := clawde.Query(ctx, "Write a haiku about programming.")
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Stream responses in real-time
	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.AssistantMessage:
			// Print text as it arrives
			fmt.Print(msg.Text())

		case *clawde.ResultMessage:
			// Print final stats
			fmt.Println("\n---")
			fmt.Printf("Completed in %dms\n", msg.DurationMS)
			fmt.Printf("Cost: $%.4f\n", msg.CostUSD)
			fmt.Printf("Turns: %d\n", msg.NumTurns)
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatal("Stream error:", err)
	}
}
