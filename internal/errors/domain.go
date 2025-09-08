package errors

import "fmt"

// ProviderError represents errors related to LLM providers.
type ProviderError struct {
	*PEError
	Provider    string `json:"provider"`
	Model       string `json:"model,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
	RetryAfter  int    `json:"retry_after,omitempty"`
}

// NewProviderError creates a new provider error.
func NewProviderError(provider, message string) *ProviderError {
	return &ProviderError{
		PEError: &PEError{
			Code:      ErrCodeProviderAPI,
			Message:   message,
			Severity:  SeverityHigh,
			Retryable: false,
			Component: "provider",
		},
		Provider: provider,
	}
}

// NewProviderAuthError creates a provider authentication error.
func NewProviderAuthError(provider, message string) *ProviderError {
	return &ProviderError{
		PEError: &PEError{
			Code:      ErrCodeProviderAuth,
			Message:   message,
			Severity:  SeverityHigh,
			Retryable: false,
			Component: "provider",
		},
		Provider: provider,
	}
}

// NewProviderRateLimitError creates a provider rate limit error.
func NewProviderRateLimitError(provider string, retryAfter int) *ProviderError {
	return &ProviderError{
		PEError: &PEError{
			Code:      ErrCodeProviderRateLimit,
			Message:   fmt.Sprintf("rate limit exceeded for provider %s", provider),
			Severity:  SeverityMedium,
			Retryable: true,
			Component: "provider",
		},
		Provider:   provider,
		RetryAfter: retryAfter,
	}
}

// NewProviderTimeoutError creates a provider timeout error.
func NewProviderTimeoutError(provider string, timeout int) *ProviderError {
	return &ProviderError{
		PEError: &PEError{
			Code:      ErrCodeProviderTimeout,
			Message:   fmt.Sprintf("timeout after %ds for provider %s", timeout, provider),
			Severity:  SeverityMedium,
			Retryable: true,
			Component: "provider",
		},
		Provider: provider,
	}
}

// WithModel adds model information to the provider error.
func (e *ProviderError) WithModel(model string) *ProviderError {
	e.Model = model
	return e
}

// WithStatusCode adds HTTP status code to the provider error.
func (e *ProviderError) WithStatusCode(code int) *ProviderError {
	e.StatusCode = code
	return e
}

// WithRequestID adds request ID to the provider error.
func (e *ProviderError) WithRequestID(id string) *ProviderError {
	e.RequestID = id
	return e
}

// InferenceError represents errors during inference operations.
type InferenceError struct {
	*PEError
	Model       string `json:"model,omitempty"`
	TokensUsed  int    `json:"tokens_used,omitempty"`
	ContextSize int    `json:"context_size,omitempty"`
}

// NewInferenceError creates a new inference error.
func NewInferenceError(message string) *InferenceError {
	return &InferenceError{
		PEError: &PEError{
			Code:      ErrCodeInferenceFailed,
			Message:   message,
			Severity:  SeverityHigh,
			Retryable: false,
			Component: "inference",
		},
	}
}

// NewInferenceTimeoutError creates an inference timeout error.
func NewInferenceTimeoutError(timeout int) *InferenceError {
	return &InferenceError{
		PEError: &PEError{
			Code:      ErrCodeInferenceTimeout,
			Message:   fmt.Sprintf("inference timeout after %ds", timeout),
			Severity:  SeverityMedium,
			Retryable: true,
			Component: "inference",
		},
	}
}

// NewInferenceContextError creates a context length error.
func NewInferenceContextError(contextSize, maxSize int) *InferenceError {
	return &InferenceError{
		PEError: &PEError{
			Code:      ErrCodeInferenceContextLen,
			Message:   fmt.Sprintf("context length %d exceeds maximum %d", contextSize, maxSize),
			Severity:  SeverityHigh,
			Retryable: false,
			Component: "inference",
		},
		ContextSize: contextSize,
	}
}

// WithModel adds model information to the inference error.
func (e *InferenceError) WithModel(model string) *InferenceError {
	e.Model = model
	return e
}

// WithTokensUsed adds token usage information to the inference error.
func (e *InferenceError) WithTokensUsed(tokens int) *InferenceError {
	e.TokensUsed = tokens
	return e
}

// OptimizationError represents errors during prompt optimization.
type OptimizationError struct {
	*PEError
	Algorithm   string  `json:"algorithm,omitempty"`
	Iteration   int     `json:"iteration,omitempty"`
	Score       float64 `json:"score,omitempty"`
	Target      float64 `json:"target,omitempty"`
	MaxRetries  int     `json:"max_retries,omitempty"`
}

// NewOptimizationError creates a new optimization error.
func NewOptimizationError(algorithm, message string) *OptimizationError {
	return &OptimizationError{
		PEError: &PEError{
			Code:      ErrCodeOptimizationFailed,
			Message:   message,
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "optimization",
		},
		Algorithm: algorithm,
	}
}

// NewOptimizationTimeoutError creates an optimization timeout error.
func NewOptimizationTimeoutError(algorithm string, timeout int) *OptimizationError {
	return &OptimizationError{
		PEError: &PEError{
			Code:      ErrCodeOptimizationTimeout,
			Message:   fmt.Sprintf("optimization timeout after %ds using %s", timeout, algorithm),
			Severity:  SeverityMedium,
			Retryable: true,
			Component: "optimization",
		},
		Algorithm: algorithm,
	}
}

// NewOptimizationConvergedError creates an optimization convergence error.
func NewOptimizationConvergedError(algorithm string, score, target float64) *OptimizationError {
	return &OptimizationError{
		PEError: &PEError{
			Code:      ErrCodeOptimizationConverged,
			Message:   fmt.Sprintf("optimization failed to converge: score %.3f < target %.3f", score, target),
			Severity:  SeverityLow,
			Retryable: false,
			Component: "optimization",
		},
		Algorithm: algorithm,
		Score:     score,
		Target:    target,
	}
}

// WithIteration adds iteration information to the optimization error.
func (e *OptimizationError) WithIteration(iteration int) *OptimizationError {
	e.Iteration = iteration
	return e
}

// EvaluationError represents errors during evaluation operations.
type EvaluationError struct {
	*PEError
	TestCase   string  `json:"test_case,omitempty"`
	Assertion  string  `json:"assertion,omitempty"`
	Expected   string  `json:"expected,omitempty"`
	Actual     string  `json:"actual,omitempty"`
	Score      float64 `json:"score,omitempty"`
	Threshold  float64 `json:"threshold,omitempty"`
}

// NewEvaluationError creates a new evaluation error.
func NewEvaluationError(message string) *EvaluationError {
	return &EvaluationError{
		PEError: &PEError{
			Code:      ErrCodeEvaluationFailed,
			Message:   message,
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "evaluation",
		},
	}
}

// NewAssertionFailedError creates an assertion failure error.
func NewAssertionFailedError(assertion, expected, actual string) *EvaluationError {
	return &EvaluationError{
		PEError: &PEError{
			Code:      ErrCodeAssertionFailed,
			Message:   fmt.Sprintf("assertion '%s' failed: expected '%s', got '%s'", assertion, expected, actual),
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "evaluation",
		},
		Assertion: assertion,
		Expected:  expected,
		Actual:    actual,
	}
}

// NewEvaluationTimeoutError creates an evaluation timeout error.
func NewEvaluationTimeoutError(timeout int) *EvaluationError {
	return &EvaluationError{
		PEError: &PEError{
			Code:      ErrCodeEvaluationTimeout,
			Message:   fmt.Sprintf("evaluation timeout after %ds", timeout),
			Severity:  SeverityMedium,
			Retryable: true,
			Component: "evaluation",
		},
	}
}

// WithTestCase adds test case information to the evaluation error.
func (e *EvaluationError) WithTestCase(testCase string) *EvaluationError {
	e.TestCase = testCase
	return e
}

// WithScore adds score information to the evaluation error.
func (e *EvaluationError) WithScore(score, threshold float64) *EvaluationError {
	e.Score = score
	e.Threshold = threshold
	return e
}

// FileError represents file and I/O related errors.
type FileError struct {
	*PEError
	Path      string `json:"path"`
	Operation string `json:"operation,omitempty"`
	Size      int64  `json:"size,omitempty"`
}

// NewFileNotFoundError creates a file not found error.
func NewFileNotFoundError(path string) *FileError {
	return &FileError{
		PEError: &PEError{
			Code:      ErrCodeFileNotFound,
			Message:   fmt.Sprintf("file not found: %s", path),
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "file",
		},
		Path:      path,
		Operation: "read",
	}
}

// NewFileReadError creates a file read error.
func NewFileReadError(path string, cause error) *FileError {
	return &FileError{
		PEError: &PEError{
			Code:      ErrCodeFileRead,
			Message:   fmt.Sprintf("failed to read file: %s", path),
			Cause:     cause,
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "file",
		},
		Path:      path,
		Operation: "read",
	}
}

// NewFileWriteError creates a file write error.
func NewFileWriteError(path string, cause error) *FileError {
	return &FileError{
		PEError: &PEError{
			Code:      ErrCodeFileWrite,
			Message:   fmt.Sprintf("failed to write file: %s", path),
			Cause:     cause,
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "file",
		},
		Path:      path,
		Operation: "write",
	}
}

// NewFileFormatError creates a file format error.
func NewFileFormatError(path, format string, cause error) *FileError {
	return &FileError{
		PEError: &PEError{
			Code:      ErrCodeFileFormat,
			Message:   fmt.Sprintf("invalid %s format in file: %s", format, path),
			Cause:     cause,
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "file",
		},
		Path: path,
	}
}

// WithSize adds file size information to the file error.
func (e *FileError) WithSize(size int64) *FileError {
	e.Size = size
	return e
}

// ModuleError represents module-related errors.
type ModuleError struct {
	*PEError
	ModuleName    string `json:"module_name,omitempty"`
	Version       string `json:"version,omitempty"`
	RepositoryURL string `json:"repository_url,omitempty"`
}

// NewModuleNotFoundError creates a module not found error.
func NewModuleNotFoundError(moduleName string) *ModuleError {
	return &ModuleError{
		PEError: &PEError{
			Code:      ErrCodeModuleNotFound,
			Message:   fmt.Sprintf("module not found: %s", moduleName),
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "module",
		},
		ModuleName: moduleName,
	}
}

// NewModuleInvalidError creates a module invalid error.
func NewModuleInvalidError(moduleName, reason string) *ModuleError {
	return &ModuleError{
		PEError: &PEError{
			Code:      ErrCodeModuleInvalid,
			Message:   fmt.Sprintf("invalid module %s: %s", moduleName, reason),
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "module",
		},
		ModuleName: moduleName,
	}
}

// NewModuleDependencyError creates a module dependency error.
func NewModuleDependencyError(moduleName, dependency string) *ModuleError {
	return &ModuleError{
		PEError: &PEError{
			Code:      ErrCodeModuleDependency,
			Message:   fmt.Sprintf("module %s has unresolved dependency: %s", moduleName, dependency),
			Severity:  SeverityMedium,
			Retryable: false,
			Component: "module",
		},
		ModuleName: moduleName,
	}
}

// WithVersion adds version information to the module error.
func (e *ModuleError) WithVersion(version string) *ModuleError {
	e.Version = version
	return e
}

// WithRepositoryURL adds repository URL to the module error.
func (e *ModuleError) WithRepositoryURL(url string) *ModuleError {
	e.RepositoryURL = url
	return e
}

// SecurityError represents security-related errors.
type SecurityError struct {
	*PEError
	SecurityCheck string `json:"security_check,omitempty"`
	RiskLevel     string `json:"risk_level,omitempty"`
	Policy        string `json:"policy,omitempty"`
}

// NewSecurityValidationError creates a security validation error.
func NewSecurityValidationError(check, message string) *SecurityError {
	return &SecurityError{
		PEError: &PEError{
			Code:      ErrCodeSecurityValidation,
			Message:   message,
			Severity:  SeverityCritical,
			Retryable: false,
			Component: "security",
		},
		SecurityCheck: check,
		RiskLevel:     "HIGH",
	}
}

// NewSecurityAuthError creates a security authentication error.
func NewSecurityAuthError(message string) *SecurityError {
	return &SecurityError{
		PEError: &PEError{
			Code:      ErrCodeSecurityAuth,
			Message:   message,
			Severity:  SeverityHigh,
			Retryable: false,
			Component: "security",
		},
		RiskLevel: "HIGH",
	}
}

// NewSecurityPolicyError creates a security policy violation error.
func NewSecurityPolicyError(policy, message string) *SecurityError {
	return &SecurityError{
		PEError: &PEError{
			Code:      ErrCodeSecurityPolicy,
			Message:   message,
			Severity:  SeverityHigh,
			Retryable: false,
			Component: "security",
		},
		Policy:    policy,
		RiskLevel: "HIGH",
	}
}

// WithRiskLevel sets the risk level for the security error.
func (e *SecurityError) WithRiskLevel(level string) *SecurityError {
	e.RiskLevel = level
	return e
}