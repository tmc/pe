package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/llm"
)

func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		options map[string]interface{}
		envKey  string
		wantErr bool
	}{
		{
			name:  "with api key in options",
			model: "gpt-4",
			options: map[string]interface{}{
				"apiKey": "test-key",
			},
			wantErr: false,
		},
		{
			name:    "with api key in env",
			model:   "gpt-4",
			options: map[string]interface{}{},
			envKey:  "test-env-key",
			wantErr: false,
		},
		{
			name:    "no api key",
			model:   "gpt-4",
			options: map[string]interface{}{},
			wantErr: true,
		},
		{
			name:  "custom base URL",
			model: "gpt-4",
			options: map[string]interface{}{
				"apiKey":  "test-key",
				"baseURL": "https://custom.api.com/v1/",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv("OPENAI_API_KEY", tt.envKey)
				defer os.Unsetenv("OPENAI_API_KEY")
			}

			provider, err := NewOpenAIProvider(tt.model, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, tt.model, provider.model)
				assert.Equal(t, "openai", provider.Name())
			}
		})
	}
}

func TestOpenAIProvider_Generate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		var req OpenAIRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Send response
		resp := OpenAIResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Test response for: " + req.Messages[0].Content,
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider, err := NewOpenAIProvider("gpt-4", map[string]interface{}{
		"apiKey":  "test-key",
		"baseURL": server.URL + "/v1",
	})
	require.NoError(t, err)

	ctx := context.Background()
	temp := 0.7
	maxTokens := 100
	options := llm.GenerateOptions{
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	resp, err := provider.Generate(ctx, "Hello, world!", options)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "Test response for: Hello, world!", resp.Text)
	assert.Equal(t, 10, resp.PromptTokens)
	assert.Equal(t, 5, resp.CompletionTokens)
	assert.Equal(t, 15, resp.TotalTokens)
	assert.Equal(t, "gpt-4", resp.Model)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.Greater(t, resp.Cost, 0.0)
}

func TestOpenAIProvider_Generate_Error(t *testing.T) {
	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			}{
				Message: "Invalid request",
				Type:    "invalid_request_error",
				Code:    "invalid_api_key",
			},
		})
	}))
	defer server.Close()

	provider, err := NewOpenAIProvider("gpt-4", map[string]interface{}{
		"apiKey":  "test-key",
		"baseURL": server.URL + "/v1",
	})
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := provider.Generate(ctx, "Hello", llm.GenerateOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid request")
	assert.Nil(t, resp)
}

func TestOpenAIProvider_GenerateStream(t *testing.T) {
	// Create test server for streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify streaming request
		var req OpenAIRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.True(t, req.Stream)

		// Send SSE response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		// Send streaming chunks
		chunks := []string{"Hello", ", ", "world", "!"}
		for i, chunk := range chunks {
			data := struct {
				ID      string `json:"id"`
				Object  string `json:"object"`
				Created int64  `json:"created"`
				Model   string `json:"model"`
				Choices []struct {
					Index int `json:"index"`
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}{
				ID:      "test-id",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "gpt-4",
				Choices: []struct {
					Index int `json:"index"`
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				}{
					{
						Index: 0,
						Delta: struct {
							Content string `json:"content"`
						}{
							Content: chunk,
						},
					},
				},
			}

			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()

			// Small delay between chunks
			if i < len(chunks)-1 {
				time.Sleep(10 * time.Millisecond)
			}
		}

		// Send done signal
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	provider, err := NewOpenAIProvider("gpt-4", map[string]interface{}{
		"apiKey":  "test-key",
		"baseURL": server.URL + "/v1",
	})
	require.NoError(t, err)

	ctx := context.Background()
	stream, err := provider.GenerateStream(ctx, "Hello", llm.GenerateOptions{})
	require.NoError(t, err)
	require.NotNil(t, stream)

	// Collect stream responses
	var texts []string
	var done bool
	for resp := range stream {
		if resp.Error != nil {
			t.Fatalf("unexpected error: %v", resp.Error)
		}
		if resp.Done {
			done = true
			break
		}
		texts = append(texts, resp.Text)
	}

	assert.True(t, done)
	assert.Equal(t, []string{"Hello", ", ", "world", "!"}, texts)
}

func TestOpenAIProvider_EvaluatePrompt(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OpenAIResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Test response",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider, err := NewOpenAIProvider("gpt-4", map[string]interface{}{
		"apiKey":  "test-key",
		"baseURL": server.URL + "/v1",
	})
	require.NoError(t, err)

	ctx := context.Background()
	vars := map[string]interface{}{
		"temperature": 0.5,
		"max_tokens":  50,
	}

	resp, err := provider.EvaluatePrompt(ctx, "Test prompt", vars)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "Test response", resp.Output)
	assert.Equal(t, int32(15), resp.TokenUsage.Total)
	assert.Equal(t, int32(10), resp.TokenUsage.Prompt)
	assert.Equal(t, int32(5), resp.TokenUsage.Completion)
	assert.Greater(t, resp.Cost, 0.0)
}

func TestCalculateOpenAICost(t *testing.T) {
	tests := []struct {
		model            string
		promptTokens     int
		completionTokens int
		expectedMin      float64
		expectedMax      float64
	}{
		{
			model:            "gpt-4",
			promptTokens:     1000,
			completionTokens: 500,
			expectedMin:      0.05,
			expectedMax:      0.06,
		},
		{
			model:            "gpt-3.5-turbo",
			promptTokens:     1000,
			completionTokens: 500,
			expectedMin:      0.002,
			expectedMax:      0.003,
		},
		{
			model:            "unknown-model",
			promptTokens:     1000,
			completionTokens: 500,
			expectedMin:      0.05, // Should default to GPT-4 pricing
			expectedMax:      0.06,
		},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			cost := calculateOpenAICost(tt.model, tt.promptTokens, tt.completionTokens)
			assert.GreaterOrEqual(t, cost, tt.expectedMin)
			assert.LessOrEqual(t, cost, tt.expectedMax)
		})
	}
}

func TestOpenAIProvider_Retry(t *testing.T) {
	attempts := 0
	// Create test server that fails first time, succeeds second time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := OpenAIResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Success after retry",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider, err := NewOpenAIProvider("gpt-4", map[string]interface{}{
		"apiKey":     "test-key",
		"baseURL":    server.URL + "/v1",
		"maxRetries": 2,
	})
	require.NoError(t, err)

	// Override timeout for faster test
	provider.client.Timeout = 5 * time.Second

	ctx := context.Background()
	resp, err := provider.Generate(ctx, "Test", llm.GenerateOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "Success after retry", resp.Text)
	assert.Equal(t, 2, attempts)
}

func TestOpenAIProvider_SupportsFeatures(t *testing.T) {
	provider, err := NewOpenAIProvider("gpt-4", map[string]interface{}{
		"apiKey": "test-key",
	})
	require.NoError(t, err)

	assert.True(t, provider.SupportsStreaming())
	assert.False(t, provider.SupportsBatch())
	assert.Equal(t, "openai", provider.Name())
	assert.Equal(t, "gpt-4", provider.Model())
}
