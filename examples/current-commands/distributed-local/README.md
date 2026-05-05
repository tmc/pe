# Distributed Local

Provider-free smoke for the local deterministic scheduler:

```bash
pe exp distributed --workers 2 --format json tasks.json
pe exp distributed --format text tasks.json
```

Run with an installed `pe`:

```bash
./smoke.sh
```

Or point the smoke at a freshly built binary:

```bash
go build -o /tmp/pe-distributed-local ./cmd/pe
PE_BIN=/tmp/pe-distributed-local ./examples/current-commands/distributed-local/smoke.sh
```
