package errors

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
)

// WrapProvider wraps an error with provider-specific context.
func WrapProvider(err error, provider, model string) error {
	if err == nil {
		return nil
	}

	// Create base provider error
	providerErr := NewProviderError(provider, err.Error())
	providerErr.Cause = err

	if model != "" {
		providerErr.WithModel(model)
	}

	// Analyze the error to determine specific type and retryability
	analyzeProviderError(providerErr, err)

	return providerErr
}

// WrapInference wraps an error with inference-specific context.
func WrapInference(err error, model string) error {
	if err == nil {
		return nil
	}

	inferenceErr := NewInferenceError(err.Error())
	inferenceErr.Cause = err

	if model != "" {
		inferenceErr.WithModel(model)
	}

	// Analyze the error to determine specific type
	analyzeInferenceError(inferenceErr, err)

	return inferenceErr
}

// WrapOptimization wraps an error with optimization-specific context.
func WrapOptimization(err error, algorithm string) error {
	if err == nil {
		return nil
	}

	optimizationErr := NewOptimizationError(algorithm, err.Error())
	optimizationErr.Cause = err

	// Analyze the error to determine specific type
	analyzeOptimizationError(optimizationErr, err)

	return optimizationErr
}

// WrapEvaluation wraps an error with evaluation-specific context.
func WrapEvaluation(err error, testCase string) error {
	if err == nil {
		return nil
	}

	evaluationErr := NewEvaluationError(err.Error())
	evaluationErr.Cause = err

	if testCase != "" {
		evaluationErr.WithTestCase(testCase)
	}

	// Analyze the error to determine specific type
	analyzeEvaluationError(evaluationErr, err)

	return evaluationErr
}

// WrapFile wraps an error with file-specific context.
func WrapFile(err error, path, operation string) error {
	if err == nil {
		return nil
	}

	// Check for specific file errors
	if os.IsNotExist(err) {
		return NewFileNotFoundError(path)
	}

	var fileErr *FileError
	switch operation {
	case "read":
		fileErr = NewFileReadError(path, err)
	case "write":
		fileErr = NewFileWriteError(path, err)
	default:
		fileErr = &FileError{
			PEError:   Wrap(err, ErrCodeFileRead, fmt.Sprintf("file operation failed: %s", path)),
			Path:      path,
			Operation: operation,
		}
	}

	// Analyze the error for additional context
	analyzeFileError(fileErr, err)

	return fileErr
}

// WrapModule wraps an error with module-specific context.
func WrapModule(err error, moduleName, version string) error {
	if err == nil {
		return nil
	}

	moduleErr := &ModuleError{
		PEError:    Wrap(err, ErrCodeModuleInvalid, fmt.Sprintf("module error: %s", moduleName)),
		ModuleName: moduleName,
		Version:    version,
	}

	// Analyze the error to determine specific type
	analyzeModuleError(moduleErr, err)

	return moduleErr
}

// WrapSecurity wraps an error with security-specific context.
func WrapSecurity(err error, check string) error {
	if err == nil {
		return nil
	}

	securityErr := NewSecurityValidationError(check, err.Error())
	securityErr.Cause = err

	// Analyze the error for security context
	analyzeSecurityError(securityErr, err)

	return securityErr
}

// WrapNetwork wraps network-related errors with appropriate context.
func WrapNetwork(err error, operation string) error {
	if err == nil {
		return nil
	}

	var code ErrorCode
	var retryable bool

	// Analyze network error types
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			code = ErrCodeNetworkTimeout
			retryable = true
		} else {
			code = ErrCodeNetworkUnavailable
			retryable = true
		}
	} else if _, ok := err.(*net.DNSError); ok {
		code = ErrCodeNetworkDNS
		retryable = true
	} else if _, ok := err.(*url.Error); ok {
		code = ErrCodeNetworkUnavailable
		retryable = true
	} else {
		code = ErrCodeNetworkUnavailable
		retryable = false
	}

	return &PEError{
		Code:      code,
		Message:   fmt.Sprintf("network %s failed: %s", operation, err.Error()),
		Cause:     err,
		Severity:  SeverityMedium,
		Retryable: retryable,
		Component: "network",
	}
}

// WrapContext wraps context-related errors.
func WrapContext(err error, operation string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		return &PEError{
			Code:      ErrCodeInternal,
			Message:   fmt.Sprintf("%s was canceled", operation),
			Cause:     err,
			Severity:  SeverityLow,
			Retryable: false,
			Component: "context",
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &PEError{
			Code:      ErrCodeNetworkTimeout,
			Message:   fmt.Sprintf("%s timed out", operation),
			Cause:     err,
			Severity:  SeverityMedium,
			Retryable: true,
			Component: "context",
		}
	}

	return Wrap(err, ErrCodeInternal, fmt.Sprintf("context error during %s", operation))
}

// analyzeProviderError analyzes provider errors to set appropriate codes and retryability.
func analyzeProviderError(providerErr *ProviderError, err error) {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "invalid api key") ||
		strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "forbidden"):
		providerErr.Code = ErrCodeProviderAuth
		providerErr.Retryable = false
		providerErr.Severity = SeverityHigh

	case strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "too many requests"):
		providerErr.Code = ErrCodeProviderRateLimit
		providerErr.Retryable = true
		providerErr.Severity = SeverityMedium

	case strings.Contains(errMsg, "quota") || strings.Contains(errMsg, "billing"):
		providerErr.Code = ErrCodeProviderQuota
		providerErr.Retryable = false
		providerErr.Severity = SeverityHigh

	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		providerErr.Code = ErrCodeProviderTimeout
		providerErr.Retryable = true
		providerErr.Severity = SeverityMedium

	case strings.Contains(errMsg, "unavailable") || strings.Contains(errMsg, "service") ||
		strings.Contains(errMsg, "connection"):
		providerErr.Code = ErrCodeProviderUnavailable
		providerErr.Retryable = true
		providerErr.Severity = SeverityMedium

	default:
		providerErr.Code = ErrCodeProviderAPI
		providerErr.Retryable = false
		providerErr.Severity = SeverityMedium
	}

	// Check for network errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			providerErr.Code = ErrCodeProviderTimeout
			providerErr.Retryable = true
		}
	}
}

// analyzeInferenceError analyzes inference errors to set appropriate codes.
func analyzeInferenceError(inferenceErr *InferenceError, err error) {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "context") && (strings.Contains(errMsg, "length") ||
		strings.Contains(errMsg, "limit") || strings.Contains(errMsg, "too long")):
		inferenceErr.Code = ErrCodeInferenceContextLen
		inferenceErr.Retryable = false
		inferenceErr.Severity = SeverityHigh

	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		inferenceErr.Code = ErrCodeInferenceTimeout
		inferenceErr.Retryable = true
		inferenceErr.Severity = SeverityMedium

	case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "malformed"):
		inferenceErr.Code = ErrCodeInferenceInvalid
		inferenceErr.Retryable = false
		inferenceErr.Severity = SeverityMedium

	default:
		inferenceErr.Code = ErrCodeInferenceFailed
		inferenceErr.Retryable = false
		inferenceErr.Severity = SeverityMedium
	}
}

// analyzeOptimizationError analyzes optimization errors to set appropriate codes.
func analyzeOptimizationError(optimizationErr *OptimizationError, err error) {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		optimizationErr.Code = ErrCodeOptimizationTimeout
		optimizationErr.Retryable = true
		optimizationErr.Severity = SeverityMedium

	case strings.Contains(errMsg, "converge") || strings.Contains(errMsg, "convergence"):
		optimizationErr.Code = ErrCodeOptimizationConverged
		optimizationErr.Retryable = false
		optimizationErr.Severity = SeverityLow

	default:
		optimizationErr.Code = ErrCodeOptimizationFailed
		optimizationErr.Retryable = false
		optimizationErr.Severity = SeverityMedium
	}
}

// analyzeEvaluationError analyzes evaluation errors to set appropriate codes.
func analyzeEvaluationError(evaluationErr *EvaluationError, err error) {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "assertion") || strings.Contains(errMsg, "assert"):
		evaluationErr.Code = ErrCodeAssertionFailed
		evaluationErr.Retryable = false
		evaluationErr.Severity = SeverityMedium

	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		evaluationErr.Code = ErrCodeEvaluationTimeout
		evaluationErr.Retryable = true
		evaluationErr.Severity = SeverityMedium

	default:
		evaluationErr.Code = ErrCodeEvaluationFailed
		evaluationErr.Retryable = false
		evaluationErr.Severity = SeverityMedium
	}
}

// analyzeFileError analyzes file errors for additional context.
func analyzeFileError(fileErr *FileError, err error) {
	// Check for specific system errors
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.EACCES:
				fileErr.Code = ErrCodeSecurityAuth
				fileErr.Message = fmt.Sprintf("permission denied: %s", fileErr.Path)
				fileErr.Severity = SeverityHigh
			case syscall.ENOSPC:
				fileErr.Code = ErrCodeFileWrite
				fileErr.Message = fmt.Sprintf("no space left on device: %s", fileErr.Path)
				fileErr.Severity = SeverityHigh
			case syscall.EMFILE, syscall.ENFILE:
				fileErr.Code = ErrCodeFileRead
				fileErr.Message = fmt.Sprintf("too many open files: %s", fileErr.Path)
				fileErr.Severity = SeverityMedium
				fileErr.Retryable = true
			}
		}
	}
}

// analyzeModuleError analyzes module errors to set appropriate codes.
func analyzeModuleError(moduleErr *ModuleError, err error) {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404"):
		moduleErr.Code = ErrCodeModuleNotFound
		moduleErr.Retryable = false
		moduleErr.Severity = SeverityMedium

	case strings.Contains(errMsg, "dependency") || strings.Contains(errMsg, "require"):
		moduleErr.Code = ErrCodeModuleDependency
		moduleErr.Retryable = false
		moduleErr.Severity = SeverityMedium

	case strings.Contains(errMsg, "registry") || strings.Contains(errMsg, "download"):
		moduleErr.Code = ErrCodeModuleRegistry
		moduleErr.Retryable = true
		moduleErr.Severity = SeverityMedium

	default:
		moduleErr.Code = ErrCodeModuleInvalid
		moduleErr.Retryable = false
		moduleErr.Severity = SeverityMedium
	}
}

// analyzeSecurityError analyzes security errors for risk level assessment.
func analyzeSecurityError(securityErr *SecurityError, err error) {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "injection") || strings.Contains(errMsg, "malicious") ||
		strings.Contains(errMsg, "exploit"):
		securityErr.Severity = SeverityCritical
		securityErr.RiskLevel = "CRITICAL"

	case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "forbidden"):
		securityErr.Code = ErrCodeSecurityAuth
		securityErr.Severity = SeverityHigh
		securityErr.RiskLevel = "HIGH"

	case strings.Contains(errMsg, "policy") || strings.Contains(errMsg, "violation"):
		securityErr.Code = ErrCodeSecurityPolicy
		securityErr.Severity = SeverityHigh
		securityErr.RiskLevel = "HIGH"

	default:
		securityErr.Severity = SeverityMedium
		securityErr.RiskLevel = "MEDIUM"
	}
}

// Chain creates a chain of errors for complex operations.
func Chain(errs ...error) error {
	var validErrs []error
	for _, err := range errs {
		if err != nil {
			validErrs = append(validErrs, err)
		}
	}

	if len(validErrs) == 0 {
		return nil
	}

	if len(validErrs) == 1 {
		return validErrs[0]
	}

	// Create a chain of errors
	chainErr := New(ErrCodeInternal, "multiple errors occurred")
	chainErr.WithContext("error_count", len(validErrs))

	var messages []string
	for i, err := range validErrs {
		messages = append(messages, fmt.Sprintf("[%d] %s", i+1, err.Error()))
		chainErr.WithContext(fmt.Sprintf("error_%d", i+1), err.Error())
	}

	chainErr.Message = strings.Join(messages, "; ")
	chainErr.Cause = validErrs[0] // First error as primary cause

	return chainErr
}

// Annotate adds contextual information to an existing error.
func Annotate(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}

	var peErr *PEError
	if errors.As(err, &peErr) {
		return peErr.WithContext(key, value)
	}

	// Wrap in PEError to add context
	wrapped := Wrap(err, ErrCodeUnknown, err.Error())
	return wrapped.WithContext(key, value)
}

// WithTimeout wraps an operation with timeout context.
func WithTimeout(ctx context.Context, timeout time.Duration, operation string, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			return WrapContext(err, operation)
		}
		return nil
	case <-ctx.Done():
		return WrapContext(ctx.Err(), operation)
	}
}
