package evaluator

import (
	"context"
	"testing"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// scriptedJudgeProvider returns a different canned response on each Generate
// call, cycling through responses; used to script per-window verdicts.
type scriptedJudgeProvider struct {
	responses []string
	calls     int
}

func (p *scriptedJudgeProvider) Name() string  { return "scripted" }
func (p *scriptedJudgeProvider) Model() string { return "scripted" }

func (p *scriptedJudgeProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	r := ""
	if len(p.responses) > 0 {
		r = p.responses[p.calls%len(p.responses)]
	}
	p.calls++
	return &llm.GenerateResponse{Text: r, Latency: time.Millisecond, FinishReason: "stop"}, nil
}

func (p *scriptedJudgeProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	resp, err := p.Generate(ctx, prompt, llm.GenerateOptions{})
	if err != nil {
		return nil, err
	}
	return &promptfoo.ProviderResponse{Output: resp.Text}, nil
}

func (p *scriptedJudgeProvider) SupportsStreaming() bool { return false }
func (p *scriptedJudgeProvider) SupportsBatch() bool     { return false }

func TestEvaluateRubric(t *testing.T) {
	for _, typ := range []AssertionType{AssertionAgentRubric, AssertionSearchRubric} {
		judge := &assertionJudgeTestProvider{response: "SCORE: 9\nREASONING: meets the rubric well"}
		ae := NewAssertionEvaluator(judge)
		a := Assertion{Type: typ, Value: "the answer must mention Paris"}
		got, err := ae.EvaluateAssertion(context.Background(), a, "The capital is Paris.", nil)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", typ, err)
		}
		if !got.Passed {
			t.Errorf("%s: expected pass at score 0.9, got %v (%s)", typ, got.Passed, got.Message)
		}
		if judge.calls != 1 {
			t.Errorf("%s: expected 1 judge call, got %d", typ, judge.calls)
		}
	}
}

func TestEvaluateRubricRequiresProvider(t *testing.T) {
	ae := NewAssertionEvaluator(nil) // no judge
	a := Assertion{Type: AssertionAgentRubric, Value: "rubric"}
	_, err := ae.EvaluateAssertion(context.Background(), a, "out", nil)
	if err == nil {
		t.Fatal("expected error when no judge provider is configured")
	}
}

func TestEvaluateRubricEmptyValue(t *testing.T) {
	judge := &assertionJudgeTestProvider{response: "SCORE: 5"}
	ae := NewAssertionEvaluator(judge)
	a := Assertion{Type: AssertionSearchRubric, Value: 42} // not a string
	got, err := ae.EvaluateAssertion(context.Background(), a, "out", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Passed {
		t.Error("expected fail when rubric value is not a string")
	}
}

func TestEvaluateConversationRelevanceSingleTurn(t *testing.T) {
	judge := &assertionJudgeTestProvider{response: `{"verdict": "yes"}`}
	ae := NewAssertionEvaluator(judge)
	a := Assertion{Type: AssertionConversationRelevance}
	meta := map[string]interface{}{"input": "What is the capital of France?"}
	got, err := ae.EvaluateAssertion(context.Background(), a, "Paris.", meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Score != 1.0 || !got.Passed {
		t.Errorf("single relevant turn: score=%v passed=%v, want 1.0/true (%s)", got.Score, got.Passed, got.Message)
	}
	if judge.calls != 1 {
		t.Errorf("expected 1 verdict call (one assistant turn), got %d", judge.calls)
	}
}

func TestEvaluateConversationRelevanceMultiTurn(t *testing.T) {
	// Two assistant turns: first relevant ("yes"), second irrelevant ("no").
	// The scripted judge alternates yes/no, so score = 1/2 = 0.5.
	judge := &scriptedJudgeProvider{responses: []string{
		`{"verdict": "yes"}`,
		`{"verdict": "no", "reason": "off topic"}`,
	}}
	ae := NewAssertionEvaluator(judge)
	a := Assertion{Type: AssertionConversationRelevance, Threshold: floatPtr(0.75)}
	meta := map[string]interface{}{
		"_conversation": []interface{}{
			map[string]interface{}{"input": "Hi", "output": "Hello, how can I help?"},
			map[string]interface{}{"input": "What meds for a sore throat?", "output": "Nice weather today!"},
		},
	}
	got, err := ae.EvaluateAssertion(context.Background(), a, "ignored when _conversation present", meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(got.Score, 0.5) {
		t.Errorf("score = %v, want 0.5 (%s)", got.Score, got.Message)
	}
	if got.Passed { // 0.5 < 0.75 threshold
		t.Errorf("expected fail at threshold 0.75, got pass (%s)", got.Message)
	}
	if judge.calls != 2 {
		t.Errorf("expected 2 verdict calls (two assistant turns), got %d", judge.calls)
	}
}

func TestModelGradedAliases(t *testing.T) {
	ids := []struct {
		id   string
		want AssertionType
	}{
		{"agent-rubric", AssertionAgentRubric},
		{"search-rubric", AssertionSearchRubric},
		{"conversation-relevance", AssertionConversationRelevance},
	}
	for _, tc := range ids {
		canonical, _, ok := normalizeAssertionType(tc.id)
		if !ok {
			t.Errorf("%q did not resolve through the alias map", tc.id)
			continue
		}
		if canonical != tc.want {
			t.Errorf("%q resolved to %q, want %q", tc.id, canonical, tc.want)
		}
		if !isModelGraded(canonical) {
			t.Errorf("%q should be model-graded (judge required)", tc.id)
		}
	}
}

func TestParseConversationVerdict(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{`{"verdict": "yes"}`, "yes"},
		{`Sure! {"verdict": "no", "reason": "x"} done`, "no"},
		{`yes`, "yes"},
		{`garbage`, ""},
	}
	for _, tc := range tests {
		if got := parseConversationVerdict(tc.text); got.Verdict != tc.want {
			t.Errorf("parseConversationVerdict(%q).Verdict = %q, want %q", tc.text, got.Verdict, tc.want)
		}
	}
}
