# CLI Reference

For the most up-to-date reference, run `pe help` or `pe [command] --help`.

## Primary Commands

| Command | Description |
| :--- | :--- |
| `pe run` | Execute a prompt immediately |
| `pe eval` | Run an evaluation suite |
| `pe view` | View evaluation results reports |
| `pe interactive` | Start interactive REPL |
| `pe experimental optimize` | Experimental metaprompt optimization |
| `pe mod` | Manage module dependencies |
| `pe exp` | Prototype command group |
| `pe experimental` | Research command group |

## Common Command Flags

Global flags vary by command. Run `pe [command] --help` for exact flags.
Common examples include:

*   `--provider`: Select LLM provider (for commands that support provider selection).
*   `--var key=val`: Set template variables (for commands that support templating).
*   `--stream`: Stream output to stdout (for commands that support streaming).

## Environment Variables

*   `OPENAI_API_KEY`: API key for OpenAI.
*   `ANTHROPIC_API_KEY`: API key for Anthropic.
*   `PE_DEBUG`: Set to `true` for verbose logs.

## Notes on Experimental Commands

Prototype commands are grouped under `pe exp` and may not expose stable subcommands yet:

```bash
pe exp --help
pe exp attest --help
pe exp distributed --help
pe exp cache --help
```
