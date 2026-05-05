package llm

import (
	"context"
	"time"

	"github.com/tmc/pe/internal/cgpt"
	"github.com/tmc/pe/internal/promptfoo"
)

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
		Cost:      resp.Cost,
		Cached:    false,
		LatencyMs: resp.LatencyMs,
		Metadata:  resp.Metadata,
	}, nil
}
