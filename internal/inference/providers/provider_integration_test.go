package providers

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tmc/pe/internal/inference"
	_ "github.com/tmc/pe/internal/inference/providers/anthropic"
	_ "github.com/tmc/pe/internal/inference/providers/cgpt"
	_ "github.com/tmc/pe/internal/inference/providers/ollama"
	_ "github.com/tmc/pe/internal/inference/providers/openai"
)

// TestProviderFactories tests that all provider factories work correctly
func TestProviderFactories(t *testing.T) {
	// Set test mode to avoid actual API calls
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name     string
		provider string
		config   map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "openai_with_api_key",
			provider: "openai",
			config: map[string]interface{}{
				"api_key": "test-key",
			},
			wantErr: false,
		},
		{
			name:     "openai_with_env_api_key",
			provider: "openai",
			config:   map[string]interface{}{},
			wantErr:  false, // May succeed if OPENAI_API_KEY env var is set
		},
		{
			name:     "anthropic_with_api_key",
			provider: "anthropic",
			config: map[string]interface{}{
				"api_key": "test-key",
			},
			wantErr: false,
		},
		{
			name:     "ollama_default",
			provider: "ollama",
			config:   nil,
			wantErr:  false,
		},
		{
			name:     "cgpt_default",
			provider: "cgpt",
			config:   nil,
			wantErr:  false,
		},
		{
			name:     "unknown_provider",
			provider: "unknown",
			config:   nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := inference.GetFactory(tt.provider)
			if factory == nil && !tt.wantErr {
				t.Errorf("Expected factory for provider %s, but got nil", tt.provider)
				return
			}
			if factory == nil && tt.wantErr {
				return // Expected error case
			}

			provider, err := factory(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Factory error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if provider != nil {
				defer provider.Close()

				// Test basic functionality
				name := provider.Name()
				if name == "" {
					t.Error("Provider name should not be empty")
				}

				// Test that name matches expected provider
				if !strings.Contains(strings.ToLower(name), tt.provider) {
					t.Errorf("Provider name %s should contain %s", name, tt.provider)
				}
			}
		})
	}
}

// TestProviderAuthenticationHandling tests authentication scenarios
func TestProviderAuthenticationHandling(t *testing.T) {
	// Skip in short mode as this may involve network calls
	if testing.Short() {
		t.Skip("Skipping authentication tests in short mode")
	}

	tests := []struct {
		name         string
		provider     string
		setupAuth    func() (cleanup func())
		expectError  bool
		errorMessage string
	}{
		{
			name:     "openai_missing_api_key",
			provider: "openai",
			setupAuth: func() func() {
				old := os.Getenv("OPENAI_API_KEY")
				os.Unsetenv("OPENAI_API_KEY")
				return func() {
					if old != "" {
						os.Setenv("OPENAI_API_KEY", old)
					}
				}
			},
			expectError:  true,
			errorMessage: "API key not provided",
		},
		{
			name:     "openai_invalid_api_key",
			provider: "openai",
			setupAuth: func() func() {
				old := os.Getenv("OPENAI_API_KEY")
				os.Setenv("OPENAI_API_KEY", "invalid-key")
				return func() {
					if old != "" {
						os.Setenv("OPENAI_API_KEY", old)
					} else {
						os.Unsetenv("OPENAI_API_KEY")
					}
				}
			},
			expectError:  false, // Factory should succeed, but API calls will fail
			errorMessage: "",
		},
		{
			name:     "anthropic_missing_api_key",
			provider: "anthropic",
			setupAuth: func() func() {
				old := os.Getenv("ANTHROPIC_API_KEY")
				os.Unsetenv("ANTHROPIC_API_KEY")
				return func() {
					if old != "" {
						os.Setenv("ANTHROPIC_API_KEY", old)
					}
				}
			},
			expectError:  true,
			errorMessage: "API key not provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupAuth()
			defer cleanup()

			factory := inference.GetFactory(tt.provider)
			if factory == nil {
				t.Skipf("Provider %s not available", tt.provider)
			}

			provider, err := factory(nil)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message to contain %q, got: %s", tt.errorMessage, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if provider != nil {
				defer provider.Close()
			}
		})
	}
}

// TestProviderConfiguration tests various configuration scenarios
func TestProviderConfiguration(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name     string
		provider string
		config   map[string]interface{}
		validate func(t *testing.T, provider inference.Provider)
	}{
		{
			name:     "openai_custom_base_url",
			provider: "openai",
			config: map[string]interface{}{
				"api_key":  "test-key",
				"base_url": "https://custom.openai.com/v1",
			},
			validate: func(t *testing.T, provider inference.Provider) {
				// In real implementation, could check internal configuration
				if provider.Name() != "openai" {
					t.Errorf("Expected openai provider, got %s", provider.Name())
				}
			},
		},
		{
			name:     "ollama_custom_host",
			provider: "ollama",
			config: map[string]interface{}{
				"host": "http://localhost:11434",
			},
			validate: func(t *testing.T, provider inference.Provider) {
				if provider.Name() != "ollama" {
					t.Errorf("Expected ollama provider, got %s", provider.Name())
				}
			},
		},
		{
			name:     "anthropic_custom_version",
			provider: "anthropic",
			config: map[string]interface{}{
				"api_key":     "test-key",
				"api_version": "2023-06-01",
			},
			validate: func(t *testing.T, provider inference.Provider) {
				if provider.Name() != "anthropic" {
					t.Errorf("Expected anthropic provider, got %s", provider.Name())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := inference.GetFactory(tt.provider)
			if factory == nil {
				t.Skipf("Provider %s not available", tt.provider)
			}

			provider, err := factory(tt.config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("Provider should not be nil")
				return
			}
			defer provider.Close()

			tt.validate(t, provider)
		})
	}
}

// TestProviderErrorHandling tests error scenarios across providers
func TestProviderErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling tests in short mode")
	}

	// Set test mode but allow some real provider testing
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	providers := []string{"openai", "anthropic", "ollama", "cgpt"}

	for _, providerName := range providers {
		t.Run(providerName, func(t *testing.T) {
			factory := inference.GetFactory(providerName)
			if factory == nil {
				t.Skipf("Provider %s not available", providerName)
			}

			// Create provider with test configuration
			config := map[string]interface{}{
				"api_key": "test-key", // For providers that need it
			}

			provider, err := factory(config)
			if err != nil {
				t.Skipf("Could not create provider %s: %v", providerName, err)
			}
			defer provider.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Test with invalid/empty request
			req := inference.Request{
				Prompt: "", // Empty prompt
				Model:  "invalid-model",
			}

			// Complete should handle gracefully
			_, err = provider.Complete(ctx, req)
			// We don't assert specific error because providers handle differently
			// Just ensure it doesn't panic

			// Stream should handle gracefully
			stream, err := provider.Stream(ctx, req)
			if stream != nil {
				// Consume stream to avoid goroutine leaks
				for range stream {
					// Just consume
				}
			}

			// Models should work even with invalid config
			models, err := provider.Models(ctx)
			// Some providers might return empty list or error, both are acceptable
			_ = models
		})
	}
}

// TestProviderConcurrency tests concurrent access to providers
func TestProviderConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency tests in short mode")
	}

	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	// Test with cgpt provider as it's most likely to be available
	factory := inference.GetFactory("cgpt")
	if factory == nil {
		t.Skip("cgpt provider not available")
	}

	provider, err := factory(nil)
	if err != nil {
		t.Skipf("Could not create cgpt provider: %v", err)
	}
	defer provider.Close()

	const numGoroutines = 5
	const numRequestsPerGoroutine = 3

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results := make(chan error, numGoroutines*numRequestsPerGoroutine)

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numRequestsPerGoroutine; j++ {
				req := inference.Request{
					Prompt: "Test concurrent request",
					Model:  "gpt-4",
				}

				_, err := provider.Complete(ctx, req)
				results <- err
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines*numRequestsPerGoroutine; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Logf("Concurrent request failed (acceptable in test mode): %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Error("Timeout waiting for concurrent requests")
			return
		}
	}
}

// TestProviderCleanup tests proper resource cleanup
func TestProviderCleanup(t *testing.T) {
	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	providers := []string{"openai", "anthropic", "ollama", "cgpt"}

	for _, providerName := range providers {
		t.Run(providerName, func(t *testing.T) {
			factory := inference.GetFactory(providerName)
			if factory == nil {
				t.Skipf("Provider %s not available", providerName)
			}

			config := map[string]interface{}{
				"api_key": "test-key",
			}

			provider, err := factory(config)
			if err != nil {
				t.Skipf("Could not create provider %s: %v", providerName, err)
			}

			// Test Close method
			err = provider.Close()
			if err != nil {
				t.Errorf("Close() failed: %v", err)
			}

			// Test that operations fail after close (if implemented)
			ctx := context.Background()
			_, err = provider.Complete(ctx, inference.Request{
				Prompt: "test after close",
			})
			// Some providers might not enforce post-close errors, so we don't assert
		})
	}
}

// TestProviderModels tests model enumeration
func TestProviderModels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping models test in short mode")
	}

	// Set test mode
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	providers := []string{"cgpt"} // Start with cgpt as it's most likely to work

	for _, providerName := range providers {
		t.Run(providerName, func(t *testing.T) {
			factory := inference.GetFactory(providerName)
			if factory == nil {
				t.Skipf("Provider %s not available", providerName)
			}

			provider, err := factory(nil)
			if err != nil {
				t.Skipf("Could not create provider %s: %v", providerName, err)
			}
			defer provider.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			models, err := provider.Models(ctx)
			if err != nil {
				t.Logf("Models() failed for %s (acceptable in test mode): %v", providerName, err)
				return
			}

			if len(models) == 0 {
				t.Logf("No models returned for %s (may be expected in test mode)", providerName)
			}

			// Validate model names are not empty
			for i, model := range models {
				if model == "" {
					t.Errorf("Model %d has empty name", i)
				}
			}
		})
	}
}