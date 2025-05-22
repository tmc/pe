# PE CLI: Comprehensive Command Reference

## Overview

PE provides a powerful command-line interface for systematic prompt engineering, implementing the latest 2024-2025 research in automated optimization. This guide covers all commands, options, and advanced usage patterns.

## Table of Contents

- [Installation & Setup](#installation--setup)
- [Core Commands](#core-commands)
- [Optimization Commands](#optimization-commands)
- [Evaluation Commands](#evaluation-commands)
- [Pipeline Commands](#pipeline-commands)
- [Development Commands](#development-commands)
- [Advanced Workflows](#advanced-workflows)
- [Configuration Reference](#configuration-reference)

## Installation & Setup

### Installation

```bash
# Install latest version
go install github.com/tmc/pe/cmd/pe@latest

# Verify installation
pe version

# Check available commands
pe --help
```

### Environment Setup

```bash
# Set API keys
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
export GOOGLE_AI_API_KEY="your-google-key"

# Optional: Set default configuration
export PE_CONFIG_PATH="~/.pe/config.yaml"
export PE_PROVIDER="openai:gpt-4"
```

### Quick Start

```bash
# Create your first evaluation
pe init my-first-eval.yaml

# Run evaluation
pe eval my-first-eval.yaml

# View results
pe view
```

## Core Commands

### `pe init`

Initialize a new configuration file with templates.

```bash
# Basic initialization
pe init config.yaml

# Initialize with specific template
pe init config.yaml --template evaluation
pe init config.yaml --template optimization
pe init config.yaml --template benchmark

# Initialize with provider-specific settings
pe init config.yaml --provider openai:gpt-4
pe init config.yaml --provider anthropic:claude-3-sonnet

# Interactive setup
pe init config.yaml --interactive
```

**Output:** Creates a YAML configuration file with appropriate templates and examples.

### `pe version`

Display version information and build details.

```bash
# Basic version info
pe version

# Detailed build information
pe version --verbose

# Check for updates
pe version --check-updates
```

### `pe help`

Comprehensive help system.

```bash
# General help
pe help

# Command-specific help
pe help optimize
pe help eval

# Show examples
pe help --examples

# Show advanced usage
pe help --advanced
```

## Optimization Commands

### `pe optimize`

Optimize prompts using cutting-edge metaprompting techniques.

#### Basic Usage

```bash
# Standard optimization
pe optimize --prompt "Analyze sentiment" --iterations 3

# PE2 optimization (meta-prompt engineering)
pe optimize --prompt "Classify text" --method pe2 --iterations 5

# APEX optimization (long prompts)
pe optimize --prompt-file system-prompt.txt --method apex --iterations 8

# TextGrad optimization (semantic gradients)
pe optimize --prompt "Solve problems" --method textgrad --iterations 6

# Hybrid optimization (multi-method)
pe optimize --prompt "Complex task" --method hybrid --iterations 10
```

#### Advanced Options

```bash
# Provider and model selection
pe optimize --prompt "Task" --provider anthropic --model claude-3-opus
pe optimize --prompt "Task" --provider openai:gpt-4-turbo

# Temperature and token control
pe optimize --prompt "Task" --temperature 0.7 --max-tokens 1500

# Output and saving
pe optimize --prompt "Task" --output optimized.json
pe optimize --prompt "Task" --save-db --id "experiment-1"

# Method-specific configuration
pe optimize --prompt "Task" --method pe2 --reasoning-template analytical
pe optimize --prompt "Task" --method apex --beam-width 7 --max-length 800
pe optimize --prompt "Task" --method textgrad --attention-layers 8,12,16
```

#### Real-World Examples

```bash
# Customer service optimization
pe optimize \
  --prompt "You are a customer service agent" \
  --method pe2 \
  --iterations 5 \
  --provider anthropic:claude-3-sonnet \
  --output customer-service-v2.json

# Code review prompt optimization
pe optimize \
  --prompt-file code-review-prompt.txt \
  --method apex \
  --beam-width 5 \
  --max-length 1000 \
  --iterations 8 \
  --save-db

# Research analysis optimization
pe optimize \
  --prompt "Analyze research papers" \
  --method textgrad \
  --attention-flow \
  --detect-drift \
  --iterations 6 \
  --output research-analyzer-v3.json
```

### `pe compose`

Component-based prompt engineering (unique to PE).

```bash
# Initialize component library
pe compose --library-init --path ./components

# Add components
pe compose --add-component context-legal.txt --category context
pe compose --add-component examples-sentiment.txt --category examples

# Compose optimized prompts
pe compose context/legal.txt instructions/analyze.txt examples/sentiment.txt

# Style-specific composition
pe compose components/ --style chain-of-thought --optimize
pe compose components/ --style few-shot --target gpt-4

# Advanced composition
pe compose \
  context/technical.txt \
  instructions/code-review.txt \
  examples/python.txt \
  --style analytical \
  --optimize coherence \
  --max-length 800 \
  --output composed-prompt.txt
```

### `pe evolve`

Evolutionary prompt optimization (unique to PE).

```bash
# Basic evolutionary optimization
pe evolve baseline.txt --generations 20 --population 15

# Multi-objective optimization
pe evolve prompt.txt --objectives accuracy,latency,cost --generations 25

# Pareto frontier analysis
pe evolve prompt.txt --multi-objective --extract-pareto-front

# Advanced evolutionary parameters
pe evolve prompt.txt \
  --algorithm nsga2 \
  --generations 30 \
  --population 20 \
  --crossover-rate 0.8 \
  --mutation-rate 0.2 \
  --elite-ratio 0.1

# Custom mutation operators
pe evolve prompt.txt \
  --operators word_substitution,phrase_insertion,structure_modification \
  --adaptive-rates \
  --generations 25
```

### `pe fusion`

Multi-model consensus optimization (unique to PE).

```bash
# Basic consensus optimization
pe fusion prompt.txt --models gpt-4,claude-3,gemini-pro

# Weighted consensus
pe fusion prompt.txt \
  --models gpt-4,claude-3,gemini-pro \
  --weights 0.4,0.35,0.25 \
  --consensus weighted

# Reflection-based consensus
pe fusion prompt.txt \
  --models gpt-4,claude-3 \
  --consensus reflection \
  --reflection-depth 3

# Adaptive weighting
pe fusion prompt.txt \
  --models gpt-4,claude-3,gemini-pro \
  --adaptive-weights \
  --learning-rate 0.1 \
  --iterations 5
```

## Evaluation Commands

### `pe eval`

Core evaluation engine with multi-provider support.

#### Basic Usage

```bash
# Simple evaluation
pe eval config.yaml

# Save results to file
pe eval config.yaml -o results.json
pe eval config.yaml -o results.csv --format csv

# Save to database for web viewing
pe eval config.yaml --save-db

# Dry run (show what would be executed)
pe eval config.yaml --dry-run
```

#### Performance Control

```bash
# Concurrency control
pe eval config.yaml --max-concurrency 8

# Timeout settings
pe eval config.yaml --timeout 60s --per-request-timeout 30s

# Memory optimization
pe eval config.yaml --memory-limit 2GB --streaming

# Rate limiting
pe eval config.yaml --rate-limit 10/minute
```

#### Advanced Features

```bash
# Statistical evaluation
pe eval config.yaml --statistical-significance --confidence 0.95

# Cross-validation
pe eval config.yaml --cross-validate --folds 5

# A/B testing
pe eval config.yaml --ab-test baseline.yaml --sample-size 1000

# Real-time monitoring
pe eval config.yaml --monitor --alerts slack://channel
```

### `pe test`

Advanced testing framework.

```bash
# Property-based testing
pe test property config.yaml --properties robustness,consistency

# Regression testing
pe test regression --baseline baseline.json --current current.json

# Security testing
pe test security config.yaml --categories injection,jailbreak,bias

# Performance testing
pe test performance config.yaml --load-test --concurrent-users 50

# Comprehensive testing suite
pe test comprehensive config.yaml \
  --property-based \
  --regression \
  --security \
  --performance \
  --statistical-validation
```

### `pe benchmark`

Performance analysis and comparison.

```bash
# Basic benchmark
pe benchmark config.yaml

# Multi-iteration benchmark
pe benchmark config.yaml --iterations 10 --warmup 2

# Concurrency testing
pe benchmark config.yaml --concurrency 1,2,4,8 --iterations 5

# Cost analysis
pe benchmark config.yaml --cost-analysis --budget 10.00

# Statistical benchmarking
pe benchmark config.yaml \
  --iterations 20 \
  --confidence-interval 0.95 \
  --outlier-detection \
  --format detailed
```

## Pipeline Commands

PE's unique Unix-style pipeline processing enables powerful data workflows.

### `pe ask`

Single-question pipeline interface.

```bash
# Basic usage
echo "What is AI?" | pe ask --provider openai:gpt-4

# With options
echo "Analyze data" | pe ask \
  --provider anthropic:claude-3-sonnet \
  --temperature 0.3 \
  --max-tokens 500

# Batch processing
cat questions.txt | pe ask --provider openai:gpt-4 --batch

# Pipeline integration
echo "Complex question" | pe ask --provider gpt-4 | pe analyze --metric quality
```

### `pe stream`

Stream processing for real-time analysis.

```bash
# Extract specific fields
pe eval config.yaml | pe stream --select response,latency,cost

# Filter streaming results
pe eval config.yaml | pe stream --filter 'latency < 2000'

# Transform streaming data
pe eval config.yaml | pe stream --transform 'cost * 1000' --select cost_mil

# Real-time aggregation
pe eval config.yaml | pe stream --aggregate 'avg(latency), sum(cost)'
```

### `pe filter`

Advanced result filtering.

```bash
# Success/failure filtering
pe eval config.yaml | pe filter --success
pe eval config.yaml | pe filter --failed

# Metric-based filtering
pe eval config.yaml | pe filter --latency "<2s" --cost "<0.01"
pe eval config.yaml | pe filter --score ">8.0" --confidence ">0.9"

# Provider-based filtering
pe eval config.yaml | pe filter --provider openai
pe eval config.yaml | pe filter --model gpt-4

# Complex conditions
pe eval config.yaml | pe filter \
  --where 'latency < 1000 AND cost < 0.005 AND score > 8.5'
```

### `pe analyze`

Statistical analysis and insights.

```bash
# Basic analysis
pe eval config.yaml | pe analyze

# Metric-specific analysis
pe eval config.yaml | pe analyze --metric latency
pe eval config.yaml | pe analyze --metric cost --group-by provider

# Statistical analysis
pe eval config.yaml | pe analyze \
  --statistical \
  --confidence 0.95 \
  --percentiles 50,90,95,99

# Trend analysis
pe eval config.yaml | pe analyze --trend --window 24h

# Comparative analysis
pe eval config.yaml | pe analyze --compare baseline.json
```

### `pe stats`

Quick statistical summaries.

```bash
# Basic stats
pe eval config.yaml | pe stats

# Detailed statistics
pe eval config.yaml | pe stats --detailed

# Grouped statistics
pe eval config.yaml | pe stats --group-by provider,model

# Format options
pe eval config.yaml | pe stats --format table
pe eval config.yaml | pe stats --format json
pe eval config.yaml | pe stats --format csv
```

## Development Commands

### `pe interactive`

Interactive REPL for prompt development.

```bash
# Start REPL
pe interactive

# With specific provider
pe interactive --provider anthropic:claude-3-sonnet

# Load configuration
pe interactive --config development.yaml

# Debug mode
pe interactive --debug --verbose

# Tutorial mode
pe interactive --tutorial
```

**REPL Commands:**
```
> .help                 # Show REPL help
> .provider openai:gpt-4 # Switch provider
> .temperature 0.7      # Set temperature
> .save prompt.txt      # Save current prompt
> .load prompt.txt      # Load prompt
> .optimize pe2         # Optimize current prompt
> .benchmark            # Benchmark current prompt
> .exit                 # Exit REPL
```

### `pe watch`

Auto-reload development mode.

```bash
# Watch current directory
pe watch config.yaml

# Watch specific patterns
pe watch config.yaml --include "*.yaml,prompts/**/*"

# Auto-save results
pe watch config.yaml --auto-save --output results/

# Live dashboard
pe watch config.yaml --dashboard --port 8080

# Debounce settings
pe watch config.yaml --debounce 500ms
```

### `pe fmt`

Format and validate configuration files.

```bash
# Format file
pe fmt config.yaml

# Convert formats
pe fmt config.yaml --output json > config.json
pe fmt config.json --output yaml > config.yaml

# Write back to file
pe fmt config.yaml --write

# Check formatting only
pe fmt config.yaml --check

# Validate configuration
pe fmt config.yaml --validate --strict
```

### `pe vet`

Configuration validation and linting.

```bash
# Basic validation
pe vet config.yaml

# Strict validation
pe vet config.yaml --strict

# Multiple files
pe vet *.yaml

# JSON Schema validation
pe vet config.yaml --schema

# Performance recommendations
pe vet config.yaml --performance-hints

# Security analysis
pe vet config.yaml --security-check
```

## Advanced Workflows

### Complete Optimization Pipeline

```bash
# Research-grade optimization workflow
pe compose \
  context/research.txt \
  instructions/analyze.txt \
  examples/papers.txt \
  --style analytical | \
pe optimize --method pe2 --iterations 5 | \
pe evolve --generations 15 --multi-objective | \
pe fusion --models gpt-4,claude-3 --consensus weighted | \
pe test comprehensive --statistical-validation | \
pe benchmark --iterations 10 --confidence 0.95
```

### Production Deployment Pipeline

```bash
# Production readiness workflow
pe eval config.yaml --save-db | \
pe test security --comprehensive | \
pe test performance --load-test | \
pe analyze --statistical --export-report | \
pe monitor --alerts production-team
```

### Research and Development Workflow

```bash
# R&D experimentation pipeline
pe init experiment.yaml --template research | \
pe optimize --method hybrid --trace-optimization | \
pe eval experiment.yaml --cross-validate --folds 10 | \
pe analyze --statistical --export-data research.csv | \
pe benchmark --academic-metrics --export-results
```

### Multi-Provider Comparison

```bash
# Comprehensive provider comparison
pe eval config.yaml \
  --providers openai:gpt-4,anthropic:claude-3,google:gemini-pro | \
pe analyze --comparative --group-by provider | \
pe stats --detailed --export-table comparison.csv | \
pe benchmark --cost-analysis --performance-analysis
```

## Configuration Reference

### Environment Variables

```bash
# API Keys
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_AI_API_KEY="..."

# Default Settings
export PE_PROVIDER="openai:gpt-4"
export PE_TEMPERATURE="0.3"
export PE_MAX_TOKENS="1000"
export PE_TIMEOUT="60s"

# Paths
export PE_CONFIG_PATH="~/.pe/config.yaml"
export PE_CACHE_PATH="~/.pe/cache"
export PE_RESULTS_PATH="~/.pe/results"

# Performance
export PE_MAX_CONCURRENCY="8"
export PE_MEMORY_LIMIT="4GB"
export PE_CACHE_SIZE="1GB"

# Monitoring
export PE_TELEMETRY_ENDPOINT="https://telemetry.example.com"
export PE_ALERT_WEBHOOK="https://hooks.slack.com/..."
```

### Global Configuration File

**~/.pe/config.yaml**
```yaml
# Default settings
defaults:
  provider: "openai:gpt-4"
  temperature: 0.3
  max_tokens: 1000
  timeout: "60s"

# Provider configurations
providers:
  openai:
    api_key: "${OPENAI_API_KEY}"
    base_url: "https://api.openai.com/v1"
    timeout: "30s"
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    base_url: "https://api.anthropic.com"
    timeout: "30s"

# Optimization settings
optimization:
  pe2:
    reasoning_template: "chain_of_thought"
    context_specification: "detailed"
    description_depth: "comprehensive"
  apex:
    beam_width: 5
    max_length: 2000
    mutation_probability: 0.3
  textgrad:
    learning_rate: 0.1
    attention_layers: [8, 12, 16]
    drift_threshold: 0.8

# Performance settings
performance:
  max_concurrency: 8
  memory_limit: "4GB"
  cache_size: "1GB"
  rate_limit: "100/minute"

# Monitoring
monitoring:
  enabled: true
  telemetry_endpoint: "${PE_TELEMETRY_ENDPOINT}"
  alert_webhook: "${PE_ALERT_WEBHOOK}"
  log_level: "info"
```

### Command Aliases

**~/.pe/aliases.yaml**
```yaml
aliases:
  opt: "optimize --method pe2 --iterations 5"
  optlong: "optimize --method apex --beam-width 5"
  eval-prod: "eval --save-db --monitor --statistical-significance"
  test-all: "test comprehensive --property-based --security --performance"
  bench: "benchmark --iterations 10 --confidence 0.95"
  
# Usage: pe opt --prompt "task" == pe optimize --method pe2 --iterations 5 --prompt "task"
```

## Performance Tips

### Speed Optimization

```bash
# Use concurrency for large evaluations
pe eval config.yaml --max-concurrency 16

# Enable caching for repeated evaluations
pe eval config.yaml --cache --cache-ttl 1h

# Stream processing for memory efficiency
pe eval config.yaml --streaming --batch-size 10

# Parallel provider evaluation
pe eval config.yaml --parallel-providers
```

### Cost Optimization

```bash
# Cost tracking and budgets
pe eval config.yaml --cost-tracking --budget 5.00

# Efficient provider selection
pe eval config.yaml --cost-optimize --fallback-providers

# Batch processing for rate limit efficiency
pe eval config.yaml --batch-requests --optimal-batching
```

### Quality Optimization

```bash
# Statistical significance testing
pe eval config.yaml --statistical-significance --min-samples 100

# Cross-validation for robust results
pe eval config.yaml --cross-validate --folds 5 --stratified

# Comprehensive testing
pe test comprehensive config.yaml --all-metrics
```

---

**PE CLI provides the most comprehensive and advanced prompt engineering interface available, implementing cutting-edge research with production-grade reliability and performance.**