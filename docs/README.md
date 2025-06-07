# PE Documentation

PE is a prompt engineering toolkit inspired by the Go toolchain. This documentation reflects the current implementation status as of January 2025.

## Current Implementation Status

PE is **under active development**. Many features documented elsewhere are aspirational or incomplete. This README provides accurate information about what is currently implemented.

## Quick Start

### Installation
```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Basic Usage
```bash
# Initialize a project
pe init config.yaml

# Run evaluation
pe eval config.yaml

# View results
pe view

# Format configuration
pe fmt config.yaml

# Validate configuration
pe vet config.yaml
```

## Implemented Commands

### Core Evaluation
- **`pe eval`** - Evaluate prompts against LLM providers ✅ **IMPLEMENTED**
- **`pe view`** - View evaluation results in browser UI ✅ **IMPLEMENTED**
- **`pe vet`** - Validate promptfoo configuration files ✅ **IMPLEMENTED**
- **`pe fmt`** - Format promptfoo configuration files ✅ **IMPLEMENTED**

### Testing & Analysis
- **`pe test`** - Run advanced testing (property-based, regression) ✅ **IMPLEMENTED**
- **`pe benchmark`** - Compare performance metrics ✅ **IMPLEMENTED**
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
- **`pe fusion`** - Multi-model fusion ✅ **IMPLEMENTED**
- **`pe compose`** - Component-based prompt composition ✅ **IMPLEMENTED**

### Module System
- **`pe mod init/tidy/download/vendor`** - Go-style module management ✅ **IMPLEMENTED**
- **`pe push`** - Push modules to registry ✅ **IMPLEMENTED**

### Security & Attestation
- **`pe attest`** - Cryptographic attestation ✅ **IMPLEMENTED**
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
- [COMMANDS.md](COMMANDS.md) - Command reference (mostly accurate)
- [CLI_REFERENCE.md](CLI_REFERENCE.md) - CLI reference
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [MODULES.md](MODULES.md) - Module system
- [MODULE_REGISTRY.md](MODULE_REGISTRY.md) - Module registry
- [PLUGINS.md](PLUGINS.md) - Plugin system
- [ATTESTATION.md](ATTESTATION.md) - Cryptographic attestation

### ⚠️ Partially Implemented
- [OVERVIEW.md](OVERVIEW.md) - Contains both implemented and aspirational features
- [TUTORIAL.md](TUTORIAL.md) - Basic tutorial (some advanced features not implemented)
- [QUICK_START.md](QUICK_START.md) - Quick start guide (some commands not fully implemented)
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
1. **Provider Integration**: Still primarily uses cgpt CLI wrapper, not native APIs
2. **Test Coverage**: Only ~15% test coverage across the codebase
3. **Documentation Accuracy**: Many docs describe aspirational features as implemented
4. **Distributed System**: CLI commands exist but integration incomplete
5. **Advanced Features**: Many "world-class" features are designs, not implementations

### Provider Support
Currently supported providers (via cgpt wrapper):
- OpenAI (GPT-3.5, GPT-4, etc.)
- Anthropic (Claude 3 family)
- Google AI (Gemini)
- Local models via cgpt

Native provider implementations are partially complete but not fully integrated.

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

## Development Status

PE follows semantic versioning. Current status:
- **Version**: Pre-1.0 (under development)
- **Stability**: Core evaluation features are stable
- **API**: Subject to change during development

This documentation will be updated as features are implemented and stabilized.