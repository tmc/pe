# PE: Prompt Engineering Toolkit Documentation

## Overview

PE is a comprehensive, production-ready toolkit for prompt engineering that rivals industry leaders like promptfoo, PromptLayer, and Mirascope. Built with Go's performance and Unix philosophy in mind, PE provides a complete solution for developing, testing, and deploying LLM applications.

## Table of Contents

- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Commands Reference](#commands-reference)
- [Configuration Guide](#configuration-guide)
- [Best Practices](#best-practices)
- [Advanced Features](#advanced-features)
- [Integration Guide](#integration-guide)
- [Examples](#examples)

## Quick Start

### Installation

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Your First Evaluation

1. Create a configuration file:
```bash
pe init my-first-eval.yaml
```

2. Run the evaluation:
```bash
pe eval my-first-eval.yaml
```

3. View results in your browser:
```bash
pe view
```

## Core Concepts

### Prompts
Prompts are the core input to language models. PE supports:
- **Template variables**: Use `{{variable}}` syntax for dynamic content
- **Multi-prompt testing**: Test multiple prompt variations simultaneously  
- **File-based prompts**: Reference external prompt files
- **Prompt composition**: Build complex prompts from reusable components

### Providers
PE supports all major LLM providers:
- **OpenAI**: GPT-4, GPT-3.5-turbo, and all variants
- **Anthropic**: Claude-3 (Opus, Sonnet, Haiku)
- **Google AI**: Gemini Pro, Gemini Ultra
- **Local models**: Via Ollama, LM Studio, or custom endpoints

### Tests
Tests define the scenarios and assertions for your prompts:
- **Variable substitution**: Test different input combinations
- **Assertions**: Verify outputs meet quality criteria
- **Success metrics**: Define what constitutes a successful response

### Results
PE provides comprehensive evaluation results:
- **Pass/fail status**: Clear indication of test outcomes
- **Performance metrics**: Latency, token usage, costs
- **Detailed outputs**: Full model responses for analysis
- **Statistical analysis**: Confidence intervals and significance testing

## Commands Reference

### Evaluation Commands

#### `pe eval`
Run prompt evaluations against configured providers and tests.

```bash
# Basic usage
pe eval config.yaml

# Save results to file
pe eval config.yaml -o results.json

# Save to database for viewing
pe eval config.yaml --save-db

# Dry run (show commands without executing)
pe eval config.yaml --dry-run

# Control concurrency
pe eval config.yaml --max-concurrency 8

# Set timeout
pe eval config.yaml --timeout 60s
```

#### `pe view`
View evaluation results in an interactive browser interface.

```bash
# View latest results
pe view

# View specific evaluation
pe view eval-123

# View from file
pe view -f results.json
```

### Development Commands

#### `pe interactive`
Start an interactive REPL for rapid prompt development.

```bash
# Start with default provider
pe interactive

# Use specific provider
pe interactive --provider anthropic:claude-3-sonnet

# Load configuration
pe interactive --config my-config.yaml
```

#### `pe watch`
Monitor files for changes and automatically re-run evaluations.

```bash
# Watch current directory
pe watch config.yaml

# Watch specific patterns
pe watch config.yaml --include "*.yaml,prompts/**/*"

# Save results on each run
pe watch config.yaml -o results.json
```

### Pipeline Commands

#### `pe ask`
Ask a single question to an LLM provider (pipeline-friendly).

```bash
# Basic usage
echo "What is AI?" | pe ask --provider openai:gpt-4

# With options
pe ask --provider anthropic:claude-3-haiku --temperature 0.7 --max-tokens 100
```

#### `pe stream`
Process evaluation results as a stream for Unix composability.

```bash
# Extract specific fields
pe eval config.yaml | pe stream --select response,latency

# Filter and analyze
pe eval config.yaml | pe stream --select cost | pe analyze --metric cost
```

#### `pe filter`
Filter evaluation results based on conditions.

```bash
# Show only successful tests
pe eval config.yaml | pe filter --success

# Filter by criteria
pe eval config.yaml | pe filter --latency "<1s" --cost "<0.01"
```

#### `pe analyze`
Analyze evaluation results with statistical methods.

```bash
# Analyze latency
pe eval config.yaml | pe analyze --metric latency

# Generate statistical report
pe eval config.yaml | pe analyze --report --confidence 0.95
```

#### `pe stats`
Show quick statistics from evaluation results.

```bash
# Basic stats
pe eval config.yaml | pe stats

# Detailed breakdown
pe eval config.yaml | pe stats --detailed
```

#### `pe diff`
Compare two evaluation results.

```bash
# Compare files
pe diff results1.json results2.json

# Compare evaluations
pe diff eval-123 eval-456

# Show only differences
pe diff results1.json results2.json --changes-only
```

### Utility Commands

#### `pe fmt`
Format and validate configuration files.

```bash
# Format file
pe fmt config.yaml

# Convert between formats
pe fmt config.yaml --output json

# Write back to file
pe fmt config.yaml --write
```

#### `pe vet`
Validate configuration files for correctness.

```bash
# Validate single file
pe vet config.yaml

# Validate multiple files
pe vet *.yaml

# Validate from stdin
cat config.yaml | pe vet
```

#### `pe convert`
Convert configuration files between formats.

```bash
# YAML to JSON
pe convert config.yaml config.json

# JSON to YAML
pe convert config.json config.yaml --output yaml
```

#### `pe benchmark`
Compare performance metrics across prompts and providers.

```bash
# Basic benchmark
pe benchmark benchmark-config.yaml

# Multiple iterations
pe benchmark config.yaml --iterations 10 --concurrency 4

# Output formats
pe benchmark config.yaml --format json -o benchmark-results.json
```

## Configuration Guide

### Basic Configuration

```yaml
# my-config.yaml
prompts:
  - "What is the capital of {{country}}?"
  - "Tell me about {{country}}'s capital city."

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-sonnet"

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

### Advanced Configuration

```yaml
# advanced-config.yaml
description: "Advanced prompt evaluation example"

prompts:
  - id: "basic-prompt"
    content: "What is the capital of {{country}}?"
  - id: "detailed-prompt"
    content: |
      You are a geography expert. When asked about a country's capital,
      provide the name and 2-3 interesting facts about the city.
      
      Country: {{country}}
      Question: What is the capital and tell me about it?

providers:
  - id: "gpt4"
    type: "openai"
    model: "gpt-4"
    config:
      temperature: 0.1
      max_tokens: 200
  - id: "claude"
    type: "anthropic"
    model: "claude-3-sonnet-20240229"
    config:
      temperature: 0.1
      max_tokens: 200

tests:
  - description: "European capitals"
    vars:
      country: "France"
    assert:
      - type: "contains"
        value: "Paris"
      - type: "length"
        min: 10
        max: 500
      - type: "cost"
        max: 0.01
      - type: "latency"
        max: "5s"
        
  - description: "Asian capitals"
    vars:
      country: "Japan"
    assert:
      - type: "contains"
        value: "Tokyo"
      - type: "not-contains"
        value: ["Kyoto", "Osaka"]  # Common mistakes
      - type: "regex"
        pattern: "Tokyo.*Japan"

# Advanced features
redteaming:
  enabled: true
  categories: ["harmful", "biased", "hallucination"]
  
metrics:
  - name: "accuracy"
    type: "custom"
    script: "./accuracy-scorer.py"
  - name: "helpfulness"
    type: "llm-graded"
    judge: "openai:gpt-4"
    criteria: "Rate the helpfulness of the response (1-5)"

output:
  format: "json"
  include_raw: true
  include_costs: true
  include_metadata: true
```

## Best Practices

### 1. Test-Driven Prompt Development

Start with failing tests, then improve your prompts:

```bash
# Create test cases first
pe init my-tests.yaml

# Run and see failures
pe eval my-tests.yaml

# Iteratively improve prompts
pe watch my-tests.yaml  # Auto-rerun on changes
```

### 2. Use Version Control

Track prompt changes over time:

```bash
# Initialize git tracking
git init
git add my-config.yaml
git commit -m "Initial prompt configuration"

# Track changes
pe eval my-config.yaml --save-db
git add . && git commit -m "Improved accuracy by 15%"
```

### 3. Monitor Costs and Performance

```bash
# Track costs over time
pe eval config.yaml | pe stream --select cost,latency | pe analyze --trend

# Set budget limits
pe eval config.yaml | pe filter --cost "<0.10" | pe stats
```

### 4. Use Multiple Providers

Compare different models systematically:

```yaml
providers:
  - "openai:gpt-4"           # High quality, expensive
  - "openai:gpt-3.5-turbo"   # Good quality, cheaper  
  - "anthropic:claude-3-haiku"  # Fast, cost-effective
```

### 5. Comprehensive Testing

Test edge cases and failure modes:

```yaml
tests:
  # Happy path
  - vars: {country: "France"}
    assert: [{type: "contains", value: "Paris"}]
    
  # Edge cases  
  - vars: {country: "Vatican City"}
    assert: [{type: "contains", value: "Vatican"}]
    
  # Error handling
  - vars: {country: "NonexistentPlace"}
    assert: [{type: "not-contains", value: "capital"}]
```

## Advanced Features

### Red-Team Testing

PE includes built-in red-team testing for safety and security:

```yaml
redteaming:
  enabled: true
  categories:
    - "harmful"           # Harmful content generation
    - "biased"           # Bias and discrimination
    - "hallucination"    # Factual inaccuracies
    - "prompt_injection" # Prompt injection attacks
    - "jailbreak"        # Jailbreak attempts
    
  custom_tests:
    - category: "custom"
      prompts: "./custom-redteam-prompts.txt"
```

### Custom Metrics

Define custom scoring functions:

```yaml
metrics:
  - name: "sentiment_score"
    type: "python"
    script: |
      from textblob import TextBlob
      def score(output):
          sentiment = TextBlob(output).sentiment.polarity
          return {"sentiment": sentiment, "pass": sentiment >= 0}
          
  - name: "fact_check"
    type: "llm-graded"
    judge: "openai:gpt-4"
    prompt: |
      Fact-check the following response for accuracy.
      Rate 1-5 where 5 is completely accurate.
      Response: {{output}}
```

### Multi-Turn Conversations

Test conversation flows:

```yaml
tests:
  - description: "Multi-turn conversation"
    conversation:
      - user: "Hello, I'm planning a trip to France."
        assistant_assert: 
          - type: "contains"
            value: "France"
      - user: "What's the capital?"
        assistant_assert:
          - type: "contains" 
            value: "Paris"
      - user: "Tell me about the weather there."
        assistant_assert:
          - type: "contains"
            value: ["weather", "climate", "temperature"]
```

### Streaming Evaluation

For real-time testing:

```bash
# Stream results as they complete
pe eval config.yaml --stream | pe filter --success --stream | pe stats --live

# Monitor in real-time
pe eval large-test-suite.yaml --stream | pe analyze --metric latency --live-plot
```

## Integration Guide

### CI/CD Integration

#### GitHub Actions

```yaml
# .github/workflows/prompt-tests.yml
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
          pe eval prompts/config.yaml -o results.json
          pe stats results.json
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

#### GitLab CI

```yaml
# .gitlab-ci.yml
prompt_tests:
  stage: test
  script:
    - go install github.com/tmc/pe/cmd/pe@latest
    - pe eval config.yaml --save-db
    - pe view --export-report report.html
  artifacts:
    reports:
      junit: report.xml
    paths:
      - report.html
```

### API Integration

Use PE programmatically:

```go
package main

import (
    "context"
    "github.com/tmc/pe/internal/evaluator"
    "github.com/tmc/pe/internal/promptfoo"
)

func main() {
    config := promptfoo.Config{
        Prompts: []string{"What is {{topic}}?"},
        Providers: []string{"openai:gpt-4"},
        Tests: []promptfoo.TestCase{
            {
                Vars: map[string]interface{}{"topic": "AI"},
                Assert: []promptfoo.Assertion{
                    {Type: "contains", Value: "artificial intelligence"},
                },
            },
        },
    }
    
    results, err := evaluator.Evaluate(config, time.Minute, false, 4, true)
    if err != nil {
        panic(err)
    }
    
    // Process results...
}
```

## Examples

### Example 1: Content Generation

```yaml
# content-generation.yaml
description: "Test content generation prompts"

prompts:
  - id: "blog-post"
    content: |
      Write a blog post about {{topic}} that is {{tone}} in tone.
      Include an introduction, 3 main points, and a conclusion.
      Target length: {{length}} words.

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-opus"

tests:
  - description: "Technical blog post"
    vars:
      topic: "machine learning"
      tone: "professional"
      length: 500
    assert:
      - type: "length"
        min: 400
        max: 600
      - type: "contains"
        value: ["introduction", "conclusion"]
      - type: "word_count"
        min: 450
        max: 550
      - type: "readability"
        min_grade_level: 8
        max_grade_level: 12
```

### Example 2: Code Generation

```yaml
# code-generation.yaml
description: "Test code generation capabilities"

prompts:
  - |
    Generate a {{language}} function that {{task}}.
    Include proper error handling and documentation.
    Function name: {{function_name}}

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-sonnet"

tests:
  - vars:
      language: "Python"
      task: "calculates the factorial of a number"
      function_name: "factorial"
    assert:
      - type: "contains"
        value: ["def factorial", "return", "if"]
      - type: "syntax_valid"
        language: "python"
      - type: "security_scan"
        rules: ["no-eval", "no-exec"]
```

### Example 3: Translation

```yaml
# translation.yaml
description: "Test translation accuracy"

prompts:
  - "Translate the following {{source_lang}} text to {{target_lang}}: {{text}}"

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-sonnet"

tests:
  - vars:
      source_lang: "English"
      target_lang: "French" 
      text: "Hello, how are you?"
    assert:
      - type: "contains"
        value: "Bonjour"
      - type: "translation_quality"
        threshold: 0.8
        reference: "Bonjour, comment allez-vous ?"
```

For more examples, see the [examples directory](../example/) in the repository.

---

## Getting Help

- **Documentation**: [docs/](.)
- **Examples**: [example/](../example/)
- **Issues**: [GitHub Issues](https://github.com/tmc/pe/issues)
- **Discussions**: [GitHub Discussions](https://github.com/tmc/pe/discussions)

## Contributing

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see [LICENSE](../LICENSE) for details.