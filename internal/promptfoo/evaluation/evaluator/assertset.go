package evaluator

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// assert-set groups child assertions and reports a single combined result,
// mirroring promptfoo's assert-set. The child assertions are taken from the
// assertion's value (a list) or config["assert"]. Each child's score is
// weighted (config/value "weight", default 1) and averaged; the set passes when
// the weighted-average score meets the set threshold. With no threshold the set
// passes only when every child passes (promptfoo's default behavior).

// normalizeAssertSetType reports whether an assertion id is the assert-set
// grouping (tolerating the same _/- and case normalization as other ids).
func normalizeAssertSetType(id string) bool {
	id = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(id)), "_", "-")
	return id == "assert-set"
}

// childAssertions decodes the list of child assertions from an assert-set's
// value or config["assert"].
func childAssertions(assert promptfoo.Assertion) []promptfoo.Assertion {
	raw := assert.Value
	if raw == nil && assert.Config != nil {
		raw = assert.Config["assert"]
	}
	list, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	children := make([]promptfoo.Assertion, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		child := promptfoo.Assertion{}
		if t, ok := m["type"].(string); ok {
			child.Type = t
		}
		if v, ok := m["value"]; ok {
			child.Value = v
		}
		if p, ok := m["provider"].(string); ok {
			child.Provider = p
		}
		if th, ok := asFloat(m["threshold"]); ok {
			child.Threshold = th
		}
		if w, ok := m["weight"]; ok {
			if child.Config == nil {
				child.Config = map[string]interface{}{}
			}
			child.Config["weight"] = w
		}
		if cfg, ok := m["config"].(map[string]interface{}); ok {
			if child.Config == nil {
				child.Config = map[string]interface{}{}
			}
			for k, v := range cfg {
				child.Config[k] = v
			}
		}
		if mt, ok := m["metric"].(string); ok {
			child.Metric = mt
		}
		if child.Type != "" {
			children = append(children, child)
		}
	}
	return children
}

// assertionWeight reads a child's weight (config["weight"], default 1).
func assertionWeight(assert promptfoo.Assertion) float64 {
	if assert.Config != nil {
		if w, ok := asFloat(assert.Config["weight"]); ok && w >= 0 {
			return w
		}
	}
	return 1
}

// evaluateAssertSet runs the child assertions and combines them into one
// weighted result.
func evaluateAssertSet(ctx context.Context, output string, assert promptfoo.Assertion, judgeProvider llm.Provider, meta map[string]interface{}) (bool, float64, string) {
	children := childAssertions(assert)
	if len(children) == 0 {
		return false, 0, "assert-set requires a non-empty list of child assertions"
	}

	totalWeight := 0.0
	weightedScore := 0.0
	allPassed := true
	failures := make([]string, 0)
	for _, child := range children {
		pass, score, reason := evaluateAssertion(ctx, output, child, judgeProvider, meta)
		w := assertionWeight(child)
		totalWeight += w
		weightedScore += w * score
		if !pass {
			allPassed = false
			failures = append(failures, fmt.Sprintf("%s: %s", child.Type, reason))
		}
	}

	avg := 0.0
	if totalWeight > 0 {
		avg = weightedScore / totalWeight
	}

	// With an explicit threshold the set passes on the weighted average; without
	// one it passes only when every child passed (promptfoo's default).
	var passed bool
	if assert.Threshold > 0 {
		passed = avg >= assert.Threshold
	} else {
		passed = allPassed
	}

	if passed {
		return true, avg, fmt.Sprintf("assert-set passed (%d assertions, score %.3f)", len(children), avg)
	}
	detail := strings.Join(failures, "; ")
	if detail == "" {
		detail = fmt.Sprintf("score %.3f below threshold %.2f", avg, assert.Threshold)
	}
	return false, avg, fmt.Sprintf("assert-set failed: %s", detail)
}
