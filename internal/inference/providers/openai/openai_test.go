package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tmc/pe/internal/inference"
)

func TestOpenAIProvider(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected auth header, got %s", r.Header.Get("Authorization"))
		}

		switch r.URL.Path {
		case "/chat/completions":
			// Decode request
			var req chatRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}

			// Check for stream
			if req.Stream {
				// Send SSE response
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)

				// Send chunks
				chunks := []string{"Hello", " from", " OpenAI"}
				for _, chunk := range chunks {
					data := streamChunk{
						ID:      "test",
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   "gpt-3.5-turbo",
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
					w.Write([]byte("data: " + string(jsonData) + "\n\n"))
					w.(http.Flusher).Flush()
				}

				// Send done
				w.Write([]byte("data: [DONE]\n\n"))
				w.(http.Flusher).Flush()
			} else {
				// Send normal response
				resp := chatResponse{
					ID:      "test",
					Object:  "chat.completion",
					Created: time.Now().Unix(),
					Model:   "gpt-3.5-turbo",
					Choices: []struct {
						Index        int         `json:"index"`
						Message      chatMessage `json:"message"`
						FinishReason string      `json:"finish_reason"`
					}{
						{
							Index: 0,
							Message: chatMessage{
								Role:    "assistant",
								Content: "Hello from OpenAI",
							},
							FinishReason: "stop",
						},
					},
					Usage: struct {
						PromptTokens     int `json:"prompt_tokens"`
						CompletionTokens int `json:"completion_tokens"`
						TotalTokens      int `json:"total_tokens"`
					}{
						PromptTokens:     10,
						CompletionTokens: 5,
						TotalTokens:      15,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create provider with test server
	provider := New("test-key", server.URL)

	t.Run("Complete", func(t *testing.T) {
		ctx := context.Background()
		req := inference.Request{
			Prompt:       "Hello",
			Model:        "gpt-3.5-turbo",
			Temperature:  0.7,
			MaxTokens:    100,
			SystemPrompt: "You are a helpful assistant",
		}

		resp, err := provider.Complete(ctx, req)
		if err != nil {
			t.Fatalf("failed to complete: %v", err)
		}

		if resp.Content != "Hello from OpenAI" {
			t.Errorf("expected 'Hello from OpenAI', got %s", resp.Content)
		}

		if resp.TokensUsed.TotalTokens != 15 {
			t.Errorf("expected 15 total tokens, got %d", resp.TokensUsed.TotalTokens)
		}
	})

	t.Run("Stream", func(t *testing.T) {
		ctx := context.Background()
		req := inference.Request{
			Prompt:      "Hello",
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
			MaxTokens:   100,
		}

		chunks, err := provider.Stream(ctx, req)
		if err != nil {
			t.Fatalf("failed to stream: %v", err)
		}

		var result strings.Builder
		for chunk := range chunks {
			if chunk.Error != nil {
				t.Fatalf("stream error: %v", chunk.Error)
			}
			if chunk.Done {
				break
			}
			result.WriteString(chunk.Delta)
		}

		if result.String() != "Hello from OpenAI" {
			t.Errorf("expected 'Hello from OpenAI', got %s", result.String())
		}
	})

	t.Run("Factory", func(t *testing.T) {
		// Test with API key in config
		p, err := Factory(map[string]interface{}{
			"api_key":  "test-key",
			"base_url": server.URL,
		})
		if err != nil {
			t.Fatalf("failed to create provider: %v", err)
		}

		if p.Name() != "openai" {
			t.Errorf("expected name 'openai', got %s", p.Name())
		}
	})

	t.Run("Models", func(t *testing.T) {
		models, err := provider.Models(context.Background())
		if err != nil {
			t.Fatalf("failed to get models: %v", err)
		}

		if len(models) == 0 {
			t.Error("expected at least one model")
		}
	})
}
