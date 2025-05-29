# Quick Start Guide

Get up and running with PE in under 5 minutes and experience the most advanced prompt engineering toolkit available.

## 🚀 Installation

```bash
# Install PE (requires Go 1.21+)
go install github.com/tmc/pe/cmd/pe@latest

# Verify installation
pe --version
```

## ⚡ 30-Second Demo

Experience PE's power immediately:

```bash
# Create a basic config
pe init demo.yaml

# Run your first evaluation
pe eval demo.yaml

# Try running a prompt directly
pe run "What is 2+2?"

# Optimize a prompt with cutting-edge TextGrad
pe optimize --prompt "Summarize this text clearly" --method textgrad --iterations 3
```

## 🎯 5-Minute Tutorial

### Step 1: Set Up Your Environment

```bash
# Set your API keys
export OPENAI_API_KEY="your-key-here"
export ANTHROPIC_API_KEY="your-key-here"  # optional

# Create a working directory
mkdir my-prompts && cd my-prompts
```

### Step 2: Create Your First Configuration

```bash
pe init advanced-demo.yaml
```

This creates a configuration with multiple providers and comprehensive testing:

```yaml
prompts:
  - "What is the capital of {{country}}?"
  - "Tell me about the capital city of {{country}}."

providers:
  - "openai:gpt-4"
  - "openai:gpt-3.5-turbo"

tests:
  - vars:
      country: "France"
    assert:
      - type: "contains"
        value: "Paris"
      - type: "latency"
        max: 3.0
      - type: "cost"
        max: 0.01
  - vars:
      country: "Japan"
    assert:
      - type: "contains"
        value: "Tokyo"
      - type: "readability"
        min: 0.6
```

### Step 3: Run Advanced Evaluation

```bash
# Run with detailed output
pe eval advanced-demo.yaml --save-db

# View results in browser
pe view

# Get quick statistics
pe eval advanced-demo.yaml | pe stats --format table
```

### Step 4: Experience TextGrad Optimization

```bash
# Optimize your prompt using cutting-edge TextGrad method
pe optimize \
  --prompt "Explain {{concept}} in simple terms suitable for beginners" \
  --method textgrad \
  --iterations 5 \
  --output optimization-results.json

# Try multi-stage optimization
pe optimize \
  --prompt "Analyze the sentiment of this text: {{text}}" \
  --method multistage \
  --iterations 6
```

### Step 5: Explore Pipeline Processing

PE's Unix-style pipelines are unique in the prompt engineering space:

```bash
# Real-time monitoring pipeline
pe eval advanced-demo.yaml --stream | \
  pe filter --success | \
  pe analyze --metric latency --live

# Cost optimization pipeline
pe eval advanced-demo.yaml | \
  pe filter --max-cost 0.005 | \
  pe stats --group-by provider

# Quality analysis pipeline
pe eval advanced-demo.yaml | \
  pe stream --select response,score,provider | \
  pe filter --min-score 0.8 | \
  pe analyze --metric score --percentiles 50,90,95,99
```

## 🎪 Interactive Development

### Interactive Mode

```bash
# Start interactive session
pe interactive

# Or use the run command for quick prompts
pe run "Your prompt here"
pe run prompt.txt --stream
```

### Watch Mode (Auto-reload)

```bash
# Watch for changes and auto-evaluate
pe watch advanced-demo.yaml

# Watch with custom patterns
pe watch advanced-demo.yaml --include "*.yaml,prompts/**/*,tests/**/*"
```

## 📊 Advanced Features Preview

### Comprehensive Assertions

```yaml
# Create advanced-assertions.yaml
tests:
  - vars:
      text: "AI will revolutionize healthcare"
    assert:
      # Quality assessments
      - type: "sentiment"
        value: "positive"
      - type: "readability" 
        min: 0.6
      - type: "coherence"
        threshold: 0.8
      
      # LLM-as-a-judge
      - type: "llm-judge"
        value: "Response is factually accurate and well-reasoned"
        threshold: 0.7
      
      # Performance constraints
      - type: "latency"
        max: 2.0
      - type: "tokens"
        max: 200
      - type: "cost"
        max: 0.005
      
      # Structure validation
      - type: "json"
        config:
          schema: "response_schema.json"
```

### Benchmarking and A/B Testing

```bash
# Comprehensive benchmark
pe benchmark advanced-demo.yaml \
  --iterations 50 \
  --concurrency 5 \
  --format json \
  --output benchmark-results.json

# A/B testing with statistical significance
pe diff baseline-results.json current-results.json \
  --significance-test \
  --confidence 0.95 \
  --metric score
```

### Security Testing

```bash
# Built-in security testing
pe security test --owasp --target advanced-demo.yaml

# Advanced red-teaming
pe security redteam --comprehensive --target advanced-demo.yaml
```

## 🛠️ Development Workflow

### 1. Rapid Prototyping

```bash
# Use run command for quick iteration
pe run "Your prompt here" --provider cgpt

# Test with streaming
pe run "Tell me a story" --stream

# Test with variables
pe run "Translate {{text}} to {{language}}" --var text="Hello" --var language="Spanish"
```

### 2. Structured Development

```bash
# Create comprehensive test suite
pe init production-config.yaml

# Add multiple test cases and assertions
# Edit production-config.yaml

# Evaluate with detailed metrics
pe eval production-config.yaml --save-db
```

### 3. Optimization

```bash
# Try different optimization methods
pe optimize --prompt "Your prompt" --method textgrad --iterations 5  
pe optimize --prompt "Your prompt" --method multistage --iterations 3
pe optimize --prompt "Your prompt" --method reflection --iterations 4
```

### 4. Production Testing

```bash
# Comprehensive benchmarking
pe benchmark production-config.yaml --iterations 100

# Performance profiling
pe profile cpu --duration 30s
pe profile memory

# Regression testing
pe diff baseline.json current.json --threshold 0.05
```

## 📁 Example Configurations

### Content Generation

```yaml
# content-generation.yaml
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
        min: 0.6
      - type: "llm-judge"
        value: "Well-structured blog post with clear main points"
        threshold: 0.8
```

### Multi-Provider Comparison

```yaml
# provider-comparison.yaml
prompts:
  - "Explain quantum computing in simple terms"

providers:
  - "openai:gpt-4"           # High quality, expensive
  - "openai:gpt-3.5-turbo"   # Good quality, cheaper

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
        max: 10.0
      - type: "readability"
        min: 0.6
```

## 🚀 Next Steps

### Learn Advanced Features

- **Optimization Methods**: See [docs/ADVANCED_OPTIMIZATION_GUIDE.md](ADVANCED_OPTIMIZATION_GUIDE.md)
- **API Reference**: See [docs/API_REFERENCE.md](API_REFERENCE.md)
- **Examples**: Explore the [example/](../example/) directory

### Plugin System

```bash
# List available plugins
pe plugin list

# Use the promptfoo plugin for compatibility
pe promptfoo import legacy-config.yaml
```

### Community and Support

- **Documentation**: [docs/](docs/)
- **Examples**: [example/](example/)  
- **Issues**: [GitHub Issues](https://github.com/tmc/pe/issues)
- **Discussions**: [GitHub Discussions](https://github.com/tmc/pe/discussions)

## 💡 Pro Tips

1. **Start Simple**: Use `pe run` for quick prompt testing
2. **Use TextGrad**: For complex prompts, TextGrad optimization is highly effective
3. **Pipeline Everything**: Leverage Unix-style pipes for powerful data processing
4. **Watch Mode**: Enable auto-reload during development for immediate feedback
5. **Comprehensive Testing**: Use multiple assertion types for robust evaluation
6. **Benchmark Regularly**: Compare providers and track performance over time
7. **Try Semantic Backprop**: For challenging optimization tasks, use `pe semantic`

---

**You're now ready to experience the most advanced prompt engineering toolkit available. PE combines cutting-edge research with practical engineering to deliver capabilities that exceed any other tool in the market.**