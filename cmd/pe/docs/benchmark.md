# Benchmark Command

The `benchmark` command allows you to compare the performance of multiple prompts and LLM providers. It runs each prompt against each provider a specified number of times and collects metrics on response time, token usage, and costs.

## Usage

```
pe benchmark [config_file] [flags]
```

## Flags

```
  -c, --config string         Path to configuration file
  -n, --concurrency int       Number of concurrent benchmark runs (default 1)
  -f, --format string         Output format: 'json', 'yaml', 'csv', or 'text' (default "json")
  -h, --help                  help for benchmark
  -i, --iterations int        Number of times to run each prompt (default 3)
  -o, --output string         Write results to file
```

## Configuration File Format

The benchmark command uses the same configuration file format as the `eval` command, with sections for prompts, providers, and optional tests:

```yaml
prompts:
  - "Your first prompt here"
  - "Your second prompt here"

providers:
  - "googleai:gemini-2.0-flash"
  - "anthropic:claude-3-haiku"

tests:
  - vars:
      some_variable: "value"
```

## Example

```bash
# Run a benchmark with default settings
pe benchmark benchmark-config.yaml

# Run a benchmark with 5 iterations per prompt/provider and 2 concurrent runs
pe benchmark benchmark-config.yaml --iterations 5 --concurrency 2

# Output results in text format
pe benchmark benchmark-config.yaml --format text

# Save results to a file
pe benchmark benchmark-config.yaml --output benchmark-results.json

# Use CSV format for easy import into spreadsheets
pe benchmark benchmark-config.yaml --format csv --output benchmark-results.csv
```

## Output

The benchmark command can output results in several formats:

### JSON

Detailed structured data including per-run metrics and summary statistics.

### YAML 

Same information as JSON but in YAML format.

### CSV

Summary statistics in CSV format, useful for importing into spreadsheets.

### Text

Human-readable summary table showing the key metrics for each prompt and provider.

## Metrics Collected

For each prompt/provider combination, the following metrics are collected:

- **Response Time**: Average, min, max, and percentiles (50th, 90th, 95th, 99th)
- **Token Usage**: Total tokens, input tokens, output tokens
- **Cost**: Estimated cost based on token usage

## Performance Considerations

- Be careful with high concurrency values as they might trigger rate limits from LLM providers
- For reliable benchmarks, use at least 5-10 iterations per prompt/provider
- Token usage and cost estimates are approximate and may vary from actual provider billing