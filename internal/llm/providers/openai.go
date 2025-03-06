package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/tmc/pe/internal/llm"
)

// Ensure OpenAIProvider implements the Provider interface.
var _ llm.Provider = (*OpenAIProvider)(nil)

// Ensure OpenAIModel implements the Model interface.
var _ llm.Model = (*OpenAIModel)(nil)

// OpenAIProvider implements the Provider interface for OpenAI.
type OpenAIProvider struct {
	apiKey     string
	models     map[string]*OpenAIModel
	modelsLock sync.RWMutex
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(config map[string]interface{}) (llm.Provider, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	provider := &OpenAIProvider{
		apiKey: apiKey,
		models: make(map[string]*OpenAIModel),
	}

	// Initialize available models
	models := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
	}

	for _, modelName := range models {
		provider.models[modelName] = &OpenAIModel{
			name:     modelName,
			provider: provider,
		}
	}

	return provider, nil
}

// Name returns the name of the provider.
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Models returns the available models for this provider.
func (p *OpenAIProvider) Models() []llm.Model {
	p.modelsLock.RLock()
	defer p.modelsLock.RUnlock()

	models := make([]llm.Model, 0, len(p.models))
	for _, model := range p.models {
		models = append(models, model)
	}
	return models
}

// GetModel returns a specific model by name.
func (p *OpenAIProvider) GetModel(name string) (llm.Model, error) {
	p.modelsLock.RLock()
	defer p.modelsLock.RUnlock()

	model, exists := p.models[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", llm.ErrModelNotFound, name)
	}
	return model, nil
}

// OpenAIModel represents a specific OpenAI model.
type OpenAIModel struct {
	name     string
	provider *OpenAIProvider
}

// Name returns the name of the model.
func (m *OpenAIModel) Name() string {
	return m.name
}

// Provider returns the provider this model belongs to.
func (m *OpenAIModel) Provider() llm.Provider {
	return m.provider
}

// Complete processes a prompt and returns a completion.
func (m *OpenAIModel) Complete(ctx context.Context, request llm.CompletionRequest) (llm.CompletionResponse, error) {
	// In a real implementation, this would call the OpenAI API
	// For this mock implementation, we'll return a simple response
	return llm.CompletionResponse{
		Text:         fmt.Sprintf("This is a mock completion from %s for: %s", m.name, request.Prompt),
		FinishReason: "length",
		UsageStats: llm.UsageStats{
			PromptTokens:     len(request.Prompt) / 4, // Rough approximation
			CompletionTokens: 20,
			TotalTokens:      len(request.Prompt)/4 + 20,
		},
		ModelInfo: llm.ModelInfo{
			Name:      m.name,
			Provider:  m.provider.Name(),
			ModelType: "text-generation",
		},
	}, nil
}

// CompleteStream processes a prompt and streams the completion.
func (m *OpenAIModel) CompleteStream(ctx context.Context, request llm.CompletionRequest) (llm.CompletionStream, error) {
	// In a real implementation, this would set up a connection to the OpenAI streaming API
	// For this mock implementation, we'll return a simple stream
	return &openAICompletionStream{
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

// openAICompletionStream implements the CompletionStream interface for OpenAI.
type openAICompletionStream struct {
	model        *OpenAIModel
	request      llm.CompletionRequest
	chunks       []string
	currentIndex int
}

// Next returns the next chunk of the completion.
func (s *openAICompletionStream) Next(ctx context.Context) (llm.CompletionChunk, error) {
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
func (s *openAICompletionStream) Close() error {
	// In a real implementation, this would close the connection to OpenAI
	return nil
}