// Package inference provides migration adapters for transitioning from llm.Provider to inference.Provider
package inference

import (
	"context"
	"fmt"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// LegacyAdapter adapts an llm.Provider to the inference.Provider interface
type LegacyAdapter struct {
	legacy llm.Provider
}

// NewLegacyAdapter creates a new adapter for an llm.Provider
func NewLegacyAdapter(provider llm.Provider) Provider {
	return &LegacyAdapter{legacy: provider}
}

// Name returns the provider name
func (l *LegacyAdapter) Name() string {
	return l.legacy.Name()
}

// Complete performs a non-streaming inference using the legacy provider
func (l *LegacyAdapter) Complete(ctx context.Context, req Request) (*Response, error) {
	// Convert Request to legacy format
	vars := make(map[string]interface{})
	if req.Temperature > 0 {
		vars["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		vars["max_tokens"] = req.MaxTokens
	}
	if req.Model != "" {
		vars["model"] = req.Model
	}
	if req.SystemPrompt != "" {
		vars["system_prompt"] = req.SystemPrompt
	}
	
	// Build the full prompt
	prompt := req.Prompt
	if req.SystemPrompt != "" {
		prompt = req.SystemPrompt + "\n\n" + prompt
	}
	if req.Prefill != "" {
		prompt = prompt + "\n\nAssistant: " + req.Prefill
	}
	
	// Call legacy provider
	resp, err := l.legacy.EvaluatePrompt(ctx, prompt, vars)
	if err != nil {
		return nil, err
	}
	
	// Convert response
	return &Response{
		Content: resp.Output,
		Model:   l.legacy.Model(),
		TokensUsed: TokenUsage{
			PromptTokens:     int(resp.TokenUsage.Prompt),
			CompletionTokens: int(resp.TokenUsage.Completion),
			TotalTokens:      int(resp.TokenUsage.Total),
		},
		Metadata: map[string]interface{}{
			"cost":   resp.Cost,
			"cached": resp.Cached,
		},
	}, nil
}

// Stream performs a streaming inference (not supported by legacy providers)
func (l *LegacyAdapter) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	// Legacy providers don't support streaming, so we simulate it
	chunks := make(chan StreamChunk)
	
	go func() {
		defer close(chunks)
		
		// Get the complete response
		resp, err := l.Complete(ctx, req)
		if err != nil {
			chunks <- StreamChunk{Error: err}
			return
		}
		
		// Send as a single chunk
		chunks <- StreamChunk{
			Delta: resp.Content,
			Done:  false,
		}
		chunks <- StreamChunk{
			Done: true,
		}
	}()
	
	return chunks, nil
}

// Models returns available models (legacy providers don't expose this)
func (l *LegacyAdapter) Models(ctx context.Context) ([]string, error) {
	// Return a default list based on provider name
	switch l.legacy.Name() {
	case "openai":
		return []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}, nil
	case "anthropic":
		return []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}, nil
	default:
		return []string{l.legacy.Model()}, nil
	}
}

// Close cleans up resources (legacy providers don't have cleanup)
func (l *LegacyAdapter) Close() error {
	return nil
}

// ModernAdapter adapts an inference.Provider to the llm.Provider interface
type ModernAdapter struct {
	modern Provider
	model  string
}

// NewModernAdapter creates a new adapter for an inference.Provider
func NewModernAdapter(provider Provider, model string) llm.Provider {
	return &ModernAdapter{
		modern: provider,
		model:  model,
	}
}

// Name returns the provider name
func (m *ModernAdapter) Name() string {
	return m.modern.Name()
}

// Model returns the model being used
func (m *ModernAdapter) Model() string {
	if m.model != "" {
		return m.model
	}
	// Try to get from provider's available models
	ctx := context.Background()
	models, err := m.modern.Models(ctx)
	if err == nil && len(models) > 0 {
		return models[0]
	}
	return "unknown"
}

// Generate generates text using the modern provider
func (m *ModernAdapter) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	// Convert to modern Request
	req := Request{
		Prompt: prompt,
		Model:  m.model,
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
	
	// Call modern provider
	startTime := time.Now()
	resp, err := m.modern.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	latency := time.Since(startTime)
	
	// Convert response
	cost := 0.0
	if c, ok := resp.Metadata["cost"].(float64); ok {
		cost = c
	}
	
	return &llm.GenerateResponse{
		Text:             resp.Content,
		PromptTokens:     resp.TokensUsed.PromptTokens,
		CompletionTokens: resp.TokensUsed.CompletionTokens,
		TotalTokens:      resp.TokensUsed.TotalTokens,
		Latency:          latency,
		Cost:             cost,
		Model:            resp.Model,
		FinishReason:     "stop",
	}, nil
}

// SupportsStreaming returns whether the provider supports streaming
func (m *ModernAdapter) SupportsStreaming() bool {
	// All modern providers should support streaming
	return true
}

// SupportsBatch returns whether the provider supports batch processing
func (m *ModernAdapter) SupportsBatch() bool {
	// Batch support is not yet implemented in modern providers
	return false
}

// EvaluatePrompt evaluates a prompt with the given variables (legacy method)
func (m *ModernAdapter) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Convert vars to Request options
	req := Request{
		Prompt:  prompt,
		Model:   m.model,
		Options: vars,
	}
	
	// Extract known options
	if temp, ok := vars["temperature"].(float64); ok {
		req.Temperature = float32(temp)
	}
	if maxTokens, ok := vars["max_tokens"].(int); ok {
		req.MaxTokens = maxTokens
	}
	if model, ok := vars["model"].(string); ok {
		req.Model = model
	}
	if systemPrompt, ok := vars["system_prompt"].(string); ok {
		req.SystemPrompt = systemPrompt
	}
	
	// Call modern provider
	resp, err := m.modern.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	
	// Convert response
	cost := 0.0
	if c, ok := resp.Metadata["cost"].(float64); ok {
		cost = c
	}
	cached := false
	if c, ok := resp.Metadata["cached"].(bool); ok {
		cached = c
	}
	
	return &promptfoo.ProviderResponse{
		Output: resp.Content,
		TokenUsage: &promptfoo.TokenUsage{
			Total:      int32(resp.TokensUsed.TotalTokens),
			Prompt:     int32(resp.TokensUsed.PromptTokens),
			Completion: int32(resp.TokensUsed.CompletionTokens),
			Cached:     0,
		},
		Cost:   cost,
		Cached: cached,
	}, nil
}

// GenerateStream generates text using streaming (for StreamingProvider interface)
func (m *ModernAdapter) GenerateStream(ctx context.Context, prompt string, options llm.GenerateOptions) (<-chan *llm.StreamResponse, error) {
	// Convert to modern Request
	req := Request{
		Prompt: prompt,
		Model:  m.model,
		Stream: true,
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
	
	// Get modern stream
	modernChunks, err := m.modern.Stream(ctx, req)
	if err != nil {
		return nil, err
	}
	
	// Convert stream chunks
	legacyChunks := make(chan *llm.StreamResponse)
	go func() {
		defer close(legacyChunks)
		startTime := time.Now()
		
		for chunk := range modernChunks {
			legacyChunks <- &llm.StreamResponse{
				Text:    chunk.Delta,
				Done:    chunk.Done,
				Latency: time.Since(startTime),
				Error:   chunk.Error,
			}
		}
	}()
	
	return legacyChunks, nil
}

// MigrateProvider converts an llm.Provider to an inference.Provider
func MigrateProvider(legacy llm.Provider) Provider {
	// Check if it's already a modern provider wrapped in ModernAdapter
	if adapter, ok := legacy.(*ModernAdapter); ok {
		return adapter.modern
	}
	
	// Wrap legacy provider
	return NewLegacyAdapter(legacy)
}

// GetLegacyProvider creates an llm.Provider from an inference.Provider
func GetLegacyProvider(modern Provider, model string) llm.Provider {
	// Check if it's already a legacy provider wrapped in LegacyAdapter
	if adapter, ok := modern.(*LegacyAdapter); ok {
		return adapter.legacy
	}
	
	// Wrap modern provider
	return NewModernAdapter(modern, model)
}

// CreateProviderFromSpec creates a Provider from a provider specification string
func CreateProviderFromSpec(spec string, config map[string]interface{}) (Provider, error) {
	// Try to create using the registry first
	provider, err := NewProvider(spec, config)
	if err == nil {
		return provider, nil
	}
	
	// Fall back to legacy provider creation
	legacy, err := llm.GetProvider(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", spec, err)
	}
	
	return MigrateProvider(legacy), nil
}