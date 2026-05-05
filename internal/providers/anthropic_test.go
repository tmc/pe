package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

func TestNewAnthropicProviderUsesInferenceProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("path = %q, want /v1/messages", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("x-api-key = %q", got)
		}
		var req struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "claude-3-haiku" {
			t.Fatalf("model = %q, want claude-3-haiku", req.Model)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]string{{"type": "text", "text": "hello"}},
			"model":   "claude-3-haiku",
			"usage":   map[string]int{"input_tokens": 2, "output_tokens": 3},
		})
	}))
	defer server.Close()

	provider, err := NewAnthropicProvider("claude-3-haiku", map[string]interface{}{
		"apiKey":  "test-key",
		"baseURL": server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := provider.Generate(context.Background(), "hi", llm.GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != "hello" || resp.TotalTokens != 5 || resp.Model != "claude-3-haiku" {
		t.Fatalf("response = %#v", resp)
	}
}

func TestNewAnthropicProviderRequiresKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	provider, err := NewAnthropicProvider("claude-3-haiku", nil)
	if err == nil || provider != nil {
		t.Fatalf("NewAnthropicProvider() = %#v, %v; want key error", provider, err)
	}
}
