package metaprompt

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

type multistageProvider struct {
	err error
}

func (p multistageProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.err != nil {
		return nil, p.err
	}
	text := "advice: keep it concise"
	if strings.Contains(prompt, "VALIDATION TASKS") {
		text = `TEST CASES:
1. input
PREDICTED PERFORMANCE:
1. 8: good
FAILURE MODES:
- vague input
OVERALL ROBUSTNESS: 8
RECOMMENDATIONS:
- Add examples
- Clarify output`
	} else if strings.Contains(prompt, "final recommendations") || strings.Contains(prompt, "FORMAT:") {
		text = "1. Use examples\n2. Monitor failures\n3. Keep policies current"
	}
	return &llm.GenerateResponse{Text: text}, nil
}

func TestMultiStageValidationFlow(t *testing.T) {
	optimizer := NewMultiStageOptimizer(multistageProvider{})
	stages := []OptimizationStage{{
		Name:        "Validation",
		Method:      "validation",
		Temperature: 0.1,
		Criteria: StageCriteria{
			MinScore:        7,
			MaxIterations:   1,
			RequiredMetrics: []string{"robustness_score"},
		},
	}}
	result, err := optimizer.OptimizeMultiStage(context.Background(), "Answer carefully", stages)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success || result.FinalPrompt != "Answer carefully" || result.OverallScore != 8 {
		t.Fatalf("result = %#v", result)
	}
	if len(result.Stages) != 1 || !result.Stages[0].PassedCriteria || len(result.Stages[0].Iterations[0].Suggestions) != 2 {
		t.Fatalf("stage = %#v", result.Stages)
	}
	if len(result.Recommendations) != 3 {
		t.Fatalf("recommendations = %#v", result.Recommendations)
	}
}

func TestMultiStageFailureAndAllowFailure(t *testing.T) {
	optimizer := NewMultiStageOptimizer(multistageProvider{})
	stages := []OptimizationStage{{
		Name:     "LowValidation",
		Method:   "validation",
		Criteria: StageCriteria{MinScore: 9, MaxIterations: 1, RequiredMetrics: []string{"robustness_score"}, GatingFunction: "allow_failure"},
	}, {
		Name:   "Unknown",
		Method: "nope",
	}}
	result, err := optimizer.OptimizeMultiStage(context.Background(), "prompt", stages)
	if err != nil {
		t.Fatal(err)
	}
	if result.Success || !strings.Contains(result.FailureReason, "Unknown") || len(result.Stages) != 1 {
		t.Fatalf("result = %#v", result)
	}

	_, err = optimizer.executeStage(context.Background(), OptimizationStage{Name: "bad", Method: "missing"}, "prompt", 0)
	if err == nil || !strings.Contains(err.Error(), "unknown stage method") {
		t.Fatalf("unknown stage error = %v", err)
	}
}

func TestMultiStageCriteriaAndHelpers(t *testing.T) {
	optimizer := NewMultiStageOptimizer(multistageProvider{})
	stageResult := &StageResult{
		StageScore: 8,
		Iterations: []IterationResult{{Iteration: 1}},
		Metrics: map[string]float64{
			"gradient_strength": 0.7,
			"coherence_avg":     0.8,
			"robustness_score":  8,
		},
	}
	criteria := StageCriteria{MinScore: 7, MaxIterations: 2, RequiredMetrics: []string{"gradient_strength", "coherence_avg", "robustness_score"}}
	if !optimizer.evaluateStageCriteria(criteria, stageResult) {
		t.Fatal("criteria should pass")
	}
	if optimizer.evaluateStageCriteria(StageCriteria{MinScore: 9}, stageResult) {
		t.Fatal("low score criteria passed")
	}
	if optimizer.evaluateStageCriteria(StageCriteria{MinScore: 1, MaxIterations: 0}, stageResult) {
		t.Fatal("max iteration criteria passed")
	}
	if optimizer.evaluateStageCriteria(StageCriteria{MinScore: 1, MaxIterations: 2, RequiredMetrics: []string{"missing"}}, stageResult) {
		t.Fatal("missing metric criteria passed")
	}
	if score := optimizer.calculateOverallScore([]StageResult{{StageScore: 4}, {StageScore: 10}}); score <= 7 {
		t.Fatalf("weighted score = %v", score)
	}
	if optimizer.calculateOverallScore(nil) != 0 {
		t.Fatal("empty score not zero")
	}
	if score := optimizer.parseValidationScore("OVERALL ROBUSTNESS: 9/10"); score != 9 {
		t.Fatalf("score = %v", score)
	}
	if score := optimizer.parseValidationScore("none"); score != 5 {
		t.Fatalf("default score = %v", score)
	}
	if got := optimizer.parseRecommendations("1. First\n2. Second\nnope"); len(got) != 2 {
		t.Fatalf("recommendations = %#v", got)
	}
	if metrics := optimizer.formatMetrics(map[string]float64{"a": 1.25}); !strings.Contains(metrics, "a") {
		t.Fatalf("metrics = %q", metrics)
	}
	if summary := optimizer.formatStagesSummary([]StageResult{{Stage: OptimizationStage{Name: "s"}, StageScore: 1, PassedCriteria: true}}); !strings.Contains(summary, "s") {
		t.Fatalf("summary = %q", summary)
	}
}

func TestMultiStageProviderErrorsAndDefaults(t *testing.T) {
	boom := errors.New("boom")
	optimizer := NewMultiStageOptimizer(multistageProvider{err: boom})
	_, err := optimizer.executeStage(context.Background(), OptimizationStage{Name: "Validation", Method: "validation"}, "prompt", 0)
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("validation error = %v", err)
	}
	if _, err := optimizer.generateNextStageAdvice(context.Background(), OptimizationStage{Name: "s"}, &StageResult{}, 0); !errors.Is(err, boom) {
		t.Fatalf("advice error = %v", err)
	}
	if _, err := optimizer.generateFinalRecommendations(context.Background(), &MultiStageResult{}); !errors.Is(err, boom) {
		t.Fatalf("recommendation error = %v", err)
	}
	stages := GetDefaultStages()
	if len(stages) != 4 || stages[0].Method != "analysis" || stages[3].Criteria.GatingFunction != "allow_failure" {
		t.Fatalf("default stages = %#v", stages)
	}
}
