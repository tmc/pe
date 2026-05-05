# Consensus Local

Provider-free smoke for deterministic weighted local consensus:

```bash
pe exp consensus --input votes.json --output consensus.json
pe exp consensus --input - <votes.json
```

Run with an installed `pe`:

```bash
./smoke.sh
```

Or point the smoke at a freshly built binary:

```bash
go build -o /tmp/pe-consensus-local ./cmd/pe
PE_BIN=/tmp/pe-consensus-local ./examples/current-commands/consensus-local/smoke.sh
```
