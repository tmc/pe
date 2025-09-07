# Unix Pipeline Examples

PE follows Unix philosophy - each command does one thing well and can be composed via pipes.

## Basic Pipeline

```bash
# Chain evaluations through filters and analysis
pe eval config.yaml | pe filter --success | pe analyze --metric latency
```

## Cost Optimization Pipeline

```bash
# Find expensive failures
pe eval config.yaml | \
  pe filter --failed --min-cost 0.10 | \
  pe stats --format json > expensive-failures.json
```

## Quality Pipeline

```bash
# Extract high-quality responses
pe eval config.yaml | \
  pe filter --min-score 0.9 | \
  pe stream --select prompt,response,score | \
  tee high-quality.jsonl | \
  pe analyze --metric score
```

## Monitoring Pipeline

```bash
# Real-time monitoring
pe eval config.yaml --stream | \
  pe filter --latency-gt 5000 | \
  pe alert --webhook https://alerts.example.com/slow
```

## Batch Processing

```bash
# Process multiple configs
for config in configs/*.yaml; do
  pe eval "$config" | pe stats --format csv
done | pe aggregate --by provider > results.csv
```

## Data Transformation

```bash
# Convert and process results
pe eval config.yaml | \
  jq '.results[]' | \
  pe filter --provider openai | \
  pe convert --format csv > openai-results.csv
```

## Advanced Composition

```bash
# Multi-stage processing
pe eval stage1.yaml | \
  pe extract --field response | \
  pe run "Improve this: {{.input}}" | \
  pe eval stage2.yaml | \
  pe diff - baseline.json
```

## Tips

1. Use `tee` to save intermediate results
2. Combine with standard Unix tools (`grep`, `jq`, `awk`)
3. Use `--stream` for real-time processing
4. Chain multiple `pe filter` commands for complex filtering
5. Use `pe stats` at the end for summaries