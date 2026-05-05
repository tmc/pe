package providers

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/llm"
)

func TestNewMockProvider(t *testing.T) {
	// Test without test mode
	_, err := NewMockProvider("test-model", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock provider only available in test mode")

	// Test with test mode
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	provider, err := NewMockProvider("test-model", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, provider)
	assert.Equal(t, "test-model", provider.model)
	assert.NotNil(t, provider.evaluationCount)
}

func TestMockProvider_Metadata(t *testing.T) {
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	provider, err := NewMockProvider("test-model", map[string]interface{}{})
	require.NoError(t, err)

	assert.Equal(t, "mock", provider.Name())
	assert.Equal(t, "test-model", provider.Model())
	assert.True(t, provider.SupportsStreaming())
	assert.False(t, provider.SupportsBatch())
}

func TestMockProvider_Generate(t *testing.T) {
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	provider, err := NewMockProvider("test-model", map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "default response",
			prompt:   "Hello world",
			expected: "Mock response for: Hello world",
		},
		{
			name:     "math question",
			prompt:   "What is 2+2?",
			expected: "4",
		},
		{
			name:     "capital question",
			prompt:   "What is the capital of France?",
			expected: "<answer>The capital of France is Paris</answer>",
		},
		{
			name:     "meaning of life",
			prompt:   "Calculate the meaning of life",
			expected: "<result>42</result>\n<explanation>The answer to life, the universe, and everything</explanation>",
		},
		{
			name:     "list items",
			prompt:   "List some items",
			expected: "<data>\n  <item>First</item>\n  <item>Second</item>\n</data>",
		},
		{
			name:     "rating prompt",
			prompt:   "Rate the prompt effectiveness on a scale from 0.0 to 1.0",
			expected: "0.85",
		},
		{
			name:     "optimized prompt rating",
			prompt:   "Rate the Optimized prompt content effectiveness on a scale from 0.0 to 1.0",
			expected: "0.97",
		},
	}

	ctx := context.Background()
	options := llm.GenerateOptions{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := provider.Generate(ctx, tt.prompt, options)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tt.expected, resp.Text)
			assert.Equal(t, "test-model", resp.Model)
			assert.Equal(t, "stop", resp.FinishReason)
			// Token counts might be 0 for very short strings
			assert.GreaterOrEqual(t, resp.PromptTokens, 0)
			assert.GreaterOrEqual(t, resp.CompletionTokens, 0)
			assert.GreaterOrEqual(t, resp.TotalTokens, 0)
			assert.GreaterOrEqual(t, resp.Latency, 10*time.Millisecond)
		})
	}
}

func TestMockProvider_GenerateStream(t *testing.T) {
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	provider, err := NewMockProvider("test-model", map[string]interface{}{})
	require.NoError(t, err)

	ctx := context.Background()
	options := llm.GenerateOptions{}

	respChan, err := provider.GenerateStream(ctx, "test prompt", options)
	require.NoError(t, err)
	require.NotNil(t, respChan)

	var responses []*llm.StreamResponse
	for resp := range respChan {
		responses = append(responses, resp)
	}

	// Should have multiple chunks plus a done signal
	assert.Greater(t, len(responses), 1)

	// Last response should be marked as done
	lastResp := responses[len(responses)-1]
	assert.True(t, lastResp.Done)

	// Combine all text
	var fullText strings.Builder
	for _, resp := range responses {
		if !resp.Done && resp.Error == nil {
			fullText.WriteString(resp.Text)
		}
	}

	// Should contain the expected mock response
	assert.Contains(t, fullText.String(), "Mock streaming response for:")
	assert.Contains(t, fullText.String(), "test prompt")
}

func TestMockProvider_GenerateStream_Cancellation(t *testing.T) {
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	provider, err := NewMockProvider("test-model", map[string]interface{}{})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	options := llm.GenerateOptions{}

	respChan, err := provider.GenerateStream(ctx, "test prompt", options)
	require.NoError(t, err)

	// Wait until the stream goroutine has started before canceling. With a
	// buffered channel, immediate cancellation can race with normal completion.
	resp := <-respChan
	require.Nil(t, resp.Error)

	cancel()

	// Should receive an error response
	var errorReceived bool
	for resp := range respChan {
		if resp.Error != nil {
			errorReceived = true
			assert.Equal(t, context.Canceled, resp.Error)
		}
	}

	assert.True(t, errorReceived, "should have received cancellation error")
}

func TestMockProvider_EvaluatePrompt(t *testing.T) {
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	provider, err := NewMockProvider("test-model", map[string]interface{}{})
	require.NoError(t, err)

	ctx := context.Background()
	vars := map[string]interface{}{
		"var1": "value1",
	}

	resp, err := provider.EvaluatePrompt(ctx, "What is 2+2?", vars)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "4", resp.Output)
	assert.NotNil(t, resp.TokenUsage)
	// Token counts might be 0 for very short strings
	assert.GreaterOrEqual(t, resp.TokenUsage.Total, int32(0))
	assert.GreaterOrEqual(t, resp.TokenUsage.Prompt, int32(0))
	assert.GreaterOrEqual(t, resp.TokenUsage.Completion, int32(0))
	assert.False(t, resp.Cached)
}

func TestMockProvider_SemanticOptimization(t *testing.T) {
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	provider, err := NewMockProvider("test-model", map[string]interface{}{})
	require.NoError(t, err)

	ctx := context.Background()
	options := llm.GenerateOptions{}

	// Test progressive scoring for accuracy objective
	for i := 0; i < 8; i++ {
		resp, err := provider.Generate(ctx, "OBJECTIVE: maximize accuracy\nRate effectiveness 0.0 to 1.0", options)
		require.NoError(t, err)

		score := resp.Text
		switch i {
		case 0:
			assert.Equal(t, "0.65", score)
		case 1:
			assert.Equal(t, "0.75", score)
		case 2:
			assert.Equal(t, "0.82", score)
		case 3:
			assert.Equal(t, "0.86", score)
		case 4:
			assert.Equal(t, "0.88", score)
		case 5:
			assert.Equal(t, "0.90", score)
		case 6:
			assert.Equal(t, "0.91", score)
		case 7:
			assert.Equal(t, "0.9105", score)
		}
	}

	// Test semantic gradient response
	resp, err := provider.Generate(ctx, "Compute semantic gradient for prompt", options)
	require.NoError(t, err)
	// Should return the gradient JSON response
	assert.Contains(t, resp.Text, "gradients")
	assert.Contains(t, resp.Text, "component")
	assert.Contains(t, resp.Text, "direction")

	// Test gradient application without triggering the "semantic gradient" pattern
	resp, err = provider.Generate(ctx, "Apply the following gradients\nORIGINAL PROMPT: test prompt", options)
	require.NoError(t, err)
	// Should be default mock response since it doesn't match our specific patterns
	expectedResponse := "Mock response for: Apply the following gradients\nORIGINAL PROMPT: test prompt"
	assert.Equal(t, expectedResponse, resp.Text)
}

func TestMockProvider_Init(t *testing.T) {
	// Set test mode for entire test
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	// Test the factory function directly
	factory := func(model string, options map[string]interface{}) (llm.Provider, error) {
		return NewMockProvider(model, options)
	}

	registry := NewRegistry()
	registry.Register("mock", factory)
	assert.True(t, registry.Exists("mock"))

	// Test creating through registry
	provider, err := registry.Create("mock:test", map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}
