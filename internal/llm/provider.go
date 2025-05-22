package llm

import (
	"context"
	"fmt"

	"github.com/tmc/pe/internal/cgpt"
	"github.com/tmc/pe/internal/promptfoo"
)

// Provider is an interface for language model providers
type Provider interface {
	EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error)
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

// CGPTProvider implements Provider using the CGPT client
type CGPTProvider struct {
	Backend string
	Model   string
}

// EvaluatePrompt evaluates a prompt with the given variables
func (p *CGPTProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Create CGPT provider configuration
	provider := cgpt.DefaultProvider()

	// Set the provider backend
	provider.Backend = p.Backend

	// Set the model if specified
	if p.Model != "" {
		provider.Model = p.Model
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
