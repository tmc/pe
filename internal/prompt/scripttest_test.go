package prompt

import (
	"testing"
)

func TestParseScriptTests(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ScriptTest
	}{
		{
			name: "simple test",
			input: `$ text="Hello"
World`,
			expected: []ScriptTest{
				{
					Variables: map[string]string{"text": "Hello"},
					Expected:  "World",
				},
			},
		},
		{
			name: "test with assertions",
			input: `$ input="test"
Output line
> contains: test
> min_length: 5`,
			expected: []ScriptTest{
				{
					Variables: map[string]string{"input": "test"},
					Expected:  "Output line",
					Assertions: []ScriptTestAssertion{
						{Type: "contains", Value: "test"},
						{Type: "min_length", Value: "5"},
					},
				},
			},
		},
		{
			name: "multiple tests",
			input: `$ a=1
One

$ b=2
Two
> equals: Two`,
			expected: []ScriptTest{
				{
					Variables: map[string]string{"a": "1"},
					Expected:  "One",
				},
				{
					Variables: map[string]string{"b": "2"},
					Expected:  "Two",
					Assertions: []ScriptTestAssertion{
						{Type: "equals", Value: "Two"},
					},
				},
			},
		},
		{
			name: "quoted values",
			input: `$ text="Hello world" name="John Doe"
Greeting for John`,
			expected: []ScriptTest{
				{
					Variables: map[string]string{
						"text": "Hello world",
						"name": "John Doe",
					},
					Expected: "Greeting for John",
				},
			},
		},
		{
			name: "multiline output",
			input: `$ prompt="test"
Line one
Line two
Line three`,
			expected: []ScriptTest{
				{
					Variables: map[string]string{"prompt": "test"},
					Expected:  "Line one\nLine two\nLine three",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseScriptTests(tt.input)
			if err != nil {
				t.Fatalf("ParseScriptTests() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d tests, got %d", len(tt.expected), len(result))
			}

			for i, test := range result {
				expected := tt.expected[i]

				// Check variables
				if len(test.Variables) != len(expected.Variables) {
					t.Errorf("Test %d: expected %d variables, got %d",
						i, len(expected.Variables), len(test.Variables))
				}
				for k, v := range expected.Variables {
					if test.Variables[k] != v {
						t.Errorf("Test %d: variable %s: expected %q, got %q",
							i, k, v, test.Variables[k])
					}
				}

				// Check expected output
				if test.Expected != expected.Expected {
					t.Errorf("Test %d: expected output %q, got %q",
						i, expected.Expected, test.Expected)
				}

				// Check assertions
				if len(test.Assertions) != len(expected.Assertions) {
					t.Errorf("Test %d: expected %d assertions, got %d",
						i, len(expected.Assertions), len(test.Assertions))
				}
			}
		})
	}
}

func TestParseVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "simple",
			input:    "a=1 b=2",
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:     "quoted",
			input:    `text="Hello world" name="John"`,
			expected: map[string]string{"text": "Hello world", "name": "John"},
		},
		{
			name:     "mixed",
			input:    `id=123 message="Test message" flag=true`,
			expected: map[string]string{"id": "123", "message": "Test message", "flag": "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := make(map[string]string)
			err := parseVariables(tt.input, vars)
			if err != nil {
				t.Fatalf("parseVariables() error = %v", err)
			}

			if len(vars) != len(tt.expected) {
				t.Fatalf("Expected %d variables, got %d", len(tt.expected), len(vars))
			}

			for k, v := range tt.expected {
				if vars[k] != v {
					t.Errorf("Variable %s: expected %q, got %q", k, v, vars[k])
				}
			}
		})
	}
}
