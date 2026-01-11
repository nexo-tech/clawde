// Example: Research Agent
// Demonstrates multi-agent coordination with subagent tracking.
// Equivalent to TypeScript SDK's research-agent/
//
// This example creates a lead agent that coordinates multiple specialized subagents:
// - Researchers: Gather information via web search
// - Data Analyst: Generate charts and visualizations
// - Report Writer: Create PDF reports
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nexo-tech/clawde"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println()
	fmt.Println(repeatString("=", 50))
	fmt.Println("  Research Agent")
	fmt.Println(repeatString("=", 50))
	fmt.Println()
	fmt.Println("Research any topic and get a comprehensive PDF")
	fmt.Println("report with data visualizations.")
	fmt.Println()
	fmt.Println("Type 'exit' to quit.")
	fmt.Println()

	// Setup session directory
	sessionDir := filepath.Join("logs", fmt.Sprintf("session_%s", time.Now().Format("20060102_150405")))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Create transcript writer
	transcriptPath := filepath.Join(sessionDir, "transcript.txt")
	transcript, err := clawde.NewTranscriptWriter(transcriptPath)
	if err != nil {
		log.Fatal(err)
	}
	defer transcript.Close()

	// Create subagent tracker
	tracker, err := clawde.NewSubagentTracker(sessionDir)
	if err != nil {
		log.Fatal(err)
	}
	tracker.TranscriptWriter = os.Stdout
	defer tracker.Close()

	// Load prompts
	leadAgentPrompt, err := loadPrompt("prompts/lead_agent.txt")
	if err != nil {
		log.Fatal(err)
	}
	researcherPrompt, err := loadPrompt("prompts/researcher.txt")
	if err != nil {
		log.Fatal(err)
	}
	dataAnalystPrompt, err := loadPrompt("prompts/data_analyst.txt")
	if err != nil {
		log.Fatal(err)
	}
	reportWriterPrompt, err := loadPrompt("prompts/report_writer.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Define specialized subagents
	agents := map[string]clawde.AgentDefinition{
		"researcher": {
			Description: "Use this agent to gather research information on any topic using web search. " +
				"Writes research findings to files/research_notes/ for later use.",
			Tools:  []string{"WebSearch", "Write"},
			Prompt: researcherPrompt,
			Model:  "haiku",
		},
		"data-analyst": {
			Description: "Use this agent AFTER researchers complete to generate quantitative analysis " +
				"and visualizations. Reads research notes and generates charts using Python/matplotlib.",
			Tools:  []string{"Glob", "Read", "Bash", "Write"},
			Prompt: dataAnalystPrompt,
			Model:  "haiku",
		},
		"report-writer": {
			Description: "Use this agent to create formal research report documents. " +
				"Reads research findings and creates professionally formatted PDF reports.",
			Tools:  []string{"Skill", "Write", "Glob", "Read", "Bash"},
			Prompt: reportWriterPrompt,
			Model:  "haiku",
		},
	}

	// Create pre/post tool use hooks
	preToolHook := clawde.MatchTool(".*", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		return tracker.PreToolUseHook(ctx, input)
	})
	postToolHook := clawde.MatchTool(".*", func(ctx context.Context, input *clawde.HookInput) (*clawde.HookOutput, error) {
		return tracker.PostToolUseHook(ctx, input)
	})

	// Create client with lead agent configuration
	client, err := clawde.NewClient(
		clawde.WithModel("haiku"),
		clawde.WithPermissionMode(clawde.PermissionBypassAll),
		clawde.WithSettingSources(clawde.SettingSourceProject),
		clawde.WithSystemPrompt(leadAgentPrompt),
		clawde.WithAllowedTools("Task"),
		clawde.WithAgents(agents),
		clawde.WithHook(clawde.HookPreToolUse, preToolHook),
		clawde.WithHook(clawde.HookPostToolUse, postToolHook),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// REPL loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "exit" || strings.ToLower(input) == "quit" || strings.ToLower(input) == "q" {
			break
		}

		// Write to transcript
		transcript.WriteString(fmt.Sprintf("\nYou: %s\n", input))

		// Send query
		stream, err := client.Query(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Print("\nAgent: ")
		transcript.WriteString("\nAgent: ")

		for stream.Next() {
			msg := stream.Current()
			switch m := msg.(type) {
			case *clawde.AssistantMessage:
				text := m.Text()
				if text != "" {
					fmt.Print(text)
					transcript.WriteString(text)
				}
			}
		}

		fmt.Println()
		transcript.WriteString("\n")

		if stream.Err() != nil {
			fmt.Printf("Error: %v\n", stream.Err())
		}
	}

	fmt.Println("\nGoodbye!")
	transcript.WriteString("\n\nGoodbye!\n")

	fmt.Printf("\nSession logs saved to: %s\n", sessionDir)
	fmt.Printf("  - Transcript: %s\n", transcriptPath)
	fmt.Printf("  - Tool calls: %s\n", filepath.Join(sessionDir, "tool_calls.jsonl"))
}

func loadPrompt(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load prompt %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
