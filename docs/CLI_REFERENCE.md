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
| [`test`](#test) | Advanced testing framework | `pe test config.yaml --type property` |
| [`security`](#security) | Security testing (OWASP) | `pe security test --target prompt.txt` |
| [`build`](#build) | Build optimized prompts | `pe build config.yaml --target anthropic` |
| [`cat`](#cat) | Display prompt files | `pe cat prompts/analyze.prompt` |
| [`mod`](#mod) | Manage prompt modules | `pe mod list` |
| [`template`](#template) | Manage prompt templates | `pe template list` |
| [`prompt`](#prompt) | Manage prompt files | `pe prompt init analyze.prompt` |
| [`attest`](#attest) | Cryptographic attestations | `pe attest verify abc123` |
| [`cache`](#cache) | Cache management | `pe cache status` |
| [`profile`](#profile) | Profiling & observability | `pe profile start --type cpu` |
| [`distributed`](#distributed) | Distributed execution | `pe distributed start --role coordinator` |
| [`doc`](#doc) | Show prompt documentation | `pe doc math-solver` |
| [`extract`](#extract) | Extract XML-like tags | `pe run prompt.txt \| pe extract --tag answer` |
| [`get`](#get) | Get prompt file fields | `pe get summarize.txt variables` |
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

## cat

Display and inspect prompt files with variable substitution.

### Synopsis

```bash
pe cat [prompt_file] [flags]
```

### Description

The `cat` command reads and displays prompt files in various formats (txt, yaml, json) with support for variable substitution, component inspection, and multiple output formats.

### Arguments

- `prompt_file`: Path to the prompt file to display

### Flags

```bash
    --components           Show all prompt components (system, user, messages, metadata)
    --format string       Output format: text, yaml, json (default "text")
-h, --help                help for cat
    --interactive         Interactively prompt for variable values
    --metadata            Show metadata information
    --raw                 Display raw file content without processing
    --set strings         Set variable values (key=value format)
    --system              Show only the system prompt component
    --variables           Show available variables
```

### Examples

```bash
# Display a simple prompt file
pe cat prompts/analyze.prompt

# Show all components of a structured prompt
pe cat --components prompts/chat.yaml

# Substitute variables interactively
pe cat --interactive prompts/template.prompt

# Set variables via command line
pe cat --set topic=AI --set style=formal prompts/template.prompt

# Show only the system prompt component
pe cat --system prompts/chat.yaml

# Output as YAML with all metadata
pe cat --components --format yaml prompts/complex.yaml

# Raw display without any processing
pe cat --raw prompts/template.prompt
```

### Variable Substitution

Variables use Go template syntax:
- `{{.variable_name}}` - Standard Go template format
- `{{variable_name}}` - Legacy format (backward compatibility)
- Use `--set key=value` to provide values
- Use `--interactive` for guided input

---

## mod

Manage prompt modules using GitHub gists as a registry.

### Synopsis

```bash
pe mod [command]
```

### Description

The `mod` command manages prompt modules, similar to Go modules. PE uses GitHub gists as a module registry, with a root gist tracking forks containing prompt modules.

### Subcommands

#### mod list

List available modules from the registry.

```bash
pe mod list
```

#### mod get

Get module information or download a module.

```bash
pe mod get [module_path]
```

#### mod init

Initialize a new prompt module.

```bash
pe mod init [module_path]
```

#### mod publish

Publish a module to the registry.

```bash
pe mod publish [flags]
```

#### mod download

Download modules specified in pe.mod file.

```bash
pe mod download
```

#### mod tidy

Add missing and remove unused modules.

```bash
pe mod tidy
```

#### mod vendor

Copy dependencies to vendor directory.

```bash
pe mod vendor
```

#### mod search

Search for modules in the registry.

```bash
pe mod search [query]
```

### Examples

```bash
# Initialize a new module
pe mod init github.com/user/my-prompts

# List available modules
pe mod list

# Get a specific module
pe mod get github.com/user/prompt-templates

# Search for modules
pe mod search "code review"

# Publish your module
pe mod publish

# Download all dependencies
pe mod download

# Clean up unused dependencies
pe mod tidy

# Vendor dependencies
pe mod vendor
```

---

## template

Manage a library of reusable prompt templates.

### Synopsis

```bash
pe template [command]
```

### Description

The `template` command manages structured, parameterized prompts for common use cases like summarization, code review, creative writing, and data analysis.

### Subcommands

#### template list

List available templates.

```bash
pe template list
```

#### template show

Show template details.

```bash
pe template show [template_name]
```

#### template apply

Apply a template with variables.

```bash
pe template apply [template_name] [flags]
```

#### template create

Create a new template.

```bash
pe template create [template_name] [flags]
```

#### template search

Search templates.

```bash
pe template search [query]
```

#### template validate

Validate template files.

```bash
pe template validate [file...]
```

#### template import

Import templates from files.

```bash
pe template import [file...]
```

#### template export

Export templates to files.

```bash
pe template export [template_name] [flags]
```

#### template interactive

Interactive template selection and application.

```bash
pe template interactive
```

### Examples

```bash
# List all templates
pe template list

# Show template details
pe template show code-review

# Apply a template
pe template apply summarize --var input=article.txt

# Create a new template
pe template create my-template --file template.yaml

# Search for templates
pe template search "data analysis"

# Interactive mode
pe template interactive

# Validate template files
pe template validate templates/*.yaml

# Import templates
pe template import my-templates.yaml

# Export a template
pe template export code-review -o exported.yaml
```

---

## build

Build optimized prompts for production deployment.

### Synopsis

```bash
pe build [config/prompt] [flags]
```

### Description

The `build` command analyzes and optimizes prompts for specific providers, validates quality, and packages them for deployment.

### Arguments

- `config/prompt`: Configuration file or prompt to build

### Flags

```bash
    --bundle              Create bundle with dependencies
    --compress            Compress output
-h, --help                help for build
    --minify              Minify prompt to reduce tokens
-o, --output string       Output file path
    --target string       Target provider for optimization
    --targets strings     Multiple target providers
    --validate            Run validation checks
    --with-metadata       Include metadata file
```

### Examples

```bash
# Build from config
pe build config.yaml

# Build with specific output
pe build config.yaml -o production.txt

# Build for specific provider
pe build config.yaml --target anthropic

# Build with validation
pe build config.yaml --validate

# Create a bundle with dependencies
pe build config.yaml --bundle -o bundle.tar.gz

# Minify for token reduction
pe build config.yaml --minify -o optimized.txt

# Multi-target build
pe build config.yaml --targets openai,anthropic,google
```

---

## test

Advanced testing framework with systematic test-driven development.

### Synopsis

```bash
pe test [config_file] [flags]
pe test [command]
```

### Description

Comprehensive testing framework implementing property-based testing, regression detection, A/B testing with Bayesian analysis, and systematic test case generation.

### Subcommands

#### test create-suite

Create systematic test suite from prompts.

```bash
pe test create-suite [prompt_file] [flags]
```

#### test generate

Generate test cases automatically.

```bash
pe test generate [config_file] [flags]
```

#### test significance

Statistical significance testing for improvements.

```bash
pe test significance [baseline] [current] [flags]
```

### Flags

```bash
-b, --baseline string          Baseline file for regression testing
    --bootstrap int            Bootstrap samples for statistical analysis (default 1000)
-c, --config string           Configuration file path
-t, --type string             Test type: property, regression, systematic, ab-test, cross-validate (default "systematic")
    --confidence float        Confidence level for statistical tests (default 0.95)
    --comprehensive           Run all testing methods combined
```

### Testing Types

- **property**: Property-based testing for robustness validation
- **regression**: Performance regression detection with statistical significance
- **systematic**: Systematic test case execution
- **ab-test**: A/B testing with Bayesian statistical analysis
- **cross-validate**: Cross-validation between methods
- **significance**: Statistical significance testing
- **comprehensive**: All testing methods combined

### Examples

```bash
# Run systematic tests
pe test config.yaml

# Property-based testing
pe test config.yaml --type property

# Regression testing against baseline
pe test config.yaml --type regression --baseline baseline.json

# A/B testing
pe test config.yaml --type ab-test

# Statistical significance test
pe test significance baseline.json current.json

# Generate test suite
pe test create-suite prompt.txt -o test-suite.yaml

# Comprehensive testing
pe test config.yaml --comprehensive

# Generate test cases
pe test generate config.yaml --count 100
```

---

## security

Comprehensive security testing implementing OWASP LLM Top 10.

### Synopsis

```bash
pe security [command]
```

### Description

Advanced security testing covering OWASP LLM Top 10, automated vulnerability discovery, prompt injection detection, bias analysis, and privacy assessment.

### OWASP LLM Top 10 Coverage

- **LLM01**: Prompt Injection (Direct, Indirect, Context Poisoning)
- **LLM02**: Insecure Output Handling (Code injection, XSS, LDAP injection)
- **LLM03**: Training Data Poisoning (Backdoor detection, bias analysis)
- **LLM04**: Model Denial of Service (Resource exhaustion, infinite loops)
- **LLM05**: Supply Chain Vulnerabilities (Model provenance, dependency checks)
- **LLM06**: Sensitive Information Disclosure (PII, credentials, training data)
- **LLM07**: Insecure Plugin Design (Authorization bypass, input validation)
- **LLM08**: Excessive Agency (Privilege escalation, unauthorized actions)
- **LLM09**: Overreliance (Human oversight, verification mechanisms)
- **LLM10**: Model Theft (IP protection, model extraction attacks)

### Subcommands

```bash
pe security test          Run security tests
pe security monitor       Real-time security monitoring
pe security report        Generate compliance reports
```

### Examples

```bash
# Complete OWASP LLM Top 10 assessment
pe security test --target system_prompt.txt --owasp-complete

# Focused prompt injection testing
pe security test --target prompt.txt --categories prompt_injection

# Sensitive information disclosure testing
pe security test --target system.txt --categories sensitive_disclosure

# Real-time security monitoring
pe security monitor --realtime --categories all --alerts high

# Compliance reporting
pe security report --format pdf --standards owasp,nist,iso27001

# Comprehensive security scan
pe security test --target config.yaml --comprehensive
```

---

## attest

Manage cryptographic attestations for prompt executions.

### Synopsis

```bash
pe attest [command]
```

### Description

The `attest` command provides cryptographic attestations that prove prompt executions. Every prompt run can be cryptographically signed and chained, creating an immutable audit trail.

### Subcommands

#### attest init

Initialize attestation store.

```bash
pe attest init
```

#### attest list

List attestations.

```bash
pe attest list [flags]
```

#### attest show

Show attestation details.

```bash
pe attest show [attestation_id]
```

#### attest verify

Verify attestations.

```bash
pe attest verify [attestation_id]
```

#### attest key

Manage attestation keys.

```bash
pe attest key [subcommand]
```

#### attest export

Export attestation chain.

```bash
pe attest export [flags]
```

### Examples

```bash
# Initialize attestation store
pe attest init

# List all attestations
pe attest list

# Show specific attestation
pe attest show abc123

# Verify an attestation
pe attest verify abc123

# Manage keys
pe attest key generate
pe attest key list

# Export attestation chain
pe attest export -o attestations.json
```

---

## cache

Cryptographically signed content-addressed caching.

### Synopsis

```bash
pe cache [command]
```

### Description

Provides secure cache sharing across distributed nodes with cryptographic signatures and witness verification.

### Subcommands

#### cache status

Show cache status and statistics.

```bash
pe cache status
```

#### cache stats

Show detailed cache statistics.

```bash
pe cache stats
```

#### cache list

List cache entries.

```bash
pe cache list [flags]
```

#### cache inspect

View detailed cache entry information.

```bash
pe cache inspect [entry_id]
```

#### cache clear

Clear cache entries.

```bash
pe cache clear [flags]
```

#### cache verify

Verify cache integrity.

```bash
pe cache verify
```

#### cache export

Export cache entries to a bundle.

```bash
pe cache export [flags]
```

#### cache import

Import a cache bundle.

```bash
pe cache import [bundle_file]
```

### Examples

```bash
# Show cache status
pe cache status

# List all cache entries
pe cache list

# Inspect specific entry
pe cache inspect entry-abc123

# Clear old entries
pe cache clear --older-than 7d

# Verify cache integrity
pe cache verify

# Export cache bundle
pe cache export -o cache-bundle.tar.gz

# Import cache bundle
pe cache import cache-bundle.tar.gz

# Show detailed statistics
pe cache stats --by-provider
```

---

## profile

Advanced profiling and observability tools.

### Synopsis

```bash
pe profile [command]
```

### Description

Provides CPU profiling, memory profiling, distributed tracing, and metrics collection to analyze performance and identify bottlenecks.

### Subcommands

#### profile start

Start profiling.

```bash
pe profile start [flags]
```

#### profile stop

Stop profiling.

```bash
pe profile stop
```

#### profile status

Show profiling status.

```bash
pe profile status
```

#### profile report

Generate profiling report.

```bash
pe profile report [flags]
```

#### profile metrics

Metrics collection tools.

```bash
pe profile metrics [subcommand]
```

#### profile trace

Distributed tracing tools.

```bash
pe profile trace [subcommand]
```

### Examples

```bash
# Start CPU profiling
pe profile start --type cpu

# Start memory profiling
pe profile start --type memory

# Show profiling status
pe profile status

# Stop profiling
pe profile stop

# Generate report
pe profile report -o profile-report.html

# Collect metrics
pe profile metrics collect --interval 1s

# Enable distributed tracing
pe profile trace start --export-to jaeger
```

---

## prompt

Manage prompt files with initialization and metadata.

### Synopsis

```bash
pe prompt [command]
```

### Description

Commands for working with prompt files, similar to `go mod` for Go modules. Provides initialization, editing, validation, and metadata management.

### Subcommands

#### prompt init

Initialize a new prompt file with shebang and structure.

```bash
pe prompt init [filename] [flags]
```

#### prompt info

Display information about a prompt file.

```bash
pe prompt info [filename]
```

#### prompt help

Show usage information for a prompt file.

```bash
pe prompt help [filename]
```

#### prompt edit

Edit prompt file defaults and metadata.

```bash
pe prompt edit [filename] [flags]
```

#### prompt tidy

Clean up and validate prompt files.

```bash
pe prompt tidy [filename...]
```

### Examples

```bash
# Initialize new prompt file
pe prompt init analyze.prompt

# Show prompt information
pe prompt info analyze.prompt

# Show usage help for a prompt
pe prompt help analyze.prompt

# Edit prompt defaults
pe prompt edit analyze.prompt --provider anthropic --temperature 0.8

# Validate and clean up prompts
pe prompt tidy *.prompt
```

---

## distributed

Manage distributed execution across multiple nodes.

### Synopsis

```bash
pe distributed [command]
```

### Description

Enables distributed execution of prompts across multiple nodes for scalability and parallel processing.

### Subcommands

#### distributed start

Start a distributed execution node.

```bash
pe distributed start [flags]
```

#### distributed join

Join an existing distributed network.

```bash
pe distributed join [network_address] [flags]
```

#### distributed status

Show distributed network status.

```bash
pe distributed status
```

#### distributed stop

Stop distributed execution.

```bash
pe distributed stop
```

### Examples

```bash
# Start a coordinator node
pe distributed start --role coordinator --port 8080

# Start a worker node
pe distributed start --role worker

# Join existing network
pe distributed join coordinator.example.com:8080

# Show network status
pe distributed status

# Stop distributed execution
pe distributed stop
```

---

## doc

Show documentation extracted from prompt files.

### Synopsis

```bash
pe doc [prompt] [variable] [flags]
```

### Description

Similar to `go doc`, this command displays documentation embedded in prompt files, including variable descriptions and examples.

### Arguments

- `prompt`: Prompt file name (without extension)
- `variable`: Specific variable to document (optional)

### Flags

```bash
    --all        Show all documentation
    --examples   Show examples (default true)
-h, --help       help for doc
    --short      Show only brief descriptions
```

### Examples

```bash
# List all documented prompts
pe doc

# Show documentation for a prompt
pe doc math-solver

# Show variable documentation
pe doc math-solver.EXPRESSION

# Show all prompts with full documentation
pe doc --all

# Show brief descriptions only
pe doc --short
```

---

## extract

Extract content from XML-like tags in LLM outputs.

### Synopsis

```bash
pe extract [flags]
```

### Description

Supports extracting single or multiple tags, nested tags, and XPath-like selectors. Can output in different formats and validate against schemas.

### Flags

```bash
    --all                Extract all occurrences of the tag
    --attr string        Attribute filter (e.g., 'type=final')
    --end string         Custom end delimiter
-f, --format string      Output format (text, json, xml) (default "text")
-h, --help               help for extract
    --nested             Include nested tags in extraction
-o, --output string      Output file (default: stdout)
    --start string       Custom start delimiter
    --stream             Stream extraction mode
    --tag string         Tag to extract (e.g., 'answer', 'thinking')
    --tags string        Multiple tags to extract (comma-separated)
    --transform string   Transform command to apply to extracted content
    --validate string    Schema file for validation
    --xpath string       XPath-like selector (e.g., '/response/answer')
```

### Examples

```bash
# Extract answer tags from stdin
echo "<answer>42</answer>" | pe extract --tag answer

# Extract multiple tags
pe run prompt.txt | pe extract --tags "thinking,answer,confidence"

# Extract with XPath
pe run prompt.txt | pe extract --xpath "/response/answer"

# Extract all occurrences
pe run prompt.txt | pe extract --tag item --all

# Output as JSON
pe run prompt.txt | pe extract --tag answer --format json

# Stream mode for large outputs
pe run long-prompt.txt | pe extract --tag chunk --stream

# Validate against schema
pe run prompt.txt | pe extract --tag output --validate schema.json
```

---

## get

Get specific fields from prompt files.

### Synopsis

```bash
pe get [prompt-file] [field] [flags]
```

### Description

Extract specific fields or all information from prompt files, including prompt text, system prompts, variables, defaults, and examples.

### Arguments

- `prompt-file`: Path to the prompt file
- `field`: Field to extract (prompt, system-prompt, variables, defaults, examples, variants, all)

### Available Fields

- **prompt**: The main prompt text
- **system-prompt**: The system prompt
- **variables**: List of template variables
- **defaults**: Default values for variables
- **examples**: Examples section
- **variants**: List all variants
- **all**: Get all information (default)

### Flags

```bash
-h, --help             help for get
    --json             Output in JSON format
    --keys             List available section keys
    --variant string   Apply variant before getting field
```

### Examples

```bash
# Get the main prompt
pe get summarize.txt prompt

# Get system prompt
pe get summarize.txt system-prompt

# Get variables used in the prompt
pe get summarize.txt variables

# Get all information as JSON
pe get summarize.txt --json

# Get specific variant
pe get summarize.txt --variant academic prompt

# List available keys
pe get summarize.txt --keys

# Get defaults
pe get summarize.txt defaults
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