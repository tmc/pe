# PE: Prompt Engineering Toolkit

## Build & Test Commands
```bash
# Build the main project
go build ./cmd/pe

# Build the workbench tool
go build ./cmd/wb

# Run all tests (including basic tests)
./scripts/run-all-tests.sh

# Run specific package tests
go test ./internal/assertutil
go test ./internal/template
go test ./internal/llm
go test ./internal/evaluator

# Run with verbose output and specific test
go test -v ./internal/template -run TestRenderTemplate

# Run basic tests only (most reliable)
go test -v ./tests/basic_tests

# Run benchmarks
./pe benchmark example/benchmark-config.yaml --iterations 5

# Validate configuration file
./pe vet config.yaml
```

## Code Style Guidelines
- **Imports**: Standard library first, followed by third-party, both alphabetically ordered
- **Error handling**: Use `fmt.Errorf()` with `%w` for wrapping, early returns
- **Naming**: CamelCase (exported), camelCase (unexported), descriptive type names
- **Documentation**: Every exported entity needs comments, complete sentences
- **Testing**: Table-driven tests with subtests (`t.Run("Case", func(t *testing.T){...})`)
- **Types**: Use interfaces for abstraction, composition over inheritance
- **Formatting**: Standard Go formatting (gofmt), 4-space indentation
- **Structure**: Keep provider implementations in separate packages
- **Error types**: Define common errors as package-level variables

## Project Structure
- **cmd/pe/**: Main command-line tool for evaluation and testing
- **cmd/wb/**: Workbench tool for interactive prompt engineering
- **internal/llm/**: Provider abstraction for multiple LLM services
- **internal/llm/providers/**: Implementations for different LLM providers (OpenAI, Anthropic, Google AI)
- **internal/cgpt/**: CGPT command-line tool integration
- **internal/template/**: Template processing with variable substitution
- **internal/evaluator/**: Evaluation logic for prompts and configurations
- **internal/assertutil/**: Assertion utilities for testing
- **tests/basic_tests/**: Core functionality unit tests

## Provider Integration
When implementing a new LLM provider:
1. Create a new file in `internal/llm/providers/`
2. Implement the `Provider` interface from `internal/llm/llm.go`
3. Register the provider in `internal/llm/providers/providers.go`
4. Add the provider to the example configuration

Follow Go best practices and maintain consistency with the existing codebase.