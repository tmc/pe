# Release Local Workflows

Offline smoke examples for release-facing local workflows:

- `pe diff --fail-on-regression`
- `pe mod tidy --json --write`
- `pe exp attest manifest` and `pe exp attest verify`
- `pe exp cache manifest put` and `pe exp cache manifest verify`

Run with an installed `pe`:

```bash
./smoke.sh
```

Or point the smoke at a freshly built binary:

```bash
go build -o /tmp/pe-release-local-workflows ./cmd/pe
PE_BIN=/tmp/pe-release-local-workflows ./examples/current-commands/release-local-workflows/smoke.sh
```
