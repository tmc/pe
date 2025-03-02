// Package googleai provides Google AI-specific LLM integration.
package googleai

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

// Provider represents a Google AI provider.
type Provider struct {
	apiKey   string
	baseURL  string
	models   map[string]llm.Model
	defModel llm.Model
}

// NewProvider creates a new Google AI provider.
func NewProvider(apiKey, baseURL string) *Provider {
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1"
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
		{"gemini-1.0-pro", 32768, false},
		{"gemini-1.5-pro", 1000000, false},
		{"gemini-2.0-flash", 1000000, true},
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
	return "googleai"
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

// Model represents a Google AI model.
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
	
	// In a real implementation, this would make an HTTP request to the Google AI API
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
		Metadata: map[string]interface{}{"provider": "googleai"},
	}, nil
}

// Stream generates streaming text completions for the given prompt.
func (m *Model) Stream(ctx context.Context, req *llm.CompletionRequest) (llm.CompletionStream, error) {
	if m.apiKey == "" {
		return nil, llm.ErrAuthFailure
	}
	
	// In a real implementation, this would create a streaming connection to the Google AI API
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

// GoogleAIFactory creates Google AI providers.
type GoogleAIFactory struct{}

// Create creates a Google AI provider.
func (f *GoogleAIFactory) Create(providerStr string) (llm.Provider, error) {
	if !strings.EqualFold(providerStr, "googleai") {
		return nil, fmt.Errorf("%w: %s", llm.ErrProviderNotFound, providerStr)
	}
	
	return NewProvider("", ""), nil
}

// SupportsProvider returns true if the factory can create the provider.
func (f *GoogleAIFactory) SupportsProvider(providerStr string) bool {
	return strings.EqualFold(providerStr, "googleai")
}

// init registers the Google AI provider factory with the default registry.
func init() {
	llm.RegisterFactory(&GoogleAIFactory{})
}