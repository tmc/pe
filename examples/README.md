# PE Examples

This directory contains runnable examples for PE. This index lists only example
directories that exist in the current tree and have files to inspect or run.

For the older, broader demo set, see [`../example/`](../example/). Treat it as
legacy development material unless a demo has been refreshed here; see
[Legacy `example/` Directory](legacy-example-directory.md).

## Basic

- [Simple Prompt](basic/simple-prompt/) - Single prompt file for `pe run`.
- [Template Variables](basic/template-vars/) - Prompt variables using Go
  template syntax such as `{{.text}}`.

## Evaluation

- [Basic Evaluation](evaluation/basic/) - Promptfoo-style math evaluation using
  OpenAI and Anthropic provider specs.
- [Assertions](evaluation/assertions/) - Assertion examples for content, regex,
  length, JSON, latency, cost, similarity, and LLM judging.
- [Local Runtime Benchmarks](benchmarks/) - Local provider benchmark config for
  MLX, Ollama, and llama.cpp.
- [Local Ollama](local-ollama/) - Local Ollama smoke test. Requires an Ollama
  daemon and the configured model.

## APIs and Extensions

- [Inference API](inference/) - Go example using PE's internal inference client.
- [Creating Modules](modules/create/) - Minimal `pe.mod` module creation example.
- [Executable Text](executable-text/) - Plain, templated, and composable text
  sketches with conservative placement policy.
- [Starlark](starlark/) - Starlark and YAML/Starlark evaluation examples.

## Pipeline

- [Unix Pipelines](pipeline/unix/) - Pipeline composition notes for `pe eval`,
  `pe filter`, `pe stats`, and standard Unix tools.

## Current Commands

- [Current Commands](current-commands/) - Offline smoke examples for `ask`,
  `template`, `prompt`, `plugin`, `profile`, `build`, `convert`, `collect`,
  `reduce`, and `watch`.
- [Distributed Local](current-commands/distributed-local/) - Provider-free
  `pe exp distributed` local scheduler smoke.
- [Consensus Local](current-commands/consensus-local/) - Provider-free
  `pe exp consensus` weighted vote smoke.
- [Eval Regression Gate](current-commands/eval-regression-gate/) - Offline
  `pe diff --fail-on-regression` gate fixture.
- [Release Local Workflows](current-commands/release-local-workflows/) -
  Provider-free smokes for `pe diff`, `pe mod tidy`, `pe exp attest`, and
  `pe exp cache`.

## Validated Commands

These commands were run from the repository root during the examples validation
pass:

```bash
go test ./example/... ./examples/...
PE_TEST_MODE=true go run ./cmd/pe run examples/basic/simple-prompt/prompt.txt --stream=false
PE_TEST_MODE=true go run ./cmd/pe run examples/basic/template-vars/translate.prompt --var text=Test --var source_lang=English --var target_lang=Italian --stream=false
go run ./cmd/pe cat examples/basic/template-vars/translate.prompt --set text=Test --set source_lang=English --set target_lang=Italian
go build -o /tmp/current-commands-pe ./cmd/pe
PE_BIN=/tmp/current-commands-pe ./examples/current-commands/smoke.sh
OPENAI_API_KEY=dummy ANTHROPIC_API_KEY=dummy go run ./cmd/pe eval examples/evaluation/basic/config.yaml --dry-run --no-progress-bar
OPENAI_API_KEY=dummy ANTHROPIC_API_KEY=dummy go run ./cmd/pe eval examples/evaluation/assertions/config.yaml --dry-run --no-progress-bar
```

Notes:

- `pe run` defaults to the `cgpt` provider. Use `PE_TEST_MODE=true` with
  `--stream=false` for offline smoke checks.
- `pe run` accepts template variables with `--var name=value`.
- `pe cat` accepts template variables with `--set name=value`.
- `pe eval --dry-run` still constructs native providers. The OpenAI and
  Anthropic examples therefore require `OPENAI_API_KEY` and `ANTHROPIC_API_KEY`
  to be set, even for dry-run smoke checks.
- Full OpenAI and Anthropic evaluation runs require valid provider credentials
  and network access.
- Local runtime examples require their local runtimes and models: Ollama for
  `ollama:*`, MLX tooling for `mlx-*`, and llama.cpp for `llama.cpp:*`.

## Current Directory Shape

```text
examples/
├── basic/
│   ├── simple-prompt/
│   └── template-vars/
├── benchmarks/
├── evaluation/
│   ├── assertions/
│   └── basic/
├── inference/
├── local-ollama/
├── current-commands/
├── executable-text/
├── modules/
│   └── create/
├── pipeline/
│   └── unix/
└── starlark/
```

Empty placeholder directories are intentionally not listed as examples until they
contain runnable files and documentation.

## Resources

- [PE Documentation](../docs/)
- [Template Syntax Guide](../docs/TEMPLATE_SYNTAX.md)
- [CLI Reference](../docs/CLI_REFERENCE.md)
