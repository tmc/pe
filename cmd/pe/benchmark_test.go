package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBenchmarkCmd_BasicUsage(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name       string
		configFile string
		args       []string
		wantErr    bool
	}{
		{
			name:       "benchmark with config",
			configFile: "bench-config.yaml",
			args:       []string{"bench-config.yaml"},
			wantErr:    false,
		},
		{
			name:    "benchmark without config",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "benchmark with nonexistent file",
			args:    []string{"nonexistent.yaml"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldWd, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldWd)

			// Create config file if specified
			if tt.configFile != "" {
				configContent := `description: "Benchmark test"
prompts:
  - "Test prompt"
providers:
  - name: "mock"
    config:
      model: "test-model"
tests:
  - vars: {}
    assert:
      - type: contains
        value: "Mock"
`
				err := os.WriteFile(tt.configFile, []byte(configContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			}

			cmd := benchmarkCmd()
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
			}
		})
	}
}

func TestBenchmarkCmd_Flags(t *testing.T) {
	cmd := benchmarkCmd()

	// Test flag existence - only test flags that actually exist
	expectedFlags := []string{"iterations", "concurrency", "output", "format"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag --%s to exist", flagName)
		}
	}
}

func TestBenchmarkCmd_RepeatFlag(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	configContent := `description: "Repeat test"
prompts:
  - "Test prompt"
providers:
  - name: "mock"
tests:
  - vars: {}
    assert:
      - type: contains
        value: "Mock"
`
	configFile := "repeat-config.yaml"
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tests := []struct {
		name       string
		iterations int
	}{
		{"single run", 1},
		{"multiple runs", 5},
		{"many runs", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{configFile, "--iterations", fmt.Sprintf("%d", tt.iterations)}

			cmd := benchmarkCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err := cmd.Execute()
			output := buf.String()

			if err != nil {
				t.Errorf("Unexpected error with iterations %d: %v. Output: %s", tt.iterations, err, output)
			}
		})
	}
}

func TestBenchmarkCmd_ConcurrencyFlag(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	configContent := `description: "Concurrency test"
prompts:
  - "Test prompt"
providers:
  - name: "mock"
tests:
  - vars: {}
    assert:
      - type: contains
        value: "Mock"
`
	configFile := "concurrent-config.yaml"
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tests := []struct {
		name        string
		concurrency int
	}{
		{"single thread", 1},
		{"multiple threads", 3},
		{"high concurrency", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{configFile, "--concurrency", fmt.Sprintf("%d", tt.concurrency)}

			cmd := benchmarkCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err := cmd.Execute()
			output := buf.String()

			if err != nil {
				t.Errorf("Unexpected error with concurrency %d: %v. Output: %s", tt.concurrency, err, output)
			}
		})
	}
}

func TestBenchmarkCmd_OutputFormats(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	configContent := `description: "Output format test"
prompts:
  - "Test prompt"
providers:
  - name: "mock"
tests:
  - vars: {}
    assert:
      - type: contains
        value: "Mock"
`
	configFile := "format-config.yaml"
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tests := []struct {
		name   string
		format string
	}{
		{"default format", ""},
		{"json format", "json"},
		{"csv format", "csv"},
		{"table format", "table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{configFile}
			if tt.format != "" {
				args = append(args, "--output", tt.format)
			}

			cmd := benchmarkCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err := cmd.Execute()
			output := buf.String()

			if err != nil {
				t.Errorf("Unexpected error with format %s: %v. Output: %s", tt.format, err, output)
			}
		})
	}
}

func TestBenchmarkCmd_ObjectProviderStructuredMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	scriptPath := filepath.Join(tmpDir, "structured-provider.sh")
	err := os.WriteFile(scriptPath, []byte(`#!/bin/sh
printf '%s\n' '{"output":"ok","prompt_tokens":11,"completion_tokens":7,"total_tokens":18,"latency_ms":42,"metrics":{"tokens_per_second":123.4}}'
`), 0o755)
	if err != nil {
		t.Fatalf("WriteFile(script) failed: %v", err)
	}

	configFile := filepath.Join(tmpDir, "benchmark.yaml")
	err = os.WriteFile(configFile, []byte(fmt.Sprintf(`description: "Structured benchmark"
prompts:
  - "Benchmark {{input}}"
providers:
  - id: cli:stub
    label: cli label
    config:
      command: %q
      parse_json_response: true
tests:
  - vars:
      input: "prompt"
`, scriptPath)), 0o644)
	if err != nil {
		t.Fatalf("WriteFile(config) failed: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "results.json")
	cmd := benchmarkCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{configFile, "--iterations", "1", "--output", outputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v\n%s", err, buf.String())
	}

	var output struct {
		Results   []BenchmarkResult  `json:"results"`
		Summaries []BenchmarkSummary `json:"summaries"`
	}
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("ReadFile(output) failed: %v", err)
	}
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}
	if len(output.Results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(output.Results))
	}
	result := output.Results[0]
	if result.Provider != "cli label" || result.ProviderID != "cli:stub" {
		t.Fatalf("provider = (%q,%q), want (cli label, cli:stub)", result.Provider, result.ProviderID)
	}
	if result.LatencyMs != 42 {
		t.Fatalf("result.LatencyMs = %.0f, want 42", result.LatencyMs)
	}
	if result.TokensInput != 11 || result.TokensOutput != 7 || result.TokensTotal != 18 {
		t.Fatalf("token counts = (%d,%d,%d), want (11,7,18)", result.TokensInput, result.TokensOutput, result.TokensTotal)
	}
	if got := result.RuntimeMetrics["tokens_per_second"]; got != 123.4 {
		t.Fatalf("runtimeMetrics[tokens_per_second] = %#v, want 123.4", got)
	}
	if len(output.Summaries) != 1 || output.Summaries[0].AvgLatencyMs != 42 {
		t.Fatalf("summary avg latency = %#v, want 42", output.Summaries)
	}
	if output.Summaries[0].Provider != "cli label" || output.Summaries[0].ProviderID != "cli:stub" {
		t.Fatalf("summary provider = (%q,%q)", output.Summaries[0].Provider, output.Summaries[0].ProviderID)
	}
}

// TestBenchmarkCmd_WarmupFlag removed - warmup flag does not exist in current implementation

// Benchmark tests using stdlib testing benchmarks
func BenchmarkRunCommand(b *testing.B) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := runCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"test prompt"})
		cmd.Execute()
	}
}

func BenchmarkTemplateSubstitution(b *testing.B) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := runCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"Hello {{.name}}", "--var", "name=World"})
		cmd.Execute()
	}
}

func BenchmarkFileReading(b *testing.B) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	// Create temp file
	tmpDir := b.TempDir()
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	err := os.WriteFile(promptFile, []byte("Test prompt content"), 0644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := runCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{promptFile})
		cmd.Execute()
	}
}
