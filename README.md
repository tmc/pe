# PE: Safe Prompting Toolchain

[![CI](https://github.com/tmc/pe/actions/workflows/ci.yml/badge.svg)](https://github.com/tmc/pe/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tmc/pe)](https://goreportcard.com/report/github.com/tmc/pe)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-current-blue)](docs/README.md)
[![Roadmap](https://img.shields.io/badge/roadmap-ROADMAP.md-lightgrey)](ROADMAP.md)
[![Coverage](https://img.shields.io/badge/coverage-48.8%25-orange)](docs/TEST_COVERAGE_REPORT.md)

**PE is a Go-like toolchain for safe prompting.** Its primary artifact is
executable, templated, composable text: plain text by default, with optional
inputs, metadata, safety policy, and placement rules when a file needs a
stronger contract.

## 🚀 Key Features (Implemented)

### Optimization
PE includes local prompt optimization experiments:
- **Semantic Backpropagation & GASO** (KAUST/IDSIA 2025): Natural language gradients for prompt optimization
- **Textual Gradient Optimization**: Gradient-style optimization through textual feedback
- **Multiple Optimization Methods**: PE2, APEX, multistage, reflection, and evolutionary approaches

### Core Capabilities (Stable Core + Prototype Extensions)
- **Multi-Provider Support**: Native OpenAI and Anthropic providers with full API implementations; see [docs/TEST_COVERAGE_REPORT.md](docs/TEST_COVERAGE_REPORT.md) for measured coverage
- **Advanced Evaluation**: Pass@N metrics, structured output validation, 15+ assertion types (some advanced types in development)
- **Unix Pipeline Philosophy**: generated CLI commands for streaming prompt processing
- **Performance**: Native Go implementation with comprehensive benchmarking
- **Module Management**: Core module system (mod init/tidy/vendor) - registry features in development
- **Security Testing**: OWASP-oriented checks with incomplete analyses reported
  explicitly or failed closed
- **Local Scheduling**: Bounded local task and consensus prototypes under `pe exp`
- **Local Attestation**: Unsigned SHA-256 manifest prototype exposed under `pe exp attest`

## 🎯 Quick Start

### Installation

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Basic Usage

```bash
# Execute a prompt - simplest form
pe run "What is 2+2?"

# Render executable text without invoking a provider
pe run-text review.prompt --var topic=release

# Validate module capability policy and executable text metadata
pe mod vet review.prompt

# From a file with template variables
pe run translate.prompt --var text="Hello" --var language="French"

# With stdin input
cat article.txt | pe run -

# Using pipe from another command
cat article.txt | pe run summarize --provider cgpt
```

### Template Variables

PE uses Go template syntax for variables:

```bash
# Using --var flags
pe run translate.prompt --var text="Hello" --var from="English" --var to="Spanish"

# Variables are replaced in the prompt
# {{.text}} becomes "Hello"
# {{.from}} becomes "English"
# {{.to}} becomes "Spanish"
```

## 📝 Prompt Format

PE uses a simple, self-documenting format. Start simple and add features as needed:

### Level 1: Just Text
```
What is the capital of France?
```

### Level 2: Templates
```
Solve this math problem: {{.expression}}
```

### Level 3: Documentation & Examples
```
Solve the following math problem step by step: {{.expression}}

-- system-prompt --
You are a helpful math tutor that explains problem-solving step by step.

-- defaults --
expression=2+2

-- examples/simple/ideal-output --
Let me solve this step by step:
15 + 27 = 42
```

### Level 4: Full Features
```
#!/usr/bin/env pe run

Translate the following {{.source_lang}} text to {{.target_lang}}:

{{.text}}

-- system-prompt --
You are a professional translator with native fluency in multiple languages.

-- defaults --
source_lang=English
target_lang=Spanish

-- variant:formal --
extend-system-prompt Use formal, professional language suitable for business.

-- config --
temperature 0.3
max-tokens 1000
```

See [docs/TEMPLATE_SYNTAX.md](docs/TEMPLATE_SYNTAX.md) for template syntax details.

## 📋 Available Commands

Run `pe --help` for the generated command list. The current top-level surface includes:

### Core Commands
- `pe run` - Execute a prompt immediately with variable substitution.
- `pe eval` - Evaluate prompt configurations against providers and assertions.
- `pe test` - Run prompt tests and generate test suites.
- `pe benchmark` - Benchmark prompts, providers, and configurations.
- `pe vet` - Validate prompt files and run their evals.
- `pe build` - Write prompts, metadata, and optional bundles.
- `pe version` - Print the PE version.

### Prompt and Configuration Commands
- `pe prompt` - Manage prompt files.
- `pe cat` - Inspect prompt files with variable substitution.
- `pe fmt` - Format prompts.
- `pe doc` - Show prompt documentation.
- `pe template` - Manage prompt templates.
- `pe convert` - Convert promptfoo configuration files.
- `pe expand` - Resolve file references and globs in configuration files.
- `pe extract` - Extract content from XML-like tags.

### Pipeline Commands
- `pe ask` - Execute prompts with optional templating.
- `pe stream` - Stream-process LLM outputs.
- `pe filter` - Filter and transform pipeline outputs.
- `pe analyze` - Analyze text with metrics.
- `pe collect` - Collect async operation results.
- `pe reduce` - Aggregate pipeline results.
- `pe stats` and `pe diff` - Summarize and compare evaluation results.

### Project and Extension Commands
- `pe init` - Initialize a PE repository.
- `pe mod` - Manage prompt modules.
- `pe work` - Manage prompt workspaces.
- `pe push` and `pe get` - Work with prompt modules and metadata.
- `pe plugin` - Manage PE plugins.
- `pe security` - Run security testing workflows.
- `pe profile`, `pe view`, and `pe watch` - Inspect, view, and rerun workflows.
- `pe interactive` - Start the prompt-development REPL.

### Experimental Commands
- `pe experimental` and `pe exp` expose prototype optimization, composition,
  local scheduling, consensus, unsigned attestation, local cache, workflow,
  import/export, and report commands.
  Treat these as active development surfaces unless their subcommand docs state
  otherwise.

### Planned Features

See [ROADMAP.md](ROADMAP.md) for the tracked roadmap and release work. Current
release-prep focus areas are documentation accuracy, example validation,
security review, build/distribution checks, and command-reference cleanup.

## 🧪 Evaluation Configuration

PE uses YAML configuration for comprehensive prompt evaluation:

```yaml
description: "Summarization evaluation"

prompts:
  - "Summarize this text: {{.text}}"

providers:
  - "openai:gpt-4o-mini"
  - id: "anthropic:claude-3-haiku-20240307"
    label: "claude haiku"
    config:
      temperature: 0
  - id: "ollama:qwen3.5:4b"
    label: "ollama raw"
    env:
      OLLAMA_HOST: "http://localhost:11434"
    delay: 150ms
    config:
      base_url: "http://localhost:11434"
      raw: true
      seed: 1
      num_predict: 200

tests:
  - vars:
      text: "Long article about AI developments..."
    assert:
      - type: contains
        value: "AI"
      - type: max_length
        value: 200
      - type: llm_rubric
        value: "Summary should be concise and capture main points"
```

`providers:` accepts either a plain provider spec string like `"openai:gpt-4o-mini"` or an object with `id`, `label`, `config`, `env`, `prompts`, and `delay`. `pe eval` and `pe benchmark` now use the same provider shape.

For a local-runtime benchmark example, see `examples/benchmarks/local-runtime-comparison.yaml`.
The design note for the provider/materialization split is in `docs/LOCAL_RUNTIME_PROVIDER_DESIGN.md`.

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

PE implements multiple optimization techniques under `pe experimental`:

```bash
# Textual gradients
pe experimental optimize --prompt "task" --method textgrad --iterations 5

# Semantic Backpropagation (2025 research)
pe experimental semantic backprop --prompt "task" --objective "accuracy"

# PE2 - Prompt Engineering Squared
pe experimental optimize --prompt "task" --method pe2 --iterations 5

# APEX long-prompt optimization
pe experimental optimize --prompt "task" --method apex --iterations 5
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

See [docs/future/README.md](docs/future/README.md) for aspirational design
directions.

## 📚 Documentation

- **[Documentation Overview](docs/README.md)** - Start here for accurate implementation status
- [Getting Started](docs/GETTING_STARTED.md)
- [Command Reference](docs/CLI_REFERENCE.md)
- [Architecture](docs/ARCHITECTURE.md)
- [API Reference](docs/API_REFERENCE.md)
- [Examples](examples/)
- [Future Features](docs/future/) - Planned designs and specifications

## 🎯 Project Status

PE has a stable core CLI, native provider support, and a comprehensive
evaluation framework, with experimental features exposed under `pe exp` and
`pe experimental`. See [RELEASE_NOTES.md](RELEASE_NOTES.md),
[docs/CURRENT_STATUS.md](docs/CURRENT_STATUS.md), and
[ROADMAP.md](ROADMAP.md) for release status and remaining work.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.
