package assertutil

import (
	"fmt"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		substr   string
		expected bool
	}{
		{
			name:     "simple substring match",
			input:    "This is a test response.",
			substr:   "test",
			expected: true,
		},
		{
			name:     "substring at beginning",
			input:    "This is a test response.",
			substr:   "This",
			expected: true,
		},
		{
			name:     "substring at end",
			input:    "This is a test response.",
			substr:   "response.",
			expected: true,
		},
		{
			name:     "full string match",
			input:    "This is a test response.",
			substr:   "This is a test response.",
			expected: true,
		},
		{
			name:     "substring not present",
			input:    "This is a test response.",
			substr:   "banana",
			expected: false,
		},
		{
			name:     "case sensitive match",
			input:    "This is a test response.",
			substr:   "this",
			expected: false,
		},
		{
			name:     "empty substring always matches",
			input:    "This is a test response.",
			substr:   "",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := New(tc.input)
			got := a.Contains(tc.substr)
			if got != tc.expected {
				t.Errorf("Contains(%q) = %v, want %v", tc.substr, got, tc.expected)
			}

			// Verify result was recorded
			if len(a.Results) != 1 {
				t.Errorf("Expected 1 result, got %d", len(a.Results))
			} else if a.Results[0].Success != tc.expected {
				t.Errorf("Result.Success = %v, want %v", a.Results[0].Success, tc.expected)
			}
		})
	}
}

func TestNotContains(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		substr   string
		expected bool
	}{
		{
			name:     "substring not present",
			input:    "This is a test response.",
			substr:   "banana",
			expected: true,
		},
		{
			name:     "substring present",
			input:    "This is a test response.",
			substr:   "test",
			expected: false,
		},
		{
			name:     "case sensitive match",
			input:    "This is a test response.",
			substr:   "this",
			expected: true,
		},
		{
			name:     "empty substring should not match",
			input:    "This is a test response.",
			substr:   "",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := New(tc.input)
			got := a.NotContains(tc.substr)
			if got != tc.expected {
				t.Errorf("NotContains(%q) = %v, want %v", tc.substr, got, tc.expected)
			}

			// Verify result was recorded
			if len(a.Results) != 1 {
				t.Errorf("Expected 1 result, got %d", len(a.Results))
			} else if a.Results[0].Success != tc.expected {
				t.Errorf("Result.Success = %v, want %v", a.Results[0].Success, tc.expected)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		want     bool
	}{
		{
			name:     "exact match",
			input:    "This is a test response.",
			expected: "This is a test response.",
			want:     true,
		},
		{
			name:     "different content",
			input:    "This is a test response.",
			expected: "This is a different response.",
			want:     false,
		},
		{
			name:     "different case",
			input:    "This is a test response.",
			expected: "this is a test response.",
			want:     false,
		},
		{
			name:     "substring match but not equal",
			input:    "This is a test response.",
			expected: "test",
			want:     false,
		},
		{
			name:     "empty string matches empty string",
			input:    "",
			expected: "",
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := New(tc.input)
			got := a.Equals(tc.expected)
			if got != tc.want {
				t.Errorf("Equals(%q) = %v, want %v", tc.expected, got, tc.want)
			}

			// Verify result was recorded
			if len(a.Results) != 1 {
				t.Errorf("Expected 1 result, got %d", len(a.Results))
			} else if a.Results[0].Success != tc.want {
				t.Errorf("Result.Success = %v, want %v", a.Results[0].Success, tc.want)
			}
		})
	}
}

func TestStartsEndsWith(t *testing.T) {
	input := "This is a test response."

	t.Run("StartsWith", func(t *testing.T) {
		tests := []struct {
			prefix string
			want   bool
		}{
			{"This", true},
			{"This is", true},
			{"This is a test response.", true},
			{"this", false}, // case sensitive
			{"test", false},
			{"", true}, // empty prefix always matches
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("prefix=%q", tc.prefix), func(t *testing.T) {
				a := New(input)
				got := a.StartsWith(tc.prefix)
				if got != tc.want {
					t.Errorf("StartsWith(%q) = %v, want %v", tc.prefix, got, tc.want)
				}
			})
		}
	})

	t.Run("EndsWith", func(t *testing.T) {
		tests := []struct {
			suffix string
			want   bool
		}{
			{"response.", true},
			{"test response.", true},
			{"This is a test response.", true},
			{"Response.", false}, // case sensitive
			{"test", false},
			{"", true}, // empty suffix always matches
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("suffix=%q", tc.suffix), func(t *testing.T) {
				a := New(input)
				got := a.EndsWith(tc.suffix)
				if got != tc.want {
					t.Errorf("EndsWith(%q) = %v, want %v", tc.suffix, got, tc.want)
				}
			})
		}
	})
}

func TestMatchesRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		want    bool
	}{
		{
			name:    "simple pattern match",
			input:   "This is a test response.",
			pattern: "test",
			want:    true,
		},
		{
			name:    "beginning anchor",
			input:   "This is a test response.",
			pattern: "^This",
			want:    true,
		},
		{
			name:    "end anchor",
			input:   "This is a test response.",
			pattern: "response\\.$",
			want:    true,
		},
		{
			name:    "character class",
			input:   "This is a test response.",
			pattern: "is [a-z] test",
			want:    true,
		},
		{
			name:    "no match",
			input:   "This is a test response.",
			pattern: "^test",
			want:    false,
		},
		{
			name:    "case insensitive",
			input:   "This is a test response.",
			pattern: "(?i)this",
			want:    true,
		},
		{
			name:    "invalid pattern",
			input:   "This is a test response.",
			pattern: "[unclosed",
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := New(tc.input)
			got := a.MatchesRegex(tc.pattern)
			if got != tc.want {
				t.Errorf("MatchesRegex(%q) = %v, want %v", tc.pattern, got, tc.want)
			}
		})
	}
}

func TestLengthAssertions(t *testing.T) {
	input := "This is a test response."
	inputLen := len(input)

	t.Run("LengthEquals", func(t *testing.T) {
		a := New(input)
		if !a.LengthEquals(inputLen) {
			t.Errorf("LengthEquals(%d) should pass", inputLen)
		}
		if a.LengthEquals(inputLen + 1) {
			t.Errorf("LengthEquals(%d) should fail", inputLen+1)
		}
	})

	t.Run("LengthGreaterThan", func(t *testing.T) {
		a := New(input)
		if !a.LengthGreaterThan(inputLen - 1) {
			t.Errorf("LengthGreaterThan(%d) should pass", inputLen-1)
		}
		if a.LengthGreaterThan(inputLen) {
			t.Errorf("LengthGreaterThan(%d) should fail", inputLen)
		}
		if a.LengthGreaterThan(inputLen + 1) {
			t.Errorf("LengthGreaterThan(%d) should fail", inputLen+1)
		}
	})

	t.Run("LengthLessThan", func(t *testing.T) {
		a := New(input)
		if !a.LengthLessThan(inputLen + 1) {
			t.Errorf("LengthLessThan(%d) should pass", inputLen+1)
		}
		if a.LengthLessThan(inputLen) {
			t.Errorf("LengthLessThan(%d) should fail", inputLen)
		}
		if a.LengthLessThan(inputLen - 1) {
			t.Errorf("LengthLessThan(%d) should fail", inputLen-1)
		}
	})

	t.Run("LengthInRange", func(t *testing.T) {
		a := New(input)
		if !a.LengthInRange(inputLen, inputLen) {
			t.Errorf("LengthInRange(%d, %d) should pass", inputLen, inputLen)
		}
		if !a.LengthInRange(inputLen-1, inputLen+1) {
			t.Errorf("LengthInRange(%d, %d) should pass", inputLen-1, inputLen+1)
		}
		if a.LengthInRange(0, inputLen-1) {
			t.Errorf("LengthInRange(%d, %d) should fail", 0, inputLen-1)
		}
		if a.LengthInRange(inputLen+1, inputLen+2) {
			t.Errorf("LengthInRange(%d, %d) should fail", inputLen+1, inputLen+2)
		}
	})
}

func TestJSONAssertions(t *testing.T) {
	t.Run("IsValidJSON", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  bool
		}{
			{
				name:  "valid simple object",
				input: `{"name": "John", "age": 30}`,
				want:  true,
			},
			{
				name:  "valid array",
				input: `[1, 2, 3, 4]`,
				want:  true,
			},
			{
				name:  "valid complex object",
				input: `{"name": "John", "age": 30, "address": {"street": "123 Main St", "city": "New York"}, "phoneNumbers": [{"type": "home", "number": "212-555-1234"}, {"type": "work", "number": "646-555-4567"}]}`,
				want:  true,
			},
			{
				name:  "valid boolean",
				input: `true`,
				want:  true,
			},
			{
				name:  "valid number",
				input: `42`,
				want:  true,
			},
			{
				name:  "valid string",
				input: `"hello"`,
				want:  true,
			},
			{
				name:  "valid null",
				input: `null`,
				want:  true,
			},
			{
				name:  "invalid syntax",
				input: `{"name": "John", "age": 30,}`,
				want:  false,
			},
			{
				name:  "incomplete object",
				input: `{"name": "John", "age": `,
				want:  false,
			},
			{
				name:  "plain text",
				input: `This is not JSON`,
				want:  false,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				a := New(tc.input)
				got := a.IsValidJSON()
				if got != tc.want {
					t.Errorf("IsValidJSON() = %v, want %v", got, tc.want)
				}
			})
		}
	})

	t.Run("MatchesJSONSchema", func(t *testing.T) {
		schema := `{
			"type": "object",
			"required": ["name", "age"],
			"properties": {
				"name": { "type": "string" },
				"age": { "type": "integer", "minimum": 0 },
				"email": { "type": "string", "format": "email" }
			}
		}`

		tests := []struct {
			name  string
			input string
			want  bool
		}{
			{
				name:  "valid schema match",
				input: `{"name": "John", "age": 30, "email": "john@example.com"}`,
				want:  true,
			},
			{
				name:  "valid minimal match",
				input: `{"name": "John", "age": 30}`,
				want:  true,
			},
			{
				name:  "missing required field",
				input: `{"name": "John"}`,
				want:  false,
			},
			{
				name:  "wrong type",
				input: `{"name": "John", "age": "thirty"}`,
				want:  false,
			},
			{
				name:  "invalid value",
				input: `{"name": "John", "age": -5}`,
				want:  false,
			},
			{
				name:  "invalid email format",
				input: `{"name": "John", "age": 30, "email": "not-an-email"}`,
				want:  false,
			},
			{
				name:  "not an object",
				input: `["John", 30]`,
				want:  false,
			},
			{
				name:  "invalid json",
				input: `{"name": "John", "age": 30,}`,
				want:  false,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				a := New(tc.input)
				got := a.MatchesJSONSchema(schema)
				if got != tc.want {
					t.Errorf("MatchesJSONSchema() = %v, want %v", got, tc.want)
				}
			})
		}
	})
}

func TestCustomAssertion(t *testing.T) {
	input := "The answer is 42 and that's final."

	t.Run("Custom passing assertion", func(t *testing.T) {
		a := New(input)
		got := a.Custom("contains number", func(s string) (bool, string) {
			matched := false
			for _, c := range s {
				if c >= '0' && c <= '9' {
					matched = true
					break
				}
			}
			if matched {
				return true, "input contains a number"
			}
			return false, "input does not contain any numbers"
		})

		if !got {
			t.Errorf("Custom assertion should pass")
		}
		if a.Results[0].Reason != "input contains a number" {
			t.Errorf("Custom assertion reason should be set correctly, got: %s", a.Results[0].Reason)
		}
	})

	t.Run("Custom failing assertion", func(t *testing.T) {
		a := New("This has no numbers")
		got := a.Custom("contains number", func(s string) (bool, string) {
			matched := false
			for _, c := range s {
				if c >= '0' && c <= '9' {
					matched = true
					break
				}
			}
			if matched {
				return true, "input contains a number"
			}
			return false, "input does not contain any numbers"
		})

		if got {
			t.Errorf("Custom assertion should fail")
		}
		if a.Results[0].Reason != "input does not contain any numbers" {
			t.Errorf("Custom assertion reason should be set correctly, got: %s", a.Results[0].Reason)
		}
	})

	t.Run("Multiple assertions", func(t *testing.T) {
		a := New(input)
		a.Contains("42")
		a.Contains("not found")
		a.MatchesRegex(`\d+`)

		if a.AllPassed() {
			t.Errorf("Not all assertions should pass")
		}

		failures := a.GetFailures()
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure, got %d", len(failures))
		}
		if failures[0].Assertion.Value != "not found" {
			t.Errorf("Expected failure for 'not found', got failure for %v", failures[0].Assertion.Value)
		}

		results := a.GetResults()
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})
}