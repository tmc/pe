# PE Test Coverage Report

Generated: 2026-05-05

## Baseline

Integrated baseline branch:

```text
exp
```

Integrated commit:

```text
03152ba
```

Commands used for the measured baseline:

```sh
GOTOOLCHAIN=go1.25.9 go test -coverprofile=/tmp/pe-coverage.out ./...
go tool cover -func=/tmp/pe-coverage.out
```

Results:

| Metric | Value |
| --- | ---: |
| Overall statement coverage | 44.4% |
| Package paths reported by `go test -cover` | 44 |
| Packages with statements | 41 |
| Packages with no statements | 3 |
| Test status | Passing integrated branch |

This supersedes older documentation claims that described overall coverage as
either about 25%, 40%, or 40.8%.

## Script Test Coverage

The script-test suite is connected to the release gate through
`GOTOOLCHAIN=go1.25.9 go test -coverprofile=/tmp/pe-coverage.out ./...`, and
the integrated run passed `github.com/tmc/pe/tests`.

Script tests validate CLI behavior by building and executing a separate `pe`
binary from `tests/scripttest_test.go`. That child process is not instrumented by
the parent `go test` coverage profile, so script tests contribute release
confidence but do not increase `cmd/pe` statement coverage percentages.

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
| `github.com/tmc/pe/internal/llm` | 31.0% |
| `github.com/tmc/pe/ext/starlark` | 38.0% |
| `github.com/tmc/pe/cmd/pe` | 39.0% |
| `github.com/tmc/pe/internal/metaprompt` | 40.8% |
| `github.com/tmc/pe/internal/promptfoo/evaluation/evaluator` | 43.0% |
| `github.com/tmc/pe/internal/module` | 44.7% |
| `github.com/tmc/pe/internal/errors` | 46.0% |
| `github.com/tmc/pe/plugins/promptfoo` | 46.4% |
| `github.com/tmc/pe/internal/inference/providers/cgpt` | 51.3% |
| `github.com/tmc/pe/internal/pemod` | 54.4% |
| `github.com/tmc/pe/internal/providers` | 58.1% |
| `github.com/tmc/pe/internal/prompt` | 58.9% |
| `github.com/tmc/pe/internal/testing/mocks` | 60.4% |
| `github.com/tmc/pe/internal/promptfoo` | 61.5% |
| `github.com/tmc/pe/internal/observability` | 62.0% |
| `github.com/tmc/pe/internal/config` | 66.8% |
| `github.com/tmc/pe/internal/inference/providers/ollama` | 68.0% |
| `github.com/tmc/pe/internal/plugin` | 68.3% |
| `github.com/tmc/pe/internal/cgpt` | 69.3% |
| `github.com/tmc/pe/internal/inference/providers/anthropic` | 73.3% |
| `github.com/tmc/pe/internal/inference/providers/openai` | 74.0% |
| `github.com/tmc/pe/internal/starlark` | 81.5% |
| `github.com/tmc/pe/internal/security` | 81.8% |
| `github.com/tmc/pe/internal/promptfoo/security/redteam` | 86.4% |
| `github.com/tmc/pe/internal/structured` | 89.3% |
| `github.com/tmc/pe/internal/optimization/localopt` | 91.2% |
| `github.com/tmc/pe/internal/templates` | 93.5% |
| `github.com/tmc/pe/internal/distributed` | 94.3% |

## Packages With No Statements

| Package |
| --- |
| `github.com/tmc/pe/example/getting-started` |
| `github.com/tmc/pe/internal/inference/providers` |
| `github.com/tmc/pe/tests` |
