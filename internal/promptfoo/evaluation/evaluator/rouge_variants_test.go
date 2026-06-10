package evaluator

import (
	"context"
	"testing"
)

func TestRougeLHelper(t *testing.T) {
	tests := []struct {
		name   string
		output string
		ref    string
		want   float64
	}{
		{"identical", "a b c", "a b c", 1.0},
		{"no-overlap", "d e f", "a b c", 0.0},
		// cand LCS=3/3, ref 3/6 -> P=1,R=0.5,F1=0.6667
		{"subsequence", "the cat sat", "the cat sat on the mat", 2.0 / 3.0},
		// LCS of "a c" within "a b c" is 2; P=2/2=1, R=2/3, F1=0.8
		{"gapped", "a c", "a b c", 0.8},
	}
	for _, tc := range tests {
		if got := rougeL(tc.output, tc.ref); !almostEqual(got, tc.want) {
			t.Errorf("rougeL(%q,%q) = %v, want %v", tc.output, tc.ref, got, tc.want)
		}
	}
}

func TestLCSLength(t *testing.T) {
	tests := []struct {
		a, b []string
		want int
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}, 3},
		{[]string{"a", "b", "c"}, []string{"d", "e", "f"}, 0},
		{[]string{"a", "b", "c", "d"}, []string{"a", "c", "d"}, 3},
		{[]string{}, []string{"a"}, 0},
	}
	for _, tc := range tests {
		if got := lcsLength(tc.a, tc.b); got != tc.want {
			t.Errorf("lcsLength(%v,%v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestRougeSHelper(t *testing.T) {
	// "a b c" skip-bigrams: {a-b, a-c, b-c} (3). Identical -> 1.0.
	if got := rougeS("a b c", "a b c"); !almostEqual(got, 1.0) {
		t.Errorf("rougeS identical = %v, want 1.0", got)
	}
	// "c b a": {c-b, c-a, b-a}; "a b c": {a-b, a-c, b-c}; no ordered pair
	// survives the reversal, so overlap is 0.
	if got := rougeS("c b a", "a b c"); !almostEqual(got, 0.0) {
		t.Errorf("rougeS reversed = %v, want 0.0 (no ordered pair survives)", got)
	}
	// "a b" vs "a b c": cand {a-b}(1), ref {a-b,a-c,b-c}(3), overlap 1.
	// P=1/1=1, R=1/3, F1=2*1*(1/3)/(1+1/3)=0.5.
	if got := rougeS("a b", "a b c"); !almostEqual(got, 0.5) {
		t.Errorf("rougeS subset = %v, want 0.5", got)
	}
}

func TestEvaluateRougeVariants(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name      string
		typ       AssertionType
		output    string
		ref       string
		threshold *float64
		wantPass  bool
	}{
		{"rouge-l-pass", AssertionRougeL, "the cat sat on the mat", "the cat sat on the mat", nil, true},
		{"rouge-l-fail-default", AssertionRougeL, "the cat sat", "the cat sat on the mat", nil, false}, // 0.667 < 0.75
		{"rouge-l-low-threshold", AssertionRougeL, "the cat sat", "the cat sat on the mat", floatPtr(0.5), true},
		{"rouge-s-pass", AssertionRougeS, "a b c", "a b c", nil, true},
		{"rouge-s-fail", AssertionRougeS, "a b", "a b c d e", floatPtr(0.75), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: tc.typ, Value: tc.ref, Threshold: tc.threshold}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.wantPass {
				t.Errorf("%s: Passed = %v, want %v (score %v, %s)", tc.typ, got.Passed, tc.wantPass, got.Score, got.Message)
			}
		})
	}
}

func TestRougeVariantAliases(t *testing.T) {
	for id, want := range map[string]AssertionType{"rouge-l": AssertionRougeL, "rouge-s": AssertionRougeS} {
		canonical, _, ok := normalizeAssertionType(id)
		if !ok || canonical != want {
			t.Errorf("%q => (%q, ok=%v), want %q", id, canonical, ok, want)
		}
	}
}
