// Example of using PE's inference API
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/inference/providers/cgpt"
)

func main() {
	ctx := context.Background()

	// Create inference client
	client := inference.NewClient()

	// Register cgpt provider
	provider := cgpt.New()
	client.Register("cgpt", provider)

	// Example 1: Simple completion
	fmt.Println("=== Example 1: Simple Completion ===")
	resp, err := client.Complete(ctx, inference.Request{
		Prompt:      "What is 2+2?",
		Temperature: 0,
	})
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}
	fmt.Printf("Response: %s\n\n", resp.Content)

	// Example 2: With system prompt
	fmt.Println("=== Example 2: With System Prompt ===")
	resp, err = client.Complete(ctx, inference.Request{
		Prompt:       "Explain quantum computing",
		SystemPrompt: "You are a teacher explaining to a 10-year-old. Keep it simple and use analogies.",
		Temperature:  0.7,
		MaxTokens:    150,
	})
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}
	fmt.Printf("Response: %s\n\n", resp.Content)

	// Example 3: Streaming
	fmt.Println("=== Example 3: Streaming Response ===")
	chunks, err := client.Stream(ctx, inference.Request{
		Prompt:      "Write a haiku about programming",
		Temperature: 0.9,
		Stream:      true,
	})
	if err != nil {
		log.Fatalf("Streaming failed: %v", err)
	}

	fmt.Print("Streaming: ")
	for chunk := range chunks {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}
		if chunk.Done {
			fmt.Println("\n[Done]")
			break
		}
		fmt.Print(chunk.Delta)
	}

	// Example 4: List available models
	fmt.Println("\n=== Example 4: Available Models ===")
	models, err := client.Models(ctx, "cgpt")
	if err != nil {
		log.Fatalf("Failed to get models: %v", err)
	}
	fmt.Println("Available models:")
	for _, model := range models {
		fmt.Printf("  - %s\n", model)
	}

	// Clean up
	if err := client.Close(); err != nil {
		log.Printf("Warning: failed to close client: %v", err)
	}
}

// To run this example:
// 1. Make sure you have cgpt installed or available via go run
// 2. Set your API key: export OPENAI_API_KEY=your-key
// 3. Run: go run examples/inference/basic.go
