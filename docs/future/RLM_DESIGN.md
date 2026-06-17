# PE-Native RLM Design

Status: v0.6+ design track. This is not a v0.5 release blocker. An initial
implementation has landed: `internal/rlm` provides the bounded combinators,
runner, deterministic aggregation, content addressing, and trace schema
(`pe.rlm.trace.v1`), and `pe exp recurse` exposes them as an experimental,
local-first command. See the Package Plan below for what each slice covers.

PE should not copy OpenProse's session-VM shape or introduce a new workflow DSL
before the core release path is stable. PE can do better by treating recursive
language-model work as a small set of Unix/Go-style operations over durable
artifacts: bounded combinators, local workers, deterministic aggregation, and
strict traces.

## Thesis

A PE-native recursive language model runner should make long-context work
auditable instead of magical. Large inputs stay outside the prompt and are
referenced by cache keys or attest manifests. Workers inspect bounded chunks.
Aggregators merge typed child results with deterministic rules. Every run emits
a JSON trace that can be checked by tests and release gates.

This design rejects arbitrary model-generated code execution. Models may produce
answers inside fixed operations, but they do not invent new programs for PE to
run.

## Future Command Shape

```sh
pe exp recurse <target-file> --prompt <instruction> --max-depth N --max-tokens N
```

The first version should remain experimental and local. Provider configuration
can follow existing PE provider conventions, but the runner must enforce its own
budgets before any provider call is made.

Required controls:

- `--max-depth`: maximum recursive expansion depth.
- `--max-tokens`: prompt or response budget per worker.
- `--workers`: bounded local worker count.
- `--trace`: output path for the JSON trace.
- `--consensus`: aggregation rule, initially deterministic majority.

## Bounded Combinators

The runner should expose only typed, bounded operations:

- `chunk`: split a target artifact into byte or line ranges.
- `map`: run one prompt over each selected chunk.
- `reduce`: combine bounded intermediate results.
- `search`: select chunks for another bounded pass.
- `recurse`: repeat a known operation while depth and budget remain.
- `aggregate`: merge child records into a parent record.
- `consensus`: resolve competing child answers with deterministic voting.

These are Go interfaces and command behaviors, not a new prompt language.

## Artifact Model

Large inputs should be stored out-of-prompt:

- cache entries hold content or chunk payloads by key.
- attest manifests record paths, sizes, and SHA-256 digests.
- prompts receive stable pointers instead of entire payloads.
- traces record every cache key and manifest used by a run.

This makes recursive context processing inspectable and reproducible without
forcing giant prompts through every model call.

## Trace Schema

Every run should emit strict JSON. Field names are stable so tests and release
checks can parse them.

```json
{
  "version": "pe.rlm.trace.v1",
  "command": "pe exp recurse",
  "target": {
    "path": "input.md",
    "manifest": "pe.unsigned_file_manifest.v1",
    "sha256": "..."
  },
  "budget": {
    "max_depth": 2,
    "max_tokens": 2000,
    "workers": 4
  },
  "calls": [
    {
      "id": "call-1",
      "parent_id": "",
      "depth": 0,
      "operation": "map",
      "snippet": {
        "cache_key": "...",
        "start": 0,
        "end": 1024
      },
      "child_prompt": "summarize this chunk",
      "provider": "provider:model",
      "cost": {
        "prompt_tokens": 0,
        "completion_tokens": 0,
        "total_tokens": 0
      },
      "termination_reason": "completed"
    }
  ],
  "aggregation": {
    "rule": "majority",
    "consensus": "selected answer",
    "votes": []
  },
  "termination_reason": "max_depth"
}
```

Minimum required trace fields:

- calls
- snippets read
- child prompts
- costs
- depth
- termination reason
- aggregation rule
- cache keys and attest manifests

## Package Plan

Implementation is split after v0.5. The following slices have landed:

1. `docs/future/RLM_DESIGN.md`: schema and contract.
2. `internal/rlm/combinators.go`: typed combinator interfaces and budgets
   (`chunk`, `Worker`, `reduce`).
3. `internal/rlm/runner.go`: bounded runner using `distributed.RunLocal`, with a
   hierarchical reduce loop that folds outputs while depth and budget remain.
4. `internal/rlm/aggregate.go`: child result aggregation through
   `distributed.Majority` (majority and concat rules).
5. `internal/rlm/storage.go`: content-key helper for out-of-prompt addressing.
6. `cmd/pe/exp_recurse.go`: experimental CLI harness.

Each slice is testable without live providers: the runner depends on a [Worker]
interface, not a concrete provider, and the package tests drive it with a
deterministic fake. Live-provider behavior belongs behind opt-in integration
tests.

Still open: a `search` combinator that selects chunks for another pass, attest
manifest references on the target, and trace redaction for provider secrets.

## Safety Model

- v0.6+ only; none of this blocks v0.5.
- No new DSL until the Go interfaces and JSON traces are proven.
- No arbitrary model-generated code execution.
- Hard limits on depth, workers, token budgets, and selected chunks.
- Deterministic aggregation for repeated child outputs.
- Trace redaction for provider secrets and environment-derived credentials.
- Replay should be possible from cache keys, attest manifests, and trace data
  when provider calls are stubbed or cached.
