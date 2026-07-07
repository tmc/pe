package workflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/tmc/pe/internal/distributed"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

// Call is one model invocation requested by a script.
type Call struct {
	Prompt   string
	Provider string // provider spec; empty means the runner default
	Model    string
	System   string

	// Temperature and MaxTokens are passed through to the provider.
	// Zero means the provider default.
	Temperature float64
	MaxTokens   int

	Label string
	Phase string
}

// Result is the outcome of one model call.
type Result struct {
	Content string
	Cost    Cost

	// Provider optionally reports the provider spec that served the call,
	// for callers that resolve an empty Call.Provider to a default. It is
	// recorded in the trace.
	Provider string
}

// Caller executes model calls on behalf of a workflow script. Callers must be
// safe for concurrent use: parallel issues calls from multiple goroutines.
type Caller interface {
	Call(ctx context.Context, call Call) (Result, error)
}

// Default limits applied when the corresponding Options fields are zero.
const (
	DefaultWorkers  = 4
	DefaultMaxCalls = 64
)

// Options configures Run. Caller is required; the zero value is not usable.
type Options struct {
	// Caller executes generate and parallel calls.
	Caller Caller

	// Args is exposed to the script as the args global. Values must be
	// convertible to Starlark: nil, bool, numbers, strings, slices, and
	// string-keyed maps.
	Args map[string]interface{}

	// Workers bounds concurrent calls in parallel. Zero means DefaultWorkers.
	Workers int

	// MaxCalls bounds the total number of model calls. Zero means
	// DefaultMaxCalls; negative disables the bound.
	MaxCalls int

	// Check, if non-nil, is consulted before every call. An error denies the
	// call.
	Check func(Call) error

	// Logf receives phase and log messages. Nil discards them.
	Logf func(format string, args ...interface{})
}

// Outcome is the result of a run. Trace is populated even when the run fails.
type Outcome struct {
	// Value is the script result: the return value of main(args) when the
	// script defines main, otherwise the value of the global named result,
	// otherwise nil.
	Value interface{}

	Trace *Trace
}

// Run executes the workflow script src. The name is used in the trace and in
// error positions. The returned Outcome carries the trace even when Run
// returns an error, so callers can persist partial traces.
func Run(ctx context.Context, name string, src []byte, opts Options) (*Outcome, error) {
	if opts.Caller == nil {
		return nil, fmt.Errorf("workflow: nil Caller")
	}
	e := &engine{ctx: ctx, opts: opts}
	outcome := &Outcome{Trace: &Trace{
		Version: TraceVersion,
		Name:    name,
		Budget:  Budget{MaxCalls: e.maxCalls(), Workers: e.workers()},
	}}
	finish := func(err error) (*Outcome, error) {
		e.mu.Lock()
		defer e.mu.Unlock()
		sort.Slice(e.calls, func(i, j int) bool { return e.calls[i].ID < e.calls[j].ID })
		outcome.Trace.Calls = e.calls
		for _, c := range e.calls {
			outcome.Trace.Totals.add(c.Cost)
		}
		if err != nil {
			outcome.Trace.Error = err.Error()
		}
		return outcome, err
	}

	argsValue, err := toStarlark(argsOrNone(opts.Args))
	if err != nil {
		return finish(fmt.Errorf("workflow args: %w", err))
	}

	thread := &starlark.Thread{
		Name:  name,
		Print: func(_ *starlark.Thread, msg string) { e.logf("%s", msg) },
	}
	stop := context.AfterFunc(ctx, func() { thread.Cancel("workflow context cancelled") })
	defer stop()

	globals, err := starlark.ExecFileOptions(fileOptions(), thread, name, src, e.predeclared(argsValue))
	if err != nil {
		return finish(scriptError(err))
	}

	value := starlark.Value(starlark.None)
	if mainValue, ok := globals["main"]; ok {
		mainFn, ok := mainValue.(starlark.Callable)
		if !ok {
			return finish(fmt.Errorf("workflow %s: main is not callable", name))
		}
		value, err = starlark.Call(thread, mainFn, starlark.Tuple{argsValue}, nil)
		if err != nil {
			return finish(scriptError(err))
		}
	} else if resultValue, ok := globals["result"]; ok {
		value = resultValue
	}

	outcome.Value, err = fromStarlark(value)
	if err != nil {
		return finish(fmt.Errorf("workflow result: %w", err))
	}
	return finish(nil)
}

// Validate compiles the workflow script src without executing it.
func Validate(name string, src []byte) error {
	predeclared := predeclaredNames()
	_, _, err := starlark.SourceProgramOptions(fileOptions(), name, src, func(s string) bool {
		return predeclared[s]
	})
	if err != nil {
		return scriptError(err)
	}
	return nil
}

// fileOptions returns the Starlark dialect for workflow scripts. Loops and
// recursion are allowed because the call budget, not syntax, bounds a run.
func fileOptions() *syntax.FileOptions {
	return &syntax.FileOptions{
		Set:             true,
		While:           true,
		TopLevelControl: true,
		GlobalReassign:  true,
		Recursion:       true,
	}
}

// scriptError unwraps Starlark evaluation errors to include their backtrace.
func scriptError(err error) error {
	var evalErr *starlark.EvalError
	if errors.As(err, &evalErr) {
		return fmt.Errorf("workflow script: %s", evalErr.Backtrace())
	}
	return fmt.Errorf("workflow script: %w", err)
}

func argsOrNone(args map[string]interface{}) interface{} {
	if args == nil {
		return nil
	}
	return args
}

// engine holds the mutable state shared by the builtins of one run.
type engine struct {
	ctx  context.Context
	opts Options

	mu    sync.Mutex
	calls []CallTrace
	spent int
	phase string
}

func (e *engine) workers() int {
	if e.opts.Workers < 1 {
		return DefaultWorkers
	}
	return e.opts.Workers
}

func (e *engine) maxCalls() int {
	switch {
	case e.opts.MaxCalls == 0:
		return DefaultMaxCalls
	case e.opts.MaxCalls < 0:
		return -1
	default:
		return e.opts.MaxCalls
	}
}

func (e *engine) logf(format string, args ...interface{}) {
	if e.opts.Logf != nil {
		e.opts.Logf(format, args...)
	}
}

func (e *engine) currentPhase() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.phase
}

func (e *engine) setPhase(title string) {
	e.mu.Lock()
	e.phase = title
	e.mu.Unlock()
	e.logf("phase: %s", title)
}

// reserve claims n call IDs against the budget before any of the calls run.
func (e *engine) reserve(n int) (start int, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if max := e.maxCalls(); max >= 0 && e.spent+n > max {
		return 0, fmt.Errorf("call budget exceeded: %d more calls with %d of %d used", n, e.spent, max)
	}
	start = e.spent
	e.spent += n
	return start, nil
}

// call executes one model call and records it in the trace.
func (e *engine) call(ctx context.Context, id int, c Call) (string, error) {
	trace := CallTrace{
		ID:           id,
		Phase:        c.Phase,
		Label:        c.Label,
		Provider:     c.Provider,
		Model:        c.Model,
		PromptSHA256: promptDigest(c.Prompt),
		PromptBytes:  len(c.Prompt),
	}
	var content string
	err := func() error {
		if e.opts.Check != nil {
			if err := e.opts.Check(c); err != nil {
				return err
			}
		}
		result, err := e.opts.Caller.Call(ctx, c)
		if err != nil {
			return err
		}
		content = result.Content
		trace.Cost = result.Cost
		if result.Provider != "" {
			trace.Provider = result.Provider
		}
		return nil
	}()
	if err != nil {
		trace.Error = err.Error()
	}
	e.mu.Lock()
	e.calls = append(e.calls, trace)
	e.mu.Unlock()
	return content, err
}

func promptDigest(prompt string) string {
	sum := sha256.Sum256([]byte(prompt))
	return hex.EncodeToString(sum[:])
}

func predeclaredNames() map[string]bool {
	return map[string]bool{
		"args":     true,
		"generate": true,
		"plan":     true,
		"parallel": true,
		"phase":    true,
		"log":      true,
		"json":     true,
		"struct":   true,
	}
}

func (e *engine) predeclared(args starlark.Value) starlark.StringDict {
	return starlark.StringDict{
		"args":     args,
		"generate": starlark.NewBuiltin("generate", e.generateFn),
		"plan":     starlark.NewBuiltin("plan", e.planFn),
		"parallel": starlark.NewBuiltin("parallel", e.parallelFn),
		"phase":    starlark.NewBuiltin("phase", e.phaseFn),
		"log":      starlark.NewBuiltin("log", e.logFn),
		"json":     starlarkjson.Module,
		"struct":   starlark.NewBuiltin("struct", starlarkstruct.Make),
	}
}

// callValue is a planned model call, produced by plan and consumed by
// parallel.
type callValue struct {
	call Call
}

func (v callValue) String() string {
	prompt := v.call.Prompt
	if len(prompt) > 32 {
		prompt = prompt[:32] + "..."
	}
	return fmt.Sprintf("plan(%q)", prompt)
}

func (v callValue) Type() string          { return "workflow_call" }
func (v callValue) Freeze()               {}
func (v callValue) Truth() starlark.Bool  { return starlark.True }
func (v callValue) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: workflow_call") }

func (e *engine) unpackCall(fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (Call, error) {
	var prompt, provider, model, system, label string
	var temperature starlark.Value
	var maxTokens int
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"prompt", &prompt,
		"provider?", &provider,
		"model?", &model,
		"system?", &system,
		"temperature?", &temperature,
		"max_tokens?", &maxTokens,
		"label?", &label,
	); err != nil {
		return Call{}, err
	}
	call := Call{
		Prompt:    prompt,
		Provider:  provider,
		Model:     model,
		System:    system,
		MaxTokens: maxTokens,
		Label:     label,
		Phase:     e.currentPhase(),
	}
	if temperature != nil && temperature != starlark.None {
		f, ok := starlark.AsFloat(temperature)
		if !ok {
			return Call{}, fmt.Errorf("%s: temperature must be a number, got %s", fn.Name(), temperature.Type())
		}
		call.Temperature = f
	}
	return call, nil
}

func (e *engine) generateFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	call, err := e.unpackCall(fn, args, kwargs)
	if err != nil {
		return nil, err
	}
	id, err := e.reserve(1)
	if err != nil {
		return nil, err
	}
	content, err := e.call(e.ctx, id, call)
	if err != nil {
		return nil, err
	}
	return starlark.String(content), nil
}

func (e *engine) planFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	call, err := e.unpackCall(fn, args, kwargs)
	if err != nil {
		return nil, err
	}
	return callValue{call: call}, nil
}

func (e *engine) parallelFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var callsValue starlark.Value
	failFast := true
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"calls", &callsValue,
		"fail_fast?", &failFast,
	); err != nil {
		return nil, err
	}

	iter := starlark.Iterate(callsValue)
	if iter == nil {
		return nil, fmt.Errorf("%s: calls must be iterable, got %s", fn.Name(), callsValue.Type())
	}
	var calls []Call
	var elem starlark.Value
	for iter.Next(&elem) {
		switch v := elem.(type) {
		case callValue:
			calls = append(calls, v.call)
		case starlark.String:
			calls = append(calls, Call{Prompt: string(v), Phase: e.currentPhase()})
		default:
			iter.Done()
			return nil, fmt.Errorf("%s: calls must contain plan() values or strings, got %s", fn.Name(), elem.Type())
		}
	}
	iter.Done()

	if len(calls) == 0 {
		return starlark.NewList(nil), nil
	}
	start, err := e.reserve(len(calls))
	if err != nil {
		return nil, err
	}

	tasks := make([]distributed.Task[*string], len(calls))
	for i, call := range calls {
		id := start + i
		call := call
		tasks[i] = distributed.Task[*string]{
			ID: fmt.Sprintf("call-%d", id),
			Run: func(ctx context.Context) (*string, error) {
				content, err := e.call(ctx, id, call)
				if err != nil {
					if failFast {
						return nil, err
					}
					return nil, nil
				}
				return &content, nil
			},
		}
	}
	results, err := distributed.RunLocal(e.ctx, e.workers(), tasks)
	if err != nil {
		return nil, err
	}
	values := make([]starlark.Value, len(results))
	for i, r := range results {
		if r.Value == nil {
			values[i] = starlark.None
			continue
		}
		values[i] = starlark.String(*r.Value)
	}
	return starlark.NewList(values), nil
}

func (e *engine) phaseFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var title string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "title", &title); err != nil {
		return nil, err
	}
	e.setPhase(title)
	return starlark.None, nil
}

func (e *engine) logFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var msg string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "msg", &msg); err != nil {
		return nil, err
	}
	e.logf("%s", msg)
	return starlark.None, nil
}
