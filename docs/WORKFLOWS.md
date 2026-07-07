# Workflow Scripts

`pe exp workflow` runs deterministic prompt-orchestration scripts: Starlark
programs that compose model calls with ordinary control flow (loops,
conditionals, functions). The script itself runs single-threaded and
deterministically; only model calls run concurrently, through `parallel()`,
under a bounded worker pool. Budgets are enforced before any call, and every
call is recorded in a strict JSON trace (schema `pe.workflow.trace.v1`).

Status: experimental (`pe exp`). The builtin surface may change while the
executable-text design settles.

## Script model

Scripts see these predeclared names:

| Name | Meaning |
|------|---------|
| `args` | Value of `--args` JSON merged with `--var` overrides and front matter input defaults |
| `generate(prompt, provider=, model=, system=, temperature=, max_tokens=, label=)` | Execute one model call, return its text |
| `plan(...)` | Same arguments as `generate`, but return a call value without executing |
| `parallel(calls, fail_fast=True)` | Execute call values (or bare prompt strings) concurrently; results keep input order; with `fail_fast=False` a failed call yields `None` |
| `phase(title)` | Group subsequent calls under `title` in the trace |
| `log(msg)` | Emit a progress message to stderr |
| `json`, `struct` | Starlark JSON module and struct constructor |

The script result is the return value of `main(args)` when the script defines
`main`, otherwise the value of the global named `result`. String results print
verbatim; other values print as JSON.

Starlark has no clock, randomness, or filesystem access, so a script's control
flow depends only on `args` and model outputs.

## Example

```python
phase("Review")
findings = parallel([
    plan(lens + ": review this diff:\n" + args["diff"], label=lens)
    for lens in ["bugs", "perf", "style"]
])

phase("Synthesize")
result = generate("Merge into one verdict:\n" + "\n".join(findings))
```

```sh
pe exp workflow run review.star --provider ollama:llama3.2:3b \
    --args '{"diff": "..."}' --trace trace.json
```

## File forms

- A plain Starlark file (conventionally `.star`) is the script verbatim.
- Executable text with `kind: pe.workflow.v1` front matter: the body is the
  script; declared `inputs` defaults seed `args`, and `safety.providers`
  allow/deny lists are enforced on every call.

`pe exp workflow validate <file>` compiles a script (and checks front matter)
without executing it.

## Budgets and policy

- `--max-calls` (default 64) caps total model calls; the cap is checked before
  a call or a whole `parallel()` batch runs, never mid-batch.
- `--workers` (default 4) bounds concurrent calls.
- Provider policy composes conservatively: `pe.mod` capability policy
  (`providers deny ...`, `placement network false`) and the script's own
  `safety.providers` lists are both enforced before every call, including
  per-call `provider=` overrides.

## Trace

`--trace <file|- >` emits `pe.workflow.trace.v1`: run budget, one record per
call (phase, label, resolved provider, model, prompt digest and size, token
cost, error), and run totals. Prompts are referenced by SHA-256 digest, not
inlined.

See `examples/workflows/multi-lens-review/` for runnable examples validated
against a live local Ollama daemon.
