package main

import (
	"bytes"
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

	// Test flag existence
	expectedFlags := []string{"repeat", "concurrency", "output", "warmup"}
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
		name   string
		repeat int
	}{
		{"single run", 1},
		{"multiple runs", 5},
		{"many runs", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{configFile, "--repeat", string(rune(tt.repeat + '0'))}

			cmd := benchmarkCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err := cmd.Execute()
			output := buf.String()

			if err != nil {
				t.Errorf("Unexpected error with repeat %d: %v. Output: %s", tt.repeat, err, output)
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
			args := []string{configFile, "--concurrency", string(rune(tt.concurrency + '0'))}

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

func TestBenchmarkCmd_WarmupFlag(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	configContent := `description: "Warmup test"
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
	configFile := "warmup-config.yaml"
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tests := []struct {
		name   string
		warmup int
	}{
		{"no warmup", 0},
		{"single warmup", 1},
		{"multiple warmups", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{configFile, "--warmup", string(rune(tt.warmup + '0'))}

			cmd := benchmarkCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err := cmd.Execute()
			output := buf.String()

			if err != nil {
				t.Errorf("Unexpected error with warmup %d: %v. Output: %s", tt.warmup, err, output)
			}
		})
	}
}

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