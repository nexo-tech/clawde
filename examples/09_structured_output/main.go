// Example: Structured Output
// Demonstrates parsing structured data from Claude's responses.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nexo-tech/clawde"
)

// Recipe represents a structured recipe
type Recipe struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
	PrepTime    string   `json:"prep_time"`
	CookTime    string   `json:"cook_time"`
}

func main() {
	ctx := context.Background()

	client, err := clawde.NewClient(
		clawde.WithSystemPrompt(`You are a helpful cooking assistant.
When asked for recipes, always respond with valid JSON in this format:
{
  "name": "Recipe Name",
  "ingredients": ["ingredient 1", "ingredient 2"],
  "steps": ["step 1", "step 2"],
  "prep_time": "10 minutes",
  "cook_time": "20 minutes"
}`),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Structured Output Demo")
	fmt.Println("---")

	stream, err := client.Query(ctx, "Give me a simple recipe for scrambled eggs. Respond only with JSON.")
	if err != nil {
		log.Fatal(err)
	}

	// Collect the full response
	text, err := stream.CollectText()
	if err != nil {
		log.Fatal(err)
	}

	// Extract JSON from response (Claude might add markdown code blocks)
	jsonStr := extractJSON(text)
	if jsonStr == "" {
		log.Fatal("No JSON found in response")
	}

	// Parse the structured response
	var recipe Recipe
	if err := json.Unmarshal([]byte(jsonStr), &recipe); err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	// Display the structured data
	fmt.Printf("Recipe: %s\n", recipe.Name)
	fmt.Printf("Prep Time: %s\n", recipe.PrepTime)
	fmt.Printf("Cook Time: %s\n", recipe.CookTime)
	fmt.Println("\nIngredients:")
	for _, ing := range recipe.Ingredients {
		fmt.Printf("  - %s\n", ing)
	}
	fmt.Println("\nSteps:")
	for i, step := range recipe.Steps {
		fmt.Printf("  %d. %s\n", i+1, step)
	}
}

// extractJSON extracts JSON from a string that might contain markdown code blocks
func extractJSON(s string) string {
	// Try to find JSON in code blocks
	re := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find raw JSON (starts with { or [)
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return s
	}

	return ""
}
