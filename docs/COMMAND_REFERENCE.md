# PE Command Reference

Complete reference for all PE commands and options.

## Core Commands

### `pe run`
Execute a prompt with an LLM provider.

```bash
pe run [prompt or file] [flags]
```

**Flags:**
- `--provider string` - Provider to use (openai, anthropic, cgpt)
- `--model string` - Specific model (e.g., gpt-4, claude-3)
- `--var key=value` - Template variables (repeatable)
- `--system string` - System prompt
- `--temperature float` - Temperature (0.0-1.0, default 0.7)
- `--max-tokens int` - Maximum response tokens
- `--stream` - Enable streaming output
- `--json` - Output in JSON format

**Examples:**
```bash
pe run "What is 2+2?"
pe run prompt.txt --provider openai
pe run template.prompt --var name="Alice" --var age="30"
```

### `pe eval`
Run evaluation against test cases.

```bash
pe eval [config.yaml] [flags]
```

**Flags:**
- `--provider string` - Override provider
- `--save-db` - Save results to database
- `--output string` - Output file (json/yaml/csv)
- `--max-cost float` - Maximum total cost
- `--concurrency int` - Parallel evaluations
- `--verbose` - Detailed output

**Examples:**
```bash
pe eval config.yaml
pe eval config.yaml --save-db --output results.json
```

### `pe optimize`
Optimize prompts using various methods.

```bash
pe optimize [prompt] [flags]
```

**Flags:**
- `--method string` - Optimization method (pe2, semantic, evolve)
- `--iterations int` - Number of iterations
- `--output string` - Save optimized prompt
- `--verbose` - Show optimization progress

**Methods:**
- `pe2` - Prompt Engineer 2 (iterative refinement)
- `semantic` - Semantic gradient descent
- `evolve` - Evolutionary optimization
- `textgrad` - TextGrad backpropagation

**Examples:**
```bash
pe optimize prompt.txt --method pe2 --iterations 3
pe optimize "Explain AI" --method semantic
```

## Evaluation Commands

### `pe view`
View evaluation results in browser.

```bash
pe view [flags]
```

**Flags:**
- `--port int` - Server port (default 3000)
- `--db string` - Database path

### `pe stats`
Show statistics from evaluation.

```bash
pe stats [results.json] [flags]
```

**Flags:**
- `--format string` - Output format (table/json/csv)
- `--group-by string` - Group by field
- `--metric string` - Focus on specific metric

### `pe diff`
Compare evaluation results.

```bash
pe diff [baseline.json] [current.json] [flags]
```

**Flags:**
- `--threshold float` - Significance threshold
- `--metric string` - Comparison metric

## Pipeline Commands

### `pe filter`
Filter evaluation results.

```bash
pe filter [flags]
```

**Flags:**
- `--success` - Only successful results
- `--failed` - Only failed results
- `--provider string` - Filter by provider
- `--min-score float` - Minimum score
- `--max-cost float` - Maximum cost
- `--min-latency int` - Minimum latency (ms)
- `--max-latency int` - Maximum latency (ms)

### `pe stream`
Stream process results.

```bash
pe stream [flags]
```

**Flags:**
- `--select string` - Select fields (comma-separated)
- `--format string` - Output format
- `--buffer-size int` - Buffer size

### `pe analyze`
Analyze results with statistics.

```bash
pe analyze [flags]
```

**Flags:**
- `--metric string` - Metric to analyze
- `--percentiles string` - Percentiles (e.g., "50,90,95,99")
- `--group-by string` - Group results

## Module Commands

### `pe mod init`
Initialize a new module.

```bash
pe mod init [module-path]
```

### `pe mod tidy`
Clean up dependencies.

```bash
pe mod tidy
```

### `pe mod vendor`
Vendor dependencies locally.

```bash
pe mod vendor
```

## Utility Commands

### `pe cat`
Display prompt with variable substitution.

```bash
pe cat [prompt-file] [flags]
```

**Flags:**
- `--var key=value` - Variables
- `--list-vars` - List required variables
- `--validate` - Validate syntax only

### `pe fmt`
Format configuration files.

```bash
pe fmt [config.yaml]
```

### `pe vet`
Validate configuration.

```bash
pe vet [config.yaml]
```

### `pe init`
Initialize configuration template.

```bash
pe init [config.yaml]
```

### `pe version`
Show version information.

```bash
pe version [flags]
```

**Flags:**
- `--verbose` - Detailed version info

## Advanced Commands

### `pe benchmark`
Run performance benchmarks.

```bash
pe benchmark [config.yaml] [flags]
```

**Flags:**
- `--iterations int` - Number of iterations
- `--concurrency int` - Parallel requests
- `--go-bench` - Output in Go benchmark format

### `pe test`
Run advanced testing.

```bash
pe test [config.yaml] [flags]
```

**Flags:**
- `--type string` - Test type (regression/property/fuzz)
- `--coverage` - Generate coverage report

### `pe security`
Security testing.

```bash
pe security [subcommand] [flags]
```

**Subcommands:**
- `test` - Run security tests
- `scan` - Scan for vulnerabilities
- `redteam` - Red team simulation

### `pe attest`
Cryptographic attestation.

```bash
pe attest [subcommand] [flags]
```

**Subcommands:**
- `sign` - Sign results
- `verify` - Verify signature
- `generate-key` - Generate signing key

## Environment Variables

- `OPENAI_API_KEY` - OpenAI API key
- `ANTHROPIC_API_KEY` - Anthropic API key
- `PE_DEFAULT_PROVIDER` - Default provider
- `PE_DEFAULT_MODEL` - Default model
- `PE_TEST_MODE` - Enable test mode
- `PE_DEBUG` - Enable debug output
- `PE_CONFIG_PATH` - Config file path
- `PE_CACHE_DIR` - Cache directory

## Configuration Files

### Basic Structure
```yaml
description: "Configuration description"

prompts:
  - "Prompt template with {{.variable}}"
  - file://path/to/prompt.txt

providers:
  - name: openai
    config:
      model: gpt-4o-mini
      temperature: 0.7

tests:
  - vars:
      variable: "value"
    assert:
      - type: contains
        value: "expected"
```

### Assertion Types

- `contains` - Text contains value
- `not-contains` - Text doesn't contain value
- `regex` - Matches regex pattern
- `length` - Length constraints
- `latency` - Response time limit
- `cost` - Cost limit
- `is-json` - Valid JSON
- `json-schema` - JSON schema validation
- `similarity` - Semantic similarity
- `llm-eval` - LLM as judge

## Output Formats

- `json` - JSON format
- `yaml` - YAML format
- `csv` - CSV format
- `table` - ASCII table
- `jsonl` - JSON lines

## Examples

### Complex Pipeline
```bash
pe eval config.yaml | \
  pe filter --success --max-cost 0.01 | \
  pe analyze --metric latency --percentiles "50,90,95,99" | \
  pe convert --format csv > results.csv
```

### Batch Processing
```bash
for file in prompts/*.txt; do
  pe run "$file" --provider openai --json
done | jq -s '.' > all-results.json
```

### CI/CD Integration
```bash
# In GitHub Actions
pe eval config.yaml --save-db
pe diff baseline.json current.json --threshold 0.05
if [ $? -ne 0 ]; then
  echo "Regression detected!"
  exit 1
fi
```

## Tips

1. Use `--help` on any command for details
2. Pipe commands for powerful workflows
3. Save intermediate results with `tee`
4. Use environment variables for API keys
5. Template syntax uses Go templates: `{{.var}}`
6. JSON output integrates well with `jq`
7. Use `--verbose` for debugging