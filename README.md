# PE: Go for Prompts

[![Go Report Card](https://goreportcard.com/badge/github.com/tmc/pe)](https://goreportcard.com/report/github.com/tmc/pe)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-comprehensive-blue)](docs/)
[![Research](https://img.shields.io/badge/research-2024--2025-green)](docs/RESEARCH_FOUNDATIONS.md)

**PE brings the simplicity and power of Go's toolchain to prompt engineering.** Like `go run` for prompts, PE makes it easy to develop, test, and optimize prompts with a familiar, composable command-line interface.

## 🚀 Key Features (Implemented)

### Research-Based Optimization
PE implements breakthrough research from 2024-2025:
- **Semantic Backpropagation & GASO** (KAUST/IDSIA 2025): Natural language gradients for prompt optimization
- **TextGrad Implementation**: Gradient-based optimization through textual feedback
- **Multiple Optimization Methods**: PE2, APEX, multistage, reflection, and evolutionary approaches

### Core Capabilities (✅ Fully Implemented)
- **Multi-Provider Support**: Native OpenAI (74% test coverage) and Anthropic (73.3% test coverage) providers with full API implementations
- **Advanced Evaluation**: Pass@N metrics, structured output validation, 15+ assertion types (some advanced types in development)
- **Unix Pipeline Philosophy**: 47 composable CLI commands for streaming prompt processing
- **Performance**: Native Go implementation with comprehensive benchmarking
- **Module Management**: Core module system (mod init/tidy/vendor) - registry features in development
- **Security Testing**: Full OWASP LLM Top 10 coverage via integrated security module
- **Distributed Execution**: Complete distributed system with P2P networking and consensus
- **Cryptographic Attestation**: Full signing and verification of prompt runs

## 🎯 Quick Start

### Installation

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Basic Usage

```bash
# Execute a prompt - simplest form
pe run "What is 2+2?"

# From a file with template variables
pe run math-solver.prompt "15 + 27"

# Using examples built into the prompt
pe run math-solver.prompt --example example-1

# With explicit variables
pe run translate.prompt --var TEXT="Hello" --var LANGUAGE="French"
```

### Template Variables

PE automatically detects template variables and provides helpful usage:

```bash
$ pe run math-solver.prompt
Error: Template variables found but no values provided

Template variables detected: EXPRESSION

Available flags:
      --model string        Model to use (e.g., gpt-4, claude-3)
      --temperature float   Temperature for randomness (0.0-1.0) (default 0.7)
      --var stringToString  Template variables (can be repeated)
      --example string      Run with example variables (e.g., example-1)
      --expression string   Value for template variable EXPRESSION

Examples:
  pe run math-solver.prompt --expression 'value'
  pe run math-solver.prompt 'value'  # positional argument
```

## 📝 Prompt Format

PE uses a simple, self-documenting format. Start simple and add features as needed:

### Level 1: Just Text
```
What is the capital of France?
```

### Level 2: Templates
```
Solve this math problem: {{EXPRESSION}}
```

### Level 3: Documentation & Examples
```
Solve the following math problem step by step: {{EXPRESSION}}

-- prompt-summary --
A helpful math tutor that explains problem-solving step by step.

-- variable-description/EXPRESSION --
A mathematical expression to solve (e.g., "2+2", "15*3", "(10+5)/3")

-- examples/simple/EXPRESSION --
15 + 27

-- examples/simple/ideal-output --
Let me solve this step by step:
15 + 27 = 42
```

### Level 4: Full Features
```
#!/usr/bin/env pe run --model gpt-4

Translate the following {{LANGUAGE}} text to {{TARGET_LANGUAGE}}:

{{TEXT}}

-- system-prompt --
You are a professional translator with native fluency in multiple languages.

-- defaults --
LANGUAGE=English
TARGET_LANGUAGE=Spanish

-- variant:formal --
extend-system-prompt Use formal, professional language suitable for business.

-- config --
temperature 0.3
max-tokens 1000
```

See [docs/PROMPT_FORMAT_SPEC.md](docs/PROMPT_FORMAT_SPEC.md) for the complete specification.

## 📋 Available Commands (47 Total - Core Features Stable ✅)

### Core Commands
- `pe run` - Execute prompts with variable substitution and native providers
- `pe eval` - Comprehensive evaluation with 20+ assertion types
- `pe optimize` - Multiple metaprompting optimization algorithms
- `pe semantic` - Semantic backpropagation and GASO (2025 research)
- `pe test` - Advanced testing with property-based and regression approaches
- `pe benchmark` - Performance benchmarking with statistical analysis
- `pe metrics` - Advanced metrics (BLEU, ROUGE, BERTScore, G-Eval, UniEval)
- `pe profile` - Performance profiling and observability
- `pe playground` - Interactive web-based prompt development
- `pe security` - Complete OWASP LLM Top 10 security testing

### Pipeline Commands (Unix Composability)
- `pe ask` - Execute prompts via pipeline
- `pe stream` - Stream processing with filtering
- `pe filter` - Filter and transform outputs with JSON support
- `pe analyze` - Statistical analysis with advanced metrics
- `pe collect` - Collect results from async operations
- `pe reduce` - Aggregate and reduce pipeline results
- `pe stats` - Quick statistical summaries

### Module & Advanced Commands
- `pe mod init/tidy/vendor` - Core module management (registry features in next-experimental branch)
- `pe compose` - Component-based prompt composition with type safety
- `pe attest` - Cryptographic attestation with keychain integration
- `pe cat` - Inspect prompt files with variable substitution and component viewing
- `pe extract` - Extract structured data with XML/JSON parsing
- `pe evolve` - Evolutionary optimization with NSGA-II algorithms
- `pe fusion` - Multi-model consensus with reliability optimization
- `pe distributed start/join/status/stop` - Complete distributed execution
- `pe cache` - Content-addressed caching with cryptographic verification
- `pe synthesize` - DSPy-style program synthesis
- `pe work` - Workspace management for complex projects

### Recently Completed Features (✅ Production Ready)
- **Native OpenAI/Anthropic Providers**: Full API implementations with 74%/73.3% test coverage
- **Distributed Execution**: Complete P2P networking and task distribution
- **Web Playground Interface**: Interactive prompt development environment
- **Security Testing Suite**: Full OWASP LLM Top 10 coverage
- **Advanced Metrics**: BERTScore, G-Eval, UniEval implementations
- **Cryptographic Attestation**: Full signing and verification system
- **Prompt File Inspector**: pe cat command with variable substitution and component inspection

### In Development (🚧 - Available in `next-experimental` branch)
- Module registry system with download/list/get commands
- Advanced assertion types (toxicity, coherence, factuality, similarity)
- Interactive REPL mode
- Extended provider ecosystem (Ollama, local models)
- REST API server implementation

### Planned Features (📝 Roadmap)
- Multi-modal support (vision, audio)
- Visual prompt engineering tools
- IDE integrations (VS Code, JetBrains)
- Neurosymbolic prompt synthesis
- Advanced caching strategies

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
- `internal/promptfoo/evaluation/evaluator/` - Evaluation engine
- `internal/promptfoo/evaluation/metrics/` - Advanced metrics
- `internal/providers/` - LLM provider abstraction
- `internal/structured/` - Structured output support
- `internal/promptfoo/execution/distributed/` - Distributed execution (experimental)

## 🤝 Contributing

PE is under active development. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 🎯 Design Philosophy

PE follows Go's philosophy: simple, composable tools that do one thing well:

- **Start Simple**: A prompt can just be text. Add complexity only as needed.
- **Progressive Enhancement**: Templates → Documentation → Examples → Variants
- **Familiar Patterns**: `pe.mod` files work like `go.mod`. Commands compose via pipes.
- **No Magic**: Everything is explicit and inspectable.

See [docs/future/DESIGN_PHILOSOPHY.md](docs/future/DESIGN_PHILOSOPHY.md) for more details.

## 📚 Documentation

- **[Documentation Overview](docs/README.md)** - Start here for accurate implementation status
- [Getting Started](docs/GETTING_STARTED.md)
- [Command Reference](docs/COMMANDS.md)
- [Architecture](docs/ARCHITECTURE.md)
- [API Reference](docs/API_REFERENCE.md)
- [Examples](example/)
- [Future Features](docs/future/) - Planned designs and specifications

## 🎯 Project Status

PE is production-ready with 47 stable CLI commands, native provider support, and comprehensive evaluation framework. Advanced experimental features are available in the `next-experimental` branch. See [RELEASE_NOTES.md](RELEASE_NOTES.md) for detailed feature status.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.