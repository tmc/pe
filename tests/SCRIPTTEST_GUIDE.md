# PE Scripttest Suite Guide

This directory contains a comprehensive scripttest suite for PE (Go for Prompts), modeled after the Go toolchain's testing approach.

## Overview

The test suite uses `rsc.io/script/scripttest` to provide declarative, readable tests for PE's command-line interface. Tests are written in a txtar-like format that combines shell commands, expected output, and file content.

## Test Organization

```
tests/
├── scripttest_test.go      # Main test runner with mock PE implementation
├── testdata/
│   └── script/            # Test files (*.txt)
│       ├── README         # Detailed test format documentation
│       ├── basic_smoke.txt # Framework verification test
│       ├── init.txt       # pe init command tests
│       ├── run.txt        # pe run command tests
│       ├── test.txt       # pe test command tests
│       ├── optimize.txt   # pe optimize command tests
│       ├── compose.txt    # pe compose command tests
│       ├── mod.txt        # pe mod (dependency management) tests
│       ├── build.txt      # pe build command tests
│       ├── semantic.txt   # pe semantic (2025 GASO/backprop) tests
│       ├── extract.txt    # pe extract (XML tag extraction) tests
│       ├── metrics.txt    # pe metrics command tests
│       ├── pipeline.txt   # Unix pipeline tests
│       ├── version_control.txt # Git-like version control tests
│       ├── security.txt   # Security and sandboxing tests
│       ├── cache.txt      # Verifiable caching tests
│       ├── plugin.txt     # Plugin system tests
│       ├── fmt.txt        # pe fmt command tests
│       └── workflow.txt   # End-to-end workflow tests
```

## Running Tests

### Run all tests:
```bash
go test -v ./tests/...
```

### Run specific test file:
```bash
go test -v ./tests/... -run TestScripts/init
```

### Debug mode (preserve work directories):
```bash
go test -v ./tests/... -testwork
```

### Update test expectations:
```bash
go test -v ./tests/... -update
```

## Test Format

Tests use the scripttest format with these key commands:

### Basic Commands
- `pe <command>` - Run PE commands
- `exec <cmd>` - Run shell commands
- `stdin <content>` - Provide stdin input
- `env KEY=value` - Set environment variables
- `cd <dir>` - Change directory

### Assertions
- `stdout <pattern>` - Check stdout contains pattern
- `stderr <pattern>` - Check stderr contains pattern
- `! stdout <pattern>` - Check stdout doesn't contain pattern
- `exists <file>` - Check file exists
- `! exists <file>` - Check file doesn't exist
- `contains <file> <pattern>` - Check file contains pattern
- `cmp <file1> <file2>` - Compare files

### File Creation
Files are created using the `--` marker:

```
-- prompt.txt --
You are a helpful assistant.

-- config.yaml --
provider: openai
model: gpt-4
```

## Example Test

```txt
# Test pe run with variables

# Create a template prompt
pe run 'Translate {{.Text}} to {{.Language}}' --var Text=Hello --var Language=Spanish
stdout 'Hola'

# Run from file
pe run prompt.txt
stdout 'assistant'
contains prompt.txt 'helpful'

-- prompt.txt --
You are a helpful assistant. Please help the user.
```

## Mock Implementation

The test suite includes a comprehensive mock PE implementation in `scripttest_test.go` that simulates:

- All major PE commands (run, test, optimize, build, etc.)
- File creation and management
- Realistic output formatting
- Error conditions

## Adding New Tests

1. Create a new `.txt` file in `testdata/script/`
2. Follow the scripttest format
3. Add mock implementations for new commands in `scripttest_test.go`
4. Run tests to verify

## Coverage

The test suite covers:

- **Core Commands**: init, run, test, build, optimize
- **Advanced Features**: semantic backprop, GASO, metaprompting
- **Toolchain Features**: mod, fmt, version control
- **Pipeline Operations**: Unix-style composition
- **Security**: Sandboxing, trust management
- **Infrastructure**: Caching, plugins, metrics

## Best Practices

1. **Test Isolation**: Each test runs in a fresh directory
2. **Clear Names**: Test files named after functionality
3. **Comprehensive Checks**: Test both success and failure cases
4. **Realistic Mocks**: Mock outputs match real PE behavior
5. **Documentation**: Comment tests to explain purpose

## Integration with CI

Add to GitHub Actions:

```yaml
- name: Run Scripttests
  run: go test -v ./tests/...
```

## Comparison with Go Toolchain Tests

This test suite follows the same patterns as the Go toolchain's `cmd/go/testdata/script/` tests:

- Same scripttest framework
- Similar command structure
- Comprehensive coverage
- Declarative test format
- Easy debugging with `-testwork`

The main difference is that PE tests focus on prompt engineering workflows rather than Go compilation.