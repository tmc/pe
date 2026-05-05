# CLI Help Audit

Last run: 2026-05-05
Branch: `exp`

Command used:

```bash
go run ./cmd/pe --help
```

Root commands from generated help, excluding Cobra's built-in `help` and
`completion`:

```text
analyze ask benchmark build cat collect convert diff doc edit eval eval-prompt exp expand experimental extract filter fmt get init interactive mod plugin profile prompt push reduce run run-text security serve stats stream template test version vet view watch work
```

Documentation reconciliation:

- `docs/CLI_REFERENCE.md` includes command-table entries for every root command
  listed above.
- `expand` and `version` were added after this audit found they were missing
  from the command inventory.
