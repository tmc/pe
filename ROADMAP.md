# PE Roadmap

Last updated: 2026-04-30

This file is the source of truth for planned PE work. Beads is deprecated for this repository: do not create or update `.beads` issues for new work. Keep roadmap changes in tracked commits with the code or documentation they describe.

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
- Test coverage claims inconsistent (25% vs 40%)
- Documentation dates are inconsistent across older 2025 snapshots and current 2026 status docs
- Multiple overlapping getting started docs
- Future docs mixed with current implementation docs
- Legacy TODO material now points at this roadmap, but current documentation still needs an accuracy pass
- Recent work not documented (scripttest, fixes, etc.)

Sub-tasks to create:
1. Reconcile test coverage numbers
2. Update all documentation dates
3. Consolidate getting started documentation
4. Update docs/CURRENT_STATUS.md with latest
5. Document recent scripttest work
6. Update RELEASE_NOTES.md
7. Clean up docs/future/ organization
8. Fix README.md accuracy
9. Review all command documentation

Priority: P1 - Blocking release


#### Release prep: Examples validation

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

- Type: `task`

**Scope**

Prepare version number and comprehensive changelog for release.

Tasks:
1. Verify v0.5.0 version consistency across source, release notes, tags, and docs
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
- Scripttest framework fixes
- Test improvements
- Documentation additions (MARKERS.md)
- .gitignore improvements


#### Release prep: README consolidation

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

- Type: `task`

**Scope**

Security review before release.

Tasks:
1. Review for hardcoded secrets or API keys
2. Check for command injection vulnerabilities
3. Verify input validation (already have tests)
4. Review file path traversal protections
5. Check dependency vulnerabilities with govulncheck and document any non-called transitive findings
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

Related: documented scripttest pipe limitations


#### Maintain roadmap with current status

- Type: `task`

**Scope**

The old TODO file is superseded. Keep this roadmap current as the tracked source of truth.

Need to:
1. Review completed items and mark appropriately
2. Add new work items to this roadmap
3. Update project status
4. Update references to point at ROADMAP.md
5. Remove obsolete TODO and Beads references as they are found


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

Tasks:
1. Document Ollama configuration and provider selection in README and docs.
2. Add examples for common local model workflows.
3. Add end-to-end validation notes for representative Ollama models.

Benefits:
- Local model support (no API keys needed)
- Privacy-focused option
- Cost-effective for development
- Support for open models (llama, mistral, etc.)


#### Improve test coverage

- Type: `task`

**Scope**

Increase test coverage across the project.

Current coverage needs a fresh measured baseline; older roadmap text and docs
disagree on 25% versus 40% overall coverage.

Priority areas:
1. Refresh package coverage data with `go test -cover ./...`.
2. Expand tests in lower-coverage packages, including internal/cli, internal/module, internal/optimization, internal/plugin, and internal/templates.
3. Add command-level integration tests for paths that currently rely on package-level tests only.

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


#### Add examples for new features

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
