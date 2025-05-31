# PE: Go for Prompts

[![Go Report Card](https://goreportcard.com/badge/github.com/tmc/pe)](https://goreportcard.com/report/github.com/tmc/pe)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-comprehensive-blue)](docs/)
[![Research](https://img.shields.io/badge/research-2024--2025-green)](docs/RESEARCH_FOUNDATIONS.md)

**PE is a Go-based toolkit for prompt engineering** that implements cutting-edge 2024-2025 research in prompt optimization. Built with Go's philosophy of simplicity and performance, PE provides tools for developing, testing, and optimizing prompts using state-of-the-art metaprompting techniques.

## 🚀 Key Features (Implemented)

### Research-Based Optimization
PE implements breakthrough research from 2024-2025:
- **Semantic Backpropagation & GASO** (KAUST/IDSIA 2025): Natural language gradients for prompt optimization
- **TextGrad Implementation**: Gradient-based optimization through textual feedback
- **Multiple Optimization Methods**: PE2, APEX, multistage, reflection, and evolutionary approaches

### Core Capabilities
- **Multi-Provider Support**: Currently supports cgpt CLI with extensible provider interface (⚠️ native providers in development)
- **Advanced Evaluation**: Pass@N metrics, structured output validation, comprehensive assertions
- **Unix Pipeline Philosophy**: Composable commands for streaming prompt processing
- **Performance**: Native Go implementation for speed and efficiency
- **Module Management**: Go-style module system (mod init/download/tidy/vendor)
- **Security Testing**: OWASP LLM Top 10 coverage via redteam module

## 🎯 Quick Start

### Installation

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Basic Usage

```bash
# Execute a prompt
pe run "Summarize this: {{.text}}" --var text="Long article..."

# Evaluate prompts
pe eval config.yaml -o results.json

# Optimize prompts using metaprompting
pe optimize --prompt "Write code" --method textgrad

# Semantic optimization (2025 research)
pe semantic backprop --prompt "task" --objective "goal"
```

## 📋 Available Commands

### Core Commands (✅ Implemented)
- `pe run` - Execute prompts with variable substitution
- `pe eval` - Comprehensive prompt evaluation with assertions
- `pe optimize` - Metaprompting-based optimization
- `pe semantic` - Semantic backpropagation and GASO
- `pe test` - Test prompts with property-based and regression testing
- `pe benchmark` - Performance benchmarking
- `pe metrics` - Calculate BLEU, ROUGE, BERTScore metrics
- `pe profile` - Performance profiling and analysis
- `pe playground` - Interactive prompt development
- `pe security` - Basic security analysis

### Pipeline Commands (✅ Implemented)
- `pe ask` - Ask a question via pipeline
- `pe stream` - Stream processing
- `pe filter` - Filter outputs
- `pe analyze` - Analyze results
- `pe stats` - Statistical analysis

### Module & Advanced Commands (✅ Implemented)
- `pe mod init/download/tidy/vendor` - Go-style module management
- `pe compose` - Component-based prompt composition
- `pe attest` - Cryptographic attestation for prompt runs
- `pe extract` - Extract structured data from prompts
- `pe passn` - Calculate pass@n metrics
- `pe evolve` - Evolutionary prompt optimization
- `pe fusion` - Multi-model fusion
- `pe distributed` - Distributed execution commands
- `pe cache` - Content-addressed caching

### In Development (🚧)
- Native OpenAI/Anthropic providers (partially implemented)
- Module registry with dependency resolution
- Full distributed execution integration
- Web dashboard and REST API

### Planned Features (📝 Roadmap)
- Ollama and local model support
- Multi-modal support (vision, audio)
- Visual prompt engineering tools
- IDE integrations
- Neurosymbolic prompt synthesis

## 🧪 Evaluation Configuration

PE uses YAML configuration for comprehensive prompt evaluation:

```yaml
prompts:
  - file: "summarize.txt"
    id: "summarizer"

providers:
  - id: "cgpt"

tests:
  - prompt: "summarizer"
    providers: ["cgpt"]
    vars:
      text: "Long article about AI..."
    assert:
      - type: contains
        value: "AI"
      - type: max-length
        value: 200
      - type: llm-rubric
        value: "Summary should be concise and capture main points"
```

### Advanced Assertions

PE supports sophisticated evaluation beyond simple string matching:

#### Pass@N Evaluation
```yaml
assert:
  - type: pass-at-n
    config:
      n: 1
      samples: 20
      temperature: 0.8
      test_cases:
        - input: "binary_search([1,3,5,7], 5)"
          expected: "2"
    threshold: 0.8
```

#### Structured Output Validation
```yaml
assert:
  - type: structured-output
    config:
      format: json
      schema:
        type: object
        properties:
          sentiment: 
            type: string
            enum: ["positive", "negative", "neutral"]
        required: ["sentiment"]
```

## 🔬 Optimization Methods

PE implements multiple state-of-the-art optimization techniques:

```bash
# TextGrad - Natural language gradients
pe optimize --prompt "task" --method textgrad --iterations 5

# Semantic Backpropagation (2025 research)
pe semantic backprop --prompt "prompt.txt" --target "accuracy"

# PE2 - Prompt Engineering Squared
pe optimize --prompt "task" --method pe2 --budget 50

# Multi-stage optimization
pe optimize --prompt "task" --method multistage --stages 3
```

## 🏗️ Architecture

PE is built with a modular architecture:

- `internal/metaprompt/` - Optimization algorithms
- `internal/evaluator/` - Evaluation engine
- `internal/metrics/` - Advanced metrics
- `internal/providers/` - LLM provider abstraction
- `internal/structured/` - Structured output support
- `internal/distributed/` - Distributed execution (experimental)

## 🤝 Contributing

PE is under active development. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📚 Documentation

- [Getting Started](docs/GETTING_STARTED.md)
- [API Reference](docs/API_REFERENCE.md)
- [Research Foundations](docs/RESEARCH_FOUNDATIONS.md)
- [Examples](example/)

## 🎯 Project Status

PE is a research implementation focusing on cutting-edge prompt optimization techniques. While core features are functional, some advanced capabilities are still in development. See [ROADMAP.md](ROADMAP.md) for planned features.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.