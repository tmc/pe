# PE Workbench (wb)

The PE Workbench (`wb`) is a command-line tool for working with prompt engineering protobuf files. It provides utilities for creating, viewing, analyzing, and manipulating prompts defined using the prompt_engineering.proto schema.

## Installation

```bash
go install github.com/tmc/pe/cmd/wb@latest
```

## Basic Usage

```bash
wb [command] [flags]
```

## Commands

### Create

Create a new prompt or prompt revision:

```bash
wb create --name "My Prompt" --system-prompt "You are a helpful assistant." --model gpt-4 --output my_prompt.pb
```

### Show

Show prompt details:

```bash
wb show my_prompt.pb
wb show my_prompt.pb --format json
wb show my_prompt.pb --revision rev-123
```

### Analyze

Analyze a prompt to identify potential issues, calculate statistics, etc.:

```bash
wb analyze my_prompt.pb
wb analyze my_prompt.pb --revision rev-123
```

### Validate

Validate a prompt against schema and semantic rules:

```bash
wb validate my_prompt.pb
```

### Convert

Convert a prompt between different formats:

```bash
wb convert my_prompt.pb my_prompt.json
wb convert my_prompt.json my_prompt.pb
```

### Evaluate

Evaluate a prompt against its test cases using an LLM provider:

```bash
wb eval my_prompt.pb
wb eval my_prompt.pb --provider openai
```

### Format

Format a prompt file according to style guidelines:

```bash
wb format my_prompt.pb
```

### Merge

Merge multiple prompts or prompt revisions into a single prompt:

```bash
wb merge prompt1.pb prompt2.pb merged_prompt.pb
```

### Diff

Show differences between two prompts or prompt revisions:

```bash
wb diff prompt1.pb prompt2.pb
```

### Export

Export a prompt to a different format:

```bash
wb export my_prompt.pb my_prompt.json
wb export my_prompt.pb my_prompt.txt --format text
```

### Import

Import a prompt from a different format:

```bash
wb import my_prompt.json my_prompt.pb
wb import my_prompt.yaml my_prompt.pb --format yaml
```

### Test

Run test cases for a prompt and report results:

```bash
wb test my_prompt.pb
```

### Stat

Show statistics for a prompt:

```bash
wb stat my_prompt.pb
```

### Search

Search for prompts matching criteria:

```bash
wb search "keyword"
wb search "keyword" --dir /path/to/prompts
```

## File Formats

The workbench supports the following file formats:

- **Proto (.pb)**: Binary protobuf format
- **JSON (.json)**: JSON representation of the protobuf message
- **YAML (.yaml)**: YAML representation of the protobuf message (for some commands)
- **Text (.txt)**: Human-readable text format (for exports only)

## Examples

### Creating and Testing a Prompt

```bash
# Create a new prompt
wb create --name "Capital Quiz" --system-prompt "You are a geography tutor. Answer questions about capital cities." --output capital_quiz.pb

# Show the prompt
wb show capital_quiz.pb

# Run test cases (if defined)
wb test capital_quiz.pb

# Export to JSON for easier editing
wb export capital_quiz.pb capital_quiz.json

# After editing, import back
wb import capital_quiz.json capital_quiz.pb

# Validate the updated prompt
wb validate capital_quiz.pb
```

### Comparing Prompt Revisions

```bash
# Create two different prompts
wb create --name "Prompt v1" --system-prompt "System prompt v1" --output prompt_v1.pb
wb create --name "Prompt v2" --system-prompt "System prompt v2" --output prompt_v2.pb

# Compare them
wb diff prompt_v1.pb prompt_v2.pb

# Merge them into a new prompt
wb merge prompt_v1.pb prompt_v2.pb merged_prompt.pb
```

## Using wb with Large Language Models

The `wb` tool can be combined with other PE tools to test and evaluate prompts:

```bash
# Create a prompt
wb create --name "Test Prompt" --system-prompt "You are a helpful assistant." --output test_prompt.pb

# Export it for the PE evaluator
wb export test_prompt.pb test_prompt.json

# Use pe to evaluate it
pe eval test_prompt.json -o results.json

# Import the results back
wb import results.json evaluated_prompt.pb
```

## Schema Reference

The `wb` tool works with the proto schema defined in `prompt_engineering.proto`. For a complete reference, see the schema documentation or run:

```bash
wb show --schema
```

This will display the full protobuf schema used by the workbench.