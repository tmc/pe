// Package promptfoo provides types and utilities for working with promptfoo configurations
// and evaluation results, aligned with the protobuf definitions in specs/promptfoo.proto.
package promptfoo

import (
	"encoding/json"
	"time"
)

// TestSuiteConfig defines a complete test configuration.
type TestSuiteConfig struct {
	Description        string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Prompts            []interface{}            `json:"prompts" yaml:"prompts"`
	Providers          []interface{}            `json:"providers" yaml:"providers"`
	Tests              []TestCase               `json:"tests" yaml:"tests"`
	DefaultTest        *TestCase                `json:"defaultTest,omitempty" yaml:"defaultTest,omitempty"`
	OutputPath         interface{}              `json:"outputPath,omitempty" yaml:"outputPath,omitempty"`
	AssertionTemplates map[string]AssertionSpec `json:"assertionTemplates,omitempty" yaml:"assertionTemplates,omitempty"`
	Scenarios          []Scenario               `json:"scenarios,omitempty" yaml:"scenarios,omitempty"`
	Sharing            bool                     `json:"sharing,omitempty" yaml:"sharing,omitempty"`
	Extensions         []string                 `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	Tags               map[string]string        `json:"tags,omitempty" yaml:"tags,omitempty"` 
	Env                map[string]string        `json:"env,omitempty" yaml:"env,omitempty"`
	Metadata           map[string]interface{}   `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TestCase represents a single test case with variables and assertions.
type TestCase struct {
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Vars        map[string]interface{} `json:"vars,omitempty" yaml:"vars,omitempty"`
	Assert      []AssertionSpec        `json:"assert,omitempty" yaml:"assert,omitempty"`
	Options     *TestOptions           `json:"options,omitempty" yaml:"options,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// AssertionSpec defines an assertion to be applied to a model response.
type AssertionSpec struct {
	Type      string      `json:"type" yaml:"type"`
	Value     interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	Threshold float64     `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	Path      string      `json:"path,omitempty" yaml:"path,omitempty"`
	Provider  string      `json:"provider,omitempty" yaml:"provider,omitempty"`
}

// TestOptions provides additional configuration for test execution.
type TestOptions struct {
	Provider     interface{}            `json:"provider,omitempty" yaml:"provider,omitempty"`
	Transform    string                 `json:"transform,omitempty" yaml:"transform,omitempty"`
	RubricPrompt interface{}            `json:"rubricPrompt,omitempty" yaml:"rubricPrompt,omitempty"`
	MaxTokens    int32                  `json:"maxTokens,omitempty" yaml:"maxTokens,omitempty"`
	Temperature  float64                `json:"temperature,omitempty" yaml:"temperature,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Scenario defines a group of related test cases.
type Scenario struct {
	Config []TestCase `json:"config" yaml:"config"`
	Tests  []TestCase `json:"tests" yaml:"tests"`
}

// TokenUsage tracks token consumption and associated costs.
type TokenUsage struct {
	Total      int32                 `json:"total" yaml:"total"`
	Prompt     int32                 `json:"prompt" yaml:"prompt"`
	Completion int32                 `json:"completion" yaml:"completion"`
	Cached     int32                 `json:"cached" yaml:"cached"`
	NumRequests int32                `json:"numRequests,omitempty" yaml:"numRequests,omitempty"`
	Details    *CompletionTokenDetails `json:"completionDetails,omitempty" yaml:"completionDetails,omitempty"`
}

// CompletionTokenDetails provides a detailed breakdown of token usage.
type CompletionTokenDetails struct {
	Reasoning          int32 `json:"reasoning" yaml:"reasoning"`
	AcceptedPrediction int32 `json:"acceptedPrediction" yaml:"acceptedPrediction"`
	RejectedPrediction int32 `json:"rejectedPrediction" yaml:"rejectedPrediction"`
}

// ProviderResponse represents a standardized response from any model provider.
type ProviderResponse struct {
	Output     string      `json:"output" yaml:"output"`
	TokenUsage *TokenUsage `json:"tokenUsage,omitempty" yaml:"tokenUsage,omitempty"`
	Cached     bool        `json:"cached,omitempty" yaml:"cached,omitempty"`
	FinishTime time.Time   `json:"finishTime,omitempty" yaml:"finishTime,omitempty"`
	Cost       float64     `json:"cost,omitempty" yaml:"cost,omitempty"`
	Error      string      `json:"error,omitempty" yaml:"error,omitempty"`
	Raw        interface{} `json:"raw,omitempty" yaml:"raw,omitempty"`
}

// PromptInfo represents a single prompt with metadata.
type PromptInfo struct {
	Raw       string            `json:"raw" yaml:"raw"`
	Label     string            `json:"label,omitempty" yaml:"label,omitempty"`
	ID        string            `json:"id,omitempty" yaml:"id,omitempty"`
	Provider  string            `json:"provider,omitempty" yaml:"provider,omitempty"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// PromptMetrics contains evaluation metrics for a specific prompt.
type PromptMetrics struct {
	Score           float64                `json:"score" yaml:"score"`
	TestPassCount   int32                  `json:"testPassCount" yaml:"testPassCount"`
	TestFailCount   int32                  `json:"testFailCount" yaml:"testFailCount"`
	TestErrorCount  int32                  `json:"testErrorCount,omitempty" yaml:"testErrorCount,omitempty"`
	AssertPassCount int32                  `json:"assertPassCount" yaml:"assertPassCount"` 
	AssertFailCount int32                  `json:"assertFailCount" yaml:"assertFailCount"`
	TotalLatencyMs  int64                  `json:"totalLatencyMs" yaml:"totalLatencyMs"`
	TokenUsage      *TokenUsage            `json:"tokenUsage,omitempty" yaml:"tokenUsage,omitempty"` 
	NamedScores     map[string]float64     `json:"namedScores,omitempty" yaml:"namedScores,omitempty"`
	NamedScoresCount map[string]int32      `json:"namedScoresCount,omitempty" yaml:"namedScoresCount,omitempty"`
	Cost            float64                `json:"cost,omitempty" yaml:"cost,omitempty"`
}

// CompletedPrompt combines a prompt with its evaluation metrics.
type CompletedPrompt struct {
	Prompt  *PromptInfo    `json:"prompt" yaml:"prompt"`
	Metrics *PromptMetrics `json:"metrics" yaml:"metrics"`
}

// GradingResult contains the evaluation results for a specific test.
type GradingResult struct {
	Pass             bool                   `json:"pass" yaml:"pass"`
	Score            float64                `json:"score" yaml:"score"`
	Reason           string                 `json:"reason,omitempty" yaml:"reason,omitempty"`
	NamedScores      map[string]float64     `json:"namedScores,omitempty" yaml:"namedScores,omitempty"`
	TokensUsed       *TokenUsage            `json:"tokensUsed,omitempty" yaml:"tokensUsed,omitempty"`
	ComponentResults []ComponentGradingResult `json:"componentResults,omitempty" yaml:"componentResults,omitempty"`
	Assertion        *AssertionSpec         `json:"assertion,omitempty" yaml:"assertion,omitempty"`
}

// ComponentGradingResult represents the result of a single component in a grading operation.
type ComponentGradingResult struct {
	Pass      bool          `json:"pass" yaml:"pass"`
	Score     float64       `json:"score" yaml:"score"`
	Reason    string        `json:"reason,omitempty" yaml:"reason,omitempty"`
	Assertion *AssertionSpec `json:"assertion,omitempty" yaml:"assertion,omitempty"`
}

// EvaluationResult represents the full results from an evaluation run.
type EvaluationResult struct {
	Version   int                    `json:"version" yaml:"version"`
	Timestamp string                 `json:"timestamp" yaml:"timestamp"`
	Prompts   []CompletedPrompt      `json:"prompts" yaml:"prompts"`
	Results   []TestResult           `json:"results" yaml:"results"`
	Stats     *EvaluateStats         `json:"stats" yaml:"stats"`
}

// TestResult contains the results of a specific test execution.
type TestResult struct {
	ID            string                 `json:"id" yaml:"id"`
	PromptID      string                 `json:"promptId" yaml:"promptId"`
	PromptIdx     int32                  `json:"promptIdx" yaml:"promptIdx"`
	TestIdx       int32                  `json:"testIdx" yaml:"testIdx"`
	Prompt        *PromptInfo            `json:"prompt" yaml:"prompt"`
	Provider      *ProviderInfo          `json:"provider" yaml:"provider"`
	Response      *ProviderResponse      `json:"response" yaml:"response"`
	LatencyMs     float64                `json:"latencyMs" yaml:"latencyMs"`
	Cost          float64                `json:"cost,omitempty" yaml:"cost,omitempty"`
	Success       bool                   `json:"success" yaml:"success"`
	Score         float64                `json:"score" yaml:"score"`
	NamedScores   map[string]float64     `json:"namedScores,omitempty" yaml:"namedScores,omitempty"`
	Vars          map[string]interface{} `json:"vars" yaml:"vars"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	FailureReason int32                  `json:"failureReason,omitempty" yaml:"failureReason,omitempty"`
	TestCase      *TestCase              `json:"testCase" yaml:"testCase"`
	GradingResult *GradingResult         `json:"gradingResult" yaml:"gradingResult"`
}

// ProviderInfo contains information about a specific provider.
type ProviderInfo struct {
	ID    string `json:"id" yaml:"id"`
	Label string `json:"label,omitempty" yaml:"label,omitempty"`
}

// EvaluateStats provides statistical summary of evaluation results.
type EvaluateStats struct {
	Successes  int32      `json:"successes" yaml:"successes"`
	Failures   int32      `json:"failures" yaml:"failures"`
	Errors     int32      `json:"errors,omitempty" yaml:"errors,omitempty"`
	TokenUsage *TokenUsage `json:"tokenUsage,omitempty" yaml:"tokenUsage,omitempty"`
}

// FullEvaluationResults represents the complete results with config and shareable URL.
type FullEvaluationResults struct {
	EvalID       string            `json:"evalId" yaml:"evalId"`
	Config       *TestSuiteConfig  `json:"config" yaml:"config"`
	Results      *EvaluationResult `json:"results" yaml:"results"`
	ShareableURL *string           `json:"shareableUrl,omitempty" yaml:"shareableUrl,omitempty"`
}

// Custom marshaling for interface{} fields that might contain file references
func (p *TestSuiteConfig) MarshalJSON() ([]byte, error) {
	type Alias TestSuiteConfig
	return json.Marshal(&struct {
		Prompts   []string `json:"prompts"`
		Providers []string `json:"providers"`
		*Alias
	}{
		Prompts:   stringifySlice(p.Prompts),
		Providers: stringifySlice(p.Providers),
		Alias:     (*Alias)(p),
	})
}

func stringifySlice(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = stringify(v)
	}
	return result
}

func stringify(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case map[string]interface{}:
		if id, ok := t["id"].(string); ok {
			return id
		}
	}
	return ""
}
