# Changelog

All notable PE changes are tracked here. The project is pre-1.0; minor releases
may still include behavior changes while the CLI and provider interfaces settle.

## v0.5.0 - Unreleased

### Added

- `pe version` command reporting `v0.5.0`.
- `exp` command root and experimental command registry tests.
- Local runtime provider materialization for string and object provider specs.
- Native Ollama provider bridge support for direct generation configuration.
- MLX local runtime provider presets and comparison examples.
- Promptfoo-compatible provider labels, prompt scoping, delays, and bounded
  concurrency.
- Promptfoo `llm-judge` assertion provider override support.
- Interactive REPL command wiring for `pe interactive`.
- mdBook/GitBook documentation scaffold.
- Script-test documentation for the custom `rsc.io/script/scripttest` runner.

### Changed

- CLI-backed provider JSON parsing is opt-in through explicit configuration.
- Direct `llm.Provider.Generate` options now preserve provider-specific values in
  `ProviderOptions` while keeping portable options typed.
- Local runtime provider docs and examples use the current object provider shape.
- ROADMAP.md replaces Beads and legacy TODO files as the active planning source.

### Fixed

- Promptfoo config conversion normalization.
- Anthropic provider registry tests are hermetic.
- Compose remains under experimental commands without mutating the parent command.
- `llm` CLI provider argument construction is covered without requiring a live
  `llm` installation.

### Validation

- Full test suite: `go test ./...`
- Scripttest suite: `go test ./tests`
- Additional release checks are tracked in `ROADMAP.md`.
