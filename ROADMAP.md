# PE Roadmap

Last updated: 2026-06-09

This file is the source of truth for planned PE work. Beads is deprecated for this repository: do not create or update `.beads` issues for new work. Keep roadmap changes in tracked commits with the code or documentation they describe.

## Document Ownership

- `ROADMAP.md`: remaining work, release blockers, priorities, and planning
  status.
- `CHANGELOG.md`: concise release deltas only.
- `RELEASE_NOTES.md`: v0.5.0 release-candidate narrative, known limitations,
  and validation checklist.
- `docs/future/PLANNED_COMMANDS.md`: aspirational command ideas, not status or
  command counts.
- `docs/archive/IMPLEMENTATION_TODOS.md`: tombstone pointing here.
- `docs/future/`: aspirational design material only.

## Priority Guide

- **P1**: Release blocking or user-facing quality work.
- **P2**: Important product, design, or maintenance work.
- **P3**: Backlog features and cleanup.

## Active Work

### P1

#### Release prep: Documentation accuracy audit

- Type: `epic`

**Scope**

Comprehensive documentation accuracy audit and update for release.

Current Issues:
- Test coverage baseline is now measured at 44.4%; remaining stale coverage
  claims should point to docs/TEST_COVERAGE_REPORT.md
- Documentation dates in release-facing status docs have been normalized; older
  research dates, model IDs, examples, future docs, and archive docs are not
  release status dates
- Multiple overlapping getting started docs
- Future docs mixed with current implementation docs
- Legacy TODO material now points at this roadmap, but current documentation still needs an accuracy pass
- Recent scripttest, provider, release-note, CLI-help-audit, and current-status
  work is documented; remaining getting-started docs still need a pass
- `docs/CLI_HELP_AUDIT.md` records the generated root command inventory from
  `GOTOOLCHAIN=go1.25.9 go run ./cmd/pe --help`; `docs/CLI_REFERENCE.md` now
  includes every root command from that inventory.

Remaining sub-tasks:
1. DONE: Replace stale coverage claims with references to docs/TEST_COVERAGE_REPORT.md
2. DONE: Update all documentation dates
3. DONE: Consolidate getting started documentation
4. DONE: Update docs/CURRENT_STATUS.md with latest
5. DONE: Review all command documentation against generated CLI help
6. DONE: Reconcile docs/README.md with README.md and current release notes

Priority: P1 - Blocking release


#### Release prep: Examples validation

- Type: `epic`

**Scope**

Validate and update all examples for release.

Ensure all examples work with current codebase:

Areas to validate:
- example/ directory examples (Go packages pass; legacy README/script/config
  samples still contain future command surfaces and should not be treated as
  release-facing)
- examples/ directory examples (Go packages pass; examples/README.md and
  current README pages refreshed)
- Code examples in documentation (README, getting started, command reference,
  command examples, and gitbook CLI checked against generated help)
- README code samples (current quick-start and optimization samples refreshed)
- Tutorial examples (getting-started guide refreshed; long tutorial still needs
  a deeper rewrite or archive pass)
- Command reference examples (high-traffic sections refreshed against generated
  help for run, eval, view, benchmark, stream, filter, analyze, stats, fmt, vet,
  init, security, and experimental commands)

Evidence:
- `go test ./cmd/pe ./internal/prompt ./internal/promptfoo ./internal/structured ./internal/providers ./ext/starlark` passed during read-only validation.
- `go test ./example/... ./examples/...` passed during read-only validation.
- `PE_TEST_MODE=true pe run ...` and `pe cat ... --set ...` smoke commands from
  examples/README.md passed during read-only validation.
- Generated help was checked for root, run, eval, view, benchmark, fmt, vet,
  init, security test, pipeline commands, and experimental optimize/semantic.
- Remaining legacy risk is concentrated in `example/` promptfoo/security/
  attestation demos and advanced long-form tutorial configs that describe
  future or incompatible config shapes.
- `GOCACHE=$(mktemp -d /tmp/pe-examples-gocache.XXXXXX) go build -o /tmp/pe ./cmd/pe` passed.
- `PE_BIN=/tmp/pe ./examples/current-commands/smoke.sh` passed for `ask`,
  `template`, `prompt`, `plugin`, `profile`, `build`, `convert`, `collect`,
  `reduce`, and `watch --help`.
- `PE_BIN=/tmp/pe ./examples/current-commands/distributed-local/smoke.sh`
  passed.
- `PE_BIN=/tmp/pe ./examples/current-commands/consensus-local/smoke.sh`
  passed.
- `PE_BIN=/tmp/pe ./examples/current-commands/release-local-workflows/smoke.sh`
  passed.

Remaining sub-tasks:
1. DONE: Keep legacy `example/` demos as non-release-facing compatibility and
   design-history material; release-facing examples live under `examples/`
2. DONE: Archive the long-form future-facing tutorial and replace
   `docs/TUTORIAL.md` with a current executable-text tutorial
3. DONE: Test representative live-provider examples with valid credentials;
   `pe run` and `pe ask` returned `PE_LIVE_PROVIDER_OK` with
   `--provider openai:gpt-4o-mini`
4. DONE: Add small runnable examples for current commands with weak coverage: `ask`,
   `template`, `prompt`, `plugin`, `profile`, `build`, `convert`, `reduce`,
   `collect`, and `watch`

Priority: P1 - User-facing quality


#### Release prep: Version and changelog

- Type: `task`

**Scope**

Prepare version number and comprehensive changelog for release.

Current status:
- `pe version` reports v0.5.0.
- `pe version` and `pe --version` both report `pe version v0.5.0 darwin/arm64`
  in the local release-prep workspace.
- RELEASE_NOTES.md and CHANGELOG.md are refreshed for the current release
  candidate.
- `docs/MIGRATION.md` records that v0.5.0 has no required migration steps.

Remaining tasks:
1. DONE: Verify v0.5.0 consistency across all current user-facing docs.
2. DONE: Document the migration decision: either add `docs/MIGRATION.md` /
   `docs/UPGRADING.md` for breaking changes, or record that v0.5.0 has no
   required migration steps.
3. Tag version in git when all P1 release blockers below are closed.


#### Release prep: README consolidation

- Type: `task`

**Scope**

Consolidate and align README files.

Current state:
- README.md points to the generated CLI surface instead of owning a command
  count.
- docs/README.md now points to current status, release notes, changelog,
  coverage, CLI help audit, migration, notice, security, and build-matrix
  artifacts.

Remaining tasks:
1. DONE: Reconcile docs/README.md with README.md
2. DONE: Update badges and links
3. DONE: Link to release notes and changelog from the documentation index
4. DONE: Update contribution guidelines reference

Key sections to update:
- Feature list with accurate status
- Installation instructions
- Quick start examples
- Documentation links
- Getting help section


#### Release prep: Security review

- Type: `task`

**Scope**

Security review before release.

Current status:
- docs/SECURITY_REVIEW.md records local secret scans, govulncheck results,
  gosec findings, and follow-up risks.
- `GOTOOLCHAIN=go1.25.9 govulncheck ./...` reports no vulnerabilities.

Remaining tasks:
1. DONE: Triage remaining gosec G204 command-execution findings
2. DONE: Add or update a concise SECURITY.md disclosure policy
3. DONE: Review file path traversal protections in config expansion, Starlark
   loading, modules, and metadata paths
4. DONE: Review provider stderr/body error propagation for secret redaction
5. DONE: Test representative commands with untrusted input

Tools to use:
- gosec (static analysis)
- go list -m all | nancy (or govulncheck)
- Manual code review of security-sensitive areas

Recent closure:
- GitHub API calls in `cmd/pe` use an explicit 30-second timeout client instead
  of `http.DefaultClient`.
- `TestGitHubHTTPClientHasTimeout` verifies the timeout client.
- Path traversal checks cover config expansion, module registry/cache paths,
  module publish paths, unsigned manifests, and local cache objects.
- `tests/testdata/script/security_untrusted.txt` covers representative CLI
  rejection of untrusted path inputs and shell-looking executable-text data.
- Generic CLI command templates now quote prompt data before shell-style
  splitting so prompt text cannot add argv entries.

Critical areas:
- Command execution (scripttest, providers)
- File operations
- Network requests
- Template evaluation


#### Release prep: Build and distribution

- Type: `task`

**Scope**

Prepare build and distribution infrastructure.

Current status:
- `go install ./cmd/pe` passes locally on darwin/arm64 with Go 1.24.13.
- `pe version` and `pe --version` report `pe version v0.5.0 darwin/arm64`;
  tagged release builds can set version, commit, and build date with
  `-ldflags`.
- Cross-compilation passes with `CGO_ENABLED=0` for darwin/arm64, darwin/amd64,
  linux/amd64, linux/arm64, and windows/amd64. See
  `docs/RELEASE_BUILD_MATRIX.md`.
- `.github/workflows/release.yml` now builds tag-triggered archives for
  darwin/arm64, darwin/amd64, linux/amd64, linux/arm64, and windows/amd64.
- `.github/workflows/release.yml` also supports manual dry-run builds with
  `workflow_dispatch` and `dry_run=true`.
- Binary dependency sanity check: `go list -deps ./cmd/pe` reports 228 packages;
  `go list -m all` reports 39 modules. See
  `docs/RELEASE_DEPENDENCY_REVIEW.md`.

Tasks:
1. DONE: Rerun and record complete cross-compilation results for every release
   target
2. BLOCKED remote: Dry-run the GitHub release workflow from a test tag or
   manual `workflow_dispatch` before publishing v0.5.0. Local `gh workflow run
   release.yml -f dry_run=true --ref exp` failed because `release.yml` is not
   present on the remote default branch (`master`); promote the workflow first
   or run the dry-run after `exp` becomes the release branch.
3. DONE locally: Verify release archive names and install commands against the
   workflow asset names in `docs/RELEASE_BUILD_MATRIX.md`
4. DONE: Do not add an install script for v0.5.0; current docs favor direct
   `go install` and release archives. See
   `docs/RELEASE_DEPENDENCY_REVIEW.md`
5. DONE locally: Verify final binary sizes after release builds complete; see
   `docs/RELEASE_BUILD_MATRIX.md`
6. DONE: Review imported package/module counts for avoidable dependencies; no
   dependency was removed in this pass. See
   `docs/RELEASE_DEPENDENCY_REVIEW.md`

Platforms to test:
- macOS (arm64 passes locally; amd64 passes with `CGO_ENABLED=0`)
- Linux (amd64, arm64)
- Windows (amd64)

Distribution methods:
- go install (primary)
- GitHub releases (binaries)
- Homebrew formula (future)
- Docker image (future)


#### Release prep: Legal and migration readiness

- Type: `task`

**Scope**

Close release-blocking legal, contributor, and migration decisions before
tagging v0.5.0.

Current status:
- `README.md` links to the tracked MIT `LICENSE` file.
- `SECURITY.md` now provides the vulnerability disclosure policy.
- `CONTRIBUTING.md` has been refreshed for current build, test, roadmap, and
  release process guidance.
- v0.5.0 has no required migration steps; `docs/MIGRATION.md` records the
  decision.
- `LICENSE` now contains the MIT license claimed by README.md.
- `docs/MIGRATION.md` records that v0.5.0 has no required migration steps.
- `docs/NOTICE_DECISION.md` records that no separate NOTICE file is required
  for the current v0.5.0 release artifacts.

Release blockers:
1. DONE: Add the missing MIT `LICENSE` file or change the release/license claim.
2. DONE: Decide whether a `NOTICE` file is required for current dependencies and
   embedded/reference material.
3. DONE: Review `CONTRIBUTING.md` for current build, test, roadmap, and release
   process guidance.
4. DONE: Review v0.5.0 changes for breaking API, CLI, provider, config, and
   module behavior; no required migration steps are known for v0.5.0.
5. DONE: Add `docs/MIGRATION.md` or `docs/UPGRADING.md` if migration steps are
   required; otherwise record the no-migration decision in release notes.

Follow-ups after release:
- Automate license inventory if needed.
- Add release checklist automation for migration-guide and legal-file presence.


#### Release prep: Post-integration validation

- Type: `epic`

**Scope**

Validate and document the integrated `exp` branch before tagging v0.5.0. This
section comes from the 2026-05-05 NotebookLM roadmap pass after syncing the
current repository as `pe-current` and the integration reports as `pe-status`.
NotebookLM claims were checked against the filesystem before inclusion here.

Current status:
- `exp` contains the integrated distributed scheduler, stats/diff regression
  gates, module tidy JSON/write behavior, localhost `serve`, experimental
  optimize, and unsigned attest/cache work.
- `cmd/pe/serve.go` sets `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`,
  and `IdleTimeout`; `docs/SECURITY_REVIEW.md` records that current behavior.
- `pe exp attest manifest` and `pe exp attest verify` are wired; future
  attestation work should focus on signatures and trust, not basic manifest
  command wiring.

Release blockers:
1. DONE: Refresh `docs/TEST_COVERAGE_REPORT.md` after the integrated branch:
   `go test -coverprofile=/tmp/pe-coverage.out ./...` then
   `go tool cover -func=/tmp/pe-coverage.out`.
2. DONE: Update command documentation from generated help for `pe diff`, `pe serve`,
   `pe mod tidy`, `pe exp optimize`, `pe exp attest`, and `pe exp cache`.
   Verification: `go run ./cmd/pe <command> --help`.
3. DONE: Reconcile `docs/SECURITY_REVIEW.md` with current `cmd/pe/serve.go`
   timeout behavior and the unsigned/local-only attest/cache caveats.
   Verification: `rg 'ReadTimeout|WriteTimeout|IdleTimeout' cmd/pe/serve.go`,
   `go run ./cmd/pe exp attest --help`, and
   `go run ./cmd/pe exp cache --help`.
4. DONE: Add release-facing examples for the new local workflows:
   `pe diff --fail-on-regression`, `pe mod tidy --json --write`,
   `pe exp attest manifest/verify`, and `pe exp cache manifest put/verify`.
   Verification: `PE_BIN=/tmp/pe ./examples/current-commands/release-local-workflows/smoke.sh`.
5. DONE: Record the final branch policy for v0.5.0: `exp` is the integrated
   launchpad unless the maintainer explicitly promotes it to `master` or
   creates `main`.

Suggested agent lanes:
- `agent/pe-coverage-refresh`: update the coverage report only.
- `agent/pe-docs-cli-sync`: update generated-help-derived command docs only.
- `agent/pe-security-review-sync`: reconcile the security review with current
  serve and attest/cache behavior.
- `agent/pe-new-examples`: add runnable examples and smoke scripts for the new
  local workflows.


### P2

#### Executable, templated, composable text

- Type: `epic`

**Scope**

Make PE the Go toolchain for safe prompting. A prompt, workflow, audit, or recursive long-context run should be a text artifact that can be as simple as a shebang file, then grow into templated, composed, validated, traced, and replayable execution when needed.

This is not a v0.5 release blocker. It is the design umbrella for future workflow, structured-output, optimization, and recursive-context work.

Current direction:
- Plain text is the baseline; a shebang can select a PE runner such as `pe run-text`.
- Text is executable only when PE can lower it into known bounded operators.
- Templates bind inputs; they do not grant execution rights.
- Text artifacts can carry metadata and declare safety policy: what data, prompts, providers, tools, and placement are allowed.
- Composition is explicit through imports, typed inputs, typed outputs, graph edges, and conservative policy merging.
- Recursive long-context execution is bounded over external context, not arbitrary model-generated code.
- Every run emits an auditable JSON trace and declared artifacts.

Initial artifacts:
1. DONE: Maintain `docs/future/EXECUTABLE_TEXT.md` as the primary design note.
2. DONE initial pass: Fold recursive-context design notes under the
   executable-text framing when they land.
3. DONE initial pass: Define `pe.text.v1`, `pe.workflow.v1`, and `pe.trace.v1`
   before broad runtime code.
4. DONE initial pass: Sketch future `pe run-text`, `pe vet-text`,
   `pe exp workflow validate`, `pe exp workflow run`, and `pe exp recurse`
   commands.

Verification:
- `test -f docs/future/EXECUTABLE_TEXT.md`
- `rg "pe run-text|pe.text.v1|pe.workflow.v1|pe.trace.v1|Operator Standard Library" docs/future/EXECUTABLE_TEXT.md`


#### Add pipe support to scripttest framework

- Type: `feature`

**Scope**

The scripttest framework now has a restricted `pipe` command for integration
tests that need Unix-style composition without enabling arbitrary shell
execution.

Impact:
- tests/testdata/script/extract.txt uses `pipe` for extraction workflows
- tests/testdata/script/pipeline.txt uses `pipe` for pipeline workflows
- The restricted implementation supports `pe`, `cat`, and `echo` stages

Options:
1. DONE: Add repository-local restricted pipe support to scripttest
2. DONE where useful: Create intermediate file approach for commands that need files
3. Watch upstream scripttest/testscript only if broader shell semantics are needed

Verification:
- `GOTOOLCHAIN=go1.25.9 go test ./tests -run 'TestPipeStages|TestScripts/(pipeline|extract)' -count=1`

**Notes**

The restricted `pipe` command intentionally does not provide arbitrary shell
semantics. Keep broader shell execution out of script tests unless there is a
specific release need and a reviewed safety model.


#### Maintain roadmap with current status

- Type: `task`

**Scope**

The old TODO file is superseded. Keep this roadmap current as the tracked source of truth.

Need to:
1. DONE current pass: Review completed items and mark appropriately
2. DONE current pass: Add new work items to this roadmap
3. DONE current pass: Update project status
4. DONE current pass: Update references to point at ROADMAP.md
5. DONE current pass: Remove obsolete TODO and Beads references as they are found


#### Document and harden Ollama provider

- Type: `feature`

**Scope**

Native Ollama support exists in internal/inference/providers/ollama and is
registered through the provider bridge. Remaining work is release polish:
documentation, examples, and end-to-end validation against common local models.

Current status:
- Native provider implements Complete, Stream, and Models.
- Provider is registered through internal/providers/init.go.
- Unit tests cover factory, completion, streaming, models, and error paths.
- Direct Generate option passthrough covers provider-specific options such as
  `raw` and `seed`.
- docs/LOCAL_RUNTIME_PROVIDER_DESIGN.md documents Ollama config.
- examples/local-ollama/ contains a local Ollama smoke example.

Tasks:
1. Add end-to-end validation notes for representative Ollama models.
2. Run the local Ollama example against a real daemon before release.

Benefits:
- Local model support (no API keys needed)
- Privacy-focused option
- Cost-effective for development
- Support for open models (llama, mistral, etc.)


#### Improve test coverage

- Type: `task`

**Scope**

Increase test coverage across the project.

Current measured baseline is 44.1% overall statement coverage. The report in
docs/TEST_COVERAGE_REPORT.md lists packages below 30%.

Priority areas:
1. Expand tests in lower-coverage packages, including internal/cli,
   internal/inference, internal/optimization, internal/testing, and example
   command packages.
2. Add command-level integration tests for paths that currently rely on package-level tests only.

Tasks:
1. Add unit tests for low-coverage packages
2. Add integration tests for end-to-end workflows
3. Add property-based tests where appropriate
4. Target 80%+ coverage for critical paths
5. Add benchmark tests for performance-sensitive code

Tools:
- go test -cover
- go test -coverprofile
- Use testify for assertions


#### Harden and document llm CLI provider backend

- Type: `epic`

**Scope**

The base adapter for Simon Willison's llm CLI exists in internal/providers/llm_cli.go
and is registered as the `llm` provider. Remaining work is hardening,
documentation, and examples.
Argument construction is covered by tests that do not require a live `llm`
installation.

Simon's llm tool (https://llm.datasette.io/) is a powerful CLI for interacting with LLMs
that supports multiple providers and models through a plugin system.

Tasks:
1. Support llm model aliases and advanced `-o` configuration.
2. Add integration docs covering setup, plugins, and provider selection.
3. Add examples for common llm model-alias workflows.

Technical Notes:
- llm supports JSON output via --json flag
- Can run: llm prompt 'text' --model model-name --json
- Supports streaming with -s flag
- Has built-in conversation/chat support
- Plugin system extends to new providers

References:
- https://llm.datasette.io/
- https://github.com/simonw/llm
- Similar to cgpt integration but more feature-rich


#### Release prep: Command documentation review

- Type: `task`

**Scope**

Review and update all command documentation for accuracy.

Review these docs:
- docs/CLI_REFERENCE.md
- docs/COMMAND_EXAMPLES_GUIDE.md
- docs/command-taxonomy.md
- docs/CLI_HELP_AUDIT.md

Verify:
1. Current generated command inventory is documented
2. Flags and options are accurate
3. Examples work
4. Help text matches documentation
5. No documented features that don't exist
6. No undocumented commands

Consider:
- Consolidate duplicate command docs
- Generate from --help output
- Add command completion examples


#### Release prep: Documentation organization

- Type: `task`

**Scope**

Reorganize documentation for clarity and discoverability.

Current issues:
- Overlapping content across getting-started, tutorial, and reference docs
- docs/future/ is now marked aspirational, but its files still need curation
- Hard to find relevant information

Proposed structure:
docs/
  README.md (index with clear sections)
  getting-started/
    installation.md
    quick-start.md
    first-steps.md
  guides/
    evaluation.md
    optimization.md
    modules.md
    security.md
  reference/
    commands.md
    api.md
    configuration.md
  architecture/
    overview.md
    providers.md
    testing.md
  future/
    (aspirational docs clearly marked)

Tasks:
1. Consolidate duplicate getting-started and tutorial docs
2. Curate docs/future/ into design notes, archived material, and promotable plans
3. Update internal links after consolidation
4. Keep docs/README.md as the current-documentation index
5. Archive outdated docs


#### Release prep: License and legal review

- Type: `task`

**Scope**

Detailed follow-up tracking for license and legal documentation. Release
blockers are tracked in P1 under "Legal and migration readiness."

Tasks:
1. Check source-file header policy after the release license file is in place
2. Review third-party dependencies and licenses
3. Verify no GPL/AGPL code included
4. Check attribution requirements
5. Update copyright years if needed

Run:
- go-licenses check (if available)
- Review go.mod dependencies
- Check for embedded code


#### Release prep: CI/CD review

- Type: `task`

**Scope**

Review and update CI/CD pipelines for release.

Tasks:
1. Review existing GitHub Actions workflows
2. Ensure tests run on all PRs
3. Add release automation
4. Set up automatic changelog generation
5. Configure semantic versioning
6. Add integration tests to CI
7. Set up code coverage reporting
8. Add linting and formatting checks
9. Configure branch protection rules
10. Document CI/CD process

Checks to add:
- go test ./...
- go vet ./...
- golint or staticcheck
- go mod tidy verification
- Documentation build (if applicable)
- Example validation


#### Release prep: Migration guide

- Type: `task`

**Scope**

Detailed migration-guide follow-up. The release-blocking decision is tracked in
P1 under "Legal and migration readiness."

Assess:
1. Review commits for breaking changes
2. Document API changes
3. Document CLI changes
4. Document configuration format changes
5. Provide migration examples
6. Create upgrade checklist

Potential breaking changes to document:
- Template syntax ({{.var}} requirement)
- Provider configuration format
- Command flag changes
- Configuration file format changes
- Module system changes

Current artifact:
- `docs/MIGRATION.md` records the v0.5.0 no-required-migration decision.

Follow-up:
- Expand `docs/MIGRATION.md` if the breaking-change review finds user-visible
  API, CLI, provider, config, or module migration steps.


#### Release prep: Performance benchmarks

- Type: `task`

**Scope**

Establish performance benchmarks for release.

Tasks:
1. Run existing Go benchmarks
2. Create baseline performance metrics
3. Document performance characteristics
4. Test with various prompt sizes
5. Test with multiple providers
6. Measure memory usage
7. Profile critical paths
8. Document optimization opportunities

Benchmarks to run:
- go test -bench=. ./...
- pe benchmark commands
- Real-world usage scenarios
- Concurrent execution tests

Document in:
- Performance section in docs
- Benchmark results in repo
- Known limitations


### P3

#### pe.mod capabilities and placement

- Type: `epic`

**Scope**

Define `pe.mod` as the module-level capability and placement contract for safe prompting. This extends the Go-like module boundary so executable text can declare what data, prompts, providers, tools, and execution placements are allowed at module scope.

This is not a v0.5 release blocker. v0.5 work is documentation only.

Current status:
- `internal/pemod` already models module identity, PE version, dependencies, replacements, trust, signing, registry, and security policy.
- `internal/pemod` parses and formats `capability`, `placement`, and `policy`
  blocks.
- `pe mod vet` validates module policy shape, executable-text typed-input
  requirements, executable-text denied provider/tool/data/prompt requests, and
  strict cached dependency policy composition.
- Remaining runtime work: enforce effective policy at provider, tool,
  file-write, network, and placement decision points.

Design direction:
1. Keep `pe.mod` Go-like and module-scoped.
2. Put module-wide `capability`, `placement`, and `policy` blocks in `pe.mod`.
3. Keep prompt-specific inputs, output schemas, shebang runners, and local metadata in per-file front matter.
4. Keep runtime facts in trace artifacts.
5. Compose policies conservatively: intersect allows, union denials, and fail strict dependencies that request denied capabilities.

Initial artifact:
- `docs/future/PEMOD_CAPABILITIES_DESIGN.md`

Version plan:
1. DONE: v0.5: docs only.
2. DONE: v0.6: parser/schema support in `internal/pemod`.
3. DONE current pass: v0.6: static validation through `pe mod vet`, including
   strict cached dependency policy composition.
4. v0.7: runtime enforcement for providers, tools, file writes, network, and placement.

Verification:
- `test -f docs/future/PEMOD_CAPABILITIES_DESIGN.md`
- `rg "capability|placement|policy|Conservative Composition" docs/future/PEMOD_CAPABILITIES_DESIGN.md`
- `go test ./cmd/pe -run TestRunModVet -count=1`


#### Review .gitignore for PE project

- Type: `task`

**Notes**

Review and update .gitignore patterns for PE project.

Current state:
- Recently added test binary patterns (*.test, test_binary)
- Has basic patterns (.claude-history.txtar, .h-*, .DS_Store, /pe, .bash_history, lastsession.txt, *.log)
- `.beads/` is deprecated and ignored; `.pe/` is ignored as local cache/state

Need to:
1. Keep `.beads/` ignored as deprecated local task history
2. Keep `.pe/` ignored as local cache/state
3. Review for other common development artifacts
4. Consider patterns for editor/IDE files
5. Ensure consistency with Go project standards


#### Module registry implementation

- Type: `feature`

**Scope**

Complete the module registry system for sharing and discovering prompts.

Current status:
- Core module management exists.
- `pe mod download`, `pe mod list`, and `pe mod search` have command
  implementations.
- `pe mod publish` is registered but publishing/indexing is not implemented in
  the current registry backends.
- Module registry support still needs release validation. Current docs now cover
  supported registry configuration, failure modes, and tidy/vendor limits.

Tasks:
1. DONE current pass: Validate the implemented module commands against a real
   or fixture registry for `pe mod download` success, missing module failure,
   version mismatch failure, and canonical cache layout.
2. DONE current pass: Document `pe mod tidy` / `pe mod vendor` current behavior
   and add a focused vendor cache-miss validation.
3. Add integrity verification, version upgrade, and dependency graph behavior.
4. DONE current pass: Document supported registry configuration and failure
   modes.
5. Decide whether a hosted registry server belongs in this repository or a
   separate project.

Related:
- internal/pemod/ has module code
- docs/MODULE_REGISTRY.md may have design docs


#### Wire experimental distributed CLI

- Type: `task`

**Scope**

Expose the integrated local deterministic scheduler through a real experimental
CLI command.

Current status:
- `internal/distributed/local.go` implements `RunLocal` and `Majority`.
- The promptfoo evaluator already uses `distributed.RunLocal` internally.
- `cmd/pe/exp_distributed.go` wires `exp distributed` to the local scheduler.

Tasks:
1. Replace the `exp distributed` stub with a small command that reads a local
   task description and runs it through `distributed.RunLocal`.
2. Keep the first command provider-free and deterministic.
3. Add tests for ordering, worker limits, cancellation, and invalid inputs at
   the CLI boundary.
4. Document that this is a local scheduler prototype, not a remote worker
   system.

Verification:
- `rg 'expDistributedCmd|RunLocal' cmd/pe internal/distributed`
- `go test ./cmd/pe ./internal/distributed`


#### Wire consensus evaluation mode

- Type: `task`

**Scope**

Make the integrated deterministic majority vote helper usable from evaluation
workflows.

Current status:
- `internal/distributed/local.go` provides `Majority`.
- Evaluator tests build consensus manually from evaluation results.
- No user-facing `pe eval` or `pe fusion` flag exposes this path.

Tasks:
1. Decide whether consensus belongs on `pe eval` or a separate `pe fusion`
   command.
2. Add the smallest CLI surface that can aggregate multiple provider or prompt
   outputs with deterministic tie handling.
3. Preserve existing eval output schemas or add an explicit versioned field for
   consensus metadata.
4. Add tests for ties, empty votes, weighted votes, and provider error rows.

Verification:
- `rg 'Majority\\(' cmd/pe internal/promptfoo/evaluation/evaluator`
- `go test ./cmd/pe ./internal/promptfoo/evaluation/evaluator`


#### Connect local optimizer to provider-backed semantic refinement

- Type: `spike`

**Scope**

Bridge the deterministic `internal/optimization/localopt` package to the
existing provider/metaprompt machinery without making the local optimizer depend
on provider packages.

Current status:
- `pe exp optimize` can refine local scored variants and promptfoo-style score
  JSON without provider calls.
- `internal/optimization/localopt` is intentionally provider-free.
- Existing metaprompt and provider code can generate candidate text, but the
  adapter boundary is not defined.

Tasks:
1. Design a small adapter interface between provider-backed candidate
   generation and `localopt.Optimizer`.
2. Prototype with a fake provider first; keep live-provider tests opt-in.
3. Preserve deterministic tests for score parsing and local refinement.
4. Document whether this graduates under `pe exp optimize` or a separate
   semantic command.

Verification:
- `rg 'localopt|Semantic|GASO|Provider' cmd/pe internal`
- `go test ./cmd/pe ./internal/optimization/localopt ./internal/metaprompt`


#### Add advanced assertion types

- Type: `feature`

**Scope**

Implement advanced evaluation assertion types beyond basic string matching.

From README - marked as 'In Development':
- toxicity detection
- coherence measurement
- factuality checking
- similarity scoring
- classification

Current:
- Basic assertions work (contains, equals, regex, etc.)
- Pass@N and structured output implemented
- LLM rubric assertions work
- Local deterministic baselines cover token-Jaccard similarity, SQL shape,
  required structure markers, toxicity term matching, and keyword
  classification.

Implementation notes:
- Rich toxicity and classification may require external models/APIs
- Coherence could use perplexity scoring
- Factuality needs knowledge base integration
- Similarity can use embeddings after the local baseline
- SQL and structure assertions need parser/schema-backed variants for richer
  guarantees

Location: internal/promptfoo/evaluation/evaluator/


#### Wire interactive REPL mode

- Type: `feature`

**Scope**

`pe interactive` now starts the existing REPLSession from the CLI entrypoint
and accepts provider, config, and temperature flags. Command wiring has tests
that do not require live provider calls. Remaining work is end-to-end behavior
verification inside the REPL loop.

Tasks:
1. Verify prompt history, context preservation, multiline input, and session save/load through the CLI entrypoint.
2. Verify provider hot-switching and temperature/token controls from the command loop.

Command: pe repl or pe interactive

Benefits:
- Faster iteration during development
- Easy experimentation
- Better user experience for exploratory work
- Teaching and demo tool

Consider:
- Use github.com/chzyer/readline for line editing
- Support .pe_history file
- Allow loading prompts from files
- Provider hot-switching


#### REST API server implementation

- Type: `feature`

**Scope**

Implement REST API server for PE toolkit.

From README - marked as 'In Development'.

Features needed:
1. HTTP server with REST endpoints
2. Authentication/authorization
3. API endpoints for main commands:
   - POST /api/v1/run - Execute prompts
   - POST /api/v1/eval - Evaluate prompts
   - GET /api/v1/modules - List modules
   - POST /api/v1/optimize - Optimize prompts
4. WebSocket support for streaming
5. OpenAPI/Swagger documentation
6. Rate limiting and quotas
7. Metrics and monitoring

Use cases:
- Web UI integration
- Third-party integrations
- Multi-user environments
- Cloud deployments

Consider:
- Use standard library net/http or gin/echo
- JWT for auth
- CORS support
- Health check endpoint


#### Sign attest manifests

- Type: `epic`

**Scope**

Add cryptographic identity and origin checks on top of the unsigned manifest and
cache workflows.

Current status:
- `pe exp attest manifest` and `pe exp attest verify` are implemented for
  deterministic local SHA-256 manifests.
- `pe exp attest keygen`, `pe exp attest sign`, and
  `pe exp attest verify-signed` are implemented for Ed25519 signed manifest
  envelopes while preserving unsigned manifest workflows.
- Help text and docs distinguish unsigned local integrity manifests from signed
  envelopes: signed envelopes bind the manifest to a public key, but do not by
  themselves prove key ownership, freshness, or remote origin.

Tasks:
1. DONE current pass: Define a signed manifest envelope that preserves the existing unsigned
   manifest payload.
2. DONE current pass: Start with Ed25519 from the standard library.
3. DONE current pass: Add explicit key-generation, signing, verification, and
   failure-mode docs.
4. DONE current pass: Keep unsigned manifests available for local integrity
   workflows.

Verification:
- `go run ./cmd/pe exp attest --help`
- `go test ./cmd/pe -run 'TestUnsignedManifest|TestExpAttest|TestSignedManifest'`
- `rg 'unsignedManifestType|signedManifestType|ed25519|signature' cmd/pe`


#### Build remote registry download path

- Type: `epic`

**Scope**

Turn the module registry code into a verified remote download path with clear
integrity and failure semantics.

Current status:
- Module commands and registry code exist.
- `pe mod tidy --json --write` is implemented for local dependency hygiene.
- Remote registry behavior has initial fixture-backed validation; live registry
  validation is still pending.

Tasks:
1. DONE current pass: Add a fixture-backed remote registry test path before
   relying on live services.
2. DONE current pass: Verify `pe mod download` extracts exactly the expected
   files and rejects path escapes.
3. Define integrity checks for downloaded modules.
4. DONE current pass: Document supported registry configuration and failure
   modes.

Verification:
- `rg 'GitHubRegistry|download|registry' internal cmd/pe docs`
- `go test ./cmd/pe ./internal/module ./internal/pemod`


#### Add DAG scheduling for optimization graphs

- Type: `spike`

**Scope**

Extend the local scheduler beyond flat task slices so it can execute dependency
graphs produced by semantic optimization work.

Current status:
- `distributed.RunLocal` executes independent tasks with bounded concurrency.
- GASO and semantic optimization code model graph-like dependencies elsewhere.

Tasks:
1. Define a minimal DAG task type without replacing the existing flat
   `RunLocal` API.
2. Add topological scheduling with cycle detection.
3. Preserve deterministic result ordering and cancellation behavior.
4. Test dependency blocking, failed parent behavior, and cycle errors.

Verification:
- `rg 'RunLocal|Dependencies|GASOComputationalGraph' internal`
- `go test ./internal/distributed ./internal/metaprompt`


#### Visualize semantic optimization graphs

- Type: `spike`

**Scope**

Provide a local visualization path for semantic/GASO optimization graphs.

Current status:
- Semantic optimization can produce graph-like structures.
- `pe serve` now has a localhost-first API foundation.
- No current UI renders optimization graph structure for debugging.

Tasks:
1. Define a stable JSON representation for optimization graph nodes and edges.
2. Add a small local-only HTML view or `pe serve` endpoint for inspection.
3. Keep the first version static and provider-free.
4. Add sample graph fixtures and snapshot-style tests.

Verification:
- `rg 'GASOComputationalGraph|serve|render' cmd/pe internal`
- `go test ./cmd/pe ./internal/metaprompt`


#### Add examples for new features

- Type: `task`

**Scope**

Create comprehensive examples for newer PE features.

Current state:
- example/ directory has many examples
- examples/ directory also exists
- May need to consolidate or organize better

Need examples for:
1. Semantic backpropagation (`pe experimental semantic backprop`)
2. GASO optimization (`pe experimental semantic gaso`)
3. Pass@N evaluation
4. Structured output validation
5. Security testing (pe security)
6. Distributed execution
7. Cryptographic attestation
8. Starlark extensions
9. Advanced metrics (BERTScore, G-Eval)
10. Prompt composition

These are additive examples for future polish. Release-blocking validation of
existing examples is tracked under P1.

Organization:
- Each example should be self-contained
- Include README with explanation
- Show both simple and advanced usage
- Include expected output
- Add to documentation


## Future Implementation Roadmap

This section groups the remaining post-cleanup implementation work into a
sequenced plan. It is intentionally broader than the v0.5 release checklist:
v0.5 should ship only after the P1 release gates are closed; later milestones
can graduate selected experimental and aspirational surfaces into stable APIs.

The cleanup branch for this work is `docs/archive-old-docs-cleanup`. Use that
branch to archive old docs, tombstone stale implementation notes, and keep
release-facing documentation honest before promoting changes back to the
release branch. Do not use `.beads` for tracking this work.

Each implementation slice should follow the same loop:

1. Sync the current checkout to the NotebookLM cleanup notebook.
2. Run a focused `generate-chat` review for either cleanup or visionary scope.
3. Verify every notebook claim against the filesystem before changing code.
4. Implement one narrow slice with fail-closed behavior for unfinished paths.
5. Run focused tests, then the full suite when the slice changes shared
   behavior.
6. Sync again and ask for a post-change review before staging.

Cleanup sessions should prefer removing stale claims, fake success, and
unreviewed execution paths. Visionary sessions should produce design notes or
roadmap entries until the API contract, dependencies, and test strategy are
small enough to implement.

Notebook session types:
- Cleanup `generate-chat` sessions audit the current tree for stale docs,
  phantom command claims, unsafe execution paths, release blockers, and old
  TODO material that should be archived or converted into this roadmap.
- Visionary `generate-chat` sessions explore future surfaces, but their output
  lands first as `docs/future/` notes or roadmap entries. No visionary session
  should create a stable command claim until a local implementation and tests
  exist.
- Post-change sessions review the exact diff from the current branch. Treat
  notebook feedback as reviewer input, not authority; confirm paths, symbols,
  command names, and test claims locally before acting.

### Milestone 0: v0.5 Release Closure

Goal: ship the current stable core without claiming unfinished behavior.

1. Run the remote release workflow dry-run after `.github/workflows/release.yml`
   is present on the remote release/default branch, or after the release branch
   is promoted.
2. Fix remaining release-facing documentation drift, including README feature
   claims about OWASP/security completeness and any roadmap entries that still
   imply module publishing is implemented.
   DONE current pass: root README, `README.md.old`, and the archived
   `docs-legacy/README.md` now describe security checks as OWASP-oriented with
   incomplete analyses reported explicitly or failed closed, rather than
   complete OWASP coverage.
3. Validate the remote registry read path with fixtures: `pe mod download`,
   `pe mod list`, and `pe mod search` should have deterministic success and
   failure tests that do not depend on live services.
   DONE current pass: `internal/module` has `httptest` coverage for HTTP
   list/get/search/download, token use, error statuses, and traversal
   rejection; `cmd/pe` now verifies `pe mod download` fails closed for missing
   modules and version mismatches, and caches successful downloads under the
   canonical `.pe/cache/modules/<module>@<version>` layout used by `vendor` and
   `verify`.
4. Run the Ollama example against a real local daemon and record model/version
   notes, error modes, and privacy caveats.
5. Archive or tombstone old documentation on the cleanup branch before release:
   release-facing docs stay current, `docs/archive/` keeps historical material,
   and `docs/future/` keeps aspirational material with clear headers.
   Archive targets include legacy root files such as `README.md.old`,
   historical Claude-era notes, old status matrices, and future-facing
   comparison/tutorial material that still reads like current product behavior.
   When a document remains in place for compatibility, add an explicit
   historical header instead of silently rewriting it into current status.
6. Keep promptfoo shell-out behavior explicitly opt-in; do not add implicit
   CLI execution paths without a reviewed command contract and tests.

Verification:
- `GOTOOLCHAIN=go1.25.11 go test ./... -count=1`
- `PE_BIN=/tmp/pe ./examples/current-commands/smoke.sh`
- `PE_BIN=/tmp/pe ./examples/current-commands/release-local-workflows/smoke.sh`
- GitHub Actions release dry-run evidence linked from release notes or build
  matrix docs.

### Milestone 1: Accuracy and Explicitness

Goal: every registered command either works, fails closed, or returns a precise
not-implemented error with matching documentation.

1. Audit all explicit not-implemented paths and classify them as stable,
   experimental, future, or removable.
2. Keep `docs/CLI_REFERENCE.md`, generated help, README, `docs/CURRENT_STATUS.md`,
   and `ROADMAP.md` synchronized for each classified command.
3. Add tests for not-implemented command surfaces so future changes cannot
   regress into fake success.
   DONE current pass: add fail-closed regression guards for unsupported
   promptfoo assertions, TypeScript/Pydantic parser and validator adapters, and
   playground not-implemented endpoints.
   DONE current pass: replace canned-success semantic flow, gradient
   visualization, drift monitor, dependency analysis, and benchmark outputs with
   precise not-implemented errors after input validation.
4. DONE current pass: remove dormant `pe exp` and `pe test` placeholders from
   the registered command surface until implementation work starts.

Current known explicit gaps:
- cross-validation summaries and richer command-level statistical analysis
  surfaces beyond the current `pe diff --statistical` and
  `pe metrics --statistical` paths
- richer prompted template-library creation beyond current starter-template API
- provider-backed/metamodel metaprompt synthesis beyond current local template,
  evolutionary, and neural-style strategies
- provider-assisted compose optimization, remote component import, and richer
  coherence validation beyond current local compose checks
- richer playground compare, security, components, history, and BERTScore
  behavior beyond current local JSON endpoints
- TypeScript and Pydantic source/runtime validation beyond local JSON-object
  data validation
- advanced/provider-backed promptfoo assertions for toxicity, coherence,
  factuality, and classification; provider-backed semantic similarity, SQL
  parsing, and rich structure validation remain future work beyond current
  local baselines
- richer data-poisoning and supply-chain security analyses beyond current
  local response-pattern checks

### Milestone 2: Build, Test, Metrics, and Structured Validation

Goal: finish the high-visibility command paths that users naturally expect from
the current command surface.

1. DONE current pass: Implement `pe build --validate` as a local validation
   pass over prompt/config syntax, front matter, and declared inputs.
2. DONE current pass: Add provider formatting only where PE can prove the
   format locally; unsupported providers return explicit errors before
   multi-target builds write partial artifacts.
   DONE current pass: add deterministic local Google build formatting alongside
   OpenAI and Anthropic, while unknown providers still fail before multi-target
   builds write partial artifacts.
3. DONE current pass: Implement local statistical primitives in the metrics
   package: summary confidence intervals, Welch/paired t-test, two-proportion
   A/B test, Mann-Whitney U, Kolmogorov-Smirnov, group comparison, effect-size
   reporting, and edge-case errors. Remaining work: command-level exposure,
   cross-validation summaries, and multiple-comparison caveats.
3a. DONE current pass: Expose an informational `pe diff --statistical` overlay
    for pass-rate, score, and latency significance on saved evaluation results.
    The output includes a multiple-comparison caveat and does not alter diff
    gate behavior.
3b. DONE current pass: Wire `pe metrics --statistical` to the local statistical
    analyzer for strict numeric group comparison from text or JSON-array
    inputs, with confidence intervals, p-values, effect sizes, finite structured
    output, and the multiple-comparison caveat.
3c. DONE current pass: Implement YAML output for semantic optimization and
    GASO results using local marshaling, with script coverage for generated
    GASO YAML artifacts.
4. Implement automatic test generation only after the input/output contract is
   narrow enough to test deterministically with a mock provider.
5. Implement TypeScript and Pydantic validation through isolated adapters with
   clear dependency and execution boundaries.
   DONE current pass: TypeScript and Pydantic formatter plugins now validate
   JSON object data locally against PE schemas without shelling out to
   TypeScript or Python runtimes. Parsing TypeScript/Python source remains out
   of scope until an isolated runtime adapter is designed.
6. Add examples and script tests for each promoted command path.
7. Refresh the roadmap and release-facing docs whenever a gap is closed so this
   section does not keep stale blockers.

Verification:
- Unit tests for each statistical primitive with edge cases for empty, tiny,
  non-finite, zero-variance, and tied samples.
- Script tests for `pe build --validate`, future statistical commands, and
  structured validation success/failure cases.
- No command shells out to user-provided tools unless the CLI contract requires
  it, the user opts in explicitly, and the path is covered by tests.

### Milestone 2a: Promptfoo CLI and External Execution Policy

Goal: make every external process boundary deliberate, documented, and tested.

Default stance: PE should implement promptfoo-compatible behavior locally in Go.
Shelling out to the promptfoo CLI is out of scope for normal evaluation,
conversion, assertion, optimization, and reporting paths. The only acceptable
exception is explicit user delegation to the underlying promptfoo CLI for a
feature that is clearly documented as promptfoo-owned, cannot yet be represented
locally, and has a tested direct-argv boundary.

1. Inventory all runtime `exec.Command` and `exec.CommandContext` paths and
   classify them as build/test-only, plugin execution, provider execution,
   platform opener, promptfoo compatibility, metric script hook, or removable.
   DONE current pass: `docs/EXTERNAL_EXECUTION_POLICY.md` records the current
   inventory command, runtime categories, build/test-only category, policy, and
   coverage anchors.
2. Keep direct promptfoo CLI delegation behind explicit user intent. The current
   acceptable shape is an option like `pe view --promptfoo`, where the help text
   names the delegation and tests prove the default path stays local.
   DONE current pass: `pe view --promptfoo` uses a tested direct argv boundary
   through `npx promptfoo view`, and default `pe view <evalId>` stays on the
   local viewer path when `--promptfoo` is absent.
   Future promptfoo delegation, if any, must use the same shape: an explicit
   flag or subcommand name, no shell interpolation, bounded execution where
   possible, clear dependency errors when promptfoo is unavailable, and tests
   proving the non-delegated path does not invoke promptfoo.
3. Decide whether metric script hooks should remain supported. If they remain,
   require explicit configuration, timeouts, argument separation, no shell
   interpolation, and tests for missing executable, timeout, stderr, and
   non-zero exit handling.
   DONE current pass: script metrics remain explicit executable-path hooks,
   run as direct argv with prompt and response as separate arguments, and fail
   closed for empty or missing executables, context cancellation, non-zero
   exits with stderr, malformed JSON output, missing fields, and non-finite
   scores. Python metric output now uses the same validated result parser.
4. Keep provider CLI adapters as provider execution, not promptfoo
   compatibility. They must validate executable names, avoid shell expansion,
   use timeouts, and document that they run local tools.
   DONE current pass: `GenericCLIProvider` validates rendered executable names
   before lookup, keeps prompt data in argv or stdin instead of shell
   expansion, and uses direct executable stubs in structured-output tests.
   DONE current pass: plugin execution runs discovered plugin paths with direct
   argv, preserves shell-looking arguments as data, and respects context
   cancellation.
5. Remove any promptfoo shell-out that is merely a convenience wrapper and can
   be replaced with local Go behavior.
6. Document the policy in release-facing docs only after tests cover every
   allowed external execution path.
7. Keep `docs/EXTERNAL_EXECUTION_POLICY.md`, `docs/PROMPTFOO_INTEGRATION.md`,
   `docs/API_REFERENCE.md`, and generated help synchronized whenever an
   external execution boundary is added, removed, or reclassified.

Verification:
- `rg -n 'exec\\.Command|CommandContext' cmd internal plugins tests`
- Unit tests for each allowed non-test external execution path.
- Script tests proving defaults do not invoke promptfoo or arbitrary shells.
- Manual smoke only for explicitly delegated promptfoo CLI behavior, because it
  depends on `npx` and the local promptfoo install.

Exit criteria:
- No implicit promptfoo CLI invocation remains.
- All promptfoo-compatible import/export/convert/eval/assertion behavior either
  runs locally or fails closed with a precise unsupported-feature error.
- Explicit promptfoo CLI delegation remains available only where it is named in
  the command contract and covered by tests.

### Milestone 3: Module Registry and Supply Chain

Goal: make modules useful beyond local examples while preserving clear trust and
integrity semantics.

1. Finish registry indexing and publishing, or explicitly split publishing into
   a separate service/repository if PE should remain client-only.
   DONE current pass: gist-backed `pemod` publishing now updates the root
   `index.json` with version metadata, file lists, size, checksum, path-cleaning
   checks, and cache refresh after creating the module gist. Remaining work:
   archive extraction guarantees, richer registry fixtures, dependency graph
   behavior, and the broader publishing-service boundary decision.
2. Define module archive format, path-cleaning rules, checksum recording, and
   extraction guarantees.
   DONE current pass: local registry publishing records deterministic module
   directory checksums, local/HTTP/GitHub downloads verify declared checksums
   after extraction, and the CLI verifier uses the same checksum primitive.
   DONE current pass: module metadata now supports `archive` as a txtar module
   archive, HTTP/GitHub downloads extract txtar archives with path-containment
   checks, empty/malformed archive rejection, 16 MiB archive limits, and
   post-extraction checksum verification. `docs/module-registry.md` records the
   current format. Remaining work: richer archive publishing ergonomics.
3. Add fixture-backed HTTP and GitHub registry tests for list, search, download,
   missing version, malformed archive, checksum mismatch, and path traversal.
   DONE current pass: HTTP and local registry tests now cover checksum mismatch
   rejection, while existing fixture tests cover HTTP list/get/search/download
   and traversal rejection. Remaining work: malformed archive and richer GitHub
   fixture coverage.
   DONE current pass: HTTP registry fixtures now cover txtar archive download,
   empty/malformed archive rejection, and archive path traversal rejection.
   Remaining work: richer GitHub fixture coverage.
4. Implement version upgrade and dependency graph behavior with predictable
   conflict reporting.
   DONE current pass: module version conflict resolution now chooses the
   highest declared version that satisfies all semver constraints and reports a
   clear conflict when requirements cannot be satisfied, instead of taking a
   lexicographic latest guess.
5. Connect `pe.mod` capability, placement, and policy blocks to static
   validation, then runtime enforcement for provider/tool/file/network access.
   DONE current pass: `pe mod vet` now checks strict cached dependency
   `pe.mod` files so dependencies cannot request provider/tool/data/prompt
   classes or network placement denied by the parent module. Remaining work:
   runtime enforcement at provider/tool/file/network decision points.
   DONE current pass: `pe run` now enforces `pe.mod` provider denials for exact
   provider names and known provider classes such as `remote` before provider
   construction. Remaining work: runtime enforcement for tool, file-write,
   network, and broader placement decision points.
   DONE current pass: `pe run` now enforces `placement { network false }` for
   known remote providers before provider construction. Remaining work:
   runtime enforcement for tool and file-write decision points, plus broader
   placement beyond provider/network classification.
   DONE current pass: component imports now enforce `tools deny write` before
   writing files under `components/`. Remaining work: runtime enforcement for
   other tool classes and broader file-write surfaces.
   DONE current pass: `pe build` now enforces `tools deny write` before build
   artifact writes while keeping `pe build --validate` read-only. Remaining
   work: runtime enforcement for other tool classes and remaining write
   surfaces.
   DONE current pass: `pe eval -o` and `pe eval --save-db` now enforce
   `tools deny write` before writing result files or promptfoo eval storage,
   while default stdout evaluation remains read-only. Remaining work: runtime
   enforcement for other tool classes and remaining write surfaces.
   DONE current pass: `pe benchmark -o` now enforces `tools deny write` before
   writing result files, while stdout and Go benchmark output remain read-only.
   Remaining work: runtime enforcement for other tool classes and remaining
   write surfaces.
   DONE current pass: `pe metrics -o` now enforces `tools deny write` before
   writing result files, while stdout metrics output remains read-only.
   Remaining work: runtime enforcement for other tool classes and remaining
   write surfaces.
   DONE current pass: `pe expand --output` now enforces `tools deny write`
   before writing expanded JSON files, while stdout expanded output remains
   read-only. Remaining work: runtime enforcement for other tool classes and
   remaining write surfaces.
6. Sign attest manifests with an Ed25519 envelope while preserving unsigned
   local integrity workflows.
   DONE current pass: `pe exp attest keygen`, `sign`, and `verify-signed`
   produce and verify Ed25519 signed manifest envelopes that preserve the
   unsigned manifest payload; tests cover wrong key, tampered payload, missing
   signature, changed local files, and CLI round trips.

Verification:
- `go test ./cmd/pe ./internal/module ./internal/pemod`
- Fixture archives prove exact file extraction and rejection of path escapes.
- Signature tests cover wrong key, tampered payload, missing signature, and
  unsigned compatibility modes.

### Milestone 4: Evaluation Quality and Safety

Goal: graduate advanced assertions and security analyses only when they have
honest scoring semantics and failure modes.

1. Implement similarity scoring with a local deterministic baseline first, then
   add optional embedding/provider-backed variants.
   DONE current pass: add local token-Jaccard similarity assertions with
   threshold support and explicit method metadata.
2. Implement coherence and factuality checks behind explicit data/model
   dependencies; avoid presenting heuristic scores as ground truth.
   DONE current pass: add local coherence scoring based on transitions and
   repetition, plus local factuality checks against required fact strings from
   `value` or `config.facts`. Provider-backed scoring and external grounding
   remain future work.
3. Implement toxicity/classification/SQL/structure assertions with clear
   provider requirements or local validators.
   DONE current pass: add local SQL shape assertions for recognized starting
   keywords, null bytes, balanced parentheses, and balanced quotes; add local
   structure assertions that check required output markers from `value` or
   `config.required`.
   DONE current pass: add local toxicity term matching and keyword
   classification assertions with explicit method metadata.
4. Replace fail-closed data-poisoning and supply-chain placeholders with
   concrete local checks, then optional remote or model-assisted analysis.
   DONE current pass: add local response-pattern checks for data-poisoning and
   supply-chain indicators with custom pattern support and safe no-evidence
   results.
5. Add benchmark fixtures and calibration docs for false positives, false
   negatives, and unsupported environments.

Verification:
- Assertion tests include supported, unsupported, and dependency-missing paths.
- Docs explain what each metric proves and what it does not prove.

### Milestone 5: Composition, Synthesis, and Optimization

Goal: make optimization and composition more than isolated experiments while
keeping provider dependencies behind small interfaces.

1. Define the adapter boundary between provider-backed candidate generation and
   deterministic local optimization.
2. Implement template/evolutionary/neural synthesis strategies only after their
   inputs, outputs, scoring, and trace format are specified.
   DONE current pass: add deterministic local template, evolutionary, and
   neural-style synthesis strategies that generate prompt programs from
   `ProgramSpec`, return quality/confidence metadata, preserve strategy
   selection, and honor context cancellation. Remaining work: provider-backed
   synthesis with calibrated scoring.
   DONE current pass: `pe synthesize --output-format yaml` now emits real YAML
   for synthesis results instead of a placeholder message, with explicit errors
   for unsupported synthesis output formats.
3. Implement compose optimization, component import, and coherence validation
   with local-first checks and provider-assisted options.
   DONE current pass: `pe compose --coherence-check` now reads component files,
   rejects missing or empty components, and reports deterministic local
   semantic-overlap and style-consistency scores. Remaining work: provider-
   assisted coherence validation, optimization, and remote component import.
   DONE current pass: `pe compose --coherence` now validates the composed
   prompt locally with the same semantic-overlap and style-consistency scoring,
   including empty-prompt and context-cancellation errors. Remaining work:
   provider-assisted coherence validation, optimization, and remote component
   import.
   DONE current pass: `pe compose --import` now imports local component files,
   directories, and txtar archives into `components/`, with path-cleaning
   checks that reject archive path escapes. Remaining work: remote component
   import and richer component-library workflows.
   DONE current pass: `pe compose --import` now imports remote HTTP(S) txtar
   bundles and single component files with a bounded response size, direct GET,
   and the same path-cleaning checks as local imports. Remaining work: richer
   component-library workflows and provider-assisted composition.
   DONE current pass: `pe compose --optimize` now applies deterministic local
   composition polish without provider calls, records optimization metadata, and
   avoids claiming TextGrad/provider execution. Remaining work: provider-assisted
   optimization and richer scoring.
4. Add DAG scheduling for semantic optimization graphs with deterministic
   topological execution, cancellation, and cycle errors.
   DONE current pass: add the local generic DAG scheduler primitive with
   upfront cycle and graph validation, bounded worker execution, input-order
   results, and dependency-failure skips. Remaining work: wire it into concrete
   semantic optimization graph execution.
   DONE current pass: wire GASO component gradient application through the
   local DAG scheduler so declared component dependencies control optimization
   order, dependency cycles fail before provider calls, and failed parents
   prevent child optimization.
5. Add static visualization for semantic/GASO graph structures through a
   local-only endpoint or generated HTML artifact.
   DONE current pass: `pe experimental semantic flow` now performs local graph
   analysis over system components and dependencies, emits JSON/YAML/text, and
   can render a static HTML/SVG graph artifact without provider calls.
   DONE current pass: `pe experimental semantic analyze --dependencies` now
   emits local JSON/YAML/text dependency reports with source/sink/isolation,
   centrality, warnings, and cycle detection.
   DONE current pass: `pe experimental semantic gradients` now emits local
   prompt-improvement gradient signals and optional static HTML visualization
   without provider calls.
   DONE current pass: `pe experimental semantic monitor` now emits local
   semantic-drift reports with token overlap, added/removed terms, severity,
   and JSON/YAML/text output.
   DONE current pass: `pe experimental semantic benchmark` now emits local
   prompt-readiness benchmark reports across named baselines with deterministic
   dimension scores and JSON/YAML/text output.

Verification:
- Fake-provider tests for candidate generation.
- Deterministic optimizer tests that do not call live providers.
- Graph tests for ordering, failed parents, cancellation, and cycles.

### Milestone 6: Local Workflow and Developer Experience

Goal: improve day-to-day workflows without widening the trusted execution
surface accidentally.

1. Verify and harden the interactive REPL: history, multiline input, context
   preservation, save/load, provider switching, and temperature/token controls.
2. Finish non-interactive template creation and template-library interactive
   creation.
   DONE current pass: implement non-interactive `pe template create` for plain
   `.prompt` artifacts and YAML/JSON template definitions, including inferred
   variables from `{{ .name }}` placeholders and fail-closed missing prompt
   validation.
   DONE current pass: implement `TemplateLibrary.CreateTemplate` as a
   deterministic starter-template creator with unique names, validation-ready
   variables, optional library-path persistence, and context cancellation.
3. Harden the playground endpoints or remove/mark unavailable endpoints from
   release-facing docs until implemented.
   DONE current pass: playground compare, security, components, history, and
   BERTScore metric requests now return local JSON results instead of 501
   placeholders. Remaining work: persistent history, richer component-library
   workflows, provider-backed comparison, and calibrated semantic metrics.
4. Expand examples for structured output, pass@N, security testing,
   distributed execution, attestation, Starlark extensions, advanced metrics,
   composition, and local providers.
5. Continue coverage work in low-coverage packages listed in
   `docs/TEST_COVERAGE_REPORT.md`.

Verification:
- REPL behavior tests where possible, with manual transcript fixtures for
  terminal-only behavior.
- Script tests for promoted template and playground paths.
- Coverage report refreshed after each major test push.

### Milestone 7: Service, IDE, and Multimodal Expansion

Goal: add larger product surfaces only after the CLI contracts are stable.

1. Define a REST API around stable command semantics, not internal package
   shapes.
2. Add authentication, authorization, rate limits, health checks, metrics, and
   streaming/WebSocket behavior before documenting multi-user deployment.
3. Design IDE integrations around executable text, traces, diagnostics, and
   module policy validation.
4. Add visual prompt engineering only after there is a stable graph/artifact
   model to render and edit.
5. Add multimodal prompt support with explicit media typing, size limits,
   provider capability detection, and fixture-backed local validation.
6. Keep neurosymbolic synthesis as a research milestone until the deterministic
   synthesis and optimization trace contracts are stable.

Verification:
- API conformance tests for each endpoint.
- Local smoke tests for streaming and cancellation.
- No stable docs claim multimodal, IDE, visual, or neurosymbolic support until
  the implementation and tests land.


## Architecture Implementation Backlog

This checklist was moved from `docs/archive/IMPLEMENTATION_TODOS.md` so roadmap work lives in one tracked file.

### PE Architecture Implementation Todo List

#### Phase 1: Provider Interface Consolidation (Weeks 1-2)

##### Audit & Analysis
- [x] Map all usages of `llm.Provider` in the codebase
- [x] Map all usages of `inference.Provider` in the codebase
- [x] Document which commands use which interface
- [x] Identify all provider implementations (OpenAI, Anthropic, cgpt, etc.)
- [x] Analyze migration complexity for each usage
- [x] Create compatibility matrix for provider features
- [x] Document breaking changes that will occur
- [x] Review test coverage for provider-dependent code

##### Migration Preparation
- [x] Create `internal/inference/migration.go` with LegacyAdapter
- [x] Implement adapter for `llm.Provider` → `inference.Provider`
- [x] Write adapter unit tests
- [x] Create migration helpers for common patterns
- [x] Add temporary compatibility layer
- [x] Document migration patterns for contributors

##### Command Migration
- [x] Update `cmd/pe/run.go` to use `inference.Provider`
- [x] Update `cmd/pe/ask.go` to use new interface
- [x] Update `cmd/pe/eval.go` for evaluation commands
- [x] Update `cmd/pe/optimize.go` and optimization commands
- [x] Update `cmd/pe/semantic.go` for semantic optimization
- [x] Update `cmd/pe/evolve.go` for evolutionary optimization
- [x] Update `cmd/pe/textgrad.go` for textual gradient optimization
- [x] Update `cmd/pe/pe2.go` for PE2 optimization
- [x] Update `cmd/pe/benchmark.go` for benchmarking
- [x] Update `cmd/pe/test.go` for testing commands
- [x] Update `cmd/pe/stream.go` for streaming
- [x] Update `cmd/pe/fusion.go` for multi-model fusion

##### Provider Implementation Updates
- [x] Update OpenAI provider to single interface
- [x] Update Anthropic provider to single interface
- [x] Update cgpt provider wrapper
- [x] Update mock provider for testing
- [x] Remove duplicate provider implementations
- [x] Consolidate provider registration logic
- [x] Update provider factory methods

##### Cleanup & Validation
- [x] Delete `internal/llm/provider.go`
- [x] Remove all legacy provider implementations
- [x] Update all import statements
- [x] Fix compilation errors
- [x] Run full test suite
- [x] Manual testing of critical paths
- [x] Performance regression testing
- [x] Update documentation

#### Phase 2: Module System Implementation (Weeks 3-4)

##### Registry Design
- [x] Research registry implementation options
- [x] Design registry API specification
- [x] Define module metadata format
- [x] Design module versioning scheme
- [x] Create module signature format
- [x] Design dependency resolution algorithm
- [x] Plan caching strategy
- [x] Document registry protocol

##### Registry Implementation
- [x] Create `internal/module/registry.go`
- [x] Implement registry client interface
- [x] Add GitHub-based registry option
- [x] Add HTTP API registry option
- [x] Implement registry authentication
- [x] Add module search functionality
- [x] Implement module metadata fetching
- [x] Add registry health checks

##### Module Resolution
- [x] Create `internal/module/resolver.go`
- [x] Implement module path parsing
- [x] Add version constraint parsing
- [x] Implement semantic version comparison
- [x] Add module cache interface
- [x] Implement file-based cache
- [x] Add cache invalidation logic
- [x] Implement module download functionality

##### Dependency Management
- [x] Create `internal/module/deps.go`
- [x] Implement dependency graph structure
- [x] Add topological sort for dependencies
- [x] Implement conflict detection
- [x] Add version resolution algorithm
- [x] Implement circular dependency detection
- [x] Add dependency pruning
- [x] Create lock file format

##### Module Commands
- [x] Fix `pe mod init` with proper initialization
- [x] Implement `pe mod download` with real registry
- [x] Complete `pe mod tidy` functionality
- [x] Implement `pe mod vendor` properly
- [x] Add `pe mod verify` for integrity checking
- [x] Implement `pe mod list` for installed modules
- [x] Add `pe mod search` for registry search
- [ ] Implement `pe mod publish` registry indexing and module publishing
- [x] Add `pe mod upgrade` for version updates
- [x] Implement `pe mod graph` for dependency visualization

##### Module Security
- [x] Implement module signing
- [x] Add signature verification
- [x] Create trust store for keys
- [x] Add checksum validation
- [x] Implement security audit command
- [x] Add vulnerability scanning

##### Initial Registry Setup
- [x] Set up registry infrastructure (GitHub/HTTP)
- [x] Create registry documentation
- [x] Publish core modules
- [x] Create example modules
- [x] Set up CI/CD for module publishing
- [x] Add module templates

#### Phase 3: Command Architecture Reorganization (Week 5)

##### Command Taxonomy Design
- [x] Generate current command inventory from CLI help
- [x] Define command categories
- [x] Create command grouping proposal
- [ ] Review with stakeholders
  - Prepared `docs/COMMAND_TAXONOMY_REVIEW.md` with the current generated
    command groups, evidence, review questions, and sign-off criteria.
  - Next action: send the packet to the stakeholder reviewer and record their
    answers, sign-off criteria decision, and approve/request-changes outcome
    before closing this item.
- [x] Finalize command hierarchy
- [x] Document command relationships

##### Command Registry Implementation
- [x] Create `cmd/pe/commands/registry.go`
- [x] Implement CommandRegistry type
- [x] Implement CommandGroup type
- [x] Add command registration methods
- [x] Implement command discovery
- [x] Add command metadata support
- [x] Create command help generator

##### Core Commands Group
- [x] Create `cmd/pe/commands/core/` directory
- [x] Move `run` command to core group
- [x] Move `build` command to core group
- [x] Move `test` command to core group
- [x] Move `ask` command to core group
- [x] Update command registrations
- [x] Add group-level help

##### Evaluation Commands Group
- [x] Create `cmd/pe/commands/evaluation/` directory
- [x] Move `eval` command to evaluation group
- [x] Move `benchmark` command to evaluation group
- [x] Move `diff` command to evaluation group
- [x] Move `stats` command to evaluation group
- [x] Move `view` command to evaluation group
- [x] Update command registrations

##### Optimization Commands Group
- [x] Create `cmd/pe/commands/optimization/` directory
- [x] Move `optimize` command to optimization group
- [x] Move `semantic` command to optimization group
- [x] Move `evolve` command to optimization group
- [x] Move `textgrad` command to optimization group
- [x] Move `pe2` command to optimization group
- [x] Move `gaso` command to optimization group
- [x] Move `fusion` command to optimization group

##### Module Commands Group
- [x] Create `cmd/pe/commands/module/` directory
- [x] Move all `mod` subcommands to module group
- [x] Move `push` command to module group
- [x] Move `get` command to module group
- [x] Update module command structure

##### Pipeline Commands Group
- [x] Create `cmd/pe/commands/pipeline/` directory
- [x] Move `stream` command to pipeline group
- [x] Move `filter` command to pipeline group
- [x] Move `extract` command to pipeline group
- [x] Move `compose` command to pipeline group
- [x] Move `cat` command to pipeline group

##### Utility Commands Group
- [x] Create `cmd/pe/commands/utility/` directory
- [x] Move `fmt` command to utility group
- [x] Move `vet` command to utility group
- [x] Move `convert` command to utility group
- [x] Move `template` command to utility group
- [x] Move `interactive` command to utility group
- [x] Move `watch` command to utility group

##### Experimental Commands Group
- [x] Create `cmd/pe/commands/experimental/` directory
- [x] Move `attest` command to experimental group
- [x] Move `security` command to experimental group
- [x] Move `profile` command to experimental group
- [x] Add experimental warning to commands

##### Command Integration
- [x] Update main.go to use command registry
- [x] Implement backward compatibility shims
- [x] Add command aliases for compatibility
- [x] Update command help system
- [x] Add command search functionality
- [x] Update shell completion scripts
- [x] Test all command paths

#### Phase 4: Testing Infrastructure (Week 6)

##### Testing Framework
- [x] Create `internal/testing/framework.go`
- [x] Implement TestFramework type
- [x] Add fixture management
- [x] Create assertion engine
- [x] Add test data generators
- [x] Implement test runners
- [x] Add parallel test support
- [x] Create test reporting

##### Mock Infrastructure
- [x] Create `internal/testing/mocks/` directory
- [x] Implement MockProvider with behavior config
- [x] Add deterministic response generation
- [x] Implement latency simulation
- [x] Add error injection capabilities
- [x] Create mock provider factory
- [x] Add call metrics tracking
- [x] Implement mock provider scenarios

##### Provider Tests
- [x] Add OpenAI provider unit tests
- [x] Add Anthropic provider unit tests
- [x] Add cgpt provider unit tests
- [x] Test provider registration
- [x] Test provider factory
- [x] Add streaming tests
- [x] Test error handling
- [x] Add timeout tests

##### Optimization Tests
- [x] Add textual gradient optimizer tests
- [x] Add PE2 optimizer tests
- [x] Add GASO optimizer tests
- [x] Add semantic optimizer tests
- [x] Add evolutionary optimizer tests
- [x] Test optimization convergence
- [x] Add property-based tests for optimization
- [x] Test optimization cancellation

##### Evaluation Tests
- [x] Add evaluation engine tests
- [x] Test all assertion types (20+)
- [x] Add Pass@N metric tests
- [x] Test parallel evaluation
- [x] Add scoring algorithm tests
- [x] Test evaluation caching
- [x] Add benchmark tests

##### Module System Tests
- [x] Add module resolution tests
- [x] Test dependency management
- [x] Add version constraint tests
- [x] Test module caching
- [x] Add registry client tests
- [x] Test module verification
- [x] Add integration tests

##### Command Tests
- [x] Add tests for core commands
- [x] Add tests for evaluation commands
- [x] Add tests for optimization commands
- [x] Add tests for module commands
- [x] Add tests for pipeline commands
- [x] Add tests for utility commands
- [x] Test command help output
- [x] Test command error handling

##### Integration Tests
- [x] Create `tests/integration/` directory
- [x] Add end-to-end workflow tests
- [x] Test optimization pipelines
- [x] Test evaluation workflows
- [x] Test module workflows
- [x] Add performance tests
- [x] Test concurrent operations
- [x] Add stress tests

##### Property-Based Tests
- [x] Add quickcheck for optimization
- [x] Test template substitution properties
- [x] Test evaluation scoring properties
- [x] Test module resolution properties
- [x] Test configuration validation
- [x] Test error handling properties

##### Test Coverage
- [x] Set up coverage reporting
- [x] Identify coverage gaps
- [x] Add tests for uncovered code
- [x] Achieve 70% coverage target
- [x] Set up coverage CI checks
- [x] Create coverage badges

#### Phase 5: Optimization Decoupling (Week 7)

##### Interface Design
- [x] Create `internal/optimization/interfaces.go`
- [x] Define Optimizer interface
- [x] Define Evaluator interface
- [x] Define GradientProvider interface
- [x] Define ObjectiveFunction interface
- [x] Add optimization context types
- [x] Document interface contracts

##### Strategy Pattern Implementation
- [x] Create `internal/optimization/strategy.go`
- [x] Implement OptimizationStrategy type
- [x] Add strategy selection logic
- [x] Implement composite strategies
- [x] Add strategy chaining
- [x] Create strategy factory
- [x] Add strategy configuration

##### Provider Adapters
- [x] Create `internal/optimization/adapters/` directory
- [x] Implement ProviderAdapter base type
- [x] Add OpenAI adapter
- [x] Add Anthropic adapter
- [x] Add generic inference adapter
- [x] Implement adapter caching
- [x] Add adapter metrics

##### Optimizer Refactoring
- [x] Refactor textual gradient optimizer to use interfaces
- [x] Refactor PE2 to use interfaces
- [x] Refactor GASO to use interfaces
- [x] Refactor semantic optimizer
- [x] Refactor evolutionary optimizer
- [x] Update hybrid optimizer
- [x] Remove provider coupling

##### Evaluation Abstraction
- [x] Create evaluation adapter interface
- [x] Implement metric-based evaluator
- [x] Add LLM-based evaluator
- [x] Implement human-in-loop evaluator
- [x] Add composite evaluator
- [x] Create evaluation pipeline

##### Testing Updates
- [x] Update optimization tests
- [x] Add adapter tests
- [x] Test strategy patterns
- [x] Add integration tests
- [x] Test with multiple providers
- [x] Verify no regressions

#### Phase 6: Configuration Management (Week 8)

##### Configuration Schema
- [x] Design configuration schema
- [x] Create `internal/config/schema.go`
- [x] Define configuration types
- [x] Add validation rules
- [x] Create default configurations
- [x] Document configuration options

##### Config Manager Implementation
- [x] Create `internal/config/manager.go`
- [x] Implement ConfigManager type
- [x] Add hierarchical lookup (CLI > ENV > File > Default)
- [x] Implement configuration sources
- [x] Add configuration caching
- [x] Implement hot reload
- [x] Add configuration watchers

##### Configuration Files
- [x] Define config file format (YAML/JSON; TOML reserved)
- [x] Create config file parser
- [x] Add config file discovery
- [x] Implement config file merging
- [x] Add config file validation
- [x] Create config migration tool

##### Environment Variables
- [x] Define environment variable schema
- [x] Implement env var parsing
- [x] Add env var validation
- [x] Create env var documentation
- [x] Add env var precedence rules

##### Provider Configuration
- [x] Update provider configs to use manager
- [x] Add API key management
- [x] Implement credential storage
- [x] Add provider-specific options
- [x] Create provider config validation

##### Command Configuration
- [x] Update commands to use config manager
- [x] Add command-specific configs
- [x] Implement config overrides
- [x] Add config profiles
- [x] Create config inheritance

##### Configuration Validation
- [x] Create `internal/config/validator.go`
- [x] Implement schema validation
- [x] Add type checking
- [x] Implement required field validation
- [x] Add custom validators
- [x] Create validation reports

##### Configuration Tools
- [x] Add `pe config get` command
- [x] Add `pe config set` command
- [x] Add `pe config list` command
- [x] Add `pe config validate` command
- [x] Add `pe config migrate` command
- [x] Create config documentation generator

#### Phase 7: Error Handling Enhancement (Week 9)

##### Error Type Definition
- [x] Create `internal/errors/types.go`
- [x] Define PEError base type
- [x] Define error codes enumeration
- [x] Create error categories
- [x] Add error metadata support
- [x] Define error severity levels

##### Domain-Specific Errors
- [x] Create ProviderError type
- [x] Create OptimizationError type
- [x] Create EvaluationError type
- [x] Create ModuleError type
- [x] Create ConfigurationError type
- [x] Create ValidationError type
- [x] Create NetworkError type
- [x] Create AuthenticationError type

##### Error Wrapping
- [x] Create `internal/errors/wrap.go`
- [x] Implement error wrapping utilities
- [x] Add context preservation
- [x] Implement error unwrapping
- [x] Add error chain support
- [x] Create error formatting

##### Error Recovery
- [x] Create `internal/errors/recovery.go`
- [x] Define RecoveryStrategy interface
- [x] Implement RetryStrategy
- [x] Implement ExponentialBackoff
- [x] Add CircuitBreaker pattern
- [x] Implement fallback strategies
- [x] Add recovery metrics

##### Error Handling Updates
- [x] Update provider error handling
- [x] Update command error handling
- [x] Update optimization error handling
- [x] Update evaluation error handling
- [x] Update module error handling
- [x] Add consistent error logging

##### Error Reporting
- [x] Create error reporting framework
- [x] Add structured error logging
- [x] Implement error aggregation
- [x] Add error metrics collection
- [x] Create error dashboards
- [x] Add error notifications

##### User-Facing Errors
- [x] Improve error messages
- [x] Add error suggestions
- [x] Create error documentation
- [x] Add error codes to docs
- [x] Implement error translation
- [x] Add troubleshooting guides

#### Phase 8: Observability (Week 10)

##### Metrics Infrastructure
- [x] Add metrics infrastructure under `internal/observability`
- [x] Implement MetricsCollector type
- [x] Add counter implementation
- [x] Add histogram implementation
- [x] Add gauge implementation
- [x] Add summary implementation
- [x] Create global metrics collector

##### Provider Metrics
- [x] Add request latency metrics
- [x] Add request count metrics
- [x] Add error rate metrics
- [x] Add token usage metrics
- [x] Add model-specific metrics
- [x] Add provider availability metrics

##### Optimization Metrics
- [x] Add optimization duration metrics
- [x] Add iteration count metrics
- [x] Add convergence metrics
- [x] Add improvement score metrics
- [x] Add resource usage metrics
- [x] Add cancellation metrics

##### Evaluation Metrics
- [x] Add evaluation latency metrics
- [x] Add test pass rate metrics
- [x] Add assertion metrics
- [x] Add parallel execution metrics
- [x] Add score distribution metrics

##### Distributed Tracing
- [x] Add tracing infrastructure under `internal/observability`
- [x] Integrate OpenTelemetry
- [x] Add trace provider setup
- [x] Implement span creation
- [x] Add context propagation
- [x] Create file trace writer
- [x] Add trace sampling

##### Command Tracing
- [x] Add tracing to run command
- [x] Add tracing to eval command
- [x] Add tracing to optimize commands
- [x] Add tracing to module commands
- [x] Add tracing to pipeline commands
- [x] Create trace visualization

##### Logging Enhancement
- [x] Implement structured logging
- [x] Add log levels
- [x] Create log formatters
- [x] Add log rotation
- [x] Implement log aggregation
- [x] Add log correlation IDs

##### Monitoring Integration
- [x] Add Prometheus exporter
- [x] Create Grafana dashboards
- [x] Add alert definitions
- [x] Create runbooks
- [x] Add SLO definitions
- [x] Implement health checks

##### Performance Profiling
- [x] Add CPU profiling
- [x] Add memory profiling
- [x] Add goroutine profiling
- [x] Add block profiling
- [x] Create profile analysis tools
- [x] Add continuous profiling

#### Cross-Cutting Concerns

##### Documentation Updates
- [x] Update architecture documentation
- [x] Update API documentation
- [x] Update command documentation
- [x] Create migration guides
- [x] Update examples
- [x] Add troubleshooting guides
- [x] Create contributor guides

##### CI/CD Updates
- [x] Update GitHub Actions workflows
- [x] Add test coverage checks
- [x] Add performance regression tests
- [x] Add security scanning
- [x] Update release process
- [x] Add automated benchmarks

##### Performance Optimization
- [x] Profile critical paths
- [x] Optimize hot loops
- [x] Add caching layers
- [x] Implement connection pooling
- [x] Optimize memory allocations
- [x] Add performance tests

##### Security Enhancements
- [x] Security audit of new code
- [x] Add input sanitization
- [x] Implement rate limiting
- [x] Add authentication checks
- [x] Implement authorization
- [x] Add security tests

##### Backward Compatibility
- [x] Create compatibility layer
- [x] Add deprecation warnings
- [x] Create migration tools
- [x] Update upgrade guides
- [x] Test compatibility paths
- [x] Document breaking changes

#### Project Management

##### Planning & Tracking
- [x] Set up project board
- [x] Create sprint plans
- [x] Define milestones
- [x] Track velocity
- [x] Update stakeholders
- [x] Manage dependencies

##### Quality Assurance
- [x] Code review process
- [x] Test plan creation
- [x] Bug tracking
- [x] Performance validation
- [x] Security review
- [x] Documentation review

##### Release Management
- [x] Version planning
- [x] Release notes preparation
- [x] Release testing
- [x] Deployment procedures
- [x] Rollback planning
- [x] Post-release monitoring

#### Total Tasks: ~450+

This comprehensive todo list can be imported into project management tools like GitHub Projects, Jira, or Linear for tracking. Each task should be assigned to appropriate team members with time estimates and dependencies mapped.
