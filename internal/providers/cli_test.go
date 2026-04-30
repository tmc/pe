package providers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

func TestGenericCLIProvider_Template(t *testing.T) {
	// Note: We can't easily test actual execution without mocking exec.CommandContext,
	// which is hard in Go without dependency injection.
	// However, we can test the provider initialization and options.

	tests := []struct {
		name    string
		options map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid command",
			options: map[string]interface{}{
				"command": "echo {{.Prompt}}",
			},
			wantErr: false,
		},
		{
			name:    "missing command",
			options: map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGenericCLIProvider("test-model", tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenericCLIProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPresets(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"mlx"},
		{"mlx-lm"},
		{"mlx-go"},
		{"mlx-go-lm"},
		{"llama-cpp"},
		{"llama.cpp"},
		{"llm-tool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := GetPresetConfig(tt.name, nil)
			if err != nil {
				t.Errorf("GetPresetConfig() error = %v", err)
				return
			}
			if config["command"] == "" {
				t.Error("preset config missing command")
			}
		})
	}
}

func TestGenericCLIProvider_Generate_Mock(t *testing.T) {
	// Using 'echo' as a mock LLM
	options := map[string]interface{}{
		"command": "echo {{.Prompt}}",
	}
	p, err := NewGenericCLIProvider("echo-model", options)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()
	prompt := "hello world"
	resp, err := p.Generate(ctx, prompt, llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if strings.TrimSpace(resp.Text) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", resp.Text)
	}
}

func TestGenericCLIProvider_Generate_StructuredJSON(t *testing.T) {
	options := map[string]interface{}{
		"command": `sh -c 'printf '\''{"output":"hi","prompt_tokens":3,"completion_tokens":5,"total_tokens":8,"latency_ms":21,"metrics":{"tokens_per_second":99.5}}'\'''`,
	}
	p, err := NewGenericCLIProvider("echo-model", options)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	resp, err := p.Generate(context.Background(), "ignored", llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Text != "hi" {
		t.Fatalf("resp.Text = %q, want hi", resp.Text)
	}
	if resp.PromptTokens != 3 || resp.CompletionTokens != 5 || resp.TotalTokens != 8 {
		t.Fatalf("token counts = (%d,%d,%d), want (3,5,8)", resp.PromptTokens, resp.CompletionTokens, resp.TotalTokens)
	}
	if resp.Latency.Milliseconds() != 21 {
		t.Fatalf("latency = %dms, want 21ms", resp.Latency.Milliseconds())
	}
	if got := resp.Metadata["tokens_per_second"]; got != 99.5 {
		t.Fatalf("metadata[tokens_per_second] = %#v, want 99.5", got)
	}
}

func TestGenericCLIProvider_Generate_ArgvAndStdin(t *testing.T) {
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "stdin-stub"), `#!/bin/sh
cat
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	p, err := NewGenericCLIProvider("test-model", map[string]interface{}{
		"executable":   "stdin-stub",
		"args":         []string{},
		"prompt_stdin": true,
	})
	if err != nil {
		t.Fatalf("NewGenericCLIProvider() failed: %v", err)
	}

	prompt := "line 1\nline 2\n"
	resp, err := p.Generate(context.Background(), prompt, llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	if resp.Text != "line 1\nline 2" {
		t.Fatalf("resp.Text = %q, want exact stdin bytes without trailing trim mismatch", resp.Text)
	}
}
