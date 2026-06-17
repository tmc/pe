package rlm

// TraceVersion identifies the trace schema. Field names are stable so tests and
// release checks can parse them.
const TraceVersion = "pe.rlm.trace.v1"

// Termination reasons recorded on calls and on the run as a whole.
const (
	TerminationCompleted = "completed"
	TerminationMaxDepth  = "max_depth"
	TerminationMaxTokens = "max_tokens"
	TerminationMaxChunks = "max_chunks"
	TerminationError     = "error"
)

// Trace is the strict JSON record emitted by a run.
type Trace struct {
	Version     string      `json:"version"`
	Command     string      `json:"command"`
	Target      Target      `json:"target"`
	Budget      Budget      `json:"budget"`
	Calls       []Call      `json:"calls"`
	Aggregation Aggregation `json:"aggregation"`

	// TerminationReason explains why the run stopped: completed, max_depth,
	// max_tokens, max_chunks, or error.
	TerminationReason string `json:"termination_reason"`
}

// Target identifies the input artifact processed by a run. The artifact is
// referenced by content digest and optional manifest, not inlined.
type Target struct {
	Path     string `json:"path"`
	Manifest string `json:"manifest,omitempty"`
	SHA256   string `json:"sha256"`
}

// Budget records the hard limits applied to a run. The runner enforces these
// before any worker call.
type Budget struct {
	MaxDepth  int `json:"max_depth"`
	MaxTokens int `json:"max_tokens"`
	Workers   int `json:"workers"`

	// MaxChunks caps the number of chunks a single pass may select. Zero means
	// no chunk cap beyond what chunking itself produces.
	MaxChunks int `json:"max_chunks,omitempty"`
}

// Call records one bounded operation against a snippet of the target.
type Call struct {
	ID          string  `json:"id"`
	ParentID    string  `json:"parent_id"`
	Depth       int     `json:"depth"`
	Operation   string  `json:"operation"`
	Snippet     Snippet `json:"snippet"`
	ChildPrompt string  `json:"child_prompt"`
	Provider    string  `json:"provider"`
	Cost        Cost    `json:"cost"`

	// TerminationReason explains why the call stopped: completed or error.
	TerminationReason string `json:"termination_reason"`

	// Error holds the failure message when TerminationReason is error.
	Error string `json:"error,omitempty"`
}

// Snippet references the bounded byte range a call inspected. The payload is
// stored out-of-prompt and addressed by CacheKey.
type Snippet struct {
	CacheKey string `json:"cache_key"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
}

// Cost records token consumption reported for a call.
type Cost struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Aggregation records how child results were merged into the final answer.
type Aggregation struct {
	Rule      string `json:"rule"`
	Consensus string `json:"consensus"`
	Votes     []Vote `json:"votes"`
}

// Vote records one child answer considered during aggregation.
type Vote struct {
	CallID string `json:"call_id"`
	Output string `json:"output"`
	Weight int    `json:"weight"`
}
