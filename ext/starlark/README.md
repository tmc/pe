# Starlark Extension for PE

This experimental extension provides Starlark-based evaluation capabilities for PE, allowing you to write dynamic tests, custom rubrics, and complex evaluation logic using a Python-like configuration language.

## Features

- **Dynamic Test Generation**: Create test matrices programmatically
- **Custom Evaluation Logic**: Write complex rubrics and scoring functions
- **Built-in Functions**: PE-specific functions for common assertion patterns
- **Flexible Results**: Return boolean, numeric scores, or detailed evaluation objects

## Built-in Functions

### Text Assertions
- `contains(text, substr)` - Check if text contains substring
- `equals(a, b)` - Exact string comparison
- `regex_match(text, pattern)` - Regular expression matching
- `min_length(text, min)` - Minimum length validation
- `max_length(text, max)` - Maximum length validation
- `word_count(text)` - Count words in text

### Utilities
- `struct(**kwargs)` - Create structured objects

## Example Usage

### Basic Test Function

```python
def test_response_quality(response):
    """Basic quality test with multiple criteria"""
    
    # Check minimum requirements
    if not min_length(response, 50):
        return {
            "pass": False,
            "reason": "Response too short (minimum 50 characters)",
            "score": 0.0
        }
    
    # Calculate score based on criteria
    score = 0.0
    details = {}
    
    # Length scoring (0-30 points)
    length_score = min(30, word_count(response) * 2)
    score += length_score
    details["length_score"] = length_score
    
    # Content scoring (0-70 points)
    content_score = 0
    if contains(response, "analysis"):
        content_score += 20
    if contains(response, "example"):
        content_score += 25
    if contains(response, "conclusion"):
        content_score += 25
    
    score += content_score
    details["content_score"] = content_score
    
    # Final result
    final_score = score / 100.0
    return {
        "pass": final_score >= 0.7,
        "score": final_score,
        "reason": "Quality score: {:.1%}".format(final_score),
        "details": details
    }
```

### Dynamic Test Generation

```python
def generate_tests():
    """Generate a matrix of tests across different criteria"""
    
    tests = []
    
    # Generate tests for different response lengths
    for min_len in [25, 50, 100, 200]:
        def length_test(response, min_length=min_len):
            passed = min_length(response, min_length)
            return {
                "pass": passed,
                "reason": f"Length check (min {min_length}): {'PASS' if passed else 'FAIL'}"
            }
        
        tests.append({
            "name": f"min_length_{min_len}",
            "test": length_test
        })
    
    # Generate content tests
    required_terms = ["analysis", "example", "conclusion", "summary"]
    for term in required_terms:
        def content_test(response, required_term=term):
            passed = contains(response, required_term)
            return {
                "pass": passed,
                "reason": f"Contains '{required_term}': {'PASS' if passed else 'FAIL'}"
            }
        
        tests.append({
            "name": f"contains_{term}",
            "test": content_test
        })
    
    return tests
```

### Multi-Dimensional Rubric

```python
def comprehensive_rubric(response):
    """Comprehensive evaluation across multiple dimensions"""
    
    dimensions = {}
    
    # Clarity (0-100)
    clarity_score = 0
    if min_length(response, 30):
        clarity_score += 25
    if word_count(response) >= 20:
        clarity_score += 25
    if contains(response, "example"):
        clarity_score += 25
    if not contains(response, "unclear"):
        clarity_score += 25
    
    dimensions["clarity"] = clarity_score
    
    # Completeness (0-100)  
    completeness_score = 0
    required_elements = ["introduction", "analysis", "conclusion"]
    for element in required_elements:
        if contains(response, element):
            completeness_score += 33
    
    dimensions["completeness"] = completeness_score
    
    # Accuracy (0-100)
    accuracy_score = 100  # Start optimistic
    if contains(response, "incorrect"):
        accuracy_score -= 30
    if contains(response, "error"):
        accuracy_score -= 20
    if contains(response, "mistake"):
        accuracy_score -= 15
    
    dimensions["accuracy"] = max(0, accuracy_score)
    
    # Calculate weighted overall score
    weights = {"clarity": 0.3, "completeness": 0.4, "accuracy": 0.3}
    overall_score = sum(dimensions[dim] * weights[dim] for dim in dimensions) / 100
    
    # Determine pass/fail
    passing = overall_score >= 0.7
    
    return {
        "pass": passing,
        "score": overall_score,
        "reason": f"Overall score: {overall_score:.1%} ({'PASS' if passing else 'FAIL'})",
        "dimensions": dimensions,
        "weights": weights
    }
```

## Usage from Go

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/tmc/pe/ext/starlark"
)

func main() {
    evaluator := starlark.NewEvaluator()
    
    // Load Starlark test file
    content, _ := os.ReadFile("test.star")
    globals, err := evaluator.EvalFile("test.star", content)
    if err != nil {
        panic(err)
    }
    
    // Run a specific test
    response := "This is a detailed analysis with examples and a clear conclusion."
    result, err := evaluator.EvaluateTest(globals, "test_response_quality", response)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Test passed: %v\n", result.Pass)
    fmt.Printf("Score: %.2f\n", result.Score)
    fmt.Printf("Reason: %s\n", result.Reason)
}
```

## Benefits

1. **Programmable**: Full programming language for complex logic
2. **Readable**: Python-like syntax familiar to many users
3. **Safe**: Sandboxed execution environment
4. **Flexible**: Support for any evaluation pattern
5. **Extensible**: Easy to add new built-in functions
6. **Self-Contained**: No external dependencies for test definitions

## Future Enhancements

- More built-in functions (sentiment analysis, readability metrics)
- Integration with external APIs and models
- Statistical functions for multi-run analysis
- Template system for common evaluation patterns
- IDE support with syntax highlighting and completion