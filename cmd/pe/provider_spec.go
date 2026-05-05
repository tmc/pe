package main

import (
	"strings"

	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/llm"
)

func commandProviderSpec(provider, model string) string {
	if provider == "" || model == "" || strings.Contains(provider, ":") {
		return provider
	}
	return provider + ":" + model
}

func commandLegacyProvider(provider, model string) (llm.Provider, error) {
	spec := commandProviderSpec(provider, model)
	inferenceProvider, err := inference.CreateProviderFromSpec(spec, nil)
	if err != nil {
		return nil, err
	}
	return inference.AsLegacyProvider(inferenceProvider, model)
}
