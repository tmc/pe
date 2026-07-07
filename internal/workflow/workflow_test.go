package workflow

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// echoCaller answers every call with "echo:" + prompt and a fixed cost. It
// fails on prompts containing "boom" and can track concurrent calls.
type echoCaller struct {
	delay time.Duration

	mu       sync.Mutex
	inflight int
	maxSeen  int
	calls    int32
}

func (c *echoCaller) Call(ctx context.Context, call Call) (Result, error) {
	atomic.AddInt32(&c.calls, 1)
	c.mu.Lock()
	c.inflight++
	if c.inflight > c.maxSeen {
		c.maxSeen = c.inflight
	}
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.inflight--
		c.mu.Unlock()
	}()

	if c.delay > 0 {
		select {
		case <-time.After(c.delay):
		case <-ctx.Done():
			return Result{}, ctx.Err()
		}
	}
	if strings.Contains(call.Prompt, "boom") {
		return Result{}, fmt.Errorf("provider failure for %q", call.Prompt)
	}
	return Result{
		Content: "echo:" + call.Prompt,
		Cost:    Cost{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
	}, nil
}

func run(t *testing.T, script string, opts Options) *Outcome {
	t.Helper()
	if opts.Caller == nil {
		opts.Caller = &echoCaller{}
	}
	outcome, err := Run(context.Background(), t.Name()+".star", []byte(script), opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return outcome
}

func TestRunResultGlobal(t *testing.T) {
	outcome := run(t, `result = generate("hi")`, Options{})
	if got, want := outcome.Value, "echo:hi"; got != want {
		t.Errorf("Value = %v, want %v", got, want)
	}
	if len(outcome.Trace.Calls) != 1 {
		t.Fatalf("trace calls = %d, want 1", len(outcome.Trace.Calls))
	}
	call := outcome.Trace.Calls[0]
	if call.PromptBytes != 2 || call.PromptSHA256 == "" {
		t.Errorf("call prompt record = %+v", call)
	}
	if outcome.Trace.Totals.TotalTokens != 2 {
		t.Errorf("Totals.TotalTokens = %d, want 2", outcome.Trace.Totals.TotalTokens)
	}
	if outcome.Trace.Version != TraceVersion {
		t.Errorf("trace version = %q", outcome.Trace.Version)
	}
}

func TestRunMain(t *testing.T) {
	script := `
def main(args):
    return generate(args["q"])
`
	outcome := run(t, script, Options{Args: map[string]interface{}{"q": "question"}})
	if got, want := outcome.Value, "echo:question"; got != want {
		t.Errorf("Value = %v, want %v", got, want)
	}
}

func TestRunNoResult(t *testing.T) {
	outcome := run(t, `x = 1`, Options{})
	if outcome.Value != nil {
		t.Errorf("Value = %v, want nil", outcome.Value)
	}
}

func TestRunParallel(t *testing.T) {
	script := `
reqs = [plan("p" + str(i)) for i in range(5)]
result = parallel(reqs)
`
	outcome := run(t, script, Options{})
	want := []interface{}{"echo:p0", "echo:p1", "echo:p2", "echo:p3", "echo:p4"}
	if !reflect.DeepEqual(outcome.Value, want) {
		t.Errorf("Value = %v, want %v", outcome.Value, want)
	}
}

func TestRunParallelStrings(t *testing.T) {
	outcome := run(t, `result = parallel(["a", "b"])`, Options{})
	want := []interface{}{"echo:a", "echo:b"}
	if !reflect.DeepEqual(outcome.Value, want) {
		t.Errorf("Value = %v, want %v", outcome.Value, want)
	}
}

func TestRunParallelConcurrencyCap(t *testing.T) {
	caller := &echoCaller{delay: 10 * time.Millisecond}
	script := `result = parallel(["a", "b", "c", "d", "e", "f"])`
	run(t, script, Options{Caller: caller, Workers: 2})
	if caller.maxSeen > 2 {
		t.Errorf("max concurrent calls = %d, want <= 2", caller.maxSeen)
	}
	if caller.calls != 6 {
		t.Errorf("calls = %d, want 6", caller.calls)
	}
}

func TestRunParallelFailFast(t *testing.T) {
	_, err := Run(context.Background(), "t.star", []byte(`result = parallel(["ok", "boom"])`), Options{Caller: &echoCaller{}})
	if err == nil || !strings.Contains(err.Error(), "provider failure") {
		t.Fatalf("err = %v, want provider failure", err)
	}
}

func TestRunParallelNoFailFast(t *testing.T) {
	outcome := run(t, `result = parallel(["ok", "boom"], fail_fast=False)`, Options{})
	want := []interface{}{"echo:ok", nil}
	if !reflect.DeepEqual(outcome.Value, want) {
		t.Errorf("Value = %v, want %v", outcome.Value, want)
	}
	var traced bool
	for _, c := range outcome.Trace.Calls {
		if c.Error != "" {
			traced = true
		}
	}
	if !traced {
		t.Error("failed call not recorded in trace")
	}
}

func TestRunCallBudget(t *testing.T) {
	script := `
for i in range(3):
    generate("p" + str(i))
`
	outcome, err := Run(context.Background(), "t.star", []byte(script), Options{Caller: &echoCaller{}, MaxCalls: 2})
	if err == nil || !strings.Contains(err.Error(), "call budget exceeded") {
		t.Fatalf("err = %v, want call budget exceeded", err)
	}
	if outcome == nil || outcome.Trace.Error == "" {
		t.Error("trace should record the budget error")
	}
	if got := len(outcome.Trace.Calls); got != 2 {
		t.Errorf("trace calls = %d, want 2", got)
	}
}

func TestRunParallelBudgetCheckedUpfront(t *testing.T) {
	caller := &echoCaller{}
	script := `result = parallel(["a", "b", "c"])`
	_, err := Run(context.Background(), "t.star", []byte(script), Options{Caller: caller, MaxCalls: 2})
	if err == nil || !strings.Contains(err.Error(), "call budget exceeded") {
		t.Fatalf("err = %v, want call budget exceeded", err)
	}
	if caller.calls != 0 {
		t.Errorf("calls = %d, want 0 (budget enforced before any call)", caller.calls)
	}
}

func TestRunCheckDenies(t *testing.T) {
	opts := Options{
		Caller: &echoCaller{},
		Check: func(c Call) error {
			if c.Provider == "denied" {
				return fmt.Errorf("provider %s is denied", c.Provider)
			}
			return nil
		},
	}
	_, err := Run(context.Background(), "t.star", []byte(`result = generate("hi", provider="denied")`), opts)
	if err == nil || !strings.Contains(err.Error(), "is denied") {
		t.Fatalf("err = %v, want denial", err)
	}
}

func TestRunPhaseAndLog(t *testing.T) {
	var logs []string
	script := `
phase("Review")
generate("a")
log("done")
result = True
`
	outcome := run(t, script, Options{Logf: func(format string, args ...interface{}) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}})
	if got, want := outcome.Trace.Calls[0].Phase, "Review"; got != want {
		t.Errorf("call phase = %q, want %q", got, want)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "phase: Review") || !strings.Contains(joined, "done") {
		t.Errorf("logs = %q", joined)
	}
}

func TestRunWhileLoop(t *testing.T) {
	script := `
n = 0
while n < 3:
    n += 1
result = n
`
	outcome := run(t, script, Options{})
	if got, want := outcome.Value, int64(3); got != want {
		t.Errorf("Value = %v, want %v", got, want)
	}
}

func TestRunNoClock(t *testing.T) {
	_, err := Run(context.Background(), "t.star", []byte(`result = time.now()`), Options{Caller: &echoCaller{}})
	if err == nil || !strings.Contains(err.Error(), "time") {
		t.Fatalf("err = %v, want undefined time", err)
	}
}

func TestRunJSONModule(t *testing.T) {
	outcome := run(t, `result = json.decode('{"a": 1}')["a"]`, Options{})
	if got, want := outcome.Value, int64(1); got != want {
		t.Errorf("Value = %v, want %v", got, want)
	}
}

func TestRunStructResult(t *testing.T) {
	outcome := run(t, `result = {"items": [1, 2], "ok": True}`, Options{})
	want := map[string]interface{}{"items": []interface{}{int64(1), int64(2)}, "ok": true}
	if !reflect.DeepEqual(outcome.Value, want) {
		t.Errorf("Value = %#v, want %#v", outcome.Value, want)
	}
}

func TestRunNilCaller(t *testing.T) {
	_, err := Run(context.Background(), "t.star", []byte(`result = 1`), Options{})
	if err == nil || !strings.Contains(err.Error(), "nil Caller") {
		t.Fatalf("err = %v, want nil Caller", err)
	}
}

func TestRunContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx, "t.star", []byte(`result = parallel(["a", "b"])`), Options{Caller: &echoCaller{delay: 50 * time.Millisecond}})
	if err == nil {
		t.Fatal("err = nil, want cancellation error")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{"valid", `result = generate("hi")`, ""},
		{"valid main", "def main(args):\n    return parallel([plan(\"a\")])\n", ""},
		{"syntax error", `result = (`, "workflow script"},
		{"undefined name", `result = fetch_url("x")`, "undefined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.name+".star", []byte(tt.script))
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
