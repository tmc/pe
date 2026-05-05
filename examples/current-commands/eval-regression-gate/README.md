# Eval Regression Gate

This example shows an offline CI gate for evaluation results:

```bash
pe stats current.jsonl
pe diff --fail-on-regression baseline.json current.jsonl
pe diff --fail-on-regression --max-pass-rate-drop 60 --max-score-drop 0.6 --max-latency-increase-ms 60 --max-failure-increase 1 baseline.json current.jsonl
```

`baseline.json` is a passing JSON result set. `current.jsonl` is a JSONL result
set with one failure, lower average score, and higher average latency.

Run the smoke script with an installed `pe`:

```bash
./smoke.sh
```

Or point it at a local binary:

```bash
go build -o /tmp/pe ./cmd/pe
PE_BIN=/tmp/pe ./examples/current-commands/eval-regression-gate/smoke.sh
```
