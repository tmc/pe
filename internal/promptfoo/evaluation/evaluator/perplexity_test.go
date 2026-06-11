package evaluator

import (
	"context"
	"math"
	"testing"
)

func TestPerplexityHelper(t *testing.T) {
	// All-confident tokens (logprob 0) -> perplexity 1.
	if pp := perplexity([]float64{0, 0, 0}); !almostEqual(pp, 1.0) {
		t.Errorf("perplexity(zeros) = %v, want 1.0", pp)
	}
	// Uniform logprob -ln(2) -> perplexity 2.
	lp := -math.Log(2)
	if pp := perplexity([]float64{lp, lp}); !almostEqual(pp, 2.0) {
		t.Errorf("perplexity(-ln2) = %v, want 2.0", pp)
	}
	// Empty -> +Inf.
	if pp := perplexity(nil); !math.IsInf(pp, 1) {
		t.Errorf("perplexity(nil) = %v, want +Inf", pp)
	}
}

func TestEvaluatePerplexity(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	lp := -math.Log(2) // perplexity 2
	meta := map[string]interface{}{"logprobs": []interface{}{lp, lp}}

	// No bound: reports and passes.
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionPerplexity}, "out", meta); !got.Passed {
		t.Errorf("perplexity with no bound should pass (%s)", got.Message)
	}
	// Max 3 >= 2: pass. Max 1 < 2: fail.
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionPerplexity, Max: floatPtr(3)}, "out", meta); !got.Passed {
		t.Errorf("perplexity<=3 should pass for pp=2 (%s)", got.Message)
	}
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionPerplexity, Max: floatPtr(1)}, "out", meta); got.Passed {
		t.Errorf("perplexity<=1 should fail for pp=2")
	}
	// Missing logprobs fails clearly.
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionPerplexity}, "out", nil); got.Passed {
		t.Errorf("perplexity with no logprobs should fail")
	}
}

func TestEvaluatePerplexityScore(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	// Confident output (logprob 0) -> perplexity 1 -> score 0.5.
	meta := map[string]interface{}{"logprobs": []interface{}{float64(0), float64(0)}}
	got, err := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionPerplexityScore}, "out", meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(got.Score, 0.5) {
		t.Errorf("perplexity-score = %v, want 0.5", got.Score)
	}
	// Default threshold 0.5: 0.5 >= 0.5 passes.
	if !got.Passed {
		t.Errorf("perplexity-score 0.5 should pass at default threshold")
	}
	// Higher threshold fails.
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionPerplexityScore, Threshold: floatPtr(0.6)}, "out", meta); got.Passed {
		t.Errorf("perplexity-score 0.5 should fail at threshold 0.6")
	}
}

func TestPerplexityAliases(t *testing.T) {
	for id, want := range map[string]AssertionType{
		"perplexity":       AssertionPerplexity,
		"perplexity-score": AssertionPerplexityScore,
	} {
		canonical, _, ok := normalizeAssertionType(id)
		if !ok || canonical != want {
			t.Errorf("%q => (%q, ok=%v), want %q", id, canonical, ok, want)
		}
	}
}
