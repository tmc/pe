package cgpt

import (
	"testing"
)

// TestSecurityValidation tests all the security validation functions
func TestSecurityValidation(t *testing.T) {
	// Test validateBackend
	t.Run("validateBackend", func(t *testing.T) {
		// Valid backends
		validBackends := []string{"openai", "anthropic", "googleai", "ollama"}
		for _, backend := range validBackends {
			if err := validateBackend(backend); err != nil {
				t.Errorf("validateBackend(%q) should be valid, got error: %v", backend, err)
			}
		}

		// Invalid backends
		invalidBackends := []string{"", "invalid", "$(whoami)", "openai; rm -rf /", "../../etc/passwd"}
		for _, backend := range invalidBackends {
			if err := validateBackend(backend); err == nil {
				t.Errorf("validateBackend(%q) should be invalid, got no error", backend)
			}
		}
	})

	// Test validateModel
	t.Run("validateModel", func(t *testing.T) {
		// Valid models
		validModels := []string{"gpt-4", "gpt-3.5-turbo", "claude-3", "gemini-pro", "llama-2-7b"}
		for _, model := range validModels {
			if err := validateModel(model); err != nil {
				t.Errorf("validateModel(%q) should be valid, got error: %v", model, err)
			}
		}

		// Invalid models
		invalidModels := []string{"", "$(whoami)", "gpt-4; rm -rf /", "model`whoami`", "model|ls", "model&echo"}
		for _, model := range invalidModels {
			if err := validateModel(model); err == nil {
				t.Errorf("validateModel(%q) should be invalid, got no error", model)
			}
		}
	})

	// Test sanitizePrompt
	t.Run("sanitizePrompt", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"Hello world", "Hello world"},
			{"What is $(whoami)?", "What is whoami)?"},
			{"Run `ls -la`", "Run ls -la"},
			{"Execute ${HOME}/script", "Execute HOME}/script"},
			{"Test && echo", "Test  echo"},
			{"Test || echo", "Test  echo"},
			{"Test ; echo", "Test  echo"},
			{"Test | echo", "Test  echo"},
			{"Test < file", "Test  file"},
			{"Test > file", "Test  file"},
			{"Test & background", "Test  background"},
			{"Normal text without issues", "Normal text without issues"},
			{"Text with\x00null bytes", "Text withnull bytes"},
		}

		for _, tc := range testCases {
			result := sanitizePrompt(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizePrompt(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		}
	})
}

// TestCommandInjectionPrevention tests that command injection attacks are prevented
func TestCommandInjectionPrevention(t *testing.T) {
	provider := DefaultProvider()

	// Test injection via backend
	t.Run("backend injection", func(t *testing.T) {
		provider.Backend = "openai; rm -rf /"
		_, _, err := provider.runCGPTCommand("test prompt", true)
		if err == nil {
			t.Error("Expected error for malicious backend, got nil")
		}
	})

	// Test injection via model
	t.Run("model injection", func(t *testing.T) {
		provider.Backend = "openai"
		provider.Model = "gpt-4`whoami`"
		_, _, err := provider.runCGPTCommand("test prompt", true)
		if err == nil {
			t.Error("Expected error for malicious model, got nil")
		}
	})

	// Test injection via prompt
	t.Run("prompt injection", func(t *testing.T) {
		provider.Backend = "openai"
		provider.Model = "gpt-4"
		
		// This should work but the dangerous characters should be sanitized
		maliciousPrompt := "Generate code that runs $(whoami) and `ls -la`"
		result, _, err := provider.runCGPTCommand(maliciousPrompt, true)
		if err != nil {
			t.Errorf("Expected no error for prompt sanitization, got: %v", err)
		}
		
		// Check that the result indicates it was sanitized (dry run mode)
		if result != "Dry run - no actual execution" {
			t.Errorf("Expected dry run result, got: %q", result)
		}
	})
}

// TestConfigValidation tests that configuration values are validated
func TestConfigValidation(t *testing.T) {
	provider := DefaultProvider()

	// Test malicious config via vars
	t.Run("malicious config", func(t *testing.T) {
		maliciousVars := map[string]interface{}{
			"backend": "openai; rm -rf /",
			"model":   "gpt-4`whoami`",
			"temperature": -1.0, // Invalid temperature
			"max_tokens": 999999, // Invalid max tokens
		}

		originalBackend := provider.Backend
		originalModel := provider.Model
		originalTemp := provider.Temperature
		originalMaxTokens := provider.MaxTokens

		provider.ApplyConfigFromVars(maliciousVars)

		// Check that malicious values were rejected
		if provider.Backend != originalBackend {
			t.Errorf("Backend should not have changed from %q, got %q", originalBackend, provider.Backend)
		}
		if provider.Model != originalModel {
			t.Errorf("Model should not have changed from %q, got %q", originalModel, provider.Model)
		}
		if provider.Temperature != originalTemp {
			t.Errorf("Temperature should not have changed from %f, got %f", originalTemp, provider.Temperature)
		}
		if provider.MaxTokens != originalMaxTokens {
			t.Errorf("MaxTokens should not have changed from %d, got %d", originalMaxTokens, provider.MaxTokens)
		}
	})

	// Test valid config
	t.Run("valid config", func(t *testing.T) {
		validVars := map[string]interface{}{
			"backend": "anthropic",
			"model":   "claude-3",
			"temperature": 0.5,
			"max_tokens": 1000,
		}

		provider.ApplyConfigFromVars(validVars)

		// Check that valid values were applied
		if provider.Backend != "anthropic" {
			t.Errorf("Expected backend 'anthropic', got %q", provider.Backend)
		}
		if provider.Model != "claude-3" {
			t.Errorf("Expected model 'claude-3', got %q", provider.Model)
		}
		if provider.Temperature != 0.5 {
			t.Errorf("Expected temperature 0.5, got %f", provider.Temperature)
		}
		if provider.MaxTokens != 1000 {
			t.Errorf("Expected max_tokens 1000, got %d", provider.MaxTokens)
		}
	})
}

// TestGetAllowedBackends tests the helper function
func TestGetAllowedBackends(t *testing.T) {
	backends := getAllowedBackends()
	if len(backends) == 0 {
		t.Error("getAllowedBackends() should return non-empty slice")
	}
	
	// Check that openai is in the list
	found := false
	for _, backend := range backends {
		if backend == "openai" {
			found = true
			break
		}
	}
	if !found {
		t.Error("getAllowedBackends() should include 'openai'")
	}
}