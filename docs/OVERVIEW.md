# PE: Safe Prompting Toolchain - Overview

PE is a safe prompting toolchain. It uses simple commands and explicit files for
prompt development, testing, module management, and local workflow prototypes.

The primary artifact is executable, templated, composable text. Plain text is
valid by default. Files can add typed inputs, metadata, safety policy, and
placement rules when they need stronger contracts.

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
- Optimized using PE optimization commands
- Deployed with confidence

### 2. **The PE Toolchain** (✅ Core + ⚠️ Prototype Commands)
```bash
pe run      # Execute prompts with native OpenAI/Anthropic providers
pe eval     # Comprehensive evaluation with 20+ assertion types
pe test     # Advanced testing (property-based, regression)
pe optimize # Metaprompting optimization (PE2, TextGrad, GASO)
pe semantic # Semantic backpropagation (2025 research)
pe fmt      # Format prompts and configurations
pe mod      # Complete module management (init/download/tidy/vendor)
pe security # OWASP LLM Top 10 security testing
pe exp attest --help      # Prototype command group
pe exp distributed --help # Prototype command group
```

### 3. **Native Formats** (✅ Implemented)

#### Prompt Files with Embedded Evaluations
PE supports self-documenting prompt files with embedded tests:
```
Solve this math problem: {{EXPRESSION}}

-- prompt-summary --
A math solver that explains step-by-step solutions.

-- variable-description/EXPRESSION --
A mathematical expression to solve (e.g., "2+2", "15*3")

-- evals --
EXPRESSION: "15 + 27"
expected: "42"
```

#### YAML Configuration
Comprehensive evaluation configurations:
```yaml
prompts:
  - file: "math-solver.txt"
providers:
  - id: "openai:gpt-4"
  - id: "anthropic:claude-3-haiku"
tests:
  - vars:
      EXPRESSION: "15 + 27"
    assert:
      - type: contains
        value: "42"
      - type: llm-rubric
        value: "Shows clear step-by-step reasoning"
```

### 4. **Security & Attestation** (✅ Security Stable, ⚠️ Attestation/Cache Prototype)

#### OWASP LLM Top 10 Security Testing
Comprehensive security testing for LLM applications:
```bash
pe security scan --prompt task.txt
pe security test --config security-tests.yaml
```

#### Cryptographic Attestation (Prototype CLI Surface)
Inspect the currently exposed prototype interface:
```bash
pe exp attest --help
```

#### Secure Caching (Prototype CLI Surface)
Inspect the currently exposed prototype interface:
```bash
pe exp cache --help
```

### 5. **Advanced Optimization** (✅ Implemented)

#### Semantic Backpropagation (2025 Research)
Implements GASO (Graph-based Agentic System Optimization) commands:
```bash
pe semantic backprop --prompt task.txt --target accuracy
pe semantic gaso --system definition.json --multi-objective
```

#### Multiple Optimization Methods
Prompt optimization commands:
```bash
pe optimize --method pe2 --iterations 10
pe optimize --method textgrad --learning-rate 0.1
pe evolve --algorithm nsga-ii --generations 20
```

### 6. **Extensibility** (✅ Implemented)

#### Plugin System
Runtime plugin discovery and management:
```bash
pe plugin list
pe plugin install plugin-name
# Automatic discovery of pe-* executables in PATH
```

#### Native Provider Integration
Fully implemented native providers; see [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md) for measured coverage.
```bash
pe run prompt.txt --provider openai:gpt-4
pe run prompt.txt --provider anthropic:claude-3-haiku
```

## Key Features

### Advanced Optimization (✅ Implemented)
Optimization methods:
- **PE2**: Meta-prompt engineering optimization
- **TextGrad**: Natural language gradient descent
- **Semantic Backpropagation**: 2025 GASO research implementation
- **APEX**: Advanced prompt optimization
- **Evolutionary Algorithms**: NSGA-II multi-objective optimization
- **Experimental Composition**: `pe experimental compose` for component composition

### Comprehensive Evaluation (✅ Implemented)
Testing and validation:
- **20+ Assertion Types**: includes pass@n, structured output, LLM rubrics
- **Property-based Testing**: Automated test generation
- **Regression Testing**: Detect performance degradation
- **Statistical Analysis**: Advanced metrics (BLEU, ROUGE, BERTScore, G-Eval)
- **Security Testing**: OWASP-oriented checks; incomplete analyses fail closed

### Infrastructure (Mixed Maturity)
Infrastructure:
- **Cryptographic Attestation**: Prototype CLI entrypoint (`pe exp attest`)
- **Content-Addressed Caching**: Prototype CLI entrypoint (`pe exp cache`)
- **Distributed Execution**: Prototype CLI entrypoint (`pe exp distributed`)
- **Native Providers**: OpenAI and Anthropic; measured coverage is tracked in [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md)
- **Module System**: Complete dependency management (init/download/tidy/vendor)

### Performance & Monitoring (✅ Implemented)
Comprehensive observability:
- **Benchmarking**: benchmark helpers; command-level significance workflows are incomplete
- **Profiling**: CPU, memory, and execution tracing
- **Metrics**: Advanced evaluation metrics and cost optimization
- **Pipeline Processing**: Unix-style composable commands for complex workflows

## Architecture (✅ Implemented)

PE is built with a modular, production-ready architecture:

```
pe (CLI - 47 Commands)
├── Core Commands (run, eval, optimize, semantic, test)
├── Pipeline Commands (ask, stream, filter, analyze, collect, reduce)
├── Core Engine
│   ├── Native Provider Interface (coverage tracked in TEST_COVERAGE_REPORT.md)
│   ├── Metaprompting Engine (PE2, TextGrad, GASO, APEX)
│   ├── Evaluation System (20+ assertion types, pass@n)
│   └── Distributed System (P2P networking, consensus)
├── Security & Attestation
│   ├── OWASP LLM Top 10 Testing
│   ├── Cryptographic Signing
│   └── Content Verification
├── Advanced Features
│   ├── Caching System (content-addressed, verifiable)
│   ├── Module Management (Go-style mod commands)
│   └── Performance Profiling
└── Plugin System (Runtime discovery of pe-* executables)
```

## Use Cases

### Development Workflow (✅ Working Examples)
```bash
# Initialize project with module support
pe init my-project
pe mod init

# Run prompts with native providers
pe run assistant.txt --provider openai:gpt-4
pe run assistant.txt --provider anthropic:claude-3-haiku

# Comprehensive evaluation
pe eval config.yaml --output results.json
pe vet *.txt  # Validate prompts and run embedded evals

# Advanced optimization
pe optimize assistant.txt --method pe2 --iterations 10
pe semantic backprop --prompt assistant.txt --target accuracy
pe evolve --prompt assistant.txt --algorithm nsga-ii

# Security testing
pe security scan --prompt assistant.txt
pe exp attest --help
```

### Production Operations (✅ Working Examples)
```bash
# Distributed execution
pe exp distributed --help
pe eval config.yaml --distributed --nodes node1,node2

# Performance monitoring
pe profile cpu --duration 30s
pe benchmark config.yaml --iterations 100
pe metrics --reference expected.txt --candidate output.txt

# Caching and verification
pe exp cache --help
pe exp cache --help
```

### Research & Experimentation (✅ Working Examples)
```bash
# Advanced evaluation with multiple metrics
pe eval config.yaml | pe analyze --metric latency,accuracy,cost
pe eval config.yaml | pe filter --success | pe stats

# Fusion and consensus
pe experimental optimize --help
pe experimental compose --components system.txt,task.txt --style dspy

# Pipeline processing
echo "analyze this text" | pe ask --provider openai:gpt-4 | pe extract --tag analysis
```

## Getting Started

1. **Install PE**: `go install github.com/tmc/pe/cmd/pe@latest`
2. **Initialize a project**: `pe init`
3. **Run your first prompt**: `pe run "Hello, world!"`
4. **Explore commands**: `pe help`

See the [Getting Started Guide](GETTING_STARTED.md) for a comprehensive tutorial.

## Next Steps

- [Installation Guide](INSTALLATION.md) - Detailed setup instructions
- [Command Reference](CLI_REFERENCE.md) - Complete command documentation
- [Tutorial](TUTORIAL.md) - Step-by-step guide
- [Architecture](ARCHITECTURE.md) - Technical deep dive
- [Plugin Development](PLUGINS.md) - Extend PE with custom plugins
