package config

import "time"

// Command returns command-specific config by name.
func (c *Config) Command(name string) (CommandConfig, bool) {
	if c == nil || c.Commands == nil {
		return CommandConfig{}, false
	}
	cfg, ok := c.Commands[name]
	return cfg, ok
}

// ProviderOptions returns provider options suitable for provider factories.
func (c *Config) ProviderOptions(name string) map[string]interface{} {
	out := make(map[string]interface{})
	if c == nil {
		return out
	}
	switch name {
	case "openai":
		addString(out, "api_key", c.Providers.OpenAI.APIKey)
		addString(out, "base_url", c.Providers.OpenAI.BaseURL)
		addString(out, "organization_id", c.Providers.OpenAI.OrganizationID)
		addString(out, "model", c.Providers.OpenAI.DefaultModel)
		addDuration(out, "timeout", c.Providers.OpenAI.Timeout)
		addInt(out, "max_retries", c.Providers.OpenAI.MaxRetries)
	case "anthropic":
		addString(out, "api_key", c.Providers.Anthropic.APIKey)
		addString(out, "base_url", c.Providers.Anthropic.BaseURL)
		addString(out, "model", c.Providers.Anthropic.DefaultModel)
		addDuration(out, "timeout", c.Providers.Anthropic.Timeout)
		addInt(out, "max_retries", c.Providers.Anthropic.MaxRetries)
	case "ollama":
		addString(out, "host", c.Providers.Ollama.Host)
		addString(out, "model", c.Providers.Ollama.DefaultModel)
		addDuration(out, "timeout", c.Providers.Ollama.Timeout)
		addDuration(out, "keep_alive", c.Providers.Ollama.KeepAlive)
	case "cgpt":
		addString(out, "binary", c.Providers.CGPT.BinaryPath)
		if len(c.Providers.CGPT.DefaultArgs) > 0 {
			out["args"] = c.Providers.CGPT.DefaultArgs
		}
		if len(c.Providers.CGPT.Environment) > 0 {
			out["environment"] = c.Providers.CGPT.Environment
		}
		addDuration(out, "timeout", c.Providers.CGPT.Timeout)
	}
	if custom, ok := c.Providers.Custom[name]; ok {
		addString(out, "type", custom.Type)
		for k, v := range custom.Config {
			out[k] = v
		}
	}
	return out
}

func addString(out map[string]interface{}, key, value string) {
	if value != "" {
		out[key] = value
	}
}

func addInt(out map[string]interface{}, key string, value int) {
	if value != 0 {
		out[key] = value
	}
}

func addDuration(out map[string]interface{}, key string, value time.Duration) {
	if value != 0 {
		out[key] = value
	}
}
