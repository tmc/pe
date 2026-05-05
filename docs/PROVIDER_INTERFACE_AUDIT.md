# Provider Interface Audit

Last run: 2026-05-05
Branch: `exp`

PE currently has two provider interfaces during migration:

- `internal/llm.Provider`: legacy interface used by most evaluation,
  optimization, promptfoo-compatible, and metaprompting paths.
- `internal/inference.Provider`: newer request/response interface used by
  `pe run`, `pe ask`, native local-runtime providers, and newer test mocks.

`internal/inference/migration.go` provides adapters in both directions:

- `NewLegacyAdapter(llm.Provider) inference.Provider`
- `NewModernAdapter(inference.Provider, model) llm.Provider`

## Commands And Interfaces

| Area | Current provider interface | Evidence |
| --- | --- | --- |
| `pe run` | `inference.Provider`, with `llm.Provider` specs adapted through `NewLegacyAdapter` | `cmd/pe/run.go` |
| `pe ask` | `inference.Provider`, with `llm.Provider` specs adapted through `NewLegacyAdapter` | `cmd/pe/pipeline.go` |
| `pe eval` and Promptfoo evaluator | `llm.Provider` | `internal/promptfoo/evaluation/evaluator/evaluator.go` |
| `pe benchmark` | `llm.Provider` materialized from promptfoo provider specs | `cmd/pe/benchmark.go`, `internal/providers/materialize.go` |
| `pe optimize`, `pe semantic`, `pe evolve` | `llm.Provider` | `cmd/pe/optimize.go`, `cmd/pe/semantic.go`, `cmd/pe/evolve.go` |
| `pe security` | `llm.Provider` | `cmd/pe/security.go` |
| `pe passn` and advanced test commands | `llm.Provider` | `cmd/pe/passn.go`, `cmd/pe/test.go` |
| Local runtime provider bridge | `inference.Provider` adapted to `llm.Provider` | `internal/providers/init.go` |

## Provider Implementations

| Provider | Current implementation path | Native interface |
| --- | --- | --- |
| OpenAI | `internal/providers/openai.go`; `internal/inference/providers/openai` also exists | Both, duplicate paths |
| Anthropic | `internal/providers/anthropic.go`; `internal/inference/providers/anthropic` also exists | Both, duplicate paths |
| cgpt | `internal/providers/cgpt.go`; `internal/inference/providers/cgpt` also exists | Both, duplicate paths |
| Ollama | `internal/inference/providers/ollama` adapted through `NewModernAdapter` | `inference.Provider` |
| Mock | `internal/providers/mock.go`; `internal/testing/mocks` also exists | Both, test-oriented |
| Generic CLI | `internal/providers/cli.go` | `llm.Provider` |
| llm CLI | `internal/providers/llm_cli.go` | `llm.Provider` |

## Migration Complexity

| Area | Complexity | Notes |
| --- | --- | --- |
| `pe run` and `pe ask` | Low | Already use `inference.Client`; provider specs are bridged through `llm.GetProviderWithOptions`. |
| Evaluator and promptfoo compatibility | High | `llm.Provider` is part of assertion, judge-provider, metric, property, regression, and materialization paths. |
| Optimization/metaprompting | High | Optimizers store `llm.Provider` directly and expose constructors around that type. |
| Native providers | Medium | OpenAI, Anthropic, and cgpt have duplicate legacy and inference implementations that need one owner. |
| Local runtimes | Medium | Ollama is native `inference.Provider`; presets are exposed through the legacy registry for compatibility. |
| Tests/mocks | Medium | Both mock systems are useful; consolidation should preserve deterministic script tests. |

## Compatibility Matrix

| Feature | `llm.Provider` | `inference.Provider` | Migration concern |
| --- | --- | --- | --- |
| Non-streaming completion | Yes | Yes | Straightforward adapter path exists. |
| Streaming | Interface method exists on several legacy providers, but usage varies | First-class `Stream` method | Preserve command behavior and tests. |
| Model listing | `Model()` returns active model only | `Models(ctx)` returns list | Adapter currently returns limited defaults. |
| Provider options | `GenerateOptions` plus provider config maps | `Request` fields plus config maps | Need consistent option names. |
| Token/cost metadata | `GenerateResponse` / `ProviderResponse` fields | `Response` metadata and token usage | Avoid losing evaluator metrics. |
| Batch support | Legacy interface has capability methods | New interface exposes capabilities through provider methods | Low current CLI reliance. |

## Breaking-Change Risk

Do not remove `internal/llm.Provider` until these are true:

1. Promptfoo evaluator and assertion code no longer require it.
2. Optimization and metaprompt constructors have migration shims or new APIs.
3. Native OpenAI, Anthropic, and cgpt have one canonical implementation each.
4. Script tests pass for `run`, `ask`, `eval`, `benchmark`, `security`, and provider-backed examples.
5. Live-provider smokes pass for at least one native remote provider.

## Verification

Commands used for this audit:

```bash
rg -n "llm\.Provider|inference\.Provider|llm\.GetProvider|llm\.GetProviderWithOptions|inference\.NewClient|inference\.NewProvider|providers\.CreateProvider|MaterializeProvider|NewLegacyAdapter|NewModernAdapter" cmd internal -g '*.go'
GOTOOLCHAIN=go1.25.9 go list ./cmd/pe ./internal/llm ./internal/inference ./internal/providers ./internal/inference/providers/...
GOTOOLCHAIN=go1.25.9 go test ./cmd/pe ./internal/inference ./internal/providers ./internal/inference/providers/... -count=1
```
