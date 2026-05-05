# Executable Text

PE treats prompt files as executable, templated, composable text. A file can be
plain text by default, or it can declare inputs, metadata, safety policy, and
placement rules for where data, prompts, providers, and tools may run.

Policy composition is conservative: child artifacts can narrow parent
constraints, but they cannot loosen them.

## Files

- `plain.prompt` - Plain executable text with no declared inputs.
- `templated.prompt` - Templated text with command-line inputs.
- `workflow.pe.yaml` - Composable text workflow sketch.
- `pe.mod` - Capability and placement policy sketch for the example directory.
- `smoke.sh` - Provider-free checks for the runnable plain and templated files.

## Run

Use an installed `pe`:

```bash
./smoke.sh
```

Or point the smoke at a local binary:

```bash
go build -o /tmp/pe-executable-text ./cmd/pe
PE_BIN=/tmp/pe-executable-text ./examples/executable-text/smoke.sh
```

## Notes

`workflow.pe.yaml` and `pe.mod` are future-facing sketches. They show the
shape of composable text and conservative placement policy without requiring a
provider or network call.
