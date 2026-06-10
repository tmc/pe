package evaluator

import (
	"context"
	"testing"
)

func TestEvaluateFinishReason(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name     string
		expected interface{}
		actual   string // metadata finishReason; "" means absent
		want     bool
	}{
		{"match-stop", "stop", "stop", true},
		{"match-case-insensitive", "STOP", "stop", true},
		{"mismatch", "length", "stop", false},
		{"absent", "stop", "", false},
		{"non-string-value", 1, "stop", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta := map[string]interface{}{}
			if tc.actual != "" {
				meta["finishReason"] = tc.actual
			}
			a := Assertion{Type: AssertionFinishReason, Value: tc.expected}
			got, err := ae.EvaluateAssertion(context.Background(), a, "output", meta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (%s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestFinishReasonAlias(t *testing.T) {
	canonical, negate, ok := normalizeAssertionType("finish-reason")
	if !ok || negate || canonical != AssertionFinishReason {
		t.Errorf("finish-reason => (%q, negate=%v, ok=%v), want (finish-reason, false, true)", canonical, negate, ok)
	}
	// not-finish-reason should invert.
	if _, negate, ok := normalizeAssertionType("not-finish-reason"); !ok || !negate {
		t.Errorf("not-finish-reason: ok=%v negate=%v, want true,true", ok, negate)
	}
}
