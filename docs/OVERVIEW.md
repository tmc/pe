# PE: Go for Prompts - Overview

PE is the unified toolchain for prompt engineering, bringing Go's philosophy of simplicity, composability, and performance to LLM development. Just as Go transformed systems programming with its elegant toolchain, PE revolutionizes prompt engineering with a comprehensive set of tools that work together seamlessly.

## Philosophy

PE embodies the Go toolchain philosophy:

- **One Tool, Many Commands**: Like `go build`, `go test`, `go fmt`, PE provides a single binary with multiple subcommands
- **Zero Configuration**: Works out of the box with sensible defaults
- **Composability**: Unix-style pipelines for complex workflows
- **Fast and Efficient**: Written in Go for maximum performance
- **Batteries Included**: Everything you need for prompt engineering

## Core Concepts

### 1. **Prompts as Code**
PE treats prompts as first-class artifacts that can be:
- Version controlled with branches and tags
- Tested with comprehensive test suites
- Optimized using state-of-the-art methods
- Deployed with confidence

### 2. **The PE Toolchain**
```bash
pe run      # Execute prompts (like go run)
pe test     # Test prompts (like go test)
pe build    # Build optimized prompts (like go build)
pe install  # Install prompt libraries (like go get)
pe fmt      # Format prompts (like go fmt)
pe mod      # Manage dependencies (like go mod)
```

### 3. **Native Formats**

#### txtar Files
PE works natively with Go's txtar format for multi-file prompt projects:
```txtar
-- prompt.txt --
You are a helpful assistant.

-- config.yaml --
provider: gpt-4
temperature: 0.7

-- tests.yaml --
- input: "Hello"
  expect: "Hi there!"
```

#### Gist Integration
Share and run prompts directly from GitHub gists:
```bash
pe run gist:username/prompt-id
pe install gist:username/library
```

### 4. **Security First**

#### macOS Sandboxing
All PE operations run in the macOS App Sandbox by default:
- File system access restrictions
- Network isolation options
- Process containment

#### Trust Model
Only explicitly trusted binaries can be executed:
```bash
pe trust add /usr/local/bin/tool
pe trust list
pe trust revoke /usr/local/bin/tool
```

### 5. **Extensibility**

#### Plugin System
Control your inference with custom providers:
```bash
pe plugin install ollama
pe plugin create my-provider
pe run prompt.txt --provider my-provider
```

#### Style Guides
Import and compose behavioral rules:
```bash
pe import style github.com/org/style-guide
pe style create my-style --inherit formal,technical
pe run prompt.txt --style my-style
```

## Key Features

### Version Control
Git-like version control for prompts:
- Branching and merging
- Forking for variants
- Remote collaboration via gists
- Complete history and lineage tracking

### Optimization
State-of-the-art optimization methods:
- **PE2**: Meta-prompt engineering
- **TextGrad**: Semantic gradient descent
- **APEX**: Long prompt optimization
- **Evolution**: Genetic algorithms
- **Fusion**: Multi-model consensus

### Testing & Validation
Comprehensive testing capabilities:
- Property-based testing
- A/B testing with statistical analysis
- Style guide compliance
- Regression testing
- Coverage reports

### Caching
Verifiable shared caching:
- Cryptographic signatures
- Content-addressed storage
- Distributed cache protocol
- Privacy-preserving options

### Benchmarking
Declarative benchmark definitions:
- Performance analysis
- Cost optimization
- Method comparison
- Statistical significance testing

## Architecture

PE is built with a modular architecture:

```
pe (CLI)
├── Commands (run, test, build, etc.)
├── Core Engine
│   ├── Provider Interface
│   ├── Optimization Engine
│   ├── Version Control
│   └── Cache System
├── Security Layer
│   ├── Sandbox Manager
│   ├── Trust Store
│   └── Signature Verification
└── Plugin System
    ├── Provider Plugins
    ├── Style Plugins
    └── Tool Plugins
```

## Use Cases

### Development Workflow
```bash
# Initialize project
pe init my-assistant

# Develop with hot reload
pe run assistant.txt --watch

# Test changes
pe test tests/

# Optimize for production
pe optimize assistant.txt --method pe2

# Deploy
pe build --output prod/
```

### Team Collaboration
```bash
# Share via gist
pe push gist:team/assistant

# Import and customize
pe pull gist:team/assistant
pe fork assistant.txt --name my-variant

# Sync caches
pe cache sync team --verify
```

### Research & Experimentation
```bash
# Compare methods
pe benchmark methods.yaml

# A/B test variants
pe ab-test variant-a variant-b

# Analyze results
pe analyze results/ --statistical
```

## Getting Started

1. **Install PE**: `go install github.com/tmc/pe/cmd/pe@latest`
2. **Initialize a project**: `pe init`
3. **Run your first prompt**: `pe run "Hello, world!"`
4. **Explore commands**: `pe help`

See the [Getting Started Guide](GETTING_STARTED.md) for a comprehensive tutorial.

## Next Steps

- [Installation Guide](INSTALLATION.md) - Detailed setup instructions
- [Command Reference](COMMANDS.md) - Complete command documentation
- [Tutorial](TUTORIAL.md) - Step-by-step guide
- [Architecture](ARCHITECTURE.md) - Technical deep dive
- [Plugin Development](PLUGINS.md) - Extend PE with custom plugins