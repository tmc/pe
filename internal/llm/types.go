// Package llm provides interfaces and utilities for working with language models.
package llm

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Common errors
var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrModelNotFound    = errors.New("model not found")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrAuthFailure      = errors.New("authentication failure")
	ErrRateLimited      = errors.New("rate limited")
)

// Provider represents an LLM provider like OpenAI, Anthropic, etc.
type Provider interface {
	// Name returns the name of the provider.
	Name() string
	
	// Models returns the available models for this provider.
	Models() []Model
	
	// GetModel returns a specific model by name.
	GetModel(name string) (Model, error)
	
	// DefaultModel returns the default model for this provider.
	DefaultModel() Model
}

// Model represents a specific language model.
type Model interface {
	// Name returns the name of the model.
	Name() string
	
	// Provider returns the provider that offers this model.
	Provider() Provider
	
	// Complete generates text completions for the given prompt.
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	
	// Stream generates streaming text completions for the given prompt.
	Stream(ctx context.Context, req *CompletionRequest) (CompletionStream, error)
	
	// MaxContextSize returns the maximum context size in tokens.
	MaxContextSize() int
	
	// IsStreamable returns true if the model supports streaming.
	IsStreamable() bool
}

// CompletionRequest represents a request to a language model.
type CompletionRequest struct {
	Prompt       string                 `json:"prompt"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	Temperature  float64                `json:"temperature,omitempty"`
	TopP         float64                `json:"top_p,omitempty"`
	StopSequences []string              `json:"stop_sequences,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// CompletionResponse represents a response from a language model.
type CompletionResponse struct {
	Text         string                 `json:"text"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	Usage        TokenUsage             `json:"usage"`
	Model        string                 `json:"model"`
	Created      time.Time              `json:"created"`
	Latency      time.Duration          `json:"latency"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Error        error                  `json:"error,omitempty"`
}

// TokenUsage represents token usage information.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CompletionStream is an interface for streaming completions.
type CompletionStream interface {
	// Next returns the next chunk of the completion.
	// It returns nil when the stream is complete or an error occurred.
	Next() (*CompletionChunk, error)
	
	// Close closes the stream.
	Close() error
}

// CompletionChunk represents a chunk of a streaming completion.
type CompletionChunk struct {
	Text      string `json:"text"`
	IsFinal   bool   `json:"is_final"`
	Error     error  `json:"error,omitempty"`
}

// ProviderFactory creates a provider from a string identification.
type ProviderFactory interface {
	// Create creates a provider from a string like "openai", "anthropic", etc.
	Create(providerStr string) (Provider, error)
	
	// SupportsProvider returns true if the factory can create the provider.
	SupportsProvider(providerStr string) bool
}

// ParseProviderAndModel parses a provider:model string (e.g., "openai:gpt-4").
func ParseProviderAndModel(s string) (providerName, modelName string) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return s, ""
}

// Format returns a string in the format "provider:model".
func Format(providerName, modelName string) string {
	if modelName == "" {
		return providerName
	}
	return providerName + ":" + modelName
}