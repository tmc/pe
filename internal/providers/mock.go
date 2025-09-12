package providers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// MockProvider implements a mock LLM provider for testing
type MockProvider struct {
	model           string
	evaluationCount map[string]int // Track evaluation calls per objective for progressive scoring
}

// NewMockProvider creates a new mock provider instance
func NewMockProvider(model string, options map[string]interface{}) (*MockProvider, error) {
	// Check if we're in test mode
	if os.Getenv("PE_TEST_MODE") != "true" {
		return nil, fmt.Errorf("mock provider only available in test mode")
	}

	return &MockProvider{
		model:           model,
		evaluationCount: make(map[string]int),
	}, nil
}

// Name returns the provider name
func (p *MockProvider) Name() string {
	return "mock"
}

// Model returns the model being used
func (p *MockProvider) Model() string {
	return p.model
}

// Generate returns a mock response
func (p *MockProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	// Generate mock response based on prompt content
	responseText := "Mock response for: " + prompt

	// Debug: print what we see in test mode
	if os.Getenv("PE_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "MOCK DEBUG: Prompt contains 'effectiveness': %v, '0.0 to 1.0': %v, 'maximize accuracy': %v\n",
			strings.Contains(prompt, "effectiveness"),
			strings.Contains(prompt, "0.0 to 1.0"),
			strings.Contains(prompt, "maximize accuracy"))
	}

	// For semantic optimization tests, return appropriate scores
	if os.Getenv("PE_MOCK_PROVIDER") == "true" || os.Getenv("PE_TEST_MODE") == "true" {
		// Debug: Check if this is an evaluation prompt
		if os.Getenv("PE_DEBUG") == "true" && strings.Contains(prompt, "OBJECTIVE:") {
			fmt.Fprintf(os.Stderr, "MOCK DEBUG: Evaluation prompt detected. Objective line: ")
			for _, line := range strings.Split(prompt, "\n") {
				if strings.Contains(line, "OBJECTIVE:") {
					fmt.Fprintf(os.Stderr, "%s\n", line)
					break
				}
			}
		}

		// Handle specific test cases
		if strings.Contains(prompt, "What is 2+2?") {
			responseText = "4"
		} else if strings.Contains(prompt, "pointer") {
			responseText = "A pointer is a variable that stores the memory address of another variable."
		} else if strings.Contains(prompt, "What is the capital of France?") {
			// For extract test - include the XML tags
			responseText = "<answer>The capital of France is Paris</answer>"
		} else if strings.Contains(prompt, "Calculate the meaning of life") {
			responseText = "<result>42</result>\n<explanation>The answer to life, the universe, and everything</explanation>"
		} else if strings.Contains(prompt, "List some items") {
			responseText = "<data>\n  <item>First</item>\n  <item>Second</item>\n</data>"
		} else if strings.Contains(prompt, "Provide analysis") {
			responseText = `<response type="preliminary">Initial thoughts</response>
<response type="final">This is the final answer</response>`
		} else if strings.Contains(prompt, "Generate data") {
			responseText = `<data>
  <content>Main content here</content>
  <metadata>
    <author>AI Assistant</author>
    <timestamp>2024-01-20</timestamp>
  </metadata>
</data>`
		} else if strings.Contains(prompt, "What is the capital?") {
			responseText = `<answer confidence="low">Maybe London?</answer>
<answer confidence="high">Definitely Paris</answer>`
		} else if strings.Contains(prompt, "List fruits") {
			responseText = "<item>Apple</item>\n<item>Banana</item>\n<item>Cherry</item>"
		} else if strings.Contains(prompt, "Write Go code") {
			responseText = `<code>
func main() {
fmt.Println("Hello")
}
</code>`
		} else if strings.Contains(prompt, "Solve the problem") {
			responseText = `<solution>{"x": 5, "y": 10}</solution>`
		} else if strings.Contains(prompt, "Process this") {
			responseText = "[[RESULT]]Custom delimited content[[/RESULT]]"
		} else if strings.Contains(prompt, "No tags here") {
			responseText = "No tags here, just plain text."
		} else if strings.Contains(prompt, "Stream some chunks") {
			responseText = "<chunk>Chunk 1</chunk>\n<chunk>Chunk 2</chunk>\n<chunk>Chunk 3</chunk>"
		} else if strings.Contains(prompt, "Count from 1 to 10") {
			responseText = "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"
		} else if strings.Contains(prompt, "List 5 random numbers") {
			responseText = "3\n7\n2\n9\n5"
		} else if strings.Contains(prompt, "Generate a paragraph about AI") {
			responseText = "Artificial Intelligence represents a transformative technology that is reshaping our world."
		} else if strings.Contains(prompt, "Be concise but explain everything") {
			responseText = "I understand you're looking for a balance. As a helpful assistant, I'll provide clear and comprehensive explanations while being concise."
		} else if strings.Contains(prompt, "Expensive computation") {
			responseText = "Result: 42"
		} else if strings.Contains(prompt, "Check status") {
			responseText = "System status: OK"
		} else if strings.Contains(prompt, "Explain quantum computing") {
			responseText = "Quantum computing uses quantum mechanics principles for computation."
		} else if strings.Contains(prompt, "Rate the prompt") || (strings.Contains(prompt, "effectiveness") && strings.Contains(prompt, "0.0 to 1.0")) {
			// Return improving scores over iterations
			if strings.Contains(prompt, "Optimized prompt content") {
				responseText = "0.97" // Improved score after optimization
			} else if strings.Contains(prompt, "maximize accuracy") || strings.Contains(prompt, "OBJECTIVE: maximize accuracy") {
				// For semantic descent test with accuracy objective
				// Progressively improve scores to simulate gradient descent
				count := p.evaluationCount["maximize_accuracy"]
				p.evaluationCount["maximize_accuracy"]++
				switch count {
				case 0:
					responseText = "0.65" // Initial score (loss = 0.35)
				case 1:
					responseText = "0.75" // Improving
				case 2:
					responseText = "0.82" // Further improvement
				case 3:
					responseText = "0.86" // Getting better
				case 4:
					responseText = "0.88" // Step 5 (loss = 0.12)
				case 5:
					responseText = "0.90" // Continue improving
				case 6:
					responseText = "0.91" // Almost converged
				case 7:
					responseText = "0.9105" // Converged (small improvement < 0.001)
				default:
					responseText = "0.92" // Converged state
				}
			} else {
				// Debug in test mode
				if os.Getenv("PE_DEBUG") == "true" {
					snippet := prompt
					if len(snippet) > 100 {
						snippet = snippet[:100]
					}
					fmt.Fprintf(os.Stderr, "MOCK DEBUG: Using default score. Prompt snippet: %s...\n", snippet)
				}
				responseText = "0.85" // Default score for semantic evaluation
			}
		} else if strings.Contains(prompt, "semantic gradient") || strings.Contains(prompt, "Provide feedback in JSON format with an array of gradients") {
			// Return mock gradient response for textgrad
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
		} else if strings.Contains(prompt, "Apply the following semantic gradients") || strings.Contains(prompt, "apply semantic gradients") || strings.Contains(prompt, "Improve this prompt based on the following textual gradients") {
			// Return optimized prompt based on the original
			if strings.Contains(prompt, "Analyze the provided code for bugs") {
				responseText = "Analyze the provided code for potential bugs and errors with improved clarity and specificity"
			} else if strings.Contains(prompt, "Analyze the sentiment") {
				responseText = "Analyze the sentiment of the provided text. Classify as positive, negative, or neutral. Include confidence score and key phrases supporting the classification."
			} else if strings.Contains(prompt, "ORIGINAL PROMPT:") {
				// Extract original prompt and return optimized version
				responseText = "Optimized prompt with improvements applied"
			} else {
				responseText = "Optimized prompt content with improvements applied"
			}
		}
	}

	return &llm.GenerateResponse{
		Text:             responseText,
		PromptTokens:     len(prompt) / 4,
		CompletionTokens: len(responseText) / 4,
		TotalTokens:      (len(prompt) + len(responseText)) / 4,
		Latency:          10 * time.Millisecond,
		Cost:             0.0,
		Model:            p.model,
		FinishReason:     "stop",
	}, nil
}

// GenerateStream returns a mock streaming response
func (p *MockProvider) GenerateStream(ctx context.Context, prompt string, options llm.GenerateOptions) (<-chan *llm.StreamResponse, error) {
	respChan := make(chan *llm.StreamResponse, 10)

	go func() {
		defer close(respChan)

		// Simulate streaming response
		words := []string{"Mock", "streaming", "response", "for:", prompt}

		for _, word := range words {
			select {
			case <-ctx.Done():
				respChan <- &llm.StreamResponse{Error: ctx.Err()}
				return
			case respChan <- &llm.StreamResponse{
				Text:    word + " ",
				Latency: 10 * time.Millisecond,
			}:
				time.Sleep(10 * time.Millisecond)
			}
		}

		respChan <- &llm.StreamResponse{
			Done:    true,
			Latency: time.Duration(len(words)) * 10 * time.Millisecond,
		}
	}()

	return respChan, nil
}

// SupportsStreaming returns whether the provider supports streaming
func (p *MockProvider) SupportsStreaming() bool {
	return true
}

// SupportsBatch returns whether the provider supports batch processing
func (p *MockProvider) SupportsBatch() bool {
	return false
}

// EvaluatePrompt implements the legacy Provider interface for backward compatibility
func (p *MockProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Convert to GenerateOptions
	options := llm.GenerateOptions{}

	// Generate response
	resp, err := p.Generate(ctx, prompt, options)
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

// init registers the mock provider when in test mode
func init() {
	if os.Getenv("PE_TEST_MODE") == "true" {
		RegisterProvider("mock", func(model string, options map[string]interface{}) (llm.Provider, error) {
			return NewMockProvider(model, options)
		})
	}
}
