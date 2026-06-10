package evaluator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tmc/pe/internal/promptfoo"
)

func TestNormalizeAssertionType(t *testing.T) {
	tests := []struct {
		id        string
		canonical AssertionType
		negate    bool
		ok        bool
	}{
		// promptfoo spellings resolve to pe canonical ids.
		{"regex", AssertionMatches, false, true},
		{"llm-rubric", AssertionLLMJudge, false, true},
		{"similar", AssertionSimilarity, false, true},
		{"similar:cosine", AssertionSimilarity, false, true},
		{"is-json", AssertionJSON, false, true},
		{"is-sql", AssertionSQL, false, true},
		{"classifier", AssertionClassify, false, true},
		{"model-graded-factuality", AssertionFactuality, false, true},
		// pe canonical ids resolve to themselves.
		{"matches", AssertionMatches, false, true},
		{"llm-judge", AssertionLLMJudge, false, true},
		{"contains", AssertionContains, false, true},
		// not-* inversion. "not-contains" is also a canonical pe id, but the
		// not- prefix is stripped first, so it resolves to contains+negate,
		// which is semantically equivalent (and still passes end to end).
		{"not-regex", AssertionMatches, true, true},
		{"not-contains", AssertionContains, true, true},
		{"not-llm-rubric", AssertionLLMJudge, true, true},
		// case, whitespace, and separator tolerance.
		{"  Regex  ", AssertionMatches, false, true},
		{"IS-JSON", AssertionJSON, false, true},
		{"llm_rubric", AssertionLLMJudge, false, true},
		// unknown ids report ok=false.
		{"answer-relevance", "", false, false},
		{"perplexity", "", false, false},
		{"made-up", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			canonical, negate, ok := normalizeAssertionType(tt.id)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.negate, negate)
			if tt.ok {
				assert.Equal(t, tt.canonical, canonical)
			}
		})
	}
}

// TestEvaluateAssertionsPromptfooIds checks that an unmodified promptfoo-spelled
// assertion config runs through pe's evaluator end to end.
func TestEvaluateAssertionsPromptfooIds(t *testing.T) {
	tests := []struct {
		name   string
		assert promptfoo.Assertion
		output string
		want   bool
	}{
		{"regex pass", promptfoo.Assertion{Type: "regex", Value: `ans.er`}, "answer", true},
		{"regex fail", promptfoo.Assertion{Type: "regex", Value: `^xyz$`}, "answer", false},
		{"is-json pass", promptfoo.Assertion{Type: "is-json"}, `{"ok":true}`, true},
		{"is-json fail", promptfoo.Assertion{Type: "is-json"}, `not json`, false},
		{"not-contains pass", promptfoo.Assertion{Type: "not-contains", Value: "missing"}, "answer", true},
		{"not-regex inverts", promptfoo.Assertion{Type: "not-regex", Value: `ans.er`}, "answer", false},
		{"unknown id falls back", promptfoo.Assertion{Type: "answer-relevance", Value: "x"}, "answer", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, _ := evaluateAssertions(tt.output, []promptfoo.Assertion{tt.assert})
			assert.Equal(t, tt.want, pass)
		})
	}
}

func TestEvaluateAssertionsPopulatesNamedScores(t *testing.T) {
	_, grading := evaluateAssertions("hello world", []promptfoo.Assertion{
		{Type: "contains", Value: "hello", Metric: "Greeting"},
		{Type: "contains", Value: "world", Metric: "Greeting"},
		{Type: "contains", Value: "missing", Metric: "Coverage"},
	})
	// Two passing "Greeting" assertions sum to 2.0; one failing "Coverage" is 0.
	if got := grading.NamedScores["Greeting"]; got != 2.0 {
		t.Fatalf("Greeting named score = %v, want 2.0", got)
	}
	if got, ok := grading.NamedScores["Coverage"]; !ok || got != 0.0 {
		t.Fatalf("Coverage named score = %v (ok=%v), want 0.0", got, ok)
	}
}

func TestPromptfooAssertionNormalizesType(t *testing.T) {
	got := promptfooAssertion(promptfoo.Assertion{Type: "regex", Value: "x"})
	assert.Equal(t, AssertionMatches, got.Type)
	require.NotNil(t, got.Value)

	// Unknown ids pass through verbatim so the caller can still inspect them.
	got = promptfooAssertion(promptfoo.Assertion{Type: "answer-relevance"})
	assert.Equal(t, AssertionType("answer-relevance"), got.Type)
}
