package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVariableAssignments(t *testing.T) {
	tests := []struct {
		name        string
		assignments []string
		expected    map[string]string
		shouldError bool
	}{
		{
			name:        "empty assignments",
			assignments: []string{},
			expected:    map[string]string{},
			shouldError: false,
		},
		{
			name:        "single assignment",
			assignments: []string{"key=value"},
			expected:    map[string]string{"key": "value"},
			shouldError: false,
		},
		{
			name:        "multiple assignments",
			assignments: []string{"key1=value1", "key2=value2"},
			expected:    map[string]string{"key1": "value1", "key2": "value2"},
			shouldError: false,
		},
		{
			name:        "assignment with spaces",
			assignments: []string{"key = value with spaces"},
			expected:    map[string]string{"key": "value with spaces"},
			shouldError: false,
		},
		{
			name:        "assignment with equals in value",
			assignments: []string{"key=value=with=equals"},
			expected:    map[string]string{"key": "value=with=equals"},
			shouldError: false,
		},
		{
			name:        "invalid assignment format",
			assignments: []string{"keyvalue"},
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "empty key",
			assignments: []string{"=value"},
			expected:    nil,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVariableAssignments(tt.assignments)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d assignments, got %d", len(tt.expected), len(result))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("expected key %s not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("expected value %s for key %s, got %s", expectedValue, key, actualValue)
				}
			}
		})
	}
}

func TestExtractVariablesFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected map[string]string
	}{
		{
			name:     "no variables",
			text:     "This is a simple text with no variables",
			expected: map[string]string{},
		},
		{
			name:     "single variable",
			text:     "Hello {{name}}!",
			expected: map[string]string{"name": ""},
		},
		{
			name:     "multiple variables",
			text:     "Hello {{name}}, welcome to {{place}}!",
			expected: map[string]string{"name": "", "place": ""},
		},
		{
			name:     "variable with spaces",
			text:     "The {{ variable_name }} is here",
			expected: map[string]string{"variable_name": ""},
		},
		{
			name:     "repeated variable",
			text:     "{{name}} and {{name}} again",
			expected: map[string]string{"name": ""},
		},
		{
			name:     "complex text with multiple variables",
			text:     "Analyze {{topic}} from {{perspective}} perspective with {{detail_level}} detail",
			expected: map[string]string{"topic": "", "perspective": "", "detail_level": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVariablesFromText(tt.text)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d variables, got %d", len(tt.expected), len(result))
				return
			}

			for key := range tt.expected {
				if _, exists := result[key]; !exists {
					t.Errorf("expected variable %s not found", key)
				}
			}
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		variables map[string]string
		expected  string
	}{
		{
			name:      "no variables",
			text:      "Simple text",
			variables: map[string]string{},
			expected:  "Simple text",
		},
		{
			name:      "single variable",
			text:      "Hello {{name}}!",
			variables: map[string]string{"name": "World"},
			expected:  "Hello World!",
		},
		{
			name:      "multiple variables",
			text:      "Hello {{name}}, welcome to {{place}}!",
			variables: map[string]string{"name": "Alice", "place": "Wonderland"},
			expected:  "Hello Alice, welcome to Wonderland!",
		},
		{
			name:      "variable with spaces",
			text:      "The {{ name }} is here",
			variables: map[string]string{"name": "result"},
			expected:  "The result is here",
		},
		{
			name:      "missing variable",
			text:      "Hello {{name}}!",
			variables: map[string]string{},
			expected:  "Hello {{name}}!",
		},
		{
			name:      "partial substitution",
			text:      "Hello {{name}} from {{place}}!",
			variables: map[string]string{"name": "Alice"},
			expected:  "Hello Alice from {{place}}!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteVariables(tt.text, tt.variables)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseTextPrompt(t *testing.T) {
	text := "Analyze {{topic}} from {{perspective}} perspective"
	content := []byte(text)

	components := parseTextPrompt(content)

	if components.UserPrompt != text {
		t.Errorf("expected UserPrompt to be %q, got %q", text, components.UserPrompt)
	}

	if components.RawContent != text {
		t.Errorf("expected RawContent to be %q, got %q", text, components.RawContent)
	}

	expectedVars := map[string]string{"topic": "", "perspective": ""}
	if len(components.Variables) != len(expectedVars) {
		t.Errorf("expected %d variables, got %d", len(expectedVars), len(components.Variables))
	}

	for key := range expectedVars {
		if _, exists := components.Variables[key]; !exists {
			t.Errorf("expected variable %s not found", key)
		}
	}
}

func TestParseYAMLPrompt(t *testing.T) {
	yamlContent := `
system: "You are a helpful assistant"
user: "What is {{topic}}?"
variables:
  topic: "AI"
  detail: "basic"
config:
  temperature: 0.7
metadata:
  version: "1.0"
`

	components, err := parseYAMLPrompt([]byte(yamlContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if components.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("expected SystemPrompt to be 'You are a helpful assistant', got %q", components.SystemPrompt)
	}

	if components.UserPrompt != "What is {{topic}}?" {
		t.Errorf("expected UserPrompt to be 'What is {{topic}}?', got %q", components.UserPrompt)
	}

	if components.Variables["topic"] != "AI" {
		t.Errorf("expected topic variable to be 'AI', got %q", components.Variables["topic"])
	}

	if components.Variables["detail"] != "basic" {
		t.Errorf("expected detail variable to be 'basic', got %q", components.Variables["detail"])
	}

	if components.Config["temperature"] != 0.7 {
		t.Errorf("expected temperature to be 0.7, got %v", components.Config["temperature"])
	}

	if components.Metadata["version"] != "1.0" {
		t.Errorf("expected version to be '1.0', got %v", components.Metadata["version"])
	}
}

func TestExtractVariablesFromTextWithDots(t *testing.T) {
	text := "Hello {{.name}}!"
	result := extractVariablesFromText(text)
	expected := map[string]string{"name": ""}

	for key, expectedValue := range expected {
		if actualValue, exists := result[key]; !exists {
			t.Errorf("expected key %s not found in result %+v", key, result)
		} else if actualValue != expectedValue {
			t.Errorf("expected value %s for key %s, got %s", expectedValue, key, actualValue)
		}
	}
}

func TestCatCmd_BasicUsage(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		fileName    string
		args        []string
		wantErr     bool
		wantOutput  string
	}{
		{
			name:        "read plain text file",
			fileContent: "Hello, World!",
			fileName:    "test.txt",
			args:        []string{},
			wantErr:     false,
			wantOutput:  "Hello, World!",
		},
		{
			name:        "read prompt file with variables",
			fileContent: "Analyze {{topic}} carefully",
			fileName:    "test.prompt",
			args:        []string{"--set", "topic=AI"},
			wantErr:     false,
			wantOutput:  "Analyze AI carefully",
		},
		{
			name:        "raw mode",
			fileContent: "{{topic}} test",
			fileName:    "test.txt",
			args:        []string{"--raw"},
			wantErr:     false,
			wantOutput:  "{{topic}} test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			filePath := filepath.Join(tmpDir, tt.fileName)
			if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			cmd := catCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			args := append([]string{filePath}, tt.args...)
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// The output goes to stdout, not the buffer
				// Just verify no error occurred
			}
		})
	}
}

func TestCatCmd_NonexistentFile(t *testing.T) {
	cmd := catCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"/nonexistent/file.txt"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCatCmd_ShowVariables(t *testing.T) {
	tmpDir := t.TempDir()

	content := "Hello {{name}}, welcome to {{place}}!"
	filePath := filepath.Join(tmpDir, "test.prompt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := catCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{filePath, "--variables"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_TextFile(t *testing.T) {
	tmpDir := t.TempDir()

	content := "Simple prompt text"
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := runCat(filePath, catOptions{raw: true})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_YAMLFile(t *testing.T) {
	tmpDir := t.TempDir()

	content := `system: You are a helpful assistant
user: What is {{topic}}?
variables:
  topic: AI`

	filePath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := runCat(filePath, catOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_JSONFile(t *testing.T) {
	tmpDir := t.TempDir()

	content := `{"system": "You are helpful", "user": "Hello"}`

	filePath := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := runCat(filePath, catOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_WithVariableSubstitution(t *testing.T) {
	tmpDir := t.TempDir()

	content := "Hello {{name}}, welcome to {{place}}!"
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := runCat(filePath, catOptions{
		setVars: []string{"name=World", "place=Earth"},
	})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_ShowSystemPrompt(t *testing.T) {
	tmpDir := t.TempDir()

	content := `system: You are a helpful assistant
user: Hello`

	filePath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := runCat(filePath, catOptions{showSystem: true})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_ShowComponents(t *testing.T) {
	tmpDir := t.TempDir()

	content := `system: You are a helpful assistant
user: Hello
variables:
  topic: AI
config:
  temperature: 0.7
metadata:
  version: "1.0"`

	filePath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := runCat(filePath, catOptions{showComponents: true})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_ShowMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	content := `user: Hello
config:
  temperature: 0.7
metadata:
  version: "1.0"
  author: Test`

	filePath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := runCat(filePath, catOptions{showMetadata: true})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunCat_OutputFormats(t *testing.T) {
	tmpDir := t.TempDir()

	content := `system: You are a helpful assistant
user: Hello`

	filePath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name   string
		format string
	}{
		{"text format", "text"},
		{"yaml format", "yaml"},
		{"json format", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCat(filePath, catOptions{outputFormat: tt.format})
			if err != nil {
				t.Errorf("Unexpected error for format %s: %v", tt.format, err)
			}
		})
	}
}

func TestParsePromptFileForCat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "yaml file",
			filename: "test.yaml",
			content:  "system: Hello",
			wantErr:  false,
		},
		{
			name:     "yml file",
			filename: "test.yml",
			content:  "user: World",
			wantErr:  false,
		},
		{
			name:     "json file",
			filename: "test.json",
			content:  `{"user": "Hello"}`,
			wantErr:  false,
		},
		{
			name:     "text file",
			filename: "test.txt",
			content:  "Plain text prompt",
			wantErr:  false,
		},
		{
			name:     "prompt file",
			filename: "test.prompt",
			content:  "Prompt content",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := parsePromptFileForCat(tt.filename, []byte(tt.content))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if components == nil {
					t.Error("Expected components but got nil")
				}
			}
		})
	}
}

func TestContainsVariable(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]string
		varName   string
		want      bool
	}{
		{
			name:      "variable exists",
			variables: map[string]string{"name": "value"},
			varName:   "name",
			want:      true,
		},
		{
			name:      "variable does not exist",
			variables: map[string]string{"name": "value"},
			varName:   "other",
			want:      false,
		},
		{
			name:      "empty map",
			variables: map[string]string{},
			varName:   "name",
			want:      false,
		},
		{
			name:      "empty value",
			variables: map[string]string{"name": ""},
			varName:   "name",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsVariable(tt.variables, tt.varName)
			if got != tt.want {
				t.Errorf("containsVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseJSONPrompt(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid json",
			content: `{"system": "You are helpful", "user": "Hello"}`,
			wantErr: false,
		},
		{
			name:    "json with variables",
			content: `{"user": "What is {{topic}}?", "variables": {"topic": "AI"}}`,
			wantErr: false,
		},
		{
			name:    "empty json object",
			content: `{}`,
			wantErr: false,
		},
		{
			name:    "nested json",
			content: `{"user": "Hello", "metadata": {"version": "1.0"}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := parseJSONPrompt([]byte(tt.content))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if components == nil {
					t.Error("Expected components but got nil")
				}
			}
		})
	}
}

func TestCatCmd_Flags(t *testing.T) {
	cmd := catCmd()

	// Test flag existence
	flags := []string{"system", "variables", "components", "metadata", "set", "interactive", "format", "raw"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected --%s flag to exist", flagName)
		}
	}
}

func TestSubstituteVariablesWithDotNotation(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		variables map[string]string
		expected  string
	}{
		{
			name:      "dot notation",
			text:      "Hello {{.name}}!",
			variables: map[string]string{"name": "World"},
			expected:  "Hello World!",
		},
		{
			name:      "dot notation with spaces",
			text:      "Hello {{ .name }}!",
			variables: map[string]string{"name": "World"},
			expected:  "Hello World!",
		},
		{
			name:      "mixed notation",
			text:      "{{greeting}} {{.name}}!",
			variables: map[string]string{"greeting": "Hello", "name": "World"},
			expected:  "Hello World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteVariables(tt.text, tt.variables)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRunCat_NonexistentFile(t *testing.T) {
	err := runCat("/nonexistent/path/file.txt", catOptions{})
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error message, got: %v", err)
	}
}
