package evaluator

import (
	"context"
	"strings"
	"testing"
)

func TestEvaluateGEval(t *testing.T) {
	threshold := 0.7
	provider := &assertionJudgeTestProvider{
		name:     "judge",
		model:    "test",
		response: "STEPS: check coherence\nREASONING: it reads well\nSCORE: 9",
	}
	evaluator := NewAssertionEvaluator(provider)

	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:      AssertionGEval,
		Value:     "the answer is coherent and correct",
		Threshold: &threshold,
	}, "Paris is the capital of France.", nil)
	if err != nil {
		t.Fatalf("EvaluateAssertion g-eval: %v", err)
	}
	if !result.Passed {
		t.Fatalf("g-eval should pass at score 9/10 with threshold 0.7, got %v", result)
	}
	if result.Score != 0.9 {
		t.Fatalf("score = %v, want 0.9", result.Score)
	}
	if result.Metadata["method"] != "g_eval_cot" {
		t.Fatalf("method = %v, want g_eval_cot", result.Metadata["method"])
	}
	// It must be a real chain-of-thought prompt (asks for steps).
	if !strings.Contains(provider.lastPrompt, "G-Eval") || !strings.Contains(provider.lastPrompt, "evaluation steps") {
		t.Fatalf("g-eval prompt is not a chain-of-thought rubric prompt:\n%s", provider.lastPrompt)
	}
}

func TestEvaluateGEvalBelowThresholdFails(t *testing.T) {
	threshold := 0.8
	provider := &assertionJudgeTestProvider{
		name:     "judge",
		model:    "test",
		response: "STEPS: x\nREASONING: weak\nSCORE: 5",
	}
	evaluator := NewAssertionEvaluator(provider)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:      AssertionGEval,
		Value:     "be excellent",
		Threshold: &threshold,
	}, "meh", nil)
	if err != nil {
		t.Fatalf("EvaluateAssertion g-eval: %v", err)
	}
	if result.Passed {
		t.Fatalf("g-eval should fail at 0.5 vs threshold 0.8, got %v", result)
	}
}

func TestEvaluateGEvalRequiresProvider(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)
	_, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:  AssertionGEval,
		Value: "be correct",
	}, "answer", nil)
	if err == nil {
		t.Fatal("g-eval without a judge provider should error")
	}
}

func TestEvaluateGEvalMissingCriteria(t *testing.T) {
	provider := &assertionJudgeTestProvider{name: "judge", model: "test", response: "SCORE: 9"}
	evaluator := NewAssertionEvaluator(provider)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:  AssertionGEval,
		Value: 123, // not a string
	}, "answer", nil)
	if err != nil {
		t.Fatalf("EvaluateAssertion g-eval: %v", err)
	}
	if result.Passed {
		t.Fatal("g-eval with non-string criteria should not pass")
	}
}
