// Example: Stderr Callback
// Demonstrates capturing CLI debug output via a callback.
// Equivalent to Python SDK's stderr_callback_example.py
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Collect stderr messages
	var stderrMessages []string
	var mu sync.Mutex

	stderrCallback := func(line string) {
		mu.Lock()
		stderrMessages = append(stderrMessages, line)
		mu.Unlock()

		// Optionally print specific messages
		if strings.Contains(line, "[ERROR]") {
			fmt.Printf("Error detected: %s\n", line)
		}
	}

	fmt.Println("Running query with stderr capture...")

	stream, err := clawde.Query(ctx, "What is 2+2?",
		clawde.WithStderrCallback(stderrCallback),
		clawde.WithExtraArg("debug-to-stderr", ""), // Enable debug output (empty value = flag only)
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Printf("Response: %s\n", msg.Text())
		}
	}
	if stream.Err() != nil {
		log.Printf("Stream error: %v", stream.Err())
	}

	// Show what we captured
	mu.Lock()
	count := len(stderrMessages)
	var firstLine string
	if count > 0 {
		firstLine = stderrMessages[0]
		if len(firstLine) > 100 {
			firstLine = firstLine[:100]
		}
	}
	mu.Unlock()

	fmt.Printf("\nCaptured %d stderr lines\n", count)
	if firstLine != "" {
		fmt.Printf("First stderr line: %s\n", firstLine)
	}
}
