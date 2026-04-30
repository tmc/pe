package providers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// MaterializedProvider couples a parsed provider spec with an executable provider.
type MaterializedProvider struct {
	Spec     promptfoo.ProviderConfig
	Executor llm.Provider
	Delay    time.Duration
}

// DisplayName returns the label when set, otherwise the provider id.
func (p MaterializedProvider) DisplayName() string {
	return p.Spec.DisplayName()
}

// AppliesToPrompt reports whether this provider should be used for the prompt.
func (p MaterializedProvider) AppliesToPrompt(prompt string) bool {
	if len(p.Spec.Prompts) == 0 {
		return true
	}
	for _, allowed := range p.Spec.Prompts {
		if allowed == prompt {
			return true
		}
	}
	return false
}

// MaterializeProvider creates an executable provider from a promptfoo provider spec.
func MaterializeProvider(spec promptfoo.ProviderConfig) (*MaterializedProvider, error) {
	options := cloneOptions(spec.Config)
	if len(spec.Env) > 0 {
		env := make(map[string]interface{}, len(spec.Env))
		for k, v := range spec.Env {
			env[k] = v
		}
		if existing, ok := options["env"].(map[string]interface{}); ok {
			for k, v := range existing {
				env[k] = v
			}
		} else if existing, ok := options["env"].(map[string]string); ok {
			for k, v := range existing {
				env[k] = v
			}
		}
		options["env"] = env
	}

	executor, err := llm.GetProviderWithOptions(spec.ID, options)
	if err != nil {
		return nil, err
	}

	delay, err := parseProviderDelay(spec.Delay)
	if err != nil {
		return nil, fmt.Errorf("provider %s delay: %w", spec.ID, err)
	}

	return &MaterializedProvider{
		Spec:     spec,
		Executor: executor,
		Delay:    delay,
	}, nil
}

// MaterializeProviders creates executable providers from provider specs.
func MaterializeProviders(specs []promptfoo.ProviderConfig) ([]*MaterializedProvider, error) {
	out := make([]*MaterializedProvider, 0, len(specs))
	for _, spec := range specs {
		provider, err := MaterializeProvider(spec)
		if err != nil {
			return nil, err
		}
		out = append(out, provider)
	}
	return out, nil
}

func cloneOptions(options map[string]interface{}) map[string]interface{} {
	if len(options) == 0 {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(options))
	for k, v := range options {
		out[k] = v
	}
	return out
}

func parseProviderDelay(delay string) (time.Duration, error) {
	delay = strings.TrimSpace(delay)
	if delay == "" {
		return 0, nil
	}
	if ms, err := strconv.Atoi(delay); err == nil {
		return time.Duration(ms) * time.Millisecond, nil
	}
	return time.ParseDuration(delay)
}
