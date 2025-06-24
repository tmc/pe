package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/inference"
)

func TestFactory(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		envVar   string
		expected string
	}{
		{
			name:     "default URL",
			config:   nil,
			expected: "http://localhost:11434",
		},
		{
			name:     "config URL",
			config:   map[string]interface{}{"base_url": "http://custom:8080"},
			expected: "http://custom:8080",
		},
		{
			name:     "environment variable",
			envVar:   "http://env:9090",
			expected: "http://env:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("OLLAMA_HOST", tt.envVar)
			}

			provider, err := Factory(tt.config)
			require.NoError(t, err)
			require.NotNil(t, provider)

			ollamaProvider := provider.(*Provider)
			assert.Equal(t, tt.expected, ollamaProvider.baseURL)
			assert.Equal(t, "ollama", provider.Name())
		})
	}
}

func TestProvider_Complete(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/generate", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req generateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "llama2", req.Model)
		assert.Equal(t, "test prompt", req.Prompt)
		assert.Equal(t, "test system", req.System)
		assert.False(t, req.Stream)

		resp := generateResponse{
			Model:           "llama2",
			CreatedAt:       "2023-12-07T09:30:35.456789Z",
			Response:        "This is a test response",
			Done:            true,
			PromptEvalCount: 10,
			EvalCount:       15,
			TotalDuration:   1000000,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := New(server.URL)
	ctx := context.Background()

	req := inference.Request{
		Prompt:       "test prompt",
		Model:        "llama2",
		SystemPrompt: "test system",
		Temperature:  0.7,
		MaxTokens:    100,
	}

	resp, err := provider.Complete(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "This is a test response", resp.Content)
	assert.Equal(t, "llama2", resp.Model)
	assert.Equal(t, 10, resp.TokensUsed.PromptTokens)
	assert.Equal(t, 15, resp.TokensUsed.CompletionTokens)
	assert.Equal(t, 25, resp.TokensUsed.TotalTokens)
	assert.Equal(t, int64(1000000), resp.Metadata["total_duration"])
}

func TestProvider_Complete_Error(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	provider := New(server.URL)
	ctx := context.Background()

	req := inference.Request{
		Prompt: "test prompt",
		Model:  "llama2",
	}

	resp, err := provider.Complete(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ollama API error 500")
}

func TestProvider_Stream(t *testing.T) {
	// Mock server for streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/generate", r.URL.Path)

		var req generateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.True(t, req.Stream)

		w.Header().Set("Content-Type", "application/json")
		
		// Send multiple streaming responses
		responses := []generateResponse{
			{Response: "This", Done: false},
			{Response: " is", Done: false},
			{Response: " a test", Done: false},
			{Response: "", Done: true},
		}

		for _, resp := range responses {
			data, _ := json.Marshal(resp)
			w.Write(data)
			w.Write([]byte("\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	provider := New(server.URL)
	ctx := context.Background()

	req := inference.Request{
		Prompt: "test prompt",
		Model:  "llama2",
		Stream: true,
	}

	ch, err := provider.Stream(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, ch)

	var chunks []string
	var done bool
	for chunk := range ch {
		require.NoError(t, chunk.Error)
		if !chunk.Done {
			chunks = append(chunks, chunk.Delta)
		} else {
			done = true
		}
	}

	assert.True(t, done)
	assert.Equal(t, []string{"This", " is", " a test"}, chunks)
}

func TestProvider_Models(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/tags", r.URL.Path)

		resp := modelsResponse{
			Models: []struct {
				Name         string `json:"name"`
				ModifiedAt   string `json:"modified_at"`
				Size         int64  `json:"size"`
				Digest       string `json:"digest"`
				Details      struct {
					Format            string   `json:"format"`
					Family            string   `json:"family"`
					Families          []string `json:"families"`
					ParameterSize     string   `json:"parameter_size"`
					QuantizationLevel string   `json:"quantization_level"`
				} `json:"details"`
			}{
				{Name: "llama2:latest"},
				{Name: "codellama:7b"},
				{Name: "mistral:latest"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := New(server.URL)
	ctx := context.Background()

	models, err := provider.Models(ctx)
	require.NoError(t, err)
	require.NotNil(t, models)

	expected := []string{"llama2:latest", "codellama:7b", "mistral:latest"}
	assert.Equal(t, expected, models)
}

func TestProvider_Models_Error(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	provider := New(server.URL)
	ctx := context.Background()

	models, err := provider.Models(ctx)
	assert.Error(t, err)
	assert.Nil(t, models)
	assert.Contains(t, err.Error(), "ollama API error 400")
}

func TestProvider_DefaultModel(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req generateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Should default to llama2 when no model specified
		assert.Equal(t, "llama2", req.Model)

		resp := generateResponse{
			Model:    "llama2",
			Response: "test",
			Done:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := New(server.URL)
	ctx := context.Background()

	req := inference.Request{
		Prompt: "test prompt",
		// No model specified - should default to llama2
	}

	resp, err := provider.Complete(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "llama2", resp.Model)
}

func TestProvider_Options(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req generateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Check that options are properly set (JSON unmarshaling converts to float64 and []interface{})
		assert.Equal(t, float64(0.8), req.Options["temperature"])
		assert.Equal(t, float64(150), req.Options["num_predict"])
		assert.Equal(t, []interface{}{"<|endoftext|>"}, req.Options["stop"])
		assert.Equal(t, "custom_value", req.Options["custom_option"])

		resp := generateResponse{
			Model:    "llama2",
			Response: "test",
			Done:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := New(server.URL)
	ctx := context.Background()

	req := inference.Request{
		Prompt:        "test prompt",
		Model:         "llama2",
		Temperature:   0.8,
		MaxTokens:     150,
		StopSequences: []string{"<|endoftext|>"},
		Options: map[string]interface{}{
			"custom_option": "custom_value",
		},
	}

	_, err := provider.Complete(ctx, req)
	require.NoError(t, err)
}

func TestProvider_Close(t *testing.T) {
	provider := New("http://localhost:11434")
	err := provider.Close()
	assert.NoError(t, err)
}

func TestProvider_ContextTimeout(t *testing.T) {
	// Mock server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		resp := generateResponse{Response: "test", Done: true}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := New(server.URL)
	
	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req := inference.Request{
		Prompt: "test prompt",
		Model:  "llama2",
	}

	resp, err := provider.Complete(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, strings.ToLower(err.Error()), "context deadline exceeded")
}