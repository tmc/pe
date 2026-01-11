# Core Concepts

PE is built on a few foundational philosophies that differentiate it from other tools.

## Prompts as Code

PE treats prompts as software artifacts. They should be:
*   **Version Controlled**: Stored in Git.
*   **Modular**: Composed of reusable parts.
*   **Tested**: Verified with automated evaluations.

## The Unix Philosophy

PE provides a suite of small, focused tools that compose via standard streams (stdin/stdout).

*   `pe ask`: Simple request/response tool.
*   `pe filter`: Filter JSON results from a stream.
*   `pe reduce`: Aggregate results.

Example:
```bash
echo "data" | pe ask "Summarize" | pe ask "Extract sentiments"
```

## Providers

PE abstracts LLM backends into **Providers**.
*   **Native Providers**: OpenAI, Anthropic (built-in, no external dependencies).
*   **CLI Providers**: Ollama, LocalAI, generic scripts (configured via YAML).

You can switch providers easily using the `--provider` flag or configuration files, allowing you to test your prompts across models without changing the prompt code.
