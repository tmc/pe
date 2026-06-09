# PE Current Implementation Status

Last Updated: 2026-05-11

## Overview

PE is a prompt engineering toolkit with a generated top-level command surface. This
document summarizes the current implementation; [../ROADMAP.md](../ROADMAP.md)
is the source of truth for planned work and release-blocking follow-ups.

## Working Features ✅

### Core Commands (Fully Functional)
- `pe run` - Execute prompts with variable substitution
- `pe run-text` - Validate and render executable, templated, composable text
  without invoking providers, tools, shell commands, or network access
- `pe ask` - Pipeline-based prompt execution
- `pe eval` - Comprehensive evaluation framework
- `pe view` - Browser-based result viewer
- `pe vet` - Configuration validation (fixed stack overflow)
- `pe fmt` - Format configuration files

### Pipeline Commands (Unix-style, All Working)
- `pe stream` - Stream processing
- `pe filter` - Filter and transform outputs
- `pe extract` - Extract structured data from responses
- `pe analyze` - Statistical analysis
- `pe collect` - Collect async results
- `pe reduce` - Aggregate results
- `pe stats` - Quick statistics

### Module System (Functional)
- `pe mod init` - Initialize modules
- `pe mod download` - Download from registry (fixed)
- `pe mod tidy` - Clean dependencies
- `pe mod vendor` - Vendor dependencies
- `pe mod list` - List available modules
- `pe mod get` - Get module information
- `pe mod search` - Search modules
- `pe mod vet` - Validate `pe.mod` capability policy and executable text
  metadata

### Template System (Functional)
- `pe template list` - List templates
- `pe template apply` - Apply templates
- `pe template create` - Create templates
- `pe template interactive` - Interactive mode; library APIs can create
  starter templates non-interactively
- `pe template show` - Show template details
- `pe template validate` - Validate templates

### Provider Support
- ✅ **OpenAI** - Native implementation; measured coverage is tracked in [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md)
- ✅ **Anthropic** - Native implementation; measured coverage is tracked in [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md)
- ✅ **cgpt** - CLI wrapper (newly registered and working)
- ✅ **mock** - For testing

### Optimization Commands
- `pe optimize` - Multiple optimization methods
- `pe semantic` - Semantic backpropagation
- `pe evolve` - Evolutionary optimization
- `pe compose` - Component composition with local coherence checks
- Local metaprompt program synthesis strategies for template, evolutionary, and
  neural-style prompt programs

### Local Workflow Prototypes
- `pe exp distributed` - Bounded local task scheduler from JSON task files
- `pe exp consensus` - Weighted local vote aggregation
- `pe exp attest` - Unsigned local SHA-256 manifest generation and verification
- `pe exp cache` - Local content-addressed cache helpers
- `pe serve` - Localhost-first HTTP API command
- `pe playground` - Local web UI with local compare, metrics, security,
  component, and in-memory history endpoints

### Executable Text
- Plain text is valid executable text input.
- `pe.text.v1` and `pe.workflow.v1` front matter can declare typed inputs,
  metadata, safety policy, placement, and local imports.
- `pe run-text` renders Go templates from explicit `--var` values.
- `{{ import "name" }}` composes local front matter imports safely.
- `pe.mod` now parses `capability`, `placement`, and `policy` blocks.
- `pe mod vet` checks denied provider/tool/data/prompt policy and typed-input
  requirements for executable text files. Strict composition also checks cached
  dependency `pe.mod` files so dependencies cannot request capabilities or
  network placement denied by the parent module.
- `pe run` enforces `pe.mod` provider denials for exact provider names and known
  provider classes such as `remote` before provider construction. Placement
  policy also blocks known remote providers when `network false` is set.
- Compose output, component library, and import writes honor `tools deny write`
  in `pe.mod`; terminal composition output remains read-only.
- `pe build` artifact writes honor `tools deny write`; `pe build --validate`
  remains read-only.
- `pe eval -o` and `pe eval --save-db` honor `tools deny write`; default eval
  stdout output remains read-only.
- `pe benchmark -o` honors `tools deny write`; stdout and Go benchmark output
  remain read-only.
- `pe metrics -o` honors `tools deny write`; stdout metrics output remains
  read-only.
- `pe expand --output` honors `tools deny write`; stdout expanded JSON remains
  read-only.
- `pe extract --output` honors `tools deny write`; stdout extracted content
  remains read-only.
- `pe evolve --output` honors `tools deny write`; terminal evolution output
  remains read-only.
- Security report output files honor `tools deny write`; terminal security
  output remains read-only.
- `pe test --output` and `pe test create-suite --output` honor `tools deny
  write`; terminal test output remains read-only.
- `pe structured` file outputs honor `tools deny write`; terminal structured
  output remains read-only.
- Template file outputs honor `tools deny write`; terminal template output
  remains read-only.
- Workspace writes to `pe.work` honor `tools deny write`; read-only workspace
  commands remain allowed.
- `pe fmt --write` honors `tools deny write`; stdout formatting and `--check`
  remain read-only.
- `pe init` honors `tools deny write` before creating `.pe/` project files.
- `pe mod init --force` honors an existing `pe.mod` `tools deny write` policy
  before overwriting the module file; initial `pe mod init` has no module
  policy file to consult.
- `pe mod tidy --write` honors `tools deny write` before rewriting `pe.mod`;
  default tidy reporting remains read-only.
- `pe mod vendor` honors `tools deny write` before creating or rewriting the
  `vendor/` dependency tree.
- `pe mod upgrade` honors `tools deny write` before rewriting `pe.mod` for
  newer dependency versions.
- `pe edit` file edits and `--module` dependency edits honor `tools deny write`;
  `--print` and `--json` output remain read-only.
- `pe prompt init`, `pe prompt edit`, and `pe prompt tidy --remove-unused`
  honor `tools deny write`; prompt info/help and validation-only tidy remain
  read-only.
- `pe convert` honors `tools deny write` before writing converted config files.
  Legacy promptfoo config fmt/init helpers also check the policy before writes.
- Interactive REPL `:save` honors `tools deny write` before writing session
  files.
- `pe plugin build` honors `tools deny write` before writing plugin artifacts.
- `pe exp consensus --output` and `pe exp optimize --output` honor
  `tools deny write` before writing result files; stdout output remains
  read-only.
- `pe profile` output directories, trace files, metrics files, and report files
  honor `tools deny write`; status and terminal reports remain read-only.

## Important Implementation Details

### Template Syntax
**CRITICAL**: PE uses Go template syntax with dots:
- ✅ Correct: `{{.variable}}`
- ❌ Wrong: `{{variable}}`

All prompts must use the dot notation for variables.

### Provider Configuration
Providers are specified as `provider:model`:
```bash
pe eval config.yaml  # Uses providers from config
pe ask "prompt" --provider openai:gpt-4
pe ask "prompt" --provider anthropic:claude-3-haiku-20240307
pe ask "prompt" --provider cgpt  # Uses cgpt CLI
```

### Module Registry
- Local registry at `~/.pe/registry/`
- HTTP and GitHub registry backends exist, but still need release validation
- Download functionality actually works (not mock)
- Publishing remains limited: local registry publishing works, gist-backed
  `pemod` publishing updates the registry index, and hosted GitHub/HTTP module
  publishing still needs a product boundary decision.

## Recent Release-Prep Fixes

1. **Executable Text** - Added `pe run-text`, front matter parsing, explicit
   template inputs, and local import composition.
2. **pe.mod Capabilities** - Added parser support and `pe mod vet` static
   validation for module capability policy.
3. **Release Legal Readiness** - Added MIT `LICENSE`, `docs/MIGRATION.md`, and
   `docs/NOTICE_DECISION.md`.
4. **Release Build Matrix** - Recorded passing local cross-compilation results
   for darwin/arm64, darwin/amd64, linux/amd64, linux/arm64, and windows/amd64.
5. **pe vet Stack Overflow** - Fixed recursive command execution.
6. **pe view JSON Schema** - Confirmed working with proper format.
7. **Provider Configuration** - Fixed provider:model parsing.
8. **Module Registry** - Created and populated with samples.
9. **Module Download** - Implemented actual downloading.
10. **Template Interactive** - Added guided selection and input.
11. **cgpt Provider** - Registered and implemented.

## Partially Implemented ⚠️

### Advanced Assertions
Basic assertions work. Local deterministic baselines now cover lexical
similarity, SQL shape checks, required structure markers, toxicity term
matching, keyword classification, coherence transitions/repetition, and
required factuality facts. These remain incomplete:
- provider-backed semantic similarity, toxicity, coherence, factuality, and
  classification
- SQL parsing and rich structure validation

### Structured Validation
JSON and YAML validation are local. TypeScript and Pydantic formatter plugins
can generate schema shapes and validate JSON object data against PE schemas, but
they do not execute TypeScript or Python runtimes or parse source files.

### Security Analysis
OWASP-oriented security checks include local response-pattern analyses for
prompt injection, insecure output handling, sensitive disclosure, data
poisoning, supply-chain indicators, denial of service, plugin design,
excessive agency, overreliance, and model theft. Remote provenance checks,
model-assisted adjudication, and calibration benchmarks remain future work.

## Not Implemented ❌

### Planned Features
- Multi-modal support (vision, audio)
- IDE integrations
- Visual prompt engineering tools
- Neurosymbolic synthesis
- Advanced caching strategies

## Testing Coverage

The measured coverage baseline is tracked in
[TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md):

- Current measured baseline is tracked in the coverage report and roadmap.

Use the coverage report instead of older approximate coverage claims.

## Configuration Examples

### Basic Evaluation
```yaml
prompts:
  - "Translate {{.text}} to {{.language}}"
  
providers:
  - openai:gpt-3.5-turbo
  - anthropic:claude-3-haiku-20240307
  
tests:
  - vars:
      text: "Hello"
      language: "Spanish"
    assert:
      - type: contains
        value: "Hola"
```

### Using cgpt Provider
```yaml
providers:
  - cgpt  # Uses default cgpt configuration
```

## Command Examples

```bash
# Run with variable substitution (note the dot syntax)
pe run prompt.txt --var name=World  # prompt must use {{.name}}

# Evaluate with multiple providers
pe eval config.yaml -o results.json

# View results in browser
pe view -f results.json

# Module operations
pe mod init myproject
pe mod download  # Actually downloads, not mock

# Template interactive mode
pe template interactive  # Now fully functional

# Use cgpt provider
pe ask "What is 2+2?" --provider cgpt
```

## Known Limitations

1. **Template Syntax**: Only Go templates with dots work
2. **Provider Defaults**: Need API keys in environment
3. **Module Registry**: Remote registry behavior still needs release validation
4. **Advanced Features**: Many documented features are designs
5. **Documentation**: Some docs describe planned features

## Getting Help

- Check this document first for accurate status
- Use `pe [command] --help` for command-specific help
- File issues for bugs or missing features
- Don't trust all documentation - some is aspirational

## Summary

PE has a solid core with a generated top-level command surface. The evaluation
framework, pipeline commands, and module system all work. The main release risk
is documentation and example drift: some older documents still mix current
behavior with planned features tracked in [../ROADMAP.md](../ROADMAP.md).
