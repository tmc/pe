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

# Evaluate prompts against live LLM providers
./pe eval ./example/getting-started/promptfooconfig.yaml --save-db

# View evaluation results in browser
./pe view <eval-id>
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

## Command Usage

### Evaluation
The `pe eval` command runs evaluations against LLM providers using the provided configuration:

```bash
# Basic usage
pe eval config.yaml

# Save results to database for later viewing
pe eval config.yaml --save-db
```

### Viewing Results
The `pe view` command provides a browser-based UI for viewing evaluation results:

```bash
# List available evaluations
pe view

# View a specific evaluation by ID
pe view <eval-id>

# View results from a specific file
pe view -f <path-to-results.json>

# Specify a custom port for the HTTP server
pe view <eval-id> -p 8888
```

The viewer provides a rich UI for exploring test results, including:
- Statistics on pass/fail rates and token usage
- Navigation of test cases by input and language
- Detailed views of LLM responses
- Assertion results and success status

## Provider Integration
When implementing a new LLM provider:
1. Create a new file in `internal/llm/providers/`
2. Implement the `Provider` interface from `internal/llm/llm.go`
3. Register the provider in `internal/llm/providers/providers.go`
4. Add the provider to the example configuration

Follow Go best practices and maintain consistency with the existing codebase.

## CGPT Integration
The toolkit integrates with CGPT for LLM provider access. It supports:

- Running evaluations with multiple LLM providers concurrently
- Handling provider-specific configurations (model, temperature, max tokens)
- Parallelizing evaluations for efficiency
- Evaluating assertions against model outputs
- Token usage tracking and cost estimation

CGPT providers are specified in configuration files using the format:
```yaml
providers:
  - googleai:gemini-2.0-flash
  - openai:gpt-4o
  - anthropic:claude-3-5-haiku-latest
```

The PE tool automatically sets up appropriate backends for each provider when running tests.