package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/tmc/pe/internal/template"
)

func TestProcessPrompt(t *testing.T) {
	// Test simple variable substitution
	t.Run("SimpleVariables", func(t *testing.T) {
		promptTemplate := "What is the capital of {{country}}?"
		vars := map[string]interface{}{
			"country": "France",
		}
		
		expected := "What is the capital of France?"
		result, err := template.ProcessPrompt(promptTemplate, vars)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
	
	// Test multiple variables
	t.Run("MultipleVariables", func(t *testing.T) {
		promptTemplate := "{{greeting}}, my name is {{name}}. I am {{age}} years old."
		vars := map[string]interface{}{
			"greeting": "Hello",
			"name":     "Alice",
			"age":      30,
		}
		
		expected := "Hello, my name is Alice. I am 30 years old."
		result, err := template.ProcessPrompt(promptTemplate, vars)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
	
	// Test missing variable
	t.Run("MissingVariable", func(t *testing.T) {
		promptTemplate := "What is the capital of {{country}}?"
		vars := map[string]interface{}{
			"city": "Paris", // Wrong variable name
		}
		
		_, err := template.ProcessPrompt(promptTemplate, vars)
		
		if err == nil {
			t.Error("Expected error for missing variable, got nil")
		}
		
		// Check if it's the expected error type
		_, ok := err.(template.ErrMissingVariable)
		if !ok {
			t.Errorf("Expected ErrMissingVariable, got %T", err)
		}
	})
	
	// Test complex variable types
	t.Run("ComplexVariableTypes", func(t *testing.T) {
		promptTemplate := "Data: {{data}}"
		vars := map[string]interface{}{
			"data": map[string]interface{}{
				"key": "value",
				"nested": map[string]interface{}{
					"inner": 123,
				},
			},
		}
		
		result, err := template.ProcessPrompt(promptTemplate, vars)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if !contains(result, "key") || !contains(result, "value") || !contains(result, "123") {
			t.Errorf("Expected result to contain complex data elements, got: %s", result)
		}
	})
}

func TestProcessPromptWithFiles(t *testing.T) {
	// Create temporary directory and files
	tmpDir := t.TempDir()
	
	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "This is test content from a file."
	err := ioutil.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test file inclusion
	t.Run("FileInclusion", func(t *testing.T) {
		promptTemplate := "File content: {{file \"" + testFile + "\"}}"
		vars := map[string]interface{}{}
		
		expected := "File content: " + testContent
		result, err := template.ProcessPromptWithFiles(promptTemplate, vars)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
	
	// Test file inclusion with variables
	t.Run("FileInclusionWithVariables", func(t *testing.T) {
		promptTemplate := "{{greeting}}, here is a file: {{file \"" + testFile + "\"}}"
		vars := map[string]interface{}{
			"greeting": "Hello",
		}
		
		expected := "Hello, here is a file: " + testContent
		result, err := template.ProcessPromptWithFiles(promptTemplate, vars)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
	
	// Test non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tmpDir, "nonexistent.txt")
		promptTemplate := "File content: {{file \"" + nonExistentFile + "\"}}"
		vars := map[string]interface{}{}
		
		_, err := template.ProcessPromptWithFiles(promptTemplate, vars)
		
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

func TestProcessMany(t *testing.T) {
	// Test processing multiple prompts
	t.Run("MultiplePrompts", func(t *testing.T) {
		promptTemplates := []string{
			"What is the capital of {{country}}?",
			"{{greeting}}, my name is {{name}}.",
		}
		
		vars := map[string]interface{}{
			"country":  "France",
			"greeting": "Hello",
			"name":     "Alice",
		}
		
		expected := []string{
			"What is the capital of France?",
			"Hello, my name is Alice.",
		}
		
		results, err := template.ProcessMany(promptTemplates, vars)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, expectedResult := range expected {
			if results[i] != expectedResult {
				t.Errorf("Result %d: expected %q, got %q", i, expectedResult, results[i])
			}
		}
	})
	
	// Test error propagation
	t.Run("ErrorPropagation", func(t *testing.T) {
		promptTemplates := []string{
			"Valid prompt with {{var}}",
			"Invalid prompt with {{missing}}",
		}
		
		vars := map[string]interface{}{
			"var": "value",
		}
		
		_, err := template.ProcessMany(promptTemplates, vars)
		
		if err == nil {
			t.Error("Expected error for missing variable, got nil")
		}
	})
}

func TestProcessManyWithFiles(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "This is test content from a file."
	err := ioutil.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test processing multiple prompts with files
	t.Run("MultiplePromptsWithFiles", func(t *testing.T) {
		promptTemplates := []string{
			"What is the capital of {{country}}?",
			"File content: {{file \"" + testFile + "\"}}",
		}
		
		vars := map[string]interface{}{
			"country": "France",
		}
		
		expected := []string{
			"What is the capital of France?",
			"File content: " + testContent,
		}
		
		results, err := template.ProcessManyWithFiles(promptTemplates, vars)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if len(results) != len(expected) {
			t.Fatalf("Expected %d results, got %d", len(expected), len(results))
		}
		
		for i, expectedResult := range expected {
			if results[i] != expectedResult {
				t.Errorf("Result %d: expected %q, got %q", i, expectedResult, results[i])
			}
		}
	})
	
	// Test error propagation
	t.Run("ErrorPropagation", func(t *testing.T) {
		promptTemplates := []string{
			"Valid prompt with {{var}}",
			"File: {{file \"" + filepath.Join(tmpDir, "nonexistent.txt") + "\"}}",
		}
		
		vars := map[string]interface{}{
			"var": "value",
		}
		
		_, err := template.ProcessManyWithFiles(promptTemplates, vars)
		
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}