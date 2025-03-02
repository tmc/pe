package llm

import (
	"fmt"
	"strings"
	"sync"
)

// Registry manages providers and factories.
type Registry struct {
	providers map[string]Provider
	factories []ProviderFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		factories: make([]ProviderFactory, 0),
	}
}

// DefaultRegistry is the default registry for the application.
var DefaultRegistry = NewRegistry()

// RegisterProvider registers a provider with the registry.
func (r *Registry) RegisterProvider(provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.providers[strings.ToLower(provider.Name())] = provider
}

// RegisterFactory registers a provider factory with the registry.
func (r *Registry) RegisterFactory(factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.factories = append(r.factories, factory)
}

// GetProvider returns a provider by name, creating it if necessary.
func (r *Registry) GetProvider(name string) (Provider, error) {
	name = strings.ToLower(name)
	
	// First, check if the provider is already registered
	r.mu.RLock()
	provider, exists := r.providers[name]
	r.mu.RUnlock()
	
	if exists {
		return provider, nil
	}
	
	// If not, try to create it using factories
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check again in case another goroutine registered it
	if provider, exists = r.providers[name]; exists {
		return provider, nil
	}
	
	// Try to create it
	for _, factory := range r.factories {
		if factory.SupportsProvider(name) {
			provider, err := factory.Create(name)
			if err != nil {
				return nil, err
			}
			
			// Register the provider
			r.providers[strings.ToLower(provider.Name())] = provider
			return provider, nil
		}
	}
	
	return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
}

// GetModelByFullName returns a model by its full name (provider:model).
func (r *Registry) GetModelByFullName(fullName string) (Model, error) {
	providerName, modelName := ParseProviderAndModel(fullName)
	
	provider, err := r.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	
	if modelName == "" {
		return provider.DefaultModel(), nil
	}
	
	return provider.GetModel(modelName)
}

// ListProviders returns a list of registered providers.
func (r *Registry) ListProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	
	return providers
}

// ListModels returns a list of all available models across all providers.
func (r *Registry) ListModels() []Model {
	var models []Model
	
	for _, provider := range r.ListProviders() {
		models = append(models, provider.Models()...)
	}
	
	return models
}

// Global convenience functions that use DefaultRegistry

// RegisterProvider registers a provider with the default registry.
func RegisterProvider(provider Provider) {
	DefaultRegistry.RegisterProvider(provider)
}

// RegisterFactory registers a provider factory with the default registry.
func RegisterFactory(factory ProviderFactory) {
	DefaultRegistry.RegisterFactory(factory)
}

// GetProvider returns a provider by name from the default registry.
func GetProvider(name string) (Provider, error) {
	return DefaultRegistry.GetProvider(name)
}

// GetModelByFullName returns a model by its full name from the default registry.
func GetModelByFullName(fullName string) (Model, error) {
	return DefaultRegistry.GetModelByFullName(fullName)
}

// ListProviders returns a list of registered providers from the default registry.
func ListProviders() []Provider {
	return DefaultRegistry.ListProviders()
}

// ListModels returns a list of all available models across all providers
// from the default registry.
func ListModels() []Model {
	return DefaultRegistry.ListModels()
}