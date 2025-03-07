# Workbench (wb) Tool Guide

The Workbench (`wb`) tool is a companion to the main `pe` toolkit, focused on the interactive process of creating and refining prompts.

## Overview

While the main `pe` tool focuses on evaluating prompts against different providers and verifying assertions, the `wb` tool is designed for the creative process of prompt engineering:

- Creating initial prompt templates
- Iterating on prompts
- Managing prompt collections
- Interactive experimentation

## Installation

```bash
go install github.com/tmc/pe/cmd/wb@latest
```

## Basic Usage

```bash
# Create a new prompt file
wb create my-prompt.json

# Edit an existing prompt
wb edit my-prompt.json

# List all prompts in a directory
wb list ./prompts/

# Test a prompt with variables
wb test my-prompt.json --var "question=What is the capital of France?"
```

## Prompt File Format

Prompts are stored in JSON format with this structure:

```json
{
  "name": "Example Prompt",
  "description": "This is an example prompt for demonstration",
  "template": "Please answer the following question: {{.question}}",
  "variables": {
    "question": {
      "description": "The question to ask",
      "type": "string",
      "required": true,
      "examples": [
        "What is the capital of France?",
        "How does photosynthesis work?"
      ]
    }
  },
  "metadata": {
    "author": "Your Name",
    "version": "1.0",
    "tags": ["example", "documentation"]
  }
}
```

The key fields are:

- **name**: A descriptive name for the prompt
- **description**: Purpose and usage notes
- **template**: The prompt template with variable placeholders
- **variables**: Schema for variables used in the template
- **metadata**: Additional information about the prompt

## Commands

### Create a Prompt

```bash
wb create my-prompt.json
```

This opens your default editor to create a new prompt file with a basic template.

### Edit a Prompt

```bash
wb edit my-prompt.json
```

Opens the prompt file in your editor.

### Test a Prompt

```bash
wb test my-prompt.json --var "question=What is the capital of France?"
```

Renders the prompt with the provided variables.

### List Prompts

```bash
wb list ./prompts/
```

Lists all prompt files in a directory with their names and descriptions.

## Advanced Features

### Template Variables

The workbench supports variable substitution in templates using Go's text/template syntax:

```
{{.variable_name}}
```

### Conditional Sections

You can include conditional sections in your templates:

```
{{if .debug}}
Include debugging information...
{{else}}
Normal output...
{{end}}
```

### Including Files

Templates can include other files:

```
{{include "header.txt"}}
Main content...
{{include "footer.txt"}}
```

## Integration with pe

The workbench is designed to work seamlessly with the main `pe` toolkit:

1. Create and refine prompts with `wb`
2. Test them across providers and validate outputs with `pe`

Example workflow:

```bash
# Create a prompt
wb create my-prompt.json

# Test it with the workbench
wb test my-prompt.json --var "question=What is the capital of France?"

# Create a pe config using this prompt
pe convert my-prompt.json config.yaml

# Evaluate with pe
pe eval config.yaml
```

## Example Prompts

The workbench includes example prompts in `cmd/wb/example/` that demonstrate various features and best practices.

## Tips for Effective Prompt Engineering

1. **Start simple**: Begin with a minimal prompt and add complexity as needed
2. **Use variables**: Make prompts flexible with variables for different use cases
3. **Include examples**: Provide examples to guide the model's output
4. **Test variations**: Try different phrasings and structures to find what works best
5. **Document your prompts**: Add clear descriptions and metadata for future reference
