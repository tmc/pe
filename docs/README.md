# PE Documentation

PE is a prompt engineering toolkit inspired by the Go toolchain. This documentation reflects the current implementation status as of March 2026.

## Current Implementation Status

PE is **under active development**. Many features documented elsewhere are aspirational or incomplete. This README provides accurate information about what is currently implemented.

## Quick Start

### Installation
```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Basic Usage
```bash
# Run a simple prompt
pe run "What is 2+2?" --provider openai

# Run with template variables
pe run translate.prompt --var text="Hello" --var to="Spanish" --provider anthropic

# Initialize an evaluation config
pe init config.yaml

# Run evaluation
pe eval config.yaml

# Validate configuration
pe vet config.yaml
```

### Important: Template Syntax
PE uses Go template syntax with dots: `{{.variable}}` not `{{variable}}`
See [TEMPLATE_SYNTAX.md](TEMPLATE_SYNTAX.md) for details.

## Implemented Commands

### Core Evaluation
- **`pe eval`** - Evaluate prompts against LLM providers ✅ **IMPLEMENTED**
- **`pe view`** - View evaluation results in browser UI ✅ **IMPLEMENTED**
- **`pe vet`** - Validate promptfoo configuration files ✅ **IMPLEMENTED**
- **`pe fmt`** - Format promptfoo configuration files ✅ **IMPLEMENTED**

### Testing & Analysis
- **`pe test`** - Run advanced testing (property-based, regression) ✅ **IMPLEMENTED**
- **`pe benchmark`** - Compare performance metrics ✅ **IMPLEMENTED**
  - NEW: `--go-bench` flag outputs in Go benchmark format for compatibility with `benchstat` and other tools
- **`pe stats`** - Show quick statistics ✅ **IMPLEMENTED**
- **`pe diff`** - Compare evaluation results ✅ **IMPLEMENTED**

### Pipeline Commands (Unix-style)
- **`pe ask`** - Ask single question to LLM provider ✅ **IMPLEMENTED**
- **`pe stream`** - Process evaluation results as stream ✅ **IMPLEMENTED**
- **`pe filter`** - Filter evaluation results ✅ **IMPLEMENTED**
- **`pe analyze`** - Analyze results with statistics ✅ **IMPLEMENTED**

### Optimization & Composition
- **`pe optimize`** - Optimize prompts using metaprompting ✅ **IMPLEMENTED**
- **`pe semantic`** - Semantic gradient descent optimization ✅ **IMPLEMENTED**
- **`pe evolve`** - Evolutionary prompt optimization ✅ **IMPLEMENTED**
- **`pe experimental compose`** - Component-based prompt composition ✅ **IMPLEMENTED**

### Module System
- **`pe mod init/tidy/download/vendor`** - Go-style module management ✅ **IMPLEMENTED**
- **`pe push`** - Push modules to registry ✅ **IMPLEMENTED**

### Security & Attestation
- **`pe exp attest`** - Cryptographic attestation ⚠️ **PROTOTYPE**
- **`pe exp distributed`** - Distributed execution ⚠️ **PROTOTYPE**
- **`pe exp cache`** - Content-addressed caching ⚠️ **PROTOTYPE**
- **`pe security`** - Security testing (OWASP LLM Top 10) ✅ **IMPLEMENTED**

### Utilities
- **`pe extract`** - Extract structured data from prompts ✅ **IMPLEMENTED**
- **`pe metrics`** - Calculate evaluation metrics ✅ **IMPLEMENTED**
- **`pe template`** - Manage prompt templates ✅ **IMPLEMENTED**
- **`pe profile`** - Profiling and observability ✅ **IMPLEMENTED**
- **`pe interactive`** - Interactive REPL mode ✅ **IMPLEMENTED**
- **`pe watch`** - Watch files for changes ✅ **IMPLEMENTED**
- **`pe convert`** - Convert between formats ✅ **IMPLEMENTED**

## Documentation by Implementation Status

### ✅ Current Implementation (Accurate)
- [GETTING_STARTED.md](GETTING_STARTED.md) - Getting started guide
- [INSTALLATION.md](INSTALLATION.md) - Installation instructions
- [CLI_REFERENCE.md](CLI_REFERENCE.md) - Command reference
- [CLI_REFERENCE.md](CLI_REFERENCE.md) - CLI reference
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [MODULES.md](MODULES.md) - Module system
- [MODULE_REGISTRY.md](MODULE_REGISTRY.md) - Module registry
- [PLUGINS.md](PLUGINS.md) - Plugin system
- [ATTESTATION.md](ATTESTATION.md) - Cryptographic attestation

### ⚠️ Partially Implemented
- [OVERVIEW.md](OVERVIEW.md) - Contains both implemented and aspirational features
- [TUTORIAL.md](TUTORIAL.md) - Basic tutorial (some advanced features not implemented)
- [GETTING_STARTED.md](GETTING_STARTED.md) - Quick start guide
- [ADVANCED_FEATURES.md](ADVANCED_FEATURES.md) - Mix of implemented and planned features
- [API_REFERENCE.md](API_REFERENCE.md) - API reference (provider interfaces exist, some incomplete)
- [OPTIMIZATION_EXAMPLES.md](OPTIMIZATION_EXAMPLES.md) - Optimization examples (methods vary in completeness)
- [STARLARK_EXTENSION.md](STARLARK_EXTENSION.md) - Starlark integration (partial)
- [PROMPTFOO_INTEGRATION.md](PROMPTFOO_INTEGRATION.md) - Integration guide (basic compatibility)
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Troubleshooting guide

### 🔮 Future/Planned Features
All documents in [future/](future/) directory describe planned features that are not yet implemented:
- Comprehensive features
- World-class tooling
- Advanced optimization guides
- Complete API references
- Competitive analysis
- Research foundations
- And more...

## Key Implementation Notes

### What Actually Works
1. **Core Evaluation**: Full promptfoo-compatible evaluation system
2. **Pipeline Commands**: Unix-style commands for composability
3. **Module System**: Go mod-style dependency management with gist registry
4. **Attestation**: Cryptographic signing and verification
5. **Plugin System**: Runtime plugin discovery and execution
6. **Optimization**: Basic metaprompting techniques (PE2, TextGrad, etc.)

### Major Limitations
1. **Provider Integration**: Native providers implemented but cgpt still default for compatibility
2. **Test Coverage**: Improved to ~25% with new test files added
3. **Documentation Accuracy**: Documentation reorganized to separate current vs future features
4. **Distributed System**: Core implementation complete, CLI integration in progress
5. **Advanced Features**: Many "world-class" features are designs, not implementations

### Provider Support
Currently supported providers:
- **Native OpenAI**: Direct API integration (GPT-3.5, GPT-4, etc.)
- **Native Anthropic**: Direct API integration (Claude 3 family)
- **cgpt wrapper**: Fallback for compatibility (supports all cgpt providers)
- Google AI (Gemini) - via cgpt
- Local models - via cgpt

Native providers are fully implemented and can be used with `--provider openai` or `--provider anthropic`.

## Configuration Format

PE uses YAML configuration files compatible with promptfoo:

```yaml
prompts:
  - "What is the capital of {{country}}?"
  
providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-haiku"
  
tests:
  - vars:
      country: "France"
    assert:
      - type: "contains"
        value: "Paris"
```

## Examples

Basic examples are available in:
- [../example/](../example/) - Working examples
- [../examples/](../examples/) - Additional examples
- [future/EXAMPLES_LIBRARY.md](future/EXAMPLES_LIBRARY.md) - Comprehensive examples (planned)

## Getting Help

- **Built-in help**: `pe help [command]`
- **Issues**: [GitHub Issues](https://github.com/tmc/pe/issues)
- **Source**: [GitHub Repository](https://github.com/tmc/pe)

## Contributing

PE is under active development. The most helpful contributions:

1. **Test Coverage**: Improve test coverage from current ~15%
2. **Provider Integration**: Complete native provider implementations
3. **Documentation Accuracy**: Fix gaps between docs and implementation
4. **Core Features**: Complete partially implemented features

## Recent Improvements (February 2025)

- **Native Provider Support**: Added direct API integrations for OpenAI and Anthropic
- **Go Benchmark Format**: Added `--go-bench` flag for compatibility with Go perf tools
- **Test Coverage**: Expanded test suite across core packages (~25% coverage)
- **Distributed Systems**: Implemented consensus and distributed execution frameworks
- **Documentation**: Reorganized to clearly separate implemented vs planned features
- **Security**: Added comprehensive security architecture documentation

## Development Status

PE follows semantic versioning. Current status:
- **Version**: Pre-1.0 (under development)
- **Stability**: Core evaluation features are stable
- **API**: Subject to change during development

This documentation will be updated as features are implemented and stabilized.
