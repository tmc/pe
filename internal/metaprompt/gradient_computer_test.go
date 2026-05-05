package metaprompt

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

type gradientProvider struct {
	err     error
	badJSON bool
}

func (p gradientProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.err != nil {
		return nil, p.err
	}
	if p.badJSON {
		return &llm.GenerateResponse{Text: "not json"}, nil
	}

	text := "1. Tighten instructions\n2. Add examples\n3. Keep measuring"
	switch {
	case strings.Contains(prompt, "Calculate loss functions"):
		text = `{"loss_functions":[{"type":"semantic","weight":0.5,"value":0.2,"gradient":-0.1,"description":"meaning"},{"type":"structural","weight":0.5,"value":0.4,"gradient":-0.2,"description":"shape"}]}`
	case strings.Contains(prompt, "compute the optimal steps"):
		text = `{"optimization_steps":[{"step_size":0.3,"direction":"modify","component":"format","modification":"add output schema","expected_gain":0.2,"confidence":0.8}]}`
	}
	return &llm.GenerateResponse{Text: text}, nil
}

func TestGradientComputerComputeGradients(t *testing.T) {
	gc := NewGradientComputer(gradientProvider{})
	history := []IterationResult{
		{Iteration: 1, Score: 2, Feedback: "too vague", Suggestions: []string{"add format"}},
		{Iteration: 2, Score: 5, Feedback: "better", Suggestions: []string{"add examples"}},
		{Iteration: 3, Score: 8, Feedback: strings.Repeat("clear ", 30), Suggestions: []string{"trim"}},
	}

	result, err := gc.ComputeGradients(context.Background(), "Answer questions", "be precise", history)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.LossFunctions) != 2 || result.LossFunctions[0].Type != "semantic" {
		t.Fatalf("loss functions = %#v", result.LossFunctions)
	}
	if len(result.OptimizationSteps) != 1 || result.OptimizationSteps[0].Component != "format" {
		t.Fatalf("optimization steps = %#v", result.OptimizationSteps)
	}
	if len(result.GradientBuffer.Gradients) != 3 {
		t.Fatalf("gradient buffer = %#v", result.GradientBuffer)
	}
	if math.Abs(result.ConvergenceMetrics.CurrentLoss-0.3) > 1e-9 {
		t.Fatalf("current loss = %v", result.ConvergenceMetrics.CurrentLoss)
	}
	if math.Abs(result.ConvergenceMetrics.LossReduction-0.6) > 1e-9 {
		t.Fatalf("loss reduction = %v", result.ConvergenceMetrics.LossReduction)
	}
	if len(result.Recommendations) != 3 {
		t.Fatalf("recommendations = %#v", result.Recommendations)
	}
}

func TestGradientComputerFallbacksAndHelpers(t *testing.T) {
	gc := NewGradientComputer(gradientProvider{badJSON: true})
	losses, err := gc.calculateLossFunctions(context.Background(), "prompt", "objective", nil)
	if err != nil || len(losses) != 3 {
		t.Fatalf("fallback losses = %#v err=%v", losses, err)
	}
	steps, err := gc.computeOptimizationSteps(context.Background(), "prompt", losses)
	if err != nil || len(steps) != 1 {
		t.Fatalf("fallback steps = %#v err=%v", steps, err)
	}
	if got := gc.formatHistory(nil); !strings.Contains(got, "No optimization history") {
		t.Fatalf("empty history = %q", got)
	}
	history := []IterationResult{
		{Iteration: 1, Score: 1, Feedback: "one"},
		{Iteration: 2, Score: 2, Feedback: "two"},
		{Iteration: 3, Score: 3, Feedback: strings.Repeat("x", 120)},
		{Iteration: 4, Score: 4, Feedback: "four"},
	}
	if got := gc.formatHistory(history); strings.Contains(got, "Iteration 4") || !strings.Contains(got, "...") {
		t.Fatalf("limited history = %q", got)
	}
	if got := gc.formatLossFunctions(losses); !strings.Contains(got, "semantic") {
		t.Fatalf("loss format = %q", got)
	}
	if got := gc.formatOptimizationSteps([]OptimizationStep{{Modification: "a"}, {Modification: "b"}, {Modification: "c"}, {Modification: "d"}}); strings.Contains(got, "Step 4") {
		t.Fatalf("step format = %q", got)
	}
	for _, tt := range []struct {
		score float64
		want  string
	}{
		{8, "positive"},
		{2, "negative"},
		{5, "neutral"},
	} {
		if got := gc.determineDirection(tt.score); got != tt.want {
			t.Fatalf("direction %.1f = %q, want %q", tt.score, got, tt.want)
		}
	}
	if got := gc.parseRecommendations("intro\n1. first\n2. second\nx\n5. fifth"); len(got) != 3 {
		t.Fatalf("recommendations = %#v", got)
	}
}

func TestGradientComputerConvergenceHelpers(t *testing.T) {
	gc := NewGradientComputer(gradientProvider{})
	history := []IterationResult{
		{Iteration: 1, Score: 2, Feedback: "low"},
		{Iteration: 2, Score: 4, Feedback: "mid"},
		{Iteration: 3, Score: 6, Feedback: "high"},
		{Iteration: 4, Score: 8, Feedback: "higher"},
	}
	buffer := gc.updateGradientBuffer(history)
	if len(buffer.Gradients) != 4 || buffer.Gradients[0].Direction != "negative" || buffer.Gradients[3].Direction != "positive" {
		t.Fatalf("buffer = %#v", buffer)
	}
	metrics := gc.calculateConvergenceMetrics([]LossFunction{{Weight: 0.5, Value: 0.2}}, history, buffer)
	if math.Abs(metrics.GradientNorm-math.Sqrt(1.2)) > 1e-9 {
		t.Fatalf("gradient norm = %v", metrics.GradientNorm)
	}
	if metrics.IterationsLeft != 2 {
		t.Fatalf("iterations left = %d", metrics.IterationsLeft)
	}
	if got := gc.calculateStability(nil); got != 0.5 {
		t.Fatalf("default stability = %v", got)
	}
	if got := gc.calculateStability([]IterationResult{{Score: 1}, {Score: 1}, {Score: 1}}); got != 0.5 {
		t.Fatalf("flat stability = %v", got)
	}
	for _, tt := range []struct {
		loss, norm, stability float64
		want                  int
	}{
		{0.05, 0.01, 0.5, 1},
		{0.5, 0.4, 0.9, 2},
		{0.5, 0.1, 0.2, 6},
		{0.5, 0.1, 0.5, 3},
	} {
		if got := gc.estimateIterationsToConvergence(tt.loss, tt.norm, tt.stability); got != tt.want {
			t.Fatalf("estimate(%v,%v,%v) = %d, want %d", tt.loss, tt.norm, tt.stability, got, tt.want)
		}
	}
	if got := gc.calculateMean(nil); got != 0 {
		t.Fatalf("empty mean = %v", got)
	}
	if got := gc.calculateVariance(nil, 1); got != 0 {
		t.Fatalf("empty variance = %v", got)
	}
}

func TestGradientComputerProviderErrors(t *testing.T) {
	want := errors.New("provider failed")
	gc := NewGradientComputer(gradientProvider{err: want})
	if _, err := gc.calculateLossFunctions(context.Background(), "prompt", "objective", nil); !errors.Is(err, want) {
		t.Fatalf("loss error = %v", err)
	}
	if _, err := gc.computeOptimizationSteps(context.Background(), "prompt", nil); !errors.Is(err, want) {
		t.Fatalf("step error = %v", err)
	}
	if _, err := gc.generateRecommendations(context.Background(), &ComputationResult{}); !errors.Is(err, want) {
		t.Fatalf("recommendation error = %v", err)
	}
	if _, err := gc.ComputeGradients(context.Background(), "prompt", "objective", nil); err == nil || !strings.Contains(err.Error(), "failed to calculate loss functions") {
		t.Fatalf("compute error = %v", err)
	}
}
