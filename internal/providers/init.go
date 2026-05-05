package providers

import (
	"github.com/tmc/pe/internal/inference"
	inferenceanthropic "github.com/tmc/pe/internal/inference/providers/anthropic"
	inferencecgpt "github.com/tmc/pe/internal/inference/providers/cgpt"
	inferenceollama "github.com/tmc/pe/internal/inference/providers/ollama"
	inferenceopenai "github.com/tmc/pe/internal/inference/providers/openai"
	"github.com/tmc/pe/internal/llm"
)

func init() {
	// Register OpenAI provider factory
	llm.RegisterProviderFactory("openai", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := inferenceopenai.Factory(remoteProviderConfig(options))
		if err != nil {
			return nil, err
		}
		return inference.NewModernAdapter(provider, providerSpec), nil
	})

	// Register Anthropic provider factory
	llm.RegisterProviderFactory("anthropic", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := inferenceanthropic.Factory(remoteProviderConfig(options))
		if err != nil {
			return nil, err
		}
		return inference.NewModernAdapter(provider, providerSpec), nil
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
	inference.Register("mock", func(config map[string]interface{}) (inference.Provider, error) {
		model, _ := config["model"].(string)
		provider, err := NewMockProvider(model, config)
		if err != nil {
			return nil, err
		}
		return inference.MigrateProvider(provider), nil
	})

	// Register cgpt provider factory
	llm.RegisterProviderFactory("cgpt", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		config := cgptProviderConfig(options)
		provider := inferencecgpt.New()
		if binary, ok := config["binary"].(string); ok && binary != "" {
			provider = inferencecgpt.NewWithBinary(binary)
		}
		return inference.NewModernAdapter(provider, providerSpec), nil
	})

	// Register generic CLI provider factory
	llm.RegisterProviderFactory("cli", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := NewGenericCLIProvider(providerSpec, options)
		if err != nil {
			return nil, err
		}
		return provider, nil
	})

	// Register Ollama using the native HTTP API implementation so request
	// options and runtime metrics are available without CLI scraping.
	llm.RegisterProviderFactory("ollama", func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
		provider, err := inferenceollama.Factory(options)
		if err != nil {
			return nil, err
		}
		return inference.NewModernAdapter(provider, providerSpec), nil
	})

	// Helper to register presets.
	registerPreset := func(names ...string) {
		if len(names) == 0 {
			return
		}
		buildName := names[0]
		for _, name := range names {
			llm.RegisterProviderFactory(name, func(providerSpec string, options map[string]interface{}) (llm.Provider, error) {
				// Get preset configuration
				config, err := GetPresetConfig(buildName, options)
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
	}

	// Register presets
	registerPreset("mlx")
	registerPreset("mlx-lm")
	registerPreset("mlx-go", "mlx-go-lm")
	registerPreset("llama-cpp", "llama.cpp")
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

func remoteProviderConfig(options map[string]interface{}) map[string]interface{} {
	config := cloneOptions(options)
	if v, ok := config["apiKey"]; ok {
		config["api_key"] = v
	}
	if v, ok := config["baseURL"]; ok {
		config["base_url"] = v
	}
	return config
}

func cgptProviderConfig(options map[string]interface{}) map[string]interface{} {
	config := cloneOptions(options)
	if v, ok := config["executable"]; ok {
		config["binary"] = v
	}
	return config
}
