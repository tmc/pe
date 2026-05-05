# Core Concepts

The current status and command surface are tracked in:

- [CURRENT_STATUS.md](../../CURRENT_STATUS.md)
- [CLI_REFERENCE.md](../../CLI_REFERENCE.md)
- [CLI_HELP_AUDIT.md](../../CLI_HELP_AUDIT.md)

Stable concepts:

- Prompts and executable text are version-controlled text artifacts.
- Templates use Go template syntax such as `{{.name}}`.
- `pe.mod` describes module dependencies and can carry capability policy.
- Commands compose through files, stdin/stdout, and explicit local artifacts.
- Experimental workflows live under `pe exp` and `pe experimental`.
