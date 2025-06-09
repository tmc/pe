# PE Starlark Extension Usage Guide

The PE Starlark extension provides powerful, programmable evaluation capabilities using the Starlark language (a Python-like configuration language). This extension allows you to write dynamic tests, custom rubrics, and complex evaluation logic.

## Quick Start

### 1. Validate a Starlark Test File

```bash
cd ext/starlark/cmd/starlark-demo
go run main.go -mode=validate -file=../../examples/basic_test.star
```

### 2. List Available Test Functions

```bash
go run main.go -mode=list -file=../../examples/basic_test.star
```

### 3. Run a Specific Test

```bash
go run main.go -mode=eval -file=../../examples/basic_test.star \
  -response="This is a comprehensive analysis with examples and conclusion" \
  -test=test_comprehensive
```

### 4. Run Complete Test Suite

```bash
go run main.go -mode=suite -file=../../examples/basic_test.star \
  -response="This analysis provides examples and evidence with clear reasoning."
```

## Example Results

### Single Test Evaluation

Running `test_comprehensive` on a good response:

```json
{
  "pass": true,
  "score": 1,
  "reason": "Score: 100.0/100 (100%)",
  "details": {
    "breakdown": [
      "Good length (100+ chars)",
      "Contains 'analysis' (+25 points)",
      "Contains 'example' (+25 points)",
      "Contains 'conclusion' (+20 points)"
    ],
    "raw_score": 100,
    "max_score": 100
  }
}
```

### Test Suite Results

Running all tests on a response:

```json
{
  "file": "../../examples/basic_test.star",
  "total_tests": 5,
  "passed_tests": 3,
  "failed_tests": 2,
  "pass_rate": 0.6,
  "average_score": 0.3024,
  "suite_passed": false,
  "results": {
    "test_comprehensive": { "pass": true, "score": 1.0 },
    "test_contains_analysis": { "pass": true },
    "test_min_length": { "pass": true },
    "test_quality_rubric": { "pass": false, "score": 0.512 },
    "test_word_count_range": { "pass": false }
  }
}
```

### Dynamic Test Generation

The `simple_dynamic.star` example shows programmatic test creation:

```json
{
  "pass": true,
  "score": 0.5,
  "reason": "Dynamic suite: 4/8 passed",
  "details": {
    "individual_results": {
      "test_min_length_very_short": { "pass": true },
      "test_min_length_short": { "pass": false },
      "test_contains_analysis": { "pass": true },
      "test_contains_example": { "pass": true }
    },
    "pass_rate": 0.5,
    "passed_tests": 4,
    "total_tests": 8
  }
}
```

## Available Built-in Functions

- `contains(text, substr)` - Check if text contains substring
- `min_length(text, min)` - Minimum length validation
- `max_length(text, max)` - Maximum length validation
- `equals(a, b)` - Exact string comparison
- `regex_match(text, pattern)` - Regular expression matching
- `word_count(text)` - Count words in text

## Writing Test Functions

Test functions must:
1. Start with `test_` prefix
2. Take a `response` parameter
3. Return either:
   - `True`/`False` for simple pass/fail
   - Dictionary with `pass`, `score`, `reason`, and optional `details`

### Simple Test

```python
def test_min_length(response):
    return min_length(response, 50)
```

### Detailed Test

```python
def test_comprehensive(response):
    score = 0
    if min_length(response, 100):
        score += 50
    if contains(response, "analysis"):
        score += 50
    
    final_score = score / 100.0
    return {
        "pass": final_score >= 0.7,
        "score": final_score,
        "reason": "Score: " + str(score) + "/100",
        "details": {"raw_score": score}
    }
```

## Dynamic Test Generation

Create test functions programmatically:

```python
def generate_length_tests():
    configs = [{"min": 25, "name": "short"}, {"min": 100, "name": "long"}]
    tests = {}
    
    for config in configs:
        def length_test(response, minimum=config["min"]):
            return min_length(response, minimum)
        tests["test_" + config["name"]] = length_test
    
    return tests
```

## Integration with PE

This extension can be integrated into PE's main command structure as:

```bash
pe starlark eval test.star "response text" [test_function]
pe starlark suite test.star "response text"
pe starlark list test.star
pe starlark validate test.star
pe starlark discover [directory]
```

## Benefits

1. **Programmable**: Full programming language for complex evaluation logic
2. **Safe**: Sandboxed execution environment
3. **Readable**: Python-like syntax familiar to many users
4. **Flexible**: Support for any evaluation pattern you can express
5. **Self-Contained**: All test logic contained in .star files
6. **Extensible**: Easy to add new built-in functions as needed

The Starlark extension transforms PE from a static configuration tool into a fully programmable evaluation platform, enabling sophisticated testing workflows while maintaining safety and simplicity.