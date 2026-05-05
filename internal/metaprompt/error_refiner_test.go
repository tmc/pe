package metaprompt

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

type refinerProvider struct{ err error }

func (p refinerProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.err != nil {
		return nil, p.err
	}
	text := "1. Monitor edge cases\n2. Keep tests current\n3. Review failures"
	switch {
	case strings.Contains(prompt, "failure analyst"):
		text = `{"error_patterns":[{"pattern_id":"ambiguous","error_type":"semantic","description":"ambiguous output","examples":["bad"],"frequency":0.7,"severity":"high","root_cause":"missing format","triggers":["edge"]}]}`
	case strings.Contains(prompt, "Generate specific fix suggestions"):
		text = `{"fix_suggestions":[{"fix_id":"fix_format","target_pattern":"ambiguous","fix_type":"addition","description":"add format","implementation":"Use exactly three bullets.","confidence":0.9,"side_effects":["less flexible"],"test_cases":["edge"]}]}`
	case strings.Contains(prompt, "Apply these fixes"):
		text = "Answer clearly. Use exactly three bullets."
	case strings.Contains(prompt, "creating regression tests"):
		text = `{"regression_tests":[{"test_id":"format","test_type":"positive","input":"x","expected_type":"contains","expected":"bullets","rationale":"format fixed"}]}`
	}
	return &llm.GenerateResponse{Text: text}, nil
}

func TestErrorDrivenRefinerRefinePrompt(t *testing.T) {
	refiner := NewErrorDrivenRefiner(refinerProvider{})
	result, err := refiner.RefinePrompt(context.Background(), "Answer the question", []string{"output was vague"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OriginalPrompt == "" || !strings.Contains(result.RefinedPrompt, "three bullets") {
		t.Fatalf("result = %#v", result)
	}
	if len(result.ErrorPatterns) != 1 || len(result.AppliedFixes) != 1 || len(result.RegressionTests) != 1 {
		t.Fatalf("analysis slices = %#v", result)
	}
	if result.ValidationResults.FixesSuccessful != 1 || result.ValidationResults.RegressionsPassed != 1 {
		t.Fatalf("validation = %#v", result.ValidationResults)
	}
	if result.ImprovementScore <= 0 || len(result.Recommendations) != 3 {
		t.Fatalf("score/recommendations = %v %#v", result.ImprovementScore, result.Recommendations)
	}
}

func TestErrorDrivenRefinerFallbackParsingAndHelpers(t *testing.T) {
	refiner := NewErrorDrivenRefiner(refinerProvider{})
	patterns, err := refiner.detectErrorPatterns(context.Background(), "prompt", nil)
	if err != nil || len(patterns) != 1 {
		t.Fatalf("patterns = %#v err=%v", patterns, err)
	}
	fallbackPatterns := refiner.parseErrorPatterns("not json")
	if len(fallbackPatterns) != 1 || fallbackPatterns[0].PatternID == "" {
		t.Fatalf("fallback patterns = %#v", fallbackPatterns)
	}
	fixes := refiner.parseFixSuggestions("not json")
	if len(fixes) != 1 || fixes[0].FixID == "" {
		t.Fatalf("fallback fixes = %#v", fixes)
	}
	tests := refiner.parseRegressionTests("not json")
	if len(tests) != 1 || tests[0].TestID == "" {
		t.Fatalf("fallback tests = %#v", tests)
	}
	if got := refiner.formatErrorExamples(nil); !strings.Contains(got, "No specific") {
		t.Fatalf("empty examples = %q", got)
	}
	if got := refiner.formatErrorExamples([]string{"a", "b", "c", "d", "e", "f"}); strings.Contains(got, "6.") {
		t.Fatalf("too many examples = %q", got)
	}
	if got := refiner.formatErrorPatterns(patterns); !strings.Contains(got, patterns[0].PatternID) {
		t.Fatalf("formatted patterns = %q", got)
	}
	if got := refiner.formatErrorPatternsForTests(patterns); !strings.Contains(got, "Triggers") {
		t.Fatalf("formatted tests = %q", got)
	}
	if got := refiner.formatFixSuggestions([]FixSuggestion{{Description: "desc", FixType: "addition", Implementation: "impl", Confidence: 0.8}}); !strings.Contains(got, "impl") {
		t.Fatalf("formatted fixes = %q", got)
	}
}

func TestErrorDrivenRefinerSelectionValidationAndScores(t *testing.T) {
	refiner := NewErrorDrivenRefiner(refinerProvider{})
	fixes := []FixSuggestion{{FixID: "low", Confidence: 0.2}, {FixID: "high", Confidence: 0.9}, {FixID: "mid", Confidence: 0.8}}
	selected := refiner.selectBestFixes(fixes)
	if len(selected) != 2 {
		t.Fatalf("selected = %#v", selected)
	}
	selected = refiner.selectBestFixes([]FixSuggestion{{FixID: "a", Confidence: 0.1}, {FixID: "b", Confidence: 0.2}})
	if len(selected) != 1 || selected[0].FixID != "b" {
		t.Fatalf("best selected = %#v", selected)
	}
	if !refiner.validateSingleFix(context.Background(), "Use exactly three bullets.", FixSuggestion{FixType: "addition", Implementation: "Use exactly three bullets."}) {
		t.Fatal("addition fix should validate")
	}
	if refiner.validateSingleFix(context.Background(), "missing", FixSuggestion{FixType: "addition", Implementation: "required"}) {
		t.Fatal("missing addition validated")
	}
	if !refiner.runRegressionTest(context.Background(), "prompt", RegressionTest{}) {
		t.Fatal("regression placeholder should pass")
	}
	result := &RefinerResult{
		ErrorPatterns:     []ErrorPattern{{Severity: "critical"}, {Severity: "high"}, {Severity: "medium"}, {Severity: "low"}},
		AppliedFixes:      []FixSuggestion{{}, {}},
		RegressionTests:   []RegressionTest{{}, {}},
		ValidationResults: ValidationResults{FixesValidated: 2, FixesSuccessful: 1, RegressionsTotal: 2, RegressionsPassed: 1},
	}
	result.QualityMetrics = refiner.calculateQualityMetrics(result)
	if result.QualityMetrics.ErrorReduction <= 0 || result.QualityMetrics.RobustnessIncrease <= 0 {
		t.Fatalf("metrics = %#v", result.QualityMetrics)
	}
	if score := refiner.calculateImprovementScore(result); score <= 0 {
		t.Fatalf("score = %v", score)
	}
	if rate := refiner.calculateValidationSuccessRate(result.ValidationResults); rate != 0.5 {
		t.Fatalf("rate = %v", rate)
	}
	if rate := refiner.calculateValidationSuccessRate(ValidationResults{}); rate != 1 {
		t.Fatalf("empty rate = %v", rate)
	}
	if recs := refiner.parseRecommendations("1. A\n2. B\ntext"); len(recs) != 2 {
		t.Fatalf("recommendations = %#v", recs)
	}
}

func TestErrorDrivenRefinerErrors(t *testing.T) {
	boom := errors.New("boom")
	refiner := NewErrorDrivenRefiner(refinerProvider{err: boom})
	if _, err := refiner.RefinePrompt(context.Background(), "prompt", nil); !errors.Is(err, boom) {
		t.Fatalf("refine error = %v", err)
	}
	if _, err := refiner.generateFixSuggestions(context.Background(), "prompt", nil); !errors.Is(err, boom) {
		t.Fatalf("fix error = %v", err)
	}
	if _, _, err := refiner.applyFixes(context.Background(), "prompt", nil); !errors.Is(err, boom) {
		t.Fatalf("apply error = %v", err)
	}
	if _, err := refiner.generateRegressionTests(context.Background(), "a", "b", nil, nil); !errors.Is(err, boom) {
		t.Fatalf("test error = %v", err)
	}
	if _, err := refiner.generateRecommendations(context.Background(), &RefinerResult{}); !errors.Is(err, boom) {
		t.Fatalf("recommendation error = %v", err)
	}
}
