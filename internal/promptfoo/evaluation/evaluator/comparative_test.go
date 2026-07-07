package evaluator

import (
	"context"
	"testing"

	"github.com/tmc/pe/internal/promptfoo"
)

func TestIsComparativeAssertion(t *testing.T) {
	for _, id := range []string{"select-best", "select_best", "MAX-SCORE", " max-score "} {
		if !isComparativeAssertion(id) {
			t.Errorf("%q should be comparative", id)
		}
	}
	if isComparativeAssertion("contains") {
		t.Errorf("contains is not comparative")
	}
}

func TestMatchingResultIndexes(t *testing.T) {
	results := []promptfoo.TestResult{
		{Vars: map[string]interface{}{"q": "a"}},
		{Vars: map[string]interface{}{"q": "b"}},
		{Vars: map[string]interface{}{"q": "a"}},
	}
	idx := matchingResultIndexes(results, map[string]interface{}{"q": "a"})
	if len(idx) != 2 || idx[0] != 0 || idx[1] != 2 {
		t.Errorf("matchingResultIndexes = %v, want [0 2]", idx)
	}
}

func TestParseSelectBest(t *testing.T) {
	tests := []struct {
		text string
		n    int
		want int
	}{
		{"INDEX: 1\nREASON: best", 3, 1},
		{"INDEX: 5\nREASON: clamp", 3, 2}, // clamp to n-1
		{"INDEX: -2", 3, 0},               // clamp to 0
		{"no index here", 3, 0},
	}
	for _, tc := range tests {
		if got, _ := parseSelectBest(tc.text, tc.n); got != tc.want {
			t.Errorf("parseSelectBest(%q) = %d, want %d", tc.text, got, tc.want)
		}
	}
}

func TestPickMaxScore(t *testing.T) {
	results := []promptfoo.TestResult{
		{GradingResult: promptfoo.GradingResult{Score: 0.4}},
		{GradingResult: promptfoo.GradingResult{Score: 0.9}},
		{GradingResult: promptfoo.GradingResult{Score: 0.7}},
	}
	winner, _ := pickMaxScore(results, []int{0, 1, 2})
	if winner != 1 {
		t.Errorf("pickMaxScore winner = %d, want 1", winner)
	}
}

func TestApplyMaxScoreRewrites(t *testing.T) {
	results := []promptfoo.TestResult{
		{Success: true, Score: 1, GradingResult: promptfoo.GradingResult{Score: 0.4, Pass: true}},
		{Success: true, Score: 1, GradingResult: promptfoo.GradingResult{Score: 0.9, Pass: true}},
	}
	assert := promptfoo.Assertion{Type: "max-score"}
	changed := applyOneComparative(context.Background(), assert, results, []int{0, 1}, nil)
	if !changed {
		t.Fatal("expected applyOneComparative to report a change")
	}
	if results[0].Success { // index 0 lost
		t.Errorf("loser (idx 0) should now fail")
	}
	if !results[1].Success { // index 1 won
		t.Errorf("winner (idx 1) should still pass")
	}
	// Each gained a component result for the comparative assertion.
	if len(results[0].GradingResult.ComponentResults) != 1 {
		t.Errorf("expected a comparative component result on the loser")
	}
}

func TestApplySelectBestWithJudge(t *testing.T) {
	judge := &assertionJudgeTestProvider{response: "INDEX: 1\nREASON: most complete"}
	results := []promptfoo.TestResult{
		{Success: true, Score: 1, Response: promptfoo.ProviderResponse{Output: "short"}, GradingResult: promptfoo.GradingResult{Pass: true}},
		{Success: true, Score: 1, Response: promptfoo.ProviderResponse{Output: "a thorough answer"}, GradingResult: promptfoo.GradingResult{Pass: true}},
	}
	assert := promptfoo.Assertion{Type: "select-best", Value: "the most complete answer"}
	changed := applyOneComparative(context.Background(), assert, results, []int{0, 1}, judge)
	if !changed {
		t.Fatal("expected change")
	}
	if results[0].Success {
		t.Errorf("idx 0 should lose select-best")
	}
	if !results[1].Success {
		t.Errorf("idx 1 should win select-best")
	}
}

func TestApplySelectBestNoJudgeFails(t *testing.T) {
	results := []promptfoo.TestResult{
		{Success: true, Response: promptfoo.ProviderResponse{Output: "x"}, GradingResult: promptfoo.GradingResult{Pass: true}},
		{Success: true, Response: promptfoo.ProviderResponse{Output: "y"}, GradingResult: promptfoo.GradingResult{Pass: true}},
	}
	assert := promptfoo.Assertion{Type: "select-best", Value: "best"}
	applyOneComparative(context.Background(), assert, results, []int{0, 1}, nil)
	// Without a judge, both candidates fail with the error surfaced.
	if results[0].Success || results[1].Success {
		t.Errorf("select-best without a judge should fail all candidates")
	}
}
