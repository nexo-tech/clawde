// Example: Budget Control
// Demonstrates setting budget limits for queries.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx := context.Background()

	// Create client with budget limit
	client, err := clawde.NewClient(
		clawde.WithMaxBudget(0.10), // $0.10 limit
		clawde.WithMaxTurns(5),      // Max 5 agentic turns
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Budget Control Demo")
	fmt.Println("Max budget: $0.10")
	fmt.Println("Max turns: 5")
	fmt.Println("---")

	stream, err := client.Query(ctx, "Write a short poem about the ocean.")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Response: ")
	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Print(msg.Text())
		}

		// Check for result to see actual usage
		if result, ok := stream.Current().(*clawde.ResultMessage); ok {
			fmt.Println("\n---")
			fmt.Printf("Actual cost: $%.4f\n", result.CostUSD)
			fmt.Printf("Total cost: $%.4f\n", result.TotalCostUSD)
			fmt.Printf("Turns used: %d\n", result.NumTurns)
		}
	}
	fmt.Println()

	if err := stream.Err(); err != nil {
		// Check for budget exceeded error
		if errors.Is(err, clawde.ErrBudgetExceeded) {
			fmt.Println("Budget limit reached!")
		} else if errors.Is(err, clawde.ErrMaxTurnsExceeded) {
			fmt.Println("Turn limit reached!")
		} else {
			log.Fatal(err)
		}
	}
}
