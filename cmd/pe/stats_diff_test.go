package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsCmdReadsEvaluationJSON(t *testing.T) {
	file := writeTempFile(t, `{
  "evalId": "eval-1",
  "results": {
    "results": [
      {
        "id": "r1",
        "provider": {"id": "openai:gpt-4"},
        "success": true,
        "score": 1,
        "latencyMs": 100,
        "response": {"tokenUsage": {"total": 12, "prompt": 5, "completion": 7}, "cost": 0.03}
      },
      {
        "id": "r2",
        "provider": {"id": "openai:gpt-4"},
        "success": false,
        "score": 0.25,
        "latencyMs": 300,
        "response": {"tokenUsage": {"total": 8, "prompt": 4, "completion": 4}, "cost": 0.01}
      }
    ]
  }
}`)

	cmd := statsCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{file})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("stats Execute() error = %v", err)
	}
	for _, want := range []string{
		"Total tests: 2",
		"Successes: 1",
		"Failures: 1",
		"Pass rate: 50.00%",
		"Average score: 0.6250",
		"Average latency: 200.00 ms",
		"Total tokens: 20",
		"openai:gpt-4: tests=2",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("stats output missing %q:\n%s", want, out.String())
		}
	}
}

func TestStatsCmdReadsJSONLFromStdin(t *testing.T) {
	cmd := statsCmd()
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(strings.Join([]string{
		`{"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":10}`,
		`{"id":"r2","provider":{"id":"mock"},"success":true,"score":0.5,"latencyMs":30}`,
		"",
	}, "\n")))
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--format", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("stats Execute() error = %v", err)
	}

	var got resultSummary
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("stats JSON output did not decode: %v\n%s", err, out.String())
	}
	if got.TotalTests != 2 || got.Successes != 2 || got.Failures != 0 {
		t.Fatalf("stats summary = %+v, want 2 successes and no failures", got)
	}
	if got.AverageLatencyMs != 20 {
		t.Fatalf("AverageLatencyMs = %v, want 20", got.AverageLatencyMs)
	}
}

func TestDiffCmdWritesJSONDelta(t *testing.T) {
	base := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":100,"response":{"tokenUsage":{"total":10}}},
  {"id":"r2","provider":{"id":"mock"},"success":false,"score":0,"latencyMs":200,"response":{"tokenUsage":{"total":20}}}
]}`)
	current := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":80,"response":{"tokenUsage":{"total":11}}},
  {"id":"r2","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":120,"response":{"tokenUsage":{"total":21}}}
]}`)

	cmd := diffCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--format", "json", base, current})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("diff Execute() error = %v", err)
	}

	var got resultDiff
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("diff JSON output did not decode: %v\n%s", err, out.String())
	}
	if got.Delta.Successes != 1 {
		t.Fatalf("Delta.Successes = %d, want 1", got.Delta.Successes)
	}
	if got.Delta.Failures != -1 {
		t.Fatalf("Delta.Failures = %d, want -1", got.Delta.Failures)
	}
	if got.Delta.AverageLatencyMs != -50 {
		t.Fatalf("Delta.AverageLatencyMs = %v, want -50", got.Delta.AverageLatencyMs)
	}
}

func TestDiffCmdFailOnRegression(t *testing.T) {
	base := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":100,"response":{"tokenUsage":{"total":10}}},
  {"id":"r2","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":100,"response":{"tokenUsage":{"total":10}}}
]}`)
	current := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":120,"response":{"tokenUsage":{"total":10}}},
  {"id":"r2","provider":{"id":"mock"},"success":false,"score":0,"latencyMs":180,"response":{"tokenUsage":{"total":10}}}
]}`)

	cmd := diffCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--fail-on-regression", base, current})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("diff Execute() error = nil, want regression gate failure")
	}
	if !strings.Contains(err.Error(), "diff gate failed") {
		t.Fatalf("diff error = %v, want gate failure", err)
	}
	output := out.String()
	for _, want := range []string{
		"Gate: fail",
		"pass rate dropped 50.00 percentage points",
		"average score dropped 0.5000",
		"average latency increased 50.00 ms",
		"failures increased +1",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("diff gate output missing %q:\n%s", want, output)
		}
	}
}

func TestDiffCmdRegressionThresholdsCanPass(t *testing.T) {
	base := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":100,"response":{"tokenUsage":{"total":10}}},
  {"id":"r2","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":100,"response":{"tokenUsage":{"total":10}}}
]}`)
	current := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":120,"response":{"tokenUsage":{"total":10}}},
  {"id":"r2","provider":{"id":"mock"},"success":false,"score":0,"latencyMs":180,"response":{"tokenUsage":{"total":10}}}
]}`)

	cmd := diffCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"--format", "json",
		"--fail-on-regression",
		"--max-pass-rate-drop", "60",
		"--max-score-drop", "0.6",
		"--max-latency-increase-ms", "60",
		"--max-failure-increase", "1",
		base, current,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("diff Execute() error = %v", err)
	}

	var got resultDiff
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("diff JSON output did not decode: %v\n%s", err, out.String())
	}
	if got.Gate == nil {
		t.Fatal("Gate = nil, want gate result")
	}
	if !got.Gate.Passed {
		t.Fatalf("Gate.Passed = false, reasons = %v", got.Gate.Reasons)
	}
}

func TestDiffCmdFailOnChange(t *testing.T) {
	base := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":100}
]}`)
	current := writeTempFile(t, `{"results":[
  {"id":"r1","provider":{"id":"mock"},"success":true,"score":1,"latencyMs":101}
]}`)

	cmd := diffCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--fail-on-change", base, current})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("diff Execute() error = nil, want change gate failure")
	}
	if !strings.Contains(out.String(), "average latency changed") {
		t.Fatalf("diff gate output missing latency reason:\n%s", out.String())
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "results.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
