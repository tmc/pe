package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

func TestAdvancedTestConfigOutputAndSummary(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(`
providers: []
prompts:
  - Say hello
options:
  temperature: 0.2
property_tests:
  - name: contains
    generator: numbers
    constraint: response.length > 0
regression_tests:
  - name: baseline
    metrics: [success_rate]
`), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadTestConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Providers) != 1 || cfg.Providers[0] != "openai:gpt-3.5-turbo" || cfg.Options["temperature"].(float64) != 0.2 {
		t.Fatalf("config = %#v", cfg)
	}
	if _, err := loadTestConfig(filepath.Join(tmpDir, "missing.yaml")); err == nil {
		t.Fatal("missing config loaded")
	}
	badConfig := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(badConfig, []byte("providers: ["), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadTestConfig(badConfig); err == nil {
		t.Fatal("bad config loaded")
	}

	results := &TestResults{
		Timestamp: time.Now(),
		Config:    configFile,
		TestType:  "regression",
		Summary: TestSummary{
			TotalTests:        2,
			PassedTests:       1,
			FailedTests:       1,
			SuccessRate:       50,
			TotalRegressions:  1,
			TotalImprovements: 1,
		},
	}
	cmd := testCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	outFile := filepath.Join(tmpDir, "results.json")
	if err := outputTestResults(results, outFile, cmd); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(outFile); err != nil || !strings.Contains(string(data), "regression") {
		t.Fatalf("results file = %q err=%v", data, err)
	}
	if err := outputTestResults(results, "", cmd); err != nil {
		t.Fatal(err)
	}
	printTestSummary(results, cmd)
	text := buf.String()
	for _, want := range []string{"Results saved", "Regression Test Results", "Total Tests", "Some tests failed"} {
		if !strings.Contains(text, want) {
			t.Fatalf("summary missing %q:\n%s", want, text)
		}
	}
}

func TestAdvancedTestSpecialCommandsAndGenerators(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte("Analyze sentiment"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := testCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := runGenerateTests(cmd, promptFile, "mock"); err != nil {
		t.Fatal(err)
	}
	if err := runGenerateTests(cmd, filepath.Join(tmpDir, "missing.txt"), "mock"); err == nil {
		t.Fatal("missing generate-tests succeeded")
	}
	cmd.Flags().Set("config-a", "a.yaml")
	cmd.Flags().Set("config-b", "b.yaml")
	if err := runABTest(cmd, "mock"); err != nil {
		t.Fatal(err)
	}
	if err := runCrossValidate(cmd, "config.yaml", "mock"); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	for _, want := range []string{"Generated test cases", "A/B Test Results", "Cross-Validation Results"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
	if tests := generateSystematicTests([]string{"p1", "p2"}, []string{"contains", "length"}); len(tests) != 4 {
		t.Fatalf("systematic tests = %#v", tests)
	}
	if cases := generateTestCases("source", "template", 3); len(cases) != 3 || cases[0]["template"] != "template" {
		t.Fatalf("cases = %#v", cases)
	}
	if cv := performCrossValidation([]string{"a"}, 2, true); cv["folds"] != 2 || cv["statistical"] != true {
		t.Fatalf("cross validation = %#v", cv)
	}
	if sig := performSignificanceTests("old", "new", 0.01, []string{"t"}); sig["alpha"] != 0.01 || sig["significant"] != true {
		t.Fatalf("significance = %#v", sig)
	}
	if ab := performABTest("a", "b", "score", true, 0.8, 0.2); ab["metric"] != "score" || ab["bayesian"] != true {
		t.Fatalf("ab = %#v", ab)
	}
}

func TestAdvancedTestSubcommands(t *testing.T) {
	tmpDir := t.TempDir()
	for _, tt := range []struct {
		name string
		cmd  *cobra.Command
		args map[string]string
	}{
		{name: "suite", cmd: createTestSuiteCmd(), args: map[string]string{"name": "suite", "description": "desc", "prompts": "p1,p2", "assertions": "contains", "output": filepath.Join(tmpDir, "suite.json")}},
		{name: "generate", cmd: generateTestsCmd(), args: map[string]string{"source": "src", "count": "2", "template": "tmpl", "output": filepath.Join(tmpDir, "gen.json")}},
		{name: "significance", cmd: significanceTestCmd(), args: map[string]string{"baseline": "old.json", "optimized": "new.json", "alpha": "0.01", "tests": "t-test", "output": filepath.Join(tmpDir, "sig.json")}},
		{name: "cross", cmd: crossValidateCmd(), args: map[string]string{"methods": "a,b", "folds": "3", "statistical": "true", "output": filepath.Join(tmpDir, "cross.json")}},
		{name: "ab", cmd: abTestCmd(), args: map[string]string{"group-a": "a", "group-b": "b", "metric": "score", "bayesian": "true", "output": filepath.Join(tmpDir, "ab.json")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			for k, v := range tt.args {
				if err := cmd.Flags().Set(k, v); err != nil {
					t.Fatal(err)
				}
			}
			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAdvancedTestCurrentResultsAndBaseline(t *testing.T) {
	provider := testLLMProvider{}
	cfg := &AdvancedTestConfig{
		Providers: []string{"mock:test-model"},
		Prompts:   []string{"a", "b"},
		Options:   map[string]interface{}{"temperature": 0.1},
	}
	results, err := generateCurrentResults(context.Background(), provider, cfg, llm.GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if results.Summary.TestCount != 2 || results.Summary.SuccessRate != 1 {
		t.Fatalf("current results = %#v", results.Summary)
	}
	if err := saveCurrentAsBaseline(&AdvancedTestConfig{}, filepath.Join(t.TempDir(), "x.json"), testCmd()); err == nil {
		t.Fatal("baseline without provider succeeded")
	}
}

type testLLMProvider struct{}

func (testLLMProvider) Name() string            { return "test" }
func (testLLMProvider) Model() string           { return "model" }
func (testLLMProvider) SupportsStreaming() bool { return false }
func (testLLMProvider) SupportsBatch() bool     { return false }
func (testLLMProvider) EvaluatePrompt(context.Context, string, map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return &promptfoo.ProviderResponse{Output: "ok"}, nil
}
func (testLLMProvider) Generate(context.Context, string, llm.GenerateOptions) (*llm.GenerateResponse, error) {
	return &llm.GenerateResponse{Text: "ok", PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2, Cost: 0.01}, nil
}
