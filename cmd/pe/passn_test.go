package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmc/pe/internal/promptfoo/evaluation/metrics"
)

func TestPassNCmdStructure(t *testing.T) {
	if passnCmd.Use != "passn" {
		t.Fatalf("Use = %q", passnCmd.Use)
	}
	for _, name := range []string{"n", "samples", "provider", "temperature", "tests", "output", "save", "data-dir", "tags", "strategy", "structured", "schema"} {
		if passnCmd.Flags().Lookup(name) == nil {
			t.Fatalf("missing flag %q", name)
		}
	}
	for _, name := range []string{"report", "search", "compare"} {
		found := false
		for _, cmd := range passnCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing subcommand %q", name)
		}
	}
}

func TestPassNHelpers(t *testing.T) {
	if !containsError("runtime error") || !containsError("Undefined name") || containsError("working output") {
		t.Fatal("containsError mismatch")
	}
	if id := generatePromptID("prompt"); !strings.HasPrefix(id, "prompt-") || len(id) != len("prompt-")+64 {
		t.Fatalf("prompt id = %q", id)
	}
	if got := expandPath("/tmp/pe"); got != "/tmp/pe" {
		t.Fatalf("expand absolute = %q", got)
	}
	if got := expandPath("~/pe-passn-test"); !strings.Contains(got, "pe-passn-test") || strings.HasPrefix(got, "~") {
		t.Fatalf("expand home = %q", got)
	}

	casesFile := filepath.Join(t.TempDir(), "cases.json")
	if err := os.WriteFile(casesFile, []byte(`[{"input":"x","expected":"y"}]`), 0644); err != nil {
		t.Fatal(err)
	}
	cases, err := loadTestCases(casesFile)
	if err != nil || len(cases) != 1 || cases[0]["input"] != "x" {
		t.Fatalf("cases = %#v err=%v", cases, err)
	}
	if _, err := loadTestCases(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("missing test cases succeeded")
	}
	badFile := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(badFile, []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadTestCases(badFile); err == nil {
		t.Fatal("bad test cases succeeded")
	}
}

func TestPassNStructuredPrompt(t *testing.T) {
	oldType, oldSchema := passnStructuredType, passnSchema
	defer func() {
		passnStructuredType = oldType
		passnSchema = oldSchema
	}()

	for _, typ := range []string{"json", "yaml", "markdown"} {
		passnStructuredType = typ
		passnSchema = ""
		prompt, err := setupStructuredPrompt("Generate output")
		if err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		if !strings.Contains(prompt, "Generate output") {
			t.Fatalf("%s prompt = %q", typ, prompt)
		}
	}

	passnStructuredType = "json"
	passnSchema = filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(passnSchema, []byte(`{"name":"answer","type":"object","properties":{"result":{"type":"string"}}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := setupStructuredPrompt("Answer"); err != nil {
		t.Fatal(err)
	}
	passnSchema = filepath.Join(t.TempDir(), "missing.json")
	if _, err := setupStructuredPrompt("Answer"); err == nil {
		t.Fatal("missing schema succeeded")
	}
	passnSchema = filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(passnSchema, []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := setupStructuredPrompt("Answer"); err == nil {
		t.Fatal("bad schema succeeded")
	}
}

func TestOutputPassNResults(t *testing.T) {
	oldFormat := passnOutputFormat
	defer func() { passnOutputFormat = oldFormat }()

	result := &metrics.PassAtNResult{
		N:          2,
		PassRate:   0.5,
		NumSamples: 2,
		NumPassed:  1,
		Duration:   time.Millisecond,
	}
	eval := metrics.PassNEvaluation{
		ID:         "eval-1",
		N:          2,
		NumSamples: 2,
		NumPassed:  1,
		PassRate:   0.5,
		Duration:   time.Millisecond,
		Samples: []metrics.PassNSample{
			{Index: 0, Content: "ok", Passed: true},
			{Index: 1, Content: strings.Repeat("x", 240), Passed: false},
		},
	}
	for _, format := range []string{"json", "table", "report"} {
		passnOutputFormat = format
		if err := outputPassNResults(result, eval); err != nil {
			t.Fatalf("%s: %v", format, err)
		}
	}
	passnOutputFormat = "bogus"
	if err := outputPassNResults(result, eval); err == nil {
		t.Fatal("unsupported format succeeded")
	}
}

func TestRunPassNWithMockProvider(t *testing.T) {
	old := os.Getenv("PE_TEST_MODE")
	oldN, oldSamples, oldProvider := passnN, passnSamples, passnProvider
	oldOutput, oldSave, oldDir := passnOutputFormat, passnSaveResults, passnDataDir
	oldTestFile, oldStructured, oldSchema := passnTestFile, passnStructuredType, passnSchema
	defer func() {
		os.Setenv("PE_TEST_MODE", old)
		passnN, passnSamples, passnProvider = oldN, oldSamples, oldProvider
		passnOutputFormat, passnSaveResults, passnDataDir = oldOutput, oldSave, oldDir
		passnTestFile, passnStructuredType, passnSchema = oldTestFile, oldStructured, oldSchema
	}()
	os.Setenv("PE_TEST_MODE", "true")
	passnN = 1
	passnSamples = 2
	passnProvider = "mock"
	passnOutputFormat = "json"
	passnSaveResults = true
	passnDataDir = t.TempDir()
	passnTestFile = ""
	passnStructuredType = ""
	passnSchema = ""

	promptFile := filepath.Join(t.TempDir(), "prompt.txt")
	if err := os.WriteFile(promptFile, []byte("write a function"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runPassN(passnCmd, []string{promptFile}); err != nil {
		t.Fatal(err)
	}
	if err := runPassN(passnCmd, nil); err == nil {
		t.Fatal("missing prompt succeeded")
	}
}
