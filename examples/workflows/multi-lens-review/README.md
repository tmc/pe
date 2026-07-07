# Workflow Scripts

Workflow scripts compose model calls with ordinary control flow. The script
runs deterministically and single-threaded; only model calls run concurrently,
through `parallel()`, under a bounded worker pool. Every call is recorded in a
strict JSON trace (`pe.workflow.trace.v1`).

## Files

- `review.star` — plain Starlark script: parallel multi-lens review, then a
  synthesis call.
- `audit.pe.md` — the same idea as executable text: `pe.workflow.v1` front
  matter declares a default input and a provider safety policy (`local` only);
  the body is the script.

## Run

With a local Ollama daemon and a small model:

```sh
pe exp workflow run review.star --provider ollama:llama3.2:3b \
    --args '{"code": "func add(a, b int) int { return a + b }"}'

pe exp workflow run audit.pe.md --provider ollama:llama3.2:3b \
    --var "topic=slice aliasing"
```

Add `--trace trace.json` to record every call (phase, label, provider, prompt
digest, token cost). Validate without running:

```sh
pe exp workflow validate review.star
```

## Policy

Provider use composes conservatively: `pe.mod` capability policy
(`providers deny ...`, `placement network false`) and the script's own
`safety.providers` allow/deny lists are both enforced before every call.
`audit.pe.md` allows only local providers, so pointing it at `openai:...` or
`anthropic:...` fails before any call is made.

## Validation record (2026-07-06)

`review.star` ran end to end against a live Ollama daemon with
`ollama:llama3.2:3b`: two review calls executed in parallel, the synthesis
call consumed their outputs, and the trace recorded per-call phases, labels,
the resolved provider, and real token costs (238 total tokens).
