package evaluation

import (
	"context"
	"testing"
)

type judgeFunc func(context.Context, string, string) (Result, error)

func (f judgeFunc) Judge(ctx context.Context, input, output string) (Result, error) {
	return f(ctx, input, output)
}

func TestMetricEvaluator(t *testing.T) {
	evaluator := MetricEvaluator{
		Threshold: 0.5,
		Metric: func(input, output string) (float64, error) {
			return 0.75, nil
		},
	}
	result, err := evaluator.Evaluate(context.Background(), "in", "out")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !result.Passed || result.Score != 0.75 {
		t.Fatalf("result = %+v, want pass at 0.75", result)
	}
}

func TestLLMEvaluator(t *testing.T) {
	evaluator := LLMEvaluator{Judge: judgeFunc(func(context.Context, string, string) (Result, error) {
		return Result{Score: 1, Passed: true}, nil
	})}
	result, err := evaluator.Evaluate(context.Background(), "in", "out")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !result.Passed {
		t.Fatal("llm result did not pass")
	}
}

func TestHumanEvaluator(t *testing.T) {
	evaluator := HumanEvaluator{Review: func(context.Context, string, string) (Result, error) {
		return Result{Score: 1, Passed: true, Message: "approved"}, nil
	}}
	result, err := evaluator.Evaluate(context.Background(), "in", "out")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result.Message != "approved" {
		t.Fatalf("message = %q, want approved", result.Message)
	}
}

func TestCompositeEvaluator(t *testing.T) {
	evaluator := CompositeEvaluator{Evaluators: []Evaluator{
		MetricEvaluator{Threshold: 0, Metric: func(string, string) (float64, error) { return 1, nil }},
		MetricEvaluator{Threshold: 0, Metric: func(string, string) (float64, error) { return 0.5, nil }},
	}}
	result, err := evaluator.Evaluate(context.Background(), "in", "out")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result.Score != 0.75 {
		t.Fatalf("score = %g, want 0.75", result.Score)
	}
}

func TestPipeline(t *testing.T) {
	pipeline := Pipeline{Evaluator: MetricEvaluator{
		Threshold: 1,
		Metric: func(input, output string) (float64, error) {
			if input == output {
				return 1, nil
			}
			return 0, nil
		},
	}}
	results, err := pipeline.Run(context.Background(), []Case{{Input: "a", Output: "a"}, {Input: "a", Output: "b"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(results) != 2 || !results[0].Passed || results[1].Passed {
		t.Fatalf("results = %+v", results)
	}
}
