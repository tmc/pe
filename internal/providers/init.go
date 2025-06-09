package providers

import (
	"github.com/tmc/pe/internal/llm"
)

func init() {
	// Register OpenAI provider factory
	llm.RegisterProviderFactory("openai", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := NewOpenAIProvider(providerSpec, options)
		if err != nil {
			return nil, err
		}
		// Use the provider directly without adapter since it already implements llm.Provider
		return provider, nil
	})

	// Register Anthropic provider factory
	llm.RegisterProviderFactory("anthropic", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := NewAnthropicProvider(providerSpec, options)
		if err != nil {
			return nil, err
		}
		// Use the provider directly without adapter since it already implements llm.Provider
		return provider, nil
	})

	// Register Mock provider factory (for testing)
	llm.RegisterProviderFactory("mock", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := NewMockProvider(providerSpec, options)
		if err != nil {
			return nil, err
		}
		// Use the provider directly without adapter since it already implements llm.Provider
		return provider, nil
	})
}
