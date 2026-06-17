package rlm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// fakeWorker is a deterministic local worker for tests. It never calls a
// provider.
type fakeWorker struct {
	name      string
	respond   func(prompt string, c Chunk) (string, Cost, error)
	callCount int
}

func (w *fakeWorker) Name() string {
	if w.name == "" {
		return "fake:test"
	}
	return w.name
}

func (w *fakeWorker) Run(ctx context.Context, prompt string, c Chunk) (string, Cost, error) {
	w.callCount++
	if err := ctx.Err(); err != nil {
		return "", Cost{}, err
	}
	if w.respond != nil {
		return w.respond(prompt, c)
	}
	return string(c.Data), Cost{TotalTokens: len(c.Data)}, nil
}

func TestChunkSplitsContiguousRanges(t *testing.T) {
	data := []byte("abcdefg")
	chunks, err := chunk(data, 3)
	if err != nil {
		t.Fatalf("chunk: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("got %d chunks, want 3", len(chunks))
	}
	want := []struct {
		start, end int
		text       string
	}{
		{0, 3, "abc"},
		{3, 6, "def"},
		{6, 7, "g"},
	}
	for i, c := range chunks {
		if c.Start != want[i].start || c.End != want[i].end {
			t.Errorf("chunk %d range = (%d,%d), want (%d,%d)", i, c.Start, c.End, want[i].start, want[i].end)
		}
		if string(c.Data) != want[i].text {
			t.Errorf("chunk %d data = %q, want %q", i, c.Data, want[i].text)
		}
		if c.CacheKey != ContentKey([]byte(want[i].text)) {
			t.Errorf("chunk %d cache key mismatch", i)
		}
	}
}

func TestChunkRejectsNonPositiveSize(t *testing.T) {
	if _, err := chunk([]byte("x"), 0); err == nil {
		t.Fatal("expected error for zero chunk size")
	}
}

func TestBatchOutputs(t *testing.T) {
	cases := []struct {
		name    string
		outputs []string
		size    int
		want    []string
	}{
		{"empty", nil, 4, nil},
		{"single", []string{"ab"}, 4, []string{"ab"}},
		{"pairs fit", []string{"a", "b", "c", "d"}, 3, []string{"a\nb", "c\nd"}},
		{"oversize own batch", []string{"abcd", "e"}, 3, []string{"abcd", "e"}},
		{"all fit one batch", []string{"a", "b"}, 8, []string{"a\nb"}},
		{"no batching when each fills size", []string{"ab", "cd", "ef"}, 2, []string{"ab", "cd", "ef"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := batchOutputs(tc.outputs, tc.size)
			if len(got) != len(tc.want) {
				t.Fatalf("batchOutputs = %q, want %q", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("batch %d = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestRunMapsEveryChunkAndRecordsTrace(t *testing.T) {
	w := &fakeWorker{name: "fake:model"}
	// Each echoed output already fills a chunk, so the outputs cannot be batched
	// further. The run completes after the single map pass while recording every
	// map call.
	trace, err := Run(context.Background(), w, []byte("abcdefg"), Options{
		Prompt:    "summarize",
		ChunkSize: 3,
		Workers:   2,
		Rule:      RuleConcat,
		Command:   "pe exp recurse",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if trace.Version != TraceVersion {
		t.Errorf("version = %q, want %q", trace.Version, TraceVersion)
	}
	if trace.Target.SHA256 != ContentKey([]byte("abcdefg")) {
		t.Error("target sha256 mismatch")
	}
	if len(trace.Calls) != 3 {
		t.Fatalf("got %d calls, want 3", len(trace.Calls))
	}
	for i, call := range trace.Calls {
		if call.Operation != "map" {
			t.Errorf("call %d operation = %q, want map", i, call.Operation)
		}
		if call.Provider != "fake:model" {
			t.Errorf("call %d provider = %q", i, call.Provider)
		}
		if call.ChildPrompt != "summarize" {
			t.Errorf("call %d child prompt = %q", i, call.ChildPrompt)
		}
		if call.TerminationReason != TerminationCompleted {
			t.Errorf("call %d termination = %q", i, call.TerminationReason)
		}
	}
	if trace.Aggregation.Consensus != "abc\ndef\ng" {
		t.Errorf("concat consensus = %q", trace.Aggregation.Consensus)
	}
	if trace.TerminationReason != TerminationCompleted {
		t.Errorf("run termination = %q, want completed", trace.TerminationReason)
	}
}

func TestRunMajorityPicksHighestWeight(t *testing.T) {
	// Three chunks: two map to "yes", one to "no". Majority should pick "yes".
	w := &fakeWorker{
		respond: func(prompt string, c Chunk) (string, Cost, error) {
			if c.Index == 1 {
				return "no", Cost{}, nil
			}
			return "yes", Cost{}, nil
		},
	}
	trace, err := Run(context.Background(), w, []byte("aaabbbccc"), Options{
		Prompt:    "vote",
		ChunkSize: 3,
		Rule:      RuleMajority,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if trace.Aggregation.Rule != RuleMajority {
		t.Errorf("rule = %q", trace.Aggregation.Rule)
	}
	if trace.Aggregation.Consensus != "yes" {
		t.Errorf("consensus = %q, want yes", trace.Aggregation.Consensus)
	}
	if len(trace.Aggregation.Votes) != 3 {
		t.Errorf("votes = %d, want 3", len(trace.Aggregation.Votes))
	}
}

func TestRunDeterministicAcrossRuns(t *testing.T) {
	mk := func() *fakeWorker {
		return &fakeWorker{respond: func(prompt string, c Chunk) (string, Cost, error) {
			return strings.ToUpper(string(c.Data)), Cost{TotalTokens: 1}, nil
		}}
	}
	first, err := Run(context.Background(), mk(), []byte("abcdef"), Options{Prompt: "p", ChunkSize: 2, Rule: RuleConcat})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	second, err := Run(context.Background(), mk(), []byte("abcdef"), Options{Prompt: "p", ChunkSize: 2, Rule: RuleConcat})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if first.Aggregation.Consensus != second.Aggregation.Consensus {
		t.Errorf("nondeterministic consensus: %q vs %q", first.Aggregation.Consensus, second.Aggregation.Consensus)
	}
	if first.Aggregation.Consensus != "AB\nCD\nEF" {
		t.Errorf("consensus = %q", first.Aggregation.Consensus)
	}
}

func TestRunReducesWideInputThroughDepths(t *testing.T) {
	// Each chunk reduces to a single byte, so a wide input collapses through
	// successive reduce passes until one chunk remains.
	w := &fakeWorker{respond: func(prompt string, c Chunk) (string, Cost, error) {
		return "x", Cost{}, nil
	}}
	trace, err := Run(context.Background(), w, []byte("abcdefgh"), Options{
		Prompt:    "fold",
		ChunkSize: 4,
		MaxDepth:  5,
		Rule:      RuleConcat,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if trace.TerminationReason != TerminationCompleted {
		t.Errorf("termination = %q, want completed", trace.TerminationReason)
	}

	// Depth 0 is a map; every deeper pass is a reduce with a parent.
	var maxDepth int
	sawReduce := false
	for _, call := range trace.Calls {
		if call.Depth == 0 {
			if call.Operation != "map" {
				t.Errorf("depth-0 call %s operation = %q, want map", call.ID, call.Operation)
			}
			continue
		}
		sawReduce = true
		if call.Operation != "reduce" {
			t.Errorf("depth-%d call %s operation = %q, want reduce", call.Depth, call.ID, call.Operation)
		}
		if call.ParentID == "" {
			t.Errorf("reduce call %s missing parent", call.ID)
		}
		maxDepth = max(maxDepth, call.Depth)
	}
	if !sawReduce {
		t.Fatal("expected at least one reduce pass")
	}
	if maxDepth < 1 {
		t.Errorf("max recorded depth = %d, want >= 1", maxDepth)
	}
}

func TestRunHaltsAtMaxDepthWithWideOutput(t *testing.T) {
	// Outputs still batch (progress is possible) but a wide input needs more
	// reduce passes than MaxDepth allows, so the run halts at the depth cap.
	w := &fakeWorker{respond: func(prompt string, c Chunk) (string, Cost, error) {
		return "z", Cost{}, nil
	}}
	trace, err := Run(context.Background(), w, []byte("0123456789abcdef"), Options{
		Prompt:    "fold",
		ChunkSize: 3,
		MaxDepth:  1,
		Rule:      RuleConcat,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if trace.TerminationReason != TerminationMaxDepth {
		t.Errorf("termination = %q, want max_depth", trace.TerminationReason)
	}
	for _, call := range trace.Calls {
		if call.Depth > 1 {
			t.Errorf("call %s at depth %d exceeds MaxDepth 1", call.ID, call.Depth)
		}
	}
}

func TestRunEnforcesMaxChunks(t *testing.T) {
	w := &fakeWorker{}
	trace, err := Run(context.Background(), w, []byte("abcdefghij"), Options{
		Prompt:    "p",
		ChunkSize: 2,
		MaxChunks: 3,
		Rule:      RuleConcat,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(trace.Calls) != 3 {
		t.Fatalf("calls = %d, want 3 (max chunks)", len(trace.Calls))
	}
	if w.callCount != 3 {
		t.Errorf("worker called %d times, want 3", w.callCount)
	}
	if trace.TerminationReason != TerminationMaxChunks {
		t.Errorf("termination = %q, want max_chunks", trace.TerminationReason)
	}
}

func TestRunEnforcesMaxTokens(t *testing.T) {
	// Each chunk reports 10 tokens; a budget of 15 is exceeded after 2 chunks.
	w := &fakeWorker{respond: func(prompt string, c Chunk) (string, Cost, error) {
		return "ok", Cost{TotalTokens: 10}, nil
	}}
	trace, err := Run(context.Background(), w, []byte("aabbcc"), Options{
		Prompt:    "p",
		ChunkSize: 2,
		MaxTokens: 15,
		Workers:   1,
		Rule:      RuleConcat,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if trace.TerminationReason != TerminationMaxTokens {
		t.Errorf("termination = %q, want max_tokens", trace.TerminationReason)
	}
}

func TestRunPropagatesWorkerError(t *testing.T) {
	wantErr := errors.New("provider unavailable")
	w := &fakeWorker{respond: func(prompt string, c Chunk) (string, Cost, error) {
		return "", Cost{}, wantErr
	}}
	trace, err := Run(context.Background(), w, []byte("abcd"), Options{Prompt: "p", ChunkSize: 2})
	if err == nil {
		t.Fatal("expected worker error")
	}
	if trace.TerminationReason != TerminationError {
		t.Errorf("termination = %q, want error", trace.TerminationReason)
	}
	foundErrCall := false
	for _, call := range trace.Calls {
		if call.TerminationReason == TerminationError && call.Error != "" {
			foundErrCall = true
		}
	}
	if !foundErrCall {
		t.Error("expected at least one call recorded with error")
	}
}

func TestRunRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := &fakeWorker{}
	_, err := Run(ctx, w, []byte("abcd"), Options{Prompt: "p", ChunkSize: 2})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestRunValidatesOptions(t *testing.T) {
	w := &fakeWorker{}
	cases := []struct {
		name string
		opts Options
	}{
		{"empty prompt", Options{ChunkSize: 1}},
		{"negative chunk size", Options{Prompt: "p", ChunkSize: -1}},
		{"negative depth", Options{Prompt: "p", ChunkSize: 1, MaxDepth: -1}},
		{"negative tokens", Options{Prompt: "p", ChunkSize: 1, MaxTokens: -1}},
		{"negative workers", Options{Prompt: "p", ChunkSize: 1, Workers: -1}},
		{"negative max chunks", Options{Prompt: "p", ChunkSize: 1, MaxChunks: -1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Run(context.Background(), w, []byte("x"), tc.opts); err == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
		})
	}
}

func TestRunZeroChunkSizeUsesDefault(t *testing.T) {
	// The zero value, with only Prompt set, must be usable: ChunkSize 0 selects
	// DefaultChunkSize rather than failing validation.
	w := &fakeWorker{}
	trace, err := Run(context.Background(), w, []byte("short input"), Options{Prompt: "p"})
	if err != nil {
		t.Fatalf("Run with zero ChunkSize: %v", err)
	}
	if trace.Budget.MaxDepth != 0 {
		t.Errorf("max depth = %d, want 0", trace.Budget.MaxDepth)
	}
	// Input shorter than DefaultChunkSize yields a single chunk and one call.
	if len(trace.Calls) != 1 {
		t.Fatalf("calls = %d, want 1 for a sub-default-size input", len(trace.Calls))
	}
	if trace.Calls[0].Snippet.End != len("short input") {
		t.Errorf("single chunk end = %d, want %d", trace.Calls[0].Snippet.End, len("short input"))
	}
}

func TestRunRejectsNilWorker(t *testing.T) {
	if _, err := Run(context.Background(), nil, []byte("x"), Options{Prompt: "p", ChunkSize: 1}); err == nil {
		t.Fatal("expected error for nil worker")
	}
}

func TestRunEmptyTarget(t *testing.T) {
	w := &fakeWorker{}
	trace, err := Run(context.Background(), w, nil, Options{Prompt: "p", ChunkSize: 4})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(trace.Calls) != 0 {
		t.Errorf("calls = %d, want 0 for empty target", len(trace.Calls))
	}
	if w.callCount != 0 {
		t.Errorf("worker called %d times for empty target", w.callCount)
	}
	if trace.TerminationReason != TerminationCompleted {
		t.Errorf("termination = %q", trace.TerminationReason)
	}
}

func TestRunRejectsUnsupportedRule(t *testing.T) {
	w := &fakeWorker{}
	// An unsupported rule is an option validation error: it is rejected before
	// any worker call, so no trace is produced.
	if _, err := Run(context.Background(), w, []byte("abcd"), Options{Prompt: "p", ChunkSize: 2, Rule: "bogus"}); err == nil {
		t.Fatal("expected unsupported rule error")
	}
	if w.callCount != 0 {
		t.Errorf("worker called %d times for invalid rule, want 0", w.callCount)
	}
}

func ExampleRun() {
	worker := &fakeWorker{name: "local:upper", respond: func(prompt string, c Chunk) (string, Cost, error) {
		return strings.ToUpper(string(c.Data)), Cost{}, nil
	}}
	trace, err := Run(context.Background(), worker, []byte("hello world"), Options{
		Prompt:    "uppercase",
		ChunkSize: 5,
		Rule:      RuleConcat,
		Command:   "pe exp recurse",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(trace.Aggregation.Consensus)
	// Output:
	// HELLO
	//  WORL
	// D
}
