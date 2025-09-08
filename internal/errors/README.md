# PE Errors Package

A comprehensive error handling system for the PE (Prompt Engineering) toolkit, providing structured error types, error codes, and contextual information for better debugging and error recovery.

## Features

- **Structured Error Types**: Domain-specific error types (Provider, Inference, Optimization, etc.)
- **Error Codes**: Standardized error codes for consistent error categorization
- **Contextual Information**: Rich context including severity, retryability, and component information
- **Error Wrapping**: Support for error chaining and unwrapping (Go 1.13+ compatible)
- **Automatic Analysis**: Intelligent error analysis to determine error codes and characteristics

## Core Types

### PEError
The base error type that provides:
- Error codes for categorization
- Severity levels (Low, Medium, High, Critical)
- Retryability information
- Component identification
- Contextual metadata

### Domain-Specific Errors
- **ProviderError**: LLM provider-related errors (auth, rate limits, timeouts)
- **InferenceError**: Inference operation errors (context length, failures)
- **OptimizationError**: Prompt optimization errors (convergence, timeouts)
- **EvaluationError**: Evaluation and assertion errors
- **FileError**: File I/O related errors
- **ModuleError**: Module management errors
- **SecurityError**: Security validation errors

## Usage Examples

### Basic Error Creation
```go
// Create a simple error
err := errors.New(errors.ErrCodeInvalidInput, "invalid prompt format")

// Create with formatted message
err := errors.Newf(errors.ErrCodeFileRead, "failed to read %s", filename)
```

### Error Wrapping
```go
// Wrap existing errors with context
if err := someOperation(); err != nil {
    return errors.WrapProvider(err, "openai", "gpt-4")
}

// Wrap file operations
if err := os.Open(filename); err != nil {
    return errors.WrapFile(err, filename, "read")
}
```

### Adding Context
```go
err := errors.New(errors.ErrCodeInferenceFailed, "model request failed")
return err.WithContext("model", "gpt-4").
    WithContext("tokens", 1024).
    WithComponent("inference-engine")
```

### Error Analysis
```go
if errors.IsCode(err, errors.ErrCodeProviderRateLimit) {
    // Handle rate limit specifically
    if errors.IsRetryable(err) {
        // Retry with backoff
    }
}

severity := errors.GetSeverity(err)
component := errors.GetComponent(err)
context := errors.GetContext(err)
```

## Error Codes

### Core System Errors
- `ErrCodeUnknown`: Unknown or unclassified errors
- `ErrCodeInternal`: Internal system errors
- `ErrCodeInvalidInput`: Invalid user input
- `ErrCodeInvalidConfig`: Invalid configuration
- `ErrCodeNotFound`: Resource not found
- `ErrCodeNotImplemented`: Feature not implemented

### Provider Errors
- `ErrCodeProviderUnavailable`: Provider service unavailable
- `ErrCodeProviderAuth`: Authentication/authorization errors
- `ErrCodeProviderQuota`: Quota or billing errors
- `ErrCodeProviderRateLimit`: Rate limit exceeded
- `ErrCodeProviderTimeout`: Request timeout
- `ErrCodeProviderAPI`: General API errors

### Inference Errors
- `ErrCodeInferenceTimeout`: Inference request timeout
- `ErrCodeInferenceFailed`: General inference failure
- `ErrCodeInferenceInvalid`: Invalid inference request
- `ErrCodeInferenceContextLen`: Context length exceeded

## Best Practices

1. **Use Specific Error Codes**: Choose the most specific error code available
2. **Add Context**: Include relevant context information for debugging
3. **Set Components**: Identify the component that generated the error
4. **Wrap Don't Replace**: Wrap existing errors to preserve the original cause
5. **Check Retryability**: Use the retryable flag to implement retry logic
6. **Handle by Severity**: Different handling strategies based on severity

## Integration

The errors package is designed to be a drop-in replacement for standard error handling:

```go
// Before
return fmt.Errorf("provider %s failed: %w", provider, err)

// After
return errors.WrapProvider(err, provider, model)
```

## Testing

The package includes comprehensive tests covering:
- Basic error creation and manipulation
- Error wrapping and unwrapping
- Domain-specific error types
- Error analysis and classification
- Utility functions

Run tests with:
```bash
go test ./internal/errors -v
```

## Coverage

Current test coverage: ~30% (focused on core functionality)
Priority areas for additional testing:
- Edge cases in error analysis
- Complex error chaining scenarios
- Integration with Go standard library errors