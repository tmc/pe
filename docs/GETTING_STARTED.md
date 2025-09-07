# Getting Started with PE

This guide will walk you through your first steps with PE, from installation to running your first evaluation.

## Prerequisites

- Go 1.21+ (for building from source)
- API keys for LLM providers (OpenAI, Anthropic, etc.)
- Basic familiarity with command line tools

## Installation

### Option 1: Install from Source (Recommended)

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Option 2: Build from Repository

```bash
git clone https://github.com/tmc/pe.git
cd pe
make build
# Or manually:
# go build -o pe ./cmd/pe
sudo mv pe /usr/local/bin/
```

### Verify Installation

```bash
pe --help
```

You should see the PE command help output.

## Setting Up API Keys

PE supports multiple LLM providers. Set up environment variables for the providers you want to use:

```bash
# OpenAI
export OPENAI_API_KEY="your-openai-api-key"

# Anthropic
export ANTHROPIC_API_KEY="your-anthropic-api-key"

# Google AI
export GOOGLE_AI_API_KEY="your-google-ai-key"
```

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.) to persist them.

## Your First Evaluation

### Step 1: Create a Configuration

Let's start with PE's initialization command:

```bash
pe init my-first-test.yaml
```

This creates a basic configuration file:

```yaml
description: "Capital cities evaluation"

prompts:
  - "What is the capital of {{.country}}?"
  - "Tell me about the capital city of {{.country}}."

providers:
  - name: "openai"
    config:
      model: "gpt-4o-mini"
  - name: "anthropic"
    config:
      model: "claude-3-haiku-20240307"

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

### Step 2: Run the Evaluation

```bash
pe eval my-first-test.yaml
```

You'll see output like:

```
Running 4 evaluations with up to 4 threads...

Evaluation: eval-2025-09-05T16:45:00
Timestamp: 2025-09-05T16:45:00Z

Success: 4, Failures: 0, Total: 4
Token Usage: 120 (Prompt: 60, Completion: 60)

ID      Prompt                          Provider    Success  Score
--      ------                          --------    -------  -----
abc123  What is the capital of France?  openai      ✓        1.00
def456  What is the capital of France?  anthropic   ✓        1.00
ghi789  Tell me about the capital...    openai      ✓        1.00
jkl012  Tell me about the capital...    anthropic   ✓        1.00
```

### Step 3: View Detailed Results

```bash
pe eval my-first-test.yaml --save-db
pe view
```

This opens an interactive browser interface where you can explore the results in detail.

## Understanding the Configuration

Let's break down the configuration file:

### Prompts
```yaml
prompts:
  - "What is the capital of {{.country}}?"  # Template with variable
  - "Tell me about {{.country}}'s capital." # Alternative prompt
```

Prompts use Go template syntax with variables in `{{.variable}}` format.

### Providers
```yaml
providers:
  - name: "openai"
    config:
      model: "gpt-4o-mini"       # OpenAI model
  - name: "anthropic"
    config:
      model: "claude-3-haiku-20240307"  # Anthropic model
```

Specify providers with their configuration including model selection.

### Tests
```yaml
tests:
  - vars:                             # Variables for template substitution
      country: "France"
    assert:                           # Assertions to verify output
      - type: "contains"              # Check if output contains text
        value: "Paris"
  - vars:
      country: "Japan"
    assert:
      - type: "contains"
        value: "Tokyo"
      - type: "length"                # Check output length
        min: 10
        max: 200
```

Tests define scenarios with variables and success criteria.

## Common Assertion Types

PE supports various assertion types:

```yaml
assert:
  # Text content checks
  - type: "contains"
    value: "expected text"
  - type: "not-contains"
    value: "unwanted text"
  - type: "regex"
    pattern: "Tokyo.*Japan"
    
  # Length checks
  - type: "length"
    min: 50
    max: 200
    
  # Performance checks
  - type: "latency"
    max: "5s"
  - type: "cost"
    max: 0.02
    
  # Quality checks (advanced)
  - type: "similarity"
    reference: "Expected output"
    threshold: 0.8
```

## Working with Multiple Tests

Create comprehensive test suites:

```yaml
prompts:
  - "Write a haiku about {{.topic}}"

providers:
  - name: "openai"
    config:
      model: "gpt-4o-mini"
  - name: "anthropic"
    config:
      model: "claude-3-haiku-20240307"

tests:
  # Test different topics
  - description: "Nature haiku"
    vars:
      topic: "cherry blossoms"
    assert:
      - type: "contains"
        value: ["cherry", "blossom"]
      - type: "regex"
        pattern: "\\n.*\\n.*\\n"  # 3 lines
        
  - description: "Technology haiku"
    vars:
      topic: "artificial intelligence"
    assert:
      - type: "length"
        min: 30
        max: 100
      - type: "not-contains"
        value: ["AI", "computer"]  # Avoid technical terms
```

## Saving and Sharing Results

### Save to File
```bash
# JSON format
pe eval config.yaml -o results.json

# YAML format  
pe eval config.yaml -o results.yaml

# CSV format for spreadsheets
pe eval config.yaml -o results.csv
```

### Save to Database
```bash
pe eval config.yaml --save-db
```

This saves results to `~/.promptfoo/evals/` for later viewing with `pe view`.

### Share Results
```bash
pe eval config.yaml --save-db --share
```

Creates a shareable URL for your evaluation results.

## Next Steps

Now that you've run your first evaluation, explore these features:

1. **Running Prompts**: `pe run prompt.txt --provider openai`
2. **Optimization**: `pe optimize prompt.txt --method pe2`
3. **Module Management**: `pe mod init myproject`
4. **Pipeline Commands**: `echo "Hello" | pe run summarize.prompt --provider cgpt`

## Common Patterns

### Development Workflow
```bash
# Create and edit config
pe init project-prompts.yaml
$EDITOR project-prompts.yaml

# Test a single prompt
pe run prompt.txt --provider openai

# Run full evaluation
pe eval project-prompts.yaml

# Optimize prompts
pe optimize prompt.txt --method pe2 --iterations 3
```

### CI/CD Integration
```bash
# Validate config
pe vet config.yaml

# Run tests
pe eval config.yaml --save-db

# Check for regressions
pe diff baseline.json current.json --threshold 0.05
```

### Cost Monitoring
```bash
# Track costs
pe eval config.yaml | pe analyze --metric cost --group-by provider

# Filter expensive runs
pe eval config.yaml | pe filter --max-cost 0.10 | pe stats
```

## Troubleshooting

### Common Issues

**API Key Not Found**
```bash
Error: openai provider requires OPENAI_API_KEY environment variable
```
Solution: Set the required environment variable:
```bash
export OPENAI_API_KEY="your-api-key"
```

**Config Validation Failed**
```bash
Error: missing required field 'prompts'
```
Solution: Ensure your config has all required fields:
```bash
pe vet config.yaml  # Check for issues
```

**Provider Not Supported**
```bash
Error: provider "unknown-provider" not found
```
Solution: Use a supported provider:
- `openai` - OpenAI API
- `anthropic` - Anthropic API
- `cgpt` - Multi-provider CLI tool

### Debug Mode

Enable verbose output for troubleshooting:

```bash
pe eval config.yaml --verbose
```

### Getting Help

- **Built-in help**: `pe help [command]`
- **Documentation**: [docs/README.md](README.md)
- **Examples**: [example/](../example/)
- **Issues**: [GitHub Issues](https://github.com/tmc/pe/issues)

## What's Next?

- Explore [Advanced Examples](README.md#advanced-examples)
- Learn about [Pipeline Commands](README.md#pipeline-processing-unix-style)
- Set up [CI/CD Integration](README.md#integration-guide)
- Try [Custom Assertions](README.md#configuration-guide)

Happy prompt engineering! 🚀