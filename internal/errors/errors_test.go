package errors

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestPEError_Basic(t *testing.T) {
	err := New(ErrCodeInvalidInput, "test message")
	if err.Code != ErrCodeInvalidInput {
		t.Errorf("expected code %s, got %s", ErrCodeInvalidInput, err.Code)
	}
	if err.Message != "test message" {
		t.Errorf("expected message 'test message', got '%s'", err.Message)
	}
	if err.Error() != "INVALID_INPUT: test message" {
		t.Errorf("unexpected error string: %s", err.Error())
	}
}

func TestPEError_WithContext(t *testing.T) {
	err := New(ErrCodeInternal, "test").WithContext("key", "value")
	if err.Context["key"] != "value" {
		t.Errorf("expected context key=value, got %v", err.Context)
	}
}

func TestPEError_WithComponent(t *testing.T) {
	err := New(ErrCodeInternal, "test").WithComponent("test-component")
	if err.Component != "test-component" {
		t.Errorf("expected component 'test-component', got '%s'", err.Component)
	}
}

func TestPEError_Wrap(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := Wrap(cause, ErrCodeInternal, "wrapped")

	if err.Cause != cause {
		t.Errorf("expected cause to be original error")
	}
	if err.Unwrap() != cause {
		t.Errorf("expected Unwrap to return original error")
	}
}

func TestPEError_Is(t *testing.T) {
	err1 := New(ErrCodeInvalidInput, "test1")
	err2 := New(ErrCodeInvalidInput, "test2")
	err3 := New(ErrCodeInternal, "test3")

	if !err1.Is(err2) {
		t.Error("expected err1.Is(err2) to be true (same code)")
	}
	if err1.Is(err3) {
		t.Error("expected err1.Is(err3) to be false (different code)")
	}
}

func TestProviderError(t *testing.T) {
	err := NewProviderError("test-provider", "test message")
	if err.Provider != "test-provider" {
		t.Errorf("expected provider 'test-provider', got '%s'", err.Provider)
	}
	if err.Code != ErrCodeProviderAPI {
		t.Errorf("expected code %s, got %s", ErrCodeProviderAPI, err.Code)
	}
}

func TestProviderError_WithModel(t *testing.T) {
	err := NewProviderError("test-provider", "test").WithModel("gpt-4")
	if err.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", err.Model)
	}
}

func TestInferenceError(t *testing.T) {
	err := NewInferenceError("test message")
	if err.Code != ErrCodeInferenceFailed {
		t.Errorf("expected code %s, got %s", ErrCodeInferenceFailed, err.Code)
	}
}

func TestOptimizationError(t *testing.T) {
	err := NewOptimizationError("textgrad", "test message")
	if err.Algorithm != "textgrad" {
		t.Errorf("expected algorithm 'textgrad', got '%s'", err.Algorithm)
	}
}

func TestEvaluationError(t *testing.T) {
	err := NewAssertionFailedError("equals", "expected", "actual")
	if err.Assertion != "equals" {
		t.Errorf("expected assertion 'equals', got '%s'", err.Assertion)
	}
	if err.Expected != "expected" {
		t.Errorf("expected Expected 'expected', got '%s'", err.Expected)
	}
	if err.Actual != "actual" {
		t.Errorf("expected Actual 'actual', got '%s'", err.Actual)
	}
}

func TestFileError(t *testing.T) {
	err := NewFileNotFoundError("/test/path")
	if err.Path != "/test/path" {
		t.Errorf("expected path '/test/path', got '%s'", err.Path)
	}
	if err.Code != ErrCodeFileNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeFileNotFound, err.Code)
	}
}

func TestModuleError(t *testing.T) {
	err := NewModuleNotFoundError("test-module")
	if err.ModuleName != "test-module" {
		t.Errorf("expected module name 'test-module', got '%s'", err.ModuleName)
	}
}

func TestSecurityError(t *testing.T) {
	err := NewSecurityValidationError("input-validation", "invalid input")
	if err.SecurityCheck != "input-validation" {
		t.Errorf("expected security check 'input-validation', got '%s'", err.SecurityCheck)
	}
	if err.Severity != SeverityCritical {
		t.Errorf("expected severity %s, got %s", SeverityCritical, err.Severity)
	}
}

func TestWrapProvider(t *testing.T) {
	originalErr := fmt.Errorf("timeout")
	err := WrapProvider(originalErr, "openai", "gpt-4")

	providerErr, ok := err.(*ProviderError)
	if !ok {
		t.Fatal("expected ProviderError")
	}

	if providerErr.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", providerErr.Provider)
	}
	if providerErr.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", providerErr.Model)
	}
	if providerErr.Cause != originalErr {
		t.Error("expected original error to be wrapped")
	}
}

func TestWrapNetwork(t *testing.T) {
	// Create a timeout error
	timeoutErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &TimeoutError{},
	}

	err := WrapNetwork(timeoutErr, "connect")

	peErr, ok := err.(*PEError)
	if !ok {
		t.Fatal("expected PEError")
	}

	if peErr.Code != ErrCodeNetworkTimeout {
		t.Errorf("expected code %s, got %s", ErrCodeNetworkTimeout, peErr.Code)
	}
	if !peErr.Retryable {
		t.Error("expected network timeout to be retryable")
	}
}

func TestWrapContext(t *testing.T) {
	// Test with context timeout
	err := WrapContext(context.DeadlineExceeded, "test operation")

	peErr, ok := err.(*PEError)
	if !ok {
		t.Fatal("expected PEError")
	}

	if peErr.Code != ErrCodeNetworkTimeout {
		t.Errorf("expected code %s, got %s", ErrCodeNetworkTimeout, peErr.Code)
	}
	if !peErr.Retryable {
		t.Error("expected timeout to be retryable")
	}

	// Test with context cancellation
	err = WrapContext(context.Canceled, "test operation")

	peErr, ok = err.(*PEError)
	if !ok {
		t.Fatal("expected PEError")
	}

	if peErr.Code != ErrCodeInternal {
		t.Errorf("expected code %s, got %s", ErrCodeInternal, peErr.Code)
	}
	if peErr.Retryable {
		t.Error("expected cancellation to not be retryable")
	}
}

func TestChain(t *testing.T) {
	err1 := New(ErrCodeInvalidInput, "error 1")
	err2 := New(ErrCodeInternal, "error 2")

	chainErr := Chain(err1, err2)

	peErr, ok := chainErr.(*PEError)
	if !ok {
		t.Fatal("expected PEError")
	}

	if peErr.Context["error_count"] != 2 {
		t.Errorf("expected error count 2, got %v", peErr.Context["error_count"])
	}
	if peErr.Cause != err1 {
		t.Error("expected first error to be primary cause")
	}
}

func TestAnnotate(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	annotatedErr := Annotate(originalErr, "test_key", "test_value")

	peErr, ok := annotatedErr.(*PEError)
	if !ok {
		t.Fatal("expected PEError")
	}

	if peErr.Context["test_key"] != "test_value" {
		t.Errorf("expected context test_key=test_value, got %v", peErr.Context)
	}
}

func TestWithTimeout(t *testing.T) {
	ctx := context.Background()

	// Test successful operation
	err := WithTimeout(ctx, time.Second, "test", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test timeout
	err = WithTimeout(ctx, time.Millisecond, "test", func(ctx context.Context) error {
		time.Sleep(time.Second)
		return nil
	})
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestUtilityFunctions(t *testing.T) {
	err := New(ErrCodeProviderAuth, "auth error").WithComponent("provider")
	err.Retryable = true

	// Test IsCode
	if !IsCode(err, ErrCodeProviderAuth) {
		t.Error("expected IsCode to return true")
	}

	// Test GetCode
	if GetCode(err) != ErrCodeProviderAuth {
		t.Errorf("expected code %s, got %s", ErrCodeProviderAuth, GetCode(err))
	}

	// Test IsRetryable
	if !IsRetryable(err) {
		t.Error("expected IsRetryable to return true")
	}

	// Test GetSeverity
	if GetSeverity(err) != SeverityMedium {
		t.Errorf("expected severity %s, got %s", SeverityMedium, GetSeverity(err))
	}

	// Test GetComponent
	if GetComponent(err) != "provider" {
		t.Errorf("expected component 'provider', got '%s'", GetComponent(err))
	}
}

func TestSuggestion(t *testing.T) {
	err := NewProviderAuthError("openai", "missing key")
	got := Suggestion(err)
	if got == "" {
		t.Fatal("Suggestion returned empty string")
	}
	if !strings.Contains(got, "credentials") {
		t.Fatalf("Suggestion = %q, want credentials guidance", got)
	}
}

// Custom timeout error for testing
type TimeoutError struct{}

func (e *TimeoutError) Error() string   { return "timeout" }
func (e *TimeoutError) Timeout() bool   { return true }
func (e *TimeoutError) Temporary() bool { return true }
