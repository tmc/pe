# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# PE: Prompt Engineering Toolkit

A Go-based toolkit for prompt engineering implementing cutting-edge metaprompting research, designed following Unix philosophy with composable commands.

## Development Commands

### Building and Testing
```bash
# Build the main binary
make build
# or
go build -o pe ./cmd/pe

# Run all Go tests
make test
# or 
go test ./...

# Run script-based integration tests
make scripttest

# Run a specific scripttest
make test-module

# Clean build artifacts
make clean
```

### Running Tests
```bash
# Run tests with coverage
go test -v -cover ./...

# Run tests for a specific package
go test ./internal/evaluator/

# Run a single test
go test -run TestSpecificFunction ./internal/package/

# Run tests with race detection
go test -race ./...
```

### Linting and Code Quality
The project currently doesn't have explicit linting commands in the Makefile. Use standard Go tools:
```bash
go vet ./...
go fmt ./...
```

## Architecture Overview

### Core Design Principles
- **Unix Philosophy**: Each command does one thing well, composable via pipes
- **Provider Abstraction**: Extensible LLM provider interface with **native OpenAI/Anthropic implementations**
- **Modular Architecture**: Clear separation between CLI, core logic, and providers

### Key Architectural Components

1. **Command Layer** (`cmd/pe/`): All CLI commands and their implementations
2. **Core Engine** (`internal/`): Business logic organized by domain:
   - `metaprompt/`: Advanced optimization algorithms (TextGrad, GASO, semantic backprop)
   - `evaluator/`: Evaluation engine with sophisticated assertion types
   - `inference/`: Provider abstraction and implementations
   - `structured/`: Schema validation and structured output handling
   - `metrics/`: Advanced metrics (BLEU, ROUGE, BERTScore, G-Eval)
   - `distributed/`: P2P networking and distributed execution
   - `consensus/`: Multi-provider consensus mechanisms
3. **Plugin System**: Runtime discovery of `pe-*` executables in PATH

### Provider System
The toolkit uses a provider abstraction for LLM calls:
- **Native providers are primary**: OpenAI (74% test coverage) and Anthropic (73.3% test coverage) fully implemented
- Production-ready implementations in `internal/inference/providers/openai/` and `internal/inference/providers/anthropic/`
- `cgpt` CLI wrapper available for compatibility (`internal/cgpt/`)
- Extensible for new providers via interface in `internal/inference/`

### Pipeline Architecture
Unix-style composable commands that can be chained:
- `ask`: Send prompts to providers
- `stream`: Stream responses
- `filter`: Filter and transform data
- `analyze`: Analyze responses
- `collect`: Aggregate results
- `reduce`: Reduce to final outputs

## Key Commands and Usage Patterns

### Evaluation System
The `pe eval` command supports sophisticated evaluation:
- Pass@N metrics for code generation tasks
- Structured output validation with JSON Schema
- 20+ assertion types for comprehensive testing
- Integration with YAML configuration files

### Metaprompting Commands  
Advanced prompt optimization based on 2024-2025 research:
- `pe optimize`: Multi-stage prompt optimization
- `pe semantic`: Semantic backpropagation and GASO optimization
- `pe evolve`: Evolutionary optimization with genetic algorithms
- `pe compose`: Component-based prompt composition

### Module System
Go-style module management:
- `pe mod init`: Initialize module
- `pe mod download`: Download dependencies  
- `pe mod tidy`: Clean up dependencies
- `pe mod vendor`: Vendor dependencies

## Testing Strategy

### Test Structure
- Unit tests alongside source files (`*_test.go`)
- Integration tests in `tests/scripttest/`
- Test coverage currently ~15%, goal is >50%

### Test Data
- Mock providers for testing without external dependencies
- Test configurations in `test-configs/`
- Example prompts and configurations in `example/`

### Running Specific Tests
```bash
# Test a specific command
go test ./cmd/pe/ -run TestCommandName

# Test with verbose output
go test -v ./internal/evaluator/

# Test with coverage report
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Organization Patterns

### Error Handling
Always wrap errors with context:
```go
if err != nil {
    return fmt.Errorf("failed to process prompt: %w", err)
}
```

### CLI Command Structure
Commands follow cobra patterns in `cmd/pe/`:
- Each command in separate file
- Shared flags and utilities in `commands.go`
- Command registration in `main.go`

### Provider Integration
New providers should implement interfaces in `internal/inference/`:
- `Provider` interface for basic LLM calls
- Registration via `internal/providers/registry.go`
- Configuration via environment variables or config files

## Development Workflow

### Adding New Commands
1. Create command file in `cmd/pe/`
2. Implement cobra.Command following existing patterns
3. Add to command registration in `main.go`
4. Add tests in `cmd/pe/*_test.go`
5. Add scripttest integration test if needed

### Adding New Providers
1. Implement Provider interface in `internal/inference/providers/`
2. Add registration in `internal/providers/register.go`
3. Add tests following existing provider test patterns
4. Update documentation

### Working with Metaprompting
The metaprompting engine in `internal/metaprompt/` implements research-based optimization:
- `optimizer.go`: Unified optimization interface
- `textgrad.go`: Natural language gradient computation  
- `semantic.go`: Semantic backpropagation implementation
- Integration via `pe optimize` and `pe semantic` commands

## Current Development Status

### Implemented ✅
- Core command structure and CLI
- Provider abstraction with cgpt implementation
- Evaluation system with pass@n metrics
- Module management commands
- Pipeline processing commands
- Metaprompting optimization engine
- Attestation and security features

### In Progress 🚧  
- Native provider implementations (OpenAI/Anthropic)
- Distributed execution integration
- Comprehensive test coverage (currently ~15%)
- Module registry implementation

### Known Limitations
- Primary dependency on cgpt CLI wrapper
- Test coverage needs improvement
- Some distributed features incomplete
- Documentation accuracy issues (some features documented but not fully implemented)

## Important Files for New Contributors

- `cmd/pe/main.go`: CLI entry point and command registration
- `cmd/pe/commands.go`: Shared command utilities
- `internal/inference/inference.go`: Provider interface definition
- `internal/evaluator/evaluator.go`: Core evaluation engine
- `internal/metaprompt/optimizer.go`: Optimization interface
- `Makefile`: Build and test commands
- `tests/scripttest/`: Integration test examples