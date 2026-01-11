// Example: Plugins
// Demonstrates how to load plugins with Claude Code SDK.
// Equivalent to Python SDK's plugin_example.py
package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/nexo-tech/clawde"
)

func pluginExample(ctx context.Context) {
	fmt.Println("=== Plugin Example ===")
	fmt.Println()

	// Get the path to the demo plugin
	// In production, you can use any path to your plugin directory
	pluginPath := filepath.Join("..", "..", "plugins", "demo-plugin")

	fmt.Printf("Loading plugin from: %s\n\n", pluginPath)

	stream, err := clawde.Query(ctx, "Hello!",
		clawde.WithPlugin("local", pluginPath),
		clawde.WithMaxTurns(1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	foundPlugins := false
	for stream.Next() {
		switch msg := stream.Current().(type) {
		case *clawde.SystemMessage:
			if msg.Subtype == "init" {
				fmt.Println("System initialized!")
				fmt.Println("Note: Plugin was passed via CLI")
				foundPlugins = true
			}
		case *clawde.AssistantMessage:
			fmt.Printf("Claude: %s\n", msg.Text())
		}
	}
	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}

	if foundPlugins {
		fmt.Println("\nPlugin successfully configured!")
	}
	fmt.Println()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pluginExample(ctx)
}
