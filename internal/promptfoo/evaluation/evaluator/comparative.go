package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
	"github.com/tmc/pe/internal/providers"
)

// Comparative assertions (select-best, max-score) rank a test row's candidate
// outputs against each other, so they cannot be evaluated from a single output.
// They run in a post-pass over all results: for each test, the results sharing
// that test's vars are gathered as candidates, the assertion picks a winner,
// and each candidate's Success/Score/GradingResult is rewritten accordingly.

// isComparativeAssertion reports whether an assertion id is a row-level
// comparison (select-best or max-score), tolerating _/- and case differences.
func isComparativeAssertion(id string) bool {
	id = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(id)), "_", "-")
	return id == "select-best" || id == "max-score"
}

// applyComparativeAssertions runs every test's comparative assertions over the
// results sharing that test's vars and rewrites those results. It reports
// whether any result was modified (so the caller recounts pass/fail).
func applyComparativeAssertions(ctx context.Context, config promptfoo.Config, results []promptfoo.TestResult, materialized []*providers.MaterializedProvider) bool {
	var judge llm.Provider
	for _, p := range materialized {
		if p != nil && p.Executor != nil {
			judge = p.Executor
			break
		}
	}

	changed := false
	for _, test := range config.Tests {
		comparatives := comparativeAsserts(test.Assert)
		if len(comparatives) == 0 {
			continue
		}
		idx := matchingResultIndexes(results, test.Vars)
		if len(idx) == 0 {
			continue
		}
		for _, assert := range comparatives {
			if applyOneComparative(ctx, assert, results, idx, judge) {
				changed = true
			}
		}
	}
	return changed
}

func comparativeAsserts(asserts []promptfoo.Assertion) []promptfoo.Assertion {
	var out []promptfoo.Assertion
	for _, a := range asserts {
		if isComparativeAssertion(a.Type) {
			out = append(out, a)
		}
	}
	return out
}

// matchingResultIndexes returns the indexes of results whose vars equal the
// test's vars (i.e. the same test row across providers/repeats).
func matchingResultIndexes(results []promptfoo.TestResult, vars map[string]interface{}) []int {
	want := canonicalVars(vars)
	var idx []int
	for i := range results {
		if canonicalVars(results[i].Vars) == want {
			idx = append(idx, i)
		}
	}
	return idx
}

func canonicalVars(vars map[string]interface{}) string {
	if len(vars) == 0 {
		return "{}"
	}
	b, err := json.Marshal(vars)
	if err != nil {
		return fmt.Sprintf("%v", vars)
	}
	return string(b)
}

// applyOneComparative evaluates a single comparative assertion over the
// candidate results (by index) and rewrites them. It returns whether anything
// changed.
func applyOneComparative(ctx context.Context, assert promptfoo.Assertion, results []promptfoo.TestResult, idx []int, judge llm.Provider) bool {
	outputs := make([]string, len(idx))
	for i, ri := range idx {
		outputs[i] = results[ri].Response.Output
	}

	var winner int
	var reason string
	switch {
	case strings.Contains(strings.ToLower(assert.Type), "max-score"):
		winner, reason = pickMaxScore(results, idx)
	default: // select-best
		var err error
		winner, reason, err = pickSelectBest(ctx, assert, outputs, judge)
		if err != nil {
			// Without a judge or on judge failure, mark every candidate failed
			// with the error so the problem is visible rather than silent.
			for _, ri := range idx {
				setComparativeResult(&results[ri], assert, false, 0, err.Error())
			}
			return true
		}
	}

	for n, ri := range idx {
		if n == winner {
			setComparativeResult(&results[ri], assert, true, 1, fmt.Sprintf("selected as best: %s", reason))
		} else {
			setComparativeResult(&results[ri], assert, false, 0, "not selected as best")
		}
	}
	return true
}

// pickMaxScore selects the candidate with the highest existing aggregate
// assertion score (GradingResult.Score), mirroring promptfoo's max-score.
func pickMaxScore(results []promptfoo.TestResult, idx []int) (int, string) {
	best, bestScore := 0, results[idx[0]].GradingResult.Score
	for n, ri := range idx {
		if s := results[ri].GradingResult.Score; s > bestScore {
			best, bestScore = n, s
		}
	}
	return best, fmt.Sprintf("highest aggregate score %.3f", bestScore)
}

// pickSelectBest asks the judge to choose the output that best meets the
// criteria (assert.Value), returning the winner's index. It mirrors promptfoo's
// select-best prompt: candidates are numbered and the judge returns the index.
func pickSelectBest(ctx context.Context, assert promptfoo.Assertion, outputs []string, judge llm.Provider) (int, string, error) {
	if judge == nil {
		return 0, "", fmt.Errorf("select-best requires a judge provider")
	}
	criteria, _ := assert.Value.(string)
	if strings.TrimSpace(criteria) == "" {
		criteria = "overall quality"
	}

	var b strings.Builder
	for i, o := range outputs {
		fmt.Fprintf(&b, "OUTPUT %d:\n%s\n\n", i, o)
	}
	prompt := fmt.Sprintf(`You are comparing several candidate outputs and selecting the single best one.

CRITERIA:
%s

CANDIDATES:
%s
TASK:
Return ONLY the index (an integer from 0 to %d) of the best output, then a
brief reason.

FORMAT (exactly):
INDEX: [integer]
REASON: [brief explanation]`, criteria, b.String(), len(outputs)-1)

	resp, err := judge.Generate(ctx, prompt, llm.GenerateOptions{})
	if err != nil {
		return 0, "", fmt.Errorf("select-best judge call failed: %w", err)
	}
	idx, reason := parseSelectBest(resp.Text, len(outputs))
	return idx, reason, nil
}

// parseSelectBest extracts the winning index (clamped to range) and reason from
// the judge reply.
func parseSelectBest(text string, n int) (int, string) {
	idx, reason := 0, ""
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "INDEX:"); ok {
			if v, err := strconv.Atoi(strings.TrimSpace(rest)); err == nil {
				idx = v
			}
		} else if rest, ok := strings.CutPrefix(line, "REASON:"); ok {
			reason = strings.TrimSpace(rest)
		}
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return idx, reason
}

// setComparativeResult records the comparative assertion's outcome on a result,
// appending a component result and updating the overall Success/Score. A failed
// comparative assertion fails the whole result, matching how other assertions
// combine.
func setComparativeResult(result *promptfoo.TestResult, assert promptfoo.Assertion, pass bool, score float64, reason string) {
	result.GradingResult.ComponentResults = append(result.GradingResult.ComponentResults, promptfoo.ComponentResult{
		Pass:      pass,
		Score:     score,
		Reason:    reason,
		Assertion: assert,
	})
	if !pass {
		result.Success = false
		result.GradingResult.Pass = false
		result.Score = 0
	}
	if assert.Metric != "" {
		if result.GradingResult.NamedScores == nil {
			result.GradingResult.NamedScores = map[string]float64{}
		}
		result.GradingResult.NamedScores[assert.Metric] += score
	}
}
