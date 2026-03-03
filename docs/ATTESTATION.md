# Cryptographic Attestation

Attestation is currently exposed as a prototype command group:

```bash
pe exp attest --help
```

## Current Status

- `pe exp attest` exists in the CLI.
- It is marked prototype and does not currently expose stable subcommands.
- Command shape and behavior may change as the feature is finalized.

## Recommended Usage Today

Use the prototype command help to inspect the current interface:

```bash
pe exp --help
pe exp attest --help
```

For production workflows, rely on stable commands (`pe run`, `pe eval`, `pe test`, `pe security`) until attestation exits prototype status.
