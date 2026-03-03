# PE Command Examples Guide

A comprehensive guide with practical examples for every PE command. Learn by doing!

## 🎯 Core Commands

### `pe run` - Execute Prompts Immediately

The most frequently used command for quick prompt execution.

#### Basic Usage
```bash
# Simple text prompt
pe run "What are the benefits of renewable energy?"

# With specific provider and model
pe run "Explain quantum computing" --provider openai --model gpt-4

# With template variables
pe run "Translate '{{.text}}' to {{.language}}" --var text="Hello world" --var language="Spanish"
```

#### File-based Prompts
```bash
# Execute prompt from file
pe run my-prompt.prompt

# With variables from file
pe run analyze.prompt --vars-file data.json

# With environment variable substitution
export ANALYSIS_TYPE="sentiment"
pe run "Perform {{.ANALYSIS_TYPE}} analysis on: {{.text}}" --var text="I love this product!"
```

#### Advanced Options
```bash
# Control output format
pe run summarize.prompt --output json --var text="Long document text..."

# Stream responses in real-time
pe run creative-writing.prompt --stream --var topic="space exploration"

# Save output to file
pe run report.prompt --output results.txt --var data="$(cat input.csv)"

# Multiple providers for comparison
pe run "Explain machine learning" --providers openai,anthropic,ollama
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

pe eval code-eval.yaml --provider openai --model gpt-4-turbo
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
# Compare providers on same tasks
pe eval comparison.yaml --providers openai,anthropic,ollama --compare-providers

# Generate comparison report
pe eval comparison.yaml --providers openai,anthropic --report comparison-report.html
```

### `pe benchmark` - Performance Testing

Measure and optimize prompt performance.

```bash
# Basic benchmarking
pe benchmark simple-prompt.prompt --iterations 20

# Comprehensive performance analysis
pe benchmark complex-eval.yaml \
  --iterations 50 \
  --providers openai,anthropic \
  --metrics latency,tokens-per-second,cost \
  --output benchmark-results.json

# Stress testing
pe benchmark load-test.yaml \
  --concurrent 10 \
  --duration 5m \
  --ramp-up 30s

# Cost analysis
pe benchmark cost-analysis.yaml \
  --providers openai,anthropic \
  --metrics cost-per-token,total-cost \
  --budget-limit 10.00
```

## 🔄 Pipeline Commands

### `pe ask` - Single Query in Pipeline

```bash
# Simple pipeline
echo "Raw customer feedback" | pe ask "Extract sentiment and key issues"

# With specific configuration
echo "Product review text" | pe ask \
  --prompt "Analyze sentiment: {{.input}}" \
  --provider anthropic \
  --model claude-3-sonnet-20240229

# JSON processing pipeline
echo '{"review": "Great product!"}' | pe ask \
  --prompt "Rate this review from 1-10: {{.input}}" \
  --input-format json \
  --output-format json
```

### `pe stream` - Process Result Streams

```bash
# Stream evaluation results
pe eval large-eval.yaml | pe stream --output streaming-results.jsonl

# Real-time filtering
pe eval continuous-eval.yaml | pe stream --filter "score > 0.8"

# Live monitoring
pe eval production-test.yaml | pe stream --monitor --alert-threshold 0.5
```

### `pe filter` - Result Filtering

```bash
# Filter successful results only
pe eval test-suite.yaml | pe filter --success

# Filter by score threshold
pe eval quality-test.yaml | pe filter --score-min 0.7

# Complex filtering with expressions
pe eval comprehensive.yaml | pe filter \
  --where "score > 0.8 AND latency < 2000 AND provider == 'openai'"

# Filter and transform
pe eval results.yaml | pe filter --success | pe filter --transform "extract_metrics"
```

### `pe analyze` - Statistical Analysis

```bash
# Basic analysis
pe eval dataset.yaml | pe analyze

# Detailed statistical report
pe eval performance.yaml | pe analyze \
  --metrics score,latency,cost \
  --group-by provider \
  --output analysis-report.html

# Trend analysis
pe eval time-series.yaml | pe analyze \
  --time-series \
  --window 24h \
  --trend-detection

# Custom analysis
pe eval custom.yaml | pe analyze \
  --script custom-analysis.py \
  --params threshold=0.8,window=100
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

### `pe optimize` - Metaprompting Optimization

```bash
# Basic optimization
pe optimize "Summarize this text: {{.text}}" \
  --target "conciseness and accuracy" \
  --iterations 5

# Advanced semantic optimization
pe optimize complex-prompt.prompt \
  --method semantic-backprop \
  --target "accuracy,latency,cost" \
  --eval-config validation.yaml \
  --iterations 10

# Multi-objective optimization
pe optimize multi-task.prompt \
  --objectives accuracy:0.4,speed:0.3,cost:0.3 \
  --method gaso \
  --convergence-threshold 0.001
```

### `pe semantic` - Advanced Optimization

```bash
# Semantic backpropagation
pe semantic backprop \
  --prompt "Analyze sentiment: {{.text}}" \
  --target "improve accuracy on edge cases" \
  --learning-rate 0.1 \
  --iterations 10

# GASO optimization
pe semantic gaso \
  --system multi-component-system.json \
  --objective "overall performance" \
  --multi-objective \
  --output optimized-system.json

# Gradient descent on prompts
pe semantic descent \
  --objective "minimize hallucination" \
  --adaptive \
  --convergence 0.001
```

## 🔧 Development Tools

### `pe fmt` - Code Formatting

```bash
# Format single config
pe fmt config.yaml

# Format all configs in directory
pe fmt configs/

# Check formatting without changing files
pe fmt --check --diff configs/

# Custom formatting options
pe fmt --style compact --sort-keys config.yaml
```

### `pe vet` - Configuration Validation

```bash
# Validate single config
pe vet config.yaml

# Validate all configs
pe vet configs/

# Strict validation
pe vet --strict --fail-on-warnings config.yaml

# Custom validation rules
pe vet --rules custom-rules.yaml config.yaml
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
pe init my-project

# With template
pe init my-project --template evaluation-suite

# Custom configuration
pe init my-project \
  --providers openai,anthropic \
  --template advanced \
  --with-examples

# Interactive setup
pe init --interactive
```

## 📊 Analysis and Reporting

### `pe stats` - Quick Statistics

```bash
# Basic stats from evaluation
pe eval results.yaml | pe stats

# Detailed statistics
pe eval results.yaml | pe stats --detailed --group-by provider

# Export statistics
pe eval results.yaml | pe stats --format csv --output stats.csv

# Real-time stats
pe eval live-test.yaml | pe stats --live --refresh 5s
```

### `pe diff` - Result Comparison

```bash
# Compare two result sets
pe diff baseline-results.json new-results.json

# Detailed comparison with metrics
pe diff old.json new.json \
  --metrics score,latency,cost \
  --threshold 0.05 \
  --output comparison-report.html

# A/B test comparison
pe diff variant-a.json variant-b.json \
  --statistical-test \
  --confidence 0.95
```

### `pe view` - Interactive Visualization

```bash
# View results in browser
pe eval results.yaml | pe view

# Custom visualization
pe view results.json \
  --chart scatter \
  --x-axis latency \
  --y-axis score \
  --color-by provider

# Dashboard mode
pe view --dashboard \
  --refresh 30s \
  --data live-results.jsonl
```

## 🔐 Security and Attestation

### `pe exp attest` - Cryptographic Attestation

```bash
# Inspect attestation prototype command surface
pe exp attest --help

# Re-check prototype interface
pe exp attest --help
```

### `pe security` - Security Testing

```bash
# OWASP LLM Top 10 testing
pe security scan prompt-file.prompt

# Custom security tests
pe security test \
  --tests injection,prompt-leaking,data-extraction \
  --severity high \
  --config security-config.yaml

# Red team testing
pe security redteam \
  --automated \
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

# Run distributed evaluation
pe eval large-test-suite.yaml \
  --distributed \
  --max-concurrent 20 \
  --timeout 5m

```

## 📦 Module Management

### `pe mod` - Module Management

```bash
# Initialize module
pe mod init github.com/myorg/prompts

# Add dependencies
pe mod download github.com/pe-community/base-prompts@v1.2.0

# List dependencies
pe mod list

# Update dependencies
pe mod tidy

# Create vendor copy
pe mod vendor

# Publish module
pe mod push --tag v1.0.0
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

# Install plugin
pe plugin install pe-promptfoo

# Use plugin
pe promptfoo convert config.yaml

# Update plugins
pe plugin update

# Remove plugin
pe plugin remove pe-promptfoo
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
pe diff baseline-results.json unit-results.json --output comparison.html

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

# 1. Optimize prompts for production
pe optimize production-prompts/ \
  --target "latency,cost" \
  --output optimized/

# 2. Validate optimized prompts
pe eval production-validation.yaml \
  --prompts optimized/ \
  --threshold 0.95

# 3. Create attestation
pe exp attest --help

# 4. Deploy with monitoring
pe exp distributed --help

# 5. Health check
pe eval health-check.yaml --distributed --alert-on-failure

echo "🚀 Production deployment complete!"
```

### Research and Development Workflow

```bash
#!/bin/bash
# research-workflow.sh

# 1. Generate prompt variations
pe evolve base-prompt.prompt \
  --generations 10 \
  --population 20 \
  --mutations creative,logical,concise

# 2. Evaluate all variations
pe eval research-eval.yaml \
  --prompts evolved-prompts/ \
  --providers openai,anthropic,ollama

# 3. Semantic optimization
pe semantic backprop \
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
# Cache frequently used results
pe run --cache my-prompt.prompt

# Parallel execution
pe eval --parallel 8 large-test-suite.yaml

# Resource limits
pe eval --memory-limit 4GB --timeout 5m heavy-eval.yaml
```

### Development Efficiency
```bash
# Watch for changes during development
pe watch development-tests.yaml --auto-reload

# Quick validation
pe vet --fast configs/

# Interactive debugging
pe run --debug --interactive problematic-prompt.prompt
```

### Production Best Practices
```bash
# Always use attestation in production
pe exp attest --help

# Monitor costs
pe benchmark --cost-tracking --budget 100.00 production-eval.yaml

# Graceful degradation
pe eval --fallback-provider anthropic --primary openai critical-eval.yaml
```

This comprehensive guide covers all PE commands with practical, real-world examples. Each command includes basic usage, advanced options, and integration patterns to help you become proficient with PE quickly!
