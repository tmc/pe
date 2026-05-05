package metrics

import (
	"context"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

func TestMetricEvaluatorBuiltinsAndErrors(t *testing.T) {
	configs := []MetricConfig{
		{Name: "length", Type: MetricTypeBuiltin, Config: map[string]interface{}{"min": float64(5), "max": float64(80)}},
		{Name: "word_count", Type: MetricTypeBuiltin, Config: map[string]interface{}{"min": float64(2), "max": float64(20)}},
		{Name: "sentiment", Type: MetricTypeBuiltin, Threshold: 0.2, Config: map[string]interface{}{"target": "positive"}},
		{Name: "readability", Type: MetricTypeBuiltin, Config: map[string]interface{}{"min_grade_level": float64(0), "max_grade_level": float64(20)}},
		{Name: "toxicity", Type: MetricTypeBuiltin, Threshold: 0.2, Config: map[string]interface{}{}},
		{Name: "coherence", Type: MetricTypeBuiltin, Threshold: 0, Config: map[string]interface{}{}},
		{Name: "relevance", Type: MetricTypeBuiltin, Threshold: 0.2, Config: map[string]interface{}{}},
		{Name: "unknown", Type: MetricTypeBuiltin},
		{Name: "badtype", Type: MetricType("nope")},
	}
	evaluator := NewMetricEvaluator(configs, nil)
	results, err := evaluator.EvaluateAll(context.Background(), "Explain renewable energy benefits", "Good renewable energy is wonderful. Therefore it helps climate.", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != len(configs) {
		t.Fatalf("results = %d, want %d", len(results), len(configs))
	}
	for i := 0; i < 7; i++ {
		if results[i].Error != "" || results[i].Reason == "" || results[i].Details == nil {
			t.Fatalf("result[%d] = %#v", i, results[i])
		}
	}
	if results[7].Error == "" || !strings.Contains(results[7].Error, "unknown builtin") {
		t.Fatalf("unknown result = %#v", results[7])
	}
	if results[8].Error == "" || !strings.Contains(results[8].Error, "unsupported metric") {
		t.Fatalf("bad type result = %#v", results[8])
	}
}

func TestMetricEvaluatorRegexAndHelpers(t *testing.T) {
	evaluator := NewMetricEvaluator(nil, nil)
	regex, err := evaluator.evaluateMetric(context.Background(), MetricConfig{
		Name:    "ticket",
		Type:    MetricTypeRegex,
		Pattern: `PE-\d+`,
		Config:  map[string]interface{}{"expected_count": float64(2)},
	}, "", "refs PE-123 and PE-456", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !regex.Pass || regex.Score != 1 || regex.Details["match_count"].(int) != 2 {
		t.Fatalf("regex = %#v", regex)
	}
	if _, err := evaluator.evaluateMetric(context.Background(), MetricConfig{Name: "bad", Type: MetricTypeRegex, Pattern: "["}, "", "x", nil); err == nil {
		t.Fatal("invalid regex did not error")
	}
	if got := countSyllables("queue"); got < 1 {
		t.Fatalf("syllables = %d", got)
	}
	keywords := extractKeywords("The quick brown fox and the quick dog")
	if !keywords["quick"] || keywords["the"] {
		t.Fatalf("keywords = %#v", keywords)
	}
}

func TestMetricEvaluatorLLMAndPassAtN(t *testing.T) {
	provider := jsonJudgeProvider{}
	evaluator := NewMetricEvaluator(nil, map[string]llm.Provider{"judge": provider})
	llmResult, err := evaluator.evaluateMetric(context.Background(), MetricConfig{
		Name:   "judge-score",
		Type:   MetricTypeLLM,
		Judge:  "judge",
		Prompt: "Grade the response.",
	}, "prompt", "response", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !llmResult.Pass || llmResult.Score != 0.75 || llmResult.Details["judge"] != "judge" {
		t.Fatalf("llm result = %#v", llmResult)
	}
	if _, err := evaluator.evaluateMetric(context.Background(), MetricConfig{Name: "missing", Type: MetricTypeLLM, Judge: "missing"}, "", "", nil); err == nil {
		t.Fatal("missing judge did not error")
	}

	pass, err := evaluator.evaluateMetric(context.Background(), MetricConfig{
		Name:      "pass-at-n",
		Type:      MetricTypePassAtN,
		Threshold: 0.5,
		Config: map[string]interface{}{
			"n":        float64(2),
			"provider": "judge",
		},
	}, "", "ok code", map[string]interface{}{"samples": []string{"ok", "error", "also ok"}})
	if err != nil {
		t.Fatal(err)
	}
	if !pass.Pass || pass.Score <= 0 || pass.Details["num_samples"].(int) != 3 {
		t.Fatalf("pass@n = %#v", pass)
	}

	withTests, err := evaluator.evaluateMetric(context.Background(), MetricConfig{
		Name: "pass-at-n-tests",
		Type: MetricTypePassAtN,
		Config: map[string]interface{}{
			"n":          float64(1),
			"provider":   "judge",
			"test_cases": []interface{}{map[string]interface{}{"input": "1", "expected": "1"}},
		},
	}, "", "ok code", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !withTests.Pass || withTests.Score != 1 {
		t.Fatalf("pass@n tests = %#v", withTests)
	}
}

type jsonJudgeProvider struct{}

func (jsonJudgeProvider) Name() string            { return "judge" }
func (jsonJudgeProvider) Model() string           { return "judge-model" }
func (jsonJudgeProvider) SupportsStreaming() bool { return false }
func (jsonJudgeProvider) SupportsBatch() bool     { return false }
func (jsonJudgeProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return &promptfoo.ProviderResponse{Output: `{"score":0.75,"pass":true,"reason":"ok"}`}, nil
}
func (jsonJudgeProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	return &llm.GenerateResponse{Text: "YES", Model: "judge-model"}, nil
}
