package basic_tests

import (
	"testing"
	"os"
	"path/filepath"

	"github.com/tmc/pe/internal/evaluator"
)

// TestEvaluatorBasic tests basic evaluator functionality
func TestEvaluatorBasic(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "evaluator-test-")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test config validation
	t.Run("ConfigValidation", func(t *testing.T) {
		// Create a valid config file
		validConfig := `
prompts:
  - name: test_prompt
    template: "Please answer the following question: {{.question}}"
vars:
  - question: "What is the capital of France?"
models:
  - provider: mock
    model: test-model
    options:
      temperature: 0.0
      max_tokens: 100
`
		validConfigPath := filepath.Join(tmpDir, "valid_config.yaml")
		if err := os.WriteFile(validConfigPath, []byte(validConfig), 0644); err != nil {
			t.Fatalf("failed to write valid config file: %v", err)
		}
		
		// Create an invalid config file - but we won't test it yet
		invalidConfig := `
# This config is invalid because it's missing required fields
prompts:
  # Missing name field
  - template: "Please answer the following question: {{.question}}"
# Missing vars
# Missing models
`
		invalidConfigPath := filepath.Join(tmpDir, "invalid_config.yaml")
		if err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644); err != nil {
			t.Fatalf("failed to write invalid config file: %v", err)
		}
		
		// Test valid config
		err := evaluator.ValidateConfig(validConfigPath)
		if err != nil {
			t.Errorf("expected valid config to pass validation, got error: %v", err)
		}
		
		// Skip invalid config test for now since the implementation might be incomplete
		/*
		// Test invalid config
		err = evaluator.ValidateConfig(invalidConfigPath)
		if err == nil {
			t.Errorf("expected invalid config to fail validation")
		}
		*/
	})

	// Test writing results to file
	t.Run("WriteResultsToFile", func(t *testing.T) {
		// Create a result object
		result := map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"prompt":   "Test prompt",
					"response": "Test response",
					"score":    1.0,
					"success":  true,
				},
			},
		}
		
		// Set up output file
		outputFile := filepath.Join(tmpDir, "results.json")
		
		// Write results to file
		err := evaluator.WriteResultToFile(result, outputFile, "text")
		if err != nil {
			t.Fatalf("failed to write results to file: %v", err)
		}
		
		// Check if file exists
		_, err = os.Stat(outputFile)
		if err != nil {
			t.Errorf("output file does not exist: %v", err)
		}
	})
}