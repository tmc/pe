package inference_test

import (
	"context"
	"testing"

	"github.com/tmc/pe/internal/inference"
)

// MockProvider is a simple mock provider for testing.
type MockProvider struct {
	name      string
	response  string
	models    []string
	streamErr error
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	return &inference.Response{
		Content: m.response,
		Model:   req.Model,
		TokensUsed: inference.TokenUsage{
			PromptTokens:     len(req.Prompt),
			CompletionTokens: len(m.response),
			TotalTokens:      len(req.Prompt) + len(m.response),
		},
	}, nil
}

func (m *MockProvider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	if m.streamErr != nil {
		return nil, m.streamErr
	}

	chunks := make(chan inference.StreamChunk, 3)
	go func() {
		defer close(chunks)

		// Simulate streaming by sending response in chunks
		parts := []string{"Hello", " ", "world!"}
		for _, part := range parts {
			chunks <- inference.StreamChunk{
				Delta: part,
				Done:  false,
			}
		}
		chunks <- inference.StreamChunk{
			Done: true,
		}
	}()

	return chunks, nil
}

func (m *MockProvider) Models(ctx context.Context) ([]string, error) {
	return m.models, nil
}

func (m *MockProvider) Close() error {
	return nil
}

func TestClient(t *testing.T) {
	ctx := context.Background()

	// Create client with mock providers
	client := inference.NewClient()

	mock1 := &MockProvider{
		name:     "mock1",
		response: "Response from mock1",
		models:   []string{"model1", "model2"},
	}

	mock2 := &MockProvider{
		name:     "mock2",
		response: "Response from mock2",
		models:   []string{"model3"},
	}

	client.Register("mock1", mock1)
	client.Register("mock2", mock2)

	// Test listing providers
	providers := client.Providers()
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}

	// Test completion with default provider
	resp, err := client.Complete(ctx, inference.Request{
		Prompt: "Test prompt",
		Model:  "test-model",
	})
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if resp.Content != "Response from mock1" {
		t.Errorf("Unexpected response: %s", resp.Content)
	}

	// Test completion with specific provider
	resp, err = client.CompleteWith(ctx, "mock2", inference.Request{
		Prompt: "Test prompt 2",
	})
	if err != nil {
		t.Fatalf("CompleteWith failed: %v", err)
	}
	if resp.Content != "Response from mock2" {
		t.Errorf("Unexpected response: %s", resp.Content)
	}

	// Test streaming
	chunks, err := client.Stream(ctx, inference.Request{
		Prompt: "Stream test",
		Stream: true,
	})
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	var collected string
	for chunk := range chunks {
		if chunk.Error != nil {
			t.Fatalf("Stream chunk error: %v", chunk.Error)
		}
		if !chunk.Done {
			collected += chunk.Delta
		}
	}

	if collected != "Hello world!" {
		t.Errorf("Unexpected streamed content: %s", collected)
	}

	// Test models
	models, err := client.Models(ctx, "mock1")
	if err != nil {
		t.Fatalf("Models failed: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	// Test invalid provider
	_, err = client.CompleteWith(ctx, "invalid", inference.Request{})
	if err == nil {
		t.Error("Expected error for invalid provider")
	}
}

func TestCollectStream(t *testing.T) {
	ctx := context.Background()

	// Create a channel with test data
	chunks := make(chan inference.StreamChunk, 4)
	chunks <- inference.StreamChunk{Delta: "Hello"}
	chunks <- inference.StreamChunk{Delta: " "}
	chunks <- inference.StreamChunk{Delta: "world!"}
	chunks <- inference.StreamChunk{Done: true}
	close(chunks)

	content, err := inference.CollectStream(ctx, chunks)
	if err != nil {
		t.Fatalf("CollectStream failed: %v", err)
	}

	if content != "Hello world!" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func TestRegistry(t *testing.T) {
	// Test provider registration
	inference.Register("test-provider", func(config map[string]interface{}) (inference.Provider, error) {
		return &MockProvider{
			name:     "test-provider",
			response: "Test response",
		}, nil
	})

	// Check if provider is registered
	providers := inference.Providers()
	found := false
	for _, p := range providers {
		if p == "test-provider" {
			found = true
			break
		}
	}

	if !found {
		t.Error("test-provider not found in registry")
	}

	// Create provider from registry
	provider, err := inference.NewProvider("test-provider", nil)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider.Name() != "test-provider" {
		t.Errorf("Unexpected provider name: %s", provider.Name())
	}
}
