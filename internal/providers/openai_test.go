package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

func TestNewOpenAIProviderUsesInferenceProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %q, want /chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q", got)
		}
		var req struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "gpt-4" {
			t.Fatalf("model = %q, want gpt-4", req.Model)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{
				"message":       map[string]string{"role": "assistant", "content": "hello"},
				"finish_reason": "stop",
			}},
			"usage": map[string]int{"prompt_tokens": 2, "completion_tokens": 3, "total_tokens": 5},
			"model": "gpt-4",
		})
	}))
	defer server.Close()

	provider, err := NewOpenAIProvider("gpt-4", map[string]interface{}{
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
	if resp.Text != "hello" || resp.TotalTokens != 5 || resp.Model != "gpt-4" {
		t.Fatalf("response = %#v", resp)
	}
}

func TestNewOpenAIProviderRequiresKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	provider, err := NewOpenAIProvider("gpt-4", nil)
	if err == nil || provider != nil {
		t.Fatalf("NewOpenAIProvider() = %#v, %v; want key error", provider, err)
	}
}
