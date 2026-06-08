package evaluator

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

type assertionJudgeTestProvider struct {
	name       string
	model      string
	response   string
	calls      int
	lastPrompt string
}

func (p *assertionJudgeTestProvider) Name() string {
	return p.name
}

func (p *assertionJudgeTestProvider) Model() string {
	return p.model
}

func (p *assertionJudgeTestProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	p.calls++
	p.lastPrompt = prompt
	return &llm.GenerateResponse{
		Text:         p.response,
		Model:        p.model,
		Latency:      time.Millisecond,
		FinishReason: "stop",
	}, nil
}

func (p *assertionJudgeTestProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	response, err := p.Generate(ctx, prompt, llm.GenerateOptions{})
	if err != nil {
		return nil, err
	}
	return &promptfoo.ProviderResponse{
		Output:    response.Text,
		LatencyMs: response.Latency.Milliseconds(),
	}, nil
}

func (p *assertionJudgeTestProvider) SupportsStreaming() bool {
	return false
}

func (p *assertionJudgeTestProvider) SupportsBatch() bool {
	return false
}

func TestAssertionUnmarshalKeepsProvider(t *testing.T) {
	var assertion Assertion
	err := json.Unmarshal([]byte(`{"type":"llm-judge","value":"be correct","provider":"openai:gpt-4o-mini"}`), &assertion)
	require.NoError(t, err)
	assert.Equal(t, AssertionLLMJudge, assertion.Type)
	assert.Equal(t, "openai:gpt-4o-mini", assertion.Provider)
}

func TestEvaluateAssertionsKeepsPromptfooAssertionProvider(t *testing.T) {
	_, grading := evaluateAssertions("output", []promptfoo.Assertion{
		{Type: "contains", Value: "output", Provider: "openai:gpt-4o-mini"},
	})

	require.Len(t, grading.ComponentResults, 1)
	assert.Equal(t, "openai:gpt-4o-mini", grading.ComponentResults[0].Assertion.Provider)
}

func TestAssertionEvaluatorCoversAllAssertionTypes(t *testing.T) {
	max := 1.0
	min := 1.0
	threshold := 0.5
	evaluator := NewAssertionEvaluator(&assertionJudgeTestProvider{
		name:     "judge",
		model:    "test",
		response: "SCORE: 9\nREASONING: ok",
	})

	tests := []struct {
		name      string
		assertion Assertion
		output    string
		metadata  map[string]interface{}
	}{
		{"contains", Assertion{Type: AssertionContains, Value: "answer"}, "answer", nil},
		{"equals", Assertion{Type: AssertionEquals, Value: "answer"}, "answer", nil},
		{"matches", Assertion{Type: AssertionMatches, Value: `ans.er`}, "answer", nil},
		{"length", Assertion{Type: AssertionLength, Min: &min}, "answer", nil},
		{"not contains", Assertion{Type: AssertionNotContains, Value: "missing"}, "answer", nil},
		{"readability", Assertion{Type: AssertionReadability, Threshold: &threshold}, "This is a short readable sentence.", nil},
		{"sentiment", Assertion{Type: AssertionSentiment, Value: "neutral"}, "fine", nil},
		{"toxicity", Assertion{Type: AssertionToxicity}, "answer", nil},
		{"coherence", Assertion{Type: AssertionCoherence}, "answer", nil},
		{"factuality", Assertion{Type: AssertionFactuality}, "answer", nil},
		{"llm judge", Assertion{Type: AssertionLLMJudge, Value: "be correct", Threshold: &threshold}, "answer", nil},
		{"classify", Assertion{Type: AssertionClassify, Value: "positive", Config: map[string]interface{}{"labels": map[string]interface{}{"positive": []interface{}{"good"}, "negative": []interface{}{"bad"}}}}, "good answer", nil},
		{"similarity", Assertion{Type: AssertionSimilarity, Value: "answer"}, "answer", nil},
		{"latency", Assertion{Type: AssertionLatency, Max: &max}, "answer", map[string]interface{}{"latency": 10 * time.Millisecond}},
		{"cost", Assertion{Type: AssertionCost, Max: &max}, "answer", map[string]interface{}{"cost": 0.01}},
		{"tokens", Assertion{Type: AssertionTokens, Max: &max}, "answer", map[string]interface{}{"tokens": 1}},
		{"json", Assertion{Type: AssertionJSON}, `{"answer":true}`, nil},
		{"sql", Assertion{Type: AssertionSQL}, "select 1", nil},
		{"code", Assertion{Type: AssertionCode, Threshold: &threshold, Config: map[string]interface{}{"language": "go"}}, "package main", nil},
		{"structure", Assertion{Type: AssertionStructure, Value: []interface{}{"answer"}}, "answer", nil},
		{"pass at n", Assertion{Type: AssertionPassAtN, Threshold: &threshold, Config: map[string]interface{}{"n": float64(2)}}, "answer", map[string]interface{}{"samples": []string{"answer", "other"}}},
		{"structured output", Assertion{Type: AssertionStructuredOutput}, `{"answer":true}`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateAssertion(context.Background(), tt.assertion, tt.output, tt.metadata)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.assertion.Type, result.Type)
		})
	}
}

func TestUnimplementedAssertionsFailClosed(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)

	tests := []struct {
		name string
		eval func() *AssertionResult
	}{
		{
			name: "coherence",
			eval: func() *AssertionResult {
				return evaluator.evaluateCoherence(context.Background(), Assertion{Type: AssertionCoherence}, "answer")
			},
		},
		{
			name: "factuality",
			eval: func() *AssertionResult {
				return evaluator.evaluateFactuality(context.Background(), Assertion{Type: AssertionFactuality}, "answer")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.eval()
			require.NotNil(t, result)
			assert.False(t, result.Passed)
			assert.Equal(t, 0.0, result.Score)
			assert.Contains(t, result.Message, "not yet implemented")
		})
	}
}

func TestAssertionEvaluatorToxicityLocalTerms(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionToxicity,
		Config: map[string]interface{}{
			"terms": []interface{}{"awful phrase", "blocked"},
		},
	}, "This response is blocked.", nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.InDelta(t, 0.5, result.Score, 0.0001)
	assert.Equal(t, "local_term_match", result.Metadata["method"])

	result, err = evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionToxicity,
		Config: map[string]interface{}{
			"terms": []interface{}{"awful phrase", "blocked"},
		},
	}, "This response is fine.", nil)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Equal(t, 0.0, result.Score)
}

func TestAssertionEvaluatorClassifyLocalKeywords(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:  AssertionClassify,
		Value: "positive",
		Config: map[string]interface{}{
			"labels": map[string]interface{}{
				"positive": []interface{}{"great", "helpful"},
				"negative": []interface{}{"broken", "bad"},
			},
		},
	}, "A great and helpful answer.", nil)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Equal(t, "positive", result.Actual)
	assert.Equal(t, "keyword_overlap", result.Metadata["method"])

	result, err = evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:  AssertionClassify,
		Value: "positive",
	}, "A great answer.", nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "requires config.labels")
}

func TestAssertionEvaluatorSimilarityLocalBaseline(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)
	threshold := 0.5
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:      AssertionSimilarity,
		Value:     "the quick brown fox",
		Threshold: &threshold,
	}, "quick brown fox jumps", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.InDelta(t, 0.6, result.Score, 0.0001)
	assert.Equal(t, "token_jaccard", result.Metadata["method"])

	result, err = evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:  AssertionSimilarity,
		Value: "",
	}, "output", nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "requires non-empty string value")
}

func TestAssertionEvaluatorSQLLocalShape(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionSQL,
	}, "select name from users where id = 1", nil)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Equal(t, 1.0, result.Score)

	result, err = evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionSQL,
	}, "select (name from users", nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "unbalanced parentheses")
}

func TestAssertionEvaluatorStructureMarkers(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)
	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:  AssertionStructure,
		Value: []interface{}{"Summary", "Details"},
	}, "Summary\n\nDetails\n\nDone", nil)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Equal(t, 1.0, result.Score)

	result, err = evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionStructure,
		Config: map[string]interface{}{
			"required": []interface{}{"Summary", "Risks"},
		},
	}, "Summary only", nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.InDelta(t, 0.5, result.Score, 0.0001)
	assert.Equal(t, []string{"Risks"}, result.Metadata["missing"])

	result, err = evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type: AssertionStructure,
	}, "anything", nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "requires required markers")
}

func TestAssertionEvaluatorLLMJudgeUsesProviderOverride(t *testing.T) {
	const providerName = "assertiontestjudge"

	var override *assertionJudgeTestProvider
	llm.RegisterProviderFactory(providerName, func(model string, options map[string]interface{}) (llm.Provider, error) {
		override = &assertionJudgeTestProvider{
			name:     providerName,
			model:    model,
			response: "SCORE: 9\nREASONING: override provider",
		}
		return override, nil
	})

	base := &assertionJudgeTestProvider{
		name:     "base",
		model:    "base",
		response: "SCORE: 0\nREASONING: base provider",
	}
	evaluator := NewAssertionEvaluator(base)
	threshold := 0.8

	result, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:      AssertionLLMJudge,
		Value:     "prefer concise answers",
		Provider:  providerName + ":strict",
		Threshold: &threshold,
	}, "a concise answer", nil)

	require.NoError(t, err)
	require.NotNil(t, override)
	assert.Equal(t, 0, base.calls)
	assert.Equal(t, 1, override.calls)
	assert.Equal(t, "strict", override.model)
	assert.Contains(t, override.lastPrompt, "CRITERIA: prefer concise answers")
	assert.True(t, result.Passed)
	assert.InDelta(t, 0.9, result.Score, 0.0001)
	assert.Equal(t, "override provider", result.Metadata["reasoning"])
}

func TestEvaluateUsesAssertionProviderOverrideForLLMJudge(t *testing.T) {
	const baseName = "assertionevalbase"
	const judgeName = "assertionevaljudge"

	base := &assertionJudgeTestProvider{
		name:     baseName,
		model:    "base",
		response: "answer",
	}
	var judge *assertionJudgeTestProvider

	llm.RegisterProviderFactory(baseName, func(model string, options map[string]interface{}) (llm.Provider, error) {
		base.model = model
		return base, nil
	})
	llm.RegisterProviderFactory(judgeName, func(model string, options map[string]interface{}) (llm.Provider, error) {
		judge = &assertionJudgeTestProvider{
			name:     judgeName,
			model:    model,
			response: "SCORE: 9\nREASONING: override provider",
		}
		return judge, nil
	})

	threshold := 0.8
	result, err := Evaluate(promptfoo.Config{
		Prompts:   []string{"question"},
		Providers: []promptfoo.ProviderConfig{{ID: baseName + ":main"}},
		Tests: []promptfoo.TestCase{
			{
				Assert: []promptfoo.Assertion{
					{
						Type:      "llm-judge",
						Value:     "be correct",
						Provider:  judgeName + ":strict",
						Threshold: threshold,
					},
				},
			},
		},
	}, 10*time.Second, false, 1, false)

	require.NoError(t, err)
	require.Len(t, result.Results.Results, 1)
	require.NotNil(t, judge)
	assert.True(t, result.Results.Results[0].Success)
	assert.Equal(t, "main", base.model)
	assert.Equal(t, 1, base.calls)
	assert.Equal(t, "strict", judge.model)
	assert.Equal(t, 1, judge.calls)
	assert.Contains(t, judge.lastPrompt, "CRITERIA: be correct")
	assert.Contains(t, result.Results.Results[0].GradingResult.ComponentResults[0].Reason, "override provider")
}

func TestAssertionEvaluatorLLMJudgeRejectsInvalidProviderOverride(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	base := &assertionJudgeTestProvider{
		name:     "base",
		model:    "base",
		response: "SCORE: 10\nREASONING: base provider",
	}
	evaluator := NewAssertionEvaluator(base)

	_, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:     AssertionLLMJudge,
		Value:    "be correct",
		Provider: "openai:gpt-4",
	}, "output", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `resolve assertion provider "openai:gpt-4"`)
}

func TestAssertionEvaluatorLLMJudgeRequiresProvider(t *testing.T) {
	evaluator := NewAssertionEvaluator(nil)

	_, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:  AssertionLLMJudge,
		Value: "be correct",
	}, "output", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "llm judge assertion requires a provider")
}

func TestAssertionEvaluatorRejectsProviderOverrideForNonJudge(t *testing.T) {
	evaluator := NewAssertionEvaluator(&assertionJudgeTestProvider{
		name:     "base",
		model:    "base",
		response: "SCORE: 10\nREASONING: base provider",
	})

	_, err := evaluator.EvaluateAssertion(context.Background(), Assertion{
		Type:     AssertionContains,
		Value:    "output",
		Provider: "openai:gpt-4",
	}, "output", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "assertion provider override is only supported for llm-judge assertions")
}
