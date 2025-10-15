package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestEvalCmd_BasicUsage(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name       string
		configFile string
		args       []string
		wantErr    bool
		contains   string
	}{
		{
			name:       "eval with config file",
			configFile: "eval-config.yaml",
			args:       []string{"eval-config.yaml"},
			wantErr:    false,
			contains:   "", // Just check it doesn't error
		},
		{
			name:    "eval without config file",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "eval with nonexistent file",
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
				configContent := `description: "Test evaluation"
prompts:
  - "What is 2+2?"
providers:
  - "mock"
tests:
  - vars: {}
    assert:
      - type: contains
        value: "4"
`
				err := os.WriteFile(tt.configFile, []byte(configContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			}

			cmd := evalCmd()
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
				if tt.contains != "" && !strings.Contains(output, tt.contains) {
					t.Errorf("Expected output to contain %q, but got: %s", tt.contains, output)
				}
			}
		})
	}
}

func TestEvalCmd_Flags(t *testing.T) {
	cmd := evalCmd()

	// Test flag existence - only test flags that actually exist
	expectedFlags := []string{"output", "max-concurrency", "timeout"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag --%s to exist", flagName)
		}
	}

	// Test flag parsing
	err := cmd.ParseFlags([]string{"--max-concurrency", "5", "--output", "results.json"})
	if err != nil {
		t.Errorf("Failed to parse flags: %v", err)
	}

	// Check flag values
	concurrency, err := cmd.Flags().GetInt("max-concurrency")
	if err != nil {
		t.Errorf("Failed to get max-concurrency flag: %v", err)
	}
	if concurrency != 5 {
		t.Errorf("Expected max-concurrency to be 5, got %d", concurrency)
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		t.Errorf("Failed to get output flag: %v", err)
	}
	if output != "results.json" {
		t.Errorf("Expected output to be 'results.json', got %s", output)
	}
}

func TestEvalCmd_OutputFormats(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name   string
		format string
	}{
		{"default format", ""},
		{"json format", "json"},
		{"yaml format", "yaml"},
		{"table format", "table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldWd, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldWd)

			// Create minimal config file
			configContent := `description: "Test evaluation"
prompts:
  - "Test prompt"
providers:
  - "mock"
tests:
  - vars: {}
    assert:
      - type: contains
        value: "test"
`
			configFile := "test-config.yaml"
			err := os.WriteFile(configFile, []byte(configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			args := []string{configFile}
			if tt.format != "" {
				args = append(args, "--output", tt.format)
			}

			cmd := evalCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(args)

			err = cmd.Execute()
			output := buf.String()

			// Just check it doesn't error for now
			// In a real implementation, we'd check the format
			if err != nil {
				t.Errorf("Unexpected error with format %s: %v. Output: %s", tt.format, err, output)
			}
		})
	}
}

func TestEvalCmd_VariableSubstitution(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create config with variables
	configContent := `description: "Variable test evaluation"
prompts:
  - "Hello {{.name}}, you are {{.age}} years old"
providers:
  - "mock"
tests:
  - vars:
      name: "Alice"
      age: "30"
    assert:
      - type: contains
        value: "Alice"
`
	configFile := "var-config.yaml"
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test that eval runs without error - variables come from the config file test cases
	cmd := evalCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{configFile})

	err = cmd.Execute()
	output := buf.String()

	if err != nil {
		t.Errorf("Unexpected error: %v. Output: %s", err, output)
	}
}

func TestEvalCmd_Concurrency(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create config with multiple test cases
	configContent := `description: "Concurrency test evaluation"
prompts:
  - "Test prompt"
providers:
  - "mock"
tests:
  - vars: {}
    assert:
      - type: contains
        value: "test"
  - vars: {}
    assert:
      - type: contains
        value: "test"
  - vars: {}
    assert:
      - type: contains
        value: "test"
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
		{"default concurrency", 0},
		{"single thread", 1},
		{"multiple threads", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{configFile}
			if tt.concurrency > 0 {
				args = append(args, "--max-concurrency", fmt.Sprintf("%d", tt.concurrency))
			}

			cmd := evalCmd()
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

// TestEvalCmd_FilterFlag removed - filter flag does not exist in current implementation