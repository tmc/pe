# Promptfoo Integration Guide

PE provides comprehensive bidirectional compatibility with [promptfoo](https://github.com/promptfoo/promptfoo), enabling seamless migration and interoperability between the two prompt engineering toolchains.

## Overview

The promptfoo integration allows you to:

- **Import** existing promptfoo configurations and test results
- **Export** PE configurations and test results to promptfoo format
- **Convert** between JSON and YAML formats
- **Preserve** assertions, providers, and test structures
- **Round-trip** configurations without loss of fidelity

## Installation

The promptfoo integration is available as:

1. **Built-in Command**: `pe promptfoo` (included in PE core)
2. **Plugin**: For advanced customization (optional)

```bash
# Using built-in command (no installation needed)
pe promptfoo --help

# Optional: Install plugin for extended features
pe plugin install promptfoo
```

## Command Reference

### Import from Promptfoo

Import existing promptfoo configurations or test results:

```bash
# Import a promptfoo configuration
pe promptfoo import promptfooconfig.yaml -o pe-config.yaml

# Import with provider preservation
pe promptfoo import config.yaml -o pe-config.yaml --preserve-providers

# Import test results
pe promptfoo import promptfoo-output.json -o pe-results.json

# Import with options
pe promptfoo import config.yaml -o pe-config.yaml \
  --preserve-providers \    # Keep original provider configs
  --expand-includes \       # Expand file:// references
  --convert-assertions      # Convert assertion formats
```

### Export to Promptfoo

Export PE data to promptfoo-compatible formats:

```bash
# Export PE config to promptfoo YAML
pe promptfoo export -c pe-config.yaml -o promptfooconfig.yaml

# Export to JSON format
pe promptfoo export -c pe-config.yaml -o promptfoo.json -f json

# Export test results
pe promptfoo export -r pe-results.json -o promptfoo-output.json

# Export with specific format
pe promptfoo export -r results.json -o output.json --format json
```

### Convert Formats

Convert between JSON and YAML:

```bash
# Convert YAML to JSON
pe promptfoo convert config.yaml -o config.json

# Convert JSON to YAML
pe promptfoo convert config.json -o config.yaml

# Auto-detect output format from extension
pe promptfoo convert input.json -o output.yaml
```

## Integration with PE Workflow

### Direct Export from Test Command

Export test results directly in promptfoo format:

```bash
# Run tests and export to promptfoo format
pe test config.yaml --export-format promptfoo -o results.json

# Run with multiple providers and export
pe test config.yaml --providers openai,anthropic \
  --export-format promptfoo -o comparison.json
```

### Pipeline Integration

Use in PE pipelines:

```bash
# Import, optimize, and export back
pe promptfoo import original.yaml -o temp.yaml | \
pe optimize temp.yaml --method textgrad | \
pe promptfoo export -c temp_optimized.yaml -o improved.yaml
```

## Format Mapping

### Configuration Mapping

| Promptfoo | PE | Notes |
|-----------|-----|-------|
| `prompts` | `prompts` | Direct mapping with template support |
| `providers` | `providers` | Optional preservation |
| `tests.vars` | `tests.variables` | Variable expansion |
| `tests.assert` | `tests.assertions` | Type conversion |
| `defaultTest` | `default_test` | Global defaults |

### Assertion Type Mapping

| Promptfoo Type | PE Type | Description |
|----------------|---------|-------------|
| `contains` | `contains` | Check substring |
| `not-contains` | `not_contains` | Negative contains |
| `equals` | `equals` | Exact match |
| `regex` | `regex` | Pattern match |
| `javascript` | `custom` | Custom JS evaluation |
| `llm-rubric` | `llm_rubric` | LLM-based evaluation |

### Provider Format Support

```yaml
# Promptfoo simple format
providers:
  - openai:gpt-4
  - anthropic:claude-3

# Promptfoo complex format  
providers:
  - id: gpt4-prod
    apiProvider: openai
    apiKey: ${OPENAI_API_KEY}
    model: gpt-4
    temperature: 0.7

# PE format (after import)
providers:
  - type: openai
    model: gpt-4
  - type: anthropic
    model: claude-3
  - id: gpt4-prod
    type: openai
    model: gpt-4
    config:
      temperature: 0.7
```

## Examples

### Basic Import/Export

```bash
# 1. Import promptfoo config
pe promptfoo import promptfoo-config.yaml -o my-eval.yaml

# 2. Run evaluation with PE
pe test my-eval.yaml --output results.json

# 3. Export results back to promptfoo format
pe promptfoo export -r results.json -o promptfoo-results.json
```

### Migration Workflow

```bash
# Migrate from promptfoo to PE
pe promptfoo import old-config.yaml -o new-config.yaml \
  --preserve-providers \
  --convert-assertions

# Use PE's advanced features
pe optimize new-config.yaml --method semantic-backprop
pe experimental compose new-config.yaml components/ --style cot

# Export back if needed
pe promptfoo export -c new-config_optimized.yaml -o updated.yaml
```

### Round-Trip Testing

```bash
# Verify round-trip compatibility
pe promptfoo import original.yaml -o intermediate.yaml
pe promptfoo export -c intermediate.yaml -o roundtrip.yaml

# Compare files (should be semantically equivalent)
diff original.yaml roundtrip.yaml
```

## Advanced Features

### Custom Assertion Conversion

PE automatically converts complex assertions:

```yaml
# Promptfoo assertion
assert:
  - type: javascript
    value: |
      output.length > 100 && 
      output.includes('example')

# PE equivalent (after import)
assertions:
  - type: custom
    value: |
      output.length > 100 && 
      output.includes('example')
```

### File Reference Expansion

When importing with `--expand-includes`:

```yaml
# Promptfoo with file reference
prompts:
  - file://prompts/assistant.txt

# PE after import (expanded)
prompts:
  - content: "You are a helpful assistant..."
    source: "prompts/assistant.txt"
```

### Provider Configuration Preservation

With `--preserve-providers`, complex provider configs are maintained:

```yaml
# Original promptfoo provider
providers:
  - id: azure-gpt4
    apiProvider: azureopenai
    apiHost: https://myinstance.openai.azure.com
    apiKey: ${AZURE_KEY}
    deploymentName: gpt-4

# PE after import (preserved)
providers:
  - id: azure-gpt4
    type: azureopenai
    config:
      apiHost: https://myinstance.openai.azure.com
      apiKey: ${AZURE_KEY}
      deploymentName: gpt-4
```

## Plugin API

For advanced customization, use the plugin API:

```go
// Custom import hook
type ImportHook func(config *promptfoo.Config) error

// Custom export transformer  
type ExportTransformer func(results *pe.Results) error

// Register custom converters
promptfoo.RegisterConverter("myformat", MyConverter{})
```

## Limitations

- **Plugins**: Promptfoo JavaScript plugins need manual conversion
- **Transforms**: Complex transforms require custom handling
- **Scenarios**: Advanced scenario features may need adaptation
- **Grading**: Custom grading functions need rewriting in PE format

## Best Practices

1. **Always verify** imported configurations before running expensive evaluations
2. **Use round-trip testing** to ensure compatibility when migrating
3. **Preserve providers** when working with complex authentication setups
4. **Export regularly** when collaborating with promptfoo users
5. **Version control** both PE and promptfoo configs during migration

## Troubleshooting

### Import Issues

```bash
# Debug import with verbose output
PE_DEBUG=1 pe promptfoo import config.yaml -o output.yaml

# Validate imported config
pe validate output.yaml
```

### Export Validation

```bash
# Verify exported format
pe promptfoo export -c config.yaml -o test.yaml
npx promptfoo@latest eval -c test.yaml --dry-run
```

### Format Detection

```bash
# Force format if auto-detection fails
pe promptfoo convert ambiguous.file -o output.json --format json
```

## See Also

- [Promptfoo Documentation](https://www.promptfoo.dev/)
- [PE Test Command](./CLI_REFERENCE.md#test)
- [PE Configuration Format](./API_REFERENCE.md#configuration)
- [Migration Guide](./MIGRATION_FROM_PROMPTFOO.md)