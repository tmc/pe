# CLI Reference

For the most up-to-date reference, run `pe help` or `pe [command] --help`.

## Primary Commands

| Command | Description |
| :--- | :--- |
| `pe run` | Execute a prompt immediately |
| `pe eval` | Run an evaluation suite |
| `pe view` | View evaluation results reports |
| `pe interactive` | Start interactive REPL |
| `pe optimize` | Automatically improve prompts |
| `pe mod` | Manage module dependencies |
| `pe serve` | Start HTTP API server |

## Global Flags

*   `--provider`: Select LLM provider (e.g., `openai`, `anthropic`).
*   `--model`: Select specific model (e.g., `gpt-4`).
*   `--var key=val`: Set string variable.
*   `--var-json key='{"a":1}'`: Set JSON variable.
*   `--stream`: Stream output to stdout.
*   `--verbose`: Enable debug logging.

## Environment Variables

*   `OPENAI_API_KEY`: API key for OpenAI.
*   `ANTHROPIC_API_KEY`: API key for Anthropic.
*   `PE_DEBUG`: Set to `true` for verbose logs.
