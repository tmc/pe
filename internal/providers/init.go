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

	// Register cgpt provider factory
	llm.RegisterProviderFactory("cgpt", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		// cgpt doesn't need a model specification
		provider, err := NewCGPTProvider(options)
		if err != nil {
			return nil, err
		}
		return provider, nil
	})

	// Register generic CLI provider factory
	llm.RegisterProviderFactory("cli", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := NewGenericCLIProvider(providerSpec, options)
		if err != nil {
			return nil, err
		}
		return provider, nil
	})

	// Helper to register presets
	registerPreset := func(name string) {
		llm.RegisterProviderFactory(name, func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
			// Get preset configuration
			config, err := GetPresetConfig(name, options)
			if err != nil {
				return nil, err
			}

			// Create generic provider with preset config
			provider, err := NewGenericCLIProvider(providerSpec, config)
			if err != nil {
				return nil, err
			}
			return provider, nil
		})
	}

	// Register presets
	registerPreset("ollama")
	registerPreset("mlx")
	registerPreset("mlx-lm")
	registerPreset("mlx-go")
	registerPreset("llama-cpp")
	registerPreset("llm-tool")

	// Register llm provider factory (Simon Willison's tool)
	llm.RegisterProviderFactory("llm", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		// providerSpec is the model name for llm
		provider, err := NewLLMCLIProvider(providerSpec, options)
		if err != nil {
			return nil, err
		}
		return provider, nil
	})
}
