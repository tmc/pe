# PE Roadmap

Last updated: 2026-05-05

This file is the source of truth for planned PE work. Beads is deprecated for this repository: do not create or update `.beads` issues for new work. Keep roadmap changes in tracked commits with the code or documentation they describe.

## Document Ownership

- `ROADMAP.md`: remaining work, release blockers, priorities, and planning
  status.
- `CHANGELOG.md`: concise release deltas only.
- `RELEASE_NOTES.md`: v0.5.0 release-candidate narrative, known limitations,
  and validation checklist.
- `docs/PLANNED_COMMANDS.md`: aspirational command ideas, not status or command
  counts.
- `docs/IMPLEMENTATION_TODOS.md`: tombstone pointing here.
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
- Recent scripttest, provider, and release-note work is documented; remaining
  command and getting-started docs still need a pass
- `docs/CLI_HELP_AUDIT.md` records the generated root command inventory from
  `go run ./cmd/pe --help`; `docs/CLI_REFERENCE.md` now includes every root
  command from that inventory.

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
- docs/COMMANDS.md
- docs/CLI_REFERENCE.md
- docs/COMMAND_REFERENCE.md (if different)
- CLI_COMMANDS_REFERENCE.md (root)
- docs/COMMAND_EXAMPLES_GUIDE.md

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
- Current docs are registry/dependency-centric. They do not define capabilities, placement, data classes, prompt provenance, provider/tool allow-deny, typed IO requirements, or conservative dependency policy composition.

Design direction:
1. Keep `pe.mod` Go-like and module-scoped.
2. Put module-wide `capability`, `placement`, and `policy` blocks in `pe.mod`.
3. Keep prompt-specific inputs, output schemas, shebang runners, and local metadata in per-file front matter.
4. Keep runtime facts in trace artifacts.
5. Compose policies conservatively: intersect allows, union denials, and fail strict dependencies that request denied capabilities.

Initial artifact:
- `docs/future/PEMOD_CAPABILITIES_DESIGN.md`

Version plan:
1. v0.5: docs only.
2. v0.6: parser/schema support in `internal/pemod`.
3. v0.6: static validation through `pe mod` / `pe vet`.
4. v0.7: runtime enforcement for providers, tools, file writes, network, and placement.

Verification:
- `test -f docs/future/PEMOD_CAPABILITIES_DESIGN.md`
- `rg "capability|placement|policy|Conservative Composition" docs/future/PEMOD_CAPABILITIES_DESIGN.md`


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
- `pe mod download`, `pe mod list`, `pe mod search`, and `pe mod publish` have
  command implementations.
- Module registry support still needs release validation and clearer user
  documentation.

Tasks:
1. Validate the implemented module commands against a real or fixture registry.
2. Complete `pe mod tidy` and `pe mod vendor` behavior.
3. Add integrity verification, version upgrade, and dependency graph behavior.
4. Document supported registry configuration and failure modes.
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

Current:
- Basic assertions work (contains, equals, regex, etc.)
- Pass@N and structured output implemented
- LLM rubric assertions work

Implementation notes:
- May require external models/APIs for toxicity
- Coherence could use perplexity scoring
- Factuality needs knowledge base integration
- Similarity can use embeddings

Location: internal/promptfoo/evaluation/metrics/


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
- Help text correctly says these manifests are unsigned and local-only: they
  detect content changes but do not prove identity, origin, or freshness.

Tasks:
1. Define a signed manifest envelope that preserves the existing unsigned
   manifest payload.
2. Start with Ed25519 from the standard library.
3. Add explicit key-generation, signing, verification, and failure-mode docs.
4. Keep unsigned manifests available for local integrity workflows.

Verification:
- `go run ./cmd/pe exp attest --help`
- `rg 'unsignedManifestType|ed25519|signature' cmd/pe`


#### Build remote registry download path

- Type: `epic`

**Scope**

Turn the module registry code into a verified remote download path with clear
integrity and failure semantics.

Current status:
- Module commands and registry code exist.
- `pe mod tidy --json --write` is implemented for local dependency hygiene.
- Remote registry behavior still needs fixture-backed validation and user docs.

Tasks:
1. Add a fixture-backed remote registry test path before relying on live
   services.
2. Verify `pe mod download` extracts exactly the expected files and rejects
   path escapes.
3. Define integrity checks for downloaded modules.
4. Document supported registry configuration and failure modes.

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


## Architecture Implementation Backlog

This checklist was moved from `docs/IMPLEMENTATION_TODOS.md` so roadmap work lives in one tracked file.

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
- [ ] Update `cmd/pe/eval.go` for evaluation commands
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
- [ ] Remove duplicate provider implementations
- [x] Consolidate provider registration logic
- [x] Update provider factory methods

##### Cleanup & Validation
- [ ] Delete `internal/llm/provider.go`
- [ ] Remove all legacy provider implementations
- [ ] Update all import statements
- [ ] Fix compilation errors
- [ ] Run full test suite
- [ ] Manual testing of critical paths
- [ ] Performance regression testing
- [ ] Update documentation

#### Phase 2: Module System Implementation (Weeks 3-4)

##### Registry Design
- [ ] Research registry implementation options
- [ ] Design registry API specification
- [ ] Define module metadata format
- [ ] Design module versioning scheme
- [ ] Create module signature format
- [ ] Design dependency resolution algorithm
- [ ] Plan caching strategy
- [ ] Document registry protocol

##### Registry Implementation
- [ ] Create `internal/module/registry.go`
- [ ] Implement registry client interface
- [ ] Add GitHub-based registry option
- [ ] Add HTTP API registry option
- [ ] Implement registry authentication
- [ ] Add module search functionality
- [ ] Implement module metadata fetching
- [ ] Add registry health checks

##### Module Resolution
- [ ] Create `internal/module/resolver.go`
- [ ] Implement module path parsing
- [ ] Add version constraint parsing
- [ ] Implement semantic version comparison
- [ ] Add module cache interface
- [ ] Implement file-based cache
- [ ] Add cache invalidation logic
- [ ] Implement module download functionality

##### Dependency Management
- [ ] Create `internal/module/deps.go`
- [ ] Implement dependency graph structure
- [ ] Add topological sort for dependencies
- [ ] Implement conflict detection
- [ ] Add version resolution algorithm
- [ ] Implement circular dependency detection
- [ ] Add dependency pruning
- [ ] Create lock file format

##### Module Commands
- [x] Fix `pe mod init` with proper initialization
- [x] Implement `pe mod download` with real registry
- [ ] Complete `pe mod tidy` functionality
- [ ] Implement `pe mod vendor` properly
- [ ] Add `pe mod verify` for integrity checking
- [x] Implement `pe mod list` for installed modules
- [x] Add `pe mod search` for registry search
- [x] Implement `pe mod publish` for module publishing
- [ ] Add `pe mod upgrade` for version updates
- [ ] Implement `pe mod graph` for dependency visualization

##### Module Security
- [ ] Implement module signing
- [ ] Add signature verification
- [ ] Create trust store for keys
- [ ] Add checksum validation
- [ ] Implement security audit command
- [ ] Add vulnerability scanning

##### Initial Registry Setup
- [ ] Set up registry infrastructure (GitHub/HTTP)
- [ ] Create registry documentation
- [ ] Publish core modules
- [ ] Create example modules
- [ ] Set up CI/CD for module publishing
- [ ] Add module templates

#### Phase 3: Command Architecture Reorganization (Week 5)

##### Command Taxonomy Design
- [ ] Generate current command inventory from CLI help
- [ ] Define command categories
- [ ] Create command grouping proposal
- [ ] Review with stakeholders
- [ ] Finalize command hierarchy
- [ ] Document command relationships

##### Command Registry Implementation
- [ ] Create `cmd/pe/commands/registry.go`
- [ ] Implement CommandRegistry type
- [ ] Implement CommandGroup type
- [ ] Add command registration methods
- [ ] Implement command discovery
- [ ] Add command metadata support
- [ ] Create command help generator

##### Core Commands Group
- [ ] Create `cmd/pe/commands/core/` directory
- [ ] Move `run` command to core group
- [ ] Move `build` command to core group
- [ ] Move `test` command to core group
- [ ] Move `ask` command to core group
- [ ] Update command registrations
- [ ] Add group-level help

##### Evaluation Commands Group
- [ ] Create `cmd/pe/commands/evaluation/` directory
- [ ] Move `eval` command to evaluation group
- [ ] Move `benchmark` command to evaluation group
- [ ] Move `diff` command to evaluation group
- [ ] Move `stats` command to evaluation group
- [ ] Move `view` command to evaluation group
- [ ] Update command registrations

##### Optimization Commands Group
- [ ] Create `cmd/pe/commands/optimization/` directory
- [ ] Move `optimize` command to optimization group
- [ ] Move `semantic` command to optimization group
- [ ] Move `evolve` command to optimization group
- [ ] Move `textgrad` command to optimization group
- [ ] Move `pe2` command to optimization group
- [ ] Move `gaso` command to optimization group
- [ ] Move `fusion` command to optimization group

##### Module Commands Group
- [ ] Create `cmd/pe/commands/module/` directory
- [ ] Move all `mod` subcommands to module group
- [ ] Move `push` command to module group
- [ ] Move `get` command to module group
- [ ] Update module command structure

##### Pipeline Commands Group
- [ ] Create `cmd/pe/commands/pipeline/` directory
- [ ] Move `stream` command to pipeline group
- [ ] Move `filter` command to pipeline group
- [ ] Move `extract` command to pipeline group
- [ ] Move `compose` command to pipeline group
- [ ] Move `cat` command to pipeline group

##### Utility Commands Group
- [ ] Create `cmd/pe/commands/utility/` directory
- [ ] Move `fmt` command to utility group
- [ ] Move `vet` command to utility group
- [ ] Move `convert` command to utility group
- [ ] Move `template` command to utility group
- [ ] Move `interactive` command to utility group
- [ ] Move `watch` command to utility group

##### Experimental Commands Group
- [ ] Create `cmd/pe/commands/experimental/` directory
- [ ] Move `attest` command to experimental group
- [ ] Move `security` command to experimental group
- [ ] Move `profile` command to experimental group
- [ ] Add experimental warning to commands

##### Command Integration
- [ ] Update main.go to use command registry
- [ ] Implement backward compatibility shims
- [ ] Add command aliases for compatibility
- [ ] Update command help system
- [ ] Add command search functionality
- [ ] Update shell completion scripts
- [ ] Test all command paths

#### Phase 4: Testing Infrastructure (Week 6)

##### Testing Framework
- [ ] Create `internal/testing/framework.go`
- [ ] Implement TestFramework type
- [ ] Add fixture management
- [ ] Create assertion engine
- [ ] Add test data generators
- [ ] Implement test runners
- [ ] Add parallel test support
- [ ] Create test reporting

##### Mock Infrastructure
- [ ] Create `internal/testing/mocks/` directory
- [ ] Implement MockProvider with behavior config
- [ ] Add deterministic response generation
- [ ] Implement latency simulation
- [ ] Add error injection capabilities
- [ ] Create mock provider factory
- [ ] Add call metrics tracking
- [ ] Implement mock provider scenarios

##### Provider Tests
- [ ] Add OpenAI provider unit tests
- [ ] Add Anthropic provider unit tests
- [ ] Add cgpt provider unit tests
- [ ] Test provider registration
- [ ] Test provider factory
- [ ] Add streaming tests
- [ ] Test error handling
- [ ] Add timeout tests

##### Optimization Tests
- [ ] Add textual gradient optimizer tests
- [ ] Add PE2 optimizer tests
- [ ] Add GASO optimizer tests
- [ ] Add semantic optimizer tests
- [ ] Add evolutionary optimizer tests
- [ ] Test optimization convergence
- [ ] Add property-based tests for optimization
- [ ] Test optimization cancellation

##### Evaluation Tests
- [ ] Add evaluation engine tests
- [ ] Test all assertion types (20+)
- [ ] Add Pass@N metric tests
- [ ] Test parallel evaluation
- [ ] Add scoring algorithm tests
- [ ] Test evaluation caching
- [ ] Add benchmark tests

##### Module System Tests
- [ ] Add module resolution tests
- [ ] Test dependency management
- [ ] Add version constraint tests
- [ ] Test module caching
- [ ] Add registry client tests
- [ ] Test module verification
- [ ] Add integration tests

##### Command Tests
- [ ] Add tests for core commands
- [ ] Add tests for evaluation commands
- [ ] Add tests for optimization commands
- [ ] Add tests for module commands
- [ ] Add tests for pipeline commands
- [ ] Add tests for utility commands
- [ ] Test command help output
- [ ] Test command error handling

##### Integration Tests
- [ ] Create `tests/integration/` directory
- [ ] Add end-to-end workflow tests
- [ ] Test optimization pipelines
- [ ] Test evaluation workflows
- [ ] Test module workflows
- [ ] Add performance tests
- [ ] Test concurrent operations
- [ ] Add stress tests

##### Property-Based Tests
- [ ] Add quickcheck for optimization
- [ ] Test template substitution properties
- [ ] Test evaluation scoring properties
- [ ] Test module resolution properties
- [ ] Test configuration validation
- [ ] Test error handling properties

##### Test Coverage
- [ ] Set up coverage reporting
- [ ] Identify coverage gaps
- [ ] Add tests for uncovered code
- [ ] Achieve 70% coverage target
- [ ] Set up coverage CI checks
- [ ] Create coverage badges

#### Phase 5: Optimization Decoupling (Week 7)

##### Interface Design
- [ ] Create `internal/optimization/interfaces.go`
- [ ] Define Optimizer interface
- [ ] Define Evaluator interface
- [ ] Define GradientProvider interface
- [ ] Define ObjectiveFunction interface
- [ ] Add optimization context types
- [ ] Document interface contracts

##### Strategy Pattern Implementation
- [ ] Create `internal/optimization/strategy.go`
- [ ] Implement OptimizationStrategy type
- [ ] Add strategy selection logic
- [ ] Implement composite strategies
- [ ] Add strategy chaining
- [ ] Create strategy factory
- [ ] Add strategy configuration

##### Provider Adapters
- [ ] Create `internal/optimization/adapters/` directory
- [ ] Implement ProviderAdapter base type
- [ ] Add OpenAI adapter
- [ ] Add Anthropic adapter
- [ ] Add generic inference adapter
- [ ] Implement adapter caching
- [ ] Add adapter metrics

##### Optimizer Refactoring
- [ ] Refactor textual gradient optimizer to use interfaces
- [ ] Refactor PE2 to use interfaces
- [ ] Refactor GASO to use interfaces
- [ ] Refactor semantic optimizer
- [ ] Refactor evolutionary optimizer
- [ ] Update hybrid optimizer
- [ ] Remove provider coupling

##### Evaluation Abstraction
- [ ] Create evaluation adapter interface
- [ ] Implement metric-based evaluator
- [ ] Add LLM-based evaluator
- [ ] Implement human-in-loop evaluator
- [ ] Add composite evaluator
- [ ] Create evaluation pipeline

##### Testing Updates
- [ ] Update optimization tests
- [ ] Add adapter tests
- [ ] Test strategy patterns
- [ ] Add integration tests
- [ ] Test with multiple providers
- [ ] Verify no regressions

#### Phase 6: Configuration Management (Week 8)

##### Configuration Schema
- [ ] Design configuration schema
- [ ] Create `internal/config/schema.go`
- [ ] Define configuration types
- [ ] Add validation rules
- [ ] Create default configurations
- [ ] Document configuration options

##### Config Manager Implementation
- [ ] Create `internal/config/manager.go`
- [ ] Implement ConfigManager type
- [ ] Add hierarchical lookup (CLI > ENV > File > Default)
- [ ] Implement configuration sources
- [ ] Add configuration caching
- [ ] Implement hot reload
- [ ] Add configuration watchers

##### Configuration Files
- [ ] Define config file format (YAML/TOML/JSON)
- [ ] Create config file parser
- [ ] Add config file discovery
- [ ] Implement config file merging
- [ ] Add config file validation
- [ ] Create config migration tool

##### Environment Variables
- [ ] Define environment variable schema
- [ ] Implement env var parsing
- [ ] Add env var validation
- [ ] Create env var documentation
- [ ] Add env var precedence rules

##### Provider Configuration
- [ ] Update provider configs to use manager
- [ ] Add API key management
- [ ] Implement credential storage
- [ ] Add provider-specific options
- [ ] Create provider config validation

##### Command Configuration
- [ ] Update commands to use config manager
- [ ] Add command-specific configs
- [ ] Implement config overrides
- [ ] Add config profiles
- [ ] Create config inheritance

##### Configuration Validation
- [ ] Create `internal/config/validator.go`
- [ ] Implement schema validation
- [ ] Add type checking
- [ ] Implement required field validation
- [ ] Add custom validators
- [ ] Create validation reports

##### Configuration Tools
- [ ] Add `pe config get` command
- [ ] Add `pe config set` command
- [ ] Add `pe config list` command
- [ ] Add `pe config validate` command
- [ ] Add `pe config migrate` command
- [ ] Create config documentation generator

#### Phase 7: Error Handling Enhancement (Week 9)

##### Error Type Definition
- [ ] Create `internal/errors/types.go`
- [ ] Define PEError base type
- [ ] Define error codes enumeration
- [ ] Create error categories
- [ ] Add error metadata support
- [ ] Define error severity levels

##### Domain-Specific Errors
- [ ] Create ProviderError type
- [ ] Create OptimizationError type
- [ ] Create EvaluationError type
- [ ] Create ModuleError type
- [ ] Create ConfigurationError type
- [ ] Create ValidationError type
- [ ] Create NetworkError type
- [ ] Create AuthenticationError type

##### Error Wrapping
- [ ] Create `internal/errors/wrap.go`
- [ ] Implement error wrapping utilities
- [ ] Add context preservation
- [ ] Implement error unwrapping
- [ ] Add error chain support
- [ ] Create error formatting

##### Error Recovery
- [ ] Create `internal/errors/recovery.go`
- [ ] Define RecoveryStrategy interface
- [ ] Implement RetryStrategy
- [ ] Implement ExponentialBackoff
- [ ] Add CircuitBreaker pattern
- [ ] Implement fallback strategies
- [ ] Add recovery metrics

##### Error Handling Updates
- [ ] Update provider error handling
- [ ] Update command error handling
- [ ] Update optimization error handling
- [ ] Update evaluation error handling
- [ ] Update module error handling
- [ ] Add consistent error logging

##### Error Reporting
- [ ] Create error reporting framework
- [ ] Add structured error logging
- [ ] Implement error aggregation
- [ ] Add error metrics collection
- [ ] Create error dashboards
- [ ] Add error notifications

##### User-Facing Errors
- [ ] Improve error messages
- [ ] Add error suggestions
- [ ] Create error documentation
- [ ] Add error codes to docs
- [ ] Implement error translation
- [ ] Add troubleshooting guides

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
- [ ] Add request latency metrics
- [ ] Add request count metrics
- [ ] Add error rate metrics
- [ ] Add token usage metrics
- [ ] Add model-specific metrics
- [ ] Add provider availability metrics

##### Optimization Metrics
- [ ] Add optimization duration metrics
- [ ] Add iteration count metrics
- [ ] Add convergence metrics
- [ ] Add improvement score metrics
- [ ] Add resource usage metrics
- [ ] Add cancellation metrics

##### Evaluation Metrics
- [ ] Add evaluation latency metrics
- [ ] Add test pass rate metrics
- [ ] Add assertion metrics
- [ ] Add parallel execution metrics
- [ ] Add score distribution metrics

##### Distributed Tracing
- [x] Add tracing infrastructure under `internal/observability`
- [ ] Integrate OpenTelemetry
- [ ] Add trace provider setup
- [x] Implement span creation
- [x] Add context propagation
- [x] Create file trace writer
- [ ] Add trace sampling

##### Command Tracing
- [ ] Add tracing to run command
- [ ] Add tracing to eval command
- [ ] Add tracing to optimize commands
- [ ] Add tracing to module commands
- [ ] Add tracing to pipeline commands
- [ ] Create trace visualization

##### Logging Enhancement
- [ ] Implement structured logging
- [ ] Add log levels
- [ ] Create log formatters
- [ ] Add log rotation
- [ ] Implement log aggregation
- [ ] Add log correlation IDs

##### Monitoring Integration
- [ ] Add Prometheus exporter
- [ ] Create Grafana dashboards
- [ ] Add alert definitions
- [ ] Create runbooks
- [ ] Add SLO definitions
- [ ] Implement health checks

##### Performance Profiling
- [x] Add CPU profiling
- [x] Add memory profiling
- [x] Add goroutine profiling
- [x] Add block profiling
- [ ] Create profile analysis tools
- [ ] Add continuous profiling

#### Cross-Cutting Concerns

##### Documentation Updates
- [ ] Update architecture documentation
- [ ] Update API documentation
- [ ] Update command documentation
- [ ] Create migration guides
- [ ] Update examples
- [ ] Add troubleshooting guides
- [ ] Create contributor guides

##### CI/CD Updates
- [ ] Update GitHub Actions workflows
- [ ] Add test coverage checks
- [ ] Add performance regression tests
- [ ] Add security scanning
- [ ] Update release process
- [ ] Add automated benchmarks

##### Performance Optimization
- [ ] Profile critical paths
- [ ] Optimize hot loops
- [ ] Add caching layers
- [ ] Implement connection pooling
- [ ] Optimize memory allocations
- [ ] Add performance tests

##### Security Enhancements
- [ ] Security audit of new code
- [ ] Add input sanitization
- [ ] Implement rate limiting
- [ ] Add authentication checks
- [ ] Implement authorization
- [ ] Add security tests

##### Backward Compatibility
- [ ] Create compatibility layer
- [ ] Add deprecation warnings
- [ ] Create migration tools
- [ ] Update upgrade guides
- [ ] Test compatibility paths
- [ ] Document breaking changes

#### Project Management

##### Planning & Tracking
- [ ] Set up project board
- [ ] Create sprint plans
- [ ] Define milestones
- [ ] Track velocity
- [ ] Update stakeholders
- [ ] Manage dependencies

##### Quality Assurance
- [ ] Code review process
- [ ] Test plan creation
- [ ] Bug tracking
- [ ] Performance validation
- [ ] Security review
- [ ] Documentation review

##### Release Management
- [ ] Version planning
- [ ] Release notes preparation
- [ ] Release testing
- [ ] Deployment procedures
- [ ] Rollback planning
- [ ] Post-release monitoring

#### Total Tasks: ~450+

This comprehensive todo list can be imported into project management tools like GitHub Projects, Jira, or Linear for tracking. Each task should be assigned to appropriate team members with time estimates and dependencies mapped.
