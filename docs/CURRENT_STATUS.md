# PE Current Implementation Status

Last Updated: 2026-05-05

## Overview

PE is a prompt engineering toolkit with a generated top-level command surface. This
document summarizes the current implementation; [../ROADMAP.md](../ROADMAP.md)
is the source of truth for planned work and release-blocking follow-ups.

## Working Features ✅

### Core Commands (Fully Functional)
- `pe run` - Execute prompts with variable substitution
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

### Template System (Functional)
- `pe template list` - List templates
- `pe template apply` - Apply templates
- `pe template create` - Create templates
- `pe template interactive` - Interactive mode (newly implemented)
- `pe template show` - Show template details
- `pe template validate` - Validate templates

### Provider Support
- ✅ **OpenAI** - Native implementation (74% test coverage)
- ✅ **Anthropic** - Native implementation (73.3% test coverage)
- ✅ **cgpt** - CLI wrapper (newly registered and working)
- ✅ **mock** - For testing

### Optimization Commands
- `pe experimental optimize` - Multiple optimization methods
- `pe experimental semantic` - Semantic backpropagation
- `pe experimental evolve` - Evolutionary optimization
- `pe experimental compose` - Component composition in experimental command group

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
- Sample modules included (greeting, math)
- Download functionality actually works (not mock)

## Recent Release-Prep Fixes

1. **pe vet Stack Overflow** - Fixed recursive command execution
2. **pe view JSON Schema** - Confirmed working with proper format
3. **Provider Configuration** - Fixed provider:model parsing
4. **Module Registry** - Created and populated with samples
5. **Module Download** - Implemented actual downloading
6. **Template Interactive** - Added guided selection and input
7. **cgpt Provider** - Registered and implemented

## Partially Implemented ⚠️

### Advanced Assertions
Basic assertions work, but these are incomplete:
- toxicity detection
- coherence scoring
- factuality checking
- similarity metrics

### Distributed Execution
Core implementation exists but CLI integration incomplete:
- `pe exp distributed --help` - Prototype command is present; operational subcommands are not yet stable

### REST API Server
- Code exists but not exposed via CLI

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

- Overall statement coverage: 40.8%
- Native OpenAI provider: 74.0%
- Native Anthropic provider: 73.3%
- `cmd/pe`: 30.0%

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
3. **Module Registry**: Only local registry, no remote yet
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
