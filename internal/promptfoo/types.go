// Package promptfoo provides data structures and utilities for working with
// prompt testing and evaluation data, compatible with the promptfoo schema.
package promptfoo

import (
	"encoding/json"
	"fmt"
)

// Config represents the promptfoo configuration structure.
type Config struct {
	Description string           `yaml:"description,omitempty" json:"description,omitempty"`
	Prompts     []string         `yaml:"prompts" json:"prompts"`
	Providers   []ProviderConfig `yaml:"providers" json:"providers"`
	Tests       []TestCase       `yaml:"tests" json:"tests"`
	DefaultTest *TestDefaults    `yaml:"defaultTest,omitempty" json:"defaultTest,omitempty"`
}

// ProviderConfig represents a provider configuration.
// It accepts either a plain string, like "openai:gpt-4",
// or an object with an id and per-provider config.
type ProviderConfig struct {
	ID     string                 `yaml:"id" json:"id"`
	Config map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
}

// String returns the provider identifier.
func (p ProviderConfig) String() string {
	return p.ID
}

// MarshalJSON preserves the compact string form when no config is present.
func (p ProviderConfig) MarshalJSON() ([]byte, error) {
	if len(p.Config) == 0 {
		return json.Marshal(p.ID)
	}
	return json.Marshal(struct {
		ID     string                 `json:"id"`
		Config map[string]interface{} `json:"config,omitempty"`
	}{
		ID:     p.ID,
		Config: p.Config,
	})
}

// UnmarshalJSON accepts either a plain string provider spec or
// an object of the form {"id":"...", "config":{...}}.
// Legacy objects using "name" and inline config keys are also accepted.
func (p *ProviderConfig) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err == nil {
		p.ID = id
		p.Config = nil
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("provider must be a string or object: %w", err)
	}

	id = getProviderObjectString(raw, "id")
	if id == "" {
		id = getProviderObjectString(raw, "name")
	}
	if id == "" {
		return fmt.Errorf("provider object requires id")
	}

	config := make(map[string]interface{})
	if nested, ok := raw["config"]; ok {
		nestedMap, ok := nested.(map[string]interface{})
		if !ok {
			return fmt.Errorf("provider config must be an object")
		}
		for k, v := range nestedMap {
			config[k] = v
		}
	}
	for k, v := range raw {
		switch k {
		case "id", "name", "config":
			continue
		default:
			config[k] = v
		}
	}

	p.ID = id
	if len(config) == 0 {
		p.Config = nil
	} else {
		p.Config = config
	}
	return nil
}

func getProviderObjectString(raw map[string]interface{}, key string) string {
	value, ok := raw[key]
	if !ok {
		return ""
	}
	s, _ := value.(string)
	return s
}

// TestCase represents a single test case in the configuration.
type TestCase struct {
	Vars    map[string]interface{} `yaml:"vars" json:"vars"`
	Assert  []Assertion            `yaml:"assert" json:"assert"`
	Options map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

// Assertion represents an assertion to validate provider output.
type Assertion struct {
	Type      string                 `yaml:"type" json:"type"`
	Value     interface{}            `yaml:"value" json:"value"`
	Provider  string                 `yaml:"provider,omitempty" json:"provider,omitempty"`
	Threshold float64                `yaml:"threshold,omitempty" json:"threshold,omitempty"`
	Config    map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
}

// EvaluationResult mirrors promptfoo's output structure.
type EvaluationResult struct {
	EvalID       string    `json:"evalId"`
	Results      ResultSet `json:"results"`
	Config       Config    `json:"config"`
	ShareableURL string    `json:"shareableUrl,omitempty"`
}

// ResultSet contains evaluation results and metadata.
type ResultSet struct {
	Version   int          `json:"version"`
	Timestamp string       `json:"timestamp"`
	Prompts   []PromptData `json:"prompts"`
	Results   []TestResult `json:"results"`
	Stats     Stats        `json:"stats"`
}

// PromptData contains metadata about a prompt.
type PromptData struct {
	Raw      string        `json:"raw"`
	Label    string        `json:"label"`
	ID       string        `json:"id"`
	Provider string        `json:"provider"`
	Metrics  PromptMetrics `json:"metrics"`
}

// PromptMetrics tracks evaluation metrics for a prompt.
type PromptMetrics struct {
	Score            float64            `json:"score"`
	TestPassCount    int                `json:"testPassCount"`
	TestFailCount    int                `json:"testFailCount"`
	TestErrorCount   int                `json:"testErrorCount"`
	AssertPassCount  int                `json:"assertPassCount"`
	AssertFailCount  int                `json:"assertFailCount"`
	TotalLatencyMs   int64              `json:"totalLatencyMs"`
	TokenUsage       TokenUsage         `json:"tokenUsage"`
	NamedScores      map[string]float64 `json:"namedScores"`
	NamedScoresCount map[string]int     `json:"namedScoresCount"`
	Cost             float64            `json:"cost,omitempty"`
}

// TestResult represents a single test evaluation result.
type TestResult struct {
	ID            string                 `json:"id"`
	PromptID      string                 `json:"promptId"`
	Prompt        map[string]string      `json:"prompt"`
	Provider      map[string]string      `json:"provider"`
	Response      ProviderResponse       `json:"response"`
	Success       bool                   `json:"success"`
	Score         float64                `json:"score"`
	Vars          map[string]interface{} `json:"vars"`
	GradingResult GradingResult          `json:"gradingResult"`
	LatencyMs     int64                  `json:"latencyMs"`
}

// ProviderResponse represents a response from an LLM provider.
type ProviderResponse struct {
	Output     string                 `json:"output"`
	TokenUsage *TokenUsage            `json:"tokenUsage,omitempty"`
	Cost       float64                `json:"cost,omitempty"`
	Cached     bool                   `json:"cached,omitempty"`
	LatencyMs  int64                  `json:"latencyMs,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TokenUsage tracks token consumption.
type TokenUsage struct {
	Total       int32              `json:"total"`
	Prompt      int32              `json:"prompt"`
	Completion  int32              `json:"completion"`
	Cached      int32              `json:"cached"`
	NumRequests int32              `json:"numRequests,omitempty"`
	Details     *CompletionDetails `json:"completionDetails,omitempty"`
}

// CompletionDetails provides details about token usage for completions
type CompletionDetails struct {
	Reasoning          int32 `json:"reasoning"`
	AcceptedPrediction int32 `json:"acceptedPrediction"`
	RejectedPrediction int32 `json:"rejectedPrediction"`
}

// GradingResult represents the outcome of assertion checks.
type GradingResult struct {
	Pass             bool              `json:"pass"`
	Score            float64           `json:"score"`
	Reason           string            `json:"reason"`
	ComponentResults []ComponentResult `json:"componentResults"`
	TokensUsed       TokenUsage        `json:"tokensUsed"`
}

// ComponentResult represents the result of a single assertion.
type ComponentResult struct {
	Pass      bool      `json:"pass"`
	Score     float64   `json:"score"`
	Reason    string    `json:"reason"`
	Assertion Assertion `json:"assertion"`
}

// Stats summarizes evaluation statistics.
type Stats struct {
	Successes  int        `json:"successes"`
	Failures   int        `json:"failures"`
	Errors     int        `json:"errors"`
	TokenUsage TokenUsage `json:"tokenUsage"`
}

// Additional types for compatibility

// Provider is kept as a compatibility alias for older code.
type Provider = ProviderConfig

// Prompt represents a prompt configuration
type Prompt struct {
	ID       string `yaml:"id" json:"id"`
	Template string `yaml:"template" json:"template"`
}

// Test represents a test configuration
type Test struct {
	Vars    map[string]interface{} `yaml:"vars" json:"vars"`
	Assert  []Assertion            `yaml:"assert" json:"assert"`
	Options map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

// TestDefaults represents default test configuration
type TestDefaults struct {
	Assert []Assertion `yaml:"assert,omitempty" json:"assert,omitempty"`
}

// EvalResults represents evaluation results
type EvalResults struct {
	Results []TestResult `json:"results"`
	Stats   Stats        `json:"stats"`
}

// BodyRow represents a row in the results table
type BodyRow struct {
	Test   TestCase   `json:"test"`
	Result TestResult `json:"result"`
}

// Note: PETestResults is defined in export.go for now
