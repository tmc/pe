# PE Release Notes - v0.5.0

## Branch: `exp`

Release prep is in progress as of May 5, 2026. These notes describe the current
release candidate state of PE.

For v0.5.0, `exp` is the integrated launchpad branch. Tag from `exp` unless the
maintainer explicitly promotes it to `master` or creates `main`.

## Release-Candidate Surface

### Core Toolchain

- `pe version` reports `v0.5.0` for this build.
- `pe --help` lists the current generated command surface, including prompt
  execution, evaluation, benchmarking, module/workspace commands, templates,
  plugins, security testing, and the experimental command groups.
- Prompt files support variable substitution, metadata, formatting, validation,
  and evaluation hooks.
- The promptfoo-compatible evaluation path supports string assertions, structured
  output checks, provider object configuration, labels, prompt scoping, provider
  delays, and bounded concurrency.

### Providers

- Native OpenAI and Anthropic providers are available through the provider
  factory system.
- Native Ollama support is registered through the local-runtime provider bridge.
- CLI-backed providers support argv-based execution and explicit JSON telemetry
  parsing when configured.
- The `llm` CLI backend is registered as provider `llm`; argument construction is
  covered by tests that do not require a live `llm` installation.
- Direct `llm.Provider.Generate` calls can now carry provider-specific options,
  including Ollama `raw` and `seed` settings.

### Evaluation

- Promptfoo assertion provider overrides are honored for `llm-judge` assertions.
- Unsupported provider overrides on non-judge assertions are rejected explicitly.
- The evaluator preserves provider labels and prompt filters in results.
- Pass@N and structured output examples are present under `example/`.

### Development Workflow

- `pe interactive` now starts the REPL session and accepts provider, config, and
  temperature flags.
- `pe run-text` validates and renders executable text without invoking
  providers, tools, shell commands, or network access.
- `pe mod vet` validates `pe.mod` capability policy and executable text
  metadata.
- Script-based CLI tests use `rsc.io/script/scripttest` and document the runner's
  limitations: no `exec`, shell pipes, redirection, heredocs, or shell job
  control.
- The roadmap in `ROADMAP.md` is the tracked source of truth for planned work;
  Beads is deprecated for this repository.
- v0.5.0 has no required migration steps; see `docs/MIGRATION.md`.

## Notable Changes Since `next`

### Features

- Added `config expand`, `version`, GitBook documentation, and the `exp` command
  root for experimental commands.
- Added local runtime provider materialization for object provider specs.
- Added MLX local runtime presets and local runtime comparison examples.
- Wired the interactive REPL command to the existing REPL session.

### Fixes

- Normalized promptfoo config conversion.
- Preserved provider-specific Generate options through the provider bridge.
- Honored Promptfoo `llm-judge` assertion provider overrides.
- Made CLI JSON response parsing explicit so ordinary JSON model output is not
  rewritten accidentally.

### Documentation and Tests

- Added mdBook/GitBook documentation scaffolding.
- Added the MIT `LICENSE` file claimed by the README.
- Recorded the v0.5.0 migration and NOTICE decisions.
- Corrected command docs for experimental command paths and current CLI surface.
- Documented local runtime provider design and scripttest runner limits.
- Added hermetic tests for provider registration, exp command registration,
  Promptfoo judge provider overrides, REPL command wiring, and `llm` CLI argv
  construction.

## Known Limitations

- Module registry support exists but still needs release validation and clearer
  user documentation.
- Advanced assertion types that require external ML services remain partial.
- Local runtime examples require the corresponding local daemon or binary
  (`ollama`, `mlx-lm`, `mlx-go-lm`, or configured CLI command).
- Documentation still contains older archived and future-facing material.
  `ROADMAP.md` tracks the remaining consolidation work.

## Validation

Current release validation should include:

- `GOTOOLCHAIN=go1.25.9 go test ./...`
- `GOTOOLCHAIN=go1.25.9 go test -cover ./...`
- `GOTOOLCHAIN=go1.25.9 go vet ./...`
- `GOTOOLCHAIN=go1.25.9 govulncheck ./...`
- `go install .`
- Example validation for `example/` and `examples/`

See `docs/TEST_COVERAGE_REPORT.md` for the current measured coverage baseline.

## Support

Report issues at: https://github.com/tmc/pe/issues
