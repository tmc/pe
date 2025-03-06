# PE: Prompt Engineering Toolkit

## Build & Test Commands
```bash
# Build the project
go build ./cmd/pe

# Run all tests
go test ./...

# Run specific package tests
go test ./internal/assertutil
go test ./internal/template

# Run with verbose output and specific test
go test -v ./internal/template -run TestRenderTemplate

# Run benchmarks
./pe benchmark example/benchmark-config.yaml --iterations 5
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

Follow Go best practices and maintain consistency with the existing codebase.