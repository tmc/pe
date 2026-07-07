// Package workflow runs deterministic prompt-orchestration scripts.
//
// A workflow script is a Starlark program that composes model calls with
// ordinary control flow: loops, conditionals, and functions. The script itself
// runs single-threaded and deterministically; only model calls run
// concurrently, through parallel, under a bounded worker pool. Budgets are
// enforced before any call, and every call is recorded in a strict JSON trace
// (schema pe.workflow.trace.v1).
//
// Scripts see these predeclared names:
//
//	args                the value passed in Options.Args
//	generate(prompt, provider=, model=, system=, temperature=, max_tokens=, label=)
//	                    execute one model call and return its text
//	plan(...)           same arguments as generate, but return a call value
//	                    without executing it
//	parallel(calls, fail_fast=True)
//	                    execute call values (or bare prompt strings)
//	                    concurrently; results keep input order. With
//	                    fail_fast=False a failed call yields None instead of
//	                    stopping the run.
//	phase(title)        group subsequent calls under title in the trace
//	log(msg)            emit a progress message
//	json, struct        Starlark JSON module and struct constructor
//
// The script result is the return value of main(args) when the script defines
// main, otherwise the value of the global named result, otherwise None.
//
// Starlark has no clock, randomness, or filesystem access, so a script's
// control flow depends only on args and model outputs. Traces make runs
// auditable and replayable.
package workflow
