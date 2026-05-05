package llm

import (
	"fmt"
	"os"
	"strings"
)

// GetProvider returns a Provider for the given backend
func GetProvider(backend string) (Provider, error) {
	return GetProviderWithOptions(backend, nil)
}

// GetProviderWithOptions returns a Provider for the given backend and provider config.
func GetProviderWithOptions(backend string, options map[string]interface{}) (Provider, error) {
	// Auto-detect provider if backend is empty
	if backend == "" {
		backend = detectDefaultProvider()
		if backend == "" {
			return nil, fmt.Errorf("no provider specified and none could be auto-detected (check OPENAI_API_KEY or ANTHROPIC_API_KEY environment variables)")
		}
	}

	// Check if we should use mock provider in test mode
	if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
		// Use the registered mock provider from providers package
		// This avoids duplication and uses the proper mock implementation
		if backend == "" || backend == "mock" {
			backend = "mock"
		}
	}

	// Parse provider:model format
	provider := backend
	model := ""
	if idx := strings.IndexByte(backend, ':'); idx != -1 {
		provider = backend[:idx]
		model = backend[idx+1:]
	}

	switch provider {
	case "mock":
		// Use native mock provider through factory
		provider, err := CreateNativeProviderFromFactory("mock", options)
		if err != nil {
			// Fall back to local mock proxy if factory fails
			return &MockProviderProxy{}, nil
		}
		return NewConfiguredProvider(provider, options), nil
	case "openai", "anthropic":
		// Use native providers through the factory system
		provider, err := CreateNativeProviderFromFactory(backend, options)
		if err != nil {
			return nil, fmt.Errorf("failed to create native provider: %w", err)
		}
		return NewConfiguredProvider(provider, options), nil
	case "cgpt":
		// Check if we can use native providers instead of cgpt
		if os.Getenv("PE_USE_NATIVE_PROVIDERS") == "true" || os.Getenv("PE_FORCE_NATIVE") == "true" {
			// Try to auto-detect and use native provider
			if os.Getenv("OPENAI_API_KEY") != "" {
				return CreateNativeProviderFromFactory("openai:gpt-4", options)
			} else if os.Getenv("ANTHROPIC_API_KEY") != "" {
				return CreateNativeProviderFromFactory("anthropic:claude-3-sonnet-20240229", options)
			}
		}
		// Fall back to CGPT wrapper
		return &CGPTProvider{
			Backend: provider,
			model:   model,
		}, nil
	case "gemini", "googleai":
		// Still use CGPT for providers not yet migrated
		return &CGPTProvider{
			Backend: provider,
			model:   model,
		}, nil
	default:
		// Try to create as native provider first
		provider, err := CreateNativeProviderFromFactory(backend, options)
		if err == nil {
			return NewConfiguredProvider(provider, options), nil
		}
		// Fall back to CGPT if it exists
		if _, err := os.Stat("/usr/local/bin/cgpt"); err == nil {
			return &CGPTProvider{
				Backend: backend,
				model:   model,
			}, nil
		}
		// Fall back to error
		return nil, fmt.Errorf("unsupported backend: %s", backend)
	}
}

// detectDefaultProvider auto-detects the default provider based on available API keys
func detectDefaultProvider() string {
	// Check for OpenAI API key
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai:gpt-4"
	}

	// Check for Anthropic API key
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "anthropic:claude-3-sonnet-20240229"
	}

	// Fall back to empty string (no provider detected)
	return ""
}

// CreateNativeProvider creates a native provider using the providers package
// This function will be used to transition away from CGPT dependency
func CreateNativeProvider(providerSpec string, options map[string]interface{}) (Provider, error) {
	// Try to create provider using factory
	provider, err := CreateNativeProviderFromFactory(providerSpec, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create native provider %s: %w", providerSpec, err)
	}
	return provider, nil
}
