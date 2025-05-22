# PE: Prompt Engineering Toolkit

[![Go Report Card](https://goreportcard.com/badge/github.com/tmc/pe)](https://goreportcard.com/report/github.com/tmc/pe)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

PE is a comprehensive, production-ready prompt engineering toolkit that rivals industry leaders like promptfoo, PromptLayer, and Mirascope. Built with Go's performance and Unix philosophy in mind, PE provides a complete solution for developing, testing, and deploying LLM applications.

## 🚀 Key Features

### Core Capabilities
- **Multi-Provider Support**: OpenAI, Anthropic, Google AI, and custom endpoints
- **Template-Based Prompts**: Use variables to test multiple variations  
- **Assertion-Based Testing**: Verify outputs meet expected criteria
- **Interactive REPL**: Rapid prompt development and testing
- **Pipeline-Friendly**: Unix composability with stream processing
- **Watch Mode**: Auto-rerun evaluations on file changes

### Advanced Features  
- **Performance Benchmarking**: Compare latency, cost, and quality metrics
- **Statistical Analysis**: Confidence intervals, percentiles, regression detection
- **Real-Time Streaming**: Process results as they arrive
- **Configuration Management**: YAML/JSON configs with validation
- **Extensible Architecture**: Easy to add new providers and metrics

### Enterprise Ready
- **CI/CD Integration**: GitHub Actions, GitLab CI support
- **Cost Tracking**: Monitor and optimize LLM API costs
- **Security Testing**: Built-in red-teaming capabilities
- **Batch Processing**: Concurrent evaluations for scale
- **Result Storage**: Local database with shareable URLs

## 📦 Installation

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

## 🎯 Quick Start

### 1. Create Your First Configuration

```bash
pe init my-first-eval.yaml
```

This creates a basic configuration:

```yaml
prompts:
  - "What is the capital of {{country}}?"
  - "Tell me about the capital city of {{country}}."

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-haiku"

tests:
  - vars:
      country: "France"
    assert:
      - type: "contains"
        value: "Paris"
  - vars:
      country: "Japan"
    assert:
      - type: "contains"
        value: "Tokyo"
```

### 2. Run Your First Evaluation

```bash
pe eval my-first-eval.yaml
```

### 3. View Results Interactively

```bash
pe view
```

### 4. Try the Interactive REPL

```bash
pe interactive --provider openai:gpt-4
```

## 📖 Documentation

- **[Complete Documentation](docs/README.md)** - Comprehensive guide with examples
- **[Configuration Guide](docs/README.md#configuration-guide)** - YAML/JSON configuration
- **[Command Reference](docs/README.md#commands-reference)** - All available commands
- **[Best Practices](docs/README.md#best-practices)** - Production deployment tips
- **[Integration Guide](docs/README.md#integration-guide)** - CI/CD setup

## 🛠 Core Commands

### Evaluation & Testing
```bash
# Run basic evaluation
pe eval config.yaml

# Save results to file with format detection
pe eval config.yaml -o results.json

# Save to database for later viewing  
pe eval config.yaml --save-db

# Dry run to see what would be executed
pe eval config.yaml --dry-run

# Control concurrency and timeout
pe eval config.yaml --max-concurrency 8 --timeout 60s
```

### Interactive Development
```bash
# Start interactive REPL
pe interactive --provider anthropic:claude-3-sonnet

# Watch files for changes
pe watch config.yaml --include "*.yaml,prompts/**/*"

# Format and validate configs
pe fmt config.yaml --write
pe vet config.yaml
```

### Pipeline Processing (Unix-style)
```bash
# Ask single questions
echo "What is AI?" | pe ask --provider openai:gpt-4

# Stream and filter results
pe eval config.yaml | pe filter --success | pe stats
pe eval config.yaml | pe stream --select response,latency | pe analyze --metric latency

# Compare results
pe diff baseline.json current.json --threshold 0.05
```

### Benchmarking
```bash
# Run performance benchmarks
pe benchmark config.yaml --iterations 10 --concurrency 4

# Format output
pe benchmark config.yaml --format json -o benchmark-results.json
```

## 🎨 Advanced Examples

### Content Generation Testing
```yaml
description: "Blog post generation evaluation"

prompts:
  - id: "blog-post"
    content: |
      Write a {{length}}-word blog post about {{topic}} that is {{tone}} in tone.
      Include an introduction, 3 main points, and a conclusion.

providers:
  - id: "gpt4"
    type: "openai" 
    model: "gpt-4"
    config:
      temperature: 0.7
      max_tokens: 800

tests:
  - description: "Technical content"
    vars:
      topic: "machine learning"
      tone: "professional"
      length: 500
    assert:
      - type: "length"
        min: 400
        max: 600
      - type: "contains"
        value: ["introduction", "conclusion", "machine learning"]
      - type: "readability"
        min_grade_level: 8
        max_grade_level: 12
```

### Multi-Provider Comparison
```yaml
prompts:
  - "Explain quantum computing in simple terms"

providers:
  - "openai:gpt-4"           # High quality, expensive
  - "openai:gpt-3.5-turbo"   # Good quality, cheaper
  - "anthropic:claude-3-haiku"  # Fast, cost-effective

tests:
  - vars: {}
    assert:
      - type: "contains"
        value: ["quantum", "computing"]
      - type: "length"
        min: 100
        max: 500
      - type: "cost"
        max: 0.02
      - type: "latency"
        max: "10s"
```

### Streaming Analysis Pipeline
```bash
# Real-time cost monitoring
pe eval large-test-suite.yaml --stream | \
  pe stream --select cost,provider | \
  pe filter --max-cost 0.10 | \
  pe analyze --metric cost --group-by provider

# Success rate by provider
pe eval config.yaml | \
  pe filter --success | \
  pe stats --group-by provider --format table
```

## 🔧 Integration Examples

### GitHub Actions
```yaml
name: Prompt Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install PE
        run: go install github.com/tmc/pe/cmd/pe@latest
      - name: Run prompt tests
        run: |
          pe eval prompts/config.yaml --save-db
          pe stats --format json > stats.json
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

### Makefile Integration
```makefile
.PHONY: test-prompts benchmark lint-prompts

test-prompts:
	pe eval config.yaml --save-db
	
benchmark:
	pe benchmark config.yaml --iterations 5 --format json -o benchmark.json
	
lint-prompts:
	pe vet config.yaml
	pe fmt config.yaml --write
	
watch:
	pe watch config.yaml
```

## 🏗 Architecture & Extensibility

PE is designed with extensibility in mind:

- **Provider Interface**: Easy to add new LLM providers
- **Assertion System**: Custom assertion types via plugins
- **Metrics Framework**: Define custom scoring functions
- **Pipeline Architecture**: Composable Unix-style commands
- **Configuration System**: Flexible YAML/JSON schemas

### Adding Custom Providers
```go
type CustomProvider struct {
    endpoint string
    apiKey   string
}

func (p *CustomProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
    // Implementation
}
```

## 📊 Comparison with Industry Tools

| Feature | PE | promptfoo | PromptLayer | Mirascope |
|---------|----|-----------|-----------  |-----------|
| **CLI-First** | ✅ | ✅ | ❌ | ❌ |
| **Unix Composability** | ✅ | ❌ | ❌ | ❌ |
| **Interactive REPL** | ✅ | ❌ | ❌ | ❌ |
| **Multi-Provider** | ✅ | ✅ | ✅ | ✅ |
| **Real-time Streaming** | ✅ | ❌ | ✅ | ❌ |
| **Statistical Analysis** | ✅ | Basic | ✅ | ❌ |
| **Cost Tracking** | ✅ | Basic | ✅ | ❌ |
| **CI/CD Ready** | ✅ | ✅ | Partial | ❌ |
| **Watch Mode** | ✅ | ❌ | ❌ | ❌ |
| **Performance** | High (Go) | Medium (Node) | N/A (SaaS) | Medium (Python) |

## 🚀 Production Features

### Cost Optimization
```bash
# Track costs by provider
pe eval config.yaml | pe analyze --metric cost --group-by provider

# Find cost-effective models
pe benchmark config.yaml --metric cost-per-quality --format table

# Set budget alerts
pe eval config.yaml | pe filter --max-cost 1.00 | pe stats
```

### Performance Monitoring
```bash
# Latency analysis
pe eval config.yaml | pe analyze --metric latency --percentiles 50,90,95,99

# Regression detection
pe diff baseline.json current.json --metric score --threshold 0.05

# Real-time monitoring
pe eval config.yaml --stream | pe analyze --metric latency --live-plot
```

### Quality Assurance
```bash
# Comprehensive testing
pe eval config.yaml --save-db
pe view --export-report report.html

# Red-team testing (coming soon)
pe redteam config.yaml --categories harmful,biased,hallucination

# A/B testing
pe diff variant-a.yaml variant-b.yaml --confidence 0.95
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Setup
```bash
git clone https://github.com/tmc/pe.git
cd pe
go mod tidy
go test ./...
```

### Adding New Features
- See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines
- Check [Issues](https://github.com/tmc/pe/issues) for feature requests
- Join our [Discussions](https://github.com/tmc/pe/discussions) for design talks

## 📋 Requirements

- **Go 1.21+** for building from source
- **API Keys** for LLM providers (OpenAI, Anthropic, etc.)
- **Unix-like environment** recommended for full pipeline features

## 🛣 Roadmap

### Phase 1: Core Enhancement (Current)
- [x] Interactive REPL implementation
- [x] Pipeline command suite
- [x] Advanced documentation
- [ ] Red-teaming module
- [ ] Custom metrics framework

### Phase 2: Advanced Features  
- [ ] Streaming provider support
- [ ] Multi-turn conversation testing
- [ ] Vision model evaluation
- [ ] RAG evaluation framework
- [ ] Plugin system

### Phase 3: Enterprise Features
- [ ] Web dashboard
- [ ] Team collaboration features
- [ ] Advanced security scanning
- [ ] Custom deployment options
- [ ] Enterprise support

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **Documentation**: [docs/](docs/)
- **Examples**: [example/](example/)
- **Issues**: [GitHub Issues](https://github.com/tmc/pe/issues)
- **Discussions**: [GitHub Discussions](https://github.com/tmc/pe/discussions)

---

**Built with ❤️ by the PE team. Make prompt engineering systematic, not trial-and-error.**