// +build ignore

// Example integration showing how to use the PE errors package
// This file demonstrates how existing code can be migrated to use
// the standardized error handling system.
package main

import (
	"fmt"
	"os"

	"github.com/tmc/pe/internal/errors"
)

// Example of migrating from fmt.Errorf to PE errors
func exampleOldErrorHandling(filename string) error {
	_, err := os.Open(filename)
	if err != nil {
		// Old way - less structured
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	return nil
}

// Example using PE errors - more structured and informative
func exampleNewErrorHandling(filename string) error {
	_, err := os.Open(filename)
	if err != nil {
		// New way - structured with error codes and context
		return errors.WrapFile(err, filename, "read").
			WithComponent("file-reader").
			WithContext("operation", "config-load")
	}
	return nil
}

// Example provider error handling
func exampleProviderError(provider, model string) error {
	// Simulate a provider API error
	apiErr := fmt.Errorf("API returned 429: rate limit exceeded")
	
	// Wrap with provider-specific context
	return errors.WrapProvider(apiErr, provider, model).
		WithStatusCode(429).
		WithRequestID("req-123456")
}

// Example inference error handling
func exampleInferenceError(model string) error {
	// Simulate a context length error
	contextErr := fmt.Errorf("context length 8192 exceeds maximum 4096")
	
	return errors.WrapInference(contextErr, model).
		WithTokensUsed(8192).
		WithContext("max_context", 4096)
}

// Example showing error analysis
func exampleErrorAnalysis(err error) {
	if err == nil {
		return
	}
	
	fmt.Printf("Error Code: %s\n", errors.GetCode(err))
	fmt.Printf("Severity: %s\n", errors.GetSeverity(err))
	fmt.Printf("Retryable: %t\n", errors.IsRetryable(err))
	fmt.Printf("Component: %s\n", errors.GetComponent(err))
	
	if context := errors.GetContext(err); context != nil {
		fmt.Printf("Context: %v\n", context)
	}
	
	// Check specific error types
	if errors.IsCode(err, errors.ErrCodeProviderRateLimit) {
		fmt.Println("This is a rate limit error - should retry with backoff")
	}
}

func main() {
	// Example usage
	
	// File error example
	if err := exampleNewErrorHandling("nonexistent.txt"); err != nil {
		fmt.Println("File error:", err)
		exampleErrorAnalysis(err)
		fmt.Println()
	}
	
	// Provider error example  
	if err := exampleProviderError("openai", "gpt-4"); err != nil {
		fmt.Println("Provider error:", err)
		exampleErrorAnalysis(err)
		fmt.Println()
	}
	
	// Inference error example
	if err := exampleInferenceError("gpt-4"); err != nil {
		fmt.Println("Inference error:", err)
		exampleErrorAnalysis(err)
		fmt.Println()
	}
	
	// Chain multiple errors
	err1 := errors.New(errors.ErrCodeFileRead, "failed to read config")
	err2 := errors.New(errors.ErrCodeInvalidConfig, "invalid yaml format")
	
	chainedErr := errors.Chain(err1, err2)
	fmt.Println("Chained error:", chainedErr)
	exampleErrorAnalysis(chainedErr)
}