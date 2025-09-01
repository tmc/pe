package cgpt

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/inference"
)

func TestProvider_Name(t *testing.T) {
	p := New()
	assert.Equal(t, "cgpt", p.Name())
}

func TestProvider_Models(t *testing.T) {
	p := New()
	models, err := p.Models(context.Background())
	require.NoError(t, err)
	
	// Check for expected models
	expectedModels := []string{
		"claude-sonnet-4-20250514",
		"gpt-4o",
		"gpt-4",
		"claude-3-7-sonnet-20250219",
		"gemini-2.0-flash",
	}
	
	for _, expected := range expectedModels {
		assert.Contains(t, models, expected)
	}
}

func TestProvider_BuildArgs(t *testing.T) {
	p := New()
	
	tests := []struct {
		name string
		req  inference.Request
	}{
		{
			name: "basic gpt-4 request",
			req: inference.Request{
				Model:  "gpt-4",
				Prompt: "Hello, world!",
			},
		},
		{
			name: "claude request with temperature",
			req: inference.Request{
				Model:       "claude-3-sonnet",
				Prompt:      "Explain AI",
				Temperature: 0.7,
			},
		},
		{
			name: "gemini request with max tokens",
			req: inference.Request{
				Model:     "gemini-2.0-flash",
				Prompt:    "Write code",
				MaxTokens: 1000,
			},
		},
		{
			name: "request with system prompt",
			req: inference.Request{
				Model:        "gpt-4o",
				Prompt:       "Hello",
				SystemPrompt: "You are a helpful assistant",
			},
		},
		{
			name: "request with prefill",
			req: inference.Request{
				Model:   "claude-3-sonnet",
				Prompt:  "Complete this:",
				Prefill: "Sure, I'll",
			},
		},
		{
			name: "request with stop sequences",
			req: inference.Request{
				Model:         "gpt-4",
				Prompt:        "Count to 10",
				StopSequences: []string{"5", "STOP"},
			},
		},
		{
			name: "request with options",
			req: inference.Request{
				Model:  "gpt-4",
				Prompt: "Hello",
				Options: map[string]interface{}{
					"verbose":            true,
					"debug":              true,
					"stream":             false,
					"completion_timeout": "30s",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := p.buildArgs(tt.req)
			argsStr := strings.Join(args, " ")
			
			// Check key expectations based on request
			if tt.req.Model != "" {
				assert.Contains(t, argsStr, "--model "+tt.req.Model)
			}
			if tt.req.Temperature > 0 {
				assert.Contains(t, argsStr, "--temperature")
			}
			if tt.req.MaxTokens > 0 {
				assert.Contains(t, argsStr, "--max-tokens")
			}
			if tt.req.SystemPrompt != "" {
				assert.Contains(t, argsStr, "--system-prompt")
			}
			if tt.req.Prefill != "" {
				assert.Contains(t, argsStr, "--prefill")
			}
			if len(tt.req.StopSequences) > 0 {
				for _, stop := range tt.req.StopSequences {
					assert.Contains(t, argsStr, "--stop "+stop)
				}
			}
			if tt.req.Prompt != "" {
				assert.Contains(t, argsStr, "--input "+tt.req.Prompt)
			}
			
			// Check backend detection
			if strings.HasPrefix(tt.req.Model, "gpt-") {
				assert.Contains(t, argsStr, "--backend openai")
			} else if strings.HasPrefix(tt.req.Model, "claude-") {
				assert.Contains(t, argsStr, "--backend anthropic")
			} else if strings.HasPrefix(tt.req.Model, "gemini-") {
				assert.Contains(t, argsStr, "--backend googleai")
			}
			
			// Check options
			if tt.req.Options != nil {
				if v, ok := tt.req.Options["verbose"].(bool); ok && v {
					assert.Contains(t, argsStr, "--verbose")
				}
				if v, ok := tt.req.Options["debug"].(bool); ok && v {
					assert.Contains(t, argsStr, "--debug")
				}
				if v, ok := tt.req.Options["stream"].(bool); ok && !v {
					assert.Contains(t, argsStr, "--stream=false")
				}
				if v, ok := tt.req.Options["completion_timeout"].(string); ok && v != "" {
					assert.Contains(t, argsStr, "--completion-timeout "+v)
				}
			}
		})
	}
}

func TestProvider_Complete_MockMode(t *testing.T) {
	// Set test mode
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")
	
	p := New()
	
	req := inference.Request{
		Model:  "gpt-4",
		Prompt: "What is 2+2?",
	}
	
	resp, err := p.Complete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	
	assert.Equal(t, "4", resp.Content)
	assert.Equal(t, "gpt-4", resp.Model)
	assert.Equal(t, 5, resp.TokensUsed.TotalTokens)
	assert.Equal(t, "cgpt", resp.Metadata["provider"])
	assert.True(t, resp.Metadata["mock"].(bool))
}

func TestProvider_Stream_MockMode(t *testing.T) {
	// Set test mode
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")
	
	p := New()
	
	req := inference.Request{
		Model:  "gpt-4",
		Prompt: "Count to 5",
	}
	
	chunks, err := p.Stream(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, chunks)
	
	var receivedChunks []inference.StreamChunk
	for chunk := range chunks {
		if chunk.Error != nil {
			t.Fatalf("Unexpected error in stream: %v", chunk.Error)
		}
		receivedChunks = append(receivedChunks, chunk)
		if chunk.Done {
			break
		}
	}
	
	// Should receive multiple chunks plus done signal
	assert.True(t, len(receivedChunks) > 1)
	assert.True(t, receivedChunks[len(receivedChunks)-1].Done)
}

func TestProvider_MockResponse_VariousPrompts(t *testing.T) {
	// Set test mode
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")
	
	p := New()
	
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "math question",
			prompt:   "What is 2+2?",
			expected: "4",
		},
		{
			name:     "translation request",
			prompt:   "Translate Hello to Spanish",
			expected: "Hola",
		},
		{
			name:     "programming question",
			prompt:   "Explain what a pointer is in one sentence.",
			expected: "A pointer is a variable that stores the memory address of another variable.",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := inference.Request{
				Model:  "gpt-4",
				Prompt: tt.prompt,
			}
			
			resp, err := p.Complete(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, resp.Content)
		})
	}
}

func TestProvider_MockResponse_WithPrefillAndStopSequences(t *testing.T) {
	// Set test mode
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")
	
	p := New()
	
	req := inference.Request{
		Model:         "claude-3-sonnet",
		Prompt:        "Count to 10",
		Prefill:       "Sure! ",
		StopSequences: []string{"5"},
	}
	
	resp, err := p.Complete(context.Background(), req)
	require.NoError(t, err)
	
	// Should include prefill and be truncated by stop sequence
	assert.Contains(t, resp.Content, "Sure!")
	// Should be truncated before reaching the stop sequence
	assert.NotContains(t, resp.Content, "5")
}

func TestProvider_ParseJSONResponse(t *testing.T) {
	jsonContent := `{"response": "test", "tokens": 10, "model": "gpt-4"}`
	
	result, err := ParseJSONResponse(jsonContent)
	require.NoError(t, err)
	
	assert.Equal(t, "test", result["response"])
	assert.Equal(t, float64(10), result["tokens"])
	assert.Equal(t, "gpt-4", result["model"])
}

func TestProvider_ParseJSONResponse_Invalid(t *testing.T) {
	invalidJSON := `{"response": "test", "incomplete"`
	
	_, err := ParseJSONResponse(invalidJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON response")
}

func TestProvider_Close(t *testing.T) {
	p := New()
	err := p.Close()
	assert.NoError(t, err)
}

func TestNewWithBinary(t *testing.T) {
	p := NewWithBinary("/custom/path/to/cgpt")
	assert.Equal(t, "/custom/path/to/cgpt", p.binaryPath)
	assert.False(t, p.useGoRun)
}

func TestProvider_Factory(t *testing.T) {
	// Test default factory
	provider, err := inference.NewProvider("cgpt", nil)
	require.NoError(t, err)
	assert.Equal(t, "cgpt", provider.Name())
	
	// Test factory with binary config
	config := map[string]interface{}{
		"binary": "/custom/cgpt",
	}
	provider, err = inference.NewProvider("cgpt", config)
	require.NoError(t, err)
	assert.Equal(t, "cgpt", provider.Name())
}