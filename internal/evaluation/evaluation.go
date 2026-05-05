package evaluation

import (
	"context"
	"fmt"
)

// Result is the outcome of evaluating one input/output pair.
type Result struct {
	Score   float64           `json:"score"`
	Passed  bool              `json:"passed"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// Evaluator evaluates output for input.
type Evaluator interface {
	Evaluate(context.Context, string, string) (Result, error)
}

// MetricFunc computes a score for input and output.
type MetricFunc func(input, output string) (float64, error)

// MetricEvaluator evaluates with a numeric metric.
type MetricEvaluator struct {
	Threshold float64
	Metric    MetricFunc
}

// Evaluate implements Evaluator.
func (e MetricEvaluator) Evaluate(ctx context.Context, input, output string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if e.Metric == nil {
		return Result{}, fmt.Errorf("metric evaluator requires metric")
	}
	score, err := e.Metric(input, output)
	if err != nil {
		return Result{}, err
	}
	return Result{Score: score, Passed: score >= e.Threshold}, nil
}

// LLMJudge judges output with an LLM or compatible scorer.
type LLMJudge interface {
	Judge(context.Context, string, string) (Result, error)
}

// LLMEvaluator evaluates with an LLM judge.
type LLMEvaluator struct {
	Judge LLMJudge
}

// Evaluate implements Evaluator.
func (e LLMEvaluator) Evaluate(ctx context.Context, input, output string) (Result, error) {
	if e.Judge == nil {
		return Result{}, fmt.Errorf("llm evaluator requires judge")
	}
	return e.Judge.Judge(ctx, input, output)
}

// HumanReviewFunc records a human review.
type HumanReviewFunc func(context.Context, string, string) (Result, error)

// HumanEvaluator evaluates with a human review callback.
type HumanEvaluator struct {
	Review HumanReviewFunc
}

// Evaluate implements Evaluator.
func (e HumanEvaluator) Evaluate(ctx context.Context, input, output string) (Result, error) {
	if e.Review == nil {
		return Result{}, fmt.Errorf("human evaluator requires review")
	}
	return e.Review(ctx, input, output)
}

// CompositeEvaluator combines multiple evaluators.
type CompositeEvaluator struct {
	Evaluators []Evaluator
}

// Evaluate implements Evaluator.
func (e CompositeEvaluator) Evaluate(ctx context.Context, input, output string) (Result, error) {
	if len(e.Evaluators) == 0 {
		return Result{}, fmt.Errorf("composite evaluator requires evaluators")
	}
	var total float64
	for _, evaluator := range e.Evaluators {
		result, err := evaluator.Evaluate(ctx, input, output)
		if err != nil {
			return Result{}, err
		}
		total += result.Score
	}
	score := total / float64(len(e.Evaluators))
	return Result{Score: score, Passed: score >= 1}, nil
}

// Case is one evaluation case.
type Case struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// Pipeline runs an evaluator over cases.
type Pipeline struct {
	Evaluator Evaluator
}

// Run evaluates all cases.
func (p Pipeline) Run(ctx context.Context, cases []Case) ([]Result, error) {
	if p.Evaluator == nil {
		return nil, fmt.Errorf("pipeline requires evaluator")
	}
	results := make([]Result, 0, len(cases))
	for _, c := range cases {
		result, err := p.Evaluator.Evaluate(ctx, c.Input, c.Output)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
