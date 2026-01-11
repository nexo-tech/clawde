// Example: Setting Sources
// Demonstrates how to control which settings are loaded.
// Equivalent to Python SDK's setting_sources.py
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexo-tech/clawde"
)

func exampleDefault(ctx context.Context) {
	fmt.Println("=== Default Behavior Example ===")
	fmt.Println("Setting sources: None (default)")
	fmt.Println("Expected: No custom slash commands will be available")
	fmt.Println()

	// No setting sources specified - isolated environment
	client, err := clawde.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	stream, err := client.Query(ctx, "What is 2 + 2?")
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.SystemMessage:
			if msg.Subtype == "init" {
				fmt.Printf("Init message received\n")
				fmt.Println("No setting sources loaded (default - isolated environment)")
			}
		case *clawde.AssistantMessage:
			fmt.Printf("Assistant: %s\n", truncate(msg.Text(), 100))
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func exampleUserOnly(ctx context.Context) {
	fmt.Println("=== User Settings Only Example ===")
	fmt.Println("Setting sources: ['user']")
	fmt.Println("Expected: Project slash commands will NOT be available")
	fmt.Println()

	client, err := clawde.NewClient(
		clawde.WithSettingSources(clawde.SettingSourceUser),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	stream, err := client.Query(ctx, "What is 2 + 2?")
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.SystemMessage:
			if msg.Subtype == "init" {
				fmt.Printf("Init message received with user settings only\n")
			}
		case *clawde.AssistantMessage:
			fmt.Printf("Assistant: %s\n", truncate(msg.Text(), 100))
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func exampleProjectAndUser(ctx context.Context) {
	fmt.Println("=== Project + User Settings Example ===")
	fmt.Println("Setting sources: ['user', 'project']")
	fmt.Println("Expected: Project slash commands WILL be available")
	fmt.Println()

	client, err := clawde.NewClient(
		clawde.WithSettingSources(clawde.SettingSourceUser, clawde.SettingSourceProject),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	stream, err := client.Query(ctx, "What is 2 + 2?")
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.SystemMessage:
			if msg.Subtype == "init" {
				fmt.Printf("Init message received with user + project settings\n")
			}
		case *clawde.AssistantMessage:
			fmt.Printf("Assistant: %s\n", truncate(msg.Text(), 100))
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Starting Claude SDK Setting Sources Examples...")
	fmt.Println("==================================================")
	fmt.Println()

	exampleDefault(ctx)
	fmt.Println("--------------------------------------------------")
	fmt.Println()

	exampleUserOnly(ctx)
	fmt.Println("--------------------------------------------------")
	fmt.Println()

	exampleProjectAndUser(ctx)
}
