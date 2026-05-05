package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/inference"
)

func TestRunCmd_BasicUsage(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantStdout string
	}{
		{
			name:       "simple math prompt",
			args:       []string{"What is 2+2?", "--stream=false"},
			wantErr:    false,
			wantStdout: "4",
		},
		{
			name:       "direct text prompt",
			args:       []string{"Hello world", "--stream=false"},
			wantErr:    false,
			wantStdout: "Mock response for: Hello world",
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := runCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", output)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, output)
				}
				if tt.wantStdout != "" && !strings.Contains(output, tt.wantStdout) {
					t.Errorf("Expected output to contain %q, but got: %s", tt.wantStdout, output)
				}
			}
		})
	}
}

func TestRunCmd_FileInput(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name        string
		fileContent string
		args        []string
		wantErr     bool
		wantStdout  string
	}{
		{
			name:        "read from file",
			fileContent: "What is the capital of France?",
			args:        []string{"prompt.txt", "--stream=false"},
			wantErr:     false,
			wantStdout:  "<answer>The capital of France is Paris</answer>",
		},
		{
			name:        "read prompt file format",
			fileContent: "Main prompt text\n\n-- system-prompt --\nYou are helpful",
			args:        []string{"prompt.txt", "--stream=false"},
			wantErr:     false,
			wantStdout:  "Mock response for: Main prompt text",
		},
		{
			name:    "nonexistent file as direct prompt",
			args:    []string{"nonexistent.txt", "--stream=false"},
			wantErr: false, // Should treat as direct text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test file if content provided
			var filePath string
			if tt.fileContent != "" {
				filePath = filepath.Join(tmpDir, "prompt.txt")
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				// Replace filename in args with full path
				for i, arg := range tt.args {
					if arg == "prompt.txt" {
						tt.args[i] = filePath
					}
				}
			}

			cmd := runCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", output)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, output)
				}
				if tt.wantStdout != "" && !strings.Contains(output, tt.wantStdout) {
					t.Errorf("Expected output to contain %q, but got: %s", tt.wantStdout, output)
				}
			}
		})
	}
}

func TestRunCmd_Variables(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name    string
		prompt  string
		vars    []string
		wantErr bool
		checkFn func(output string) bool
	}{
		{
			name:    "template substitution",
			prompt:  "Hello {{.name}}",
			vars:    []string{"--var", "name=World", "--stream=false"},
			wantErr: false,
			checkFn: func(output string) bool {
				return strings.Contains(output, "Hello World") || strings.Contains(output, "Mock response for: Hello World")
			},
		},
		{
			name:    "multiple variables",
			prompt:  "{{.greeting}} {{.name}}!",
			vars:    []string{"--var", "greeting=Hello", "--var", "name=Alice", "--stream=false"},
			wantErr: false,
			checkFn: func(output string) bool {
				return strings.Contains(output, "Hello Alice!") || strings.Contains(output, "Mock response for: Hello Alice!")
			},
		},
		{
			name:    "no template variables",
			prompt:  "Simple prompt",
			vars:    []string{"--var", "unused=value", "--stream=false"},
			wantErr: false,
			checkFn: func(output string) bool {
				return strings.Contains(output, "Mock response for: Simple prompt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{tt.prompt}, tt.vars...)

			cmd := runCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err := cmd.Execute()
			output := buf.String()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", output)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, output)
				}
				if tt.checkFn != nil && !tt.checkFn(output) {
					t.Errorf("Check function failed for output: %s", output)
				}
			}
		})
	}
}

func TestRunCmd_Providers(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name     string
		prompt   string
		provider string
		wantErr  bool
	}{
		{
			name:     "default provider",
			prompt:   "Test prompt",
			provider: "",
			wantErr:  false,
		},
		{
			name:     "cgpt provider",
			prompt:   "Test prompt",
			provider: "cgpt",
			wantErr:  false,
		},
		{
			name:     "mock provider",
			prompt:   "Test prompt",
			provider: "mock",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{tt.prompt}
			if tt.provider != "" {
				args = append(args, "--provider", tt.provider)
			}

			cmd := runCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err := cmd.Execute()
			output := buf.String()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", output)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, output)
				}
			}
		})
	}
}

func TestRunCmd_StreamFlag(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	cmd := runCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"test prompt", "--stream"})

	err := cmd.Execute()
	output := buf.String()

	if err != nil {
		t.Errorf("Unexpected error: %v. Output: %s", err, output)
	}
}

func TestRunCmd_FlagParsing(t *testing.T) {
	cmd := runCmd()

	// Test flag existence
	providerFlag := cmd.Flags().Lookup("provider")
	if providerFlag == nil {
		t.Error("Expected --provider flag to exist")
	}

	varFlag := cmd.Flags().Lookup("var")
	if varFlag == nil {
		t.Error("Expected --var flag to exist")
	}

	streamFlag := cmd.Flags().Lookup("stream")
	if streamFlag == nil {
		t.Error("Expected --stream flag to exist")
	}

	// Test flag types
	if providerFlag != nil && providerFlag.Value.Type() != "string" {
		t.Errorf("Expected provider flag to be string, got %s", providerFlag.Value.Type())
	}

	if streamFlag != nil && streamFlag.Value.Type() != "bool" {
		t.Errorf("Expected stream flag to be bool, got %s", streamFlag.Value.Type())
	}
}

func TestRegisterLLMProviderSpec(t *testing.T) {
	client := inference.NewClient()
	if err := registerLLMProviderSpec(client, "mock:test"); err != nil {
		t.Fatalf("registerLLMProviderSpec() failed: %v", err)
	}
	resp, err := client.CompleteWith(context.Background(), "mock:test", inference.Request{Prompt: "hello"})
	if err != nil {
		t.Fatalf("CompleteWith() failed: %v", err)
	}
	if resp.Content == "" {
		t.Fatal("empty response from registered provider spec")
	}
}
