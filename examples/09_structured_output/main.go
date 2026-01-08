// Package main demonstrates structured output with Clawde SDK.
// Shows how to get JSON-formatted responses using output schemas.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nexo-tech/clawde"
)

// WeatherResponse defines the expected structure for weather data.
type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	Unit        string  `json:"unit"`
	Conditions  string  `json:"conditions"`
	Forecast    string  `json:"forecast"`
}

// CodeReviewResponse defines the expected structure for code review.
type CodeReviewResponse struct {
	Summary     string   `json:"summary"`
	Issues      []Issue  `json:"issues"`
	Score       int      `json:"score"`
	Suggestions []string `json:"suggestions"`
}

// Issue represents a code review issue.
type Issue struct {
	Line     int    `json:"line"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func main() {
	ctx := context.Background()

	// Example 1: Weather response (simulated)
	fmt.Println("=== Structured Output: Weather ===")
	if err := weatherExample(ctx); err != nil {
		log.Printf("Weather example failed: %v", err)
	}

	fmt.Println()

	// Example 2: Code review response
	fmt.Println("=== Structured Output: Analysis ===")
	if err := analysisExample(ctx); err != nil {
		log.Printf("Analysis example failed: %v", err)
	}
}

// weatherExample demonstrates getting structured weather data.
func weatherExample(ctx context.Context) error {
	// Define the JSON schema
	schema := `{
		"type": "object",
		"properties": {
			"location": {"type": "string", "description": "City name"},
			"temperature": {"type": "number", "description": "Temperature value"},
			"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]},
			"conditions": {"type": "string", "description": "Weather conditions"},
			"forecast": {"type": "string", "description": "Brief forecast"}
		},
		"required": ["location", "temperature", "unit", "conditions"]
	}`

	client, err := clawde.NewClient(
		clawde.WithOutputFormat(json.RawMessage(schema)),
		clawde.WithSystemPrompt("You provide weather information. Make up realistic data for demonstration."),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	stream, err := client.Query(ctx, "What's the weather like in Tokyo today?")
	if err != nil {
		return err
	}

	var resultData json.RawMessage
	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			fmt.Print(m.Text())
		case *clawde.ResultMessage:
			if len(m.StructuredOutput) > 0 {
				resultData = m.StructuredOutput
			}
		}
	}

	if err := stream.Err(); err != nil {
		return err
	}

	// Parse structured output
	if len(resultData) > 0 {
		var weather WeatherResponse
		if err := json.Unmarshal(resultData, &weather); err != nil {
			fmt.Printf("\nCould not parse structured output: %v\n", err)
			fmt.Printf("Raw output: %s\n", string(resultData))
		} else {
			fmt.Println("\n\n--- Parsed Structured Data ---")
			fmt.Printf("Location: %s\n", weather.Location)
			fmt.Printf("Temperature: %.1f %s\n", weather.Temperature, weather.Unit)
			fmt.Printf("Conditions: %s\n", weather.Conditions)
			if weather.Forecast != "" {
				fmt.Printf("Forecast: %s\n", weather.Forecast)
			}
		}
	}

	return nil
}

// analysisExample demonstrates getting structured analysis data.
func analysisExample(ctx context.Context) error {
	// Define the JSON schema for analysis
	schema := `{
		"type": "object",
		"properties": {
			"summary": {"type": "string", "description": "Brief summary"},
			"score": {"type": "integer", "minimum": 0, "maximum": 100},
			"issues": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"line": {"type": "integer"},
						"severity": {"type": "string", "enum": ["low", "medium", "high"]},
						"message": {"type": "string"}
					}
				}
			},
			"suggestions": {
				"type": "array",
				"items": {"type": "string"}
			}
		},
		"required": ["summary", "score"]
	}`

	client, err := clawde.NewClient(
		clawde.WithOutputFormat(json.RawMessage(schema)),
		clawde.WithSystemPrompt("You are a code reviewer. Analyze code and provide structured feedback."),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	// Sample code to review
	sampleCode := `
func calculateSum(numbers []int) int {
    sum := 0
    for i := 0; i <= len(numbers); i++ {
        sum += numbers[i]
    }
    return sum
}
`

	query := fmt.Sprintf("Review this Go code for issues:\n```go%s```", sampleCode)
	stream, err := client.Query(ctx, query)
	if err != nil {
		return err
	}

	var resultData json.RawMessage
	for stream.Next() {
		msg := stream.Current()
		switch m := msg.(type) {
		case *clawde.AssistantMessage:
			fmt.Print(m.Text())
		case *clawde.ResultMessage:
			if len(m.StructuredOutput) > 0 {
				resultData = m.StructuredOutput
			}
		}
	}

	if err := stream.Err(); err != nil {
		return err
	}

	// Parse structured output
	if len(resultData) > 0 {
		var review CodeReviewResponse
		if err := json.Unmarshal(resultData, &review); err != nil {
			fmt.Printf("\nCould not parse structured output: %v\n", err)
		} else {
			fmt.Println("\n\n--- Code Review Results ---")
			fmt.Printf("Summary: %s\n", review.Summary)
			fmt.Printf("Score: %d/100\n", review.Score)

			if len(review.Issues) > 0 {
				fmt.Println("\nIssues Found:")
				for i, issue := range review.Issues {
					fmt.Printf("  %d. [%s] Line %d: %s\n", i+1, issue.Severity, issue.Line, issue.Message)
				}
			}

			if len(review.Suggestions) > 0 {
				fmt.Println("\nSuggestions:")
				for i, suggestion := range review.Suggestions {
					fmt.Printf("  %d. %s\n", i+1, suggestion)
				}
			}
		}
	}

	return nil
}
