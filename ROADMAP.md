# PE Roadmap

Last updated: 2026-04-30

This file is the source of truth for planned PE work. Beads is deprecated for this repository: do not create or update `.beads` issues for new work. Keep roadmap changes in tracked commits with the code or documentation they describe.

The active work items below were migrated from the legacy Beads database. Legacy IDs are retained only for traceability.

## Priority Guide

- **P1**: Release blocking or user-facing quality work.
- **P2**: Important product, design, or maintenance work.
- **P3**: Backlog features and cleanup.

## Active Work

### P1

#### Release prep: Documentation accuracy audit

- Legacy ID: `pe-30`
- Type: `epic`

**Scope**

Comprehensive documentation accuracy audit and update for release.

Current Issues:
- Test coverage claims inconsistent (25% vs 40%)
- Dates outdated (Jan/Feb 2025, should be Oct 2025)
- Multiple overlapping getting started docs
- Future docs mixed with current implementation docs
- TODO.md severely outdated
- Recent work not documented (scripttest, fixes, etc.)

Sub-tasks to create:
1. Reconcile test coverage numbers
2. Update all documentation dates
3. Consolidate getting started documentation
4. Update docs/CURRENT_STATUS.md with latest
5. Migrate TODO.md items to this roadmap
6. Document recent scripttest work
7. Update RELEASE_NOTES.md
8. Clean up docs/future/ organization
9. Fix README.md accuracy
10. Review all command documentation

Priority: P1 - Blocking release


#### Release prep: Examples validation

- Legacy ID: `pe-31`
- Type: `epic`

**Scope**

Validate and update all examples for release.

Ensure all examples work with current codebase:

Areas to validate:
- example/ directory examples
- examples/ directory examples
- Code examples in documentation
- README code samples
- Tutorial examples
- Command reference examples

Sub-tasks:
1. Test all example/ directory samples
2. Test all examples/ directory samples
3. Verify documentation code examples
4. Update broken examples
5. Add missing examples for new features
6. Create examples index/README
7. Validate template syntax ({{.var}} not {{var}})
8. Test with all supported providers

Priority: P1 - User-facing quality


#### Release prep: Version and changelog

- Legacy ID: `pe-33`
- Type: `task`

**Scope**

Prepare version number and comprehensive changelog for release.

Tasks:
1. Determine version number (0.1.0? 0.5.0? 1.0.0?)
2. Review all commits since last release
3. Categorize changes:
   - Breaking changes
   - New features
   - Bug fixes
   - Documentation improvements
   - Testing improvements
4. Update RELEASE_NOTES.md
5. Create/update CHANGELOG.md
6. Document migration path if breaking changes
7. Tag version in git

Recent work to include:
- Scripttest framework fixes (pe-16)
- Test improvements (pe-12, pe-10, pe-11)
- Documentation additions (MARKERS.md)
- .gitignore improvements


#### Release prep: README consolidation

- Legacy ID: `pe-36`
- Type: `task`

**Scope**

Consolidate and align README files.

Current state:
- README.md (root) - main project README
- docs/README.md - documentation index
- Potential inconsistencies between them

Tasks:
1. Review README.md for accuracy
2. Update with latest features and status
3. Ensure consistency with docs/README.md
4. Update badges and links
5. Add clear feature status indicators (✅ Implemented, ⚠️ Partial, 📝 Planned)
6. Link to release notes and changelog
7. Update contribution guidelines reference
8. Add clear next steps for users

Key sections to update:
- Feature list with accurate status
- Installation instructions
- Quick start examples
- Documentation links
- Getting help section


#### Release prep: Security review

- Legacy ID: `pe-39`
- Type: `task`

**Scope**

Security review before release.

Tasks:
1. Review for hardcoded secrets or API keys
2. Check for command injection vulnerabilities
3. Verify input validation (already have tests)
4. Review file path traversal protections
5. Check dependency vulnerabilities (go list -m all)
6. Review security.md if exists
7. Test with untrusted input
8. Review error messages for info disclosure

Tools to use:
- gosec (static analysis)
- go list -m all | nancy (or govulncheck)
- Manual code review of security-sensitive areas

Critical areas:
- Command execution (scripttest, providers)
- File operations
- Network requests
- Template evaluation


#### Release prep: Build and distribution

- Legacy ID: `pe-40`
- Type: `task`

**Scope**

Prepare build and distribution infrastructure.

Tasks:
1. Verify 'go install' works correctly
2. Test cross-compilation (multiple platforms)
3. Set up GitHub releases workflow (if not exists)
4. Prepare release binaries
5. Document build process
6. Create install script if needed
7. Test installation from various sources
8. Verify binary size is reasonable
9. Check for unnecessary dependencies in binary

Platforms to test:
- macOS (arm64, amd64)
- Linux (amd64, arm64)
- Windows (amd64) if supported

Distribution methods:
- go install (primary)
- GitHub releases (binaries)
- Homebrew formula (future)
- Docker image (future)


### P2

#### Add pipe support to scripttest framework

- Legacy ID: `pe-6`
- Type: `feature`

**Scope**

The scripttest framework doesn't currently support shell pipes, which prevents testing commands that use pipes.

Impact:
- tests/testdata/script/extract.txt - All tests commented out (require pipes)
- Future workflow tests will be limited

Options:
1. Add native pipe support to scripttest
2. Create intermediate file approach (command1 output to file, command2 reads file)
3. Wait for upstream scripttest to add pipe support

Currently extract.txt has a minimal test using grep to verify the file works.

**Notes**

Pipe support blocked by scripttest framework limitations

Current situation:
- scripttest framework doesn't support shell pipes (|)
- Multiple test files need pipe support:
  * pipeline.txt - Entire file commented out
  * extract.txt - Pipe-dependent tests commented out
  * Future workflow tests will be limited

Options:
1. Add native pipe support to scripttest framework
   - Requires modifying rogpeppe/go-internal/testscript
   - Complex implementation
   - Would benefit upstream project

2. Intermediate file approach (workaround)
   - command1 writes to temp file
   - command2 reads from temp file
   - Works but verbose and less intuitive

3. Wait for upstream testscript
   - Monitor rogpeppe/go-internal for pipe support
   - Contribute PR to upstream if motivated

4. Alternative test framework
   - Consider bash-based integration tests
   - Use bats or similar
   - Maintain both scripttest and bash tests

Recommendation: Option 2 (workaround) for immediate needs, watch Option 1 (upstream PR) for long term.

Related: pe-16 (completed - documented pipe limitations)
Related: pe-19 (scripttest docs should mention this limitation)


#### Maintain roadmap with current status

- Legacy ID: `pe-18`
- Type: `task`

**Scope**

The old TODO file is superseded. Keep this roadmap current as the tracked source of truth.

Need to:
1. Review completed items and mark appropriately
2. Add new work items to this roadmap
3. Update project status
4. Update references to point at ROADMAP.md
5. Remove obsolete TODO and Beads references as they are found


#### Add scripttest framework documentation

- Legacy ID: `pe-19`
- Type: `task`

**Scope**

Create comprehensive documentation for scripttest framework used in tests/.

Should cover:
1. Overview of scripttest framework and philosophy
2. Available commands (pe, cat, echo, exists, env, grep, etc.)
3. File creation with -- markers
4. Variable substitution
5. Output matching (stdout, stderr)
6. Test organization and best practices
7. Limitations (no exec, no pipes currently)
8. Examples from tests/testdata/script/

Reference:
- tests/scripttest_test.go
- tests/testdata/script/*.txt examples
- MARKERS.md for prompt format context


#### Implement Ollama provider

- Legacy ID: `pe-20`
- Type: `feature`

**Scope**

Add support for Ollama as a local LLM provider.

Current status:
- Provider interface exists in internal/inference/providers/
- Ollama provider directory exists with basic test (73.9% coverage)
- Need to complete implementation

Tasks:
1. Review existing ollama provider code
2. Implement complete API integration
3. Add configuration support
4. Add tests for various models
5. Document usage in README and docs
6. Add examples

Benefits:
- Local model support (no API keys needed)
- Privacy-focused option
- Cost-effective for development
- Support for open models (llama, mistral, etc.)


#### Improve test coverage

- Legacy ID: `pe-25`
- Type: `task`

**Scope**

Increase test coverage across the project.

Current coverage (from go test output):
- Most packages have good coverage (70%+)
- Some packages lack tests entirely ([no test files])

Priority areas:
1. internal/cli - No test files
2. internal/module - No test files
3. internal/optimization - No test files
4. internal/plugin - No test files
5. internal/templates - No test files

Tasks:
1. Add unit tests for untested packages
2. Add integration tests for end-to-end workflows
3. Add property-based tests where appropriate
4. Target 80%+ coverage for critical paths
5. Add benchmark tests for performance-sensitive code

Tools:
- go test -cover
- go test -coverprofile
- Use testify for assertions


#### Document scripttest limitations in code comments

- Legacy ID: `pe-28`
- Type: `task`

**Scope**

Add inline documentation about scripttest framework limitations.

Based on pe-16 work, add comments to tests/scripttest_test.go explaining:

1. Why exec is disabled (security)
2. Why pipes aren't supported (testscript framework limitation)
3. Why certain shell features are unavailable:
   - Background jobs (&)
   - Shell redirection (>, <, >>)
   - Complex shell features (process substitution, etc.)

4. Workarounds and alternatives:
   - Use -- file markers instead of heredoc
   - Use intermediate files instead of pipes
   - Use direct commands instead of exec

5. Link to upstream testscript docs

This will help future developers understand the constraints and make better test design decisions.

Related:
- pe-16: Completed scripttest fixes
- pe-6: Pipe support investigation
- pe-19: Scripttest documentation (should reference this)


#### Add llm CLI support as provider backend

- Legacy ID: `pe-29`
- Type: `epic`

**Scope**

Add support for Simon Willison's llm CLI tool as a provider backend.

Simon's llm tool (https://llm.datasette.io/) is a powerful CLI for interacting with LLMs
that supports multiple providers and models through a plugin system.

Epic Goals:
1. Implement llm provider adapter
2. Support llm's model aliases and configuration
3. Enable access to llm's extensive plugin ecosystem
4. Document llm integration

Benefits:
- Access to 40+ LLM providers via llm plugins
- Local model support (llama-cpp, mlc, etc.)
- Consistent interface for diverse backends
- Leverage llm's quality and community

Implementation Tasks:
- Research llm CLI interface and JSON output
- Design provider adapter for llm backend
- Implement LLMProvider in internal/inference/providers/
- Add llm-specific configuration support
- Test with various llm providers/plugins
- Document llm setup and usage
- Add examples for common llm workflows

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

- Legacy ID: `pe-35`
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
1. All 47 commands documented
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

- Legacy ID: `pe-37`
- Type: `task`

**Scope**

Reorganize documentation for clarity and discoverability.

Current issues:
- 47 files in docs/ directory
- Overlapping content (3 getting started docs, multiple tutorials)
- docs/future/ mixed with current docs
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
1. Create new directory structure
2. Consolidate duplicate docs
3. Move files to appropriate locations
4. Update all internal links
5. Create comprehensive index in docs/README.md
6. Add navigation/breadcrumbs
7. Archive outdated docs


#### Release prep: License and legal review

- Legacy ID: `pe-38`
- Type: `task`

**Scope**

Review license and legal documentation for release.

Tasks:
1. Verify LICENSE file is present and correct (MIT)
2. Check all source files have proper license headers
3. Review third-party dependencies and licenses
4. Ensure NOTICE file if required by dependencies
5. Verify no GPL/AGPL code included
6. Check attribution requirements
7. Update copyright years if needed
8. Review CONTRIBUTING.md for legal clarity

Run:
- go-licenses check (if available)
- Review go.mod dependencies
- Check for embedded code


#### Release prep: CI/CD review

- Legacy ID: `pe-41`
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

- Legacy ID: `pe-42`
- Type: `task`

**Scope**

Create migration guide if there are breaking changes.

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

Create docs/MIGRATION.md or docs/UPGRADING.md


#### Release prep: Performance benchmarks

- Legacy ID: `pe-43`
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


#### Fix dependency security vulnerabilities

- Legacy ID: `pe-44`
- Type: `task`

**Scope**

CRITICAL: Pre-commit hook detected security vulnerabilities in dependencies.

Pre-commit output:
❌ Security vulnerabilities detected in dependencies!
Run 'govulncheck ./...' for details

Actions:
1. Run govulncheck ./... to identify vulnerable packages
2. Review vulnerability details
3. Update affected dependencies
4. Test that updates don't break functionality
5. Re-run security check
6. Document any vulnerabilities that can't be fixed

This is BLOCKING for release.

Related: pe-39 (security review)

**Notes**

govulncheck results: NO CODE VULNERABILITIES

Output:
'Your code is affected by 0 vulnerabilities.'

Details:
- 0 vulnerabilities in packages we import
- 2 vulnerabilities in modules we require BUT code doesn't call them
- These are transitive dependencies not actually used

Action: Downgrade from P0 to P2
This is not blocking - just needs documentation and monitoring.

Pre-commit hook may be too strict - consider adjusting threshold.


#### Extend GenerateOptions for provider-specific options

- Legacy ID: `pe-46`
- Type: `feature`

**Scope**

NotebookLM audit 2026-04-30 found that direct llm.Provider.Generate calls carry only internal/llm/provider.go GenerateOptions. The evaluator configuration path preserves provider-specific parameters through a map, but direct Generate calls drop options such as Ollama seed and raw.

Evidence:
- internal/llm/provider.go
- internal/llm/configured_provider.go

Add a minimal provider-specific extension path for direct generation without widening provider APIs unnecessarily.

**Acceptance Criteria**

- Direct Generate can pass provider-specific options needed by existing providers, including Ollama seed and raw.
- Existing evaluator configuration behavior remains unchanged.
- Regression tests cover direct Generate option passthrough.
- Package docs describe which options are portable and which are provider-specific.

**Notes**

Created from NotebookLM design audit notebook 50b86925-3d62-4d09-bd1a-59d7ced9a523 conversation 3a1136a7-93aa-4b91-a8df-9966ffc88a5c.


#### Honor Promptfoo assertion provider override

- Legacy ID: `pe-47`
- Type: `bug`

**Scope**

NotebookLM audit 2026-04-30 found that internal/promptfoo/types.go records Assertion.Provider, but internal/promptfoo/evaluation/evaluator/assertions.go binds AssertionEvaluator to a single llm.Provider. evaluateLLMJudge always uses the bound provider, so assertion-level provider overrides are parsed but not honored.

Add assertion-level judge provider resolution or reject unsupported overrides explicitly.

**Acceptance Criteria**

- Promptfoo assertions with provider set route LLM judge calls through that provider.
- Missing or invalid assertion providers return clear errors.
- Existing default judge provider behavior remains unchanged when provider is unset.
- Regression tests cover assertion-level provider selection.

**Notes**

Created from NotebookLM design audit notebook 50b86925-3d62-4d09-bd1a-59d7ced9a523 conversation 3a1136a7-93aa-4b91-a8df-9966ffc88a5c.


### P3

#### Review .gitignore for PE project

- Legacy ID: `pe-15`
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

- Legacy ID: `pe-21`
- Type: `feature`

**Scope**

Complete the module registry system for sharing and discovering prompts.

Current status (from README):
- Core module management works (mod init/tidy/vendor)
- Registry features in next-experimental branch
- Local registry exists at ~/.pe/registry

Tasks:
1. Implement remote registry support
2. Add mod publish command
3. Add mod search with filtering
4. Add mod download from remote sources
5. Add versioning and dependency resolution
6. Create registry server (optional)
7. Documentation and examples

Related:
- internal/pemod/ has module code
- docs/MODULE_REGISTRY.md may have design docs


#### Add advanced assertion types

- Legacy ID: `pe-22`
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


#### Add interactive REPL mode

- Legacy ID: `pe-23`
- Type: `feature`

**Scope**

Implement interactive REPL (Read-Eval-Print Loop) for prompt experimentation.

Features needed:
1. Interactive prompt input
2. Live execution with provider selection
3. History and editing (readline support)
4. Variable management
5. Context preservation across runs
6. Save session to file
7. Multiline input support

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

- Legacy ID: `pe-24`
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


#### Add examples for new features

- Legacy ID: `pe-27`
- Type: `task`

**Scope**

Create comprehensive examples for newer PE features.

Current state:
- example/ directory has many examples
- examples/ directory also exists
- May need to consolidate or organize better

Need examples for:
1. Semantic backpropagation (pe semantic backprop)
2. GASO optimization (pe semantic gaso)
3. Pass@N evaluation
4. Structured output validation
5. Security testing (pe security)
6. Distributed execution
7. Cryptographic attestation
8. Starlark extensions
9. Advanced metrics (BERTScore, G-Eval)
10. Prompt composition

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
- [ ] Map all usages of `llm.Provider` in the codebase
- [ ] Map all usages of `inference.Provider` in the codebase
- [ ] Document which commands use which interface
- [ ] Identify all provider implementations (OpenAI, Anthropic, cgpt, etc.)
- [ ] Analyze migration complexity for each usage
- [ ] Create compatibility matrix for provider features
- [ ] Document breaking changes that will occur
- [ ] Review test coverage for provider-dependent code

##### Migration Preparation
- [ ] Create `internal/inference/migration.go` with LegacyAdapter
- [ ] Implement adapter for `llm.Provider` → `inference.Provider`
- [ ] Write adapter unit tests
- [ ] Create migration helpers for common patterns
- [ ] Add temporary compatibility layer
- [ ] Document migration patterns for contributors

##### Command Migration
- [ ] Update `cmd/pe/run.go` to use `inference.Provider`
- [ ] Update `cmd/pe/ask.go` to use new interface
- [ ] Update `cmd/pe/eval.go` for evaluation commands
- [ ] Update `cmd/pe/optimize.go` and optimization commands
- [ ] Update `cmd/pe/semantic.go` for semantic optimization
- [ ] Update `cmd/pe/evolve.go` for evolutionary optimization
- [ ] Update `cmd/pe/textgrad.go` for TextGrad optimization
- [ ] Update `cmd/pe/pe2.go` for PE2 optimization
- [ ] Update `cmd/pe/benchmark.go` for benchmarking
- [ ] Update `cmd/pe/test.go` for testing commands
- [ ] Update `cmd/pe/stream.go` for streaming
- [ ] Update `cmd/pe/fusion.go` for multi-model fusion

##### Provider Implementation Updates
- [ ] Update OpenAI provider to single interface
- [ ] Update Anthropic provider to single interface
- [ ] Update cgpt provider wrapper
- [ ] Update mock provider for testing
- [ ] Remove duplicate provider implementations
- [ ] Consolidate provider registration logic
- [ ] Update provider factory methods

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
- [ ] Fix `pe mod init` with proper initialization
- [ ] Implement `pe mod download` with real registry
- [ ] Complete `pe mod tidy` functionality
- [ ] Implement `pe mod vendor` properly
- [ ] Add `pe mod verify` for integrity checking
- [ ] Implement `pe mod list` for installed modules
- [ ] Add `pe mod search` for registry search
- [ ] Implement `pe mod publish` for module publishing
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
- [ ] Analyze current 48+ commands
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
- [ ] Add TextGrad optimizer tests
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
- [ ] Refactor TextGrad to use interfaces
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
- [ ] Create `internal/metrics/` directory
- [ ] Implement MetricsCollector type
- [ ] Add counter implementation
- [ ] Add histogram implementation
- [ ] Add gauge implementation
- [ ] Add summary implementation
- [ ] Create metrics registry

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
- [ ] Create `internal/tracing/` directory
- [ ] Integrate OpenTelemetry
- [ ] Add trace provider setup
- [ ] Implement span creation
- [ ] Add context propagation
- [ ] Create trace exporters
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
- [ ] Add CPU profiling
- [ ] Add memory profiling
- [ ] Add goroutine profiling
- [ ] Add block profiling
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
