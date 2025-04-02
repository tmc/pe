// Package promptfoo provides data structures and utilities for working with
// prompt testing and evaluation data, compatible with the promptfoo schema.
package promptfoo

// Config represents the top-level promptfoo configuration structure.
type Config struct {
	Prompts   []string   `yaml:"prompts"`
	Providers []string   `yaml:"providers"`
	Tests     []TestCase `yaml:"tests"`
}

// TestCase represents a single test case within the promptfoo configuration.
type TestCase struct {
	Vars   map[string]string `yaml:"vars"`
	Assert []Assertion       `yaml:"assert"`
}

// Assertion represents an assertion to be checked against the provider's output.
type Assertion struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
	// TODO: Add other assertion fields as needed (e.g., threshold, weight, not)
}

// ProviderResponse represents a response from an LLM provider.
type ProviderResponse struct {
	Output     string      `json:"output"`
	TokenUsage *TokenUsage `json:"tokenUsage,omitempty"`
	FinishTime interface{} `json:"finishTime,omitempty"`
	Cost       float64     `json:"cost,omitempty"`
}

// TokenUsage represents the token counts for a model response.
type TokenUsage struct {
	Total       int32                   `json:"total"`
	Prompt      int32                   `json:"prompt"`
	Completion  int32                   `json:"completion"`
	Cached      int32                   `json:"cached"`
	NumRequests int32                   `json:"numRequests"`
	Details     *CompletionTokenDetails `json:"details,omitempty"`
}

// CompletionTokenDetails provides a breakdown of token usage in model completions.
type CompletionTokenDetails struct {
	Reasoning          int32 `json:"reasoning"`
	AcceptedPrediction int32 `json:"acceptedPrediction"`
	RejectedPrediction int32 `json:"rejectedPrediction"`
}

// PromptMetrics contains metrics for a specific prompt.
type PromptMetrics struct {
	Score            float64            `json:"score"`
	TestPassCount    int32              `json:"testPassCount"`
	TestFailCount    int32              `json:"testFailCount"`
	AssertPassCount  int32              `json:"assertPassCount"`
	AssertFailCount  int32              `json:"assertFailCount"`
	TokenUsage       *TokenUsage        `json:"tokenUsage,omitempty"`
	NamedScores      map[string]float64 `json:"namedScores,omitempty"`
	NamedScoresCount map[string]int32   `json:"namedScoresCount,omitempty"`
}
