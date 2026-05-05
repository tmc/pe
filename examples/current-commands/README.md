# Current Command Examples

This directory contains small offline examples for current commands that are
easy to miss in the larger evaluation examples. The smoke script uses a temp
directory for generated files and does not require provider credentials.

## Files

- `build.yaml` - Minimal input for `pe build`.
- `convert.yaml` - Minimal input for `pe convert`.
- `eval-regression-gate/` - Regression gate fixture for `pe diff --fail-on-regression`.
- `prompt.txt` - Plain prompt file used by prompt-oriented commands.
- `release-local-workflows/` - Release-facing local workflow smokes for `diff`,
  `mod tidy`, `exp attest`, and `exp cache`.
- `smoke.sh` - Runs the commands below with `PE_TEST_MODE=true`.

## Run

Use an installed `pe`:

```bash
./smoke.sh
```

Or point the script at a local binary:

```bash
go build -o /tmp/pe ./cmd/pe
PE_BIN=/tmp/pe ./examples/current-commands/smoke.sh
```

## Commands Covered

```bash
PE_TEST_MODE=true pe ask "What is the capital of France?" --provider mock
pe template list --format json
pe prompt init demo.prompt --force
pe plugin list
pe profile status
pe build build.yaml --output built.txt --minify
pe convert convert.yaml converted.json
pe collect --jobs 3
pe reduce --sum sum
pe watch --help
PE_BIN=/tmp/pe ./eval-regression-gate/smoke.sh
PE_BIN=/tmp/pe ./release-local-workflows/smoke.sh
```

`pe watch` is represented by `--help` because the normal command is a
long-running file watcher.
