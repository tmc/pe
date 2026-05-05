# Command Taxonomy Review

Date: 2026-05-05
Branch: `exp`

This packet records the current command hierarchy for stakeholder review. It is
not a replacement for review sign-off; `ROADMAP.md` remains open until a
stakeholder approves or requests changes.

## Evidence

- Generated help command: `GOTOOLCHAIN=go1.25.9 go run ./cmd/pe --help`
- Registry implementation: `cmd/pe/commands/registry.go`
- Root registration source: `cmd/pe/root_registry.go`
- Group metadata source: `cmd/pe/root_metadata.go`
- Current taxonomy document: `docs/command-taxonomy.md`
- Full verification gate: `GOTOOLCHAIN=go1.25.9 go test ./... -coverprofile=/tmp/pe-coverage.out -count=1`
- Current coverage result: `70.3%`

## Current Groups

Core:
`ask`, `build`, `doc`, `edit`, `init`, `prompt`, `run`, `run-text`, `serve`,
`test`, `work`

Evaluation:
`benchmark`, `diff`, `eval`, `eval-prompt`, `stats`, `view`

Experimental:
`exp`, `experimental`, `profile`, `security`

Module:
`get`, `mod`, `push`

Optimization:
`evolve`, `fusion`, `gaso`, `optimize`, `pe2`, `semantic`, `textgrad`

Pipeline:
`analyze`, `cat`, `collect`, `compose`, `expand`, `extract`, `filter`,
`reduce`, `stream`

Plugin:
`plugin`

Utility:
`config`, `convert`, `fmt`, `interactive`, `template`, `version`, `vet`,
`watch`

## Review Questions

1. Should `ask` remain in `core`, or move to `pipeline` with the other
   Unix-style text-flow commands?
2. Should `prompt` and `edit` remain in `core`, or move to `module` with prompt
   package management?
3. Should `security` remain in `experimental`, or become a release-facing
   `evaluation` command?
4. Should optimization commands remain top-level, or move under
   `experimental` until their APIs stabilize?
5. Should `completion` and Cobra `help` stay excluded from release-facing
   command inventories?

## Proposed Sign-Off Criteria

- The generated root help and `docs/command-taxonomy.md` list the same
  release-facing root commands, excluding Cobra's built-in `help` and
  `completion`.
- Every registered root command has a group annotation in `cmd/pe/root_metadata.go`.
- Existing command names remain stable; taxonomy changes only affect discovery
  and help grouping unless a separate compatibility plan is approved.
- Any command moved between groups gets a matching documentation update in
  `docs/CLI_REFERENCE.md` and `docs/command-taxonomy.md`.
