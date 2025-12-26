package providers

import (
	"context"
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
		{"ollama"},
		{"mlx"},
		{"llama-cpp"},
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
