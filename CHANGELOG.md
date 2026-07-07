# Changelog

All notable PE changes are tracked here. The project is pre-1.0; minor releases
may still include behavior changes while the CLI and provider interfaces settle.

## v0.5.0 - 2026-07-07

### Added

- `pe version` command reporting `v0.5.0`.
- `exp` command root and experimental command registry tests.
- Local runtime provider materialization for string and object provider specs.
- Native Ollama provider bridge support for direct generation configuration.
- Provider specs with colon-bearing local model tags, such as
  `ollama:llama3.2:3b`.
- MLX local runtime provider presets and comparison examples.
- Promptfoo-compatible provider labels, prompt scoping, delays, and bounded
  concurrency.
- Promptfoo `llm-judge` assertion provider override support.
- Interactive REPL command wiring for `pe interactive`.
- mdBook/GitBook documentation scaffold.
- Script-test documentation for the custom `rsc.io/script/scripttest` runner.
- `pe run-text` for safe executable text rendering.
- `pe mod vet` for static `pe.mod` capability and executable text policy
  validation.
- MIT `LICENSE`, v0.5.0 migration decision, and NOTICE decision documents.

### Changed

- CLI-backed provider JSON parsing is opt-in through explicit configuration.
- Direct `llm.Provider.Generate` options now preserve provider-specific values in
  `ProviderOptions` while keeping portable options typed.
- Local runtime provider docs and examples use the current object provider shape.
- Local Ollama example validation now covers a real daemon with
  `ollama:llama3.2:3b`.
- ROADMAP.md replaces Beads and legacy TODO files as the active planning source.

### Fixed

- Promptfoo config conversion normalization.
- Anthropic provider registry tests are hermetic.
- Compose remains under experimental commands without mutating the parent command.
- `llm` CLI provider argument construction is covered without requiring a live
  `llm` installation.
- The native Ollama provider always sends an explicit `stream` field, so
  non-streaming completions return the full response and token counts instead
  of one truncated chunk.
- `pe eval` resolves `file://` prompt references relative to the config file
  instead of sending the reference string to the provider.
- Release archives package the binary as `pe` (`pe.exe` on Windows); the CLI
  refuses to run under a `pe-*` name because that prefix is reserved for
  plugin dispatch.
- `pe eval` JSON output checks the marshal error where it occurs.

### Security

- Executable-text imports enforce conservative safety composition: an
  imported artifact cannot allow values its parent denies or exceed a parent
  allow list.
- `pe.mod` `tools deny exec` is enforced before plugin discovery/execution,
  `pe view --promptfoo` delegation, and browser opening.

### Validation

- Full test suite: `go test ./...`
- Scripttest suite: `go test ./tests`
- Linux race-enabled CI suite on `exp`: `go test -race ./...`
- Tag-triggered release workflow validated end to end with `v0.5.0-rc.1` and
  `v0.5.0-rc.2`, including packaged-binary smoke tests.
- Additional release checks are tracked in `ROADMAP.md`.
