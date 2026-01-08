// Package main demonstrates budget control with Clawde SDK.
// Shows how to set cost limits and track spending.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx := context.Background()

	// Example 1: Set maximum budget
	fmt.Println("=== Budget Control Demo ===")
	if err := budgetControlExample(ctx); err != nil {
		log.Printf("Budget control example failed: %v", err)
	}

	fmt.Println()

	// Example 2: Track costs across queries
	fmt.Println("=== Cost Tracking Demo ===")
	if err := costTrackingExample(ctx); err != nil {
		log.Printf("Cost tracking example failed: %v", err)
	}
}

// budgetControlExample shows how to set a maximum budget.
func budgetControlExample(ctx context.Context) error {
	// Set a very low budget for demonstration
	// In real usage, you'd set this to a reasonable limit
	budget := 0.10 // $0.10

	fmt.Printf("Setting maximum budget: $%.2f\n\n", budget)

	client, err := clawde.NewClient(
		clawde.WithMaxBudget(budget),
		clawde.WithSystemPrompt("Be very concise. Answer in one sentence."),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	// Send a query
	stream, err := client.Query(ctx, "What is the meaning of life?")
	if err != nil {
		return err
	}

	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			fmt.Print(m.Text())
		case *clawde.ResultMessage:
			fmt.Printf("\n\n--- Query Stats ---\n")
			fmt.Printf("Cost: $%.6f\n", m.TotalCostUSD)
			fmt.Printf("Remaining budget: $%.6f\n", budget-m.TotalCostUSD)
			fmt.Printf("Duration: %dms\n", m.DurationMS)
		}
	}

	if err := stream.Err(); err != nil {
		if clawde.IsBudgetExceeded(err) {
			fmt.Println("\n⚠️  Budget limit reached!")
		}
		return err
	}

	return nil
}

// costTrackingExample tracks costs across multiple queries.
func costTrackingExample(ctx context.Context) error {
	var totalCost float64
	var queryCount int

	client, err := clawde.NewClient(
		clawde.WithSystemPrompt("Be extremely concise. One word answers preferred."),
		clawde.WithMaxTurns(1),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	queries := []string{
		"What color is the sky?",
		"What is 2+2?",
		"Capital of France?",
	}

	fmt.Println("Sending multiple queries and tracking costs:\n")

	for i, query := range queries {
		queryCount++
		fmt.Printf("Query %d: %s\n", i+1, query)

		stream, err := client.Query(ctx, query)
		if err != nil {
			return err
		}

		var queryCost float64
		for stream.Next() {
			msg := stream.Current()
			switch m := msg.(type) {
			case *clawde.AssistantMessage:
				fmt.Printf("  Answer: %s\n", m.Text())
			case *clawde.ResultMessage:
				queryCost = m.TotalCostUSD
				totalCost += queryCost
			}
		}

		if err := stream.Err(); err != nil {
			return err
		}

		fmt.Printf("  Cost: $%.6f\n\n", queryCost)
	}

	// Summary
	fmt.Println("═══════════════════════════════════")
	fmt.Printf("Total queries: %d\n", queryCount)
	fmt.Printf("Total cost: $%.6f\n", totalCost)
	fmt.Printf("Average cost per query: $%.6f\n", totalCost/float64(queryCount))
	fmt.Println("═══════════════════════════════════")

	return nil
}
