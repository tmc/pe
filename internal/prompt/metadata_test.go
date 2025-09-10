package prompt

import (
	"reflect"
	"testing"
)

func TestExtractDefaults(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name: "simple defaults",
			content: `#!/usr/bin/env pe run
Main prompt

---defaults---
name: Alice
age: 30
city: New York`,
			expected: map[string]string{
				"name": "Alice",
				"age":  "30",
				"city": "New York",
			},
		},
		{
			name: "no defaults section",
			content: `#!/usr/bin/env pe run
Just a prompt`,
			expected: map[string]string{},
		},
		{
			name: "defaults with other sections",
			content: `Main prompt

---defaults---
var1: value1
var2: value2

---system---
System prompt

---tests---
Test cases`,
			expected: map[string]string{
				"var1": "value1",
				"var2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDefaults(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractDefaults() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "single line description",
			content: `#!/usr/bin/env pe run
# This is a test prompt
Main content`,
			expected: "This is a test prompt",
		},
		{
			name: "multi-line description",
			content: `#!/usr/bin/env pe run
# Line 1 of description
# Line 2 of description
# Line 3 of description

Main content`,
			expected: "Line 1 of description\nLine 2 of description\nLine 3 of description",
		},
		{
			name: "no description",
			content: `#!/usr/bin/env pe run
Main content directly`,
			expected: "",
		},
		{
			name: "description without shebang",
			content: `# Description here
Main content`,
			expected: "Description here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDescription(tt.content)
			if result != tt.expected {
				t.Errorf("ExtractDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "simple variables",
			content: `Process {{.input}} and return {{.output}}`,
			expected: []string{"input", "output"},
		},
		{
			name: "variables with sections",
			content: `Main: {{.var1}} and {{.var2}}

---defaults---
var1: default1
var2: default2

---system---
System with {{.ignored}}`,
			expected: []string{"var1", "var2"},
		},
		{
			name: "duplicate variables",
			content: `First {{.name}}, second {{.name}}, third {{.age}}`,
			expected: []string{"name", "age"},
		},
		{
			name: "variables with pipes",
			content: `Format: {{.data | json}}`,
			expected: []string{"data"},
		},
		{
			name: "no variables",
			content: `Plain text without variables`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVariables(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractVariables() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractSystemPrompt(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "simple system prompt",
			content: `Main prompt

---system---
You are a helpful assistant.`,
			expected: "You are a helpful assistant.",
		},
		{
			name: "multi-line system prompt",
			content: `Main prompt

---system---
You are an expert.
Be concise.
Be accurate.

---defaults---
foo: bar`,
			expected: "You are an expert.\nBe concise.\nBe accurate.",
		},
		{
			name: "no system prompt",
			content: `Main prompt

---defaults---
foo: bar`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSystemPrompt(tt.content)
			if result != tt.expected {
				t.Errorf("ExtractSystemPrompt() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractMetadata(t *testing.T) {
	content := `#!/usr/bin/env pe run --provider=openai
# Test Prompt
# This is a test description

Process {{.input}} with {{.method}}

---defaults---
input: test data
method: analysis

---system---
You are helpful.

---tests---
- test case

---variants---
verbose: detailed output`

	meta := ExtractMetadata(content)

	if meta.Description != "Test Prompt\nThis is a test description" {
		t.Errorf("Description = %q", meta.Description)
	}

	if !reflect.DeepEqual(meta.Variables, []string{"input", "method"}) {
		t.Errorf("Variables = %v", meta.Variables)
	}

	expectedDefaults := map[string]string{
		"input":  "test data",
		"method": "analysis",
	}
	if !reflect.DeepEqual(meta.Defaults, expectedDefaults) {
		t.Errorf("Defaults = %v", meta.Defaults)
	}

	if meta.SystemPrompt != "You are helpful." {
		t.Errorf("SystemPrompt = %q", meta.SystemPrompt)
	}

	if !meta.HasShebang {
		t.Error("HasShebang should be true")
	}

	if meta.Provider != "openai" {
		t.Errorf("Provider = %q, want openai", meta.Provider)
	}

	expectedSections := []string{"defaults", "system", "tests", "variants"}
	if !reflect.DeepEqual(meta.Sections, expectedSections) {
		t.Errorf("Sections = %v, want %v", meta.Sections, expectedSections)
	}
}