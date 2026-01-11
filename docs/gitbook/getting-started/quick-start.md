# Quick Start

In this guide, you will create, run, and evaluate your first prompt with PE.

## 1. Run a Simple Prompt

The `pe run` command is the equivalent of `go run`. It executes a prompt immediately.

```bash
pe run "What is the capital of France?" --provider openai
```

## 2. Use Prompt Files & Templates

PE encourages storing prompts in files. It supports Go text templates for dynamic variables.

Create a file named `translate.prompt`:

```text
Translate the following text to {{.language}}:

Text: {{.text}}
```

Run it with variables:

```bash
pe run translate.prompt \
  --var language="Spanish" \
  --var text="Hello, world!"
```

## 3. Initialize a Module

Organize your prompts into a module (project):

```bash
mkdir my-prompts
cd my-prompts
pe mod init github.com/username/my-prompts
```

This creates a `pe.mod` file, similar to `go.mod`, to track dependencies.

## 4. Run an Evaluation

Create a configuration file `pe-config.yaml` to define test cases:

```yaml
prompts: [translate.prompt]
providers: [openai:gpt-4]
tests:
  - vars:
      language: French
      text: "Good morning"
    assert:
      - type: contains
        value: "Bonjour"
```

Run the evaluation:

```bash
pe eval pe-config.yaml
```

You will see a report of the test results in your terminal.
