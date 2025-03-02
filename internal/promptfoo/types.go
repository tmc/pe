package promptfoo

import (
	"encoding/json"
)

type PromptfooConfig struct {
	Description        string               `json:"description,omitempty" yaml:"description,omitempty" csv:"description"`
	Prompts            []interface{}        `json:"prompts" yaml:"prompts" csv:"prompts"`
	Providers          []interface{}        `json:"providers" yaml:"providers" csv:"providers"`
	Tests              []TestCase           `json:"tests" yaml:"tests" csv:"tests"`
	DefaultTest        *TestCase            `json:"defaultTest,omitempty" yaml:"defaultTest,omitempty" csv:"defaultTest"`
	OutputPath         interface{}          `json:"outputPath,omitempty" yaml:"outputPath,omitempty" csv:"outputPath"`
	AssertionTemplates map[string]Assertion `json:"assertionTemplates,omitempty" yaml:"assertionTemplates,omitempty" csv:"assertionTemplates"`
	Scenarios          []Scenario           `json:"scenarios,omitempty" yaml:"scenarios,omitempty" csv:"scenarios"`
	NunjucksFilters    map[string]string    `json:"nunjucksFilters,omitempty" yaml:"nunjucksFilters,omitempty" csv:"nunjucksFilters"`
	DerivedMetrics     []DerivedMetric      `json:"derivedMetrics,omitempty" yaml:"derivedMetrics,omitempty" csv:"derivedMetrics"`
	Extensions         []string             `json:"extensions,omitempty" yaml:"extensions,omitempty" csv:"extensions"`
}

type TestCase struct {
	Description string                 `json:"description,omitempty" yaml:"description,omitempty" csv:"description"`
	Vars        map[string]interface{} `json:"vars,omitempty" yaml:"vars,omitempty" csv:"vars"`
	Assert      []Assertion            `json:"assert,omitempty" yaml:"assert,omitempty" csv:"assert"`
	Options     *TestOptions           `json:"options,omitempty" yaml:"options,omitempty" csv:"options"`
}

type Assertion struct {
	Type      string      `json:"type" yaml:"type" csv:"type"`
	Value     interface{} `json:"value,omitempty" yaml:"value,omitempty" csv:"value"`
	Threshold float64     `json:"threshold,omitempty" yaml:"threshold,omitempty" csv:"threshold"`
}

type TestOptions struct {
	Provider     interface{} `json:"provider,omitempty" yaml:"provider,omitempty" csv:"provider"`
	Transform    string      `json:"transform,omitempty" yaml:"transform,omitempty" csv:"transform"`
	RubricPrompt interface{} `json:"rubricPrompt,omitempty" yaml:"rubricPrompt,omitempty" csv:"rubricPrompt"`
}

type Scenario struct {
	Config []TestCase `json:"config" yaml:"config" csv:"config"`
	Tests  []TestCase `json:"tests" yaml:"tests" csv:"tests"`
}

type DerivedMetric struct {
	Name  string      `json:"name" yaml:"name" csv:"name"`
	Value interface{} `json:"value" yaml:"value" csv:"value"`
}

type EvaluateResult struct {
	Version int      `json:"version" yaml:"version" csv:"version"`
	Results []Result `json:"results" yaml:"results" csv:"results"`
	Stats   Stats    `json:"stats" yaml:"stats" csv:"stats"`
	Table   Table    `json:"table" yaml:"table" csv:"table"`
}

type Result struct {
	Prompt   Prompt                 `json:"prompt" yaml:"prompt" csv:"prompt"`
	Vars     map[string]interface{} `json:"vars" yaml:"vars" csv:"vars"`
	Response Response               `json:"response" yaml:"response" csv:"response"`
	Success  bool                   `json:"success" yaml:"success" csv:"success"`
}

type Prompt struct {
	Raw     string `json:"raw" yaml:"raw" csv:"raw"`
	Display string `json:"display" yaml:"display" csv:"display"`
}

type Response struct {
	Output     string     `json:"output" yaml:"output" csv:"output"`
	TokenUsage TokenUsage `json:"tokenUsage" yaml:"tokenUsage" csv:"tokenUsage"`
}

type TokenUsage struct {
	Total      int `json:"total" yaml:"total" csv:"total"`
	Prompt     int `json:"prompt" yaml:"prompt" csv:"prompt"`
	Completion int `json:"completion" yaml:"completion" csv:"completion"`
	Cached     int `json:"cached" yaml:"cached" csv:"cached"`
}

type Stats struct {
	Successes  int        `json:"successes" yaml:"successes" csv:"successes"`
	Failures   int        `json:"failures" yaml:"failures" csv:"failures"`
	TokenUsage TokenUsage `json:"tokenUsage" yaml:"tokenUsage" csv:"tokenUsage"`
}

type Table struct {
	Head TableHead      `json:"head" yaml:"head" csv:"head"`
	Body []TableBodyRow `json:"body" yaml:"body" csv:"body"`
}

type TableHead struct {
	Prompts []string `json:"prompts" yaml:"prompts" csv:"prompts"`
	Vars    []string `json:"vars" yaml:"vars" csv:"vars"`
}

type TableBodyRow struct {
	Outputs []string `json:"outputs" yaml:"outputs" csv:"outputs"`
	Vars    []string `json:"vars" yaml:"vars" csv:"vars"`
}

// Custom marshaling for interface{} fields that might contain file references
func (p *PromptfooConfig) MarshalJSON() ([]byte, error) {
	type Alias PromptfooConfig
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
