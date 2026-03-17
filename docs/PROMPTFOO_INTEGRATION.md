# Promptfoo Integration Guide

PE exposes Promptfoo compatibility through the `pe-promptfoo` plugin. Once the plugin binary is on your `PATH`, PE discovers it automatically and makes it available as `pe promptfoo`.

## Installation

Build the plugin from this repository and place it on your `PATH`:

```bash
go build -o ~/bin/pe-promptfoo ./plugins/promptfoo
pe plugin list
pe promptfoo --help
```

## Supported Scope

The current plugin supports configuration-file workflows for the common evaluation schema:

- `description`
- `prompts`
- `providers`
- `tests`
- `defaultTest`

When exporting PE configs, the plugin normalizes common PE key variants:

- `variables` -> `vars`
- `assertions` -> `assert`
- assertion types like `not_contains` -> `not-contains`

When importing Promptfoo configs, provider objects are reshaped into PE-friendly entries using:

- `apiProvider` + `model` -> `id` (for example `openai:gpt-4o-mini`)
- remaining provider fields -> `config`

## Commands

Import a Promptfoo-compatible config into PE's compatibility shape:

```bash
pe promptfoo import promptfooconfig.yaml -o pe-config.yaml
pe promptfoo import promptfooconfig.yaml -o pe-config.json -f json
```

Export a PE config to Promptfoo-friendly keys:

```bash
pe promptfoo export pe-config.yaml -o promptfooconfig.yaml
pe promptfoo export pe-config.yaml -o promptfooconfig.json -f json
```

Convert in either direction:

```bash
pe promptfoo convert promptfooconfig.yaml pe-config.yaml
pe promptfoo convert pe-config.yaml promptfooconfig.json --direction pe-to-promptfoo
```

Formats default from file extensions. You can override them explicitly:

```bash
pe promptfoo convert input.yaml output.json --direction pe-to-promptfoo --output-format json
pe promptfoo convert input.json output.yaml --input-format json --output-format yaml
```

## Example

Promptfoo-style input:

```yaml
prompts:
  - "Answer {{question}}"
providers:
  - id: gpt4-mini
    apiProvider: openai
    model: gpt-4o-mini
tests:
  - vars:
      question: "What is 2+2?"
    assert:
      - type: llm-rubric
        value: "Says 4"
```

Imported PE-compatible output:

```yaml
version: "1.0"
prompts:
  - "Answer {{question}}"
providers:
  - id: openai:gpt-4o-mini
tests:
  - vars:
      question: "What is 2+2?"
    assert:
      - type: llm-rubric
        value: "Says 4"
```

## Current Limits

The plugin does not currently implement full Promptfoo results import/export, file include expansion, or deep provider-schema translation beyond the common config shapes above.
