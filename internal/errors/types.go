// Package errors provides standardized error types and utilities for the PE toolkit.
// It implements a comprehensive error handling system with error codes, wrapping,
// and recovery strategies following Go best practices.
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents standardized error codes for different types of failures.
type ErrorCode string

// Standard error codes for the PE toolkit.
const (
	// Core system errors
	ErrCodeUnknown        ErrorCode = "UNKNOWN"
	ErrCodeInternal       ErrorCode = "INTERNAL"
	ErrCodeInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrCodeInvalidConfig  ErrorCode = "INVALID_CONFIG"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"

	// Provider-related errors
	ErrCodeProviderUnavailable ErrorCode = "PROVIDER_UNAVAILABLE"
	ErrCodeProviderAuth        ErrorCode = "PROVIDER_AUTH"
	ErrCodeProviderQuota       ErrorCode = "PROVIDER_QUOTA"
	ErrCodeProviderRateLimit   ErrorCode = "PROVIDER_RATE_LIMIT"
	ErrCodeProviderTimeout     ErrorCode = "PROVIDER_TIMEOUT"
	ErrCodeProviderAPI         ErrorCode = "PROVIDER_API"

	// Inference-related errors
	ErrCodeInferenceTimeout    ErrorCode = "INFERENCE_TIMEOUT"
	ErrCodeInferenceFailed     ErrorCode = "INFERENCE_FAILED"
	ErrCodeInferenceInvalid    ErrorCode = "INFERENCE_INVALID"
	ErrCodeInferenceContextLen ErrorCode = "INFERENCE_CONTEXT_LENGTH"

	// Optimization-related errors
	ErrCodeOptimizationFailed    ErrorCode = "OPTIMIZATION_FAILED"
	ErrCodeOptimizationTimeout   ErrorCode = "OPTIMIZATION_TIMEOUT"
	ErrCodeOptimizationConverged ErrorCode = "OPTIMIZATION_CONVERGED"

	// Evaluation-related errors
	ErrCodeEvaluationFailed  ErrorCode = "EVALUATION_FAILED"
	ErrCodeEvaluationTimeout ErrorCode = "EVALUATION_TIMEOUT"
	ErrCodeAssertionFailed   ErrorCode = "ASSERTION_FAILED"

	// File and I/O errors
	ErrCodeFileNotFound ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFileRead     ErrorCode = "FILE_READ"
	ErrCodeFileWrite    ErrorCode = "FILE_WRITE"
	ErrCodeFileFormat   ErrorCode = "FILE_FORMAT"

	// Network errors
	ErrCodeNetworkUnavailable ErrorCode = "NETWORK_UNAVAILABLE"
	ErrCodeNetworkTimeout     ErrorCode = "NETWORK_TIMEOUT"
	ErrCodeNetworkDNS         ErrorCode = "NETWORK_DNS"

	// Security errors
	ErrCodeSecurityValidation ErrorCode = "SECURITY_VALIDATION"
	ErrCodeSecurityAuth       ErrorCode = "SECURITY_AUTH"
	ErrCodeSecurityPolicy     ErrorCode = "SECURITY_POLICY"

	// Module-related errors
	ErrCodeModuleNotFound    ErrorCode = "MODULE_NOT_FOUND"
	ErrCodeModuleInvalid     ErrorCode = "MODULE_INVALID"
	ErrCodeModuleDependency  ErrorCode = "MODULE_DEPENDENCY"
	ErrCodeModuleRegistry    ErrorCode = "MODULE_REGISTRY"

	// Template and parsing errors
	ErrCodeTemplateInvalid ErrorCode = "TEMPLATE_INVALID"
	ErrCodeTemplateRender  ErrorCode = "TEMPLATE_RENDER"
	ErrCodeParseError      ErrorCode = "PARSE_ERROR"
)

// Severity represents the severity level of an error.
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// PEError is the base error type for all PE toolkit errors.
// It provides structured error information with error codes, context,
// and support for error wrapping and unwrapping.
type PEError struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"-"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Severity  Severity               `json:"severity"`
	Retryable bool                   `json:"retryable"`
	Component string                 `json:"component,omitempty"`
}

// Error implements the error interface.
func (e *PEError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error, supporting Go 1.13+ error unwrapping.
func (e *PEError) Unwrap() error {
	return e.Cause
}

// Is reports whether the target error matches this error's type or code.
func (e *PEError) Is(target error) bool {
	if target == nil {
		return false
	}

	var peErr *PEError
	if errors.As(target, &peErr) {
		return e.Code == peErr.Code
	}

	return false
}

// WithContext adds context information to the error.
func (e *PEError) WithContext(key string, value interface{}) *PEError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithComponent sets the component that generated this error.
func (e *PEError) WithComponent(component string) *PEError {
	e.Component = component
	return e
}

// IsRetryable returns whether this error indicates a retryable condition.
func (e *PEError) IsRetryable() bool {
	return e.Retryable
}

// GetSeverity returns the severity level of this error.
func (e *PEError) GetSeverity() Severity {
	return e.Severity
}

// New creates a new PEError with the specified code and message.
func New(code ErrorCode, message string) *PEError {
	return &PEError{
		Code:      code,
		Message:   message,
		Severity:  SeverityMedium,
		Retryable: false,
	}
}

// Newf creates a new PEError with formatted message.
func Newf(code ErrorCode, format string, args ...interface{}) *PEError {
	return &PEError{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Severity:  SeverityMedium,
		Retryable: false,
	}
}

// Wrap wraps an existing error with a PEError.
func Wrap(err error, code ErrorCode, message string) *PEError {
	if err == nil {
		return nil
	}

	return &PEError{
		Code:      code,
		Message:   message,
		Cause:     err,
		Severity:  SeverityMedium,
		Retryable: false,
	}
}

// Wrapf wraps an existing error with a PEError and formatted message.
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *PEError {
	if err == nil {
		return nil
	}

	return &PEError{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Cause:     err,
		Severity:  SeverityMedium,
		Retryable: false,
	}
}

// IsCode checks if an error has a specific error code.
func IsCode(err error, code ErrorCode) bool {
	var peErr *PEError
	if errors.As(err, &peErr) {
		return peErr.Code == code
	}
	return false
}

// GetCode extracts the error code from an error.
func GetCode(err error) ErrorCode {
	var peErr *PEError
	if errors.As(err, &peErr) {
		return peErr.Code
	}
	return ErrCodeUnknown
}

// IsRetryable checks if an error indicates a retryable condition.
func IsRetryable(err error) bool {
	var peErr *PEError
	if errors.As(err, &peErr) {
		return peErr.Retryable
	}
	return false
}

// GetSeverity extracts the severity level from an error.
func GetSeverity(err error) Severity {
	var peErr *PEError
	if errors.As(err, &peErr) {
		return peErr.Severity
	}
	return SeverityMedium
}

// GetComponent extracts the component name from an error.
func GetComponent(err error) string {
	var peErr *PEError
	if errors.As(err, &peErr) {
		return peErr.Component
	}
	return ""
}

// GetContext extracts context information from an error.
func GetContext(err error) map[string]interface{} {
	var peErr *PEError
	if errors.As(err, &peErr) {
		return peErr.Context
	}
	return nil
}