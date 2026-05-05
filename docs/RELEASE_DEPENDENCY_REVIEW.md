# Release Dependency Review

Last run: 2026-05-05
Branch: `exp`
Toolchain: Go 1.25.9

This review records the dependency surface for the v0.5.0 release candidate.

## Commands

```bash
GOTOOLCHAIN=go1.25.9 go list -deps ./cmd/pe | wc -l
GOTOOLCHAIN=go1.25.9 go list -m all | wc -l
GOTOOLCHAIN=go1.25.9 go list -m all
```

## Results

- `go list -deps ./cmd/pe` reports 228 packages.
- `go list -m all` reports 39 modules.
- The executable-text work uses the standard library and the existing
  `gopkg.in/yaml.v3` dependency; it does not add a module.
- No dependency was removed in this pass.

## Install Script Decision

Do not add an install script for v0.5.0.

The supported installation paths are:

- `go install github.com/tmc/pe/cmd/pe@latest` after a public tag contains
  `cmd/pe`.
- `go install ./cmd/pe` from a checked-out tree.
- GitHub release archives for macOS, Linux, and Windows.

An install script would add another release artifact, trust boundary, and test
matrix. The current release should keep distribution simple and use the Go
module toolchain plus signed GitHub release assets.
