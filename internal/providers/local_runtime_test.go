package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

func TestOllamaProviderConfigPropagation(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")

	var got struct {
		Model   string                 `json:"model"`
		Prompt  string                 `json:"prompt"`
		Raw     bool                   `json:"raw"`
		Options map[string]interface{} `json:"options"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() failed: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"model":             got.Model,
			"response":          "ok",
			"done":              true,
			"total_duration":    12_000_000,
			"prompt_eval_count": 11,
			"eval_count":        7,
		})
	}))
	defer server.Close()

	provider, err := llm.GetProviderWithOptions("ollama:test-model", map[string]interface{}{
		"base_url":    server.URL,
		"raw":         true,
		"seed":        1,
		"num_predict": 100,
	})
	if err != nil {
		t.Fatalf("GetProviderWithOptions() failed: %v", err)
	}

	resp, err := provider.EvaluatePrompt(context.Background(), "exact prompt", nil)
	if err != nil {
		t.Fatalf("EvaluatePrompt() failed: %v", err)
	}

	if got.Model != "test-model" {
		t.Fatalf("request model = %q, want test-model", got.Model)
	}
	if got.Prompt != "exact prompt" {
		t.Fatalf("request prompt = %q, want exact prompt", got.Prompt)
	}
	if !got.Raw {
		t.Fatal("request raw = false, want true")
	}
	if got.Options["seed"] != float64(1) {
		t.Fatalf("request seed = %#v, want 1", got.Options["seed"])
	}
	if got.Options["num_predict"] != float64(100) {
		t.Fatalf("request num_predict = %#v, want 100", got.Options["num_predict"])
	}
	if resp.TokenUsage == nil || resp.TokenUsage.Total != 18 {
		t.Fatalf("response token usage = %#v, want total 18", resp.TokenUsage)
	}
	if resp.LatencyMs != 12 {
		t.Fatalf("response latency = %d, want 12", resp.LatencyMs)
	}
}

func TestMLXGoPresetConfigPropagation(t *testing.T) {
	binDir := t.TempDir()
	stubPath := filepath.Join(binDir, "mlx-lm-generate")
	writeExecutable(t, stubPath, `#!/bin/sh
printf "%s\n" "$@"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	provider, err := llm.GetProviderWithOptions("mlx-go-lm:test-model", map[string]interface{}{
		"args":       []string{"--ignore-chat-template", "--kv-size", "4096"},
		"max_tokens": 123,
	})
	if err != nil {
		t.Fatalf("GetProviderWithOptions() failed: %v", err)
	}

	resp, err := provider.Generate(context.Background(), "prompt text", llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	output := resp.Text
	for _, want := range []string{"--ignore-chat-template", "--kv-size", "4096", "--model", "test-model", "--prompt", "prompt text", "--max-tokens", "123"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}

func TestMLXGoPresetExplicitCommandOverridesDefaults(t *testing.T) {
	binDir := t.TempDir()
	stubPath := filepath.Join(binDir, "mlx-explicit")
	writeExecutable(t, stubPath, `#!/bin/sh
printf "%s\n" "$@"
`)

	provider, err := llm.GetProviderWithOptions("mlx-go-lm:test-model", map[string]interface{}{
		"executable": stubPath,
		"args": []string{
			"--model",
			"{{.Model}}",
			"--prompt",
			"{{.Prompt}}",
			"--raw-mode",
		},
	})
	if err != nil {
		t.Fatalf("GetProviderWithOptions() failed: %v", err)
	}

	resp, err := provider.Generate(context.Background(), "prompt text", llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	output := resp.Text
	if strings.Contains(output, "--max-tokens") || strings.Contains(output, "--temperature") {
		t.Fatalf("output %q contains preset defaults that should have been overridden", output)
	}
	for _, want := range []string{"--model", "test-model", "--prompt", "prompt text", "--raw-mode"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}

func TestLlamaCPPPresetConfigPropagation(t *testing.T) {
	binDir := t.TempDir()
	stubPath := filepath.Join(binDir, "llama-cli")
	writeExecutable(t, stubPath, `#!/bin/sh
printf "%s\n" "$@"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	provider, err := llm.GetProviderWithOptions("llama.cpp:/models/test.gguf", map[string]interface{}{
		"args": []string{"--no-conversation", "-ngl", "999", "-t", "8", "-fa"},
	})
	if err != nil {
		t.Fatalf("GetProviderWithOptions() failed: %v", err)
	}

	resp, err := provider.Generate(context.Background(), "prompt text", llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	output := resp.Text
	for _, want := range []string{"--no-conversation", "-ngl", "999", "-t", "8", "-fa", "-m", "/models/test.gguf"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}
