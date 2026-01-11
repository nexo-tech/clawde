// Example: System Prompt Configurations
// Demonstrates different ways to configure the system prompt.
// Equivalent to Python SDK's system_prompt.py
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexo-tech/clawde"
)

func noSystemPrompt(ctx context.Context) {
	fmt.Println("=== No System Prompt (Vanilla Claude) ===")

	stream, err := clawde.Query(ctx, "What is 2 + 2?")
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Printf("Claude: %s\n", msg.Text())
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func stringSystemPrompt(ctx context.Context) {
	fmt.Println("=== String System Prompt ===")

	stream, err := clawde.Query(ctx, "What is 2 + 2?",
		clawde.WithSystemPrompt("You are a pirate assistant. Respond in pirate speak."),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Printf("Claude: %s\n", msg.Text())
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func presetSystemPrompt(ctx context.Context) {
	fmt.Println("=== Preset System Prompt (Default) ===")

	stream, err := clawde.Query(ctx, "What is 2 + 2?",
		clawde.WithSystemPromptPreset("claude_code", ""),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Printf("Claude: %s\n", msg.Text())
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func presetWithAppend(ctx context.Context) {
	fmt.Println("=== Preset System Prompt with Append ===")

	stream, err := clawde.Query(ctx, "What is 2 + 2?",
		clawde.WithSystemPromptPreset("claude_code", "Always end your response with a fun fact."),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Printf("Claude: %s\n", msg.Text())
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}
	fmt.Println()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	noSystemPrompt(ctx)
	stringSystemPrompt(ctx)
	presetSystemPrompt(ctx)
	presetWithAppend(ctx)
}
