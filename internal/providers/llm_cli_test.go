package providers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

func TestLLMCLIProvider(t *testing.T) {
	binDir := t.TempDir()
	executable := filepath.Join(binDir, "llm")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\nprintf mocked\n"), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", executable, err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	p, err := NewLLMCLIProvider("gpt-4", nil)
	if err != nil {
		t.Fatalf("NewLLMCLIProvider() failed: %v", err)
	}

	if p.Name() != "llm" {
		t.Errorf("Name() = %q, want llm", p.Name())
	}

	if p.Model() != "gpt-4" {
		t.Errorf("Model() = %q, want gpt-4", p.Model())
	}
}

func TestLLMCLIProvider_GenerateArgs(t *testing.T) {
	temperature := 0.25
	maxTokens := 128

	tests := []struct {
		name    string
		model   string
		prompt  string
		options llm.GenerateOptions
		want    []string
	}{
		{
			name:   "model temperature max tokens and prompt",
			model:  "gpt-4o-mini",
			prompt: `Say "hello"`,
			options: llm.GenerateOptions{
				Temperature: &temperature,
				MaxTokens:   &maxTokens,
			},
			want: []string{
				"-m", "gpt-4o-mini",
				"-o", "temperature", "0.250000",
				"-o", "max_tokens", "128",
				`Say "hello"`,
			},
		},
		{
			name:   "default model uses prompt only",
			model:  "default",
			prompt: "hello",
			want:   []string{"hello"},
		},
		{
			name:   "empty model uses prompt only",
			model:  "",
			prompt: "hello",
			want:   []string{"hello"},
		},
		{
			name:   "prompt remains one argument",
			model:  "gpt-4",
			prompt: "line 1\nline 2 with spaces",
			want:   []string{"-m", "gpt-4", "line 1\nline 2 with spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotCommand string
			var gotArgs []string
			old := llmCLICommandContext
			llmCLICommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
				gotCommand = command
				gotArgs = append([]string(nil), args...)

				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestLLMCLIProviderHelperProcess", "--")
				cmd.Env = append(os.Environ(),
					"PE_WANT_LLM_CLI_HELPER_PROCESS=1",
					"PE_TEST_OUTPUT=Mocked output",
				)
				return cmd
			}
			t.Cleanup(func() {
				llmCLICommandContext = old
			})

			p := &LLMCLIProvider{executable: "llm", model: tt.model}
			resp, err := p.Generate(context.Background(), tt.prompt, tt.options)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}
			if resp.Text != "Mocked output" {
				t.Fatalf("resp.Text = %q, want Mocked output", resp.Text)
			}
			if gotCommand != "llm" {
				t.Fatalf("command = %q, want llm", gotCommand)
			}
			if !slices.Equal(gotArgs, tt.want) {
				t.Fatalf("args = %#v, want %#v", gotArgs, tt.want)
			}
		})
	}
}

func TestLLMCLIProviderHelperProcess(t *testing.T) {
	if os.Getenv("PE_WANT_LLM_CLI_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Print(os.Getenv("PE_TEST_OUTPUT"))
	os.Exit(0)
}

func TestLLMCLIProvider_Integration(t *testing.T) {
	if os.Getenv("PE_INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	options := map[string]interface{}{
		"executable": "echo", // Use echo to succeed immediately
	}

	p, err := NewLLMCLIProvider("test-model", options)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// This will fail because NewLLMCLIProvider checks for executable existence via LookPath
	// and "echo" might not be in the path depending on the environment, or the args might handle differently.
	// But since we use exec.CommandContext, using "echo" should work if it's in PATH.

	resp, err := p.Generate(ctx, "hello", llm.GenerateOptions{})
	if err != nil {
		t.Logf("Generate failed as expected (since we're using echo or it's missing): %v", err)
	} else {
		t.Logf("Generate output: %s", resp.Text)
	}
}
