package security

import "strings"

// InputSanitizer sanitizes and validates user inputs.
type InputSanitizer struct {
	maxPromptLength int
	maxFileSize     int64
	allowedChars    string
}

// NewInputSanitizer creates an input sanitizer with default limits.
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{
		maxPromptLength: 10000,
		maxFileSize:     1024 * 1024 * 10,
		allowedChars:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.,!?;:()[]{}\"'- \n\r\t",
	}
}

// SanitizePrompt removes unsafe control characters from input.
func (s *InputSanitizer) SanitizePrompt(input string) (string, error) {
	if len(input) > s.maxPromptLength {
		return "", NewSecurityError("prompt_too_long", "Prompt exceeds maximum length")
	}
	cleaned := strings.ReplaceAll(input, "\x00", "")
	var result strings.Builder
	for _, r := range cleaned {
		if r < 32 {
			if r == '\n' || r == '\r' || r == '\t' {
				result.WriteRune(r)
			}
			continue
		}
		if r == 127 {
			continue
		}
		result.WriteRune(r)
	}
	return result.String(), nil
}

// ValidateFilePath validates file paths to prevent traversal.
func (s *InputSanitizer) ValidateFilePath(path string) error {
	if strings.Contains(path, "\x00") {
		return NewSecurityError("invalid_path", "File path contains null bytes")
	}
	if len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
		if (path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z') {
			return NewSecurityError("path_traversal", "File path contains traversal patterns")
		}
	}
	for _, pattern := range []string{"../", "..\\", "/..", "\\..", "..."} {
		if strings.Contains(path, pattern) {
			return NewSecurityError("path_traversal", "File path contains traversal patterns")
		}
	}
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "/tmp/") && !strings.HasPrefix(path, "/var/tmp/") {
		return NewSecurityError("absolute_path", "Absolute paths outside temp directories not allowed")
	}
	return nil
}

// ValidateAPIKey validates API key shape.
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
	if strings.Contains(apiKey, " ") || strings.Contains(apiKey, "\n") || strings.Contains(apiKey, "\r") {
		return NewSecurityError("invalid_api_key", "API key contains invalid whitespace")
	}
	return nil
}

// SecurityError is a security-related error.
type SecurityError struct {
	Code    string
	Message string
}

func (e *SecurityError) Error() string {
	return e.Message
}

// NewSecurityError creates a security error.
func NewSecurityError(code, message string) *SecurityError {
	return &SecurityError{Code: code, Message: message}
}
