# PE Documentation

Welcome to the PE (Go for Prompts) documentation. PE is the unified toolchain for prompt engineering, bringing Go's philosophy of simplicity, composability, and performance to LLM development.

## Quick Links

- **[Overview](OVERVIEW.md)** - Understand PE's philosophy and capabilities
- **[Installation](INSTALLATION.md)** - Get PE running on your system
- **[Tutorial](TUTORIAL.md)** - Step-by-step guide from basics to advanced
- **[Commands](COMMANDS.md)** - Complete reference for all PE commands
- **[Architecture](ARCHITECTURE.md)** - Technical deep dive into PE's design
- **[Plugins](PLUGINS.md)** - Extend PE with custom functionality

## Getting Started

```bash
# Install PE
go install github.com/tmc/pe/cmd/pe@latest

# Initialize a project
pe init

# Run your first prompt
pe run "Hello, PE!"

# Explore help
pe help
```

## Documentation Guide

### For New Users

1. Start with the **[Overview](OVERVIEW.md)** to understand PE's approach
2. Follow the **[Installation Guide](INSTALLATION.md)** to set up PE
3. Work through the **[Tutorial](TUTORIAL.md)** for hands-on learning
4. Reference the **[Command Reference](COMMANDS.md)** as needed

### For Developers

1. Review the **[Architecture Guide](ARCHITECTURE.md)** for system design
2. Learn **[Plugin Development](PLUGINS.md)** to extend PE
3. Check the **[API Reference](../pkg/)** for programmatic usage
4. See **[Contributing Guidelines](../CONTRIBUTING.md)** to contribute

### For Teams

1. Learn about **[Shared Caching](TUTORIAL.md#part-7-team-collaboration)** for collaboration
2. Set up **[Style Guides](COMMANDS.md#style--composition-commands)** for consistency
3. Configure **[Security Policies](INSTALLATION.md#security-model)** for your organization
4. Implement **[CI/CD Integration](TUTORIAL.md#continuous-integration)** for automation

## Core Concepts

### The PE Toolchain

PE provides a complete toolchain similar to Go:

| PE Command | Go Equivalent | Purpose |
|------------|---------------|---------|
| `pe run` | `go run` | Execute prompts immediately |
| `pe test` | `go test` | Test prompts with assertions |
| `pe build` | `go build` | Build optimized prompts |
| `pe install` | `go get` | Install prompt libraries |
| `pe fmt` | `go fmt` | Format prompt files |
| `pe mod` | `go mod` | Manage dependencies |

### Key Features

- **🔧 Unified Toolchain**: One tool for all prompt engineering needs
- **🚀 Performance**: Written in Go for maximum speed and efficiency
- **🔒 Security First**: macOS sandboxing and trusted execution
- **🔌 Extensible**: Plugin system for custom providers and tools
- **📦 Native Formats**: Works with txtar files and GitHub gists
- **🔄 Version Control**: Git-like branching and history
- **🧬 Advanced Optimization**: State-of-the-art methods (PE2, TextGrad, APEX)
- **📊 Comprehensive Testing**: Property-based, A/B, and regression testing
- **💾 Verifiable Caching**: Cryptographically signed shared caches
- **📈 Benchmarking**: Performance and cost analysis tools

## Example Workflows

### Basic Development

```bash
# Create and test a prompt
echo "You are a helpful assistant" > assistant.txt
pe test assistant.txt --assert "polite"

# Optimize it
pe optimize assistant.txt --method pe2

# Build for production
pe build assistant.txt --output prod/
```

### Team Collaboration

```bash
# Share via gist
pe push gist:team/assistant

# Import and customize
pe pull gist:team/assistant
pe fork assistant.txt --name my-variant

# Share cache
pe cache export --sign > team-cache.tar
```

### Advanced Optimization

```bash
# Multi-stage optimization
pe compose components/ | \
  pe optimize --method textgrad | \
  pe optimize --method evolve | \
  pe test --comprehensive
```

## Community Resources

- **GitHub**: [github.com/tmc/pe](https://github.com/tmc/pe)
- **Discord**: [Join our community](https://discord.gg/pe-prompts)
- **Examples**: [Example projects](../examples/)
- **Blog**: [PE Blog](https://pe.dev/blog)

## Documentation Versions

This documentation is for PE v1.0. For other versions:

- [Latest](https://docs.pe.dev/latest)
- [v1.0](https://docs.pe.dev/v1.0) (current)
- [Development](https://docs.pe.dev/dev)

## Contributing to Docs

We welcome documentation improvements! To contribute:

1. Fork the repository
2. Make your changes in the `docs/` directory
3. Submit a pull request

See [Contributing Guidelines](../CONTRIBUTING.md) for details.

## License

PE is open source under the MIT License. See [LICENSE](../LICENSE) for details.