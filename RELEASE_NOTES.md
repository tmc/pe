# PE Release Notes - Initial Public Release

## Branch: `next` 

This branch contains the stable, production-ready features of PE (Prompt Engineering toolkit).

## ✅ Production-Ready Features

### Core Functionality
- **47 CLI Commands**: Full suite of composable Unix-style commands
- **Native Provider Support**: 
  - OpenAI provider (74% test coverage)
  - Anthropic provider (73.3% test coverage)
  - cgpt CLI integration (updated for v0.4.4)
- **Evaluation Framework**: 
  - Pass@N metrics
  - 15+ assertion types (contains, regex, JSON, length, etc.)
  - Comprehensive evaluation reports
- **Module Management**:
  - `pe mod init` - Initialize modules
  - `pe mod tidy` - Clean up dependencies
  - `pe mod vendor` - Vendor dependencies

### Advanced Features
- **Optimization Methods**:
  - PE2 optimization
  - APEX method
  - Multistage optimization
  - Reflection-based improvement
  - Evolutionary algorithms (NSGA-II)
- **Security Testing**: Full OWASP LLM Top 10 coverage
- **Cryptographic Attestation**: Sign and verify prompt runs
- **Structured Output**: JSON/YAML schema validation
- **Pipeline Support**: Unix-style composability

## ⚠️ Features Moved to `next-experimental` Branch

The following features are under development and have been moved to the `next-experimental` branch:

### Module Registry (Pending Implementation)
- `pe mod list` - List registry modules
- `pe mod get` - Download specific modules  
- `pe mod download` - Batch download dependencies
- Registry URLs (registry.pe.dev) are placeholders

### Assertion Types (Placeholders)
- Toxicity detection
- Coherence evaluation
- Factuality checking
- Classification
- Similarity scoring
- SQL validation
- Document structure validation

These assertions will return informative error messages indicating they're not yet implemented.

### Other Incomplete Features
- Interactive REPL mode
- Playground web interface
- Some statistical functions (Shapiro-Wilk test)
- Distributed execution edge cases

## Migration Guide

If you were using any of the experimental features:

1. **Module Registry**: Use local file paths or Git repositories for now
2. **Advanced Assertions**: Stick to the implemented assertion types (contains, regex, JSON, etc.)
3. **Interactive Mode**: Use the standard CLI commands instead

## Known Limitations

- Some script tests may fail (does not affect core functionality)
- Module registry uses placeholder configuration
- Some caching features are incomplete

## Testing

Core functionality has been thoroughly tested:
- Unit test coverage: ~40% overall
- Provider coverage: 70%+
- Build system: Fully functional
- Core commands: Production ready

## Next Steps

The `next-experimental` branch contains ongoing work on:
- Full module registry implementation
- Advanced assertion types with ML integration
- Interactive development environment
- Extended statistical analysis

For production use, stick with the `next` branch. For bleeding-edge features and contributing to development, check out `next-experimental`.

## Support

Report issues at: https://github.com/tmc/pe/issues

---

*Generated: September 2025*
*Branch: next (180+ commits ahead of master)*
*Repository size: ~17MB (optimized from 307MB)*