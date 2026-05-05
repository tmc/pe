package optimization

import (
	"context"
	"errors"
	"testing"
)

type strategyTestOptimizer struct {
	calls  int
	scores []float64
	err    error
}

func (o *strategyTestOptimizer) Name() string { return "test" }

func (o *strategyTestOptimizer) Optimize(ctx context.Context, config OptimizationConfig) (*OptimizationResult, error) {
	o.calls++
	if o.err != nil {
		return nil, o.err
	}
	score := float64(o.calls)
	if len(o.scores) >= o.calls {
		score = o.scores[o.calls-1]
	}
	prompt := config.InitialPrompt + "-optimized"
	return &OptimizationResult{
		OriginalPrompt:  config.InitialPrompt,
		OptimizedPrompt: prompt,
		Iterations: []IterationResult{{
			Iteration: o.calls,
			Prompt:    prompt,
			Score:     score,
		}},
		ImprovementScore: score,
	}, nil
}

func (o *strategyTestOptimizer) SupportsObjective(string) bool { return true }
func (o *strategyTestOptimizer) SupportsBatch() bool           { return false }

func TestOptimizationEngineStrategies(t *testing.T) {
	engine := NewOptimizationEngine(nil)
	names := engine.ListStrategies()
	for _, want := range []string{"single", "hybrid", "multi", "comparative"} {
		if _, err := engine.GetStrategy(want); err != nil {
			t.Fatalf("missing strategy %q in %v: %v", want, names, err)
		}
	}
	if _, err := engine.GetStrategy("missing"); err == nil {
		t.Fatal("GetStrategy accepted missing strategy")
	}
}

func TestOptimizationEngineOptimizeDefault(t *testing.T) {
	engine := NewOptimizationEngine(nil)
	optimizer := &strategyTestOptimizer{}
	result, err := engine.Optimize(context.Background(), optimizer, OptimizationConfig{
		InitialPrompt: "prompt",
		Iterations:    1,
	})
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}
	if optimizer.calls != 1 {
		t.Fatalf("optimizer calls = %d, want 1", optimizer.calls)
	}
	if result.OptimizedPrompt != "prompt-optimized" {
		t.Fatalf("OptimizedPrompt = %q", result.OptimizedPrompt)
	}
}

func TestHybridStrategy(t *testing.T) {
	optimizer := &strategyTestOptimizer{scores: []float64{0.5, 0.9}}
	result, err := (&HybridStrategy{}).Execute(context.Background(), optimizer, OptimizationConfig{
		InitialPrompt: "prompt",
		Iterations:    4,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if optimizer.calls != 2 {
		t.Fatalf("optimizer calls = %d, want 2", optimizer.calls)
	}
	if result.Method != "hybrid" {
		t.Fatalf("Method = %q, want hybrid", result.Method)
	}
	if result.ImprovementScore != 0.9 {
		t.Fatalf("ImprovementScore = %v, want 0.9", result.ImprovementScore)
	}
}

func TestHybridStrategyReturnsPhaseOneOnPhaseTwoError(t *testing.T) {
	optimizer := &strategyTestOptimizer{scores: []float64{0.5}}
	calls := 0
	optimizerFunc := OptimizerFunc(func(ctx context.Context, config OptimizationConfig) (*OptimizationResult, error) {
		calls++
		if calls == 2 {
			return nil, errors.New("phase two")
		}
		return optimizer.Optimize(ctx, config)
	})

	result, err := (&HybridStrategy{}).Execute(context.Background(), optimizerFunc, OptimizationConfig{
		InitialPrompt: "prompt",
		Iterations:    2,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if result.ImprovementScore != 0.5 {
		t.Fatalf("ImprovementScore = %v, want 0.5", result.ImprovementScore)
	}
}

func TestComparativeAndIterativeStrategies(t *testing.T) {
	optimizer := &strategyTestOptimizer{scores: []float64{0.5, 0.7, 0.7001}}
	comparative, err := (&ComparativeStrategy{}).Execute(context.Background(), optimizer, OptimizationConfig{
		InitialPrompt: "prompt",
		Iterations:    2,
	})
	if err != nil {
		t.Fatalf("comparative Execute: %v", err)
	}
	if len(comparative.Iterations) != 2 {
		t.Fatalf("comparative iterations = %d, want 2", len(comparative.Iterations))
	}

	optimizer = &strategyTestOptimizer{scores: []float64{0.5, 0.7, 0.7001}}
	iterative := NewIterativeImprovementStrategy(nil)
	result, err := iterative.Execute(context.Background(), optimizer, OptimizationConfig{
		InitialPrompt:        "prompt",
		MaxIterations:        3,
		ConvergenceThreshold: 0.01,
	})
	if err != nil {
		t.Fatalf("iterative Execute: %v", err)
	}
	if !result.Converged {
		t.Fatal("iterative result did not converge")
	}
}

type OptimizerFunc func(context.Context, OptimizationConfig) (*OptimizationResult, error)

func (f OptimizerFunc) Name() string { return "func" }
func (f OptimizerFunc) Optimize(ctx context.Context, config OptimizationConfig) (*OptimizationResult, error) {
	return f(ctx, config)
}
func (f OptimizerFunc) SupportsObjective(string) bool { return true }
func (f OptimizerFunc) SupportsBatch() bool           { return false }
