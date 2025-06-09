package llm

import (
	"fmt"
	"strings"
	"sync"
)

// ProviderFactoryFunc is a function that creates a provider
type ProviderFactoryFunc func(providerSpec string, options map[string]interface{}) (Provider, error)

// providerFactoryRegistry holds registered provider factories
var (
	providerFactoryMu sync.RWMutex
	providerFactories = make(map[string]ProviderFactoryFunc)
)

// RegisterProviderFactory registers a provider factory
func RegisterProviderFactory(name string, factory ProviderFactoryFunc) {
	providerFactoryMu.Lock()
	defer providerFactoryMu.Unlock()
	providerFactories[name] = factory
}

// GetProviderFactory returns a registered provider factory
func GetProviderFactory(name string) (ProviderFactoryFunc, bool) {
	providerFactoryMu.RLock()
	defer providerFactoryMu.RUnlock()
	factory, ok := providerFactories[name]
	return factory, ok
}

// CreateNativeProviderFromFactory creates a provider using registered factories
func CreateNativeProviderFromFactory(providerSpec string, options map[string]interface{}) (Provider, error) {
	// Parse provider name from spec
	providerName := providerSpec
	if idx := strings.IndexByte(providerSpec, ':'); idx != -1 {
		providerName = providerSpec[:idx]
	}

	// Get factory
	factory, ok := GetProviderFactory(providerName)
	if !ok {
		return nil, fmt.Errorf("no factory registered for provider: %s", providerName)
	}

	// Create provider
	return factory(providerSpec, options)
}
