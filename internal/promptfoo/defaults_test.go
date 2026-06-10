package promptfoo

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func TestApplyDefaults(t *testing.T) {
	cfg := Config{
		DefaultTest: &TestDefaults{
			Vars:   map[string]interface{}{"lang": "en", "tone": "formal"},
			Assert: []Assertion{{Type: "contains", Value: "hello"}},
		},
	}
	tests := []TestCase{
		{
			Vars:   map[string]interface{}{"tone": "casual"}, // overrides default
			Assert: []Assertion{{Type: "regex", Value: "hi"}},
		},
	}

	got := cfg.ApplyDefaults(tests)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	tc := got[0]
	if tc.Vars["lang"] != "en" {
		t.Fatalf("lang = %v, want en (from default)", tc.Vars["lang"])
	}
	if tc.Vars["tone"] != "casual" {
		t.Fatalf("tone = %v, want casual (test overrides default)", tc.Vars["tone"])
	}
	// Default assert is prepended, test assert follows.
	if len(tc.Assert) != 2 || tc.Assert[0].Type != "contains" || tc.Assert[1].Type != "regex" {
		t.Fatalf("asserts = %+v, want [contains, regex]", tc.Assert)
	}
	// Original input is not mutated.
	if len(tests[0].Assert) != 1 {
		t.Fatalf("input test mutated: %d asserts", len(tests[0].Assert))
	}
}

func TestApplyDefaultsDisableDefaultAsserts(t *testing.T) {
	cfg := Config{
		DefaultTest: &TestDefaults{
			Vars:   map[string]interface{}{"lang": "en"},
			Assert: []Assertion{{Type: "contains", Value: "hello"}},
		},
	}
	tests := []TestCase{
		{
			Assert:                []Assertion{{Type: "regex", Value: "hi"}},
			DisableDefaultAsserts: true,
		},
	}
	got := cfg.ApplyDefaults(tests)
	// Default vars still apply, but default asserts are skipped.
	if got[0].Vars["lang"] != "en" {
		t.Fatalf("lang = %v, want en even with disableDefaultAsserts", got[0].Vars["lang"])
	}
	if len(got[0].Assert) != 1 || got[0].Assert[0].Type != "regex" {
		t.Fatalf("asserts = %+v, want only [regex]", got[0].Assert)
	}
}

func TestApplyDefaultsNoDefaultTest(t *testing.T) {
	cfg := Config{}
	tests := []TestCase{{Vars: map[string]interface{}{"a": 1}}}
	got := cfg.ApplyDefaults(tests)
	if len(got) != 1 || got[0].Vars["a"] != 1 {
		t.Fatalf("ApplyDefaults with no defaultTest should pass tests through: %+v", got)
	}
}

func TestConfigParsesDefaultTest(t *testing.T) {
	input := []byte(`
prompts: ["{{input}}"]
providers: [openai:gpt-4o-mini]
defaultTest:
  vars:
    lang: en
  assert:
    - type: contains
      value: hello
tests:
  - vars: { input: hi }
    disableDefaultAsserts: true
    assert:
      - type: regex
        value: hi
`)
	var cfg Config
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if cfg.DefaultTest == nil || cfg.DefaultTest.Vars["lang"] != "en" || len(cfg.DefaultTest.Assert) != 1 {
		t.Fatalf("defaultTest not parsed: %+v", cfg.DefaultTest)
	}
	if !cfg.Tests[0].DisableDefaultAsserts {
		t.Fatal("disableDefaultAsserts not parsed")
	}
	merged := cfg.ApplyDefaults(cfg.Tests)
	if merged[0].Vars["lang"] != "en" {
		t.Fatalf("merged lang = %v, want en", merged[0].Vars["lang"])
	}
	if len(merged[0].Assert) != 1 || merged[0].Assert[0].Type != "regex" {
		t.Fatalf("merged asserts = %+v, want only [regex] (default asserts disabled)", merged[0].Assert)
	}
}
