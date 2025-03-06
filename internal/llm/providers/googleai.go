package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/tmc/pe/internal/llm"
)

// Ensure GoogleAIProvider implements the Provider interface.
var _ llm.Provider = (*GoogleAIProvider)(nil)

// Ensure GoogleAIModel implements the Model interface.
var _ llm.Model = (*GoogleAIModel)(nil)

// GoogleAIProvider implements the Provider interface for Google AI.
type GoogleAIProvider struct {
	apiKey     string
	projectID  string
	models     map[string]*GoogleAIModel
	modelsLock sync.RWMutex
}

// NewGoogleAIProvider creates a new Google AI provider.
func NewGoogleAIProvider(config map[string]interface{}) (llm.Provider, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, errors.New("Google AI API key is required")
	}

	projectID, _ := config["project_id"].(string)

	provider := &GoogleAIProvider{
		apiKey:    apiKey,
		projectID: projectID,
		models:    make(map[string]*GoogleAIModel),
	}

	// Initialize available models
	models := []string{
		"gemini-pro",
		"gemini-ultra",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
	}

	for _, modelName := range models {
		provider.models[modelName] = &GoogleAIModel{
			name:     modelName,
			provider: provider,
		}
	}

	return provider, nil
}

// Name returns the name of the provider.
func (p *GoogleAIProvider) Name() string {
	return "googleai"
}

// Models returns the available models for this provider.
func (p *GoogleAIProvider) Models() []llm.Model {
	p.modelsLock.RLock()
	defer p.modelsLock.RUnlock()

	models := make([]llm.Model, 0, len(p.models))
	for _, model := range p.models {
		models = append(models, model)
	}
	return models
}

// GetModel returns a specific model by name.
func (p *GoogleAIProvider) GetModel(name string) (llm.Model, error) {
	p.modelsLock.RLock()
	defer p.modelsLock.RUnlock()

	model, exists := p.models[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", llm.ErrModelNotFound, name)
	}
	return model, nil
}

// GoogleAIModel represents a specific Google AI model.
type GoogleAIModel struct {
	name     string
	provider *GoogleAIProvider
}

// Name returns the name of the model.
func (m *GoogleAIModel) Name() string {
	return m.name
}

// Provider returns the provider this model belongs to.
func (m *GoogleAIModel) Provider() llm.Provider {
	return m.provider
}

// Complete processes a prompt and returns a completion.
func (m *GoogleAIModel) Complete(ctx context.Context, request llm.CompletionRequest) (llm.CompletionResponse, error) {
	// In a real implementation, this would call the Google AI API
	// For this mock implementation, we'll return a simple response
	return llm.CompletionResponse{
		Text:         fmt.Sprintf("This is a mock completion from %s for: %s", m.name, request.Prompt),
		FinishReason: "length",
		UsageStats: llm.UsageStats{
			PromptTokens:     len(request.Prompt) / 4, // Rough approximation
			CompletionTokens: 15,
			TotalTokens:      len(request.Prompt)/4 + 15,
		},
		ModelInfo: llm.ModelInfo{
			Name:      m.name,
			Provider:  m.provider.Name(),
			ModelType: "text-generation",
		},
	}, nil
}

// CompleteStream processes a prompt and streams the completion.
func (m *GoogleAIModel) CompleteStream(ctx context.Context, request llm.CompletionRequest) (llm.CompletionStream, error) {
	// In a real implementation, this would set up a connection to the Google AI streaming API
	// For this mock implementation, we'll return a simple stream
	return &googleAICompletionStream{
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

// googleAICompletionStream implements the CompletionStream interface for Google AI.
type googleAICompletionStream struct {
	model        *GoogleAIModel
	request      llm.CompletionRequest
	chunks       []string
	currentIndex int
}

// Next returns the next chunk of the completion.
func (s *googleAICompletionStream) Next(ctx context.Context) (llm.CompletionChunk, error) {
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
func (s *googleAICompletionStream) Close() error {
	// In a real implementation, this would close the connection to Google AI
	return nil
}