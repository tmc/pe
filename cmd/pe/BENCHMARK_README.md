# Benchmark Feature for PE Command

This document explains the new benchmark feature added to the `pe` command, which allows comparing multiple prompts and providers for performance metrics.

## Implementation Overview

The benchmark feature adds a new subcommand to the `pe` CLI tool that:

1. Accepts a configuration file specifying prompts and providers
2. Runs each prompt against each provider multiple times
3. Measures and reports response time, token usage, and costs
4. Outputs results in various formats (text, JSON, CSV, YAML)

## Key Files

The implementation consists of the following key files:

- `benchmark.go`: The main implementation of the benchmark command
- `example/benchmark-config.yaml`: A sample configuration file for benchmarking
- `docs/benchmark.md`: Documentation for the benchmark command
- `run_benchmark.sh`: A demo script to run the benchmark command

## Usage

```bash
pe benchmark [config_file] [flags]
```

### Flags

```
  -c, --config string         Path to configuration file
  -n, --concurrency int       Number of concurrent benchmark runs (default 1)
  -f, --format string         Output format: 'json', 'yaml', 'csv', or 'text' (default "json")
  -h, --help                  Help for benchmark
  -i, --iterations int        Number of times to run each prompt (default 3)
  -o, --output string         Write results to file
```

## Example Configuration File

```yaml
prompts:
  - "Explain the concept of prompt engineering in one paragraph."
  - "What are 3 best practices for writing effective prompts for large language models?"
  - "Compare and contrast zero-shot, one-shot, and few-shot prompting approaches."

providers:
  - "googleai:gemini-2.0-flash"
  - "googleai:gemini-2.0-pro"
  - "anthropic:claude-3-haiku"
  - "anthropic:claude-3-sonnet"

tests:
  - vars:
      topic: "prompt engineering"
```

## Metrics Collected

For each prompt and provider combination, the following metrics are measured:

- **Response Time**: Average, minimum, maximum, and percentiles (50th, 90th, 95th, 99th)
- **Token Usage**: Total tokens, input tokens, output tokens
- **Cost**: Estimated cost based on token usage

## Output Formats

The benchmark results can be output in several formats:

- **JSON**: Detailed output with all metrics and individual run data
- **YAML**: Same information as JSON but in YAML format
- **CSV**: Summary statistics in CSV format for easy import into spreadsheets
- **Text**: Human-readable summary tables organized by prompt

## Performance and Concurrency

The benchmark command supports concurrent execution of benchmark runs through the `--concurrency` flag, which controls how many prompts can be evaluated simultaneously. This can significantly speed up the benchmarking process, but may trigger rate limits if set too high.

## Usage Examples

```bash
# Run a simple benchmark
pe benchmark benchmark-config.yaml

# Run with 5 iterations per prompt/provider
pe benchmark benchmark-config.yaml --iterations 5

# Run with 3 concurrent evaluations
pe benchmark benchmark-config.yaml --concurrency 3

# Save results as CSV
pe benchmark benchmark-config.yaml --format csv --output results.csv

# Human-readable output
pe benchmark benchmark-config.yaml --format text
```