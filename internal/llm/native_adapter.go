package llm

import (
	"context"
	"fmt"

	"github.com/tmc/pe/internal/promptfoo"
)

// NativeProviderAdapter adapts native providers to the legacy Provider interface
type NativeProviderAdapter struct {
	provider Provider
}

// NewNativeProviderAdapter creates a new adapter for native providers
func NewNativeProviderAdapter(provider Provider) (*NativeProviderAdapter, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}

	return &NativeProviderAdapter{
		provider: provider,
	}, nil
}

// Name returns the provider name
func (a *NativeProviderAdapter) Name() string {
	return a.provider.Name()
}

// Model returns the model being used
func (a *NativeProviderAdapter) Model() string {
	return a.provider.Model()
}

// Generate generates text using the native provider
func (a *NativeProviderAdapter) Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error) {
	return a.provider.Generate(ctx, prompt, options)
}

// SupportsStreaming returns whether the provider supports streaming
func (a *NativeProviderAdapter) SupportsStreaming() bool {
	return a.provider.SupportsStreaming()
}

// SupportsBatch returns whether the provider supports batch processing
func (a *NativeProviderAdapter) SupportsBatch() bool {
	return a.provider.SupportsBatch()
}

// EvaluatePrompt implements the legacy Provider interface for backward compatibility
func (a *NativeProviderAdapter) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Extract generation options from vars
	options := GenerateOptions{}

	if temp, ok := vars["temperature"].(float64); ok {
		options.Temperature = &temp
	}

	if maxTokens, ok := vars["max_tokens"].(int); ok {
		options.MaxTokens = &maxTokens
	} else if maxTokens, ok := vars["max_tokens"].(float64); ok {
		mt := int(maxTokens)
		options.MaxTokens = &mt
	}

	if topP, ok := vars["top_p"].(float64); ok {
		options.TopP = &topP
	}

	if topK, ok := vars["top_k"].(int); ok {
		options.TopK = &topK
	} else if topK, ok := vars["top_k"].(float64); ok {
		tk := int(topK)
		options.TopK = &tk
	}

	if stop, ok := vars["stop"].([]string); ok {
		options.Stop = stop
	} else if stop, ok := vars["stop"].(string); ok {
		options.Stop = []string{stop}
	}

	// Generate response
	resp, err := a.provider.Generate(ctx, prompt, options)
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

// GenerateStream implements streaming generation if supported
func (a *NativeProviderAdapter) GenerateStream(ctx context.Context, prompt string, options GenerateOptions) (<-chan *StreamResponse, error) {
	if sp, ok := a.provider.(StreamingProvider); ok {
		return sp.GenerateStream(ctx, prompt, options)
	}
	return nil, fmt.Errorf("provider %s does not support streaming", a.provider.Name())
}

// GenerateBatch implements batch generation if supported
func (a *NativeProviderAdapter) GenerateBatch(ctx context.Context, prompts []string, options GenerateOptions) ([]*GenerateResponse, error) {
	if bp, ok := a.provider.(BatchProvider); ok {
		return bp.GenerateBatch(ctx, prompts, options)
	}
	return nil, fmt.Errorf("provider %s does not support batch processing", a.provider.Name())
}
