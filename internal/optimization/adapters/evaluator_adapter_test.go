package adapters

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/optimization"
)

type evaluatorProvider struct {
	text string
	err  error
}

func (p evaluatorProvider) Name() string { return "eval" }

func (p evaluatorProvider) Model() string { return "model" }

func (p evaluatorProvider) SupportsStreaming() bool { return false }

func (p evaluatorProvider) SupportsBatch() bool { return false }

func (p evaluatorProvider) Generate(context.Context, string, optimization.GenerationOptions) (*optimization.GenerationResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &optimization.GenerationResponse{Text: p.text}, nil
}

func TestStandardEvaluatorEvaluatePromptAndSystem(t *testing.T) {
	evaluator := NewStandardEvaluator(evaluatorProvider{text: "Score: 1.2"}).(*StandardEvaluator)
	score, err := evaluator.EvaluatePrompt(context.Background(), "prompt", "objective")
	if err != nil {
		t.Fatal(err)
	}
	if score != 1 {
		t.Fatalf("score = %v", score)
	}
	evaluator.provider = evaluatorProvider{text: "-0.4"}
	score, err = evaluator.EvaluatePrompt(context.Background(), "prompt", "objective")
	if err != nil {
		t.Fatal(err)
	}
	if score != 0 {
		t.Fatalf("clamped score = %v", score)
	}

	evaluator.provider = evaluatorProvider{text: "0.75"}
	system := optimization.SystemDefinition{
		Name:        "sys",
		Description: "system",
		Components: []optimization.SystemComponent{
			{Name: "planner", Type: "prompt", Content: "Plan work"},
			{Name: "runner", Type: "agent", Content: "Run work"},
		},
		Dependencies: []optimization.ComponentDependency{{From: "planner", To: "runner", Type: "control"}},
	}
	score, err = evaluator.EvaluateSystem(context.Background(), system, "ship")
	if err != nil || score != 0.75 {
		t.Fatalf("system score=%v err=%v", score, err)
	}
}

func TestStandardEvaluatorComparativeAndErrors(t *testing.T) {
	evaluator := NewStandardEvaluator(evaluatorProvider{text: "WINNER: B\nSCORE_A: 0.4\nSCORE_B: 0.8\nREASONING: clearer"})
	result, err := evaluator.EvaluateComparative(context.Background(), "a", "b", "objective")
	if err != nil {
		t.Fatal(err)
	}
	if result.Winner != 2 || result.Score1 != 0.4 || result.Score2 != 0.8 || result.Reasoning != "clearer" {
		t.Fatalf("comparison = %#v", result)
	}

	want := errors.New("provider failed")
	evaluator = NewStandardEvaluator(evaluatorProvider{err: want})
	if _, err := evaluator.EvaluatePrompt(context.Background(), "p", "o"); err == nil || !strings.Contains(err.Error(), "failed to evaluate prompt") {
		t.Fatalf("prompt error = %v", err)
	}
	if _, err := evaluator.EvaluateComparative(context.Background(), "a", "b", "o"); err == nil || !strings.Contains(err.Error(), "failed to compare prompts") {
		t.Fatalf("comparative error = %v", err)
	}
	if _, err := evaluator.EvaluateSystem(context.Background(), optimization.SystemDefinition{}, "o"); err == nil || !strings.Contains(err.Error(), "failed to evaluate system") {
		t.Fatalf("system error = %v", err)
	}

	evaluator = NewStandardEvaluator(evaluatorProvider{text: "not a score"})
	if _, err := evaluator.EvaluatePrompt(context.Background(), "p", "o"); err == nil || !strings.Contains(err.Error(), "failed to parse evaluation score") {
		t.Fatalf("parse prompt error = %v", err)
	}
	if _, err := evaluator.EvaluateSystem(context.Background(), optimization.SystemDefinition{}, "o"); err == nil || !strings.Contains(err.Error(), "failed to parse system evaluation score") {
		t.Fatalf("parse system error = %v", err)
	}
}

func TestEvaluatorAdapterHelpers(t *testing.T) {
	for _, tt := range []struct {
		text string
		want float64
		err  bool
	}{
		{"0.5", 0.5, false},
		{"score 0.75 now", 0.75, false},
		{"\n\nbad", 0, true},
	} {
		got, err := parseScore(tt.text)
		if tt.err {
			if err == nil {
				t.Fatalf("parseScore(%q) succeeded", tt.text)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Fatalf("parseScore(%q) = %v err=%v", tt.text, got, err)
		}
	}
	for _, text := range []string{
		"WINNER: A\nSCORE_A: 0.9\nSCORE_B: 0.1\nREASONING: a",
		"WINNER: TIE\nSCORE_A: 0.5\nSCORE_B: 0.5\nREASONING: tie",
		"SCORE_A: 0.1\nSCORE_B: 0.2",
	} {
		if _, err := parseComparisonResult(text); err != nil {
			t.Fatal(err)
		}
	}
	components := formatSystemComponents([]optimization.SystemComponent{{Name: "a", Type: "prompt", Content: "content"}})
	if !strings.Contains(components, "a (prompt): content") {
		t.Fatalf("components = %q", components)
	}
	deps := formatSystemDependencies([]optimization.ComponentDependency{{From: "a", To: "b", Type: "data"}})
	if !strings.Contains(deps, "a -> b (data)") {
		t.Fatalf("deps = %q", deps)
	}
}
