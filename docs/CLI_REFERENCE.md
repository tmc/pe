# PE CLI Reference

Complete reference for all PE commands, options, and usage patterns.

## Global Options

These options are available for all commands:

```bash
pe [global-options] command [command-options]
```

Global options:
- `--help, -h`: Show help information
- `--version`: Show version information

## Commands Overview

| Command | Purpose | Example |
|---------|---------|---------|
| [`run`](#run) | Execute prompt immediately | `pe run "What is AI?"` |
| [`eval`](#eval) | Run prompt evaluations | `pe eval config.yaml` |
| [`view`](#view) | View results in browser | `pe view` |
| [`interactive`](#interactive) | Start REPL mode | `pe interactive` |
| [`watch`](#watch) | Auto-rerun on changes | `pe watch config.yaml` |
| [`ask`](#ask) | Single prompt query | `pe ask "What is AI?"` |
| [`benchmark`](#benchmark) | Performance testing | `pe benchmark config.yaml` |
| [`stream`](#stream) | Process results stream | `pe eval config.yaml \| pe stream` |
| [`filter`](#filter) | Filter results | `pe eval config.yaml \| pe filter --success` |
| [`analyze`](#analyze) | Statistical analysis | `pe eval config.yaml \| pe analyze` |
| [`stats`](#stats) | Quick statistics | `pe eval config.yaml \| pe stats` |
| [`diff`](#diff) | Compare results | `pe diff old.json new.json` |
| [`fmt`](#fmt) | Format configs | `pe fmt config.yaml` |
| [`vet`](#vet) | Validate configs | `pe vet config.yaml` |
| [`convert`](#convert) | Convert formats | `pe convert config.yaml config.json` |
| [`init`](#init) | Create config template | `pe init new-config.yaml` |
| [`plugin`](#plugin) | Manage plugins | `pe plugin list` |

---

## run

Execute a prompt immediately using the inference API.

### Synopsis

```bash
pe run [prompt or file] [flags]
```

### Description

The `run` command executes prompts directly without needing a configuration file. It's designed for quick testing and iteration, similar to `go run` for Go programs. Supports both direct prompt strings and prompt files.

### Arguments

- `prompt or file`: Either a prompt string or path to a file containing the prompt

### Flags

```bash
-p, --provider string        Inference provider to use (default "cgpt")
-m, --model string           Model to use (e.g., gpt-4, claude-3)
-t, --temperature float32    Temperature for randomness (0.0-1.0) (default 0.7)
    --max-tokens int         Maximum tokens in response
-s, --system string          System prompt
    --stream                 Stream the response
    --var stringToString     Template variables (can be repeated)
```

### Examples

```bash
# Simple prompt
pe run "What is 2+2?"

# From file
pe run prompt.txt

# With variables
pe run "Translate {{.Text}} to {{.Language}}" \
  --var Text="Hello world" \
  --var Language="French"

# With specific model and temperature
pe run "Write a haiku about coding" \
  --temperature 0.9 \
  --model gpt-4

# Streaming mode
pe run "Tell me a story about a robot" \
  --stream \
  --max-tokens 200

# With system prompt
pe run "Explain recursion" \
  --system "You are a computer science teacher. Use simple examples."

# From stdin
echo "What are the benefits of Go?" | pe run -
```

### Template Variables

The run command supports Go template syntax for variables:
- Use `{{.VarName}}` in your prompt
- Pass values with `--var VarName=value`
- Multiple variables can be specified

---

## eval

Run prompt evaluations against configured providers and tests.

### Synopsis

```bash
pe eval [config_file] [flags]
```

### Description

The `eval` command is the core of PE. It reads a configuration file, executes prompts against specified LLM providers, and evaluates the results against defined assertions.

### Arguments

- `config_file`: Path to YAML/JSON configuration file (optional, defaults to `promptfooconfig.yaml`)

### Flags

```bash
-c, --config string           Path to configuration file
-o, --output string          Write results to file (format inferred from extension)
-t, --timeout duration       Timeout for entire test run (default "30s")
    --dry-run               Show commands without executing them
    --save-db               Save results to promptfoo database
    --share                 Create shareable URL for results
-j, --max-concurrency int    Maximum concurrent API calls (default 4)
    --no-progress-bar       Disable progress bar display
```

### Examples

```bash
# Basic evaluation
pe eval config.yaml

# Save results in different formats
pe eval config.yaml -o results.json
pe eval config.yaml -o results.csv
pe eval config.yaml -o results.yaml

# Control execution
pe eval config.yaml --timeout 60s --max-concurrency 8

# Save to database for viewing
pe eval config.yaml --save-db

# Dry run to see what would execute
pe eval config.yaml --dry-run

# Create shareable results
pe eval config.yaml --save-db --share
```

### Output Formats

- **Table** (default): Human-readable table format
- **JSON** (`.json`): Machine-readable JSON
- **YAML** (`.yaml`, `.yml`): YAML format
- **CSV** (`.csv`): Spreadsheet-compatible CSV

---

## view

View evaluation results in an interactive browser interface.

### Synopsis

```bash
pe view [eval_id] [flags]
```

### Description

Opens a web browser interface to explore evaluation results. Can view results from the database or from a specific file.

### Arguments

- `eval_id`: Specific evaluation ID to view (optional)

### Flags

```bash
-f, --file string    View results from specific file
-p, --port int       Port for local server (default 8080)
    --no-open        Don't automatically open browser
```

### Examples

```bash
# View latest results
pe view

# View specific evaluation
pe view eval-123

# View from file
pe view -f results.json

# Custom port
pe view --port 9000

# Don't open browser automatically
pe view --no-open
```

---

## interactive

Start an interactive REPL (Read-Eval-Print Loop) for rapid prompt development.

### Synopsis

```bash
pe interactive [flags]
```

### Description

Launches an interactive session where you can test prompts in real-time, switch providers, adjust parameters, and save sessions.

### Flags

```bash
-p, --provider string        Default LLM provider (default "openai:gpt-4")
-c, --config string         Load configuration file
-t, --temperature float64   Default temperature (default 0.7)
```

### Examples

```bash
# Start with default provider
pe interactive

# Use specific provider
pe interactive --provider anthropic:claude-3-sonnet

# Load existing configuration
pe interactive --config my-config.yaml

# Set default temperature
pe interactive --temperature 0.9
```

### Interactive Commands

Once in interactive mode, use these commands:

```bash
:help, :h                    Show help
:quit, :q                    Exit
:provider <provider>         Switch provider
:temp <value>                Set temperature (0.0-2.0)
:tokens <value>              Set max tokens
:save <filename>             Save session
:load <filename>             Load configuration
:history                     Show command history
:clear                       Clear screen
:multiline                   Enter multiline mode
:benchmark <n>               Benchmark prompt n times
:status                      Show current settings
:models                      List available models
```

---

## watch

Monitor configuration files and automatically re-run evaluations when changes are detected.

### Synopsis

```bash
pe watch [config_file] [flags]
```

### Description

Watches configuration and prompt files for changes and automatically re-runs evaluations. Great for development workflows.

### Arguments

- `config_file`: Configuration file to watch (optional, searches for defaults)

### Flags

```bash
-c, --config string         Configuration file path
-o, --output string         Write results to file on each run
-i, --include string        Comma-separated glob patterns to watch (default "*.yaml,*.yml,*.json,prompts/**/*")
```

### Examples

```bash
# Watch current directory
pe watch config.yaml

# Watch with output file
pe watch config.yaml -o results.json

# Watch specific patterns
pe watch config.yaml --include "*.yaml,prompts/**/*,tests/**/*"

# Auto-detect config file
pe watch
```

### Default Config Search

If no config file is specified, PE searches for:
1. `promptfooconfig.yaml`
2. `pe-config.yaml`
3. `config.yaml`

---

## ask

Ask a single question to an LLM provider. Pipeline-friendly for Unix-style composition.

### Synopsis

```bash
pe ask [prompt] [flags]
echo "prompt" | pe ask [flags]
```

### Description

Send a single prompt to an LLM provider and get a response. Designed for pipeline use and quick queries.

### Arguments

- `prompt`: The prompt to send (optional if reading from stdin)

### Flags

```bash
-p, --provider string        LLM provider to use (default "openai:gpt-4")
-t, --temperature float64    Temperature for generation (default -1, uses provider default)
-m, --max-tokens int         Maximum tokens to generate (default 0, uses provider default)
-f, --format string          Output format: json, yaml, text (default "json")
```

### Examples

```bash
# Direct prompt
pe ask "What is the capital of France?" --provider openai:gpt-4

# From stdin
echo "Explain quantum computing" | pe ask --provider anthropic:claude-3-haiku

# With parameters
pe ask "Write a haiku" --temperature 0.9 --max-tokens 100

# Different formats
pe ask "Hello" --format text
pe ask "Hello" --format json
```

---

## benchmark

Run performance benchmarks comparing prompts, providers, and configurations.

### Synopsis

```bash
pe benchmark [config_file] [flags]
```

### Description

Runs multiple iterations of evaluations to measure performance characteristics like latency, cost, and quality metrics.

### Arguments

- `config_file`: Configuration file for benchmarking

### Flags

```bash
-i, --iterations int         Number of benchmark iterations (default 5)
-c, --concurrency int       Concurrent requests (default 2)
-f, --format string         Output format: text, json, csv (default "text")
-o, --output string         Write results to file
    --warmup int            Warmup iterations (default 1)
    --timeout duration      Timeout per iteration (default "60s")
```

### Examples

```bash
# Basic benchmark
pe benchmark config.yaml

# Multiple iterations
pe benchmark config.yaml --iterations 10 --concurrency 4

# Save results
pe benchmark config.yaml --format json -o benchmark.json

# With warmup
pe benchmark config.yaml --warmup 3 --iterations 10
```

---

## stream

Process evaluation results as a stream for pipeline processing.

### Synopsis

```bash
pe eval config.yaml | pe stream [flags]
```

### Description

Processes evaluation results line by line, supporting field selection and format conversion. Designed for Unix pipeline composition.

### Flags

```bash
-s, --select string         Comma-separated fields to output
-f, --format string         Output format: json, ndjson, csv, tsv (default "json")
    --filter string         Filter expression
```

### Examples

```bash
# Select specific fields
pe eval config.yaml | pe stream --select response,latency,cost

# Convert to NDJSON
pe eval config.yaml | pe stream --format ndjson

# Filter and select
pe eval config.yaml | pe stream --select cost,provider --filter "success=true"
```

---

## filter

Filter evaluation results based on various conditions.

### Synopsis

```bash
pe eval config.yaml | pe filter [flags]
```

### Description

Filters evaluation results based on success/failure, provider, performance metrics, and other criteria.

### Flags

```bash
    --success               Filter for successful results only
    --failure               Filter for failed results only
    --provider string       Filter by provider
    --min-score float       Minimum score threshold (default -1)
    --max-latency int       Maximum latency in ms (default -1)
```

### Examples

```bash
# Show only successful tests
pe eval config.yaml | pe filter --success

# Filter by provider
pe eval config.yaml | pe filter --provider "openai:gpt-4"

# Performance filters
pe eval config.yaml | pe filter --max-latency 1000 --min-score 0.8

# Show only failures
pe eval config.yaml | pe filter --failure
```

---

## analyze

Compute statistics and insights from evaluation results.

### Synopsis

```bash
pe eval config.yaml | pe analyze [flags]
```

### Description

Performs statistical analysis on evaluation results, supporting grouping, percentile calculations, and various metrics.

### Flags

```bash
-m, --metric string         Metric to analyze: score, latency, tokens, cost (default "score")
-g, --group-by string       Group results by field: provider, prompt, test
-p, --percentiles string    Comma-separated percentiles (default "50,90,95,99")
```

### Examples

```bash
# Analyze scores
pe eval config.yaml | pe analyze --metric score

# Group by provider
pe eval config.yaml | pe analyze --metric latency --group-by provider

# Custom percentiles
pe eval config.yaml | pe analyze --percentiles "25,50,75,95"

# Cost analysis
pe eval config.yaml | pe analyze --metric cost --group-by provider
```

---

## stats

Show quick statistics from evaluation results.

### Synopsis

```bash
pe eval config.yaml | pe stats [flags]
```

### Description

Provides quick statistical summaries including success rate, average scores, latency metrics, and cost summaries.

### Flags

```bash
-f, --format string    Output format: table, json, yaml (default "table")
```

### Examples

```bash
# Quick stats
pe eval config.yaml | pe stats

# JSON format
pe eval config.yaml | pe stats --format json

# YAML format
pe eval config.yaml | pe stats --format yaml
```

---

## diff

Compare two evaluation results to detect changes and regressions.

### Synopsis

```bash
pe diff [baseline_file] [current_file] [flags]
pe eval config.yaml | pe diff baseline.json [flags]
```

### Description

Compares evaluation results to identify differences in scores, latency, costs, and other metrics. Useful for regression testing.

### Arguments

- `baseline_file`: Baseline results file
- `current_file`: Current results file (optional if reading from stdin)

### Flags

```bash
-m, --metric string         Primary metric for comparison (default "score")
-t, --threshold float       Regression threshold, e.g., 0.05 = 5% (default 0.05)
```

### Examples

```bash
# Compare files
pe diff baseline.json current.json

# From pipeline
pe eval config.yaml | pe diff baseline.json

# Custom threshold
pe diff old.json new.json --threshold 0.10

# Different metric
pe diff old.json new.json --metric latency
```

---

## fmt

Format and optionally convert configuration files.

### Synopsis

```bash
pe fmt [file...] [flags]
```

### Description

Formats configuration files for consistency and optionally converts between YAML and JSON formats.

### Arguments

- `file...`: Configuration files to format

### Flags

```bash
-w, --write               Write result to source file instead of stdout
-o, --output string       Output format: yaml, json (default is input format)
```

### Examples

```bash
# Format to stdout
pe fmt config.yaml

# Format in-place
pe fmt config.yaml --write

# Convert YAML to JSON
pe fmt config.yaml --output json

# Format multiple files
pe fmt *.yaml --write
```

---

## vet

Validate configuration files for correctness.

### Synopsis

```bash
pe vet [file...] [flags]
cat config.yaml | pe vet
```

### Description

Validates configuration files against the PE schema, checking for required fields, syntax errors, and common issues.

### Arguments

- `file...`: Configuration files to validate

### Examples

```bash
# Validate single file
pe vet config.yaml

# Validate multiple files
pe vet *.yaml

# Validate from stdin
cat config.yaml | pe vet

# Validate all configs in directory
find . -name "*.yaml" -exec pe vet {} \;
```

---

## convert

Convert configuration files between YAML and JSON formats.

### Synopsis

```bash
pe convert [input_file] [output_file] [flags]
```

### Description

Converts configuration files between different formats while preserving structure and content.

### Arguments

- `input_file`: Source configuration file
- `output_file`: Target configuration file

### Flags

```bash
-o, --output string    Output format: yaml, json (default determined by file extension)
```

### Examples

```bash
# YAML to JSON
pe convert config.yaml config.json

# JSON to YAML
pe convert config.json config.yaml

# Explicit format
pe convert input.txt output.txt --output yaml
```

---

## init

Create a new configuration file with a basic template.

### Synopsis

```bash
pe init [output_file] [flags]
```

### Description

Creates a new configuration file with a basic template to get started quickly.

### Arguments

- `output_file`: Output file path (default: "pe-config.yaml")

### Flags

```bash
-f, --format string    Output format: yaml, json (default "yaml")
    --force           Overwrite existing file
```

### Examples

```bash
# Create default config
pe init

# Custom filename
pe init my-config.yaml

# JSON format
pe init config.json --format json

# Overwrite existing
pe init existing-config.yaml --force
```

---

## Environment Variables

PE respects these environment variables:

### Provider API Keys

```bash
OPENAI_API_KEY          # OpenAI API key
ANTHROPIC_API_KEY       # Anthropic API key  
GOOGLE_AI_API_KEY       # Google AI API key
```

### Configuration

```bash
PE_CONFIG_FILE          # Default configuration file
PE_PROVIDER             # Default provider for ask/interactive commands
PE_TEMPERATURE          # Default temperature
PE_MAX_TOKENS           # Default max tokens
PE_TIMEOUT              # Default timeout
PE_CONCURRENCY          # Default concurrency
```

### Output

```bash
PE_OUTPUT_FORMAT        # Default output format
PE_NO_COLOR             # Disable colored output
PE_QUIET                # Reduce output verbosity
```

---

## Configuration File Format

PE supports YAML and JSON configuration files with the following structure:

### Basic Structure

```yaml
# Optional description
description: "My prompt evaluation"

# Prompts to test (required)
prompts:
  - "Simple prompt"
  - "Prompt with {{variable}}"
  - id: "named-prompt"
    content: "Complex prompt content"

# Providers to test against (required)
providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-sonnet"
  - id: "custom-provider"
    type: "openai"
    model: "gpt-3.5-turbo"
    config:
      temperature: 0.7

# Test cases (required)
tests:
  - vars:
      variable: "value"
    assert:
      - type: "contains"
        value: "expected"

# Optional advanced features
redteaming:
  enabled: true
  categories: ["harmful", "biased"]

metrics:
  - name: "custom-metric"
    type: "python"
    script: "./custom-scorer.py"

output:
  format: "json"
  include_raw: true
```

## Exit Codes

PE uses these exit codes:

- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Provider error
- `4`: Evaluation failure
- `5`: File not found
- `6`: Permission denied

---

## plugin

Manage PE plugins and extensions.

### Synopsis

```bash
pe plugin [subcommand] [flags]
```

### Description

The `plugin` command manages PE plugins. Plugins are discovered as `pe-*` executables in your PATH and can extend PE's functionality.

### Subcommands

#### plugin list

List all discovered plugins.

```bash
pe plugin list
```

Example output:
```
Installed plugins:
  promptfoo - Promptfoo compatibility layer for PE
    Version: 0.1.0
```

#### plugin run

Run a plugin command explicitly.

```bash
pe plugin run <plugin-name> [args...]
```

Example:
```bash
pe plugin run promptfoo import config.yaml
```

### Plugin Discovery

Plugins are automatically discovered as executables with the pattern `pe-*` in your PATH:
- `pe-promptfoo` → available as `pe promptfoo`
- `pe-custom` → available as `pe custom`

### Direct Plugin Invocation

Once discovered, plugins can be invoked directly:

```bash
# These are equivalent:
pe plugin run promptfoo import config.yaml
pe promptfoo import config.yaml
```

### Creating Plugins

To create a PE plugin:

1. Create an executable named `pe-yourplugin`
2. Place it in your PATH
3. Implement `--pe-plugin-info` flag that returns JSON metadata:

```json
{
  "description": "Your plugin description",
  "version": "1.0.0",
  "commands": [
    {
      "name": "command",
      "description": "What it does",
      "usage": "pe yourplugin command [args]"
    }
  ]
}
```

### Available Plugins

#### promptfoo

The promptfoo compatibility plugin provides import/export functionality:

```bash
# Import promptfoo configuration
pe promptfoo import promptfoo-config.yaml -o pe-config.yaml

# Export to promptfoo format
pe promptfoo export pe-config.yaml -o promptfoo-config.yaml

# Convert between formats
pe promptfoo convert input.yaml output.json
```

---

## Pipeline Examples

PE is designed for Unix-style composition:

```bash
# Basic pipeline
pe eval config.yaml | pe filter --success | pe stats

# Complex analysis
pe eval config.yaml | \
  pe stream --select provider,score,cost | \
  pe filter --min-score 0.8 | \
  pe analyze --metric score --group-by provider

# Cost monitoring
pe eval config.yaml | \
  pe filter --max-cost 0.10 | \
  pe stream --select cost,provider | \
  pe analyze --metric cost

# A/B testing
pe eval variant-a.yaml -o a.json
pe eval variant-b.yaml -o b.json
pe diff a.json b.json --threshold 0.05
```

For more examples and advanced usage, see the [main documentation](README.md).