package errors

import (
	"fmt"
	"testing"
)

func TestProviderAuthError(t *testing.T) {
	err := NewProviderAuthError("openai", "Invalid API key")
	if err.Code != ErrCodeProviderAuth {
		t.Errorf("expected code %s, got %s", ErrCodeProviderAuth, err.Code)
	}
	if err.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", err.Provider)
	}
}

func TestProviderRateLimitError(t *testing.T) {
	err := NewProviderRateLimitError("anthropic", 60)
	if err.Code != ErrCodeProviderRateLimit {
		t.Errorf("expected code %s, got %s", ErrCodeProviderRateLimit, err.Code)
	}
	if err.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got '%s'", err.Provider)
	}
}

func TestProviderTimeoutError(t *testing.T) {
	err := NewProviderTimeoutError("openai", 30)
	if err.Code != ErrCodeProviderTimeout {
		t.Errorf("expected code %s, got %s", ErrCodeProviderTimeout, err.Code)
	}
}

func TestProviderError_WithStatusCode(t *testing.T) {
	err := NewProviderError("openai", "test").WithStatusCode(429)
	if err.StatusCode != 429 {
		t.Errorf("expected status code 429, got %d", err.StatusCode)
	}
}

func TestProviderError_WithRequestID(t *testing.T) {
	err := NewProviderError("openai", "test").WithRequestID("req-123")
	if err.RequestID != "req-123" {
		t.Errorf("expected request ID 'req-123', got '%s'", err.RequestID)
	}
}

func TestInferenceTimeoutError(t *testing.T) {
	err := NewInferenceTimeoutError(30)
	if err.Code != ErrCodeInferenceTimeout {
		t.Errorf("expected code %s, got %s", ErrCodeInferenceTimeout, err.Code)
	}
}

func TestInferenceContextError(t *testing.T) {
	err := NewInferenceContextError(10000, 8000)
	if err.Code != ErrCodeInferenceContextLen {
		t.Errorf("expected code %s, got %s", ErrCodeInferenceContextLen, err.Code)
	}
	if err.ContextSize != 10000 {
		t.Errorf("expected context size 10000, got %d", err.ContextSize)
	}
}

func TestInferenceError_WithModel(t *testing.T) {
	err := NewInferenceError("test").WithModel("gpt-4")
	if err.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", err.Model)
	}
}

func TestInferenceError_WithTokensUsed(t *testing.T) {
	err := NewInferenceError("test").WithTokensUsed(500)
	if err.TokensUsed != 500 {
		t.Errorf("expected tokens used 500, got %d", err.TokensUsed)
	}
}

func TestOptimizationTimeoutError(t *testing.T) {
	err := NewOptimizationTimeoutError("pe2", 60)
	if err.Code != ErrCodeOptimizationTimeout {
		t.Errorf("expected code %s, got %s", ErrCodeOptimizationTimeout, err.Code)
	}
	if err.Algorithm != "pe2" {
		t.Errorf("expected algorithm 'pe2', got '%s'", err.Algorithm)
	}
}

func TestOptimizationConvergedError(t *testing.T) {
	err := NewOptimizationConvergedError("textgrad", 0.75, 0.90)
	if err.Algorithm != "textgrad" {
		t.Errorf("expected algorithm 'textgrad', got '%s'", err.Algorithm)
	}
	if err.Score != 0.75 {
		t.Errorf("expected score 0.75, got %f", err.Score)
	}
	if err.Target != 0.90 {
		t.Errorf("expected target 0.90, got %f", err.Target)
	}
}

func TestOptimizationError_WithIteration(t *testing.T) {
	err := NewOptimizationError("pe2", "test").WithIteration(10)
	if err.Iteration != 10 {
		t.Errorf("expected iteration 10, got %d", err.Iteration)
	}
}

func TestNewEvaluationError(t *testing.T) {
	err := NewEvaluationError("test case failed")
	if err.Code != ErrCodeEvaluationFailed {
		t.Errorf("expected code %s, got %s", ErrCodeEvaluationFailed, err.Code)
	}
}

func TestEvaluationTimeoutError(t *testing.T) {
	err := NewEvaluationTimeoutError(30)
	if err.Code != ErrCodeEvaluationTimeout {
		t.Errorf("expected code %s, got %s", ErrCodeEvaluationTimeout, err.Code)
	}
}

func TestEvaluationError_WithTestCase(t *testing.T) {
	err := NewEvaluationError("test").WithTestCase("test-001")
	if err.TestCase != "test-001" {
		t.Errorf("expected test case 'test-001', got '%s'", err.TestCase)
	}
}

func TestEvaluationError_WithScore(t *testing.T) {
	err := NewEvaluationError("test").WithScore(0.85, 0.90)
	if err.Score != 0.85 {
		t.Errorf("expected score 0.85, got %f", err.Score)
	}
	if err.Threshold != 0.90 {
		t.Errorf("expected threshold 0.90, got %f", err.Threshold)
	}
}

func TestFileReadError(t *testing.T) {
	cause := fmt.Errorf("permission denied")
	err := NewFileReadError("/path/to/file", cause)
	if err.Code != ErrCodeFileRead {
		t.Errorf("expected code %s, got %s", ErrCodeFileRead, err.Code)
	}
	if err.Path != "/path/to/file" {
		t.Errorf("expected path '/path/to/file', got '%s'", err.Path)
	}
}

func TestFileWriteError(t *testing.T) {
	cause := fmt.Errorf("disk full")
	err := NewFileWriteError("/path/to/file", cause)
	if err.Code != ErrCodeFileWrite {
		t.Errorf("expected code %s, got %s", ErrCodeFileWrite, err.Code)
	}
}

func TestFileFormatError(t *testing.T) {
	cause := fmt.Errorf("invalid syntax")
	err := NewFileFormatError("/path/to/file", "yaml", cause)
	if err.Code != ErrCodeFileFormat {
		t.Errorf("expected code %s, got %s", ErrCodeFileFormat, err.Code)
	}
}

func TestFileError_WithSize(t *testing.T) {
	err := NewFileNotFoundError("/path").WithSize(1024)
	if err.Size != 1024 {
		t.Errorf("expected size 1024, got %d", err.Size)
	}
}

func TestModuleInvalidError(t *testing.T) {
	err := NewModuleInvalidError("test-module", "invalid format")
	if err.Code != ErrCodeModuleInvalid {
		t.Errorf("expected code %s, got %s", ErrCodeModuleInvalid, err.Code)
	}
	if err.ModuleName != "test-module" {
		t.Errorf("expected module name 'test-module', got '%s'", err.ModuleName)
	}
}

func TestConfigurationError(t *testing.T) {
	err := NewConfigurationError("providers.default", "invalid provider").WithSource("pe.yaml")
	if err.Code != ErrCodeInvalidConfig {
		t.Errorf("expected code %s, got %s", ErrCodeInvalidConfig, err.Code)
	}
	if err.Key != "providers.default" {
		t.Errorf("expected key providers.default, got %s", err.Key)
	}
	if err.Source != "pe.yaml" {
		t.Errorf("expected source pe.yaml, got %s", err.Source)
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("eval.max_concurrency", "must be positive").WithValue("0")
	if err.Code != ErrCodeValidation {
		t.Errorf("expected code %s, got %s", ErrCodeValidation, err.Code)
	}
	if err.Field != "eval.max_concurrency" {
		t.Errorf("expected field eval.max_concurrency, got %s", err.Field)
	}
}

func TestNetworkError(t *testing.T) {
	err := NewNetworkError("dial", "network unavailable").WithEndpoint("https://example.com")
	if err.Code != ErrCodeNetworkUnavailable {
		t.Errorf("expected code %s, got %s", ErrCodeNetworkUnavailable, err.Code)
	}
	if !err.Retryable {
		t.Error("expected network error to be retryable")
	}
	if err.Endpoint != "https://example.com" {
		t.Errorf("expected endpoint, got %s", err.Endpoint)
	}
}

func TestAuthenticationError(t *testing.T) {
	err := NewAuthenticationError("user", "missing token").WithMethod("bearer")
	if err.Code != ErrCodeAuthentication {
		t.Errorf("expected code %s, got %s", ErrCodeAuthentication, err.Code)
	}
	if err.Method != "bearer" {
		t.Errorf("expected method bearer, got %s", err.Method)
	}
}

func TestModuleDependencyError(t *testing.T) {
	err := NewModuleDependencyError("test-module", "dep-module")
	if err.Code != ErrCodeModuleDependency {
		t.Errorf("expected code %s, got %s", ErrCodeModuleDependency, err.Code)
	}
	if err.ModuleName != "test-module" {
		t.Errorf("expected module name 'test-module', got '%s'", err.ModuleName)
	}
}

func TestModuleError_WithVersion(t *testing.T) {
	err := NewModuleNotFoundError("test").WithVersion("v1.2.3")
	if err.Version != "v1.2.3" {
		t.Errorf("expected version 'v1.2.3', got '%s'", err.Version)
	}
}

func TestModuleError_WithRepositoryURL(t *testing.T) {
	err := NewModuleNotFoundError("test").WithRepositoryURL("https://github.com/test/repo")
	if err.RepositoryURL != "https://github.com/test/repo" {
		t.Errorf("expected repository URL 'https://github.com/test/repo', got '%s'", err.RepositoryURL)
	}
}

func TestSecurityAuthError(t *testing.T) {
	err := NewSecurityAuthError("token expired")
	if err.Code != ErrCodeSecurityAuth {
		t.Errorf("expected code %s, got %s", ErrCodeSecurityAuth, err.Code)
	}
	if err.RiskLevel != "HIGH" {
		t.Errorf("expected risk level 'HIGH', got '%s'", err.RiskLevel)
	}
}

func TestSecurityPolicyError(t *testing.T) {
	err := NewSecurityPolicyError("content-policy", "forbidden content detected")
	if err.Code != ErrCodeSecurityPolicy {
		t.Errorf("expected code %s, got %s", ErrCodeSecurityPolicy, err.Code)
	}
	if err.Policy != "content-policy" {
		t.Errorf("expected policy 'content-policy', got '%s'", err.Policy)
	}
}

func TestSecurityError_WithRiskLevel(t *testing.T) {
	err := NewSecurityValidationError("test", "validation failed").WithRiskLevel("high")
	if err.RiskLevel != "high" {
		t.Errorf("expected risk level 'high', got '%s'", err.RiskLevel)
	}
}

func TestPEError_IsRetryable(t *testing.T) {
	err := New(ErrCodeNetworkTimeout, "timeout")
	err.Retryable = true
	if !err.IsRetryable() {
		t.Error("expected IsRetryable to return true")
	}
}

func TestPEError_GetSeverity(t *testing.T) {
	err := New(ErrCodeInternal, "test")
	err.Severity = SeverityCritical
	if err.GetSeverity() != SeverityCritical {
		t.Errorf("expected severity %s, got %s", SeverityCritical, err.GetSeverity())
	}
}

func TestNewf(t *testing.T) {
	err := Newf(ErrCodeInvalidInput, "invalid value: %d", 42)
	if err.Message != "invalid value: 42" {
		t.Errorf("expected message 'invalid value: 42', got '%s'", err.Message)
	}
}

func TestWrapf(t *testing.T) {
	cause := New(ErrCodeInternal, "original")
	err := Wrapf(cause, ErrCodeInvalidInput, "wrapped: %s", "test")
	if err.Message != "wrapped: test" {
		t.Errorf("expected message 'wrapped: test', got '%s'", err.Message)
	}
	if err.Cause != cause {
		t.Error("expected cause to be original error")
	}
}

func TestGetContext(t *testing.T) {
	err := New(ErrCodeInternal, "test").WithContext("key", "value")
	ctx := GetContext(err)
	if ctx["key"] != "value" {
		t.Errorf("expected context key=value, got %v", ctx)
	}
}
