// Package inference provides a provider registry for inference providers.
package inference

import (
	"fmt"
	"sync"
)

// ProviderFactory is a function that creates a new provider instance.
type ProviderFactory func(config map[string]interface{}) (Provider, error)

// Registry manages provider factories.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

// globalRegistry is the default registry instance.
var globalRegistry = &Registry{
	factories: make(map[string]ProviderFactory),
}

// Register adds a provider factory to the global registry.
func Register(name string, factory ProviderFactory) {
	globalRegistry.Register(name, factory)
}

// NewProvider creates a provider from the global registry.
func NewProvider(name string, config map[string]interface{}) (Provider, error) {
	return globalRegistry.NewProvider(name, config)
}

// Providers returns all registered provider names from the global registry.
func Providers() []string {
	return globalRegistry.Providers()
}

// Register adds a provider factory to the registry.
func (r *Registry) Register(name string, factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// NewProvider creates a provider instance using a registered factory.
func (r *Registry) NewProvider(name string, config map[string]interface{}) (Provider, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider %q not registered", name)
	}

	return factory(config)
}

// Providers returns all registered provider names.
func (r *Registry) Providers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// DefaultClient creates a client with all registered providers.
func DefaultClient() (*Client, error) {
	client := NewClient()

	// Register all available providers with default configs
	for _, name := range Providers() {
		provider, err := NewProvider(name, nil)
		if err != nil {
			// Skip providers that fail with nil config
			continue
		}
		client.Register(name, provider)
	}

	return client, nil
}

// MustRegister is like Register but panics on error.
// It's intended for use in init functions.
func MustRegister(name string, factory ProviderFactory) {
	if err := validateProviderName(name); err != nil {
		panic(fmt.Sprintf("invalid provider name %q: %v", name, err))
	}
	Register(name, factory)
}

// validateProviderName checks if a provider name is valid.
func validateProviderName(name string) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	// Add more validation as needed
	return nil
}
