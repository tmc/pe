# External Execution Policy

This document records PE's current external process boundaries. PE does not use
a general shell to interpret prompt text. External execution is allowed only for
the categories below, with direct argv boundaries or explicit user delegation.

## Inventory Command

```sh
rg -n 'exec\.Command|CommandContext|llmCLICommandContext' cmd internal plugins -g'*.go'
```

Re-run the inventory after adding any runtime process execution path, then
classify the new path here and add focused tests.

## Allowed Runtime Boundaries

| Category | Location | Contract | Current coverage |
| --- | --- | --- | --- |
| Promptfoo viewer delegation | `cmd/pe/view.go` | Only runs through explicit `pe view --promptfoo`; invokes `npx promptfoo view` with direct argv. Default `pe view` stays local. | `TestRunPromptfooViewUsesExplicitArgv`, `TestViewCmd_MissingEvalDoesNotRunPromptfoo` |
| Local browser opener | `cmd/pe/view.go` | Opens the local viewer URL with the platform opener after `pe view --file` or saved-eval lookup. It does not receive prompt text. | Covered by command review and local-viewer tests; keep as platform integration. |
| Generic CLI provider | `internal/providers/cli.go` | Runs trusted local provider configuration. Prefer `executable` plus `args`, optionally with prompt on stdin. Rendered executable names are validated before PATH lookup. Prompt text is passed as argv or stdin, not through shell expansion. | `TestGenericCLIProvider_CommandTemplateQuotesPrompt`, `TestGenericCLIProvider_InvalidExecutableName`, `TestGenericCLIProvider_Generate_ArgvAndStdin`, structured-output tests |
| LLM CLI provider | `internal/providers/llm_cli.go` | Runs the configured `llm` executable with direct argv. Prompt text is one argument. | `TestLLMCLIProvider_GenerateArgs` |
| cgpt providers | `internal/cgpt`, `internal/inference/providers/cgpt` | Runs `cgpt`, `go tool cgpt`, or `go run github.com/tmc/cgpt/cmd/cgpt` as provider execution. Model/backend values are validated in the legacy provider; configured binary paths are rejected if empty or control-character-bearing; stderr is redacted in the inference provider. | `internal/cgpt` package tests, `TestProvider_InvalidBinaryPath`, and `internal/inference/providers/cgpt` package tests |
| Plugin execution | `internal/plugin/plugin.go` | Discovers only executable `pe-*` files under `PE_PLUGIN_PATH`. Discovery does not scan PATH. Plugin execution is explicit and direct argv. | `TestManager_DiscoverIgnoresPATH`, `TestManager_discoverInDir_DoesNotExecuteMetadata`, `TestManager_Execute_PassesArgsWithoutShell`, `TestManager_Execute_ContextCancellation` |
| Metric script hooks | `internal/promptfoo/evaluation/metrics/metrics.go` | Runs explicit executable paths from trusted metric config. Script metrics pass prompt and response as separate argv entries. Python metrics run `python3 -c` from trusted metric config. Output must be valid finite JSON metric data. | `TestMetricEvaluatorScriptMetric`, `TestMetricEvaluatorScriptMetricErrors`, `TestMetricEvaluatorScriptMetricCancellation` |

## Build And Test-Only Boundaries

Test files may build helper binaries or run the local `pe` test binary. These
paths are not product runtime behavior, but they still should use direct argv
and temporary directories.

Current examples include `tests/scripttest_test.go`, `tests/integration`, and
`cmd/pe/exp_cache_test.go`.

## Policy

- Do not add implicit promptfoo CLI execution. Promptfoo CLI delegation must be
  opt-in and named in help text.
- Do not pass prompt text through a shell command string.
- Prefer `exec.CommandContext` over `exec.Command` for runtime paths that can
  block or call user-controlled tools.
- Validate executable names before lookup when they can be templated or
  configured.
- Preserve stderr when it is useful for diagnostics, but redact secrets before
  returning provider or subprocess errors that may reach logs.
- Tests for new external execution paths must cover argv separation, missing
  executable or dependency, non-zero exit, and cancellation where applicable.
