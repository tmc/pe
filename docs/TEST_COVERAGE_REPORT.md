# PE Test Coverage Report

Generated: 2026-05-05

## Baseline

Integrated baseline branch:

```text
exp
```

Integrated commit:

```text
edb843e
```

Commands used for the measured baseline:

```sh
GOTOOLCHAIN=go1.25.9 go test -coverprofile=/tmp/pe-full-coverage.out ./...
go tool cover -func=/tmp/pe-full-coverage.out
```

Results:

| Metric | Value |
| --- | ---: |
| Overall statement coverage | 48.8% |
| Test status | Passing integrated branch |
| Coverage CI | `.github/workflows/ci.yml` runs tests and `make coverage-check` |
| Local coverage target | `make coverage-check` enforces the current floor; `make coverage` writes `coverage.out` and `coverage.html` |

This report is the source of truth for release coverage numbers. Older roadmap
backlog entries with lower percentages are historical.

The 70% target is not met yet. Coverage is improving, but the next work should
focus on command/plugin paths and packages with executable code that currently
show 0% statement coverage.

## Script Test Coverage

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
| `github.com/tmc/pe/internal/promptfoo/evaluation/testing` | 0.0% |
| `github.com/tmc/pe/internal/testing` | 0.0% |
| `github.com/tmc/pe/test` | 0.0% |
| `github.com/tmc/pe/internal/promptfoo/evaluation/metrics` | 27.9% |

## Executable Text Safety Gates

PE is the Go toolchain for safe prompting: executable, templated, composable
text. Plain text is valid by default. Optional declarations can define template
inputs, metadata, safety policy, and placement rules for what data, prompts,
providers, and tools may run where.

The first coverage targets for that direction are test-plan gates, not runtime
expansion:

| Gate | First coverage target |
| --- | --- |
| Executable text parsing | Plain text remains valid; front matter is parsed only when present; unknown required fields fail closed. |
| Template rendering | Declared inputs are required; rendered text is deterministic; missing inputs and undeclared secret values fail validation. |
| Static capability checks | `pe.mod` capability, placement, and policy declarations reject denied data, prompt, provider, and tool classes before execution. |
| Conservative policy composition | Child artifacts and dependencies cannot loosen parent constraints; effective allows narrow and denials accumulate. |

Coverage reports should track these gates as they move from documentation into
parser and static-validation packages. Script tests should cover release-facing
CLI behavior; unit tests should cover parser, renderer, and policy-composition
edge cases directly.

## Other Measured Packages

| Package | Coverage |
| --- | ---: |
| `github.com/tmc/pe/internal/optimization/optimizers` | 30.7% |
| `github.com/tmc/pe/internal/llm` | 31.0% |
| `github.com/tmc/pe/ext/starlark` | 38.0% |
| `github.com/tmc/pe/cmd/pe` | 40.6% |
| `github.com/tmc/pe/internal/metaprompt` | 40.8% |
| `github.com/tmc/pe/internal/optimization/adapters` | 44.6% |
| `github.com/tmc/pe/plugins/promptfoo` | 46.4% |
| `github.com/tmc/pe/internal/errors` | 46.0% |
| `github.com/tmc/pe/internal/pemod` | 54.4% |
| `github.com/tmc/pe/internal/providers` | 57.9% |
| `github.com/tmc/pe/internal/module` | 58.4% |
| `github.com/tmc/pe/internal/prompt` | 58.9% |
| `github.com/tmc/pe/internal/testing/mocks` | 61.1% |
| `github.com/tmc/pe/internal/observability` | 62.0% |
| `github.com/tmc/pe/internal/config` | 66.8% |
| `github.com/tmc/pe/internal/inference/providers/ollama` | 68.0% |
| `github.com/tmc/pe/internal/plugin` | 68.3% |
| `github.com/tmc/pe/internal/cgpt` | 69.3% |
| `github.com/tmc/pe/internal/promptfoo` | 69.2% |
| `github.com/tmc/pe/internal/inference` | 71.2% |
| `github.com/tmc/pe/internal/inference/providers/anthropic` | 73.3% |
| `github.com/tmc/pe/internal/inference/providers/openai` | 74.0% |
| `github.com/tmc/pe/internal/exectext` | 75.0% |
| `github.com/tmc/pe/internal/promptfoo/evaluation/evaluator` | 77.8% |
| `github.com/tmc/pe/internal/starlark` | 81.5% |
| `github.com/tmc/pe/internal/security` | 81.8% |
| `github.com/tmc/pe/internal/promptfoo/security/redteam` | 86.4% |
| `github.com/tmc/pe/internal/structured` | 89.3% |
| `github.com/tmc/pe/internal/optimization/localopt` | 91.2% |
| `github.com/tmc/pe/cmd/pe/commands` | 91.4% |
| `github.com/tmc/pe/internal/templates` | 93.5% |
| `github.com/tmc/pe/internal/distributed` | 94.3% |

## Next Coverage Work

- Add direct tests for `internal/cli` or remove it if it is dead code.
- Add tests for the examples that are expected to stay buildable.
- Add focused unit tests for `internal/optimization` constructors and strategy helpers.
- Add command-level tests for uncovered `plugins/promptfoo` command branches.
- Decide whether `test/run-scripttest.go` should be tested directly or excluded from the coverage target as a helper command.

## Packages With No Statements

| Package |
| --- |
| `github.com/tmc/pe/cmd/pe/commands/core` |
| `github.com/tmc/pe/cmd/pe/commands/evaluation` |
| `github.com/tmc/pe/cmd/pe/commands/experimental` |
| `github.com/tmc/pe/cmd/pe/commands/module` |
| `github.com/tmc/pe/cmd/pe/commands/optimization` |
| `github.com/tmc/pe/cmd/pe/commands/pipeline` |
| `github.com/tmc/pe/cmd/pe/commands/utility` |
| `github.com/tmc/pe/example/getting-started` |
| `github.com/tmc/pe/internal/inference/providers` |
| `github.com/tmc/pe/tests` |
