# Running Prompts

The most basic operation in PE is executing a prompt.

## The `pe run` Command

Use `pe run` to send a prompt to an LLM provider.

### Basic Usage

```bash
pe run "Explain quantum physics" --provider openai
```

### Using Files

We recommend storing prompts in files for version control.

```bash
pe run prompts/explain.prompt
```

### Template Variables

Pass variables to your prompt templates using `--var`:

```bash
pe run translate.prompt --var lang=French --var text="Hello"
```

### Streaming

For long responses, use the --stream flag to see output in real-time:

```bash
pe run write-essay.prompt --stream
```

## Interactive Mode (`pe interactive`)

For rapid iteration, use the interactive REPL:

```bash
pe interactive --provider openai
```

This opens a session where you can:
*   Type prompts and get immediate responses.
*   modify settings like temperature.
*   Load/save sessions.
