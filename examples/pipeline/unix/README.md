# Unix Pipeline Examples

PE follows Unix philosophy - each command does one thing well and can be composed via pipes.

## Basic Pipeline

```bash
# Chain evaluations through filters and analysis
pe eval config.yaml | pe filter --contains pass | pe analyze --metrics readability
```

## Pattern Filtering Pipeline

```bash
# Save rows that mention failures
pe eval config.yaml | \
  pe filter --contains fail | \
  tee failures.txt | \
  pe stats
```

## Text Transformation Pipeline

```bash
# Normalize matching output
pe eval config.yaml | \
  pe filter --pattern PASS | \
  pe filter --transform lowercase
```

## Stream Pipeline

```bash
# Pass evaluation output through the stream command
pe eval config.yaml | pe stream
```

## Batch Processing

```bash
# Process multiple configs
for config in configs/*.yaml; do
  pe eval "$config" | pe filter --contains pass
done | pe stats
```

## Data Transformation

```bash
# Extract a field from newline-delimited JSON
printf '{"provider":"openai","score":1}\n' | pe filter --json .provider
```

## Advanced Composition

```bash
# Multi-stage processing
pe eval stage1.yaml | \
  pe filter --contains pass | \
  pe ask --template "Improve this: {{.}}" | \
  pe stream
```

## Tips

1. Use `tee` to save intermediate results
2. Combine with standard Unix tools (`grep`, `jq`, `awk`)
3. Use `pe stream` for line-oriented pass-through
4. Chain multiple `pe filter` commands for complex filtering
5. Use `pe stats` at the end for summaries
