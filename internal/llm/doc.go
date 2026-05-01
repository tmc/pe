// Package llm defines language model provider interfaces.
//
// GenerateOptions separates portable options, such as Temperature and
// MaxTokens, from ProviderOptions. ProviderOptions carries backend-specific
// values such as Ollama raw and seed settings.
package llm
