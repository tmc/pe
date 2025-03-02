// Package anthropic provides Anthropic-specific LLM integration.
package anthropic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// Provider represents an Anthropic provider.
type Provider struct {
	apiKey   string
	baseURL  string
	models   map[string]llm.Model
	defModel llm.Model
}

// NewProvider creates a new Anthropic provider.
func NewProvider(apiKey, baseURL string) *Provider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	
	provider := &Provider{
		apiKey:  apiKey,
		baseURL: baseURL,
		models:  make(map[string]llm.Model),
	}
	
	// Create models
	models := []struct {
		name           string
		maxContextSize int
		isDefault      bool
	}{
		{"claude-3-opus", 200000, true},
		{"claude-3-sonnet", 200000, false},
		{"claude-3-haiku", 200000, false},
		{"claude-2", 100000, false},
		{"claude-instant", 100000, false},
	}
	
	for _, m := range models {
		model := &Model{
			name:           m.name,
			provider:       provider,
			maxContextSize: m.maxContextSize,
			streamable:     true,
			apiKey:         apiKey,
			baseURL:        baseURL,
		}
		
		provider.models[m.name] = model
		
		if m.isDefault {
			provider.defModel = model
		}
	}
	
	return provider
}

// Name returns the name of the provider.
func (p *Provider) Name() string {
	return "anthropic"
}

// Models returns the available models.
func (p *Provider) Models() []llm.Model {
	models := make([]llm.Model, 0, len(p.models))
	for _, model := range p.models {
		models = append(models, model)
	}
	return models
}

// GetModel returns a specific model by name.
func (p *Provider) GetModel(name string) (llm.Model, error) {
	model, exists := p.models[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", llm.ErrModelNotFound, name)
	}
	return model, nil
}

// DefaultModel returns the default model.
func (p *Provider) DefaultModel() llm.Model {
	return p.defModel
}

// Model represents an Anthropic model.
type Model struct {
	name           string
	provider       *Provider
	maxContextSize int
	streamable     bool
	apiKey         string
	baseURL        string
}

// Name returns the name of the model.
func (m *Model) Name() string {
	return m.name
}

// Provider returns the provider that offers this model.
func (m *Model) Provider() llm.Provider {
	return m.provider
}

// Complete generates text completions for the given prompt.
func (m *Model) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.apiKey == "" {
		return nil, llm.ErrAuthFailure
	}
	
	// In a real implementation, this would make an HTTP request to the Anthropic API
	// For now, we'll return a mock response
	startTime := time.Now()
	
	// Get prompt tokens (rough estimate)
	promptTokens := len(strings.Split(req.Prompt, " ")) / 3 * 4
	
	// Mock response text
	responseText := fmt.Sprintf("This is a mock response from %s model for prompt: %s", m.name, req.Prompt)
	
	// Get completion tokens (rough estimate)
	completionTokens := len(strings.Split(responseText, " ")) / 3 * 4
	
	return &llm.CompletionResponse{
		Text:         responseText,
		FinishReason: "stop",
		Usage: llm.TokenUsage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
		Model:    m.name,
		Created:  time.Now(),
		Latency:  time.Since(startTime),
		Metadata: map[string]interface{}{"provider": "anthropic"},
	}, nil
}

// Stream generates streaming text completions for the given prompt.
func (m *Model) Stream(ctx context.Context, req *llm.CompletionRequest) (llm.CompletionStream, error) {
	if m.apiKey == "" {
		return nil, llm.ErrAuthFailure
	}
	
	// In a real implementation, this would create a streaming connection to the Anthropic API
	// For now, we'll return a mock stream
	responseText := fmt.Sprintf("This is a mock response from %s model for prompt: %s", m.name, req.Prompt)
	
	// Split the response into chunks
	words := strings.Split(responseText, " ")
	chunks := make([]string, 0)
	
	for i := 0; i < len(words); i += 2 {
		end := i + 2
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}
	
	// Create and return the stream
	return &mockStream{chunks: chunks}, nil
}

// MaxContextSize returns the maximum context size in tokens.
func (m *Model) MaxContextSize() int {
	return m.maxContextSize
}

// IsStreamable returns true if the model supports streaming.
func (m *Model) IsStreamable() bool {
	return m.streamable
}

// mockStream implements the CompletionStream interface for testing.
type mockStream struct {
	chunks  []string
	index   int
	closed  bool
}

// Next returns the next chunk of the completion.
func (s *mockStream) Next() (*llm.CompletionChunk, error) {
	if s.closed {
		return nil, errors.New("stream is closed")
	}
	
	if s.index >= len(s.chunks) {
		return nil, io.EOF
	}
	
	chunk := &llm.CompletionChunk{
		Text:    s.chunks[s.index],
		IsFinal: s.index == len(s.chunks)-1,
	}
	
	s.index++
	return chunk, nil
}

// Close closes the stream.
func (s *mockStream) Close() error {
	s.closed = true
	return nil
}

// AnthropicFactory creates Anthropic providers.
type AnthropicFactory struct{}

// Create creates an Anthropic provider.
func (f *AnthropicFactory) Create(providerStr string) (llm.Provider, error) {
	if strings.ToLower(providerStr) != "anthropic" {
		return nil, fmt.Errorf("%w: %s", llm.ErrProviderNotFound, providerStr)
	}
	
	return NewProvider("", ""), nil
}

// SupportsProvider returns true if the factory can create the provider.
func (f *AnthropicFactory) SupportsProvider(providerStr string) bool {
	return strings.ToLower(providerStr) == "anthropic"
}

// init registers the Anthropic provider factory with the default registry.
func init() {
	llm.RegisterFactory(&AnthropicFactory{})
}