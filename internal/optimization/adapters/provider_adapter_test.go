package adapters

import (
	"context"
	"testing"
	"time"

	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/optimization"
	"github.com/tmc/pe/internal/promptfoo"
)

// MockInferenceProvider implements inference.Provider for testing
type MockInferenceProvider struct {
	name string
}

func (m *MockInferenceProvider) Name() string {
	return m.name
}

func (m *MockInferenceProvider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	return &inference.Response{
		Content: "mock response",
		Model:   "mock-model",
		TokensUsed: inference.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *MockInferenceProvider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	ch := make(chan inference.StreamChunk)
	close(ch)
	return ch, nil
}

func (m *MockInferenceProvider) Models(ctx context.Context) ([]string, error) {
	return []string{"mock-model"}, nil
}

func (m *MockInferenceProvider) Close() error {
	return nil
}

// MockLLMProvider implements llm.Provider for testing
type MockLLMProvider struct {
	name  string
	model string
	calls int
}

func (m *MockLLMProvider) Name() string {
	return m.name
}

func (m *MockLLMProvider) Model() string {
	return m.model
}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	m.calls++
	return &llm.GenerateResponse{
		Text:             "mock response",
		PromptTokens:     10,
		CompletionTokens: 5,
		TotalTokens:      15,
		Latency:          100 * time.Millisecond,
		Cost:             0.001,
		Model:            m.model,
		FinishReason:     "stop",
	}, nil
}

func (m *MockLLMProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return &promptfoo.ProviderResponse{
		Output: "mock response",
	}, nil
}

func (m *MockLLMProvider) SupportsStreaming() bool {
	return true
}

func (m *MockLLMProvider) SupportsBatch() bool {
	return false
}

func TestNewInferenceProviderAdapter(t *testing.T) {
	mock := &MockInferenceProvider{name: "test-provider"}
	adapter := NewInferenceProviderAdapter(mock)

	if adapter == nil {
		t.Fatal("NewInferenceProviderAdapter returned nil")
	}

	// Verify it implements the interface
	var _ optimization.LanguageModelProvider = adapter
}

func TestInferenceProviderAdapter_Name(t *testing.T) {
	mock := &MockInferenceProvider{name: "test-provider"}
	adapter := NewInferenceProviderAdapter(mock)

	name := adapter.Name()
	if name != "test-provider" {
		t.Errorf("expected name 'test-provider', got '%s'", name)
	}
}

func TestInferenceProviderAdapter_Model(t *testing.T) {
	mock := &MockInferenceProvider{name: "test-provider"}
	adapter := NewInferenceProviderAdapter(mock)

	model := adapter.Model()
	if model != "default" {
		t.Errorf("expected model 'default', got '%s'", model)
	}
}

func TestInferenceProviderAdapter_Generate(t *testing.T) {
	mock := &MockInferenceProvider{name: "test-provider"}
	adapter := NewInferenceProviderAdapter(mock)

	temp := 0.7
	maxTokens := 100
	options := optimization.GenerationOptions{
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Stop:        []string{"END"},
	}

	resp, err := adapter.Generate(context.Background(), "test prompt", options)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Text != "mock response" {
		t.Errorf("expected text 'mock response', got '%s'", resp.Text)
	}
	if resp.TotalTokens != 15 {
		t.Errorf("expected total tokens 15, got %d", resp.TotalTokens)
	}
}

func TestInferenceProviderAdapter_SupportsStreaming(t *testing.T) {
	mock := &MockInferenceProvider{name: "test-provider"}
	adapter := NewInferenceProviderAdapter(mock)

	if !adapter.SupportsStreaming() {
		t.Error("expected SupportsStreaming to return true")
	}
}

func TestInferenceProviderAdapter_SupportsBatch(t *testing.T) {
	mock := &MockInferenceProvider{name: "test-provider"}
	adapter := NewInferenceProviderAdapter(mock)

	if adapter.SupportsBatch() {
		t.Error("expected SupportsBatch to return false")
	}
}

func TestNewLLMProviderAdapter(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	adapter := NewLLMProviderAdapter(mock)

	if adapter == nil {
		t.Fatal("NewLLMProviderAdapter returned nil")
	}

	var _ optimization.LanguageModelProvider = adapter
}

func TestLLMProviderAdapter_Name(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	adapter := NewLLMProviderAdapter(mock)

	name := adapter.Name()
	if name != "test-llm" {
		t.Errorf("expected name 'test-llm', got '%s'", name)
	}
}

func TestLLMProviderAdapter_Model(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	adapter := NewLLMProviderAdapter(mock)

	model := adapter.Model()
	if model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", model)
	}
}

func TestLLMProviderAdapter_Generate(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	adapter := NewLLMProviderAdapter(mock)

	temp := 0.7
	options := optimization.GenerationOptions{
		Temperature: &temp,
	}

	resp, err := adapter.Generate(context.Background(), "test prompt", options)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Text != "mock response" {
		t.Errorf("expected text 'mock response', got '%s'", resp.Text)
	}
	if resp.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", resp.Model)
	}
}

func TestLLMProviderAdapter_SupportsStreaming(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	adapter := NewLLMProviderAdapter(mock)

	if !adapter.SupportsStreaming() {
		t.Error("expected SupportsStreaming to return true")
	}
}

func TestLLMProviderAdapter_SupportsBatch(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	adapter := NewLLMProviderAdapter(mock)

	if adapter.SupportsBatch() {
		t.Error("expected SupportsBatch to return false")
	}
}

func TestNewProviderAdapterFactory(t *testing.T) {
	factory := NewProviderAdapterFactory()
	if factory == nil {
		t.Fatal("NewProviderAdapterFactory returned nil")
	}
}

func TestProviderAdapterFactory_CreateAdapter_InferenceProvider(t *testing.T) {
	factory := NewProviderAdapterFactory()
	mock := &MockInferenceProvider{name: "test-provider"}

	adapter, err := factory.CreateAdapter(mock)
	if err != nil {
		t.Fatalf("CreateAdapter failed: %v", err)
	}

	if adapter.Name() != "test-provider" {
		t.Errorf("expected name 'test-provider', got '%s'", adapter.Name())
	}
}

func TestProviderAdapterFactory_CreateAdapter_LLMProvider(t *testing.T) {
	factory := NewProviderAdapterFactory()
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}

	adapter, err := factory.CreateAdapter(mock)
	if err != nil {
		t.Fatalf("CreateAdapter failed: %v", err)
	}

	if adapter.Name() != "test-llm" {
		t.Errorf("expected name 'test-llm', got '%s'", adapter.Name())
	}
}

func TestProviderAdapterFactory_CreateAdapter_AlreadyAdapter(t *testing.T) {
	factory := NewProviderAdapterFactory()
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	existingAdapter := NewLLMProviderAdapter(mock)

	adapter, err := factory.CreateAdapter(existingAdapter)
	if err != nil {
		t.Fatalf("CreateAdapter failed: %v", err)
	}

	// Should return the same adapter
	if adapter != existingAdapter {
		t.Error("expected same adapter to be returned")
	}
}

func TestProviderAdapterFactory_CreateAdapter_UnsupportedType(t *testing.T) {
	factory := NewProviderAdapterFactory()

	_, err := factory.CreateAdapter("not a provider")
	if err == nil {
		t.Error("expected error for unsupported provider type")
	}
}

func TestProviderAdapterFactory_CreateAdapterFromSpec_Invalid(t *testing.T) {
	factory := NewProviderAdapterFactory()

	_, err := factory.CreateAdapterFromSpec("nonexistent:model", nil)
	if err == nil {
		t.Error("expected error for invalid provider spec")
	}
}

func TestNativeProviderAdapters(t *testing.T) {
	tests := []struct {
		name string
		got  optimization.LanguageModelProvider
		want string
	}{
		{"openai", NewOpenAIAdapter("test-key", ""), "openai"},
		{"anthropic", NewAnthropicAdapter("test-key", ""), "anthropic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.Name() != tt.want {
				t.Fatalf("Name = %q, want %q", tt.got.Name(), tt.want)
			}
		})
	}
}

func TestCachedProviderAdapterCachesResponses(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	base := NewLLMProviderAdapter(mock)
	adapter := NewCachedProviderAdapter(base)

	first, err := adapter.Generate(context.Background(), "test prompt", optimization.GenerationOptions{})
	if err != nil {
		t.Fatalf("Generate first: %v", err)
	}
	second, err := adapter.Generate(context.Background(), "test prompt", optimization.GenerationOptions{})
	if err != nil {
		t.Fatalf("Generate second: %v", err)
	}
	if mock.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", mock.calls)
	}
	if first == second {
		t.Fatal("cached adapter returned shared response pointer")
	}
	if second.Text != first.Text {
		t.Fatalf("cached response text = %q, want %q", second.Text, first.Text)
	}
}

func TestMetricsProviderAdapterRecordsMetrics(t *testing.T) {
	mock := &MockLLMProvider{name: "test-llm", model: "gpt-4"}
	base := NewLLMProviderAdapter(mock)
	adapter := NewMetricsProviderAdapter(base)

	if _, err := adapter.Generate(context.Background(), "test prompt", optimization.GenerationOptions{}); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	metrics := adapter.Metrics()
	if metrics.Requests != 1 {
		t.Fatalf("Requests = %d, want 1", metrics.Requests)
	}
	if metrics.Errors != 0 {
		t.Fatalf("Errors = %d, want 0", metrics.Errors)
	}
	if metrics.TotalTokens != 15 {
		t.Fatalf("TotalTokens = %d, want 15", metrics.TotalTokens)
	}
	if metrics.TotalCost != 0.001 {
		t.Fatalf("TotalCost = %v, want 0.001", metrics.TotalCost)
	}
	if metrics.LastFinishCode != "stop" {
		t.Fatalf("LastFinishCode = %q, want stop", metrics.LastFinishCode)
	}
}
