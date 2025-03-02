package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tmc/pe/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary YAML file
	yamlData := []byte(`
prompts:
  - "What is the capital of {{country}}?"
providers:
  - "openai:gpt-4"
tests:
  - vars:
      country: "France"
    assert:
      - type: "contains"
        value: "Paris"
`)

	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(yamlFile, yamlData, 0644); err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Create a temporary JSON file
	jsonData := []byte(`{
  "prompts": ["What is the capital of {{country}}?"],
  "providers": ["openai:gpt-4"],
  "tests": [
    {
      "vars": {"country": "France"},
      "assert": [{"type": "contains", "value": "Paris"}]
    }
  ]
}`)

	jsonFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Test loading from YAML
	yamlConfig, err := config.LoadConfig(yamlFile)
	if err != nil {
		t.Errorf("Failed to load YAML config: %v", err)
	}
	if yamlConfig == nil {
		t.Fatalf("YAML config is nil")
	}
	if len(yamlConfig.Prompts) != 1 || yamlConfig.Prompts[0] != "What is the capital of {{country}}?" {
		t.Errorf("YAML config has incorrect prompts: %v", yamlConfig.Prompts)
	}

	// Test loading from JSON
	jsonConfig, err := config.LoadConfig(jsonFile)
	if err != nil {
		t.Errorf("Failed to load JSON config: %v", err)
	}
	if jsonConfig == nil {
		t.Fatalf("JSON config is nil")
	}
	if len(jsonConfig.Prompts) != 1 || jsonConfig.Prompts[0] != "What is the capital of {{country}}?" {
		t.Errorf("JSON config has incorrect prompts: %v", jsonConfig.Prompts)
	}
}

func TestConfigValidation(t *testing.T) {
	// Valid config should pass validation
	validConfig := &config.Config{
		Prompts:   []string{"test prompt"},
		Providers: []string{"test-provider"},
		Tests: []map[string]interface{}{
			{
				"vars": map[string]interface{}{"var1": "value1"},
				"assert": []interface{}{
					map[string]interface{}{"type": "contains", "value": "expected"},
				},
			},
		},
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}

	// Test missing prompts
	invalidConfig1 := &config.Config{
		Prompts:   []string{},
		Providers: []string{"test-provider"},
		Tests: []map[string]interface{}{
			{"vars": map[string]interface{}{"var1": "value1"}, "assert": []interface{}{map[string]interface{}{"type": "contains", "value": "expected"}}},
		},
	}
	if err := invalidConfig1.Validate(); err == nil {
		t.Error("Expected validation error for missing prompts, but got none")
	}

	// Test missing providers
	invalidConfig2 := &config.Config{
		Prompts:   []string{"test prompt"},
		Providers: []string{},
		Tests: []map[string]interface{}{
			{"vars": map[string]interface{}{"var1": "value1"}, "assert": []interface{}{map[string]interface{}{"type": "contains", "value": "expected"}}},
		},
	}
	if err := invalidConfig2.Validate(); err == nil {
		t.Error("Expected validation error for missing providers, but got none")
	}

	// Test missing tests
	invalidConfig3 := &config.Config{
		Prompts:   []string{"test prompt"},
		Providers: []string{"test-provider"},
		Tests:     []map[string]interface{}{},
	}
	if err := invalidConfig3.Validate(); err == nil {
		t.Error("Expected validation error for missing tests, but got none")
	}

	// Test missing vars in test
	invalidConfig4 := &config.Config{
		Prompts:   []string{"test prompt"},
		Providers: []string{"test-provider"},
		Tests: []map[string]interface{}{
			{"assert": []interface{}{map[string]interface{}{"type": "contains", "value": "expected"}}},
		},
	}
	if err := invalidConfig4.Validate(); err == nil {
		t.Error("Expected validation error for test missing vars, but got none")
	}

	// Test missing assert in test
	invalidConfig5 := &config.Config{
		Prompts:   []string{"test prompt"},
		Providers: []string{"test-provider"},
		Tests: []map[string]interface{}{
			{"vars": map[string]interface{}{"var1": "value1"}},
		},
	}
	if err := invalidConfig5.Validate(); err == nil {
		t.Error("Expected validation error for test missing assert, but got none")
	}
}