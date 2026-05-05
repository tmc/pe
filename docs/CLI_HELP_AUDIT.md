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
analyze ask benchmark build cat collect compose config convert diff doc edit eval eval-prompt evolve exp expand experimental extract filter fmt fusion gaso get init interactive mod optimize pe2 plugin profile prompt push reduce run run-text security semantic serve stats stream template test textgrad version vet view watch work
```

Documentation reconciliation:

- `docs/CLI_REFERENCE.md` includes command-table entries for every root command
  listed above.
- `expand` and `version` were added after this audit found they were missing
  from the command inventory.
- `compose`, `config`, `evolve`, `fusion`, `gaso`, `optimize`, `pe2`,
  `semantic`, and `textgrad` are current root commands and should remain visible
  in release-facing command inventories.
