# pe: Prompt Engineering Toolkit

`pe` is a comprehensive toolkit for prompt engineering tasks, supporting evaluation, validation, formatting and viewing of prompt configurations and results across multiple LLM providers.

## Features

- **Template-based prompts:** Use variables to test multiple variations
- **Multiple provider support:** Test across OpenAI, Anthropic, Google AI and more  
- **Assertion-based testing:** Verify outputs meet expected criteria
- **Dry run mode:** Preview commands without making API calls
- **Extensible architecture:** Easily add new providers and assertion types

## Installation

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

## Quick Start

Create a YAML configuration file:

```yaml
# example.yaml
prompts:
  - "What is the capital of {{country}}?"

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

Run an evaluation:

```bash
pe eval example.yaml -o results.json
```

Generate shell commands to run manually:

```bash
pe eval example.yaml --dry-run > commands.sh
chmod +x commands.sh
./commands.sh  # Run the commands if desired
```

## Commands

* **eval**: Evaluate prompt configurations against LLM providers
  * `pe eval test-config.yaml`
  * `pe eval -c test-config.yaml -o results.json`

* **view**: View evaluation results in the promptfoo browser-based UI
  * `pe view`
  * `pe view [evalId]`
  * `pe view -f results.json`

* **vet**: Validate promptfoo configuration files
  * `pe vet config.yaml`
  * `cat config.yaml | pe vet`

* **fmt**: Format promptfoo configuration files
  * `pe fmt config.yaml --output yaml`
  * `pe fmt --write config.yaml`

* **convert**: Convert between YAML and JSON formats
  * `pe convert input.yaml output.json`
  * `pe convert config.json config.yaml --output yaml`

## External Evaluators

`pe` supports external evaluator executables for running evaluations against real LLM providers:

1. Provider-specific evaluators: `pe-eval-provider-{provider}` (e.g., `pe-eval-provider-openai`)
2. Generic evaluator: `pe-eval`
3. Default CGPT-based implementation: `pe-eval-provider-cgpt`

When running `pe eval`, the tool will search your PATH for these executables in order, using the first one found. If none are found, it falls back to the built-in implementation.

## Directory Structure

* `cmd/pe/`: Main command-line tool implementation
* `internal/promptfoo/`: Types and utilities for promptfoo integration
* `internal/cgpt/`: CGPT integration for prompt evaluation
* `specs/`: Specifications and protocol definitions
* `example/`: Example configuration files and evaluator scripts
* `docs/`: Documentation

## Testing

```bash
# Run Go tests
go test ./...
```

## Report Issues / Send Patches

https://github.com/tmc/pe/issues