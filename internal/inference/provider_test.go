package inference

import (
	"context"
	"errors"
	"testing"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	name       string
	responses  map[string]*Response
	streamData map[string][]StreamChunk
	models     []string
	errors     map[string]error
	closed     bool
}

// NewMockProvider creates a new mock provider for testing
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:       name,
		responses:  make(map[string]*Response),
		streamData: make(map[string][]StreamChunk),
		models:     []string{"mock-model-1", "mock-model-2"},
		errors:     make(map[string]error),
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	if m.closed {
		return nil, errors.New("provider is closed")
	}

	// Check for configured errors
	if err, exists := m.errors[req.Prompt]; exists {
		return nil, err
	}

	// Return configured response
	if response, exists := m.responses[req.Prompt]; exists {
		return response, nil
	}

	// Default response
	return &Response{
		Content: "Mock response for: " + req.Prompt,
		Model:   req.Model,
		TokensUsed: TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

func (m *MockProvider) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	if m.closed {
		return nil, errors.New("provider is closed")
	}

	ch := make(chan StreamChunk, 10)

	// Check for configured errors
	if err, exists := m.errors[req.Prompt]; exists {
		close(ch)
		return ch, err
	}

	go func() {
		defer close(ch)

		// Use configured stream data or default
		chunks := m.streamData[req.Prompt]
		if chunks == nil {
			chunks = []StreamChunk{
				{Delta: "Mock", Done: false},
				{Delta: " stream", Done: false},
				{Delta: " response", Done: true},
			}
		}

		for _, chunk := range chunks {
			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (m *MockProvider) Models(ctx context.Context) ([]string, error) {
	if m.closed {
		return nil, errors.New("provider is closed")
	}

	if err, exists := m.errors["models"]; exists {
		return nil, err
	}

	return m.models, nil
}

func (m *MockProvider) Close() error {
	m.closed = true
	return nil
}

// Helper methods for test configuration

func (m *MockProvider) SetResponse(prompt string, response *Response) {
	m.responses[prompt] = response
}

func (m *MockProvider) SetStreamData(prompt string, chunks []StreamChunk) {
	m.streamData[prompt] = chunks
}

func (m *MockProvider) SetError(operation string, err error) {
	m.errors[operation] = err
}

func (m *MockProvider) SetModels(models []string) {
	m.models = models
}

// Test Provider Interface Compliance

func TestProviderInterface(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
	}{
		{
			name:     "mock_provider",
			provider: NewMockProvider("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test Name method
			name := tt.provider.Name()
			if name == "" {
				t.Error("Name() should return non-empty string")
			}

			// Test Complete method
			req := Request{
				Prompt:      "test prompt",
				Model:       "test-model",
				Temperature: 0.7,
				MaxTokens:   100,
			}

			resp, err := tt.provider.Complete(ctx, req)
			if err != nil {
				t.Errorf("Complete() failed: %v", err)
			}
			if resp == nil {
				t.Error("Complete() should return non-nil response")
			}
			if resp != nil && resp.Content == "" {
				t.Error("Complete() should return response with content")
			}

			// Test Stream method
			stream, err := tt.provider.Stream(ctx, req)
			if err != nil {
				t.Errorf("Stream() failed: %v", err)
			}
			if stream == nil {
				t.Error("Stream() should return non-nil channel")
			}

			// Consume stream
			var chunks []StreamChunk
			for chunk := range stream {
				chunks = append(chunks, chunk)
				if chunk.Done {
					break
				}
			}
			if len(chunks) == 0 {
				t.Error("Stream() should produce at least one chunk")
			}

			// Test Models method
			models, err := tt.provider.Models(ctx)
			if err != nil {
				t.Errorf("Models() failed: %v", err)
			}
			if len(models) == 0 {
				t.Error("Models() should return at least one model")
			}

			// Test Close method
			err = tt.provider.Close()
			if err != nil {
				t.Errorf("Close() failed: %v", err)
			}
		})
	}
}

func TestMockProviderConfiguration(t *testing.T) {
	mock := NewMockProvider("test")

	// Test custom response
	customResp := &Response{
		Content: "Custom response",
		Model:   "custom-model",
		TokensUsed: TokenUsage{
			PromptTokens:     5,
			CompletionTokens: 10,
			TotalTokens:      15,
		},
	}
	mock.SetResponse("custom prompt", customResp)

	ctx := context.Background()
	resp, err := mock.Complete(ctx, Request{Prompt: "custom prompt"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.Content != "Custom response" {
		t.Errorf("Expected custom response, got: %s", resp.Content)
	}

	// Test custom error
	customErr := errors.New("custom error")
	mock.SetError("error prompt", customErr)

	_, err = mock.Complete(ctx, Request{Prompt: "error prompt"})
	if err == nil {
		t.Error("Expected error but got none")
	}
	if err.Error() != "custom error" {
		t.Errorf("Expected custom error, got: %v", err)
	}

	// Test custom models
	customModels := []string{"model-1", "model-2", "model-3"}
	mock.SetModels(customModels)

	models, err := mock.Models(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(models))
	}
	for i, model := range models {
		if model != customModels[i] {
			t.Errorf("Expected model %s, got %s", customModels[i], model)
		}
	}
}

func TestProviderErrorHandling(t *testing.T) {
	mock := NewMockProvider("error-test")

	// Test error after close
	mock.Close()

	ctx := context.Background()
	_, err := mock.Complete(ctx, Request{Prompt: "test"})
	if err == nil {
		t.Error("Expected error from closed provider")
	}

	_, err = mock.Stream(ctx, Request{Prompt: "test"})
	if err == nil {
		t.Error("Expected error from closed provider")
	}

	_, err = mock.Models(ctx)
	if err == nil {
		t.Error("Expected error from closed provider")
	}
}

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr bool
	}{
		{
			name: "valid_request",
			request: Request{
				Prompt:      "Valid prompt",
				Model:       "test-model",
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr: false,
		},
		{
			name: "empty_prompt",
			request: Request{
				Prompt:    "",
				Model:     "test-model",
				MaxTokens: 100,
			},
			wantErr: false, // Empty prompt might be valid in some cases
		},
		{
			name: "negative_temperature",
			request: Request{
				Prompt:      "Test prompt",
				Model:       "test-model",
				Temperature: -1.0,
				MaxTokens:   100,
			},
			wantErr: false, // Provider should handle validation
		},
		{
			name: "high_temperature",
			request: Request{
				Prompt:      "Test prompt",
				Model:       "test-model",
				Temperature: 2.5,
				MaxTokens:   100,
			},
			wantErr: false, // Provider should handle validation
		},
		{
			name: "zero_max_tokens",
			request: Request{
				Prompt:      "Test prompt",
				Model:       "test-model",
				Temperature: 0.7,
				MaxTokens:   0,
			},
			wantErr: false, // Provider should handle validation
		},
	}

	mock := NewMockProvider("validation-test")
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mock.Complete(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponseValidation(t *testing.T) {
	mock := NewMockProvider("response-test")

	// Test response structure
	ctx := context.Background()
	resp, err := mock.Complete(ctx, Request{
		Prompt: "test prompt",
		Model:  "test-model",
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	// Validate response fields
	if resp.Content == "" {
		t.Error("Response content should not be empty")
	}

	if resp.Model == "" {
		t.Error("Response model should not be empty")
	}

	// Validate token usage
	if resp.TokensUsed.TotalTokens <= 0 {
		t.Error("Total tokens should be positive")
	}

	if resp.TokensUsed.TotalTokens != resp.TokensUsed.PromptTokens+resp.TokensUsed.CompletionTokens {
		t.Error("Total tokens should equal prompt + completion tokens")
	}
}

func TestConcurrentProviderAccess(t *testing.T) {
	mock := NewMockProvider("concurrent-test")
	ctx := context.Background()

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Test concurrent Complete calls
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := Request{
				Prompt: "concurrent test",
				Model:  "test-model",
			}
			_, err := mock.Complete(ctx, req)
			results <- err
		}(i)
	}

	// Check results
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent access failed: %v", err)
		}
	}

	// Test concurrent Stream calls
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := Request{
				Prompt: "stream test",
				Model:  "test-model",
			}
			stream, err := mock.Stream(ctx, req)
			if err != nil {
				results <- err
				return
			}

			// Consume stream
			for range stream {
				// Just consume
			}
			results <- nil
		}(i)
	}

	// Check results
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent stream access failed: %v", err)
		}
	}
}