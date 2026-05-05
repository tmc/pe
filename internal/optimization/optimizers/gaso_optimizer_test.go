package optimizers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/optimization"
)

type gasoProvider struct {
	responses []string
	err       error
	calls     []string
}

func (p *gasoProvider) Name() string            { return "gaso-test" }
func (p *gasoProvider) Model() string           { return "test-model" }
func (p *gasoProvider) SupportsStreaming() bool { return false }
func (p *gasoProvider) SupportsBatch() bool     { return false }
func (p *gasoProvider) Generate(ctx context.Context, prompt string, options optimization.GenerationOptions) (*optimization.GenerationResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.err != nil {
		return nil, p.err
	}
	p.calls = append(p.calls, prompt)
	text := "optimized component content with details"
	if strings.Contains(prompt, "Response format (JSON)") {
		text = `prefix {"component":"content","feedback":"too vague","suggestions":["add examples"],"confidence":1.5,"priority":-0.2,"direction":"clarify","magnitude":2} suffix`
	}
	if len(p.responses) > 0 {
		text = p.responses[0]
		p.responses = p.responses[1:]
	}
	return &optimization.GenerationResponse{Text: text}, nil
}

type gasoEvaluator struct {
	scores []float64
	err    error
	calls  int
}

func (e *gasoEvaluator) EvaluatePrompt(ctx context.Context, prompt string, objective string) (float64, error) {
	return 0.8, ctx.Err()
}
func (e *gasoEvaluator) EvaluateComparative(ctx context.Context, prompt1, prompt2 string, objective string) (optimization.ComparisonResult, error) {
	return optimization.ComparisonResult{Winner: 2, Score1: 0.4, Score2: 0.8}, ctx.Err()
}
func (e *gasoEvaluator) EvaluateSystem(ctx context.Context, system optimization.SystemDefinition, objective string) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if e.err != nil {
		return 0, e.err
	}
	e.calls++
	if len(e.scores) == 0 {
		return 0.8, nil
	}
	score := e.scores[0]
	e.scores = e.scores[1:]
	return score, nil
}

func TestGASOOptimizerOptimizeSystem(t *testing.T) {
	provider := &gasoProvider{}
	evaluator := &gasoEvaluator{scores: []float64{0.4, 0.9}}
	optimizer := NewGASOOptimizer(provider, evaluator)
	if optimizer.Name() != "gaso" || !optimizer.SupportsBatch() || !optimizer.SupportsObjective("system clarity") {
		t.Fatalf("basic optimizer metadata failed")
	}

	result, err := optimizer.OptimizeSystem(context.Background(), gasoSystem(), optimization.OptimizationConfig{
		Objective:            "clarity",
		MaxIterations:        1,
		ConvergenceThreshold: 0.0001,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OverallPerformance != 0.9 || result.Iterations != 1 {
		t.Fatalf("result = %#v", result)
	}
	if len(result.OptimizedComponents) != 2 || !strings.Contains(result.OptimizedComponents[0].Content, "optimized") {
		t.Fatalf("components = %#v", result.OptimizedComponents)
	}
	if len(result.OptimizationHistory) != 1 || len(result.OptimizationHistory[0].Changes) == 0 {
		t.Fatalf("history = %#v", result.OptimizationHistory)
	}
	if len(provider.calls) == 0 || evaluator.calls != 2 {
		t.Fatalf("provider calls %d evaluator calls %d", len(provider.calls), evaluator.calls)
	}
}

func TestGASOOptimizerOptimizeSinglePrompt(t *testing.T) {
	provider := &gasoProvider{}
	evaluator := &gasoEvaluator{scores: []float64{0.3, 0.7}}
	optimizer := NewGASOOptimizer(provider, evaluator)
	result, err := optimizer.Optimize(context.Background(), optimization.OptimizationConfig{
		InitialPrompt: "summarize",
		Objective:     "clarity",
		Iterations:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Method != "gaso" || result.OriginalPrompt != "summarize" || result.OptimizedPrompt == "summarize" {
		t.Fatalf("result = %#v", result)
	}
}

func TestGASOOptimizerFallbackAndHelpers(t *testing.T) {
	optimizer := &GASOOptimizer{provider: &gasoProvider{responses: []string{"not json", "applied"}}, evaluator: &gasoEvaluator{}}
	gradients, err := optimizer.ComputeSystemGradients(context.Background(), gasoSystem(), "clarity")
	if err != nil {
		t.Fatal(err)
	}
	if len(gradients) != 2 || gradients[0].Gradient.Component != "content optimization" {
		t.Fatalf("fallback gradients = %#v", gradients)
	}
	if deps := optimizer.formatDependencies("missing", gasoSystem()); deps != "No direct dependencies" {
		t.Fatalf("deps = %q", deps)
	}
	priority := optimizer.calculateComponentPriority(gasoSystem().Components[0], gasoSystem())
	if priority <= 0.5 || priority > 1 {
		t.Fatalf("priority = %v", priority)
	}
	parsed, err := optimizer.parseComponentGradient(`{"component":"x","confidence":-1,"priority":2,"magnitude":2}`)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Confidence != 0 || parsed.Priority != 1 || parsed.Magnitude != 1 {
		t.Fatalf("clamped gradient = %#v", parsed)
	}
	if _, err := optimizer.parseComponentGradient("none"); err == nil {
		t.Fatal("parse without json did not error")
	}
}

func TestGASOOptimizerErrors(t *testing.T) {
	boom := errors.New("boom")
	_, err := NewGASOOptimizer(&gasoProvider{}, &gasoEvaluator{err: boom}).OptimizeSystem(context.Background(), gasoSystem(), optimization.OptimizationConfig{Iterations: 1})
	if err == nil || !strings.Contains(err.Error(), "initial system evaluation") {
		t.Fatalf("initial error = %v", err)
	}
	_, err = NewGASOOptimizer(&gasoProvider{err: boom}, &gasoEvaluator{scores: []float64{0.4}}).OptimizeSystem(context.Background(), gasoSystem(), optimization.OptimizationConfig{Iterations: 1})
	if err == nil || !strings.Contains(err.Error(), "gradient computation") {
		t.Fatalf("gradient error = %v", err)
	}
	_, err = NewGASOOptimizer(&gasoProvider{responses: []string{`{"component":"x"}`, "applied"}}, &gasoEvaluator{scores: []float64{0.4}, err: boom}).Optimize(context.Background(), optimization.OptimizationConfig{InitialPrompt: "x", Iterations: 1})
	if err == nil || !strings.Contains(err.Error(), "system optimization failed") {
		t.Fatalf("optimize error = %v", err)
	}
}

func TestGASOConvergenceAnalysis(t *testing.T) {
	optimizer := &GASOOptimizer{}
	short := optimizer.analyzeConvergence(nil)
	if short.Converged || short.FinalGradient != 1 {
		t.Fatalf("short convergence = %#v", short)
	}
	history := []optimization.SystemIterationResult{
		{Performance: 0.80, Improvement: 0.010},
		{Performance: 0.81, Improvement: 0.005},
		{Performance: 0.82, Improvement: 0.004},
	}
	analysis := optimizer.analyzeConvergence(history)
	if analysis.FinalGradient <= 0 || analysis.StabilityMetric <= 0 {
		t.Fatalf("analysis = %#v", analysis)
	}
}

func gasoSystem() optimization.SystemDefinition {
	return optimization.SystemDefinition{
		Name:        "system",
		Description: "two component system",
		Components: []optimization.SystemComponent{
			{ID: "planner", Name: "Planner", Type: "agent", Content: "plan briefly"},
			{ID: "writer", Name: "Writer", Type: "prompt", Content: "write answer"},
		},
		Dependencies: []optimization.ComponentDependency{{From: "planner", To: "writer", Type: "control", Weight: 0.8}},
	}
}
