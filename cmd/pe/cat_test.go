package main

import (
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
