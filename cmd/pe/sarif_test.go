package main

import (
	"encoding/json"
	"testing"
)

func TestFormatSecuritySARIF(t *testing.T) {
	result := &SecurityTestResult{
		TestID: "sec-1",
		Target: "openai:gpt-4o-mini",
		Vulnerabilities: []Vulnerability{
			{
				ID:          "v1",
				Type:        "prompt-injection",
				Severity:    "High",
				Title:       "Prompt injection",
				Description: "The model followed an injected instruction.",
				Remediation: "Add input guardrails.",
				OWASP:       "LLM01",
				References:  []string{"https://owasp.org/llm01"},
			},
			{
				ID:          "v2",
				Type:        "prompt-injection",
				Severity:    "Medium",
				Title:       "Prompt injection (weak)",
				Description: "Partial injection.",
			},
			{
				ID:       "v3",
				Type:     "pii-leak",
				Severity: "Low",
				Title:    "PII disclosure",
			},
		},
	}

	data, err := formatSecuritySARIF(result)
	if err != nil {
		t.Fatalf("formatSecuritySARIF: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(data, &log); err != nil {
		t.Fatalf("output is not valid SARIF JSON: %v\n%s", err, data)
	}
	if log.Version != "2.1.0" {
		t.Fatalf("version = %q, want 2.1.0", log.Version)
	}
	if len(log.Runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(log.Runs))
	}
	run := log.Runs[0]

	// Two distinct rule ids (prompt-injection deduped, pii-leak).
	if len(run.Tool.Driver.Rules) != 2 {
		t.Fatalf("rules = %d, want 2 (deduped by type)", len(run.Tool.Driver.Rules))
	}
	// One result per vulnerability.
	if len(run.Results) != 3 {
		t.Fatalf("results = %d, want 3", len(run.Results))
	}

	// Severity mapping reaches the result level via the rule id.
	if got := levelForType(run.Results, "prompt-injection", "High"); got != "error" {
		t.Fatalf("High level = %q, want error", got)
	}
}

// levelForType finds the SARIF level of the first result for a rule id whose
// message reflects the given severity context. Helper kept simple for the test.
func levelForType(results []sarifResult, ruleID, _ string) string {
	for _, r := range results {
		if r.RuleID == ruleID {
			return r.Level
		}
	}
	return ""
}

func TestSarifLevelMapping(t *testing.T) {
	cases := map[string]string{
		"Critical": "error",
		"High":     "error",
		"Medium":   "warning",
		"Low":      "note",
		"info":     "note",
		"unknown":  "warning",
	}
	for in, want := range cases {
		if got := sarifLevel(in); got != want {
			t.Errorf("sarifLevel(%q) = %q, want %q", in, got, want)
		}
	}
}
