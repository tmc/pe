# Local Ollama Example

This example runs a prompt against a model served by a local Ollama daemon.
PE uses the native Ollama provider registered through the provider bridge.

No API key is required. Ollama must be running, and the configured model must
be available locally.

## Files

- `config.yaml` - evaluation config using provider spec `ollama:llama3.2:3b`
- `prompt.txt` - prompt template used by the evaluation

## Setup

Start Ollama if it is not already running:

```bash
ollama serve
```

In another shell, pull the model used by the config:

```bash
ollama pull llama3.2:3b
```

If your daemon listens somewhere other than `http://localhost:11434`, update
`base_url` in `config.yaml`. If you use a different model, update the
`ollama:model` provider spec and pull that model first.

## Validation

Before release, run this example against at least one small instruct model and
one second model family already available on the target machine. The goal is to
verify provider wiring and option passthrough, not model quality.

Recommended smoke matrix:

- `llama3.2:3b`: default small local model.
- `qwen2.5:7b-instruct-q4_0` or a local `qwen3.5:*`: alternate tokenizer and
  response style.
- Optional larger local model such as `gemma4:*` when available, to check the
  same config path under longer startup and generation times.

For each model:

1. Confirm it appears in `ollama list`.
2. Set the provider id in `config.yaml` to `ollama:<model>`.
3. Run `pe eval config.yaml`.
4. Run `pe run prompt.txt --provider ollama:<model> --var question="What is one benefit of local inference?"`.
5. Confirm the command returns model text, not an Ollama HTTP error, and that
   `raw`, `seed`, and `num_predict` remain present in the provider config.

## Run

```bash
cd examples/local-ollama
pe eval config.yaml
```

To run only the prompt:

```bash
pe run prompt.txt --provider ollama:llama3.2:3b --var question="What is one benefit of local inference?"
```

## Validation record (2026-07-06)

Validated against a local Ollama daemon, version 0.30.6, on darwin/arm64:

- `llama3.2:3b`: `pe eval config.yaml` passed (1/1, token usage reported:
  48 prompt / 128 completion) and `pe run prompt.txt --provider
  ollama:llama3.2:3b` returned model text.
- `qwen2.5:7b-instruct-q4_0`: `pe eval` (provider id swapped in config) and
  `pe run` both returned model text.
- `gemma4:e2b-it-qat`: `pe run` returned model text through the same config
  path.
- Missing model (`ollama:no-such-model:1b`): fails closed with exit code 1 and
  `ollama API error 404: {"error":"model 'no-such-model:1b' not found"}`.

Notes from this pass:

- Privacy: prompts and completions stay on the local machine; no API key is
  read or sent. Model pulls (`ollama pull`) are the only network operation.
- The daemon must be started separately (`ollama serve`); when it is down,
  commands fail with a connection error rather than hanging.
- `pe eval` keeps a persistent response cache under `.pe/cache` by default.
  Use `--no-cache` (or remove the directory) when revalidating provider
  behavior, otherwise cached responses mask provider changes.
