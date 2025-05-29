package prompt

import (
	"testing"
)

func TestParseDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name: "single line format",
			input: `Hello {{.name}}

-- defaults --
name=World&greeting=Hello`,
			expected: map[string]string{
				"name":     "World",
				"greeting": "Hello",
			},
		},
		{
			name: "single line with quotes",
			input: `Translate {{.text}}

-- defaults --
'source_lang=English&target_lang=Spanish'`,
			expected: map[string]string{
				"source_lang": "English",
				"target_lang": "Spanish",
			},
		},
		{
			name: "multi-line format",
			input: `API call

-- defaults --
host: api.example.com
port: 443
protocol: https`,
			expected: map[string]string{
				"host":     "api.example.com",
				"port":     "443",
				"protocol": "https",
			},
		},
		{
			name: "mixed format",
			input: `Test

-- defaults --
key1=value1&key2=value2
key3: value3
# Comment
key4: value4`,
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(p.Defaults) != len(tt.expected) {
				t.Errorf("Expected %d defaults, got %d", len(tt.expected), len(p.Defaults))
			}

			for k, v := range tt.expected {
				if p.Defaults[k] != v {
					t.Errorf("Default %s: expected %q, got %q", k, v, p.Defaults[k])
				}
			}
		})
	}
}

func TestPromptWithDefaults(t *testing.T) {
	input := `Translate "{{.text}}" from {{.source}} to {{.target}}

-- defaults --
source=English&target=Spanish

-- system-prompt --
You are a translator.

-- evals --
$ text="Hello"
Hola`

	p, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check main content
	if p.Main != `Translate "{{.text}}" from {{.source}} to {{.target}}` {
		t.Errorf("Unexpected main prompt: %q", p.Main)
	}

	// Check defaults
	if p.Defaults["source"] != "English" {
		t.Errorf("Expected source default to be English, got %q", p.Defaults["source"])
	}
	if p.Defaults["target"] != "Spanish" {
		t.Errorf("Expected target default to be Spanish, got %q", p.Defaults["target"])
	}

	// Check system prompt
	if p.SystemPrompt != "You are a translator." {
		t.Errorf("Unexpected system prompt: %q", p.SystemPrompt)
	}

	// Check evals section
	if _, ok := p.Sections["evals"]; !ok {
		t.Error("Expected evals section to exist")
	}
}