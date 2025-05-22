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
go build -o pe cmd/pe/main.go
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

### Step 2: Run the Evaluation

```bash
pe eval my-first-test.yaml
```

You'll see output like:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                Evaluation Results                            │
├─────────────────────────────────────────────────────────────────────────────┤
│ Prompt                                │ Provider               │ Pass │ Score │
├─────────────────────────────────────────────────────────────────────────────┤
│ What is the capital of {{country}}?   │ openai:gpt-4          │  ✓   │ 1.00  │
│ What is the capital of {{country}}?   │ anthropic:claude-3-haiku │  ✓   │ 1.00  │
│ Tell me about the capital city...     │ openai:gpt-4          │  ✓   │ 1.00  │
│ Tell me about the capital city...     │ anthropic:claude-3-haiku │  ✓   │ 1.00  │
└─────────────────────────────────────────────────────────────────────────────┘

Summary: 4/4 tests passed (100.0%)
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
  - "What is the capital of {{country}}?"  # Template with variable
  - "Tell me about {{country}}'s capital." # Alternative prompt
```

Prompts can use template variables in `{{variable}}` syntax.

### Providers
```yaml
providers:
  - "openai:gpt-4"                    # Provider:model format
  - "anthropic:claude-3-haiku"        # Different provider
  - "openai:gpt-3.5-turbo"           # Different model
```

Specify providers in `provider:model` format.

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
  - "Write a haiku about {{topic}}"

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-haiku"

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

1. **Interactive Mode**: `pe interactive --provider openai:gpt-4`
2. **Watch Mode**: `pe watch config.yaml` (auto-rerun on changes)
3. **Benchmarking**: `pe benchmark config.yaml --iterations 5`
4. **Pipeline Commands**: `echo "Hello" | pe ask --provider openai:gpt-4`

## Common Patterns

### Development Workflow
```bash
# Create and edit config
pe init project-prompts.yaml
$EDITOR project-prompts.yaml

# Test interactively
pe interactive --config project-prompts.yaml

# Run full evaluation
pe eval project-prompts.yaml --save-db

# View results
pe view
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
Error: unsupported provider: unknown-provider
```
Solution: Use a supported provider format:
- `openai:gpt-4`
- `anthropic:claude-3-sonnet`
- `googleai:gemini-pro`

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