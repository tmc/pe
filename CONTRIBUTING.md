# Contributing to PE

Thank you for your interest in contributing to PE! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful, inclusive, and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- API keys for LLM providers (for testing)
- Basic familiarity with prompt engineering concepts

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/pe.git
   cd pe
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

3. **Set Up Environment**
   ```bash
   # Copy example environment file
   cp .env.example .env
   
   # Add your API keys
   export OPENAI_API_KEY="your-key"
   export ANTHROPIC_API_KEY="your-key"
   ```

4. **Build and Test**
   ```bash
   # Build the project
   go build -o pe cmd/pe/main.go
   
   # Run tests
   go test ./...
   
   # Test the CLI
   ./pe --help
   ```

5. **Verify Installation**
   ```bash
   # Create a test configuration
   ./pe init test-config.yaml
   
   # Run a test evaluation
   ./pe eval test-config.yaml --dry-run
   ```

## Contributing Guidelines

### Types of Contributions

We welcome various types of contributions:

- **Bug Reports**: Help us identify and fix issues
- **Feature Requests**: Suggest new functionality
- **Code Contributions**: Implement features, fix bugs, improve performance
- **Documentation**: Improve docs, add examples, fix typos
- **Testing**: Add test cases, improve test coverage
- **Examples**: Create example configurations and use cases

### Before You Start

1. **Check Existing Issues**: Look for existing issues related to your contribution
2. **Create an Issue**: For significant changes, create an issue first to discuss the approach
3. **Review Roadmap**: Check our [roadmap](README.md#roadmap) to see planned features
4. **Read Documentation**: Familiarize yourself with the project structure and architecture

## Pull Request Process

### 1. Prepare Your Changes

1. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-number
   ```

2. **Make Your Changes**
   - Follow the coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed
   - Ensure your changes don't break existing functionality

3. **Test Your Changes**
   ```bash
   # Run all tests
   go test ./...
   
   # Run specific tests
   go test ./internal/promptfoo/evaluation/evaluator
   
   # Test CLI functionality
   ./pe eval example/config.yaml
   
   # Run integration tests
   ./scripts/test-integration.sh
   ```

### 2. Commit Your Changes

Use clear, descriptive commit messages:

```bash
# Good commit messages
git commit -m "feat: add streaming support to eval command"
git commit -m "fix: handle timeout errors in provider calls"
git commit -m "docs: add examples for custom metrics"

# Follow conventional commits format when possible
git commit -m "type(scope): description"
```

### 3. Submit Your Pull Request

1. **Push Your Branch**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request**
   - Use the pull request template
   - Provide a clear description of changes
   - Link related issues
   - Add screenshots for UI changes
   - Request review from maintainers

3. **Pull Request Template**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Tests pass locally
   - [ ] Added tests for new functionality
   - [ ] Manual testing completed
   
   ## Checklist
   - [ ] Code follows style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] No breaking changes (or marked as such)
   ```

### 4. Code Review Process

1. **Automated Checks**: CI will run tests and checks
2. **Maintainer Review**: A maintainer will review your code
3. **Address Feedback**: Make requested changes
4. **Approval**: Once approved, your PR will be merged

## Issue Guidelines

### Bug Reports

When reporting bugs, please include:

```markdown
## Bug Description
Clear description of what went wrong

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should have happened

## Actual Behavior
What actually happened

## Environment
- OS: [e.g., macOS 12.6]
- Go version: [e.g., 1.21.0]
- PE version: [e.g., v1.2.3]
- Provider: [e.g., OpenAI GPT-4]

## Configuration
```yaml
# Your configuration file (remove sensitive data)
```

## Error Output
```
Error messages and stack traces
```

## Additional Context
Any other relevant information
```

### Feature Requests

For feature requests, please include:

```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Why is this feature needed? What problem does it solve?

## Proposed Solution
How should this feature work?

## Alternatives Considered
What other approaches have you considered?

## Additional Context
Mock-ups, examples, related issues, etc.
```

## Development Workflow

### Project Structure

```
pe/
├── cmd/pe/                 # Main CLI application
│   ├── main.go            # Entry point
│   ├── commands.go        # Command definitions
│   ├── eval.go            # Evaluation command
│   ├── pipeline.go        # Pipeline commands
│   └── repl.go            # Interactive REPL
├── internal/              # Internal packages
│   ├── promptfoo/         # Promptfoo compatibility
│   │   ├── evaluation/    # Evaluation components
│   │   │   ├── evaluator/ # Core evaluation logic
│   │   │   ├── metrics/   # Custom metrics framework
│   │   │   └── testing/   # Testing utilities
│   │   ├── execution/     # Execution components
│   │   │   ├── consensus/ # Multi-provider consensus
│   │   │   └── distributed/ # Distributed execution
│   │   └── security/      # Security components
│   │       ├── attestation/ # Cryptographic attestation
│   │       └── redteam/   # Red-teaming module
│   ├── llm/               # LLM provider interfaces
│   └── providers/         # Provider implementations
├── docs/                  # Documentation
├── example/               # Example configurations
├── testdata/              # Test data
└── specs/                 # API specifications
```

### Coding Standards

#### Go Style Guidelines

1. **Follow Go conventions**
   ```go
   // Good: Use camelCase for unexported functions
   func evaluatePrompt() error {
       // implementation
   }
   
   // Good: Use PascalCase for exported functions
   func EvaluatePrompt() error {
       // implementation
   }
   ```

2. **Error Handling**
   ```go
   // Good: Wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to evaluate prompt: %w", err)
   }
   
   // Good: Use specific error types when appropriate
   type ValidationError struct {
       Field string
       Value interface{}
   }
   
   func (e ValidationError) Error() string {
       return fmt.Sprintf("validation failed for field %s: %v", e.Field, e.Value)
   }
   ```

3. **Documentation**
   ```go
   // Good: Document exported functions
   // EvaluatePrompt runs a prompt evaluation against the specified provider.
   // It returns the evaluation result or an error if the evaluation fails.
   func EvaluatePrompt(prompt string, provider Provider) (*Result, error) {
       // implementation
   }
   ```

4. **Testing**
   ```go
   // Good: Use table-driven tests
   func TestEvaluatePrompt(t *testing.T) {
       tests := []struct {
           name     string
           prompt   string
           provider Provider
           want     *Result
           wantErr  bool
       }{
           {
               name:     "successful evaluation",
               prompt:   "test prompt",
               provider: &mockProvider{},
               want:     &Result{Score: 1.0},
               wantErr:  false,
           },
           // more test cases...
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               got, err := EvaluatePrompt(tt.prompt, tt.provider)
               if (err != nil) != tt.wantErr {
                   t.Errorf("EvaluatePrompt() error = %v, wantErr %v", err, tt.wantErr)
                   return
               }
               if !reflect.DeepEqual(got, tt.want) {
                   t.Errorf("EvaluatePrompt() = %v, want %v", got, tt.want)
               }
           })
       }
   }
   ```

#### CLI Guidelines

1. **Command Structure**
   ```go
   // Good: Use cobra.Command structure
   cmd := &cobra.Command{
       Use:   "command [args]",
       Short: "Short description",
       Long:  `Long description with examples`,
       Args:  cobra.ExactArgs(1),
       RunE:  runCommand,
   }
   ```

2. **Flag Naming**
   ```go
   // Good: Use consistent flag naming
   cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
   cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show commands without executing")
   ```

3. **Output Formatting**
   ```go
   // Good: Support multiple output formats
   switch outputFormat {
   case "json":
       return json.MarshalIndent(result, "", "  ")
   case "yaml":
       return yaml.Marshal(result)
   case "table":
       return formatTable(result)
   default:
       return fmt.Errorf("unsupported format: %s", outputFormat)
   }
   ```

### Adding New Features

#### 1. New Commands

1. Create command file in `cmd/pe/`
2. Implement cobra.Command
3. Add to main.go
4. Add tests
5. Update documentation

Example:
```go
// cmd/pe/newcommand.go
func newCommandCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "newcommand",
        Short: "Description of new command",
        RunE:  runNewCommand,
    }
    return cmd
}

func runNewCommand(cmd *cobra.Command, args []string) error {
    // implementation
    return nil
}
```

#### 2. New Providers

1. Implement `llm.Provider` interface
2. Add to provider registry
3. Add configuration support
4. Add tests
5. Update documentation

Example:
```go
// internal/providers/newprovider.go
type NewProvider struct {
    apiKey   string
    endpoint string
}

func (p *NewProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
    // implementation
    return &promptfoo.ProviderResponse{}, nil
}
```

#### 3. New Assertion Types

1. Add to assertion package
2. Implement evaluation logic
3. Add to assertion registry
4. Add tests
5. Update documentation

#### 4. New Metrics

1. Add to metrics package
2. Implement metric interface
3. Add configuration support
4. Add tests
5. Update documentation

## Testing

### Test Categories

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **CLI Tests**: Test command-line interface
4. **Provider Tests**: Test LLM provider integrations

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/promptfoo/evaluation/evaluator

# Run tests with race detection
go test -race ./...

# Run integration tests
./scripts/test-integration.sh

# Run benchmarks
go test -bench=. ./...
```

### Writing Tests

1. **Use descriptive test names**
   ```go
   func TestEvaluatePrompt_WithValidInput_ReturnsSuccess(t *testing.T) {
       // test implementation
   }
   ```

2. **Use table-driven tests for multiple scenarios**
3. **Mock external dependencies**
4. **Test error conditions**
5. **Add benchmarks for performance-critical code**

### Test Coverage

- Aim for >80% test coverage
- Focus on critical paths
- Test error handling
- Include edge cases

## Documentation

### Types of Documentation

1. **Code Documentation**: Godoc comments
2. **User Documentation**: README, guides, examples
3. **API Documentation**: OpenAPI specs
4. **Architecture Documentation**: Design decisions

### Documentation Standards

1. **README Updates**: Update for new features
2. **Godoc Comments**: Document exported functions
3. **Examples**: Provide working examples
4. **Changelog**: Document changes

### Writing Documentation

1. **Be Clear and Concise**
2. **Include Examples**
3. **Update Related Docs**
4. **Use Consistent Formatting**

## Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

1. **Update Version**
2. **Update Changelog**
3. **Run Full Test Suite**
4. **Update Documentation**
5. **Create Release Notes**
6. **Tag Release**
7. **Publish Binaries**

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: General discussion, questions
- **Documentation**: Comprehensive guides and examples

### Maintainer Contact

For questions about contributing:

1. Create an issue for bugs or features
2. Start a discussion for general questions
3. Review existing documentation first

## Recognition

Contributors will be recognized in:

- **CONTRIBUTORS.md**: List of contributors
- **Release Notes**: Major contributions
- **Documentation**: Example credits

Thank you for contributing to PE! Your efforts help make prompt engineering more systematic and accessible to everyone.