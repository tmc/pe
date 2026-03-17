package llm

import (
	"context"

	"github.com/tmc/pe/internal/promptfoo"
)

// ConfiguredProvider applies fixed default options to an underlying provider.
type ConfiguredProvider struct {
	base     Provider
	defaults map[string]interface{}
}

// NewConfiguredProvider wraps a provider with default per-request options.
func NewConfiguredProvider(base Provider, defaults map[string]interface{}) Provider {
	if base == nil || len(defaults) == 0 {
		return base
	}
	copied := make(map[string]interface{}, len(defaults))
	for k, v := range defaults {
		copied[k] = v
	}
	return &ConfiguredProvider{
		base:     base,
		defaults: copied,
	}
}

func (p *ConfiguredProvider) Name() string            { return p.base.Name() }
func (p *ConfiguredProvider) Model() string           { return p.base.Model() }
func (p *ConfiguredProvider) SupportsStreaming() bool { return p.base.SupportsStreaming() }
func (p *ConfiguredProvider) SupportsBatch() bool     { return p.base.SupportsBatch() }

func (p *ConfiguredProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error) {
	merged := options
	if merged.Temperature == nil {
		if v, ok := getFloat64Default(p.defaults, "temperature"); ok {
			merged.Temperature = &v
		}
	}
	if merged.MaxTokens == nil {
		if v, ok := getIntDefault(p.defaults, "max_tokens"); ok {
			merged.MaxTokens = &v
		} else if v, ok := getIntDefault(p.defaults, "num_predict"); ok {
			merged.MaxTokens = &v
		}
	}
	if merged.TopP == nil {
		if v, ok := getFloat64Default(p.defaults, "top_p"); ok {
			merged.TopP = &v
		}
	}
	if merged.TopK == nil {
		if v, ok := getIntDefault(p.defaults, "top_k"); ok {
			merged.TopK = &v
		}
	}
	if len(merged.Stop) == 0 {
		if stop, ok := p.defaults["stop"].([]string); ok {
			merged.Stop = append([]string(nil), stop...)
		}
	}
	return p.base.Generate(ctx, prompt, merged)
}

func (p *ConfiguredProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	merged := make(map[string]interface{}, len(p.defaults)+len(vars))
	for k, v := range p.defaults {
		merged[k] = v
	}
	for k, v := range vars {
		merged[k] = v
	}
	return p.base.EvaluatePrompt(ctx, prompt, merged)
}

func getIntDefault(values map[string]interface{}, key string) (int, bool) {
	value, ok := values[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func getFloat64Default(values map[string]interface{}, key string) (float64, bool) {
	value, ok := values[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}
