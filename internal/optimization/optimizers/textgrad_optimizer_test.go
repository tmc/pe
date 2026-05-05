package optimizers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/quick"

	"github.com/tmc/pe/internal/optimization"
)

type testProvider struct {
	response string
}

func (p testProvider) Name() string  { return "test" }
func (p testProvider) Model() string { return "test-model" }
func (p testProvider) Generate(ctx context.Context, prompt string, options optimization.GenerationOptions) (*optimization.GenerationResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	text := p.response
	if strings.Contains(prompt, "Improve this prompt") {
		text = "improved prompt with specific examples"
	}
	if text == "" {
		text = `{"gradients":[{"component":"body","feedback":"too vague","suggestions":["be specific"],"confidence":0.8,"priority":0.9,"direction":"improve","magnitude":0.7}]}`
	}
	return &optimization.GenerationResponse{Text: text}, nil
}
func (p testProvider) SupportsStreaming() bool { return false }
func (p testProvider) SupportsBatch() bool     { return false }

type testEvaluator struct{}

func (testEvaluator) EvaluatePrompt(ctx context.Context, prompt string, objective string) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return 0.8, nil
}
func (testEvaluator) EvaluateComparative(ctx context.Context, prompt1, prompt2 string, objective string) (optimization.ComparisonResult, error) {
	if err := ctx.Err(); err != nil {
		return optimization.ComparisonResult{}, err
	}
	return optimization.ComparisonResult{Winner: 2, Score1: 0.4, Score2: 0.8}, nil
}
func (testEvaluator) EvaluateSystem(ctx context.Context, system optimization.SystemDefinition, objective string) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return 0.8, nil
}

func TestTextGradOptimizerUsesInterfaces(t *testing.T) {
	optimizer := NewTextGradOptimizer(testProvider{}, testEvaluator{})
	var _ optimization.GradientOptimizer = optimizer

	result, err := optimizer.Optimize(context.Background(), optimization.OptimizationConfig{
		InitialPrompt: "summarize",
		Objective:     "clarity",
		Iterations:    1,
	})
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}
	if result.OptimizedPrompt == "summarize" {
		t.Fatalf("OptimizedPrompt = %q, want changed prompt", result.OptimizedPrompt)
	}
}

func TestTextGradOptimizerHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	optimizer := NewTextGradOptimizer(testProvider{}, testEvaluator{})
	_, err := optimizer.Optimize(ctx, optimization.OptimizationConfig{
		InitialPrompt: "summarize",
		Objective:     "clarity",
		Iterations:    1,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Optimize error = %v, want context.Canceled", err)
	}
}

func TestTextGradParseGradientsPropertyClampsScores(t *testing.T) {
	optimizer := &TextGradOptimizer{}
	property := func(confidence, priority, magnitude int16) bool {
		response := `{"gradients":[{"component":"body","feedback":"vague","suggestions":["clarify"],"confidence":` +
			floatLiteral(confidence) + `,"priority":` + floatLiteral(priority) + `,"direction":"improve","magnitude":` + floatLiteral(magnitude) + `}]}`
		gradients, err := optimizer.parseGradients(response)
		if err != nil || len(gradients) != 1 {
			return false
		}
		g := gradients[0]
		return between01(g.Confidence) && between01(g.Priority) && between01(g.Magnitude)
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func floatLiteral(v int16) string {
	return fmt.Sprintf("%g", float64(v)/10)
}

func between01(v float64) bool {
	return v >= 0 && v <= 1
}
