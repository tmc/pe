package mocks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tmc/pe/internal/inference"
)

func TestMockProviderBasicFunctionality(t *testing.T) {
	config := MockProviderConfig{
		Name:           "test-provider",
		Models:         []string{"test-model-1", "test-model-2"},
		DefaultLatency: 10 * time.Millisecond,
	}

	provider := NewMockProvider(config)
	defer provider.Close()

	// Test Name method
	if provider.Name() != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got '%s'", provider.Name())
	}

	// Test Models method
	ctx := context.Background()
	models, err := provider.Models(ctx)
	if err != nil {
		t.Fatalf("Models() failed: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	// Test Complete method with default response
	req := inference.Request{
		Prompt: "Test prompt",
		Model:  "test-model-1",
	}

	resp, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete() failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Model != req.Model {
		t.Errorf("Expected model '%s', got '%s'", req.Model, resp.Model)
	}

	// Test Stream method
	chunks, err := provider.Stream(ctx, req)
	if err != nil {
		t.Fatalf("Stream() failed: %v", err)
	}

	// Collect all chunks
	var allChunks []inference.StreamChunk
	for chunk := range chunks {
		allChunks = append(allChunks, chunk)
		if chunk.Done {
			break
		}
	}

	if len(allChunks) < 2 { // Should have content chunks + done chunk
		t.Errorf("Expected at least 2 chunks, got %d", len(allChunks))
	}

	// Verify the last chunk is done
	lastChunk := allChunks[len(allChunks)-1]
	if !lastChunk.Done {
		t.Error("Expected last chunk to be done")
	}
}

func TestMockProviderCustomResponses(t *testing.T) {
	provider := NewMockProvider(MockProviderConfig{Name: "custom-test"})
	defer provider.Close()

	// Set custom response
	customResponse := &inference.Response{
		Content: "Custom response content",
		Model:   "custom-model",
		TokensUsed: inference.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	prompt := "Custom prompt"
	provider.SetResponse(prompt, customResponse)

	// Test the custom response
	ctx := context.Background()
	req := inference.Request{Prompt: prompt}

	resp, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete() failed: %v", err)
	}

	if resp.Content != customResponse.Content {
		t.Errorf("Expected content '%s', got '%s'", customResponse.Content, resp.Content)
	}
	if resp.TokensUsed.TotalTokens != customResponse.TokensUsed.TotalTokens {
		t.Errorf("Expected total tokens %d, got %d",
			customResponse.TokensUsed.TotalTokens, resp.TokensUsed.TotalTokens)
	}
}

func TestMockProviderDeterministicResponse(t *testing.T) {
	provider := NewMockProvider(MockProviderConfig{Name: "det", Deterministic: true})
	defer provider.Close()

	ctx := context.Background()
	req := inference.Request{Prompt: "same", Model: "m"}
	first, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("first Complete: %v", err)
	}
	second, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("second Complete: %v", err)
	}
	if first.Metadata["request_id"] != second.Metadata["request_id"] {
		t.Fatalf("request IDs differ: %v vs %v", first.Metadata["request_id"], second.Metadata["request_id"])
	}
	if first.Content != second.Content {
		t.Fatalf("content differs: %q vs %q", first.Content, second.Content)
	}
}

func TestMockProviderErrorHandling(t *testing.T) {
	provider := NewMockProvider(MockProviderConfig{Name: "error-test"})
	defer provider.Close()

	// Set error for Complete operation
	testErr := fmt.Errorf("simulated error")
	provider.SetError("Complete", testErr)

	ctx := context.Background()
	req := inference.Request{Prompt: "test"}

	_, err := provider.Complete(ctx, req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err != testErr {
		t.Errorf("Expected specific error, got %v", err)
	}
}

func TestMockProviderCallCounting(t *testing.T) {
	provider := NewMockProvider(MockProviderConfig{Name: "count-test"})
	defer provider.Close()

	ctx := context.Background()
	req := inference.Request{Prompt: "test"}

	// Make several calls
	provider.Complete(ctx, req)
	provider.Complete(ctx, req)
	provider.Models(ctx)

	// Check call counts
	completeCount := provider.GetCallCount("Complete")
	if completeCount != 2 {
		t.Errorf("Expected 2 Complete calls, got %d", completeCount)
	}

	modelsCount := provider.GetCallCount("Models")
	if modelsCount != 1 {
		t.Errorf("Expected 1 Models call, got %d", modelsCount)
	}
}

func TestAdvancedMockProvider(t *testing.T) {
	config := MockProviderConfig{
		Name:   "advanced-test",
		Models: []string{"advanced-model"},
	}

	provider := NewAdvancedMockProvider(config)
	defer provider.Close()

	// Test with failure rate
	provider.SetFailureRate(0.5) // 50% failure rate
	provider.SetRandomResponses(true)

	ctx := context.Background()
	req := inference.Request{Prompt: "test"}

	// Make multiple requests to test random behavior
	successCount := 0
	errorCount := 0

	for i := 0; i < 20; i++ {
		_, err := provider.Complete(ctx, req)
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// With 50% failure rate, we should have some successes and failures
	if successCount == 0 {
		t.Error("Expected some successful requests")
	}
	if errorCount == 0 {
		t.Error("Expected some failed requests")
	}

	// Check metrics
	metrics := provider.GetMetrics()
	if metrics.TotalRequests != 20 {
		t.Errorf("Expected 20 total requests, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulRequests+metrics.FailedRequests != metrics.TotalRequests {
		t.Error("Successful + failed requests should equal total requests")
	}
}

func TestChaosProvider(t *testing.T) {
	config := MockProviderConfig{Name: "chaos-test"}
	chaosConfig := ChaosConfig{
		CorruptionProbability:   0.1, // Low corruption rate for testing
		SlowResponseProbability: 0.1, // Low slow response rate
		SlowResponseDelay:       10 * time.Millisecond,
	}

	provider := NewChaosProvider(config, chaosConfig)
	defer provider.Close()

	ctx := context.Background()
	req := inference.Request{Prompt: "chaos test"}

	// Make a request - should work most of the time
	resp, err := provider.Complete(ctx, req)

	// Either we get a valid response or a chaos-induced error
	if err == nil && resp == nil {
		t.Error("Expected either response or error")
	}
}

func TestProviderInterfaceCompliance(t *testing.T) {
	// Test that all mock providers implement the interface correctly
	var providers []inference.Provider

	providers = append(providers, NewMockProvider(MockProviderConfig{Name: "basic"}))
	providers = append(providers, NewAdvancedMockProvider(MockProviderConfig{Name: "advanced"}))
	providers = append(providers, NewChaosProvider(MockProviderConfig{Name: "chaos"}, ChaosConfig{}))

	ctx := context.Background()
	req := inference.Request{Prompt: "interface test", Model: "test-model"}

	for i, provider := range providers {
		t.Run(provider.Name(), func(t *testing.T) {
			// Test all interface methods
			name := provider.Name()
			if name == "" {
				t.Errorf("Provider %d returned empty name", i)
			}

			models, err := provider.Models(ctx)
			if err != nil {
				t.Errorf("Provider %d Models() failed: %v", i, err)
			}
			if len(models) == 0 {
				t.Errorf("Provider %d returned no models", i)
			}

			resp, err := provider.Complete(ctx, req)
			if err != nil {
				t.Logf("Provider %d Complete() error (may be intentional): %v", i, err)
			} else if resp == nil {
				t.Errorf("Provider %d returned nil response without error", i)
			}

			chunks, err := provider.Stream(ctx, req)
			if err != nil {
				t.Logf("Provider %d Stream() error (may be intentional): %v", i, err)
			} else {
				// Drain the channel
				for range chunks {
				}
			}

			err = provider.Close()
			if err != nil {
				t.Errorf("Provider %d Close() failed: %v", i, err)
			}
		})
	}
}
