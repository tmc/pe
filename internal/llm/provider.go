package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/pe/internal/cgpt"
	"github.com/tmc/pe/internal/promptfoo"
)

// GenerateOptions contains options for text generation
type GenerateOptions struct {
	Temperature *float64
	MaxTokens   *int
	TopP        *float64
	TopK        *int
	Stop        []string
}

// GenerateResponse contains the response from a generation request
type GenerateResponse struct {
	Text             string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Latency          time.Duration
	Cost             float64
	Model            string
	FinishReason     string
}

// StreamResponse contains a streaming response chunk
type StreamResponse struct {
	Text    string
	Done    bool
	Latency time.Duration
	Error   error
}

// Provider is an interface for language model providers
type Provider interface {
	// Legacy method for backward compatibility
	EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error)
	
	// Enhanced methods for native providers
	Name() string
	Model() string
	Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error)
	SupportsStreaming() bool
	SupportsBatch() bool
}

// StreamingProvider extends Provider with streaming capabilities
type StreamingProvider interface {
	Provider
	GenerateStream(ctx context.Context, prompt string, options GenerateOptions) (<-chan *StreamResponse, error)
}

// BatchProvider extends Provider with batch processing capabilities
type BatchProvider interface {
	Provider
	GenerateBatch(ctx context.Context, prompts []string, options GenerateOptions) ([]*GenerateResponse, error)
}

// GetProvider returns a Provider for the given backend
func GetProvider(backend string) (Provider, error) {
	// Check if we should use mock provider in test mode
	if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
		// Import providers package to get the mock provider
		// This is a hack to avoid circular dependency - in production we'd refactor
		// to have the providers package register with llm package
		return &MockProviderProxy{}, nil
	}
	
	// Parse provider:model format
	provider := backend
	model := ""
	if idx := strings.IndexByte(backend, ':'); idx != -1 {
		provider = backend[:idx]
		model = backend[idx+1:]
	}
	
	switch provider {
	case "cgpt", "openai", "anthropic", "gemini", "googleai":
		return &CGPTProvider{
			Backend: provider,
			model:   model,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported backend: %s", backend)
	}
}

// CreateNativeProvider creates a native provider using the providers package
// This function will be used to transition away from CGPT dependency
func CreateNativeProvider(providerSpec string, options map[string]interface{}) (Provider, error) {
	// Import providers package dynamically to avoid circular dependencies
	// This is a placeholder - in real implementation we'd use the registry
	return nil, fmt.Errorf("native providers integration in progress - use GetProvider for now")
}

// CGPTProvider implements Provider using the CGPT client
type CGPTProvider struct {
	Backend string
	model   string
}

// Name returns the provider name
func (p *CGPTProvider) Name() string {
	return p.Backend
}

// Model returns the model being used
func (p *CGPTProvider) Model() string {
	return p.model
}

// Generate generates text using the CGPT provider
func (p *CGPTProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error) {
	startTime := time.Now()
	
	// Create CGPT provider configuration
	provider := cgpt.DefaultProvider()
	provider.Backend = p.Backend
	if p.model != "" {
		provider.Model = p.model
	}

	// Convert options to vars
	vars := make(map[string]interface{})
	if options.Temperature != nil {
		vars["temperature"] = *options.Temperature
	}
	if options.MaxTokens != nil {
		vars["max_tokens"] = *options.MaxTokens
	}

	// Evaluate the prompt
	resp, err := provider.EvaluatePromptWithOptions(prompt, vars, false)
	if err != nil {
		return nil, err
	}

	latency := time.Since(startTime)

	return &GenerateResponse{
		Text:             resp.Output,
		PromptTokens:     int(resp.TokenUsage.Prompt),
		CompletionTokens: int(resp.TokenUsage.Completion),
		TotalTokens:      int(resp.TokenUsage.Total),
		Latency:          latency,
		Cost:             resp.Cost,
		Model:            p.Model(),
		FinishReason:     "stop", // CGPT doesn't provide finish reason
	}, nil
}

// SupportsStreaming returns whether the provider supports streaming
func (p *CGPTProvider) SupportsStreaming() bool {
	return false // CGPT wrapper doesn't support streaming
}

// SupportsBatch returns whether the provider supports batch processing
func (p *CGPTProvider) SupportsBatch() bool {
	return false
}

// EvaluatePrompt evaluates a prompt with the given variables (legacy method)
func (p *CGPTProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Create CGPT provider configuration
	provider := cgpt.DefaultProvider()

	// Set the provider backend
	provider.Backend = p.Backend

	// Set the model if specified
	if p.Model() != "" {
		provider.Model = p.Model()
	} else if model, ok := vars["model"].(string); ok {
		provider.Model = model
	}

	// Evaluate the prompt
	resp, err := provider.EvaluatePromptWithOptions(prompt, vars, false)
	if err != nil {
		return nil, err
	}

	// Convert to promptfoo response format
	return &promptfoo.ProviderResponse{
		Output: resp.Output,
		TokenUsage: &promptfoo.TokenUsage{
			Total:      resp.TokenUsage.Total,
			Prompt:     resp.TokenUsage.Prompt,
			Completion: resp.TokenUsage.Completion,
			Cached:     resp.TokenUsage.Cached,
		},
		Cost:   resp.Cost,
		Cached: false,
	}, nil
}

// MockProviderProxy is a simple mock provider for testing
type MockProviderProxy struct{}

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
	
	// For optimization tests, return appropriate scores
	if strings.Contains(prompt, "Rate") && strings.Contains(prompt, "scale of 0.0 to 1.0") {
		// Return improving scores over iterations
		if strings.Contains(prompt, "Optimized prompt content") {
			responseText = "0.97" // Improved score after optimization
		} else {
			responseText = "0.85" // Default score for semantic evaluation
		}
	} else if strings.Contains(prompt, "semantic gradient") {
		// Return mock gradient response
		responseText = `{
			"gradients": [
				{
					"component": "overall clarity",
					"direction": "improve specificity and structure", 
					"magnitude": 0.7,
					"reasoning": "General improvement needed",
					"confidence": 0.6
				}
			]
		}`
	} else if strings.Contains(prompt, "Apply the following semantic gradients") {
		// Return optimized prompt
		responseText = "Optimized prompt content"
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
