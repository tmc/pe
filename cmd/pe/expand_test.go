package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolvePrompts(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pe-expand-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create dummy prompt files
	prompt1 := filepath.Join(tempDir, "prompt1.txt")
	if err := os.WriteFile(prompt1, []byte("Content 1"), 0644); err != nil {
		t.Fatal(err)
	}
	prompt2 := filepath.Join(tempDir, "prompt2.txt")
	if err := os.WriteFile(prompt2, []byte("Content 2"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		input   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name:  "string literal",
			input: "Verify this",
			want:  []interface{}{"Verify this"},
		},
		{
			name:  "string file",
			input: "prompt1.txt",
			want:  []interface{}{"Content 1"},
		},
		{
			name:  "file protocol prefix",
			input: "file://prompt1.txt",
			want:  []interface{}{"Content 1"},
		},
		{
			name:  "glob pattern",
			input: "*.txt",
			want:  []interface{}{"Content 1", "Content 2"},
		},
		{
			name:  "list of mixed",
			input: []interface{}{"Static", "prompt1.txt"},
			want:  []interface{}{"Static", "Content 1"},
		},
		{
			name:  "map input",
			input: map[string]interface{}{"id": "foo", "prompt": "bar"},
			want:  []interface{}{map[string]interface{}{"id": "foo", "prompt": "bar"}},
		},
		{
			name:    "missing file",
			input:   "missing.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePrompts(tt.input, tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePrompts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Normalize result for comparison if needed, or check subsets for globs since order might vary
				if len(got) != len(tt.want) {
					t.Errorf("resolvePrompts() got %v, want %v", got, tt.want)
				}
				// For globs, sorting might be needed, but filepath.Glob returns sorted matches usually.
				// However, let's just do deep equal for simple cases
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("resolvePrompts() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestResolveTests(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pe-expand-test-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testsJSON := filepath.Join(tempDir, "tests.json")
	if err := os.WriteFile(testsJSON, []byte(`[{"vars": {"a": 1}}]`), 0644); err != nil {
		t.Fatal(err)
	}
	testsYAML := filepath.Join(tempDir, "tests.yaml")
	if err := os.WriteFile(testsYAML, []byte("- vars:\n    b: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		input   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name:  "load json",
			input: "tests.json",
			want: []interface{}{
				map[string]interface{}{"vars": map[string]interface{}{"a": float64(1)}},
			},
		},
		{
			name:  "load yaml",
			input: "tests.yaml",
			want: []interface{}{
				map[string]interface{}{"vars": map[string]interface{}{"b": 2}},
			},
		},
		{
			name:  "list files",
			input: []interface{}{"tests.json", "tests.yaml"},
			want: []interface{}{
				map[string]interface{}{"vars": map[string]interface{}{"a": float64(1)}},
				map[string]interface{}{"vars": map[string]interface{}{"b": 2}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTests(tt.input, tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveTests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// JSON decoding uses float64 for numbers, YAML might use int.
			// Let's do a JSON roundtrip on want/got to normalize types if needed, or relax check.
			// Reflect.DeepEqual for interface{} maps is tricky with numbers.
			// Let's try direct comparison first, assume yaml/json unmarshal behavior is consistent enough for simple values
			// or use json marshal for comparison.

			gotBytes, _ := json.Marshal(got)
			wantBytes, _ := json.Marshal(tt.want)

			if string(gotBytes) != string(wantBytes) {
				t.Errorf("resolveTests() = %s, want %s", gotBytes, wantBytes)
			}
		})
	}
}

func TestLoadAndExpandConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pe-expand-config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
description: Test Config
prompts: [prompt.txt]
providers: [openai:gpt-4]
tests: [tests.json]
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "prompt.txt"), []byte("Hello world"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "tests.json"), []byte(`[{"assert": []}]`), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := loadAndExpandConfig(configFile)
	if err != nil {
		t.Fatalf("loadAndExpandConfig() error = %v", err)
	}

	if got.Description != "Test Config" {
		t.Errorf("Description = %q, want 'Test Config'", got.Description)
	}
	if len(got.Prompts) != 1 || got.Prompts[0] != "Hello world" {
		t.Errorf("Prompts expanded incorrectly: %v", got.Prompts)
	}
	if len(got.Tests) != 1 {
		t.Errorf("Tests expanded incorrectly: %v", got.Tests)
	}
}

func TestExpandCmdOutputDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("config.yaml", []byte(expandPolicyTestConfig()), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(expandPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := expandCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"config.yaml", "--output", "expanded.json"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("expand error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("expanded.json"); !os.IsNotExist(err) {
		t.Fatalf("expand wrote output despite write policy: %v", err)
	}
}

func TestExpandCmdStdoutAllowedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("config.yaml", []byte(expandPolicyTestConfig()), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(expandPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := expandCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"config.yaml"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expand stdout: %v", err)
	}
	if _, err := os.Stat("expanded.json"); !os.IsNotExist(err) {
		t.Fatalf("expand created unexpected output: %v", err)
	}
}

func expandPolicyTestConfig() string {
	return `description: Test Config
prompts: ["hello"]
providers: [mock]
tests: []
`
}

func expandPolicyTestModule() string {
	return `module example.com/prompts

pe 1

capability {
    tools deny write
}
`
}

func TestSafeConfigPathRejectsTraversal(t *testing.T) {
	base := t.TempDir()
	tests := []string{
		"../secret.txt",
		"file:///tmp/secret.txt",
		filepath.Join("..", "secret.txt"),
	}
	for _, ref := range tests {
		t.Run(ref, func(t *testing.T) {
			if _, err := safeConfigPath(base, strings.TrimPrefix(ref, "file://")); err == nil {
				t.Fatalf("safeConfigPath(%q) succeeded, want error", ref)
			}
		})
	}
}
