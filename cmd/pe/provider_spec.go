package main

import "strings"

func commandProviderSpec(provider, model string) string {
	if provider == "" || model == "" || strings.Contains(provider, ":") {
		return provider
	}
	return provider + ":" + model
}
