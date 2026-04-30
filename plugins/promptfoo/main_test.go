package main

import "testing"

func TestConvertToPromptfooNormalizesProvidersAndTests(t *testing.T) {
	in := map[string]interface{}{
		"description": "test config",
		"prompts":     []interface{}{"Say hi"},
		"providers": []interface{}{
			map[string]interface{}{
				"name": "openai",
				"config": map[string]interface{}{
					"model":       "gpt-4o-mini",
					"temperature": 0.1,
				},
			},
			"anthropic:claude-3-haiku",
		},
		"default_test": map[string]interface{}{
			"assertions": []interface{}{
				map[string]interface{}{
					"type":  "not_contains",
					"value": "error",
				},
			},
		},
		"tests": []interface{}{
			map[string]interface{}{
				"variables": map[string]interface{}{"topic": "Go"},
				"assertions": []interface{}{
					map[string]interface{}{
						"type":  "llm_rubric",
						"value": "Mentions goroutines",
					},
				},
			},
		},
	}

	got := convertToPromptfoo(in)

	providers, ok := got["providers"].([]interface{})
	if !ok || len(providers) != 2 {
		t.Fatalf("providers = %#v, want 2 entries", got["providers"])
	}

	firstProvider, ok := providers[0].(map[string]interface{})
	if !ok {
		t.Fatalf("first provider = %#v, want map", providers[0])
	}
	if firstProvider["apiProvider"] != "openai" {
		t.Fatalf("apiProvider = %#v, want openai", firstProvider["apiProvider"])
	}
	if firstProvider["model"] != "gpt-4o-mini" {
		t.Fatalf("model = %#v, want gpt-4o-mini", firstProvider["model"])
	}

	tests, ok := got["tests"].([]interface{})
	if !ok || len(tests) != 1 {
		t.Fatalf("tests = %#v, want 1 entry", got["tests"])
	}

	testCase, ok := tests[0].(map[string]interface{})
	if !ok {
		t.Fatalf("test case = %#v, want map", tests[0])
	}
	if _, ok := testCase["vars"]; !ok {
		t.Fatalf("vars missing from %#v", testCase)
	}

	asserts, ok := testCase["assert"].([]interface{})
	if !ok || len(asserts) != 1 {
		t.Fatalf("assert = %#v, want 1 entry", testCase["assert"])
	}

	assertion, ok := asserts[0].(map[string]interface{})
	if !ok {
		t.Fatalf("assertion = %#v, want map", asserts[0])
	}
	if assertion["type"] != "llm-rubric" {
		t.Fatalf("assertion type = %#v, want llm-rubric", assertion["type"])
	}

	defaults, ok := got["defaultTest"].(map[string]interface{})
	if !ok {
		t.Fatalf("defaultTest = %#v, want map", got["defaultTest"])
	}
	defaultAsserts, ok := defaults["assert"].([]interface{})
	if !ok || len(defaultAsserts) != 1 {
		t.Fatalf("defaultTest.assert = %#v, want 1 entry", defaults["assert"])
	}
}

func TestConvertFromPromptfooNormalizesProviders(t *testing.T) {
	in := map[string]interface{}{
		"description": "promptfoo config",
		"providers": []interface{}{
			map[string]interface{}{
				"id":          "prod",
				"apiProvider": "openai",
				"model":       "gpt-4.1",
				"temperature": 0.2,
			},
			"anthropic:claude-3-haiku",
		},
		"tests": []interface{}{
			map[string]interface{}{
				"vars": map[string]interface{}{"question": "hi"},
				"assert": []interface{}{
					map[string]interface{}{
						"type":  "not-contains",
						"value": "error",
					},
				},
			},
		},
		"defaultTest": map[string]interface{}{
			"assert": []interface{}{
				map[string]interface{}{
					"type":  "contains",
					"value": "hello",
				},
			},
		},
	}

	got := convertFromPromptfoo(in)

	if got["version"] != "1.0" {
		t.Fatalf("version = %#v, want 1.0", got["version"])
	}

	providers, ok := got["providers"].([]interface{})
	if !ok || len(providers) != 2 {
		t.Fatalf("providers = %#v, want 2 entries", got["providers"])
	}

	firstProvider, ok := providers[0].(map[string]interface{})
	if !ok {
		t.Fatalf("first provider = %#v, want map", providers[0])
	}
	if firstProvider["name"] != "openai" {
		t.Fatalf("name = %#v, want openai", firstProvider["name"])
	}
	if firstProvider["id"] != "prod" {
		t.Fatalf("id = %#v, want prod", firstProvider["id"])
	}

	config, ok := firstProvider["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("config = %#v, want map", firstProvider["config"])
	}
	if config["model"] != "gpt-4.1" {
		t.Fatalf("model = %#v, want gpt-4.1", config["model"])
	}

	tests, ok := got["tests"].([]interface{})
	if !ok || len(tests) != 1 {
		t.Fatalf("tests = %#v, want 1 entry", got["tests"])
	}

	testCase, ok := tests[0].(map[string]interface{})
	if !ok {
		t.Fatalf("test case = %#v, want map", tests[0])
	}
	if _, ok := testCase["vars"]; !ok {
		t.Fatalf("vars missing from %#v", testCase)
	}
	if _, ok := testCase["assert"]; !ok {
		t.Fatalf("assert missing from %#v", testCase)
	}
}
