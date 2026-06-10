package evaluator

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/promptfoo"
)

func TestFormatResultsAsJUnit(t *testing.T) {
	results := promptfoo.EvaluationResult{
		EvalID: "eval-test",
		Results: promptfoo.ResultSet{
			Timestamp: "2026-06-09T00:00:00Z",
			Results: []promptfoo.TestResult{
				{
					ID:        "t1",
					Prompt:    map[string]string{"raw": "say hi"},
					Provider:  map[string]string{"id": "openai:gpt-4o-mini"},
					Response:  promptfoo.ProviderResponse{Output: "hi"},
					Success:   true,
					LatencyMs: 120,
				},
				{
					ID:        "t2",
					Prompt:    map[string]string{"raw": "say bye"},
					Provider:  map[string]string{"id": "openai:gpt-4o-mini"},
					Response:  promptfoo.ProviderResponse{Output: "nope"},
					Success:   false,
					LatencyMs: 80,
					GradingResult: promptfoo.GradingResult{
						Reason: "Some assertions failed",
						ComponentResults: []promptfoo.ComponentResult{
							{Pass: false, Reason: "expected output to contain bye", Assertion: promptfoo.Assertion{Type: "contains"}},
						},
					},
				},
			},
		},
	}

	data, err := FormatResults(results, "junit")
	if err != nil {
		t.Fatalf("FormatResults junit: %v", err)
	}

	// It must be well-formed XML.
	var report junitTestSuites
	if err := xml.Unmarshal(data, &report); err != nil {
		t.Fatalf("junit output is not valid XML: %v\n%s", err, data)
	}
	if report.Tests != 2 {
		t.Fatalf("tests = %d, want 2", report.Tests)
	}
	if report.Failures != 1 {
		t.Fatalf("failures = %d, want 1", report.Failures)
	}
	if len(report.Suites) != 1 || len(report.Suites[0].Cases) != 2 {
		t.Fatalf("unexpected suite shape: %+v", report.Suites)
	}

	// The failing case carries the assertion reason.
	var failing *junitTestCase
	for i := range report.Suites[0].Cases {
		if report.Suites[0].Cases[i].Failure != nil {
			failing = &report.Suites[0].Cases[i]
		}
	}
	if failing == nil {
		t.Fatal("expected one failing testcase with a <failure>")
	}
	if !strings.Contains(failing.Failure.Message, "contains") {
		t.Fatalf("failure message = %q, want it to mention the assertion type", failing.Failure.Message)
	}
	if !strings.Contains(string(data), xml.Header) {
		t.Fatal("expected XML header in output")
	}
}
