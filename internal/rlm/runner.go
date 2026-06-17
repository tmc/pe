package rlm

import (
	"context"
	"fmt"

	"github.com/tmc/pe/internal/distributed"
)

// DefaultChunkSize is the chunk byte length used when Options.ChunkSize is zero.
const DefaultChunkSize = 4096

// Options configure a single recursive run. All limits are hard budgets that the
// runner enforces before any worker call.
type Options struct {
	// Prompt is the instruction applied to each chunk.
	Prompt string

	// ChunkSize is the maximum byte length of each chunk. Zero selects
	// DefaultChunkSize; a negative value is rejected.
	ChunkSize int

	// MaxDepth is the maximum recursive reduction depth. Depth 0 runs a single
	// map pass with no recursion.
	MaxDepth int

	// MaxTokens is the per-worker token budget. A worker call whose reported
	// cost would exceed the remaining budget terminates the run with
	// max_tokens. Zero disables the token budget.
	MaxTokens int

	// Workers bounds local concurrency for the map pass.
	Workers int

	// MaxChunks caps the chunks a single pass may process. Zero means no cap.
	MaxChunks int

	// Rule names the aggregation rule, e.g. majority or concat. Empty defaults
	// to majority.
	Rule string

	// Command labels the run in the trace, e.g. "pe exp recurse".
	Command string

	// TargetPath records the input path in the trace. It is metadata only; the
	// runner never reads it.
	TargetPath string

	// TargetManifest optionally records an attest manifest type for the target.
	TargetManifest string
}

// withDefaults resolves zero-valued fields to their defaults so the zero Options
// (with only Prompt set) is usable.
func (o Options) withDefaults() Options {
	if o.ChunkSize == 0 {
		o.ChunkSize = DefaultChunkSize
	}
	return o
}

func (o Options) validate() error {
	if o.Prompt == "" {
		return fmt.Errorf("rlm: prompt is required")
	}
	if o.ChunkSize < 0 {
		return fmt.Errorf("rlm: chunk size must not be negative, got %d", o.ChunkSize)
	}
	if o.MaxDepth < 0 {
		return fmt.Errorf("rlm: max depth must not be negative, got %d", o.MaxDepth)
	}
	if o.MaxTokens < 0 {
		return fmt.Errorf("rlm: max tokens must not be negative, got %d", o.MaxTokens)
	}
	if o.Workers < 0 {
		return fmt.Errorf("rlm: workers must not be negative, got %d", o.Workers)
	}
	if o.MaxChunks < 0 {
		return fmt.Errorf("rlm: max chunks must not be negative, got %d", o.MaxChunks)
	}
	switch o.Rule {
	case "", RuleMajority, RuleConcat:
	default:
		return fmt.Errorf("rlm: unsupported aggregation rule %q", o.Rule)
	}
	return nil
}

// Run processes target through bounded map/reduce passes with worker and emits a
// strict trace. It never inlines target into prompts: each chunk is referenced
// by content key, and only the bounded chunk payload reaches the worker.
//
// Budgets are enforced before any worker call. The returned trace records every
// call, the aggregation rule and result, and a termination reason. A worker
// error terminates the run; the partial trace is still returned alongside the
// error.
func Run(ctx context.Context, worker Worker, target []byte, opts Options) (*Trace, error) {
	if worker == nil {
		return nil, fmt.Errorf("rlm: worker is required")
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	opts = opts.withDefaults()

	workers := max(opts.Workers, 1)

	trace := &Trace{
		Version: TraceVersion,
		Command: opts.Command,
		Target: Target{
			Path:     opts.TargetPath,
			Manifest: opts.TargetManifest,
			SHA256:   ContentKey(target),
		},
		Budget: Budget{
			MaxDepth:  opts.MaxDepth,
			MaxTokens: opts.MaxTokens,
			Workers:   workers,
			MaxChunks: opts.MaxChunks,
		},
		Calls: []Call{},
	}

	chunks, err := chunk(target, opts.ChunkSize)
	if err != nil {
		return nil, err
	}
	if opts.MaxChunks > 0 && len(chunks) > opts.MaxChunks {
		chunks = chunks[:opts.MaxChunks]
		trace.TerminationReason = TerminationMaxChunks
	}
	if len(chunks) == 0 {
		trace.Aggregation = Aggregation{Rule: ruleName(opts.Rule), Votes: []Vote{}}
		if trace.TerminationReason == "" {
			trace.TerminationReason = TerminationCompleted
		}
		return trace, nil
	}

	r := &run{
		ctx:     ctx,
		worker:  worker,
		opts:    opts,
		workers: workers,
		trace:   trace,
		budget:  opts.MaxTokens,
	}

	results, err := r.reducePasses(chunks)
	if err != nil {
		return trace, err
	}

	agg, err := aggregate(opts.Rule, results)
	if err != nil {
		trace.Aggregation = Aggregation{Rule: ruleName(opts.Rule), Votes: []Vote{}}
		trace.TerminationReason = TerminationError
		return trace, err
	}
	trace.Aggregation = agg
	if trace.TerminationReason == "" {
		trace.TerminationReason = TerminationCompleted
	}
	return trace, nil
}

// reducePasses runs a map pass over chunks at depth 0, then folds the outputs
// through successive reduce passes while depth and budget remain. Each reduce
// pass groups the prior outputs into batches that fit ChunkSize and reduces each
// batch to one output, so a wide input collapses hierarchically toward a single
// answer.
//
// The loop stops when one of the following holds, in order: a worker error or
// budget limit is recorded (returned immediately), a single output remains
// (TerminationCompleted), the outputs can no longer be batched because each
// already fills a chunk (TerminationCompleted, since no further progress is
// possible), or the depth budget is exhausted with more than one output
// remaining (TerminationMaxDepth).
func (r *run) reducePasses(chunks []Chunk) ([]mapResult, error) {
	depth := 0
	results, err := r.mapPass(chunks, "", depth)
	if err != nil {
		return results, err
	}

	for {
		// A recorded budget or chunk limit halts recursion: the current level's
		// results are the final ones.
		if r.trace.TerminationReason == TerminationMaxTokens || r.trace.TerminationReason == TerminationMaxChunks {
			return results, nil
		}

		outputs := successfulOutputs(results)
		if len(outputs) <= 1 {
			return results, nil
		}

		batches := batchOutputs(outputs, r.opts.ChunkSize)
		// One batch per output means nothing can be combined: every output
		// already fills a chunk on its own, so further passes cannot collapse
		// them. Stop here rather than spin.
		if len(batches) >= len(outputs) {
			return results, nil
		}
		if depth >= r.opts.MaxDepth {
			r.trace.TerminationReason = TerminationMaxDepth
			return results, nil
		}

		next := make([]Chunk, len(batches))
		for i, b := range batches {
			payload := []byte(b)
			next[i] = Chunk{
				Index:    i,
				Start:    0,
				End:      len(payload),
				CacheKey: ContentKey(payload),
				Data:     payload,
			}
		}

		depth++
		results, err = r.mapPass(next, r.lastCallID(), depth)
		if err != nil {
			return results, err
		}
	}
}

// successfulOutputs returns the outputs of non-error results in chunk order.
func successfulOutputs(results []mapResult) []string {
	outputs := make([]string, 0, len(results))
	for _, r := range results {
		if r.err == nil {
			outputs = append(outputs, r.output)
		}
	}
	return outputs
}

// batchOutputs groups outputs in order into newline-joined batches whose length
// does not exceed size. An output larger than size occupies its own batch. The
// grouping is deterministic and order-preserving.
func batchOutputs(outputs []string, size int) []string {
	var batches []string
	var cur []string
	curLen := 0
	for _, o := range outputs {
		add := len(o)
		if len(cur) > 0 {
			add++ // newline separator
		}
		if len(cur) > 0 && curLen+add > size {
			batches = append(batches, joinLines(cur))
			cur = nil
			curLen = 0
			add = len(o)
		}
		cur = append(cur, o)
		curLen += add
	}
	if len(cur) > 0 {
		batches = append(batches, joinLines(cur))
	}
	return batches
}

// lastCallID returns the ID of the most recently recorded call, used as the
// parent for the next reduce level. It is empty when no call has been recorded.
func (r *run) lastCallID() string {
	if n := len(r.trace.Calls); n > 0 {
		return r.trace.Calls[n-1].ID
	}
	return ""
}

// run holds the mutable state threaded through a single Run.
type run struct {
	ctx     context.Context
	worker  Worker
	opts    Options
	workers int
	trace   *Trace
	budget  int // remaining token budget; <=0 disables when opts.MaxTokens==0
	calls   int
}

// mapPass runs the prompt over each chunk concurrently and records one call per
// chunk. parentID and depth label the calls for the trace. The depth-0 pass is
// a map over the target; deeper passes are reduces over prior outputs.
func (r *run) mapPass(chunks []Chunk, parentID string, depth int) ([]mapResult, error) {
	operation := "map"
	if depth > 0 {
		operation = "reduce"
	}

	ids := make([]string, len(chunks))
	for i := range chunks {
		r.calls++
		ids[i] = fmt.Sprintf("call-%d", r.calls)
	}

	tasks := make([]distributed.Task[mapResult], len(chunks))
	for i, c := range chunks {
		tasks[i] = distributed.Task[mapResult]{
			ID: ids[i],
			Run: func(ctx context.Context) (mapResult, error) {
				out, cost, err := r.worker.Run(ctx, r.opts.Prompt, c)
				return mapResult{chunk: c, output: out, cost: cost, err: err}, err
			},
		}
	}

	raw, runErr := distributed.RunLocal(r.ctx, r.workers, tasks)

	results := make([]mapResult, 0, len(raw))
	for i, res := range raw {
		mr := res.Value
		call := Call{
			ID:        ids[i],
			ParentID:  parentID,
			Depth:     depth,
			Operation: operation,
			Snippet: Snippet{
				CacheKey: chunks[i].CacheKey,
				Start:    chunks[i].Start,
				End:      chunks[i].End,
			},
			ChildPrompt: r.opts.Prompt,
			Provider:    r.worker.Name(),
			Cost:        mr.cost,
		}
		if res.Err != nil {
			call.TerminationReason = TerminationError
			call.Error = res.Err.Error()
		} else {
			call.TerminationReason = TerminationCompleted
		}
		r.trace.Calls = append(r.trace.Calls, call)
		results = append(results, mr)

		if res.Err == nil && r.opts.MaxTokens > 0 {
			r.budget -= mr.cost.TotalTokens
			if r.budget < 0 {
				r.trace.TerminationReason = TerminationMaxTokens
			}
		}
	}

	if runErr != nil {
		r.trace.TerminationReason = TerminationError
		return results, runErr
	}
	return results, nil
}

func ruleName(rule string) string {
	if rule == "" {
		return RuleMajority
	}
	return rule
}
