package evaluator

import (
	"context"
	"strings"
	"testing"
)

func TestEvaluateAnswerRelevance(t *testing.T) {
	threshold := 0.7
	provider := &assertionJudgeTestProvider{name: "judge", model: "test", response: "SCORE: 9\nREASONING: on topic"}
	evaluator := NewAssertionEvaluator(provider)

	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:      AssertionAnswerRelevance,
		Threshold: &threshold,
	}, "Paris is the capital of France.", map[string]interface{}{"input": "What is the capital of France?"})
	if err != nil {
		t.Fatalf("answer-relevance: %v", err)
	}
	if !result.Passed || result.Score != 0.9 {
		t.Fatalf("answer-relevance result = %+v, want pass at 0.9", result)
	}
	if !strings.Contains(provider.lastPrompt, "QUESTION:") || !strings.Contains(provider.lastPrompt, "capital of France") {
		t.Fatalf("answer-relevance prompt missing the question:\n%s", provider.lastPrompt)
	}
}

func TestAnswerRelevanceRequiresQuestion(t *testing.T) {
	provider := &assertionJudgeTestProvider{name: "judge", model: "test", response: "SCORE: 9"}
	evaluator := NewAssertionEvaluator(provider)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionAnswerRelevance,
	}, "answer", nil) // no input in metadata
	if err != nil {
		t.Fatalf("answer-relevance: %v", err)
	}
	if result.Passed {
		t.Fatal("answer-relevance without a question should not pass")
	}
}

func TestEvaluateContextMetrics(t *testing.T) {
	threshold := 0.6
	for _, typ := range []AssertionType{
		AssertionContextFaithfulness,
		AssertionContextRecall,
		AssertionContextRelevance,
	} {
		t.Run(string(typ), func(t *testing.T) {
			provider := &assertionJudgeTestProvider{name: "judge", model: "test", response: "SCORE: 8\nREASONING: ok"}
			evaluator := NewAssertionEvaluator(provider)
			result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
				Type:      typ,
				Threshold: &threshold,
			}, "The Eiffel Tower is in Paris.", map[string]interface{}{
				"input":   "Where is the Eiffel Tower?",
				"context": "The Eiffel Tower is a landmark located in Paris, France.",
			})
			if err != nil {
				t.Fatalf("%s: %v", typ, err)
			}
			if !result.Passed || result.Score != 0.8 {
				t.Fatalf("%s result = %+v, want pass at 0.8", typ, result)
			}
			if !strings.Contains(provider.lastPrompt, "CONTEXT:") {
				t.Fatalf("%s prompt missing context section:\n%s", typ, provider.lastPrompt)
			}
		})
	}
}

func TestContextMetricRequiresContext(t *testing.T) {
	provider := &assertionJudgeTestProvider{name: "judge", model: "test", response: "SCORE: 9"}
	evaluator := NewAssertionEvaluator(provider)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionContextFaithfulness,
	}, "answer", map[string]interface{}{"input": "q"}) // no context var
	if err != nil {
		t.Fatalf("context-faithfulness: %v", err)
	}
	if result.Passed {
		t.Fatal("context-faithfulness without a context var should not pass")
	}
}

func TestRAGMetricsRequireProvider(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)
	_, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionAnswerRelevance,
	}, "answer", map[string]interface{}{"input": "q"})
	if err == nil {
		t.Fatal("answer-relevance without a judge provider should error")
	}
}
