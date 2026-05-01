# PE Test Coverage Report

Generated: 2026-04-30

## Baseline

Command used for the measured baseline:

```sh
go test -coverprofile=/tmp/pe-coverage.out ./...
go tool cover -func=/tmp/pe-coverage.out
```

Verification command:

```sh
go test -cover ./...
```

Results:

| Metric | Value |
| --- | ---: |
| Overall statement coverage | 40.8% |
| Package paths reported by `go test -cover` | 41 |
| Packages with statements | 37 |
| Packages with no statements | 4 |
| Test status | Pass |

This supersedes older documentation claims that described overall coverage as
either about 25% or about 40%.

## Low-Coverage Packages

Packages below 30% statement coverage are the current priority for test
expansion.

| Package | Coverage |
| --- | ---: |
| `github.com/tmc/pe/example/structured/with-go-structs` | 0.0% |
| `github.com/tmc/pe/examples/inference` | 0.0% |
| `github.com/tmc/pe/ext/starlark/cmd/starlark-demo` | 0.0% |
| `github.com/tmc/pe/internal/cli` | 0.0% |
| `github.com/tmc/pe/internal/optimization` | 0.0% |
| `github.com/tmc/pe/internal/optimization/optimizers` | 0.0% |
| `github.com/tmc/pe/internal/promptfoo/evaluation/testing` | 0.0% |
| `github.com/tmc/pe/internal/testing` | 0.0% |
| `github.com/tmc/pe/test` | 0.0% |
| `github.com/tmc/pe/internal/inference` | 19.6% |
| `github.com/tmc/pe/internal/promptfoo/evaluation/metrics` | 27.9% |
| `github.com/tmc/pe/internal/optimization/adapters` | 28.4% |

## Other Measured Packages

| Package | Coverage |
| --- | ---: |
| `github.com/tmc/pe/cmd/pe` | 30.0% |
| `github.com/tmc/pe/internal/llm` | 31.0% |
| `github.com/tmc/pe/ext/starlark` | 38.0% |
| `github.com/tmc/pe/internal/module` | 39.7% |
| `github.com/tmc/pe/internal/metaprompt` | 40.8% |
| `github.com/tmc/pe/internal/promptfoo/evaluation/evaluator` | 41.0% |
| `github.com/tmc/pe/internal/errors` | 46.0% |
| `github.com/tmc/pe/plugins/promptfoo` | 46.4% |
| `github.com/tmc/pe/internal/pemod` | 48.7% |
| `github.com/tmc/pe/internal/inference/providers/cgpt` | 51.3% |
| `github.com/tmc/pe/internal/providers` | 58.1% |
| `github.com/tmc/pe/internal/prompt` | 58.9% |
| `github.com/tmc/pe/internal/testing/mocks` | 60.4% |
| `github.com/tmc/pe/internal/promptfoo` | 61.5% |
| `github.com/tmc/pe/internal/observability` | 62.0% |
| `github.com/tmc/pe/internal/config` | 66.8% |
| `github.com/tmc/pe/internal/inference/providers/ollama` | 68.0% |
| `github.com/tmc/pe/internal/cgpt` | 69.3% |
| `github.com/tmc/pe/internal/inference/providers/anthropic` | 73.3% |
| `github.com/tmc/pe/internal/inference/providers/openai` | 74.0% |
| `github.com/tmc/pe/internal/starlark` | 81.5% |
| `github.com/tmc/pe/internal/plugin` | 83.6% |
| `github.com/tmc/pe/internal/promptfoo/security/redteam` | 86.4% |
| `github.com/tmc/pe/internal/structured` | 89.3% |
| `github.com/tmc/pe/internal/templates` | 93.5% |

## Packages With No Statements

| Package |
| --- |
| `github.com/tmc/pe/example/getting-started` |
| `github.com/tmc/pe/internal/inference/providers` |
| `github.com/tmc/pe/internal/security` |
| `github.com/tmc/pe/tests` |
