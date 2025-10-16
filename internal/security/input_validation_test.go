package security

import (
	"strings"
	"testing"
)

// InputSanitizer provides methods to sanitize and validate user inputs
type InputSanitizer struct {
	maxPromptLength int
	maxFileSize     int64
	allowedChars    string
}

// NewInputSanitizer creates a new input sanitizer with default settings
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{
		maxPromptLength: 10000, // 10KB max prompt
		maxFileSize:     1024 * 1024 * 10, // 10MB max file
		allowedChars:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.,!?;:()[]{}\"'- \n\r\t",
	}
}

// SanitizePrompt cleans user input prompts
func (s *InputSanitizer) SanitizePrompt(input string) (string, error) {
	// Check length
	if len(input) > s.maxPromptLength {
		return "", NewSecurityError("prompt_too_long", "Prompt exceeds maximum length")
	}

	// Remove null bytes
	cleaned := strings.ReplaceAll(input, "\x00", "")

	// Remove other control characters except allowed ones
	var result strings.Builder
	for _, r := range cleaned {
		if r < 32 {
			// Allow specific whitespace characters
			if r == '\n' || r == '\r' || r == '\t' {
				result.WriteRune(r)
			}
			// Skip other control characters
		} else if r == 127 {
			// Skip DEL character
		} else {
			result.WriteRune(r)
		}
	}

	return result.String(), nil
}

// ValidateFilePath validates file paths to prevent directory traversal
func (s *InputSanitizer) ValidateFilePath(path string) error {
	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return NewSecurityError("invalid_path", "File path contains null bytes")
	}

	// Check for Windows drive paths (e.g., C:\, D:\, etc.)
	// These are absolute paths and potentially dangerous
	if len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
		// Single letter followed by colon indicates Windows drive
		if (path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z') {
			return NewSecurityError("path_traversal", "File path contains traversal patterns")
		}
	}

	// Check for directory traversal patterns
	dangerous := []string{"../", "..\\", "/..", "\\..", "..."}
	for _, pattern := range dangerous {
		if strings.Contains(path, pattern) {
			return NewSecurityError("path_traversal", "File path contains traversal patterns")
		}
	}

	// Check for absolute paths that might be dangerous
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "/tmp/") && !strings.HasPrefix(path, "/var/tmp/") {
		return NewSecurityError("absolute_path", "Absolute paths outside temp directories not allowed")
	}

	return nil
}

// ValidateAPIKey validates API key format and characteristics
func (s *InputSanitizer) ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return NewSecurityError("empty_api_key", "API key cannot be empty")
	}

	if len(apiKey) < 10 {
		return NewSecurityError("api_key_too_short", "API key too short")
	}

	if len(apiKey) > 500 {
		return NewSecurityError("api_key_too_long", "API key too long")
	}

	// Check for suspicious patterns
	if strings.Contains(apiKey, " ") {
		return NewSecurityError("invalid_api_key", "API key contains spaces")
	}

	if strings.Contains(apiKey, "\n") || strings.Contains(apiKey, "\r") {
		return NewSecurityError("invalid_api_key", "API key contains newlines")
	}

	return nil
}

// SecurityError represents a security-related error
type SecurityError struct {
	Code    string
	Message string
}

func (e *SecurityError) Error() string {
	return e.Message
}

// NewSecurityError creates a new security error
func NewSecurityError(code, message string) *SecurityError {
	return &SecurityError{
		Code:    code,
		Message: message,
	}
}

// Test Input Sanitization

func TestSanitizePrompt(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "normal_text",
			input:    "Hello, world! How are you?",
			expected: "Hello, world! How are you?",
			wantErr:  false,
		},
		{
			name:     "with_newlines",
			input:    "Line 1\nLine 2\rLine 3",
			expected: "Line 1\nLine 2\rLine 3",
			wantErr:  false,
		},
		{
			name:     "with_tabs",
			input:    "Column1\tColumn2\tColumn3",
			expected: "Column1\tColumn2\tColumn3",
			wantErr:  false,
		},
		{
			name:     "with_null_bytes",
			input:    "Hello\x00World",
			expected: "HelloWorld",
			wantErr:  false,
		},
		{
			name:     "with_control_chars",
			input:    "Hello\x01\x02\x03World",
			expected: "HelloWorld",
			wantErr:  false,
		},
		{
			name:     "with_del_char",
			input:    "Hello\x7fWorld",
			expected: "HelloWorld",
			wantErr:  false,
		},
		{
			name:    "too_long_prompt",
			input:   strings.Repeat("a", 15000),
			wantErr: true,
		},
		{
			name:     "unicode_chars",
			input:    "Hello 世界! 🌍",
			expected: "Hello 世界! 🌍",
			wantErr:  false,
		},
		{
			name:     "empty_input",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "special_chars",
			input:    `!"#$%&'()*+,-./:;<=>?@[\]^_{|}~`,
			expected: `!"#$%&'()*+,-./:;<=>?@[\]^_{|}~`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.SanitizePrompt(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name     string
		path     string
		wantErr  bool
		errorCode string
	}{
		{
			name:    "normal_file",
			path:    "config.yaml",
			wantErr: false,
		},
		{
			name:    "with_subdirectory",
			path:    "configs/test.yaml",
			wantErr: false,
		},
		{
			name:      "directory_traversal_unix",
			path:      "../../../etc/passwd",
			wantErr:   true,
			errorCode: "path_traversal",
		},
		{
			name:      "directory_traversal_windows",
			path:      "..\\..\\windows\\system32",
			wantErr:   true,
			errorCode: "path_traversal",
		},
		{
			name:      "hidden_traversal",
			path:      "configs/./../../secret.txt",
			wantErr:   true,
			errorCode: "path_traversal",
		},
		{
			name:      "null_byte_injection",
			path:      "config.txt\x00.exe",
			wantErr:   true,
			errorCode: "invalid_path",
		},
		{
			name:    "temp_directory_allowed",
			path:    "/tmp/pe_temp_file.txt",
			wantErr: false,
		},
		{
			name:    "var_tmp_allowed",
			path:    "/var/tmp/pe_cache.dat",
			wantErr: false,
		},
		{
			name:      "absolute_path_not_allowed",
			path:      "/etc/passwd",
			wantErr:   true,
			errorCode: "absolute_path",
		},
		{
			name:      "windows_drive",
			path:      "C:\\Windows\\System32\\config",
			wantErr:   true,
			errorCode: "path_traversal",
		},
		{
			name:      "triple_dot_traversal",
			path:      "...config",
			wantErr:   true,
			errorCode: "path_traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.ValidateFilePath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if secErr, ok := err.(*SecurityError); ok {
					if tt.errorCode != "" && secErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, secErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name      string
		apiKey    string
		wantErr   bool
		errorCode string
	}{
		{
			name:    "valid_openai_key",
			apiKey:  "sk-1234567890abcdef1234567890abcdef12345678",
			wantErr: false,
		},
		{
			name:    "valid_anthropic_key",
			apiKey:  "ant-api03-1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:      "empty_key",
			apiKey:    "",
			wantErr:   true,
			errorCode: "empty_api_key",
		},
		{
			name:      "too_short_key",
			apiKey:    "sk-123",
			wantErr:   true,
			errorCode: "api_key_too_short",
		},
		{
			name:      "too_long_key",
			apiKey:    strings.Repeat("a", 600),
			wantErr:   true,
			errorCode: "api_key_too_long",
		},
		{
			name:      "key_with_spaces",
			apiKey:    "sk-1234567890 abcdef1234567890",
			wantErr:   true,
			errorCode: "invalid_api_key",
		},
		{
			name:      "key_with_newlines",
			apiKey:    "sk-1234567890\nabcdef1234567890",
			wantErr:   true,
			errorCode: "invalid_api_key",
		},
		{
			name:      "key_with_carriage_return",
			apiKey:    "sk-1234567890\rabcdef1234567890",
			wantErr:   true,
			errorCode: "invalid_api_key",
		},
		{
			name:    "key_with_special_chars",
			apiKey:  "sk-1234567890_abcdef.1234567890-abcdef",
			wantErr: false,
		},
		{
			name:    "minimal_valid_key",
			apiKey:  "1234567890",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.ValidateAPIKey(tt.apiKey)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if secErr, ok := err.(*SecurityError); ok {
					if tt.errorCode != "" && secErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, secErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test Security Error Handling

func TestSecurityError(t *testing.T) {
	err := NewSecurityError("test_code", "Test message")

	if err.Code != "test_code" {
		t.Errorf("Expected code 'test_code', got %s", err.Code)
	}

	if err.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", err.Message)
	}

	if err.Error() != "Test message" {
		t.Errorf("Expected Error() to return 'Test message', got %s", err.Error())
	}
}

// Benchmark Tests

func BenchmarkSanitizePrompt(b *testing.B) {
	sanitizer := NewInputSanitizer()
	input := "Hello, world! This is a test prompt with some special characters: @#$%^&*()."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sanitizer.SanitizePrompt(input)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkValidateFilePath(b *testing.B) {
	sanitizer := NewInputSanitizer()
	path := "configs/test/prompt.yaml"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sanitizer.ValidateFilePath(path)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkValidateAPIKey(b *testing.B) {
	sanitizer := NewInputSanitizer()
	apiKey := "sk-1234567890abcdef1234567890abcdef12345678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sanitizer.ValidateAPIKey(apiKey)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// Edge Case Tests

func TestInputSanitization_EdgeCases(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name  string
		test  func(t *testing.T)
	}{
		{
			name: "very_large_input",
			test: func(t *testing.T) {
				// Test with input at the boundary
				input := strings.Repeat("a", sanitizer.maxPromptLength)
				_, err := sanitizer.SanitizePrompt(input)
				if err != nil {
					t.Errorf("Should accept input at max length: %v", err)
				}

				// Test with input over the boundary
				input = strings.Repeat("a", sanitizer.maxPromptLength+1)
				_, err = sanitizer.SanitizePrompt(input)
				if err == nil {
					t.Error("Should reject input over max length")
				}
			},
		},
		{
			name: "mixed_control_characters",
			test: func(t *testing.T) {
				input := "Hello\x01\x02\x03\x04\x05World\x06\x07\x08\x09Test"
				result, err := sanitizer.SanitizePrompt(input)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// \x09 is tab which is preserved, other control chars are removed
				if result != "HelloWorld\tTest" {
					t.Errorf("Expected 'HelloWorld\\tTest', got %q", result)
				}
			},
		},
		{
			name: "unicode_normalization",
			test: func(t *testing.T) {
				// Test with various Unicode characters
				input := "Café naïve résumé"
				result, err := sanitizer.SanitizePrompt(input)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != input {
					t.Errorf("Unicode characters should be preserved: got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestFilePath_EdgeCases(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name     string
		path     string
		wantErr  bool
	}{
		{
			name:    "empty_path",
			path:    "",
			wantErr: false,
		},
		{
			name:    "single_dot",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "current_dir_with_file",
			path:    "./config.yaml",
			wantErr: false,
		},
		{
			name:    "multiple_slashes",
			path:    "config//file.yaml",
			wantErr: false,
		},
		{
			name:    "trailing_slash",
			path:    "config/",
			wantErr: false,
		},
		{
			name:    "very_long_path",
			path:    strings.Repeat("a/", 100) + "file.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Concurrency Tests

func TestInputSanitization_Concurrent(t *testing.T) {
	sanitizer := NewInputSanitizer()
	const numGoroutines = 10
	const numOperations = 100

	results := make(chan error, numGoroutines*numOperations)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// Test sanitization
				_, err := sanitizer.SanitizePrompt("Test prompt")
				results <- err

				// Test file path validation
				err = sanitizer.ValidateFilePath("test.txt")
				results <- err

				// Test API key validation
				err = sanitizer.ValidateAPIKey("sk-1234567890abcdef")
				results <- err
			}
		}(i)
	}

	for i := 0; i < numGoroutines*numOperations*3; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}
}