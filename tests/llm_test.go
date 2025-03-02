package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// Mock implementations for testing

type MockProvider struct {
	name   string
	models []llm.Model
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) Models() []llm.Model {
	return p.models
}

func (p *MockProvider) GetModel(name string) (llm.Model, error) {
	for _, model := range p.models {
		if model.Name() == name {
			return model, nil
		}
	}
	return nil, llm.ErrModelNotFound
}

func (p *MockProvider) DefaultModel() llm.Model {
	if len(p.models) == 0 {
		return nil
	}
	return p.models[0]
}

type MockModel struct {
	name             string
	provider         llm.Provider
	maxContextSize   int
	streamable       bool
	completeResponse *llm.CompletionResponse
	completeError    error
	streamResponse   []llm.CompletionChunk
	streamError      error
}

func (m *MockModel) Name() string {
	return m.name
}

func (m *MockModel) Provider() llm.Provider {
	return m.provider
}

func (m *MockModel) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.completeError != nil {
		return nil, m.completeError
	}
	return m.completeResponse, nil
}

func (m *MockModel) Stream(ctx context.Context, req *llm.CompletionRequest) (llm.CompletionStream, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}
	return &MockStream{chunks: m.streamResponse}, nil
}

func (m *MockModel) MaxContextSize() int {
	return m.maxContextSize
}

func (m *MockModel) IsStreamable() bool {
	return m.streamable
}

type MockStream struct {
	chunks []llm.CompletionChunk
	index  int
	closed bool
}

func (s *MockStream) Next() (*llm.CompletionChunk, error) {
	if s.closed {
		return nil, errors.New("stream closed")
	}
	
	if s.index >= len(s.chunks) {
		return nil, nil
	}
	
	chunk := &s.chunks[s.index]
	s.index++
	return chunk, nil
}

func (s *MockStream) Close() error {
	s.closed = true
	return nil
}

type MockFactory struct {
	providers map[string]llm.Provider
}

func (f *MockFactory) Create(providerStr string) (llm.Provider, error) {
	provider, exists := f.providers[providerStr]
	if !exists {
		return nil, llm.ErrProviderNotFound
	}
	return provider, nil
}

func (f *MockFactory) SupportsProvider(providerStr string) bool {
	_, exists := f.providers[providerStr]
	return exists
}

// Tests

func TestRegistry(t *testing.T) {
	registry := llm.NewRegistry()
	
	// Create mock providers and models
	openaiProvider := &MockProvider{name: "openai"}
	gpt4 := &MockModel{
		name:             "gpt-4",
		provider:         openaiProvider,
		maxContextSize:   8192,
		streamable:       true,
		completeResponse: &llm.CompletionResponse{Text: "GPT-4 response"},
	}
	openaiProvider.models = []llm.Model{gpt4}
	
	anthropicProvider := &MockProvider{name: "anthropic"}
	claude := &MockModel{
		name:             "claude",
		provider:         anthropicProvider,
		maxContextSize:   100000,
		streamable:       true,
		completeResponse: &llm.CompletionResponse{Text: "Claude response"},
	}
	anthropicProvider.models = []llm.Model{claude}
	
	// Register providers
	registry.RegisterProvider(openaiProvider)
	
	// Test GetProvider
	t.Run("GetProvider", func(t *testing.T) {
		provider, err := registry.GetProvider("openai")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if provider.Name() != "openai" {
			t.Errorf("Expected provider name 'openai', got: %s", provider.Name())
		}
		
		// Test case insensitivity
		provider, err = registry.GetProvider("OpenAI")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if provider.Name() != "openai" {
			t.Errorf("Expected provider name 'openai', got: %s", provider.Name())
		}
		
		// Test non-existent provider
		_, err = registry.GetProvider("nonexistent")
		if !errors.Is(err, llm.ErrProviderNotFound) {
			t.Errorf("Expected ErrProviderNotFound, got: %v", err)
		}
	})
	
	// Test ListProviders
	t.Run("ListProviders", func(t *testing.T) {
		providers := registry.ListProviders()
		if len(providers) != 1 {
			t.Errorf("Expected 1 provider, got: %d", len(providers))
		}
		if providers[0].Name() != "openai" {
			t.Errorf("Expected provider name 'openai', got: %s", providers[0].Name())
		}
	})
	
	// Test GetModelByFullName
	t.Run("GetModelByFullName", func(t *testing.T) {
		model, err := registry.GetModelByFullName("openai:gpt-4")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if model.Name() != "gpt-4" {
			t.Errorf("Expected model name 'gpt-4', got: %s", model.Name())
		}
		
		// Test default model
		model, err = registry.GetModelByFullName("openai")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if model.Name() != "gpt-4" {
			t.Errorf("Expected default model name 'gpt-4', got: %s", model.Name())
		}
		
		// Test non-existent model
		_, err = registry.GetModelByFullName("openai:nonexistent")
		if !errors.Is(err, llm.ErrModelNotFound) {
			t.Errorf("Expected ErrModelNotFound, got: %v", err)
		}
	})
	
	// Test factory
	t.Run("Factory", func(t *testing.T) {
		factory := &MockFactory{
			providers: map[string]llm.Provider{
				"anthropic": anthropicProvider,
			},
		}
		
		registry.RegisterFactory(factory)
		
		// Test creating a provider through factory
		provider, err := registry.GetProvider("anthropic")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if provider.Name() != "anthropic" {
			t.Errorf("Expected provider name 'anthropic', got: %s", provider.Name())
		}
		
		// Test ListModels after factory registration
		models := registry.ListModels()
		if len(models) != 2 {
			t.Errorf("Expected 2 models, got: %d", len(models))
		}
	})
}

func TestProviderAndModelParsing(t *testing.T) {
	t.Run("ParseProviderAndModel", func(t *testing.T) {
		tests := []struct{
			input          string
			wantProvider   string
			wantModel      string
		}{
			{"openai:gpt-4", "openai", "gpt-4"},
			{"anthropic:claude-3", "anthropic", "claude-3"},
			{"openai", "openai", ""},
			{"", "", ""},
			{"with:multiple:colons", "with", "multiple:colons"},
		}
		
		for _, tt := range tests {
			provider, model := llm.ParseProviderAndModel(tt.input)
			if provider != tt.wantProvider {
				t.Errorf("ParseProviderAndModel(%q) provider = %q, want %q", tt.input, provider, tt.wantProvider)
			}
			if model != tt.wantModel {
				t.Errorf("ParseProviderAndModel(%q) model = %q, want %q", tt.input, model, tt.wantModel)
			}
		}
	})
	
	t.Run("Format", func(t *testing.T) {
		tests := []struct{
			provider    string
			model       string
			want        string
		}{
			{"openai", "gpt-4", "openai:gpt-4"},
			{"anthropic", "claude-3", "anthropic:claude-3"},
			{"openai", "", "openai"},
			{"", "", ""},
		}
		
		for _, tt := range tests {
			result := llm.Format(tt.provider, tt.model)
			if result != tt.want {
				t.Errorf("Format(%q, %q) = %q, want %q", tt.provider, tt.model, result, tt.want)
			}
		}
	})
}

func TestDefaultRegistry(t *testing.T) {
	// Create mock provider and model
	provider := &MockProvider{name: "test-provider"}
	model := &MockModel{
		name:             "test-model",
		provider:         provider,
		maxContextSize:   1000,
		streamable:       true,
		completeResponse: &llm.CompletionResponse{Text: "Test response"},
	}
	provider.models = []llm.Model{model}
	
	// Register provider with default registry
	llm.RegisterProvider(provider)
	
	// Test global functions
	t.Run("GlobalFunctions", func(t *testing.T) {
		// GetProvider
		p, err := llm.GetProvider("test-provider")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if p.Name() != "test-provider" {
			t.Errorf("Expected provider name 'test-provider', got: %s", p.Name())
		}
		
		// GetModelByFullName
		m, err := llm.GetModelByFullName("test-provider:test-model")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if m.Name() != "test-model" {
			t.Errorf("Expected model name 'test-model', got: %s", m.Name())
		}
		
		// ListProviders
		providers := llm.ListProviders()
		found := false
		for _, p := range providers {
			if p.Name() == "test-provider" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find 'test-provider' in ListProviders result")
		}
		
		// ListModels
		models := llm.ListModels()
		found = false
		for _, m := range models {
			if m.Name() == "test-model" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find 'test-model' in ListModels result")
		}
	})
}

func TestCompletionAndStreaming(t *testing.T) {
	// Create mock provider and model with responses
	provider := &MockProvider{name: "test-provider"}
	
	completionResponse := &llm.CompletionResponse{
		Text:         "This is a test response",
		FinishReason: "stop",
		Usage: llm.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
		Model:    "test-model",
		Created:  time.Now(),
		Latency:  100 * time.Millisecond,
		Metadata: map[string]interface{}{"test": "value"},
	}
	
	streamChunks := []llm.CompletionChunk{
		{Text: "This ", IsFinal: false},
		{Text: "is ", IsFinal: false},
		{Text: "a ", IsFinal: false},
		{Text: "test ", IsFinal: false},
		{Text: "response", IsFinal: true},
	}
	
	model := &MockModel{
		name:             "test-model",
		provider:         provider,
		maxContextSize:   1000,
		streamable:       true,
		completeResponse: completionResponse,
		streamResponse:   streamChunks,
	}
	
	provider.models = []llm.Model{model}
	
	// Test completion
	t.Run("Completion", func(t *testing.T) {
		req := &llm.CompletionRequest{
			Prompt:       "Test prompt",
			MaxTokens:    100,
			Temperature:  0.7,
			TopP:         1.0,
			StopSequences: []string{"\n"},
		}
		
		resp, err := model.Complete(context.Background(), req)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if resp.Text != completionResponse.Text {
			t.Errorf("Expected text %q, got: %q", completionResponse.Text, resp.Text)
		}
		
		if resp.FinishReason != completionResponse.FinishReason {
			t.Errorf("Expected finish reason %q, got: %q", completionResponse.FinishReason, resp.FinishReason)
		}
		
		if resp.Usage.TotalTokens != completionResponse.Usage.TotalTokens {
			t.Errorf("Expected total tokens %d, got: %d", completionResponse.Usage.TotalTokens, resp.Usage.TotalTokens)
		}
	})
	
	// Test streaming
	t.Run("Streaming", func(t *testing.T) {
		req := &llm.CompletionRequest{
			Prompt:      "Test prompt",
			MaxTokens:   100,
			Temperature: 0.7,
		}
		
		stream, err := model.Stream(context.Background(), req)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		defer stream.Close()
		
		var fullText string
		var chunkCount int
		
		for {
			chunk, err := stream.Next()
			if err != nil {
				t.Errorf("Error from stream: %v", err)
				break
			}
			
			if chunk == nil {
				break
			}
			
			fullText += chunk.Text
			chunkCount++
			
			if chunk.IsFinal && chunkCount < len(streamChunks) {
				t.Errorf("Expected IsFinal on last chunk only, got it on chunk %d", chunkCount)
			}
		}
		
		expectedText := "This is a test response"
		if fullText != expectedText {
			t.Errorf("Expected full text %q, got: %q", expectedText, fullText)
		}
		
		if chunkCount != len(streamChunks) {
			t.Errorf("Expected %d chunks, got: %d", len(streamChunks), chunkCount)
		}
	})
	
	// Test error conditions
	t.Run("ErrorHandling", func(t *testing.T) {
		// Create model with errors
		errorModel := &MockModel{
			name:          "error-model",
			provider:      provider,
			completeError: errors.New("completion error"),
			streamError:   errors.New("stream error"),
		}
		
		// Test completion error
		_, err := errorModel.Complete(context.Background(), &llm.CompletionRequest{})
		if err == nil {
			t.Error("Expected completion error, got nil")
		}
		
		// Test stream error
		_, err = errorModel.Stream(context.Background(), &llm.CompletionRequest{})
		if err == nil {
			t.Error("Expected stream error, got nil")
		}
	})
}