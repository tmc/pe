# PE Command Examples Guide

A comprehensive guide with practical examples for every PE command. Learn by doing!

## 🎯 Core Commands

### `pe run` - Execute Prompts Immediately

The most frequently used command for quick prompt execution.

#### Basic Usage
```bash
# Simple text prompt
pe run "What are the benefits of renewable energy?"

# With the default provider
pe run "Explain quantum computing"

# With template variables
pe run "Translate '{{.text}}' to {{.language}}" --var text="Hello world" --var language="Spanish"
```

#### File-based Prompts
```bash
# Execute prompt from file
pe run my-prompt.prompt

# With environment variable substitution
export ANALYSIS_TYPE="sentiment"
pe run "Perform {{.ANALYSIS_TYPE}} analysis on: {{.text}}" --var text="I love this product!"
```

#### Advanced Options
```bash
# Stream responses in real-time
pe run creative-writing.prompt --stream --var topic="space exploration"

# Save output with shell redirection
pe run report.prompt --var data="$(cat input.csv)" > results.txt
```

### `pe eval` - Comprehensive Evaluation

Systematic testing and validation of prompts.

#### Basic Evaluation
```bash
# Simple evaluation config
cat > basic-eval.yaml << 'EOF'
prompts:
  - "Summarize in one sentence: {{.text}}"

tests:
  - vars:
      text: "Artificial intelligence is transforming industries worldwide."
    assert:
      - type: contains
        value: "AI"
      - type: max-length
        value: 100
EOF

pe eval basic-eval.yaml
```

#### Advanced Evaluation Scenarios

**Pass@N for Code Generation:**
```bash
cat > code-eval.yaml << 'EOF'
prompts:
  - "Write a Python function to {{.task}}"

tests:
  - vars:
      task: "calculate factorial of a number"
    assert:
      - type: pass-at-n
        config:
          n: 3
          samples: 10
          test_cases:
            - input: "factorial(5)"
              expected: "120"
            - input: "factorial(0)"
              expected: "1"
        threshold: 0.8
EOF

pe eval code-eval.yaml --output code-results.json
```

**Structured Output Validation:**
```bash
cat > structured-eval.yaml << 'EOF'
prompts:
  - "Extract entities from: {{.text}}"

tests:
  - vars:
      text: "Apple Inc. was founded by Steve Jobs in Cupertino, California."
    assert:
      - type: structured-output
        config:
          format: json
          schema:
            type: object
            properties:
              companies:
                type: array
                items:
                  type: string
              people:
                type: array
                items:
                  type: string
              locations:
                type: array
                items:
                  type: string
            required: ["companies", "people", "locations"]
EOF

pe eval structured-eval.yaml --output results.json
```

**Multi-Provider Comparison:**
```bash
# Compare providers by listing providers in the config file
pe eval comparison.yaml --output comparison-results.json
```

### `pe benchmark` - Performance Testing

Measure and optimize prompt performance.

```bash
# Basic benchmarking
pe benchmark benchmark-config.yaml --iterations 20

# Comprehensive performance analysis
pe benchmark complex-eval.yaml \
  --iterations 50 \
  --output benchmark-results.json

# Stress testing
pe benchmark load-test.yaml \
  --concurrency 10 \
  --iterations 20

# Go benchmark format
pe benchmark cost-analysis.yaml --go-bench
```

## 🔄 Pipeline Commands

### `pe ask` - Single Query in Pipeline

```bash
# Simple pipeline
echo "Raw customer feedback" | pe ask "Extract sentiment and key issues"

# With a template
echo "Product review text" | pe ask --template "Analyze sentiment: {{.}}"

# JSON processing pipeline
echo '{"review": "Great product!"}' | pe ask --template "Rate this review from 1-10: {{.}}"
```

### `pe stream` - Process Result Streams

```bash
# Stream evaluation results
pe eval large-eval.yaml | pe stream > streaming-results.txt

# Real-time filtering
pe eval continuous-eval.yaml | pe stream | pe filter --contains pass

# Live monitoring
pe eval production-test.yaml | pe stream
```

### `pe filter` - Result Filtering

```bash
# Filter lines containing pass
pe eval test-suite.yaml | pe filter --contains pass

# Filter by pattern or substring
pe eval quality-test.yaml | pe filter --contains pass

# Filter and transform
pe eval results.yaml | pe filter --contains pass | pe filter --transform lowercase
```

### `pe analyze` - Statistical Analysis

```bash
# Basic analysis
pe eval dataset.yaml | pe analyze

# Select metrics and output format
pe eval performance.yaml | pe analyze \
  --metrics readability,sentiment \
  --format json
```

## 🎨 Composition and Optimization

### `pe experimental compose` - Prompt Composition

```bash
# Simple composition
pe experimental compose base-system.prompt task-specific.prompt \
  --style structured \
  --output composed.prompt

# Style-specific composition
pe experimental compose --style cot \
  reasoning-base.prompt \
  math-problem.prompt

# Advanced composition with validation
pe experimental compose \
  --components system.prompt,context.prompt,task.prompt \
  --style dspy \
  --coherence \
  --output composed-prompt.prompt
```

### `pe experimental optimize` - Metaprompting Optimization

```bash
# TextGrad optimization
pe experimental optimize --prompt "Summarize this text: {{.text}}" \
  --method textgrad \
  --iterations 5

# PE2 optimization
pe experimental optimize --prompt "Classify sentiment: {{.text}}" \
  --method pe2 \
  --iterations 5

# APEX optimization
pe experimental optimize --prompt "Review this code: {{.code}}" \
  --method apex \
  --iterations 5
```

### `pe experimental semantic` - Advanced Optimization

```bash
# Semantic backpropagation
pe experimental semantic backprop \
  --prompt "Analyze sentiment: {{.text}}" \
  --objective "improve accuracy on edge cases" \
  --iterations 10

# GASO optimization
pe experimental semantic gaso \
  --system multi-component-system.json \
  --objective "overall performance" \
  --multi-objective \
  --output optimized-system.json

# Gradient descent on prompts
pe experimental semantic descent \
  --objective "minimize hallucination" \
  --adaptive \
  --convergence 0.001
```

## 🔧 Development Tools

### `pe fmt` - Code Formatting

```bash
# Format single prompt
pe fmt prompt.txt

# Format prompts in place
pe fmt *.prompt --write --style standard

# Check formatting without changing files
pe fmt prompt.txt --check

# Custom formatting options
pe fmt prompt.txt --style openai --fix
```

### `pe vet` - Configuration Validation

```bash
# Validate single config
pe vet config.yaml

# Validate all configs
pe vet configs/

# Show validation help
pe vet --help
```

### `pe convert` - Format Conversion

```bash
# YAML to JSON
pe convert config.yaml config.json

# Promptfoo to PE format
pe convert promptfoo-config.yaml pe-config.yaml --from promptfoo

# Batch conversion
pe convert --batch input-dir/ output-dir/ --format json

# With validation
pe convert --validate input.yaml output.json
```

### `pe init` - Project Initialization

```bash
# Basic project setup
mkdir my-project
cd my-project
pe init

# Reinitialize an existing .pe directory
pe init --force
```

## 📊 Analysis and Reporting

### `pe stats` - Quick Statistics

```bash
# Basic stats from evaluation
pe eval results.yaml | pe stats

# Stats currently reads evaluation data from stdin.
pe eval results.yaml | pe stats
```

### `pe diff` - Result Comparison

```bash
# Compare two result sets
pe diff baseline-results.json new-results.json

# JSON output for automation
pe diff --format json baseline-results.json new-results.json

# Fail CI on regressions outside explicit thresholds
pe diff --fail-on-regression \
  --max-pass-rate-drop 5 \
  --max-latency-increase-ms 100 \
  baseline-results.json new-results.json
```

### `pe view` - Interactive Visualization

```bash
# View results in browser
pe view --file results.json

# Choose a port
pe view --file results.json --port 8081
```

## 🔐 Security and Attestation

### `pe exp attest` - Unsigned Local Manifests

```bash
# Create a deterministic unsigned manifest
pe exp attest manifest prompts/ > manifest.json

# Verify files against the manifest
pe exp attest verify manifest.json
```

### `pe exp cache` - Local Content Cache

```bash
# Print a file cache key
pe exp cache key prompt.txt

# Store and verify a file by SHA-256
pe exp cache put prompt.txt
pe exp cache verify <sha256>

# Store a canonical unsigned manifest
pe exp cache manifest put manifest.json
```

### `pe security` - Security Testing

```bash
# OWASP LLM Top 10 testing
pe security scan prompt-file.prompt

# Category-specific security test
pe security test prompt-file.prompt --category prompt_injection

# Red team testing
pe security redteam \
  --target prompt-file.prompt \
  --duration 1h \
  --output security-report.json
```

## 🌐 Distributed Computing

### `pe exp distributed` - Distributed Execution

```bash
# Inspect distributed prototype command surface
pe exp distributed --help

# Re-check prototype interface
pe exp distributed --help

# Run a larger evaluation with core concurrency controls
pe eval large-test-suite.yaml \
  --max-concurrency 20 \
  --timeout 5m
```

## 📦 Module Management

### `pe mod` - Module Management

```bash
# Initialize module
pe mod init github.com/myorg/prompts

# Download dependencies listed in pe.mod
pe mod download

# List dependencies
pe mod list

# Update dependencies
pe mod tidy

# Create vendor copy
pe mod vendor

# Publish module
pe mod publish
```

### `pe get` - Download Modules

```bash
# Get specific module
pe get github.com/pe-community/nlp-prompts

# Get specific version
pe get github.com/user/prompts@v2.1.0

# Get with custom path
pe get github.com/user/prompts --path custom-location/
```

## 🔌 Plugin Management

### `pe plugin` - Plugin System

```bash
# List available plugins
pe plugin list

# Build and expose the Promptfoo plugin
go build -o ~/bin/pe-promptfoo ./plugins/promptfoo

# Use plugin
pe promptfoo convert config.yaml output.json --direction pe-to-promptfoo

# Update plugins
pe plugin update

# Remove plugin
rm -f ~/bin/pe-promptfoo
```

## 💡 Advanced Workflows

### End-to-End Testing Pipeline

```bash
#!/bin/bash
# complete-testing-pipeline.sh

# 1. Validate configurations
pe vet configs/

# 2. Run unit tests
pe eval unit-tests.yaml --output unit-results.json

# 3. Performance benchmarks
pe benchmark performance-tests.yaml --output perf-results.json

# 4. Security scanning
pe security scan prompts/ --output security-results.json

# 5. Compare with baseline
pe diff --format json baseline-results.json unit-results.json

# 6. Generate final report
pe analyze \
  --input unit-results.json,perf-results.json \
  --template comprehensive-report \
  --output final-report.html

echo "✅ Complete testing pipeline finished!"
echo "📊 View results: open final-report.html"
```

### Production Deployment Workflow

```bash
#!/bin/bash
# deploy-prompts.sh

# 1. Select a prompt variant from local deterministic scores
pe exp optimize --input variants.json --output selected.json

# 2. Validate optimized prompts
pe eval production-validation.yaml \
  --config production-validation.yaml \
  --output validation-results.json

# 3. Create an unsigned local manifest
pe exp attest manifest prompts/ > manifest.json

# 4. Deploy with monitoring
pe exp distributed --help

# 5. Health check
pe eval health-check.yaml --max-concurrency 20 --timeout 5m

echo "🚀 Production deployment complete!"
```

### Research and Development Workflow

```bash
#!/bin/bash
# research-workflow.sh

# 1. Generate prompt variations
pe experimental evolve base-prompt.prompt \
  --generations 10 \
  --population 20 \
  --mutations creative,logical,concise

# 2. Evaluate all variations
pe eval research-eval.yaml \
  --config research-eval.yaml \
  --output research-results.json

# 3. Semantic optimization
pe experimental semantic backprop \
  --best-prompts top-5/ \
  --target "research quality" \
  --iterations 15

# 4. Statistical analysis
pe analyze research-results.json \
  --correlations \
  --feature-importance \
  --output research-insights.html

echo "🔬 Research analysis complete!"
echo "📈 Insights: open research-insights.html"
```

## 🚀 Pro Tips

### Performance Optimization
```bash
# Parallel execution
pe eval large-test-suite.yaml --max-concurrency 8

# Resource limits
pe eval heavy-eval.yaml --timeout 5m
```

### Development Efficiency
```bash
# Watch for changes during development
pe watch --help

# Quick validation
pe vet configs/

# Interactive debugging
pe interactive --help
```

### Production Best Practices
```bash
# Create unsigned local manifests before deployment
pe exp attest manifest prompts/ > manifest.json

# Monitor benchmark output
pe benchmark production-eval.yaml --format json --output production-benchmark.json

# Graceful degradation
pe eval critical-eval.yaml --timeout 5m --max-concurrency 4
```

This comprehensive guide covers all PE commands with practical, real-world examples. Each command includes basic usage, advanced options, and integration patterns to help you become proficient with PE quickly!
