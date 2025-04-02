# Testing Guide

This document outlines the testing strategy for the Prompt Engineering toolkit.

## Testing Philosophy

The PE toolkit follows these testing principles:

1. **Functionality First**: Tests focus on the behavior of components
2. **Interface Stability**: Tests should be resilient to internal implementation changes
3. **Reproducibility**: Tests must be reproducible and deterministic

## Testing Layers

### 1. Unit Tests

Unit tests verify the behavior of individual functions and methods:

- Located alongside the code they test
- Fast and focused on single components
- Independent of external services

Examples:
- Template processing tests
- Assertion utility tests
- Configuration validation tests

### 2. Component Tests

Component tests verify the behavior of entire packages or subsystems:

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

## Test Organization

### Directory Structure

- **internal/*/\*_test.go**: Unit tests alongside the code
- **tests/\*_test.go**: Additional specialized tests

### Naming Conventions

- Test functions: `TestXxx` where Xxx describes what's being tested
- Test suites/subtests: `t.Run("Description", func(t *testing.T) {...})`

## Running Tests

### Run All Tests

```bash
go test ./...
```

### Run Specific Tests

```bash
# Run tests in a specific package
go test ./internal/template

# Run a specific test
go test ./internal/template -run TestRenderTemplate
```

## Mocking

When testing components that depend on external services, use mocks or fakes:

1. **Interface-Based Mocking**: Create implementations of interfaces for testing
2. **Dependency Injection**: Pass mocks to the code being tested

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
    if !ok {
        return nil, fmt.Errorf("no mock response for prompt: %s", req.Prompt)
    }
    
    return &llm.CompletionResponse{
        Text: response,
        Model: req.Model,
    }, nil
}
```