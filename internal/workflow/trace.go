package workflow

// TraceVersion identifies the trace schema. Field names are stable so tests
// and release checks can parse them.
const TraceVersion = "pe.workflow.trace.v1"

// Trace is the strict JSON record emitted by a run.
type Trace struct {
	Version string      `json:"version"`
	Name    string      `json:"name"`
	Budget  Budget      `json:"budget"`
	Calls   []CallTrace `json:"calls"`
	Totals  Cost        `json:"totals"`

	// Error holds the failure message when the run stopped with an error.
	Error string `json:"error,omitempty"`
}

// Budget records the hard limits applied to a run. The runner enforces these
// before any model call.
type Budget struct {
	// MaxCalls caps the total number of model calls. Negative means no cap.
	MaxCalls int `json:"max_calls"`
	Workers  int `json:"workers"`
}

// CallTrace records one model call made by a script. The prompt is referenced
// by digest, not inlined.
type CallTrace struct {
	ID           int    `json:"id"`
	Phase        string `json:"phase,omitempty"`
	Label        string `json:"label,omitempty"`
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	PromptSHA256 string `json:"prompt_sha256"`
	PromptBytes  int    `json:"prompt_bytes"`
	Cost         Cost   `json:"cost"`

	// Error holds the failure message when the call failed.
	Error string `json:"error,omitempty"`
}

// Cost records token consumption reported for a call.
type Cost struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (c *Cost) add(other Cost) {
	c.PromptTokens += other.PromptTokens
	c.CompletionTokens += other.CompletionTokens
	c.TotalTokens += other.TotalTokens
}
