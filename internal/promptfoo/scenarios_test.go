package promptfoo

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func TestExpandScenarios(t *testing.T) {
	cfg := Config{
		Tests: []TestCase{
			{Vars: map[string]interface{}{"input": "base"}},
		},
		Scenarios: []Scenario{
			{
				Config: []TestCase{
					{Vars: map[string]interface{}{"language": "Spanish"}, Assert: []Assertion{{Type: "is-json"}}},
					{Vars: map[string]interface{}{"language": "French"}},
				},
				Tests: []TestCase{
					{Vars: map[string]interface{}{"input": "Hello"}, Assert: []Assertion{{Type: "contains", Value: "x"}}},
					{Vars: map[string]interface{}{"input": "World"}},
				},
			},
		},
	}

	got := cfg.ExpandScenarios()
	// 1 standalone test + (2 config x 2 tests) = 5.
	if len(got) != 5 {
		t.Fatalf("ExpandScenarios() len = %d, want 5", len(got))
	}

	// The standalone test is preserved first.
	if got[0].Vars["input"] != "base" {
		t.Fatalf("got[0].input = %v, want base", got[0].Vars["input"])
	}

	// First product: Spanish config x Hello test. Test vars override; asserts
	// are config-then-test.
	first := got[1]
	if first.Vars["language"] != "Spanish" || first.Vars["input"] != "Hello" {
		t.Fatalf("first product vars = %v, want language=Spanish input=Hello", first.Vars)
	}
	if len(first.Assert) != 2 || first.Assert[0].Type != "is-json" || first.Assert[1].Type != "contains" {
		t.Fatalf("first product asserts = %+v, want [is-json, contains]", first.Assert)
	}
}

func TestMergeTestCasesVarsOverride(t *testing.T) {
	base := TestCase{Vars: map[string]interface{}{"a": 1, "b": 2}}
	override := TestCase{Vars: map[string]interface{}{"b": 99, "c": 3}}
	got := mergeTestCases(base, override)
	if got.Vars["a"] != 1 || got.Vars["b"] != 99 || got.Vars["c"] != 3 {
		t.Fatalf("merged vars = %v, want a=1 b=99 c=3", got.Vars)
	}
	// Inputs are not mutated.
	if base.Vars["b"] != 2 {
		t.Fatalf("base mutated: b=%v, want 2", base.Vars["b"])
	}
}

// TestConfigParsesPromptfooKeys confirms an unmodified promptfoo config that
// uses scenarios, derivedMetrics, outputPath, and redteam loads through the
// same loader pe eval uses (sigs.k8s.io/yaml, JSON-tag based).
func TestConfigParsesPromptfooKeys(t *testing.T) {
	input := []byte(`
description: compat check
prompts:
  - "Translate {{input}} to {{language}}"
providers:
  - openai:gpt-4o-mini
outputPath: out.json
scenarios:
  - description: greetings
    config:
      - vars: { language: Spanish }
    tests:
      - vars: { input: Hello }
        assert:
          - type: is-json
derivedMetrics:
  - name: DoubleScore
    value: "Score * 2"
redteam:
  purpose: "test the assistant"
  numTests: 5
  plugins:
    - harmful
    - pii
  strategies:
    - jailbreak
`)
	var cfg Config
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}
	if cfg.OutputPath != "out.json" {
		t.Fatalf("OutputPath = %q, want out.json", cfg.OutputPath)
	}
	if len(cfg.Scenarios) != 1 || len(cfg.Scenarios[0].Config) != 1 || len(cfg.Scenarios[0].Tests) != 1 {
		t.Fatalf("scenarios not parsed: %+v", cfg.Scenarios)
	}
	if len(cfg.DerivedMetrics) != 1 || cfg.DerivedMetrics[0].Name != "DoubleScore" {
		t.Fatalf("derivedMetrics not parsed: %+v", cfg.DerivedMetrics)
	}
	if cfg.Redteam == nil || cfg.Redteam.NumTests != 5 || len(cfg.Redteam.Plugins) != 2 || len(cfg.Redteam.Strategies) != 1 {
		t.Fatalf("redteam not parsed: %+v", cfg.Redteam)
	}
	// Scenario expansion yields the one product test.
	expanded := cfg.ExpandScenarios()
	if len(expanded) != 1 {
		t.Fatalf("ExpandScenarios() len = %d, want 1", len(expanded))
	}
	if expanded[0].Vars["language"] != "Spanish" || expanded[0].Vars["input"] != "Hello" {
		t.Fatalf("expanded vars = %v", expanded[0].Vars)
	}
}
