# Starlark Integration for PE

This package provides Starlark scripting support for PE prompt files, enabling programmable test generation, custom evaluation logic, and dynamic configuration.

## Architecture

```
internal/starlark/
├── starlark.go          # Main Starlark interpreter integration
├── builtins.go          # Built-in PE functions for Starlark
├── evaluator.go         # Starlark-based evaluation engine
├── converter.go         # Convert between Starlark and PE configs
├── validator.go         # Starlark script validation
└── stdlib/              # Standard library modules
    ├── assertions.go    # Assertion helpers
    ├── metrics.go       # Metric calculation functions
    ├── judges.go        # LLM judge templates
    └── analysis.go      # Response analysis utilities
```

## Implementation Overview

### Core Components

1. **Starlark Interpreter** (`starlark.go`)
   - Wraps go.starlark.net interpreter
   - Manages execution environment
   - Handles imports and module loading

2. **Built-in Functions** (`builtins.go`)
   - PE-specific functions exposed to Starlark
   - Assertion helpers (contains, regex_match, etc.)
   - Evaluation utilities (analyze_response, etc.)
   - Scoring functions

3. **Evaluator Engine** (`evaluator.go`)
   - Executes Starlark-defined evaluations
   - Manages evaluation context
   - Collects and aggregates results

4. **Config Converter** (`converter.go`)
   - Converts Starlark values to PE config structures
   - Handles type conversions and validation
   - Supports partial Starlark (embedded in YAML)

## Usage Examples

### Basic Integration

```go
package main

import (
    "github.com/tmc/pe/internal/starlark"
)

// Load and execute a Starlark config
func loadStarlarkConfig(filename string) (*Config, error) {
    interpreter := starlark.New()
    
    // Execute the Starlark file
    result, err := interpreter.ExecFile(filename)
    if err != nil {
        return nil, err
    }
    
    // Convert to PE config
    return starlark.ToConfig(result)
}

// Evaluate using Starlark rubric
func evaluateWithStarlark(output string, rubricFile string) (*EvalResult, error) {
    evaluator := starlark.NewEvaluator()
    
    // Load the rubric
    if err := evaluator.LoadRubric(rubricFile); err != nil {
        return nil, err
    }
    
    // Run evaluation
    return evaluator.Evaluate(output, context)
}
```

### Built-in Functions Implementation

```go
// builtins.go
package starlark

import (
    "go.starlark.net/starlark"
)

// registerBuiltins adds PE-specific functions to Starlark
func registerBuiltins(env starlark.StringDict) {
    env["contains"] = starlark.NewBuiltin("contains", containsImpl)
    env["min_length"] = starlark.NewBuiltin("min_length", minLengthImpl)
    env["analyze_response"] = starlark.NewBuiltin("analyze_response", analyzeImpl)
    // ... more built-ins
}

// Implementation of 'contains' assertion
func containsImpl(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
    var substring string
    if err := starlark.UnpackArgs("contains", args, kwargs, "substring", &substring); err != nil {
        return nil, err
    }
    
    // Return assertion function
    return &assertion{
        name: "contains",
        check: func(output string) bool {
            return strings.Contains(output, substring)
        },
    }, nil
}
```

### Custom Types

```go
// Custom Starlark types for PE

// assertion represents a test assertion
type assertion struct {
    name  string
    check func(string) bool
}

// Starlark interface methods
func (a *assertion) String() string { return fmt.Sprintf("assertion(%s)", a.name) }
func (a *assertion) Type() string   { return "assertion" }
func (a *assertion) Freeze()        {} // assertions are immutable
func (a *assertion) Truth() starlark.Bool { return true }
func (a *assertion) Hash() (uint32, error) { 
    return starlark.String(a.name).Hash() 
}

// rubric represents an evaluation rubric
type rubric struct {
    name     string
    evaluate func(output string, context map[string]interface{}) RubricResult
}

// Similar Starlark interface implementation...
```

## Integration Points

### 1. Command Line

```go
// cmd/pe/test.go
func runTest(cmd *cobra.Command, args []string) error {
    configFile := args[0]
    
    // Check if Starlark file
    if strings.HasSuffix(configFile, ".star") {
        config, err := starlark.LoadConfig(configFile)
        if err != nil {
            return fmt.Errorf("loading Starlark config: %w", err)
        }
        return runTestsWithConfig(config)
    }
    
    // Regular YAML/JSON handling...
}
```

### 2. Embedded Starlark in YAML

```go
// When parsing YAML config
type TestConfig struct {
    Prompt string
    Tests  TestsField
}

type TestsField struct {
    // Either static tests or Starlark script
    Static   []Test
    Starlark string
}

func (t *TestsField) UnmarshalYAML(value *yaml.Node) error {
    // Try to unmarshal as list first
    var tests []Test
    if err := value.Decode(&tests); err == nil {
        t.Static = tests
        return nil
    }
    
    // Try as Starlark string
    var script string
    if err := value.Decode(&script); err == nil {
        t.Starlark = script
        return nil
    }
    
    // Handle map with 'starlark' key
    var m map[string]string
    if err := value.Decode(&m); err == nil {
        if script, ok := m["starlark"]; ok {
            t.Starlark = script
            return nil
        }
    }
    
    return fmt.Errorf("invalid tests field")
}
```

### 3. Runtime Evaluation

```go
// During test execution
func (r *Runner) runTest(test Test, provider Provider) (*TestResult, error) {
    // Get response from provider
    response, err := provider.Complete(test.Prompt, test.Vars)
    if err != nil {
        return nil, err
    }
    
    // Run assertions
    for _, assertion := range test.Assertions {
        // Check if assertion is Starlark-based
        if sa, ok := assertion.(*starlarkAssertion); ok {
            passed := sa.check(response.Text)
            // Record result...
        }
    }
    
    // Run Starlark evaluations if configured
    if r.starlarkEvaluator != nil {
        evalResult, err := r.starlarkEvaluator.Evaluate(response.Text, test.Context)
        // Merge results...
    }
}
```

## Standard Library Modules

### assertions.star

```python
# Standard assertion library
load("re", "re")

def contains(substring):
    """Check if output contains substring"""
    return _create_assertion("contains", lambda out: substring in out)

def regex_match(pattern):
    """Check if output matches regex pattern"""
    compiled = re.compile(pattern)
    return _create_assertion("regex_match", lambda out: compiled.match(out) != None)

def length_between(min_len, max_len):
    """Check if output length is in range"""
    def check(out):
        l = len(out)
        return min_len <= l <= max_len
    return _create_assertion("length_between", check)

# Helper to create assertion objects
def _create_assertion(name, check_fn):
    return struct(
        type = "assertion",
        name = name,
        check = check_fn
    )
```

### metrics.star

```python
# Metrics calculation library

def flesch_kincaid_score(text):
    """Calculate Flesch-Kincaid readability score"""
    sentences = count_sentences(text)
    words = count_words(text)
    syllables = count_syllables(text)
    
    if sentences == 0 or words == 0:
        return 0.0
    
    score = 206.835 - 1.015 * (words / sentences) - 84.6 * (syllables / words)
    return max(0.0, min(100.0, score))

def response_metrics(output):
    """Calculate comprehensive response metrics"""
    return {
        "length": len(output),
        "words": count_words(output),
        "sentences": count_sentences(output),
        "readability": flesch_kincaid_score(output),
        "unique_words": len(set(tokenize(output))),
        "avg_sentence_length": count_words(output) / max(1, count_sentences(output))
    }
```

## Testing

### Unit Tests

```go
// starlark_test.go
func TestStarlarkAssertion(t *testing.T) {
    script := `
def test_contains():
    assert contains("hello")("hello world") == True
    assert contains("goodbye")("hello world") == False
    
test_contains()
`
    
    interp := New()
    _, err := interp.ExecString(script)
    assert.NoError(t, err)
}

func TestStarlarkRubric(t *testing.T) {
    script := `
def quality_rubric(output, context):
    score = 0.0
    if len(output) > 100:
        score += 0.5
    if "example" in output.lower():
        score += 0.5
    return {"score": score, "pass": score >= 0.5}
`
    
    interp := New()
    result, err := interp.ExecString(script)
    assert.NoError(t, err)
    
    rubric := result["quality_rubric"]
    // Test the rubric function...
}
```

## Performance Considerations

1. **Script Caching**: Cache compiled Starlark scripts
2. **Sandboxing**: Starlark is already sandboxed, but limit resource usage
3. **Timeout**: Set execution timeouts for user scripts
4. **Memory Limits**: Monitor memory usage during execution

## Security

Starlark provides built-in security features:
- No file I/O
- No network access  
- No system calls
- Deterministic execution
- Memory safe

Additional PE-specific restrictions:
- Limited computation time
- Restricted built-in functions
- No access to sensitive PE internals

## Possible Enhancements

1. **Module System**: Support for importing shared Starlark modules
2. **Debugging**: Step-through debugging for Starlark scripts
3. **Type Checking**: Static type analysis for Starlark configs
4. **IDE Support**: Language server for Starlark PE files
5. **Performance**: JIT compilation for hot paths
