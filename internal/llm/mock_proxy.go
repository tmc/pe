package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/pe/internal/promptfoo"
)

// MockProviderProxy is a simple mock provider for testing
type MockProviderProxy struct{}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Name returns the provider name
func (p *MockProviderProxy) Name() string {
	return "mock"
}

// Model returns the model being used
func (p *MockProviderProxy) Model() string {
	return "mock-model"
}

// Generate returns a mock response
func (p *MockProviderProxy) Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error) {
	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	// Generate mock response based on prompt content
	responseText := "Mock response for: " + prompt

	// Debug output in test mode
	if os.Getenv("PE_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "MOCK DEBUG: Received prompt (first 200 chars): %s\n", prompt[:min(200, len(prompt))])
	}

	// For optimization tests, return appropriate scores
	if (strings.Contains(prompt, "Rate") && strings.Contains(prompt, "scale of 0.0 to 1.0")) ||
		(strings.Contains(prompt, "Evaluate this prompt") && strings.Contains(prompt, "scale from 0.0 to 1.0")) {
		// Return improving scores over iterations
		if strings.Contains(prompt, "Analyze the sentiment of the provided text. Classify as positive") {
			responseText = "0.92" // Improved score after optimization
		} else if strings.Contains(prompt, "Optimized prompt content") {
			responseText = "0.97" // Improved score after optimization
		} else {
			responseText = "0.85" // Default score for semantic evaluation
		}
	} else if strings.Contains(prompt, "semantic gradient") || strings.Contains(prompt, "Compute semantic gradients") || strings.Contains(prompt, "textual gradients") || strings.Contains(prompt, "Provide feedback in JSON format with an array of gradients") {
		// Return mock gradient response in the format expected by textgrad
		responseText = `{
			"gradients": [
				{
					"component": "overall clarity",
					"feedback": "The prompt lacks specificity and clear output requirements",
					"suggestions": ["Specify the sentiment categories (positive/negative/neutral)", "Add output format instructions"],
					"confidence": 0.8,
					"priority": 0.9,
					"gradient": "add structure and specificity",
					"magnitude": 0.7,
					"direction": "improve"
				}
			]
		}`
	} else if strings.Contains(prompt, "Apply the following semantic gradients") || strings.Contains(prompt, "Improve this prompt based on the following textual gradients") {
		// Return optimized prompt based on the gradients
		if strings.Contains(prompt, "Analyze the sentiment") {
			responseText = "Analyze the sentiment of the provided text. Classify as positive, negative, or neutral. Include confidence score and key phrases supporting the classification."
		} else {
			responseText = "Optimized prompt content with improvements applied"
		}
	} else if strings.Contains(prompt, "Generate a better version") {
		// PE2 optimization response
		responseText = "Analyze the sentiment of the given text with improved clarity and specificity."
	} else if strings.Contains(prompt, "STEP-BY-STEP REASONING TEMPLATE") || strings.Contains(prompt, "Analyze the provided prompt") {
		// PE2 meta-prompt response
		responseText = `Analysis: [Your systematic analysis of the current prompt, identifying specific strengths and areas for improvement]
Reasoning: [Your step-by-step reasoning process for the optimization, showing how you arrived at each improvement]
Improvements: [Brief summary of key improvements made and expected impact]

OPTIMIZED PROMPT:
Analyze the sentiment of the given text with improved clarity and specificity.

SCORE: 10.0`
	}

	return &GenerateResponse{
		Text:             responseText,
		PromptTokens:     len(prompt) / 4,
		CompletionTokens: len(responseText) / 4,
		TotalTokens:      (len(prompt) + len(responseText)) / 4,
		Latency:          10 * time.Millisecond,
		Cost:             0.0,
		Model:            p.Model(),
		FinishReason:     "stop",
	}, nil
}

// SupportsStreaming returns whether the provider supports streaming
func (p *MockProviderProxy) SupportsStreaming() bool {
	return false
}

// SupportsBatch returns whether the provider supports batch processing
func (p *MockProviderProxy) SupportsBatch() bool {
	return false
}

// EvaluatePrompt implements the legacy Provider interface for backward compatibility
func (p *MockProviderProxy) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Generate response
	resp, err := p.Generate(ctx, prompt, GenerateOptions{})
	if err != nil {
		return nil, err
	}

	// Convert to promptfoo response format
	return &promptfoo.ProviderResponse{
		Output: resp.Text,
		TokenUsage: &promptfoo.TokenUsage{
			Total:      int32(resp.TotalTokens),
			Prompt:     int32(resp.PromptTokens),
			Completion: int32(resp.CompletionTokens),
			Cached:     0,
		},
		Cost:   resp.Cost,
		Cached: false,
	}, nil
}
