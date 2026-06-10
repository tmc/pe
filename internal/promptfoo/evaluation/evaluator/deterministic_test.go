package evaluator

import (
	"context"
	"math"
	"testing"
)

func floatPtr(f float64) *float64 { return &f }

// almostEqual reports whether a and b are within a small epsilon, for comparing
// floating-point assertion scores.
func almostEqual(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestEvaluateLevenshtein(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name      string
		output    string
		value     string
		threshold *float64
		want      bool
	}{
		{"identical", "hello", "hello", nil, true},
		{"within-default", "hello", "hella", nil, true},   // distance 1 <= 5
		{"at-default-edge", "abcdef", "uvwxyz", nil, false}, // distance 6 > 5
		{"custom-threshold-pass", "kitten", "sitting", floatPtr(3), true},
		{"custom-threshold-fail", "kitten", "sitting", floatPtr(2), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionLevenshtein, Value: tc.value, Threshold: tc.threshold}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (message: %s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"café", "cafe", 1}, // rune-aware
	}
	for _, tc := range tests {
		if got := levenshtein(tc.a, tc.b); got != tc.want {
			t.Errorf("levenshtein(%q,%q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestEvaluateRougeN(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name      string
		output    string
		value     string
		n         int // 0 means default (config omitted)
		threshold *float64
		wantScore float64
		wantPass  bool
	}{
		// Values cross-checked against js-rouge (the routine promptfoo calls).
		{"unigram-default", "the cat sat", "the cat sat on the mat", 0, nil, 0.6666666666666666, false},
		{"bigram-config", "the cat sat", "the cat sat on the mat", 2, nil, 0.5714285714285715, false},
		{"identical", "hello world", "hello world", 0, nil, 1.0, true},
		{"case-sensitive-miss", "The Cat", "the cat", 0, nil, 0.0, false},
		{"no-overlap", "completely different", "the cat sat", 0, nil, 0.0, false},
		{"low-threshold-pass", "the cat sat", "the cat sat on the mat", 0, floatPtr(0.5), 0.6666666666666666, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionRougeN, Value: tc.value, Threshold: tc.threshold}
			if tc.n != 0 {
				a.Config = map[string]interface{}{"n": tc.n}
			}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !almostEqual(got.Score, tc.wantScore) {
				t.Errorf("Score = %v, want %v", got.Score, tc.wantScore)
			}
			if got.Passed != tc.wantPass {
				t.Errorf("Passed = %v, want %v", got.Passed, tc.wantPass)
			}
		})
	}
}

func TestEvaluateBLEU(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name     string
		output   string
		value    string
		wantPass bool
	}{
		{"identical", "the quick brown fox jumps", "the quick brown fox jumps", true},
		{"no-overlap", "completely unrelated text here", "the quick brown fox jumps", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionBLEU, Value: tc.value}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.wantPass {
				t.Errorf("Passed = %v, want %v (score %v)", got.Passed, tc.wantPass, got.Score)
			}
		})
	}
	// An identical candidate must score a perfect 1.0.
	if s := bleu("a b c d e", "a b c d e", 4); !almostEqual(s, 1.0) {
		t.Errorf("bleu(identical) = %v, want 1.0", s)
	}
}

func TestEvaluateWordCount(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name     string
		output   string
		value    interface{}
		min, max *float64
		want     bool
	}{
		{"exact-pass", "one two three", float64(3), nil, nil, true},
		{"exact-fail", "one two three", float64(4), nil, nil, false},
		{"min-only-pass", "one two three four", nil, floatPtr(3), nil, true},
		{"min-only-fail", "one two", nil, floatPtr(3), nil, false},
		{"max-only-pass", "one two", nil, nil, floatPtr(3), true},
		{"max-only-fail", "one two three four", nil, nil, floatPtr(3), false},
		{"range-pass", "one two three", nil, floatPtr(2), floatPtr(4), true},
		{"range-fail-high", "one two three four five", nil, floatPtr(2), floatPtr(4), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionWordCount, Value: tc.value, Min: tc.min, Max: tc.max}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (message: %s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestEvaluateIsValidFunctionCall(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"bare-valid", `{"name":"get_weather","arguments":"{\"city\":\"SF\"}"}`, true},
		{"nested-valid", `{"function_call":{"name":"f","arguments":"{}"}}`, true},
		{"missing-name", `{"arguments":"{}"}`, false},
		{"args-not-string", `{"name":"f","arguments":{"city":"SF"}}`, false},
		{"args-bad-json", `{"name":"f","arguments":"{not json}"}`, false},
		{"not-json", `not json at all`, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionIsValidFunctionCall}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (message: %s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestEvaluateIsValidToolsCall(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"array-valid", `[{"type":"function","function":{"name":"f","arguments":"{}"}}]`, true},
		{"nested-tool_calls", `{"tool_calls":[{"type":"function","function":{"name":"f","arguments":"{\"a\":1}"}}]}`, true},
		{"empty-array", `[]`, false},
		{"wrong-type", `[{"type":"other","function":{"name":"f","arguments":"{}"}}]`, false},
		{"missing-function", `[{"type":"function"}]`, false},
		{"bad-inner-args", `[{"type":"function","function":{"name":"f","arguments":"{bad}"}}]`, false},
		{"not-array", `{"foo":"bar"}`, false},
		{"not-json", `nope`, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionIsValidToolsCall}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (message: %s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

// TestDeterministicAliases confirms the new assertion ids resolve through the
// promptfoo->pe alias map so unmodified promptfoo configs route to these
// evaluators.
func TestDeterministicAliases(t *testing.T) {
	ids := []struct {
		id   string
		want AssertionType
	}{
		{"levenshtein", AssertionLevenshtein},
		{"rouge-n", AssertionRougeN},
		{"bleu", AssertionBLEU},
		{"word-count", AssertionWordCount},
		{"is-valid-openai-function-call", AssertionIsValidFunctionCall},
		{"is-valid-openai-tools-call", AssertionIsValidToolsCall},
	}
	for _, tc := range ids {
		canonical, _, ok := normalizeAssertionType(tc.id)
		if !ok {
			t.Errorf("%q did not resolve through the alias map", tc.id)
			continue
		}
		if canonical != tc.want {
			t.Errorf("%q resolved to %q, want %q", tc.id, canonical, tc.want)
		}
	}
}
