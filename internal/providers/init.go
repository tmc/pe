package providers

import (
	"github.com/tmc/pe/internal/llm"
)

func init() {
	// Register OpenAI provider factory
	llm.RegisterProviderFactory("openai", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		// The providerSpec here is just the model name, not the full spec
		// The provider name has already been extracted by the caller
		model := providerSpec
		provider, err := NewOpenAIProvider(model, options)
		if err != nil {
			return nil, err
		}
		// Use the provider directly without adapter since it already implements llm.Provider
		return provider, nil
	})

	// Register Anthropic provider factory
	llm.RegisterProviderFactory("anthropic", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		// The providerSpec here is just the model name, not the full spec
		// The provider name has already been extracted by the caller
		model := providerSpec
		provider, err := NewAnthropicProvider(model, options)
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
