# PE Scripttest Tests

This directory contains tests for PE (Go for Prompts) using `rsc.io/script/scripttest`.

## Test Structure

```
tests/
├── run-tests.sh              # Shell wrapper for go test
├── scripttest_test.go        # Go test runner using rsc.io/script/scripttest
├── scripttest/
│   ├── basic/               # Basic workflow tests
│   │   ├── 01-init-and-run.txt
│   │   ├── 02-test-and-validate.txt
│   │   ├── 03-version-control.txt
│   │   ├── 04-optimization.txt
│   │   └── 05-cache-and-security.txt
│   ├── advanced/            # Advanced feature tests
│   └── benchmarks/          # Performance benchmarks
│       ├── 01-performance-benchmark.txt
│       ├── 02-optimization-benchmark.txt
│       └── 03-cost-efficiency-benchmark.txt
```

## Running Tests

### Run all tests:
```bash
./run-tests.sh
# or
go test -v ./tests/...
```

### Update test expectations:
```bash
go test -v ./tests/... -update
```

### Run with specific pattern:
```bash
go test -v ./tests/... -run TestScripts/basic
```

## Test Format

Tests use scripttest format where:
- Lines starting with `$` are commands to execute
- Lines without `$` are expected output
- Comments start with `#`

Example:
```
# Test: Basic prompt execution
$ pe run "What is 2+2?"
4
```

## Writing New Tests

1. Create a new `.txt` file in the appropriate directory
2. Use scripttest format for commands and expected output
3. Include comments to explain what's being tested
4. Run the test to verify it works

### Test Categories

#### Basic Tests
- Core command functionality
- Standard workflows
- Error handling

#### Advanced Tests  
- Complex features
- Integration scenarios
- Edge cases

#### Benchmark Tests
- Performance measurements
- Cost analysis
- Optimization comparisons

## Benchmark Descriptions

Benchmarks can be defined declaratively:

```yaml
# benchmark.yaml
name: "My Benchmark"
scenarios:
  - prompt: "Test prompt"
    providers: [gpt-4, claude-3]
    iterations: 10
metrics:
  - latency
  - accuracy
  - cost
```

Then run with:
```bash
pe benchmark benchmark.yaml
```

## Continuous Integration

Add to your CI pipeline:

```yaml
# .github/workflows/test.yml
- name: Run PE Tests
  run: |
    cd tests
    ./run-tests.sh
```

## Mock Mode

For testing without real API calls, use mock providers:

```bash
export PE_MOCK_MODE=true
./run-tests.sh
```

This will use predefined responses instead of making actual API calls.