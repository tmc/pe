package prompt

import (
	"strings"
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

func TestPromptFormatVariantsConfigAndExamples(t *testing.T) {
	input := `#!/usr/bin/env pe run --provider=openai --stream --model gpt-4
Answer {{QUESTION}}

-- system-prompt --
You are precise.

-- prompt-summary --
Answers questions.

-- config --
prefill content/prefill
stop-sequence "END"
stop-sequence 'STOP'

-- content/prefill --
Start here

-- variants/strict --
prepend-system-prompt "Follow policy."
extend-system-prompt "No guessing."
prepend-prompt "Read carefully."
extend-prompt "Cite sources."
set-flag temperature 0
unknown ignored

-- tests/smoke --
QUESTION=what?

-- examples/example-1/QUESTION --
What is Go?

-- examples/example-1/ideal-output --
A language.

-- variable-description/QUESTION --
The user question.`

	p, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if p.Shebang == "" || p.Flags["provider"] != "openai" || p.Flags["stream"] != "true" || p.Flags["model"] != "gpt-4" {
		t.Fatalf("flags = %#v shebang=%q", p.Flags, p.Shebang)
	}
	if p.PromptSummary != "Answers questions." {
		t.Fatalf("summary = %q", p.PromptSummary)
	}
	if p.Config.Prefill != "Start here" || len(p.Config.StopSequence) != 2 {
		t.Fatalf("config = %#v", p.Config)
	}
	if got, ok := p.GetTest("smoke"); !ok || got == "" {
		t.Fatalf("test = %q %v", got, ok)
	}
	if got, ok := p.GetVariant("strict"); !ok || got == "" {
		t.Fatalf("variant = %q %v", got, ok)
	}
	if names := p.ListExamples(); len(names) != 1 || names[0] != "example-1" {
		t.Fatalf("examples = %#v", names)
	}
	vars, ideal, ok := p.GetExample("example-1")
	if !ok || vars["QUESTION"] != "What is Go?" || ideal != "A language." {
		t.Fatalf("example vars=%#v ideal=%q ok=%v", vars, ideal, ok)
	}
	if _, _, ok := p.GetExample("missing"); ok {
		t.Fatal("missing example exists")
	}
	if p.VariableDescriptions["QUESTION"] != "The user question." {
		t.Fatalf("descriptions = %#v", p.VariableDescriptions)
	}

	strict, err := p.ApplyVariant("strict")
	if err != nil {
		t.Fatal(err)
	}
	if strict.Flags["temperature"] != "0" || !containsAll(strict.SystemPrompt, "Follow policy.", "No guessing.") {
		t.Fatalf("strict = %#v", strict)
	}
	if !containsAll(strict.Main, "Read carefully.", "Cite sources.") {
		t.Fatalf("strict main = %q", strict.Main)
	}
	if strict.Examples["example-1"]["QUESTION"] != "What is Go?" {
		t.Fatalf("strict examples = %#v", strict.Examples)
	}
	if _, err := p.ApplyVariant("missing"); err == nil {
		t.Fatal("missing variant succeeded")
	}
	if formatted := strict.Format(); !containsAll(formatted, "System:", "User:") {
		t.Fatalf("formatted = %q", formatted)
	}
	if strict.Minimal() != strict.Main {
		t.Fatalf("minimal = %q", strict.Minimal())
	}
}

func TestPromptParseConfigDirectPrefill(t *testing.T) {
	p := &Prompt{Sections: map[string]string{}}
	p.ParseConfig(`
# comment
prefill "direct text"
stop-sequence "\n\n"
bad
`)
	if p.Config.Prefill != "direct text" || len(p.Config.StopSequence) != 1 {
		t.Fatalf("config = %#v", p.Config)
	}
	p.Defaults = map[string]string{}
	p.ParseDefaults(`
ignored
quoted="value"
`)
	if p.Defaults["quoted"] != "value" {
		t.Fatalf("defaults = %#v", p.Defaults)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
