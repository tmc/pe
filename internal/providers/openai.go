package providers

import (
	"fmt"

	"github.com/tmc/pe/internal/inference"
	inferenceopenai "github.com/tmc/pe/internal/inference/providers/openai"
	"github.com/tmc/pe/internal/llm"
)

// NewOpenAIProvider creates an OpenAI provider using the shared inference implementation.
func NewOpenAIProvider(model string, options map[string]interface{}) (llm.Provider, error) {
	cfg := providerConfig(options)
	if v, ok := cfg["apiKey"]; ok {
		cfg["api_key"] = v
	}
	if v, ok := cfg["baseURL"]; ok {
		cfg["base_url"] = v
	}
	provider, err := inferenceopenai.Factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("openai provider: %w", err)
	}
	return inference.NewModernAdapter(provider, model), nil
}

func providerConfig(options map[string]interface{}) map[string]interface{} {
	cfg := make(map[string]interface{}, len(options))
	for k, v := range options {
		cfg[k] = v
	}
	return cfg
}
