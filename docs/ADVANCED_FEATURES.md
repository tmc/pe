# Advanced Features Guide

This guide describes PE features beyond the basic `pe run` and `pe eval`
workflows: prompt optimization, richer assertions, pipelines, providers,
plugins, and observability.

## Optimization Methods

### TextGrad Optimization

PE implements TextGrad-style optimization using natural language feedback as
"textual gradients."

```bash
# Use TextGrad optimization
pe optimize --prompt "Summarize the key points" --method textgrad --iterations 5

# Hybrid approach combining standard + TextGrad
pe optimize --prompt "Analyze sentiment" --method hybrid --iterations 6
```

**How TextGrad Works:**

1. **Computation Graph**: PE builds a computation graph representing your prompt components
2. **Textual Gradients**: LLMs provide natural language feedback (gradients) for each component
3. **Gradient Application**: Feedback is systematically applied to improve the prompt
4. **Iterative Refinement**: Each iteration builds on previous improvements

**Properties:**
- More nuanced feedback than simple scoring
- Better handling of complex, multi-step reasoning
- Natural language gradients are human-interpretable
- Iterative changes can be inspected between runs

### Standard Optimization

Traditional iterative refinement with structured suggestions and evaluation.

```bash
# Standard optimization method
pe optimize --prompt "Generate code" --method standard --iterations 3
```

### Hybrid Optimization

Combines two approaches:
1. Standard optimization for initial improvements
2. TextGrad for fine-tuning and advanced refinement

```bash
pe optimize --prompt "Complex reasoning task" --method hybrid --iterations 8
```

## Evaluation Framework

PE supports assertion-based evaluation with string, structure, scoring, and
LLM-judged checks.

### Basic Assertions

```yaml
tests:
  - vars: {text: "Hello world"}
    assert:
      - type: "contains"
        value: "hello"
      - type: "length"
        min: 5
        max: 20
      - type: "matches"
        value: "^[A-Za-z\\s]+$"
```

### Quality-Based Assertions

```yaml
tests:
  - vars: {topic: "AI"}
    assert:
      - type: "readability"
        min: 0.6  # Flesch reading ease
      - type: "sentiment"
        value: "positive"
      - type: "coherence"
        threshold: 0.8
```

### LLM-as-a-Judge Evaluations

```yaml
tests:
  - vars: {question: "Explain quantum computing"}
    assert:
      - type: "llm-judge"
        value: "Response is scientifically accurate and explains concepts clearly"
        threshold: 0.7
      - type: "factuality"
        threshold: 0.9
```

### Performance Assertions

```yaml
tests:
  - vars: {prompt: "Quick summary"}
    assert:
      - type: "latency"
        max: 2.0  # seconds
      - type: "cost"
        max: 0.01  # dollars
      - type: "tokens"
        max: 500
```

### Advanced Structure Assertions

```yaml
tests:
  - vars: {data: "user input"}
    assert:
      - type: "json"
        config:
          schema: "response_schema.json"
      - type: "code"
        config:
          language: "python"
        threshold: 0.8
      - type: "sql"
        config:
          dialect: "postgresql"
```

## 📊 Advanced Analytics and Benchmarking

### Statistical Analysis

PE provides local text metrics, benchmarking helpers, and metrics-package
statistical primitives. Command-level statistical workflows are still future
work.

```bash
# Comprehensive benchmarking
pe benchmark config.yaml --iterations 100 --confidence 0.95

# Compare two result files
pe diff baseline.json variant.json

# Performance regression detection
pe diff historical.json current.json --regression-threshold 0.1
```

### Real-time Monitoring

```bash
# Live performance monitoring
pe eval config.yaml --stream | pe analyze --metric latency --live-plot

# Cost optimization analysis
pe eval config.yaml | pe analyze --metric cost-per-quality --percentiles 50,90,95,99
```

## 🔄 Pipeline Processing and Unix Composability

PE's Unix-style pipeline commands enable powerful data processing workflows.

### Stream Processing

```bash
# Real-time evaluation pipeline
pe eval large-dataset.yaml --stream | \
  pe filter --success | \
  pe stream --select response,score,latency | \
  pe analyze --metric score --group-by provider

# Quality monitoring pipeline
pe eval prompts.yaml | \
  pe filter --min-score 0.8 | \
  pe stats --group-by model --format table
```

### Advanced Filtering

```bash
# Multi-criteria filtering
pe eval config.yaml | \
  pe filter --success --min-score 0.7 --max-latency 2000 | \
  pe analyze --metric cost

# Outlier detection
pe eval benchmark.yaml | \
  pe filter --remove-outliers --percentile 95 | \
  pe stats
```

## 🛡️ Security and Red-teaming

PE includes built-in security testing capabilities for responsible AI deployment.

```bash
# Red-team testing command is not implemented in the current CLI
pe redteam config.yaml --categories harmful,biased,hallucination,privacy

# Toxicity detection
pe eval config.yaml --assertions toxicity.yaml

# Bias evaluation
pe eval config.yaml --bias-tests demographic,gender,racial
```

## 🎯 Interactive Development

### REPL Mode

PE's interactive REPL provides the fastest prompt development experience.

```bash
# Start interactive session
pe interactive --provider anthropic:claude-3-sonnet

# Commands available in REPL:
# /optimize - optimize current prompt
# /eval - evaluate against test cases  
# /save - save current session
# /load - load previous session
# /providers - list available providers
# /help - show all commands
```

### Watch Mode

Automatic re-evaluation on file changes for rapid iteration.

```bash
# Watch for changes
pe watch config.yaml --include "*.yaml,prompts/**/*"

# Watch with custom output
pe watch config.yaml --output results.json --debounce 500ms
```

## 🔧 Extensibility and Custom Providers

### Custom Provider Integration

```go
type CustomProvider struct {
    endpoint string
    apiKey   string
}

func (p *CustomProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (*GenerateResponse, error) {
    // Custom implementation
}

// Register the provider
providers.Register("custom", func(config map[string]interface{}) (Provider, error) {
    return &CustomProvider{
        endpoint: config["endpoint"].(string),
        apiKey:   config["api_key"].(string),
    }, nil
})
```

### Custom Assertions

```go
type CustomAssertion struct {
    criteria string
}

func (a *CustomAssertion) Evaluate(output string) (*AssertionResult, error) {
    // Custom evaluation logic
    return &AssertionResult{
        Passed: true,
        Score:  0.85,
        Message: "Custom assertion passed",
    }, nil
}
```

## 📈 Performance Optimization

### Concurrency Control

```bash
# Control concurrent evaluations
pe eval config.yaml --max-concurrency 10 --timeout 30s

# Batch processing optimization
pe eval large-dataset.yaml --batch-size 50 --parallel-batches 4
```

### Caching and Optimization

```bash
# Enable response caching
pe eval config.yaml --cache --cache-ttl 3600

# Profile performance
pe profile --cpu --memory eval config.yaml
```

## 🌐 Integration and Deployment

### CI/CD Integration

```yaml
# GitHub Actions example
name: Prompt Tests
on: [push, pull_request]

jobs:
  test-prompts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install PE
        run: go install github.com/tmc/pe/cmd/pe@latest
      - name: Run comprehensive tests
        run: |
          pe eval prompts/config.yaml --save-db
          pe benchmark prompts/benchmark.yaml --format json > benchmark.json
          pe diff baseline.json benchmark.json --threshold 0.05
```

### Docker Integration

```dockerfile
FROM golang:1.21-alpine AS builder
RUN go install github.com/tmc/pe/cmd/pe@latest

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/pe /usr/local/bin/
ENTRYPOINT ["pe"]
```

## 🎓 Best Practices

### Prompt Development Workflow

1. **Start with REPL**: Use interactive mode for rapid prototyping
2. **Add Assertions**: Define comprehensive test cases
3. **Optimize**: Use TextGrad or hybrid methods for improvement
4. **Benchmark**: Compare against alternatives
5. **Monitor**: Set up continuous evaluation in production

### Performance Guidelines

- Use `--max-concurrency` to balance speed and resource usage
- Enable caching for repeated evaluations
- Use streaming for large datasets
- Profile regularly to identify bottlenecks

### Security Considerations

- Always test for toxicity and bias before production
- Use red-teaming for high-risk applications
- Monitor cost and token usage to prevent abuse
- Validate structured outputs (JSON, SQL, etc.)

## 🔄 Migration from Other Tools

### From promptfoo

PE is fully compatible with promptfoo configurations:

```bash
# Direct migration
pe eval promptfoo-config.yaml

# Convert and enhance
pe convert promptfoo-config.yaml pe-config.yaml
pe optimize --config pe-config.yaml --method textgrad
```

### From LangSmith

```bash
# Export from LangSmith and import to PE
pe import langsmith-export.json --format pe-config
pe eval pe-config.yaml --save-db
```

### From Custom Tools

PE can convert selected external configuration formats into PE evaluation
configuration:

```bash
# Custom format conversion
pe convert custom-config.json pe-config.yaml --input-format custom
pe eval pe-config.yaml
```

---

Use this guide as a map of implemented advanced surfaces. Check each command's
help output before scripting experimental commands.
