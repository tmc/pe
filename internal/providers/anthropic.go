package providers

import (
	"fmt"

	"github.com/tmc/pe/internal/inference"
	inferenceanthropic "github.com/tmc/pe/internal/inference/providers/anthropic"
	"github.com/tmc/pe/internal/llm"
)

// NewAnthropicProvider creates an Anthropic provider using the shared inference implementation.
func NewAnthropicProvider(model string, options map[string]interface{}) (llm.Provider, error) {
	cfg := providerConfig(options)
	if v, ok := cfg["apiKey"]; ok {
		cfg["api_key"] = v
	}
	if v, ok := cfg["baseURL"]; ok {
		cfg["base_url"] = v
	}
	provider, err := inferenceanthropic.Factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("anthropic provider: %w", err)
	}
	return inference.NewModernAdapter(provider, model), nil
}
