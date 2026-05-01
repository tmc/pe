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
