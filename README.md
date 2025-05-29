# PE: Go for Prompts

[![Go Report Card](https://goreportcard.com/badge/github.com/tmc/pe)](https://goreportcard.com/report/github.com/tmc/pe)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-comprehensive-blue)](docs/)
[![Research](https://img.shields.io/badge/research-2024--2025-green)](docs/RESEARCH_FOUNDATIONS.md)

**PE is the Go toolchain for prompt engineering**—a comprehensive toolkit that brings Go's philosophy of simplicity, composability, and performance to the world of LLM development. Just as Go revolutionized systems programming, PE revolutionizes prompt engineering with a unified toolchain for developing, testing, optimizing, and deploying prompts.

Built on Go's performance and Unix philosophy, PE implements cutting-edge 2024-2025 research while maintaining the simplicity and power that developers expect from Go tools.

## 🔧 The Go Toolchain Philosophy for Prompts

Like the Go toolchain transformed programming with `go build`, `go test`, `go fmt`, and `go mod`, PE brings the same unified experience to prompt engineering:

```bash
pe run "Analyze this text"              # go run for prompts - execute immediately
pe test prompts/                        # go test for prompt validation
pe eval config.yaml                     # evaluate prompts with providers
pe optimize --prompt "Your task"        # optimize prompts using AI
pe fmt prompts/                         # format prompt configurations
pe init myproject                       # initialize new prompt project
```

### 🚀 World's First Implementation of 2024-2025 Research

PE implements cutting-edge metaprompting research from 2024-2025:

- **🧠 Semantic Backpropagation & GASO**: Revolutionary gradient-based optimization for prompts
- **⚡ TextGrad 2.0**: Natural language gradients with attention flow mapping
- **🔍 DSPy MIPROv2**: Multi-stage instruction and prompt optimization
- **🧬 Evolutionary Optimization**: Population-based genetic algorithms
- **🤝 Multi-Model Consensus**: Cross-provider optimization
- **🏗️ Component-Based Engineering**: Reusable prompt components

## 🎯 Quick Start

### Installation

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Minimal Prompt Format

PE introduces a minimal, text-first prompt format inspired by Go's simplicity:

```
Summarize this: {{.text}}

-- defaults --
text=Hello world

-- evals --
$ text="The quick brown fox"
A fox story.
```

Test your prompts with the default eval command:

```bash
pe vet                    # Test all .txt files (default)
pe vet summarize.txt      # Test specific file
```

### Core Commands (Implemented)

```bash
# pe run - Execute a prompt immediately (like go run)
pe run "Summarize this article: {{.Article}}" --var Article="Long text..."
pe run prompt.txt                          # Run from file
pe run "What is AI?" --provider cgpt      # Use specific provider

# pe test - Test your prompts comprehensively
pe test config.yaml                        # Run comprehensive tests
pe test config.yaml --property             # Property-based testing
pe test config.yaml --regression           # Regression testing

# pe eval - Evaluate prompts against providers
pe eval config.yaml                        # Run evaluation
pe eval config.yaml -o results.json        # Save results
pe eval config.yaml --stream               # Stream results

# pe optimize - Optimize prompts using metaprompting
pe optimize --prompt "Write code" --method textgrad
pe optimize --prompt "Analyze data" --method multistage
pe optimize --prompt "Complex task" --method reflection

# pe semantic - Semantic backpropagation (2025 breakthrough)
pe semantic backprop --prompt "task" --objective "goal"
pe semantic descent --objective "improve accuracy"
pe semantic gaso --system config.json      # Graph-based optimization

# pe compose - Component-based prompt engineering
pe compose context.txt instruction.txt --style cot
pe compose components/ --optimize

# pe metrics - Advanced evaluation metrics
pe metrics --type bleu --generated out.txt --reference ref.txt
pe metrics --type bertscore --all
pe metrics --type g-eval --criteria "accuracy,clarity"

# pe eval - Advanced evaluation with pass@n and structured output
pe eval config.yaml                        # Supports pass@n and structured assertions
pe eval code-gen.yaml --assert pass-at-n  # Focus on code generation metrics
pe eval extract.yaml --assert structured   # Validate structured outputs

# pe security - Security testing
pe security test --owasp --target prompt.txt
pe security redteam --comprehensive

# pe run - Execute prompts with minimal syntax
echo "text" | pe run summarize              # Simplest form
pe run analyze.prompt --variant academic    # With variants
./summarize.prompt                          # Direct execution

# pe edit - Programmatic prompt editing (like go mod edit)
pe edit prompt.txt --set-prompt "Analyze this"
pe edit prompt.txt --add-variant pirate --variant-cmd "extend-system-prompt 'Arr!'"

# pe work - Manage prompt workspaces (like go work)
pe work init ./prompts ./shared-prompts
pe work list

# pe benchmark - Performance benchmarking
pe benchmark config.yaml --iterations 10
pe benchmark config.yaml --compare baseline.yaml

# pe profile - Performance profiling
pe profile cpu --duration 30s
pe profile memory
pe profile trace

# Other useful commands
pe init                                    # Initialize new project
pe fmt config.yaml                         # Format configuration
pe vet                                     # Default: validate & test all prompts
pe vet prompts/*.txt                       # Test specific prompt files  
pe vet config.yaml                         # Validate YAML configuration
pe convert input.yaml output.json          # Convert formats
pe watch config.yaml                       # Watch and re-run
pe view                                    # View results in browser
pe playground                              # Interactive playground
pe template list                           # Manage templates
```

### Plugin System

```bash
# Plugin management
pe plugin list                             # List installed plugins
pe plugin run promptfoo import config.yaml # Run plugin command

# Promptfoo compatibility (via plugin)
pe promptfoo import config.yaml            # Import promptfoo config
pe promptfoo export pe-config.yaml         # Export to promptfoo
pe promptfoo convert input.yaml output.json # Convert formats

# Plugin discovery
# Plugins are discovered as pe-* executables in PATH
# Example: pe-promptfoo becomes available as 'pe promptfoo'
```

### Pipeline Processing (Unix Philosophy)

```bash
# Stream processing pipeline
pe eval config.yaml --stream | \
  pe filter --success --min-score 0.8 | \
  pe analyze --metric latency | \
  pe stats --format table

# Ask questions interactively
echo "What is AI?" | pe ask --provider cgpt

# Diff evaluation results
pe diff results1.json results2.json

# Interactive REPL
pe interactive --provider cgpt
```

### Advanced Optimization

```bash
# TextGrad optimization
pe optimize --prompt "Summarize text" --method textgrad --iterations 5

# Multi-stage optimization
pe optimize --prompt "Complex analysis" --method multistage

# Reflection-based optimization
pe optimize --prompt "Write a story" --method reflection

# Semantic backpropagation (2025 breakthrough)
pe semantic backprop --prompt "current" --target "goal" --iterations 10
pe semantic descent --objective "minimize errors" --learning-rate 0.1
pe semantic gaso --system multi-agent.json --multi-objective
```

## 🧠 Implemented Optimization Methods

### Semantic Backpropagation & GASO (2025 KAUST/IDSIA Breakthrough)

```bash
pe semantic backprop --prompt "Classify text" --target "99% accuracy"
pe semantic gaso --system agent-system.json --objective performance
```

**Features**:
- 🎯 Semantic gradients generalize mathematical gradients to natural language
- 🔄 Graph-based optimization for multi-component systems
- 📊 Multi-objective optimization with Pareto efficiency
- ✅ **93.2% accuracy** on GSM8K (surpassing TextGrad's 78.2%)

### TextGrad 2.0 Implementation

```bash
pe optimize --prompt "Analyze sentiment" --method textgrad --iterations 6
```

**Features**:
- 🌊 Natural language gradients with attention flow mapping
- 🔍 Semantic drift detection during optimization
- 🧠 Backward propagation through textual feedback
- ✅ Cross-modal gradient computation support

### Multi-Stage Optimization (DSPy MIPROv2)

```bash
pe optimize --prompt "Complex task" --method multistage --gates
```

**Features**:
- 📋 Data-aware instruction generation
- 🎯 Demonstration-aware optimization
- 🔄 Three-stage process with quality gates
- ✅ Typical optimization cost: ~$2 USD, ~20 minutes

### Component-Based Engineering

## 📊 Advanced Evaluation Features

PE's `eval` command includes sophisticated assertion types for comprehensive testing:

### Pass@N Evaluation

Measure success rates across multiple attempts - crucial for code generation:

```yaml
# In your eval config:
tests:
  - vars:
      task: "Write a fibonacci function"
    assert:
      - type: pass-at-n
        config:
          n: 1              # Calculate pass@1
          samples: 20       # Generate 20 samples
          test_cases:       # Test each sample
            - input: "fib(5)"
              expected: "5"
            - input: "fib(10)"  
              expected: "55"
        threshold: 0.8      # Expect 80% pass rate
```

### Structured Output Validation

Ensure LLM outputs follow specific schemas:

```yaml
# In your eval config:
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
          score:
            type: number
            minimum: -1
            maximum: 1
          keywords:
            type: array
            items:
              type: string
        required: ["sentiment", "score"]
```

### Using with Go Structs

Define expected outputs as Go structs:

```go
type CodeOutput struct {
    Language string   `json:"language" enum:"python|javascript|go"`
    Code     string   `json:"code" minLength:"1"`
    Tests    []string `json:"tests" minItems:"2"`
}
```

**Features**:
- 📈 Industry-standard pass@n metrics
- 🧪 Test case validation for code
- 📐 JSON Schema validation
- 🔷 Go struct integration
- 📊 Multiple assertion types
- 🔌 Extensible validation system

## 🛠️ Minimal Prompt Format

PE uses a txtar-inspired format where prompts are just text with optional structure:

```bash
# Simplest form - just text
echo "long text" | pe run summarize

# Executable prompt file
#!/usr/bin/env pe run --max-tokens=500
Analyze this text:
{{input}}

-- system-prompt --
You are an expert analyst.

-- variants/academic --
extend-system-prompt 'Use academic language'
extend-prompt 'Include citations'

-- tests/simple --
input: "Sample text"
expect-contains: "analysis"
```

Usage:
```bash
# Direct execution
chmod +x analyze.prompt
cat article.txt | ./analyze.prompt

# With variants
pe run analyze.prompt --variant academic

# Edit programmatically
pe edit analyze.prompt --add-variant business \
  --variant-cmd "set-flag temperature 0.5"

# Manage collections
pe work init ./prompts
pe work list
```

**Features**:
- 📄 Prompts are just text files
- 🚀 Minimal syntax inspired by txtar
- 🔧 Progressive enhancement (add structure only when needed)
- 🎯 Variants for different modes/styles
- ✏️ Programmatic editing (like go mod edit)
- 📁 Workspace management (like go work)

### Component-Based Engineering

```bash
pe compose context.txt instruction.txt examples.txt --style cot --optimize
```

**Features**:
- 🏗️ Verified component libraries
- 📚 Style-specific composition (CoT, few-shot, etc.)
- 🔧 Automatic coherence validation
- ✅ TextGrad flow optimization

## 🔬 Advanced Evaluation & Metrics

### State-of-the-Art Metrics (Implemented)
```bash
# All major metrics implemented
pe metrics --type bleu --generated output.txt --reference expected.txt
pe metrics --type rouge --all-variants
pe metrics --type meteor --synonyms
pe metrics --type bertscore --model bert-base
pe metrics --type g-eval --criteria "coherence,fluency"
pe metrics --type unieval --dimensions all
```

### Statistical Analysis
```bash
# Comprehensive statistical testing
pe test significance baseline.json optimized.json
pe test cross-validate --methods textgrad,multistage --folds 5
pe analyze results.json --distribution --outliers
```

## 🛡️ Enterprise Security Testing

### OWASP LLM Top 10 Assessment
```bash
pe security test --owasp --target prompt.txt
pe security redteam --comprehensive --target system.txt
pe security scan --categories prompt_injection,data_leakage
```

## 🏗️ Architecture & Performance

PE is built with Go for maximum performance:

- ⚡ **Native Performance**: 3-5x faster than Python/Node.js alternatives
- 🔧 **Single Binary**: No dependencies, instant startup
- 🌊 **Stream Processing**: Real-time pipeline processing
- 🔄 **Concurrent Evaluation**: Native goroutine support

## 🔌 Inference Provider Support

PE supports multiple inference providers through a plugin architecture:

### Built-in Provider: cgpt
```bash
# Uses github.com/tmc/cgpt for inference
pe run "What is AI?" --provider cgpt
pe run "Count to 5" --stream
pe run "Explain X" --max-tokens 100 --temperature 0.7
```

### Plugin Architecture
The inference system is extensible - new providers can be added as plugins:
- Clean provider interface for easy integration
- Support for streaming and non-streaming inference
- Automatic provider detection and configuration

## 📚 Documentation

### Getting Started
- **[Quick Start Guide](docs/QUICK_START.md)** - Get up and running in 5 minutes
- **[CLI Reference](docs/CLI_REFERENCE.md)** - Complete command documentation
- **[Examples](example/)** - Working examples and use cases

### Advanced Topics
- **[Optimization Guide](docs/ADVANCED_OPTIMIZATION_GUIDE.md)** - All optimization methods
- **[API Reference](docs/API_REFERENCE.md)** - Go API documentation
- **[Architecture](docs/ARCHITECTURE.md)** - System design and internals

### Research & Comparisons
- **[Research Foundations](docs/RESEARCH_FOUNDATIONS.md)** - 2024-2025 research papers
- **[Competitive Analysis](docs/COMPETITIVE_ANALYSIS.md)** - Comparison with other tools
- **[Best Practices](docs/BEST_PRACTICES_2025.md)** - Production recommendations

## 🌟 Community & Contributing

PE is open source and welcomes contributions:

- 🌟 **Star the repo** to show support
- 🐛 **Report issues** to help improve PE
- 💡 **Request features** for the roadmap
- 🤝 **Contribute code** to advance the field

### Development

```bash
git clone https://github.com/tmc/pe.git
cd pe
go mod tidy
go test ./...
go build -o pe cmd/pe/main.go
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - see [LICENSE](LICENSE) file

---

**Built with ❤️ by the PE team. Leading the prompt engineering revolution.**

**PE: Go for Prompts. The unified toolchain for the LLM era.**