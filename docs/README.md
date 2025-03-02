# PE: Prompt Engineering Toolkit

PE is a command-line utility for testing, evaluating, and optimizing prompts for large language models (LLMs) like GPT-4, Claude, and Gemini.

## Features

- **Template-based prompts:** Use variables in your prompts to test multiple variations
- **Multiple provider support:** Test across different LLM providers (OpenAI, Anthropic, Google AI)  
- **Assertion-based testing:** Verify that model outputs meet expected criteria
- **Dry run mode:** Preview commands without making API calls
- **Extensible architecture:** Easy to add new providers and assertion types

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
  - "Tell me about the capital city of {{country}}."

providers:
  - "openai:gpt-4"
  - "anthropic:claude-3-haiku"
  - "googleai:gemini-2.0-flash"

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

See the commands that would be executed without running them:

```bash
pe eval example.yaml --dry-run > commands.sh
chmod +x commands.sh
./commands.sh  # Run the commands if desired
```

## Configuration Format

The PE configuration uses a YAML format with these key sections:

- `prompts`: List of prompt templates
- `providers`: List of LLM providers to test with
- `tests`: Test cases with variables and assertions

For complete documentation on the configuration format, see [Configuration Guide](docs/configuration.md).

## Assertion Types

PE supports various assertion types to validate model outputs:

- `contains`: Output includes the specified text
- `not-contains`: Output does not include the specified text
- `equals`: Output exactly matches the specified text
- `regex`: Output matches the regular expression
- `starts-with`: Output begins with the specified text
- `ends-with`: Output ends with the specified text

## Command Reference

- `pe eval [config]`: Evaluate prompts according to the config file
  - `--output, -o`: Write results to a file
  - `--format, -f`: Output format (json, yaml, text)
  - `--dry-run`: Show commands without executing them
  - `--timeout, -t`: Set timeout for the evaluation

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT