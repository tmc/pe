# Testing Approach for PE Toolkit

This document outlines the testing strategy and approach used in the Prompt Engineering toolkit.

## Testing Philosophy

The PE toolkit follows these testing principles:

1. **Functionality First**: Tests focus on the behavior of components, not implementation details
2. **Interface Stability**: Tests should be resilient to internal implementation changes
3. **Real-World Scenarios**: Tests cover actual use cases and workflows
4. **Isolation**: Components are tested in isolation with clear boundaries
5. **Reproducibility**: Tests must be reproducible and deterministic

## Testing Layers

### 1. Unit Tests

Unit tests verify the behavior of individual functions and methods. These tests are:

- Located alongside the code they test
- Fast and focused on single components
- Independent of external services

Examples:
- Template processing tests
- Assertion utility tests
- Configuration validation tests

### 2. Component Tests

Component tests verify the behavior of entire packages or subsystems:

- Located in the `tests/basic_tests/` directory
- Test public APIs and interfaces
- May use mocks for external dependencies

Examples:
- Template engine tests
- Evaluator package tests
- Provider integration tests

### 3. Integration Tests

Integration tests verify that components work together correctly:

- Test multiple components together
- May require external services or APIs
- Focus on real-world workflows

Examples:
- End-to-end evaluation flows
- Multi-provider testing
- Configuration processing and validation

### 4. Benchmark Tests

Benchmark tests measure performance and identify bottlenecks:

- Focus on performance-critical paths
- Useful for comparing different implementations
- Help identify regressions

Examples:
- Template rendering performance
- Provider request handling
- Large configuration processing

## Test Organization

### Directory Structure

- **internal/*/\*_test.go**: Unit tests alongside the code
- **tests/basic_tests/**: Component and integration tests
- **tests/\*_test.go**: Additional specialized tests
- **scripts/run-all-tests.sh**: Test runner script

### Naming Conventions

- Test functions: `TestXxx` where Xxx describes what's being tested
- Test suites/subtests: `t.Run("Description", func(t *testing.T) {...})`
- Benchmark functions: `BenchmarkXxx`

### Testing Tools and Framework

The toolkit uses the standard Go testing package with these enhancements:

1. **Table-Driven Tests**: For testing multiple scenarios with the same logic
2. **Subtests**: For organizing tests into logical groups
3. **Script-Based Tests**: For testing file-based operations and CLI tools
4. **Basic Assertion Tools**: For complex validation logic

## Running Tests

### Run All Tests

```bash
./scripts/run-all-tests.sh
```

### Run Specific Tests

```bash
# Run tests in a specific package
go test ./internal/template

# Run a specific test
go test ./internal/template -run TestRenderTemplate

# Run with verbose output
go test -v ./internal/template

# Run with coverage
go test -cover ./internal/template
```

## Test Data Management

The toolkit uses several approaches for test data:

1. **Inline Test Data**: Small test cases defined directly in test files
2. **Generated Test Data**: Programmatically generated data for comprehensive testing
3. **Test Files**: External files for testing file operations
4. **Mock Responses**: Simulated LLM responses for predictable testing

## Writing Good Tests

### Test Structure

Each test should follow this general structure:

1. **Setup**: Prepare the test environment and inputs
2. **Execute**: Run the code being tested
3. **Verify**: Check the results against expected outcomes
4. **Teardown**: Clean up any resources (often handled by defer)

Example:

```go
func TestTemplateRendering(t *testing.T) {
    // Setup
    tmpl := template.NewTemplate(
        "Hello, {{.name}}\!",
        map[string]interface{}{"name": "World"}
    )
    
    // Execute
    result, err := tmpl.Process()
    
    // Verify
    if err \!= nil {
        t.Errorf("unexpected error: %v", err)
    }
    if result \!= "Hello, World\!" {
        t.Errorf("expected 'Hello, World\!', got %q", result)
    }
    
    // Teardown (implicit with Go GC)
}
```

### Table-Driven Tests

For testing multiple scenarios, use table-driven tests:

```go
func TestTemplateConditionals(t *testing.T) {
    tests := []struct {
        name     string
        template string
        vars     map[string]interface{}
        expected string
    }{
        {
            name:     "simple conditional",
            template: "{{if .show}}Visible{{else}}Hidden{{end}}",
            vars:     map[string]interface{}{"show": true},
            expected: "Visible",
        },
        {
            name:     "nested conditional",
            template: "{{if .outer}}{{if .inner}}Both{{else}}Outer{{end}}{{else}}None{{end}}",
            vars:     map[string]interface{}{"outer": true, "inner": false},
            expected: "Outer",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpl := template.NewTemplate(tt.template, tt.vars)
            result, err := tmpl.Process()
            
            if err \!= nil {
                t.Errorf("unexpected error: %v", err)
            }
            if result \!= tt.expected {
                t.Errorf("expected %q, got %q", tt.expected, result)
            }
        })
    }
}
```

### Testing Error Conditions

Always test error conditions and edge cases:

```go
func TestTemplateErrors(t *testing.T) {
    tests := []struct {
        name        string
        template    string
        vars        map[string]interface{}
        expectError bool
    }{
        {
            name:        "missing variable",
            template:    "Hello, {{.missing}}\!",
            vars:        map[string]interface{}{},
            expectError: true,
        },
        {
            name:        "syntax error",
            template:    "Hello, {{.name\!",
            vars:        map[string]interface{}{"name": "World"},
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpl := template.NewTemplate(tt.template, tt.vars)
            _, err := tmpl.Process()
            
            if (err \!= nil) \!= tt.expectError {
                t.Errorf("expectError=%v, got error: %v", tt.expectError, err)
            }
        })
    }
}
```

## Mocking

When testing components that depend on external services, use mocks or fakes:

1. **Interface-Based Mocking**: Create implementations of interfaces for testing
2. **Dependency Injection**: Pass mocks to the code being tested
3. **Fake Implementations**: Simplified versions of real components

Example mock for a Provider:

```go
type MockProvider struct {
    models []llm.Model
    responses map[string]string
}

func (p *MockProvider) Name() string {
    return "mock"
}

func (p *MockProvider) Models() []llm.Model {
    return p.models
}

func (p *MockProvider) GetModel(name string) (llm.Model, error) {
    for _, m := range p.models {
        if m.Name() == name {
            return m, nil
        }
    }
    return nil, llm.ErrModelNotFound
}

func (p *MockProvider) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
    // Return pre-configured response
    response, ok := p.responses[req.Prompt]
    if \!ok {
        return nil, fmt.Errorf("no mock response for prompt: %s", req.Prompt)
    }
    
    return &llm.CompletionResponse{
        Text: response,
        Model: req.Model,
    }, nil
}
```
