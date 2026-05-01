# Local Ollama Example

This example runs a prompt against a model served by a local Ollama daemon.
PE uses the native Ollama provider registered through the provider bridge.

No API key is required. Ollama must be running, and the configured model must
be available locally.

## Files

- `config.yaml` - evaluation config using provider spec `ollama:llama3.2`
- `prompt.txt` - prompt template used by the evaluation

## Setup

Start Ollama if it is not already running:

```bash
ollama serve
```

In another shell, pull the model used by the config:

```bash
ollama pull llama3.2
```

If your daemon listens somewhere other than `http://localhost:11434`, update
`base_url` in `config.yaml`. If you use a different model, update the
`ollama:model` provider spec and pull that model first.

## Run

```bash
cd examples/local-ollama
pe eval config.yaml --verbose
```

To run only the prompt:

```bash
pe run prompt.txt --provider ollama:llama3.2 --var question="What is one benefit of local inference?"
```
