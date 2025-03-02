# PE Workbench (wb)

The PE Workbench (`wb`) is a command-line tool for working with prompt engineering protobuf files, based on the schema defined in [prompt_engineering.proto](../../specs/prompt_engineering.proto).

## Overview

The `wb` tool provides functionality for:

- Creating, viewing, and analyzing prompts
- Validating prompts against schema and semantic rules
- Converting between different formats (protobuf, JSON, etc.)
- Testing and evaluating prompts against test cases
- Comparing and merging different prompts or revisions
- Searching for prompts matching specific criteria

## Example Usage

Create a new prompt:
```bash
wb create --name "My Prompt" --system-prompt "You are a helpful assistant." --output my_prompt.pb
```

Show prompt details:
```bash
wb show my_prompt.pb
```

Analyze a prompt:
```bash
wb analyze my_prompt.pb
```

For a full list of commands and options, see the [workbench documentation](../../docs/workbench.md).

## Example Files

The `example` directory contains sample prompt files that demonstrate the schema:

- `example_prompt.json`: A capital city prompt with variables, examples, and test cases
