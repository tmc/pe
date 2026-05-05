// Package adapters provides adapters to decouple optimization algorithms from specific provider implementations.
package adapters

import (
	"context"
	"fmt"

	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/optimization"
)

// InferenceProviderAdapter adapts inference.Provider to optimization.LanguageModelProvider
type InferenceProviderAdapter struct {
	provider inference.Provider
}

// NewInferenceProviderAdapter creates a new adapter for inference providers
func NewInferenceProviderAdapter(provider inference.Provider) optimization.LanguageModelProvider {
	return &InferenceProviderAdapter{
		provider: provider,
	}
}

// Name returns the provider identifier
func (a *InferenceProviderAdapter) Name() string {
	return a.provider.Name()
}

// Model returns the model being used
func (a *InferenceProviderAdapter) Model() string {
	// inference.Provider doesn't have a Model() method directly
	// We'll need to extract it from the provider or use a default
	return "default"
}

// Generate generates text given a prompt and options
func (a *InferenceProviderAdapter) Generate(ctx context.Context, prompt string, options optimization.GenerationOptions) (*optimization.GenerationResponse, error) {
	// Convert optimization options to inference request
	req := inference.Request{
		Prompt: prompt,
	}

	if options.Temperature != nil {
		req.Temperature = float32(*options.Temperature)
	}
	if options.MaxTokens != nil {
		req.MaxTokens = *options.MaxTokens
	}
	if len(options.Stop) > 0 {
		req.StopSequences = options.Stop
	}

	// Use Complete method from inference provider
	resp, err := a.provider.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("inference provider failed: %w", err)
	}

	// Convert inference response to optimization response
	return &optimization.GenerationResponse{
		Text:             resp.Content,
		PromptTokens:     resp.TokensUsed.PromptTokens,
		CompletionTokens: resp.TokensUsed.CompletionTokens,
		TotalTokens:      resp.TokensUsed.TotalTokens,
		Model:            resp.Model,
		FinishReason:     "stop", // inference doesn't provide finish reason
	}, nil
}

// SupportsStreaming indicates if the provider supports streaming
func (a *InferenceProviderAdapter) SupportsStreaming() bool {
	// All inference providers support streaming via Stream method
	return true
}

// SupportsBatch indicates if the provider supports batch processing
func (a *InferenceProviderAdapter) SupportsBatch() bool {
	// inference.Provider doesn't have explicit batch support
	return false
}

// LLMProviderAdapter adapts llm.Provider to optimization.LanguageModelProvider
type LLMProviderAdapter struct {
	provider llm.Provider
}

// NewLLMProviderAdapter creates a new adapter for LLM providers
func NewLLMProviderAdapter(provider llm.Provider) optimization.LanguageModelProvider {
	return &LLMProviderAdapter{
		provider: provider,
	}
}

// Name returns the provider identifier
func (a *LLMProviderAdapter) Name() string {
	return a.provider.Name()
}

// Model returns the model being used
func (a *LLMProviderAdapter) Model() string {
	return a.provider.Model()
}

// Generate generates text given a prompt and options
func (a *LLMProviderAdapter) Generate(ctx context.Context, prompt string, options optimization.GenerationOptions) (*optimization.GenerationResponse, error) {
	// Convert optimization options to LLM options
	llmOptions := llm.GenerateOptions{
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
		TopP:        options.TopP,
		TopK:        options.TopK,
		Stop:        options.Stop,
	}

	// Use Generate method from LLM provider
	resp, err := a.provider.Generate(ctx, prompt, llmOptions)
	if err != nil {
		return nil, fmt.Errorf("LLM provider failed: %w", err)
	}

	// Convert LLM response to optimization response
	return &optimization.GenerationResponse{
		Text:             resp.Text,
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
		Latency:          resp.Latency,
		Cost:             resp.Cost,
		Model:            resp.Model,
		FinishReason:     resp.FinishReason,
	}, nil
}

// SupportsStreaming indicates if the provider supports streaming
func (a *LLMProviderAdapter) SupportsStreaming() bool {
	return a.provider.SupportsStreaming()
}

// SupportsBatch indicates if the provider supports batch processing
func (a *LLMProviderAdapter) SupportsBatch() bool {
	return a.provider.SupportsBatch()
}

// ProviderAdapterFactory creates appropriate adapters based on provider type
type ProviderAdapterFactory struct{}

// NewProviderAdapterFactory creates a new provider adapter factory
func NewProviderAdapterFactory() *ProviderAdapterFactory {
	return &ProviderAdapterFactory{}
}

// CreateAdapter creates an appropriate adapter for the given provider
func (f *ProviderAdapterFactory) CreateAdapter(provider interface{}) (optimization.LanguageModelProvider, error) {
	switch p := provider.(type) {
	case inference.Provider:
		return NewInferenceProviderAdapter(p), nil
	case llm.Provider:
		return NewLLMProviderAdapter(p), nil
	case optimization.LanguageModelProvider:
		// Already the right interface
		return p, nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %T", provider)
	}
}

// CreateAdapterFromSpec creates an adapter from a provider specification
func (f *ProviderAdapterFactory) CreateAdapterFromSpec(providerSpec string, options map[string]interface{}) (optimization.LanguageModelProvider, error) {
	provider, err := inference.CreateProviderFromSpec(providerSpec, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider from spec %q: %w", providerSpec, err)
	}
	return f.CreateAdapter(provider)
}
