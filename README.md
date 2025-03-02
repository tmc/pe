# pe

pe is a comprehensive toolkit for prompt engineering tasks, supporting evaluation, validation, formatting and viewing of prompt configurations and results.

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

## Specifications

* [prompt_engineering.proto](./prompt_engineering.proto): Protobuf definitions for prompt
  engineering concepts.

## Directory Structure

* `cmd/pe/`: Main command-line tool implementation
* `promptfoo/`: Types and utilities for promptfoo integration
* `example/`: Example configuration files and evaluator scripts

## Installation

The easiest way to install is to run:

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

You can also clone the repository and build it locally:

```bash
git clone https://github.com/tmc/pe.git
cd pe
go build -o pe ./cmd/pe
```

## Testing

```bash
# Run Go tests
go test ./...
```

## Dependencies

* Go 1.18 or higher
* Node.js and npm (for the promptfoo UI integration)

## External Evaluators

pe supports external evaluator executables for running evaluations against real LLM providers:

1. Provider-specific evaluators: `pe-eval-provider-{provider}` (e.g., `pe-eval-provider-openai`)
2. Generic evaluator: `pe-eval`
3. Default CGPT-based implementation: `pe-eval-provider-cgpt`

When running `pe eval`, the tool will search your PATH for these executables in order, using the first one found. If none are found, it falls back to the mock implementation.

The system follows a discovery model similar to protocol buffers' protoc compiler, where provider implementations can be added independently by placing appropriate executables in your PATH.

Example evaluator scripts are provided in the `example/` directory.

## Report Issues / Send Patches

https://github.com/tmc/pe/issues
