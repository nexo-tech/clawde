// Example: Resume Generator
// Demonstrates web search and document generation with custom system prompts.
// Equivalent to TypeScript SDK's resume-generator/resume-generator.ts
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nexo-tech/clawde"
)

const systemPrompt = `You are a professional resume writer. Research a person and create a 1-page .docx resume.

WORKFLOW:
1. WebSearch for the person's background (LinkedIn, GitHub, company pages)
2. Create a .docx file using the docx library

OUTPUT:
- Script: agent/custom_scripts/generate_resume.js
- Resume: agent/custom_scripts/resume.docx

PAGE FIT (must be exactly 1 page):
- 0.5 inch margins, Name 24pt, Headers 12pt, Body 10pt
- 2-3 bullet points per job, ~80-100 chars each
- Max 3 job roles, 2-line summary, 2-line skills`

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go \"Person Name\"")
		fmt.Println("Example: go run main.go \"Jane Doe\"")
		os.Exit(1)
	}
	personName := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Printf("\nGenerating resume for: %s\n", personName)
	fmt.Println(repeatString("=", 50))

	// Ensure output directory exists
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	outputDir := filepath.Join(cwd, "agent", "custom_scripts")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Create client with settings
	client, err := clawde.NewClient(
		clawde.WithModel("sonnet"),
		clawde.WithMaxTurns(30),
		clawde.WithWorkingDir(cwd),
		clawde.WithAllowedTools(
			"Skill", "WebSearch", "WebFetch", "Bash", "Write", "Read", "Glob",
		),
		clawde.WithSettingSources(clawde.SettingSourceProject),
		clawde.WithSystemPrompt(systemPrompt),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Create prompt
	prompt := fmt.Sprintf(`Research "%s" and create a professional 1-page resume as a .docx file. Search for their professional background, experience, education, and skills.`, personName)

	fmt.Println("\nResearching and creating resume...")

	stream, err := client.Query(ctx, prompt)
	if err != nil {
		log.Fatal(err)
	}

	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			// Print text content
			text := m.Text()
			if text != "" {
				fmt.Println(text)
			}

			// Check for tool uses
			for _, tu := range m.ToolUses() {
				toolName := tu.Name
				if toolName == "WebSearch" {
					// Extract query from input
					var input struct {
						Query string `json:"query"`
					}
					if err := json.Unmarshal(tu.Input, &input); err == nil && input.Query != "" {
						fmt.Printf("\nSearching: \"%s\"\n", input.Query)
					}
				} else {
					fmt.Printf("\nUsing tool: %s\n", toolName)
				}
			}

		case *clawde.ResultMessage:
			// Show completion info
			if m.Subtype == "success" {
				fmt.Printf("\nCompleted in %dms, cost: $%.4f\n", m.DurationMS, m.TotalCostUSD)
			}
		}
	}

	if stream.Err() != nil {
		log.Printf("Error: %v", stream.Err())
	}

	// Check if resume was created
	expectedPath := filepath.Join(outputDir, "resume.docx")
	if _, err := os.Stat(expectedPath); err == nil {
		fmt.Println("\n" + repeatString("=", 50))
		fmt.Printf("Resume saved to: %s\n", expectedPath)
		fmt.Println(repeatString("=", 50))
	} else {
		fmt.Println("\nResume file was not created. Check the output above for errors.")
	}
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
