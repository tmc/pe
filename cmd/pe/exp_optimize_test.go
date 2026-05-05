package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpOptimizeSelectsDeterministicBest(t *testing.T) {
	out := runExpOptimizeCommand(t, `{
  "seed": {"source":"seed", "prompt":"say hi", "score":0.1},
  "variants": [
    {"source":"a", "prompt":"say hello", "score":0.4},
    {"source":"b", "prompt":"say hello clearly", "score":0.9}
  ]
}`)

	var got expOptimizeOutput
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v\n%s", err, out)
	}
	if got.Selected.Source != "b" {
		t.Fatalf("selected source = %q, want b", got.Selected.Source)
	}
	if got.Selected.Score != 0.9 {
		t.Fatalf("selected score = %v, want 0.9", got.Selected.Score)
	}
	if len(got.Trajectory) != 4 {
		t.Fatalf("trajectory length = %d, want 4", len(got.Trajectory))
	}
	if !got.Trajectory[3].Accepted {
		t.Fatalf("accepted step missing from trajectory")
	}
}

func TestExpOptimizeConvergesOnMinimumImprovement(t *testing.T) {
	out := runExpOptimizeCommand(t, `{
  "seed": {"source":"seed", "prompt":"base", "score":0.5},
  "min_improvement": 0.2,
  "variants": [
    {"source":"small", "prompt":"small improvement", "score":0.6}
  ]
}`)

	var got expOptimizeOutput
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !got.Converged {
		t.Fatalf("converged = false, want true")
	}
	if got.Selected.Source != "seed" {
		t.Fatalf("selected source = %q, want seed", got.Selected.Source)
	}
}

func TestExpOptimizeHonorsIterationLimit(t *testing.T) {
	out := runExpOptimizeCommand(t, `{
  "seed": {"source":"seed", "prompt":"base", "score":0.1},
  "max_rounds": 1,
  "rounds": [
    [{"source":"first", "prompt":"first", "score":0.2}],
    [{"source":"second", "prompt":"second", "score":0.9}]
  ]
}`)

	var got expOptimizeOutput
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got.Selected.Source != "first" {
		t.Fatalf("selected source = %q, want first", got.Selected.Source)
	}
	if got.Iterations != 1 {
		t.Fatalf("iterations = %d, want 1", got.Iterations)
	}
}

func TestExpOptimizeRejectsInvalidInput(t *testing.T) {
	cmd := expOptimizeCmd()
	cmd.SetIn(strings.NewReader(`{
  "seed": {"source":"seed", "prompt":"base", "score":0.1},
  "variants": [
    {"source":"missing", "prompt":"missing score"}
  ]
}`))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected missing score error")
	}
	if !strings.Contains(err.Error(), `score missing for "missing"`) {
		t.Fatalf("error = %v, want missing score", err)
	}
}

func TestExpOptimizeReadsAndWritesFiles(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.json")
	outputPath := filepath.Join(dir, "output.json")
	input := `{
  "seed": {"source":"seed", "prompt":"base", "score":0.1},
  "variants": [
    {"source":"candidate", "prompt":"from file", "score":0.7}
  ]
}`
	if err := os.WriteFile(inputPath, []byte(input), 0666); err != nil {
		t.Fatalf("write input: %v", err)
	}

	cmd := expOptimizeCmd()
	cmd.SetArgs([]string{"--input", inputPath, "--output", outputPath})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute optimize: %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want empty when writing file", out.String())
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var got expOptimizeOutput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal output file: %v", err)
	}
	if got.Selected.Source != "candidate" {
		t.Fatalf("selected source = %q, want candidate", got.Selected.Source)
	}
}

func TestExpOptimizeJSONOutputStable(t *testing.T) {
	out := runExpOptimizeCommand(t, `{
  "seed": {"source":"seed", "prompt":"base", "score":0.25},
  "variants": [
    {"source":"candidate", "prompt":"better", "score":0.75}
  ]
}`)

	want := `{
  "selected": {
    "round": 1,
    "source": "candidate",
    "prompt": "better",
    "score": 0.75,
    "reason": "provided local score",
    "metrics": {
      "provided": 0.75
    },
    "improvement": 0.5,
    "accepted": true
  },
  "initial": {
    "round": 0,
    "source": "seed",
    "prompt": "base",
    "score": 0.25,
    "reason": "provided local score",
    "metrics": {
      "provided": 0.25
    }
  },
  "trajectory": [
    {
      "round": 0,
      "source": "seed",
      "prompt": "base",
      "score": 0.25,
      "reason": "provided local score",
      "metrics": {
        "provided": 0.25
      }
    },
    {
      "round": 1,
      "source": "candidate",
      "prompt": "better",
      "score": 0.75,
      "reason": "provided local score",
      "metrics": {
        "provided": 0.75
      },
      "improvement": 0.5
    },
    {
      "round": 1,
      "source": "candidate",
      "prompt": "better",
      "score": 0.75,
      "reason": "provided local score",
      "metrics": {
        "provided": 0.75
      },
      "improvement": 0.5,
      "accepted": true
    }
  ],
  "converged": false,
  "iterations": 1
}
`
	if out != want {
		t.Fatalf("stable JSON mismatch\ngot:\n%s\nwant:\n%s", out, want)
	}
}

func runExpOptimizeCommand(t *testing.T, input string) string {
	t.Helper()

	cmd := expOptimizeCmd()
	cmd.SetIn(strings.NewReader(input))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute optimize: %v\n%s", err, out.String())
	}
	return out.String()
}
