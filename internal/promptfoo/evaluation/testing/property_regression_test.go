package testing

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

type packageProvider struct {
	text string
	err  error
}

func (p packageProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return &promptfoo.ProviderResponse{Output: p.text}, p.err
}

func (p packageProvider) Name() string { return "package" }

func (p packageProvider) Model() string { return "test" }

func (p packageProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &llm.GenerateResponse{
		Text:             p.text,
		CompletionTokens: len(strings.Fields(p.text)),
		Latency:          5 * time.Millisecond,
	}, nil
}

func (p packageProvider) SupportsStreaming() bool { return false }

func (p packageProvider) SupportsBatch() bool { return false }

func TestPropertyTesterRunAndHelpers(t *testing.T) {
	pt := NewPropertyTester(packageProvider{text: "answer contains needle"}, llm.GenerateOptions{})
	result, err := pt.RunPropertyTest(context.Background(), PropertyTest{
		Name:       "contains",
		Property:   "needle appears",
		Generator:  "sentences",
		Constraint: `response.contains("needle")`,
		Iterations: 2,
		Options:    map[string]interface{}{"count": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || len(result.Failures) != 0 || result.AverageLatency <= 0 {
		t.Fatalf("result = %#v", result)
	}

	for _, tt := range []struct {
		generator string
		options   map[string]interface{}
		contains  string
	}{
		{"random_text", map[string]interface{}{"min_length": 1, "max_length": 1}, " "},
		{"random_question", nil, "?"},
		{"random_code", map[string]interface{}{"language": "go"}, "go code"},
		{"random_json", map[string]interface{}{"depth": 1}, "{"},
		{"numbers", map[string]interface{}{"min": 7, "max": 7, "count": 2}, "7, 7"},
	} {
		got, err := pt.generateInput(tt.generator, tt.options)
		if err != nil {
			t.Fatalf("%s error = %v", tt.generator, err)
		}
		if tt.contains != " " && !strings.Contains(got, tt.contains) {
			t.Fatalf("%s = %q", tt.generator, got)
		}
	}
	if _, err := pt.generateInput("missing", nil); err == nil {
		t.Fatal("unknown generator succeeded")
	}
	if got := getIntOption(map[string]interface{}{"n": "12"}, "n", 1); got != 12 {
		t.Fatalf("int option = %d", got)
	}
	if got := getStringOption(map[string]interface{}{"s": "ok"}, "s", "x"); got != "ok" {
		t.Fatalf("string option = %q", got)
	}
}

func TestPropertyTesterFailuresAndConstraints(t *testing.T) {
	pt := NewPropertyTester(packageProvider{text: "short"}, llm.GenerateOptions{})
	for _, tt := range []struct {
		constraint string
		want       bool
	}{
		{"response.length > 3", true},
		{"response.length < 3", false},
		{`response.contains("SHORT")`, true},
		{"response.tokens < 2", true},
		{"response.latency < 10", true},
		{"response.length > prompt.length * 0.5", false},
		{"unparsed constraint", true},
	} {
		resp := &llm.GenerateResponse{CompletionTokens: 1, Latency: 5 * time.Millisecond}
		if got := pt.checkConstraint(tt.constraint, "prompt text", "short", resp); got != tt.want {
			t.Fatalf("%q = %v, want %v", tt.constraint, got, tt.want)
		}
	}

	result, err := pt.RunPropertyTest(context.Background(), PropertyTest{
		Name:       "too short",
		Generator:  "numbers",
		Constraint: "response.length > 100",
		Iterations: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || len(result.Failures) != 1 {
		t.Fatalf("constraint result = %#v", result)
	}

	want := errors.New("boom")
	result, err = NewPropertyTester(packageProvider{err: want}, llm.GenerateOptions{}).RunPropertyTest(context.Background(), PropertyTest{
		Name:       "provider error",
		Generator:  "numbers",
		Constraint: "anything",
		Iterations: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || !strings.Contains(result.Failures[0].Reason, want.Error()) {
		t.Fatalf("provider error result = %#v", result)
	}
}

func TestRegressionTesterRunAndHelpers(t *testing.T) {
	rt := NewRegressionTester(packageProvider{}, llm.GenerateOptions{})
	baseline := BaselineData{
		Provider: "p",
		Model:    "m",
		TestCases: map[string]TestCaseResult{
			"latency": {Latency: 100 * time.Millisecond, Score: 0.8, Cost: 0.01, Tokens: TokenMetrics{Total: 10}, Success: true},
			"score":   {Latency: 100 * time.Millisecond, Score: 0.8, Cost: 0.01, Tokens: TokenMetrics{Total: 10}, Success: true},
			"cost":    {Latency: 100 * time.Millisecond, Score: 0.8, Cost: 0.01, Tokens: TokenMetrics{Total: 10}, Success: true},
		},
		Summary: SummaryMetrics{SuccessRate: 1, AverageLatency: 0.1, AverageScore: 0.8, TotalCost: 0.03, TotalTokens: 30, TestCount: 3},
	}
	current := baseline
	current.Summary.AverageScore = 0.7
	current.TestCases = map[string]TestCaseResult{
		"latency": {Latency: 200 * time.Millisecond, Score: 0.8, Cost: 0.01, Tokens: TokenMetrics{Total: 10}, Success: true},
		"score":   {Latency: 100 * time.Millisecond, Score: 0.4, Cost: 0.01, Tokens: TokenMetrics{Total: 10}, Success: true},
		"cost":    {Latency: 100 * time.Millisecond, Score: 1.0, Cost: 0.02, Tokens: TokenMetrics{Total: 10}, Success: true},
	}

	file := t.TempDir() + "/baseline.json"
	if err := rt.SaveBaseline(file, baseline); err != nil {
		t.Fatal(err)
	}
	result, err := rt.RunRegressionTest(context.Background(), RegressionTest{
		Name:          "regression",
		Baseline:      file,
		Tolerance:     10,
		ThresholdType: "relative",
		Metrics:       []string{"success_rate", "average_score", "total_tokens", "unknown"},
	}, current)
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || len(result.Regressions) != 3 || len(result.Improvements) == 0 {
		t.Fatalf("regression result = %#v", result)
	}
	if result.Metrics["unknown"].Baseline != 0 {
		t.Fatalf("unknown metric = %#v", result.Metrics["unknown"])
	}
	if result.OverallChange <= 0 {
		t.Fatalf("overall change = %v", result.OverallChange)
	}
}

func TestRegressionTesterEdges(t *testing.T) {
	rt := NewRegressionTester(packageProvider{}, llm.GenerateOptions{})
	if _, err := rt.RunRegressionTest(context.Background(), RegressionTest{Baseline: "missing"}, BaselineData{}); err == nil {
		t.Fatal("missing baseline succeeded")
	}
	badFile := t.TempDir() + "/bad.json"
	if err := os.WriteFile(badFile, []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.loadBaseline(badFile); err == nil {
		t.Fatal("bad baseline parsed")
	}
	if err := rt.SaveBaseline(t.TempDir(), BaselineData{}); err == nil {
		t.Fatal("directory write succeeded")
	}
	for _, tt := range []struct {
		metric            string
		baseline, current float64
		want              bool
	}{
		{"latency", 1, 2, true},
		{"cost", 1, 2, true},
		{"score", 2, 1, true},
		{"other", 1, 2, false},
		{"latency", 0, 2, false},
	} {
		if got := rt.isRegression(tt.metric, tt.baseline, tt.current, 10); got != tt.want {
			t.Fatalf("isRegression(%q,%v,%v) = %v, want %v", tt.metric, tt.baseline, tt.current, got, tt.want)
		}
	}
	for _, tt := range []struct {
		baseline, current float64
		want              string
	}{
		{0, 1, "unknown"},
		{1, 1.6, "critical"},
		{1, 1.3, "high"},
		{1, 1.2, "medium"},
		{1, 1.05, "low"},
	} {
		if got := rt.calculateSeverity(tt.baseline, tt.current); got != tt.want {
			t.Fatalf("severity(%v,%v) = %q, want %q", tt.baseline, tt.current, got, tt.want)
		}
	}
	if got := rt.calculateOverallChange(nil); got != 0 {
		t.Fatalf("empty overall change = %v", got)
	}
	created := CreateBaselineFromEvaluation("p", "m", map[string]TestCaseResult{
		"a": {Latency: time.Second, Score: 1, Cost: 0.1, Tokens: TokenMetrics{Total: 3}, Success: true},
		"b": {Latency: 3 * time.Second, Score: 3, Cost: 0.2, Tokens: TokenMetrics{Total: 7}},
	})
	if created.Summary.SuccessRate != 0.5 || created.Summary.AverageLatency != 2 || created.Summary.TotalTokens != 10 {
		t.Fatalf("created baseline = %#v", created.Summary)
	}
}
