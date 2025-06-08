package providers

import (
	"fmt"
	"sync"

	"github.com/tmc/pe/internal/llm"
)

// ProviderFactory is a function that creates a new provider instance
type ProviderFactory func(model string, options map[string]interface{}) (llm.Provider, error)

// Registry manages provider registration and creation
type Registry struct {
	factories map[string]ProviderFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	r := &Registry{
		factories: make(map[string]ProviderFactory),
	}

	// Register built-in providers
	r.registerBuiltinProviders()

	return r
}

// registerBuiltinProviders registers the built-in providers
func (r *Registry) registerBuiltinProviders() {
	r.Register("openai", func(model string, options map[string]interface{}) (llm.Provider, error) {
		return NewOpenAIProvider(model, options)
	})

	r.Register("anthropic", func(model string, options map[string]interface{}) (llm.Provider, error) {
		return NewAnthropicProvider(model, options)
	})

	// Register mock provider for testing
	r.Register("mock", func(model string, options map[string]interface{}) (llm.Provider, error) {
		return NewMockProvider(model, options)
	})
}

// Register registers a new provider factory
func (r *Registry) Register(name string, factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Create creates a new provider instance
func (r *Registry) Create(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
	providerName, model, err := ParseProviderString(providerSpec)
	if err != nil {
		return nil, err
	}

	// Normalize the model name
	model = NormalizeModelName(providerName, model)

	r.mu.RLock()
	factory, exists := r.factories[providerName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	return factory(model, options)
}

// List returns a list of available provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []string
	for name := range r.factories {
		providers = append(providers, name)
	}

	return providers
}

// Exists checks if a provider exists in the registry
func (r *Registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.factories[name]
	return exists
}

// GetSupportedModels returns a list of commonly supported models for each provider
func (r *Registry) GetSupportedModels() map[string][]string {
	return map[string][]string{
		"openai": {
			"gpt-4",
			"gpt-4-turbo",
			"gpt-4-turbo-preview",
			"gpt-4-0125-preview",
			"gpt-4-1106-preview",
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-0125",
			"gpt-3.5-turbo-instruct",
		},
		"anthropic": {
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
			"claude-2.1",
			"claude-2.0",
			"claude-instant-1.2",
		},
	}
}

// ValidateProvider validates a provider specification
func (r *Registry) ValidateProvider(providerSpec string) error {
	providerName, model, err := ParseProviderString(providerSpec)
	if err != nil {
		return err
	}

	if !r.Exists(providerName) {
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	// Validate model is supported (optional check)
	supportedModels := r.GetSupportedModels()
	if models, exists := supportedModels[providerName]; exists {
		normalizedModel := NormalizeModelName(providerName, model)
		for _, supportedModel := range models {
			if supportedModel == normalizedModel {
				return nil
			}
		}
		// Don't return error for unsupported models, just warn
		// This allows for newer models that haven't been added to the list yet
	}

	return nil
}

// Default registry instance
var defaultRegistry = NewRegistry()

// CreateProvider creates a provider using the default registry
func CreateProvider(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
	return defaultRegistry.Create(providerSpec, options)
}

// RegisterProvider registers a provider in the default registry
func RegisterProvider(name string, factory ProviderFactory) {
	defaultRegistry.Register(name, factory)
}

// ListProviders returns available providers from the default registry
func ListProviders() []string {
	return defaultRegistry.List()
}

// ValidateProviderSpec validates a provider specification using the default registry
func ValidateProviderSpec(providerSpec string) error {
	return defaultRegistry.ValidateProvider(providerSpec)
}

// GetSupportedModelsMap returns supported models from the default registry
func GetSupportedModelsMap() map[string][]string {
	return defaultRegistry.GetSupportedModels()
}
