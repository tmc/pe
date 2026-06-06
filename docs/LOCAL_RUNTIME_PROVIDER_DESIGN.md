# Local Runtime Provider Design

`pe` now adopts the richer provider-spec ideas already present in the Promptfoo compatibility layer, but it does not route execution through the plugin itself.

## Decision

Reuse `pf` concepts, not `pf` as the backend.

## Why

- The Promptfoo plugin in this repository already models provider configs as richer objects and normalizes schema differences well.
- That plugin is a config translation layer. It does not provide a runtime execution stack for Ollama, `mlx-go-lm`, or `llama.cpp`.
- `pe` already has working execution backends, including a native Ollama HTTP adapter that preserves token counts and timing metadata.
- Reusing the provider-spec shape while keeping execution in `pe` avoids adding another provider abstraction or a second runtime path.

## Resulting Structure

- `internal/promptfoo/types.go`
  - owns the provider spec shape used by config loading
  - supports string or object providers
  - object providers can carry `id`, `label`, `config`, `env`, `prompts`, and `delay`

- `internal/providers/materialize.go`
  - separates provider materialization from execution
  - merges provider-local env into executor options
  - parses execution delay
  - returns executable providers with labels and prompt scoping preserved

- `internal/providers/cli.go`
  - executes argv directly when configured with `executable` and `args`
  - keeps shell-string execution only as a compatibility fallback
  - decodes structured JSON responses for output, token counts, latency, and runtime metadata when `parse_json_response` is enabled

- `internal/inference/providers/ollama/ollama.go`
  - remains the local Ollama execution path
  - preserves `raw`, `seed`, `num_predict`, base URL, and timing/token counters

## Ollama Configuration

Native Ollama execution is registered through the provider bridge. Use
`ollama:model` as the provider spec; the text after `ollama:` is the Ollama
model name sent to the local runtime.

Example provider object:

```yaml
providers:
  - id: "ollama:llama3.2:3b"
    label: "local ollama"
    config:
      base_url: "http://localhost:11434"
      raw: true
      seed: 1
      num_predict: 128
```

`base_url` selects the Ollama daemon URL when `OLLAMA_HOST` is not set. If
omitted, the native provider uses `http://localhost:11434`; `OLLAMA_HOST`
overrides the configured or default URL. `raw` controls Ollama raw prompting.
`seed` and `num_predict` are forwarded as Ollama generation options.

No API key is required. An Ollama daemon must be running, and the configured
model must already be available to that daemon.

The release smoke test for this path lives in
`examples/local-ollama/README.md`. It records the representative local-model
matrix and the focused commands to run against a real Ollama daemon.

## Preset Override Rules

Local runtime presets build `executable` plus `args` defaults for each supported runtime.

- `command` bypasses the preset argv builder and is kept as a compatibility escape hatch.
- `executable` plus `args` replaces the preset argv entirely.
- `executable` by itself changes only the binary name and keeps the preset arguments.
- `args` by itself appends extra arguments after the preset arguments.
- `parse_json_response` is opt-in; without it, valid JSON on stdout is returned as model text.

## Tradeoff

This keeps `pe` and the Promptfoo plugin aligned on provider schema without forcing them to share an execution backend that does not exist yet. The design stays explicit and small: one provider spec model, one materialization step, and existing executors underneath.
