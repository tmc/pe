package main

import (
	"strings"
	"testing"

	"github.com/tmc/pe/internal/promptfoo"
)

func TestFormatResultsAsEnhancedTable(t *testing.T) {
	result := promptfoo.EvaluationResult{
		EvalID: "eval-1",
		Results: promptfoo.ResultSet{
			Stats: promptfoo.Stats{
				Successes:  1,
				Failures:   1,
				Errors:     1,
				TokenUsage: promptfoo.TokenUsage{Total: 30, Prompt: 10, Completion: 20},
			},
			Results: []promptfoo.TestResult{
				{
					ID:       "2",
					Prompt:   map[string]string{"label": "convert"},
					Provider: map[string]string{"id": "provider-b-with-a-long-name"},
					Response: promptfoo.ProviderResponse{Output: "this output is deliberately long"},
					Success:  false,
					Vars:     map[string]interface{}{"input": "world", "language": "French"},
				},
				{
					ID:       "1",
					Prompt:   map[string]string{"label": "translate"},
					Provider: map[string]string{"id": "provider-a"},
					Response: promptfoo.ProviderResponse{Output: "Bonjour"},
					Success:  true,
					Vars:     map[string]interface{}{"input": "hello", "language": "French"},
				},
			},
		},
	}

	out := string(formatResultsAsEnhancedTable(result))
	for _, want := range []string{
		"Evaluation complete. ID: eval-1",
		"Successes: 1",
		"Failures: 1",
		"Errors: 1",
		"Pass Rate: 50.00%",
		"Total tokens: 30 / Prompt tokens: 10 / Completion tokens: 20 / Cached tokens: 0",
		"provider-b-wi...",
		"Bonjour",
		"this output ...",
		"[N/A]",
		"Done.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "hello") > strings.Index(out, "world") {
		t.Fatalf("rows not sorted by language/input:\n%s", out)
	}
}

func TestFormatResultsAsEnhancedTableDefaultPassRate(t *testing.T) {
	out := string(formatResultsAsEnhancedTable(promptfoo.EvaluationResult{
		EvalID:  "empty",
		Results: promptfoo.ResultSet{Stats: promptfoo.Stats{}},
	}))
	if !strings.Contains(out, "Pass Rate: 100.00%") {
		t.Fatalf("output = %s", out)
	}
}
