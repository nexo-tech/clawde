// Package main demonstrates event-driven patterns with Clawde SDK.
// Shows how to use hooks for event monitoring and metrics collection.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nexo-tech/clawde"
)

// EventCollector collects and tracks events during execution.
type EventCollector struct {
	mu     sync.Mutex
	events []Event
	start  time.Time
}

// Event represents a tracked event.
type Event struct {
	Time     time.Duration
	Type     string
	ToolName string
	Details  string
}

// NewEventCollector creates a new event collector.
func NewEventCollector() *EventCollector {
	return &EventCollector{
		events: make([]Event, 0),
		start:  time.Now(),
	}
}

// Add adds an event to the collector.
func (c *EventCollector) Add(eventType, toolName, details string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.events = append(c.events, Event{
		Time:     time.Since(c.start),
		Type:     eventType,
		ToolName: toolName,
		Details:  details,
	})
}

// Summary prints a summary of all events.
func (c *EventCollector) Summary() {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Event Summary                           ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")

	if len(c.events) == 0 {
		fmt.Println("║  No events recorded                                       ║")
	} else {
		for _, e := range c.events {
			fmt.Printf("║  [%8s] %-8s %-12s %s\n",
				e.Time.Round(time.Millisecond),
				e.Type,
				e.ToolName,
				truncate(e.Details, 20))
		}
	}

	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Printf("\nTotal events: %d\n", len(c.events))
	fmt.Printf("Total duration: %s\n", time.Since(c.start).Round(time.Millisecond))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func main() {
	ctx := context.Background()

	// Create event collector
	collector := NewEventCollector()

	// Create hooks for event tracking
	preToolHook := clawde.MatchAllTools(func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		collector.Add("START", input.ToolName, fmt.Sprintf("input: %d bytes", len(input.ToolInput)))
		fmt.Printf("  ⏵ Starting %s\n", input.ToolName)
		return clawde.ContinueHook(), nil
	})

	postToolHook := clawde.MatchAllTools(func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		collector.Add("END", input.ToolName, fmt.Sprintf("output: %d bytes", len(input.ToolOutput)))
		fmt.Printf("  ⏹ Finished %s\n", input.ToolName)
		return clawde.ContinueHook(), nil
	})

	// Create client with event hooks
	client, err := clawde.NewClient(
		clawde.WithHook(clawde.HookPreToolUse, preToolHook),
		clawde.WithHook(clawde.HookPostToolUse, postToolHook),
		clawde.WithAllowedTools("Bash", "Read", "Glob"),
		clawde.WithSystemPrompt("You are a helpful assistant. Execute commands when asked."),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Connect
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=== Event Tracking Demo ===")
	fmt.Println("Events will be tracked as tools are used.\n")

	// Test queries that will trigger tools
	queries := []string{
		"List the Go files in the current directory using glob",
		"Show me the current working directory",
	}

	for i, query := range queries {
		collector.Add("QUERY", "user", query)
		fmt.Printf("\n>>> Query %d: %s\n\n", i+1, query)

		stream, err := client.Query(ctx, query)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}

		for stream.Next() {
			msg := stream.Current()
			switch m := msg.(type) {
			case *clawde.AssistantMessage:
				text := m.Text()
				if text != "" {
					fmt.Print(text)
				}
			case *clawde.ResultMessage:
				collector.Add("RESULT", "session", fmt.Sprintf("$%.4f", m.TotalCostUSD))
			}
		}

		if err := stream.Err(); err != nil {
			collector.Add("ERROR", "stream", err.Error())
		}

		fmt.Println()
	}

	// Show event summary
	collector.Summary()
}
