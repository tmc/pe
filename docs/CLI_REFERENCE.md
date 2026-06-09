<!-- Legacy draft: this file contains generated and historical command notes. Verify current behavior with `pe --help` and command-specific help before relying on listed flags or examples. -->

# PE CLI Reference

Complete reference for all PE commands, options, and usage patterns.

## Global Options

These options are available for all commands:

```bash
pe [global-options] command [command-options]
```

Global options:
- `--help, -h`: Show help information

Use `pe version` to show version information.

## Commands Overview

| Command | Purpose | Example |
|---------|---------|---------|
| [`run`](#run) | Execute prompt immediately | `pe run "What is AI?"` |
| [`run-text`](#run-text) | Render executable text safely | `pe run-text review.prompt --var topic=release` |
| [`eval`](#eval) | Run prompt evaluations | `pe eval config.yaml` |
| [`view`](#view) | View results in browser | `pe view` |
| [`serve`](#serve) | Serve localhost API | `pe serve --addr 127.0.0.1:8080` |
| [`interactive`](#interactive) | Start REPL mode | `pe interactive` |
| [`watch`](#watch) | Auto-rerun on changes | `pe watch config.yaml` |
| [`ask`](#ask) | Single prompt query | `pe ask "What is AI?"` |
| [`benchmark`](#benchmark) | Performance testing | `pe benchmark config.yaml` |
| [`test`](#test) | Advanced testing framework | `pe test config.yaml --type property` |
| [`version`](#version) | Print version information | `pe version` |
| [`security`](#security) | Security testing (OWASP) | `pe security test prompt.txt` |
| [`build`](#build) | Write prompts, metadata, and bundles | `pe build config.yaml --target anthropic` |
| [`compose`](#compose) | Compose prompt components | `pe compose context.txt instruction.txt --style cot` |
| [`config`](#config) | Inspect PE configuration | `pe config list` |
| [`cat`](#cat) | Display prompt files | `pe cat prompts/analyze.prompt` |
| [`mod`](#mod) | Manage prompt modules | `pe mod list` |
| [`template`](#template) | Manage prompt templates | `pe template list` |
| [`prompt`](#prompt) | Manage prompt files | `pe prompt init analyze.prompt` |
| [`attest`](#attest) | Unsigned local file manifests | `pe exp attest manifest .` |
| [`cache`](#cache) | Local content-addressed cache | `pe exp cache key file.txt` |
| [`profile`](#profile) | Profiling & observability | `pe profile start --type cpu` |
| [`distributed`](#distributed) | Local bounded task scheduler | `pe exp distributed tasks.json --workers 2` |
| [`consensus`](#consensus) | Local weighted vote aggregation | `pe exp consensus --input votes.json` |
| [`doc`](#doc) | Show prompt documentation | `pe doc math-solver` |
| [`expand`](#expand) | Resolve config file references and globs | `pe expand config.yaml` |
| [`extract`](#extract) | Extract XML-like tags | `pe run prompt.txt \| pe extract --tag answer` |
| [`get`](#get) | Get prompt file fields | `pe get summarize.txt variables` |
| [`collect`](#collect) | Gather parallel results | `pe collect --jobs 10` |
| [`completion`](#completion) | Shell completions | `pe completion bash` |
| [`edit`](#edit) | Edit prompt files | `pe edit prompt.txt --set-prompt "text"` |
| [`eval-prompt`](#eval-prompt) | Run prompt evals | `pe eval-prompt prompt.txt` |
| [`exp`](#exp) | Prototype command group | `pe exp --help` |
| [`experimental`](#experimental) | Research commands | `pe experimental optimize config.yaml` |
| [`optimize`](#optimize) | Optimize prompts with metaprompting | `pe optimize --prompt "Summarize" --method pe2` |
| [`evolve`](#evolve) | Optimize prompts with evolutionary search | `pe evolve prompt.txt --population 10` |
| [`semantic`](#semantic) | Semantic optimization commands | `pe semantic analyze prompt.txt` |
| [`fusion`](#fusion) | Aggregate provider outputs | `pe fusion --input votes.json` |
| [`gaso`](#gaso) | Optimize agentic systems | `pe gaso --system system.yaml --objective accuracy` |
| [`pe2`](#pe2) | PE2 prompt optimization shortcut | `pe pe2 --prompt "Classify text"` |
| [`textgrad`](#textgrad) | Textual-gradient optimization shortcut | `pe textgrad --prompt "Classify text"` |
| [`push`](#push) | Push module to registry | `pe push tmc/hello` |
| [`reduce`](#reduce) | Aggregate results | `pe reduce --sum cost` |
| [`work`](#work) | Workspace management | `pe work init` |
| [`help`](#help) | Get command help | `pe help run` |
| [`stream`](#stream) | Process results stream | `pe eval config.yaml \| pe stream` |
| [`filter`](#filter) | Filter results | `pe eval config.yaml \| pe filter --contains pass` |
| [`analyze`](#analyze) | Statistical analysis | `pe eval config.yaml \| pe analyze` |
| [`stats`](#stats) | Quick statistics | `pe eval config.yaml \| pe stats` |
| [`diff`](#diff) | Compare results | `pe diff old.json new.json` |
| [`fmt`](#fmt) | Format prompt files | `pe fmt prompt.txt` |
| [`vet`](#vet) | Validate configs | `pe vet config.yaml` |
| [`convert`](#convert) | Convert formats | `pe convert config.yaml config.json` |
| [`init`](#init) | Initialize `.pe/` project files | `pe init` |
| [`plugin`](#plugin) | Manage plugins | `pe plugin list` |

---

## run-text

Validate and render executable text without invoking providers, tools, shell
commands, or network requests.

### Synopsis

```bash
pe run-text [file|-] [flags]
```

### Description

Plain text is valid by default. A file may add `pe.text.v1` or `pe.workflow.v1`
front matter to declare inputs, metadata, safety policy, and placement.
`run-text` validates that contract and renders Go template variables from
explicit `--var` bindings. Front matter `imports` can name local text
components, which are rendered through the safe template function
`{{ import "name" }}`.

### Flags

```bash
    --check                 Validate without rendering
    --var stringToString    Template variables
```

### Examples

```bash
pe run-text review.prompt --var topic=release
pe run-text review.prompt --check
cat review.prompt | pe run-text - --var topic=release
```

---

## run

Execute a prompt immediately using the inference API.

### Synopsis

```bash
pe run [prompt or file] [flags]
```

### Description

The `run` command executes prompts directly without needing a configuration file. It's designed for quick testing and iteration, similar to `go run` for Go programs. Supports both direct prompt strings and prompt files.

If `pe.mod` exists in the current directory, `pe run` enforces provider
denials before creating a provider. Exact provider denials and known provider
classes such as `remote` are rejected at runtime. `placement { network false }`
also rejects known remote providers.

### Arguments

- `prompt or file`: Either a prompt string or path to a file containing the prompt

### Flags

```bash
    --provider string        Provider to use (default "cgpt")
    --stream                 Stream the response (default true)
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

# Streaming mode
pe run "Tell me a story about a robot" --stream=false

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

If `pe.mod` denies the `write` tool capability, `pe eval -o` and
`pe eval --save-db` fail before writing files. Default stdout output remains
read-only and is allowed.

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

Opens a local web browser interface to explore evaluation results. Can view
results saved by eval ID or from a specific file. Use `--promptfoo` to
explicitly delegate to the promptfoo CLI viewer.

### Arguments

- `eval_id`: Specific evaluation ID to view (optional)

### Flags

```bash
-f, --file string     View results from specific file
-p, --port int        Port for local server (default 8080)
    --promptfoo       Open the promptfoo CLI viewer instead of the local viewer
-y, --yes             Pass -y to promptfoo when used with --promptfoo
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

# Explicitly open promptfoo's viewer
pe view --promptfoo --yes
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

If `pe.mod` denies the `write` tool capability, `pe benchmark -o` fails
before writing the result file. Stdout and `--go-bench` output remain read-only.

### Arguments

- `config_file`: Configuration file for benchmarking

### Flags

```bash
-i, --iterations int        Number of benchmark iterations (default 3)
-n, --concurrency int       Number of concurrent benchmark runs (default 1)
-c, --config string         Path to configuration file
-f, --format string         Output format: json, yaml, csv, or text (default "json")
-o, --output string         Write results to file
    --go-bench              Output in Go benchmark format
```

### Examples

```bash
# Basic benchmark
pe benchmark config.yaml

# Multiple iterations
pe benchmark config.yaml --iterations 10 --concurrency 4

# Save results
pe benchmark config.yaml --format json -o benchmark.json

# Go benchmark format
pe benchmark config.yaml --go-bench
```

---

## stream

Process evaluation results as a stream for pipeline processing.

### Synopsis

```bash
pe eval config.yaml | pe stream [flags]
```

### Description

Passes input through line by line. Designed for simple Unix pipeline composition.

### Flags

This command currently has no command-specific flags.

### Examples

```bash
# Pass evaluation output through
pe eval config.yaml | pe stream
```

---

## filter

Filter evaluation results based on various conditions.

### Synopsis

```bash
pe eval config.yaml | pe filter [flags]
```

### Description

Filters text streams by pattern, substring, simple matches, or JSON fields.

### Flags

```bash
    --pattern string        Filter by substring pattern
    --contains string       Filter by substring
    --match string          Match pattern
    --json string           Extract JSON field
    --field string          Extract field
    --transform string      Transform output
    --if-contains string    Conditional contains
    --then string           Then command
    --else string           Else command
```

### Examples

```bash
# Keep lines containing pass
pe eval config.yaml | pe filter --contains pass

# Filter by pattern
pe eval config.yaml | pe filter --pattern openai

# Extract a JSON field from newline-delimited JSON
cat results.ndjson | pe filter --json .response

# Transform text
echo "HELLO" | pe filter --transform lowercase
```

---

## analyze

Compute statistics and insights from evaluation results.

### Synopsis

```bash
pe eval config.yaml | pe analyze [flags]
```

### Description

Analyzes text from stdin. It can report basic counts, selected text metrics, or JSON output.

### Flags

```bash
    --type string           Type of analysis
    --metrics string        Comma-separated metrics
    --format string         Output format
```

### Examples

```bash
# Basic counts
pe eval config.yaml | pe analyze

# Text metrics
cat response.txt | pe analyze --metrics readability,sentiment

# JSON output
cat response.txt | pe analyze --format json
```

---

## stats

Show quick statistics from evaluation results.

### Synopsis

```bash
pe eval config.yaml | pe stats [flags]
```

### Description

Displays quick statistics from evaluation results.

### Flags

This command currently has no command-specific flags.

### Examples

```bash
# Quick stats
pe eval config.yaml | pe stats

pe eval config.yaml | pe stats
```

---

## serve

Serve a small localhost-first HTTP API.

### Synopsis

```bash
pe serve [flags]
```

### Description

Starts a local HTTP API. The server binds to `127.0.0.1:8080` by default.
Use `--addr` explicitly to bind somewhere else.

### Flags

```bash
--addr string   listen address (default "127.0.0.1:8080")
```

### Examples

```bash
pe serve
pe serve --addr 127.0.0.1:9090
```

---

## diff

Compare two evaluation results to detect changes and regressions.

### Synopsis

```bash
pe diff [baseline] [current] [flags]
```

### Description

Compares evaluation results to identify differences in scores, latency, costs, and other metrics. Useful for regression testing.

### Arguments

- `baseline`: Baseline results file
- `current`: Current results file, or `-` to read from stdin

### Flags

```bash
--fail-on-change                  Exit non-zero if any compared metric changes
--fail-on-regression              Exit non-zero if regression metrics exceed thresholds
-f, --format string               Output format: text or json (default "text")
--max-error-increase int          Allowed error count increase with --fail-on-regression
--max-failure-increase int        Allowed failure count increase with --fail-on-regression
--max-latency-increase-ms float   Allowed average latency increase in milliseconds with --fail-on-regression
--max-pass-rate-drop float        Allowed pass-rate drop in percentage points with --fail-on-regression
--max-score-drop float            Allowed average score drop with --fail-on-regression
--max-token-increase int32        Allowed token total increase with --fail-on-regression
--statistical                     Print statistical significance for pass rate, score, and latency
```

### Examples

```bash
# Compare files
pe diff baseline.json current.json

# From pipeline
pe eval config.yaml -o current.json && pe diff baseline.json current.json

# JSON output
pe diff --format json baseline.json current.json

# CI regression gate with explicit thresholds
pe diff --fail-on-regression --max-pass-rate-drop 5 --max-latency-increase-ms 100 baseline.json current.json

# Add informational significance checks
pe diff --statistical baseline.json current.json
```

---

## fmt

Format and optionally convert configuration files.

### Synopsis

```bash
pe fmt [file...] [flags]
```

### Description

Formats prompt files according to the selected style.

### Arguments

- `file...`: Prompt files to format

### Flags

```bash
-w, --write               Write result to source file instead of stdout
    --check               Check whether files are formatted
    --style string        Formatting style: anthropic, openai, standard
    --fix                 Automatically fix common issues
```

### Examples

```bash
# Format to stdout
pe fmt prompt.txt

# Format in-place
pe fmt prompt.txt --write

# Check formatting
pe fmt prompt.txt --check

# Format multiple files
pe fmt *.prompt --write --style standard
```

---

## version

Print PE version information.

### Synopsis

```bash
pe version
```

### Examples

```bash
pe version
pe --version
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
pe init [flags]
```

### Description

Creates `.pe/` project files, similar to `git init`.

### Flags

```bash
    --force           Force reinitialization even if `.pe` exists
```

### Examples

```bash
# Create default config
pe init

# Overwrite existing
pe init --force
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
  openai (built-in)
  anthropic (built-in)
  promptfoo v0.1.0
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

The Promptfoo compatibility plugin is exposed when a `pe-promptfoo` executable is on your `PATH`.

Build it from this repository:

```bash
go build -o ~/bin/pe-promptfoo ./plugins/promptfoo
```

Once installed, the plugin provides config import/export/conversion commands:

```bash
# Import promptfoo configuration
pe promptfoo import promptfoo-config.yaml -o pe-config.yaml

# Export to promptfoo format
pe promptfoo export pe-config.yaml -o promptfoo-config.yaml

# Convert between formats
pe promptfoo convert input.yaml output.json --direction pe-to-promptfoo
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

Manage prompt modules using a local, HTTP, or GitHub registry.

### Synopsis

```bash
pe mod [command]
```

### Description

The `mod` command manages prompt modules, similar to Go modules. PE defaults to
a local registry at `$HOME/.pe/registry`. Set `PE_REGISTRY_TYPE` to `local`,
`http`, or `github` to select a registry backend.

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

Registered command for future registry publishing. Current registry backends do
not publish modules.

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
pe mod tidy [flags]
```

Flags:

```bash
--json        Write dependency audit report as JSON
-w, --write   Update pe.mod; add only refs with one explicit version
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

#### mod vet

Validate `pe.mod` capability, placement, and policy blocks. With file
arguments, also validate executable text metadata against the module policy.
When `policy { composition strict }` is set, cached dependency `pe.mod` files
are checked so dependencies cannot request providers, tools, data, prompts, or
network placement denied by the parent module.

```bash
pe mod vet [file...]
```

### Examples

```bash
# Initialize a new module
pe mod init github.com/user/my-prompts

# Validate module policy and an executable text file
pe mod vet review.prompt

# List available modules
pe mod list

# Get a specific module
pe mod get github.com/user/prompt-templates

# Search for modules
pe mod search "code review"

# Publishing is registered for future registry backends.
# Current registry backends return an explicit error.

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

Write prompts, metadata, and optional bundles for deployment.

### Synopsis

```bash
pe build [config/prompt] [flags]
```

### Description

The `build` command reads a prompt or config, applies supported provider
formatting, writes output files, and can package a directory into a bundle.
Validation is local and deterministic. Some provider-specific formatting paths
return explicit not-implemented errors.

If `pe.mod` denies the `write` tool capability, build artifact writes fail.
`pe build --validate` remains read-only and is allowed.

### Arguments

- `config/prompt`: Configuration file or prompt to build

### Flags

```bash
    --bundle              Create bundle with dependencies
    --compress            Compress output
-h, --help                help for build
    --minify              Minify prompt to reduce tokens
-o, --output string       Output file path
    --target string       Target provider for formatting
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

# Multi-target build with supported formatters
pe build config.yaml --targets openai,anthropic
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

Comprehensive testing framework implementing property-based testing, regression detection, and systematic test suite creation.

### Subcommands

#### test create-suite

Create systematic test suite from prompts.

```bash
pe test create-suite [prompt_file] [flags]
```

### Flags

```bash
-b, --baseline string          Baseline file for regression testing
-c, --config string           Configuration file path
-t, --type string             Test type: property, regression, systematic, comprehensive (default "systematic")
    --comprehensive           Run all testing methods combined
```

### Testing Types

- **property**: Property-based testing for robustness validation
- **regression**: Performance regression detection
- **systematic**: Systematic test case execution
- **comprehensive**: All testing methods combined

### Examples

```bash
# Run systematic tests
pe test config.yaml

# Property-based testing
pe test config.yaml --type property

# Regression testing against baseline
pe test config.yaml --type regression --baseline baseline.json

# Generate test suite
pe test create-suite prompt.txt -o test-suite.yaml

# Comprehensive testing
pe test config.yaml --comprehensive
```

---

## security

OWASP-oriented security testing with incomplete analyses reported explicitly or
failed closed.

### Synopsis

```bash
pe security [command]
```

### Description

Security testing covers prompt-injection and related local checks. Data
poisoning and supply-chain checks use local response-pattern indicators; remote
provenance checks and model-assisted adjudication remain future work.

### OWASP-Oriented Checks

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
# OWASP-oriented assessment
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

Create, sign, and verify deterministic SHA-256 manifests for local files.

### Synopsis

```bash
pe exp attest [command]
```

### Description

`pe exp attest` writes and verifies unsigned local manifests. It can also wrap
an unsigned manifest in an Ed25519 signed envelope. Unsigned manifests detect
file content changes, missing files, and manifest tampering. Signed envelopes
bind the manifest payload to a public key, but do not by themselves prove key
ownership, origin, or freshness.

### Commands

```bash
pe exp attest manifest [file-or-dir...] [--root .]
pe exp attest verify <manifest.json> [--root .]
pe exp attest keygen
pe exp attest sign <manifest.json> --private-key <hex>
pe exp attest verify-signed <signed-manifest.json> [--root .] [--public-key <hex>]
```

### Examples

```bash
pe exp attest manifest prompts/ > manifest.json
pe exp attest verify manifest.json
pe exp attest keygen > attest-key.json
pe exp attest sign --private-key "$(jq -r .private_key attest-key.json)" manifest.json > signed-manifest.json
pe exp attest verify-signed --public-key "$(jq -r .public_key attest-key.json)" signed-manifest.json
```

---

## cache

Inspect a local content-addressed cache for files and unsigned manifests.

### Synopsis

```bash
pe exp cache [command]
```

### Description

`pe exp cache` stores and verifies content by SHA-256 digest. The cache is
local only and unsigned; it does not prove identity, origin, or freshness.

### Commands

```bash
pe exp cache key <file> [--manifest]
pe exp cache put <file> [--cache-dir .pe/cache]
pe exp cache get <sha256> [--cache-dir .pe/cache]
pe exp cache verify <sha256> [--cache-dir .pe/cache]
pe exp cache manifest put <manifest.json> [--cache-dir .pe/cache]
pe exp cache manifest verify <sha256> [--cache-dir .pe/cache] [--root .]
```

### Examples

```bash
pe exp cache key prompt.txt
pe exp cache put prompt.txt
pe exp cache verify <sha256>
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

Run local deterministic tasks with bounded concurrency.

### Synopsis

```bash
pe exp distributed <tasks.json> [flags]
```

### Description

`pe exp distributed` runs a local scheduler prototype from a JSON task file.
It does not start daemons, open network connections, or coordinate remote
workers.

### Flags

```bash
--format string    output format: json or text (default "json")
--timeout-ms int   optional timeout in milliseconds
--workers int      maximum concurrent local tasks (default 1)
```

### Examples

```bash
pe exp distributed tasks.json --workers 2 --format text
pe exp distributed tasks.json --workers 2 --timeout-ms 5000
```

---

## consensus

Aggregate local provider votes deterministically.

### Synopsis

```bash
pe exp consensus [--input votes.json] [flags]
```

### Description

`pe exp consensus` reads local vote JSON. Rows with an error are reported and
excluded; successful rows are passed to the local weighted majority aggregator.

### Flags

```bash
-i, --input string    input JSON file, or - for stdin
-o, --output string   output JSON file, or - for stdout (default "-")
```

### Examples

```bash
pe exp consensus --input votes.json --output consensus.json
pe exp consensus --input votes.json --output -
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

## expand

Expand a configuration file by resolving file references and globs.

### Synopsis

```bash
pe expand [config_file] [flags]
```

### Flags

```bash
-o, --output string   Output file (JSON)
```

### Examples

```bash
pe expand config.yaml
pe expand config.yaml --output expanded.json
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

## collect

Gather results from parallel or asynchronous operations.

### Synopsis

```bash
pe collect [flags]
```

### Description

The `collect` command gathers results from parallel/async operations, useful when running multiple prompts concurrently.

### Flags

```bash
-h, --help       help for collect
    --jobs int   Number of jobs to collect
```

### Examples

```bash
# Collect results from parallel operations
pe collect

# Collect specific number of jobs
pe collect --jobs 10
```

---

## completion

Generate shell completion scripts for pe.

### Synopsis

```bash
pe completion [command]
```

### Description

Generate autocompletion scripts for various shells to enable tab completion for PE commands and flags.

### Available Commands

```bash
bash        Generate the autocompletion script for bash
fish        Generate the autocompletion script for fish
powershell  Generate the autocompletion script for powershell
zsh         Generate the autocompletion script for zsh
```

### Examples

```bash
# Bash completion
pe completion bash > /etc/bash_completion.d/pe

# Zsh completion
pe completion zsh > "${fpath[1]}/_pe"

# Fish completion
pe completion fish > ~/.config/fish/completions/pe.fish

# PowerShell completion
pe completion powershell | Out-String | Invoke-Expression
```

---

## edit

Edit prompt files programmatically with scriptable modifications.

### Synopsis

```bash
pe edit [prompt-file] [flags]
```

### Description

Inspired by `go mod edit`, provides a programmatic interface for editing prompt files without manual text editing. All modifications preserve the prompt format.

### Flags

```bash
    --add-default string         Add a default (format: key=value)
    --add-section string         Add a section
    --add-variant string         Add a variant
    --append-prompt string       Append text to the main prompt
    --fmt                        Format the prompt file
-h, --help                       help for edit
    --json                       Output the prompt in JSON format
    --module string              Add module dependency to go.mod
    --prepend-prompt string      Prepend text to the main prompt
    --print                      Print the result instead of writing to file
    --remove-default string      Remove a default by key
    --remove-section string      Remove a section
    --remove-variant string      Remove a variant
    --section-content string     Content for the section (use with --add-section)
    --set-defaults string        Set defaults (format: key1=val1&key2=val2)
    --set-examples string        Set the examples section
    --set-prompt string          Set the main prompt text
    --set-system-prompt string   Set the system prompt
    --variant-cmd string         Commands for the variant (use with --add-variant)
```

### Examples

```bash
# Set the main prompt
pe edit prompt.txt --set-prompt "Summarize this document"

# Set the system prompt
pe edit prompt.txt --set-system-prompt "You are a helpful assistant"

# Add a variant
pe edit prompt.txt --add-variant academic --variant-cmd "extend-system-prompt 'Use academic language'"

# Add a section
pe edit prompt.txt --add-section examples --section-content "Example 1: ..."

# Output as JSON
pe edit prompt.txt --json

# Set defaults
pe edit prompt.txt --set-defaults "lang=en&model=gpt-4"
pe edit prompt.txt --add-default lang=en

# Format the file
pe edit prompt.txt --fmt
```

---

## eval-prompt

Run evaluations defined in a prompt file's evals section.

### Synopsis

```bash
pe eval-prompt [prompt-file] [flags]
```

### Description

Executes test cases defined in the `-- evals --` section of a prompt file. Tests should be in YAML format with variables and assertions.

### Flags

```bash
-h, --help              help for eval-prompt
    --json              Output results as JSON
    --output string     Output file for results
    --provider string   LLM provider to use
    --variant string    Apply variant before running evals
```

### Evals Format

```yaml
-- evals --
tests:
  - vars:
      input: "test input"
    assert:
      - type: contains
        value: "expected"
  - vars:
      input: "another test"
    assert:
      - type: llm_rubric
        value: "Should be concise"
```

### Examples

```bash
# Run evals from prompt file
pe eval-prompt prompt.txt

# Use specific provider
pe eval-prompt prompt.txt --provider anthropic:claude-3-sonnet

# Apply variant before evaluating
pe eval-prompt prompt.txt --variant formal

# Output results as JSON
pe eval-prompt prompt.txt --json -o results.json
```

---

## compose

Compose prompts from modular components.

### Synopsis

```bash
pe compose [component files...] [flags]
```

### Common Flags

```bash
    --style string       Composition style (default, cot, few-shot, structured, conversational, dspy)
    --output string      Output file for composed prompt
    --validate           Validate component compatibility
    --coherence          Add local coherence score to composed output
    --coherence-check    Report local semantic-overlap and style-consistency scores
    --import string      Import local component file, directory, or txtar archive
    --optimize           Apply local composition polish after composition
```

`pe compose --import` writes component files under `components/`. If `pe.mod`
denies the `write` tool capability, imports fail before writing files.

### Examples

```bash
pe compose context.txt instruction.txt examples.txt --style cot
pe compose components/ --style few-shot --validate --output prompt.txt
pe compose context.txt instruction.txt --coherence
pe compose context.txt instruction.txt --coherence-check
pe compose --import components.txtar
```

---

## config

Inspect, validate, and update PE configuration.

### Synopsis

```bash
pe config [command]
```

### Subcommands

```bash
docs        Print configuration reference
get         Print one configuration value
list        Print merged configuration
migrate     Migrate a config file to canonical PE YAML
set         Set one configuration value in a config file
validate    Validate merged configuration
```

### Examples

```bash
pe config list
pe config get provider.default
pe config validate pe.yaml
```

---

## exp

Access prototype commands under active development.

### Synopsis

```bash
pe exp [command]
```

### Description

`pe exp` groups prototype commands. Current entries include `attest`, `cache`,
`compose`, `consensus`, `distributed`, and `optimize`.

### Examples

```bash
pe exp --help
pe exp optimize --help
pe exp compose --help
pe exp distributed --help
pe exp consensus --help
pe exp attest manifest prompts/ > manifest.json
```

### Local Optimization

`pe exp optimize` selects prompt variants with local deterministic scores. It
accepts input JSON with a seed plus variants or rounds, or promptfoo evaluation
JSON from `results.prompts[].metrics.score`.

```bash
pe exp optimize --input input.json --output result.json
pe exp optimize --scores eval-results.json --max-rounds 3
```

---

## optimize

Optimize prompts using metaprompting methods.

### Synopsis

```bash
pe optimize [prompt-file] [flags]
```

### Common Flags

```bash
    --prompt string      Initial prompt to optimize
-m, --method string      Method: standard, pe2, apex, textgrad, hybrid
-i, --iterations int     Number of optimization iterations
    --provider string    Provider name
-o, --output string      Output JSON file
```

### Examples

```bash
pe optimize --prompt "Analyze sentiment" --method pe2 --iterations 5
pe optimize complex-system.txt --method apex --output results.json
```

---

## evolve

Optimize prompts with evolutionary algorithms.

### Synopsis

```bash
pe evolve [prompt-file] [flags]
```

### Common Flags

```bash
-n, --population int        Population size
-g, --generations int       Number of generations
    --objectives string     Comma-separated objectives
-o, --output string         Output JSON file
```

### Examples

```bash
pe evolve prompt.txt
pe evolve prompt.txt --objectives accuracy,conciseness --generations 50
```

---

## semantic

Run semantic analysis and optimization subcommands.

### Synopsis

```bash
pe semantic [command]
```

### Subcommands

```bash
analyze     Analyze semantic structure and dependencies
backprop    Apply semantic backpropagation
benchmark   Benchmark semantic optimization
descent     Apply semantic gradient descent
flow        Analyze semantic information flow
gaso        Graph-based Agentic System Optimization
gradients   Compute semantic gradients
monitor     Monitor semantic drift
```

### Examples

```bash
pe semantic analyze prompt.txt
pe semantic gradients prompt.txt
```

---

## fusion

Aggregate provider outputs with deterministic consensus.

### Synopsis

```bash
pe fusion [--input votes.json] [flags]
```

### Flags

```bash
-i, --input string     Input JSON file, or - for stdin
-o, --output string    Output JSON file, or - for stdout
```

### Examples

```bash
pe fusion --input votes.json
cat votes.json | pe fusion --input - --output result.json
```

---

## gaso

Optimize multi-component agentic systems with semantic gradients.

### Synopsis

```bash
pe gaso [flags]
```

### Common Flags

```bash
-s, --system string       System definition file
-b, --objective string    Optimization objective
-i, --iterations int      Number of iterations
-f, --format string       Output format: json, yaml, table
-o, --output string       Output file
```

### Examples

```bash
pe gaso --system system.yaml --objective accuracy
pe gaso --system system.yaml --multi-objective --format table
```

---

## pe2

Optimize prompts with the PE2 method.

### Synopsis

```bash
pe pe2 [prompt-file] [flags]
```

### Examples

```bash
pe pe2 --prompt "Classify text"
pe pe2 prompt.txt --iterations 5 --output result.json
```

---

## textgrad

Optimize prompts with textual gradients.

### Synopsis

```bash
pe textgrad [prompt-file] [flags]
```

### Examples

```bash
pe textgrad --prompt "Classify text"
pe textgrad prompt.txt --iterations 5 --output result.json
```

---

## experimental

Access experimental prompt engineering research commands.

### Synopsis

```bash
pe experimental [command]
```

### Description

Experimental commands for prompt engineering research and prototypes. These
commands may be unstable, slow, or produce inconsistent results. Do not depend
on them in production automation without pinning and testing the exact release.

### Available Commands

```bash
analyze     Analyze text with various metrics
compose     Compose prompts from verified components with type-safe composition
evolve      Optimize prompts using evolutionary algorithms (NSGA-II)
extract     Extract content from XML-like tags
filter      Filter and transform pipeline outputs
metrics     Calculate advanced evaluation metrics for generated text
optimize    Optimize prompts using metaprompting techniques
playground  Launch interactive web playground for prompt engineering
plugin      Manage PE plugins
semantic    Semantic backpropagation and Graph-based Agentic System Optimization (GASO)
stream      Stream process LLM outputs
synthesize  Generate prompts using structured program synthesis
```

### Examples

```bash
# Optimize prompts using metaprompting
pe experimental optimize config.yaml

# Evolutionary optimization with NSGA-II
pe experimental evolve --population 10

# Semantic analysis
pe experimental semantic analyze prompt.txt

# Launch interactive playground
pe experimental playground

# Prompt synthesis workflow
pe experimental synthesize --help
```

---

## push

Push a prompt module to the GitHub gist registry.

### Synopsis

```bash
pe push [module] [flags]
```

### Description

Creates or updates a GitHub gist with your module content and registers it in the root registry gist. Requires `GITHUB_TOKEN` environment variable.

### Arguments

- `module`: Module name (e.g., `tmc/hello`, `myorg/summarize`)

### Flags

```bash
    --draft    Save as draft without pushing
-h, --help     help for push
    --public   Make the gist public
    --update   Update existing gist
```

### Examples

```bash
# Push a module
pe push tmc/hello

# Push as public gist
pe push myorg/summarize --public

# Update existing module
pe push tmc/hello --update

# Save as draft
pe push myorg/test --draft
```

---

## reduce

Aggregate pipeline results using various reduction operations.

### Synopsis

```bash
pe reduce [flags]
```

### Description

Reduces and aggregates evaluation results from a pipeline, supporting various mathematical and statistical operations.

### Flags

```bash
-h, --help         help for reduce
    --sum string   Sum operation
```

### Examples

```bash
# Sum costs from evaluations
pe eval config.yaml | pe reduce --sum cost

# Aggregate streamed results
pe eval config.yaml | pe stream | pe reduce --sum cost
```

---

## work

Workspace support for developing multiple related prompts.

### Synopsis

```bash
pe work [command]
```

### Description

A `pe.work` file in the root of your workspace lets you develop multiple prompt modules together, similar to `go.work` files in Go.

### Subcommands

#### work init

Initialize a workspace.

```bash
pe work init
```

#### work list

List workspace contents.

```bash
pe work list
```

#### work sync

Sync workspace prompt dependencies.

```bash
pe work sync
```

#### work use

Add directories to workspace.

```bash
pe work use [directory...]
```

#### work edit

Edit pe.work file programmatically.

```bash
pe work edit [flags]
```

### Examples

```bash
# Initialize workspace
pe work init

# Add directories to workspace
pe work use ./prompts ./templates ./experiments

# List workspace contents
pe work list

# Sync dependencies
pe work sync

# Edit workspace programmatically
pe work edit --add ./new-prompts
```

---

## help

Get help for any PE command.

### Synopsis

```bash
pe help [command] [flags]
```

### Description

Displays help information for any command in the application. Simply type `pe help [path to command]` for full details.

### Examples

```bash
# General help
pe help

# Help for specific command
pe help run

# Help for subcommand
pe help mod list
pe help experimental optimize
```

---

## Pipeline Examples

PE is designed for Unix-style composition:

```bash
# Basic pipeline
pe eval config.yaml | pe filter --contains pass | pe stats

# Complex analysis
pe eval config.yaml | \
  pe stream | \
  pe filter --contains pass | \
  pe analyze --metrics readability

# Cost monitoring
pe eval config.yaml | \
  pe filter --contains cost | \
  pe stream | \
  pe analyze

# A/B testing
pe eval variant-a.yaml -o a.json
pe eval variant-b.yaml -o b.json
pe diff --fail-on-regression --max-score-drop 0.05 a.json b.json
```

For more examples and advanced usage, see the [main documentation](README.md).
