// Example: Interactive
// Demonstrates a REPL-style interactive session with Claude.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx := context.Background()

	client, err := clawde.NewClient(
		clawde.WithSystemPrompt("You are a helpful assistant. Keep responses concise."),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Interactive Claude Session")
	fmt.Println("Type 'quit' to exit")
	fmt.Println("---")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nYou: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Send and stream response
		stream, err := client.Query(ctx, input)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Print("Claude: ")
		for stream.Next() {
			if msg, ok := stream.Current().(*clawde.AssistantMessage); ok {
				fmt.Print(msg.Text())
			}
		}
		fmt.Println()

		if stream.Err() != nil {
			fmt.Println("Error:", stream.Err())
		}
	}
}
