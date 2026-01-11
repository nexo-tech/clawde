// Example: Quickstart
// Demonstrates the simplest usage of the Clawde SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx := context.Background()

	// Simplest usage: one-shot query
	text, err := clawde.QueryText(ctx, "What is 2+2? Reply with just the number.")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", text)

	// Alternative: streaming response
	stream, err := clawde.Query(ctx, "Tell me a short joke.")
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	fmt.Print("\nJoke: ")
	for stream.Next() {
		if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
			fmt.Print(msg.Text())
		}
	}
	fmt.Println()

	if stream.Err() != nil {
		fmt.Fprintln(os.Stderr, "Error:", stream.Err())
	}
}
