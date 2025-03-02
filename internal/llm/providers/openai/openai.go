// Package openai provides OpenAI-specific LLM integration.
package openai

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

// Provider represents an OpenAI provider.
type Provider struct {
	apiKey   string
	baseURL  string
	models   map[string]llm.Model
	defModel llm.Model
}

// NewProvider creates a new OpenAI provider.
func NewProvider(apiKey, baseURL string) *Provider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
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
		{"gpt-4", 8192, true},
		{"gpt-4-turbo", 128000, false},
		{"gpt-4-32k", 32768, false},
		{"gpt-3.5-turbo", 4096, false},
		{"gpt-3.5-turbo-16k", 16384, false},
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
	return "openai"
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

// Model represents an OpenAI model.
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
	
	// In a real implementation, this would make an HTTP request to the OpenAI API
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
		Metadata: map[string]interface{}{"provider": "openai"},
	}, nil
}

// Stream generates streaming text completions for the given prompt.
func (m *Model) Stream(ctx context.Context, req *llm.CompletionRequest) (llm.CompletionStream, error) {
	if m.apiKey == "" {
		return nil, llm.ErrAuthFailure
	}
	
	// In a real implementation, this would create a streaming connection to the OpenAI API
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

// OpenAIFactory creates OpenAI providers.
type OpenAIFactory struct{}

// Create creates an OpenAI provider.
func (f *OpenAIFactory) Create(providerStr string) (llm.Provider, error) {
	if strings.ToLower(providerStr) != "openai" {
		return nil, fmt.Errorf("%w: %s", llm.ErrProviderNotFound, providerStr)
	}
	
	return NewProvider("", ""), nil
}

// SupportsProvider returns true if the factory can create the provider.
func (f *OpenAIFactory) SupportsProvider(providerStr string) bool {
	return strings.ToLower(providerStr) == "openai"
}

// init registers the OpenAI provider factory with the default registry.
func init() {
	llm.RegisterFactory(&OpenAIFactory{})
}