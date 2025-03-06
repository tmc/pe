// Package providers manages the registration and initialization of LLM providers.
package providers

import (
	"fmt"
	"sync"

	"github.com/tmc/pe/internal/llm"
)

var (
	// registry is a map of provider names to their initializer functions.
	registry = make(map[string]ProviderInitFunc)
	
	// instances stores initialized provider instances.
	instances = make(map[string]llm.Provider)
	
	// mu protects access to the registry and instances maps.
	mu sync.RWMutex
)

// ProviderInitFunc is a function that initializes a provider with the given configuration.
type ProviderInitFunc func(config map[string]interface{}) (llm.Provider, error)

// Register adds a provider initializer to the registry.
func Register(name string, initFn ProviderInitFunc) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = initFn
}

// Initialize creates an instance of a provider with the given configuration.
func Initialize(name string, config map[string]interface{}) (llm.Provider, error) {
	mu.RLock()
	initFn, exists := registry[name]
	mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("provider %q not registered", name)
	}
	
	provider, err := initFn(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider %q: %w", name, err)
	}
	
	mu.Lock()
	instances[name] = provider
	mu.Unlock()
	
	return provider, nil
}

// Get returns an initialized provider by name.
func Get(name string) (llm.Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	
	provider, exists := instances[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not initialized", name)
	}
	
	return provider, nil
}

// List returns all registered provider names.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	
	var names []string
	for name := range registry {
		names = append(names, name)
	}
	
	return names
}

// ListInitialized returns all initialized provider names.
func ListInitialized() []string {
	mu.RLock()
	defer mu.RUnlock()
	
	var names []string
	for name := range instances {
		names = append(names, name)
	}
	
	return names
}

// Reset clears all initialized providers.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	instances = make(map[string]llm.Provider)
}

// init registers all built-in providers.
func init() {
	// Register built-in providers
	Register("openai", NewOpenAIProvider)
	Register("anthropic", NewAnthropicProvider)
	Register("googleai", NewGoogleAIProvider)
}