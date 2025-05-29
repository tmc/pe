# PE Benchmarks

This directory contains benchmark definitions and tests for PE (Go for Prompts).

## Running Benchmarks

### Quick benchmark:
```bash
pe benchmark example-benchmark.yaml
```

### Detailed benchmark with analysis:
```bash
pe benchmark example-benchmark.yaml \
  --output results/ \
  --format json,markdown \
  --visualize
```

### Compare multiple benchmarks:
```bash
pe benchmark compare \
  results/benchmark-1.json \
  results/benchmark-2.json \
  --metric accuracy,cost
```

## Benchmark Definition Format

Benchmarks are defined in YAML with the following structure:

```yaml
name: "Benchmark Name"
scenarios:
  - id: scenario_1
    base_prompt: "..."
    test_cases: [...]
    
optimization_methods:
  - method: pe2
    iterations: 5
    
providers:
  - name: gpt-4
    temperature: 0.7
    
metrics:
  primary: [accuracy, latency, cost]
  
configuration:
  iterations: 10
  parallel_jobs: 4
```

## Built-in Benchmarks

### Performance Benchmark
Tests raw performance across providers:
```bash
pe benchmark performance --providers all
```

### Optimization Benchmark  
Compares optimization methods:
```bash
pe benchmark optimization --methods all
```

### Cost Efficiency Benchmark
Analyzes cost vs quality tradeoffs:
```bash
pe benchmark cost --scenarios production
```

## Creating Custom Benchmarks

1. Define scenarios that match your use case
2. Select relevant optimization methods
3. Choose appropriate metrics
4. Set success criteria
5. Run and analyze results

## Benchmark Results

Results include:
- Raw measurements
- Statistical analysis
- Visualizations
- Recommendations
- Export in multiple formats

## Integration with CI/CD

```yaml
# .github/workflows/benchmark.yml
- name: Run benchmarks
  run: |
    pe benchmark production.yaml
    pe benchmark compare --baseline main