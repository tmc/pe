package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/inference"
)

func TestAnthropicProvider(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected auth header, got %s", r.Header.Get("x-api-key"))
		}

		// Check version header
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected version header, got %s", r.Header.Get("anthropic-version"))
		}

		switch r.URL.Path {
		case "/v1/messages":
			// Decode request
			var req messageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}

			// Check for stream
			if req.Stream {
				// Send SSE response
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)

				// Send chunks
				chunks := []string{"Hello", " from", " Anthropic"}
				for _, chunk := range chunks {
					// Send event type
					w.Write([]byte("event: content_block_delta\n"))

					// Send data
					data := contentBlockDelta{
						Type:  "content_block_delta",
						Index: 0,
						Delta: struct {
							Type string `json:"type"`
							Text string `json:"text"`
						}{
							Type: "text_delta",
							Text: chunk,
						},
					}

					jsonData, _ := json.Marshal(data)
					w.Write([]byte("data: " + string(jsonData) + "\n\n"))
					w.(http.Flusher).Flush()
				}

				// Send done
				w.Write([]byte("event: message_stop\n"))
				w.Write([]byte("data: {\"type\":\"message_stop\"}\n\n"))
				w.(http.Flusher).Flush()
			} else {
				// Send normal response
				resp := messageResponse{
					ID:   "test",
					Type: "message",
					Role: "assistant",
					Content: []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{
						{
							Type: "text",
							Text: "Hello from Anthropic",
						},
					},
					Model:      "claude-3-sonnet-20240229",
					StopReason: "end_turn",
					Usage: struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
					}{
						InputTokens:  10,
						OutputTokens: 5,
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
			Model:        "claude-3-sonnet-20240229",
			Temperature:  0.7,
			MaxTokens:    100,
			SystemPrompt: "You are a helpful assistant",
		}

		resp, err := provider.Complete(ctx, req)
		if err != nil {
			t.Fatalf("failed to complete: %v", err)
		}

		if resp.Content != "Hello from Anthropic" {
			t.Errorf("expected 'Hello from Anthropic', got %s", resp.Content)
		}

		if resp.TokensUsed.TotalTokens != 15 {
			t.Errorf("expected 15 total tokens, got %d", resp.TokensUsed.TotalTokens)
		}
	})

	t.Run("Stream", func(t *testing.T) {
		ctx := context.Background()
		req := inference.Request{
			Prompt:      "Hello",
			Model:       "claude-3-sonnet-20240229",
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

		if result.String() != "Hello from Anthropic" {
			t.Errorf("expected 'Hello from Anthropic', got %s", result.String())
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

		if p.Name() != "anthropic" {
			t.Errorf("expected name 'anthropic', got %s", p.Name())
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

	t.Run("DefaultMaxTokens", func(t *testing.T) {
		ctx := context.Background()
		req := inference.Request{
			Prompt: "Hello",
			// MaxTokens not set - should default to 4096
		}

		// This test just ensures no panic when MaxTokens is 0
		_, err := provider.Complete(ctx, req)
		if err != nil {
			t.Fatalf("failed with default max tokens: %v", err)
		}
	})
}
