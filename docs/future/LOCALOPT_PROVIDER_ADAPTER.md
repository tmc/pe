# Localopt Provider Adapter

Status: deferred after v0.5.

`internal/optimization/localopt` is intentionally provider-free. It accepts a
scorer and a refiner, then selects the best improving prompt variant with
deterministic local control flow. Provider-backed semantic refinement should sit
outside that package.

## Boundary

Keep the dependency direction:

- `internal/optimization/localopt` defines local optimization data types and
  interfaces.
- `internal/metaprompt` or a CLI layer adapts model providers into
  `localopt.Refiner`.
- Provider registry packages construct providers, but are not imported by
  `localopt`.

The adapter contract should be:

```go
type CandidateGenerator interface {
	GenerateCandidates(context.Context, localopt.Variant) ([]localopt.Variant, error)
}
```

A provider-backed implementation can format a semantic refinement prompt, call
the existing LLM provider abstraction, parse JSON candidates, and return
`localopt.Variant` values. Scoring remains supplied separately so offline eval
results, promptfoo metrics, and model-generated candidates can be composed
without coupling local optimization to a provider runtime.

## Candidate JSON

Provider output should be constrained to a small JSON object:

```json
{
  "candidates": [
    {
      "prompt": "candidate prompt text",
      "source": "provider-or-strategy-label",
      "metadata": {
        "reason": "short rationale"
      }
    }
  ]
}
```

The adapter should reject empty candidate sets, ignore blank prompt candidates,
and wrap provider or parse errors with local context.

## Deferral Reason

This is not a v0.5 release requirement. NotebookLM Phase 11 direction prioritizes
the promptfoo script fix, distributed CLI integration, consensus integration,
and final release gates before more semantic optimization code. Adding the
adapter now would increase API surface before the v0.5 release path is stable.

Revisit after v0.5 when the promptfoo gate and integrated CLI surfaces are green.
