package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/tmc/pe/internal/llm"
)

// Ensure AnthropicProvider implements the Provider interface.
var _ llm.Provider = (*AnthropicProvider)(nil)

// Ensure AnthropicModel implements the Model interface.
var _ llm.Model = (*AnthropicModel)(nil)

// AnthropicProvider implements the Provider interface for Anthropic.
type AnthropicProvider struct {
	apiKey     string
	models     map[string]*AnthropicModel
	modelsLock sync.RWMutex
}

// NewAnthropicProvider creates a new Anthropic provider.
func NewAnthropicProvider(config map[string]interface{}) (llm.Provider, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, errors.New("Anthropic API key is required")
	}

	provider := &AnthropicProvider{
		apiKey: apiKey,
		models: make(map[string]*AnthropicModel),
	}

	// Initialize available models
	models := []string{
		"claude-2",
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"claude-3.5-sonnet",
	}

	for _, modelName := range models {
		provider.models[modelName] = &AnthropicModel{
			name:     modelName,
			provider: provider,
		}
	}

	return provider, nil
}

// Name returns the name of the provider.
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Models returns the available models for this provider.
func (p *AnthropicProvider) Models() []llm.Model {
	p.modelsLock.RLock()
	defer p.modelsLock.RUnlock()

	models := make([]llm.Model, 0, len(p.models))
	for _, model := range p.models {
		models = append(models, model)
	}
	return models
}

// GetModel returns a specific model by name.
func (p *AnthropicProvider) GetModel(name string) (llm.Model, error) {
	p.modelsLock.RLock()
	defer p.modelsLock.RUnlock()

	model, exists := p.models[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", llm.ErrModelNotFound, name)
	}
	return model, nil
}

// AnthropicModel represents a specific Anthropic model.
type AnthropicModel struct {
	name     string
	provider *AnthropicProvider
}

// Name returns the name of the model.
func (m *AnthropicModel) Name() string {
	return m.name
}

// Provider returns the provider this model belongs to.
func (m *AnthropicModel) Provider() llm.Provider {
	return m.provider
}

// Complete processes a prompt and returns a completion.
func (m *AnthropicModel) Complete(ctx context.Context, request llm.CompletionRequest) (llm.CompletionResponse, error) {
	// In a real implementation, this would call the Anthropic API
	// For this mock implementation, we'll return a simple response
	return llm.CompletionResponse{
		Text:         fmt.Sprintf("This is a mock completion from %s for: %s", m.name, request.Prompt),
		FinishReason: "length",
		UsageStats: llm.UsageStats{
			PromptTokens:     len(request.Prompt) / 4, // Rough approximation
			CompletionTokens: 25,
			TotalTokens:      len(request.Prompt)/4 + 25,
		},
		ModelInfo: llm.ModelInfo{
			Name:      m.name,
			Provider:  m.provider.Name(),
			ModelType: "text-generation",
		},
	}, nil
}

// CompleteStream processes a prompt and streams the completion.
func (m *AnthropicModel) CompleteStream(ctx context.Context, request llm.CompletionRequest) (llm.CompletionStream, error) {
	// In a real implementation, this would set up a connection to the Anthropic streaming API
	// For this mock implementation, we'll return a simple stream
	return &anthropicCompletionStream{
		model:   m,
		request: request,
		chunks: []string{
			"This is ",
			"a mock ",
			"streaming ",
			"completion ",
			fmt.Sprintf("from %s ", m.name),
			fmt.Sprintf("for: %s", request.Prompt),
		},
		currentIndex: 0,
	}, nil
}

// anthropicCompletionStream implements the CompletionStream interface for Anthropic.
type anthropicCompletionStream struct {
	model        *AnthropicModel
	request      llm.CompletionRequest
	chunks       []string
	currentIndex int
}

// Next returns the next chunk of the completion.
func (s *anthropicCompletionStream) Next(ctx context.Context) (llm.CompletionChunk, error) {
	if s.currentIndex >= len(s.chunks) {
		return llm.CompletionChunk{}, io.EOF
	}

	chunk := llm.CompletionChunk{
		Text:    s.chunks[s.currentIndex],
		IsFinal: s.currentIndex == len(s.chunks)-1,
	}

	if chunk.IsFinal {
		chunk.FinishReason = "stop"
	}

	s.currentIndex++
	return chunk, nil
}

// Close closes the stream.
func (s *anthropicCompletionStream) Close() error {
	// In a real implementation, this would close the connection to Anthropic
	return nil
}