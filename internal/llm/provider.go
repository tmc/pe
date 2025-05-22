package llm

import (
	"context"
	"fmt"
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
	switch backend {
	case "cgpt", "openai", "anthropic", "gemini", "googleai":
		return &CGPTProvider{
			Backend: backend,
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
