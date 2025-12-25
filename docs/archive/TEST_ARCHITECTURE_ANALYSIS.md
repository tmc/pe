# Test Architecture Analysis and Recommendations

## Current State Analysis

### 1. Code Organization Issues

#### cmd/pe/ Directory (50+ files, ~1600 items)
**Problems:**
- Monolithic structure with 40+ command files in single directory
- Each command is 300-1000+ lines
- Mixed concerns: commands, helpers, types all in same package
- Duplicate functions across files (extractDefaults, extractDescription, etc.)

#### internal/ Directory
**Good:**
- Well-organized with domain-specific packages
- Clear separation of concerns

**Issues:**
- Some packages lack corresponding tests
- No clear boundary between public API and implementation details

### 2. Test Organization Issues

#### Current Test Structure:
```
tests/
├── scripttest/
│   ├── basic/
│   │   ├── *.txt        # Old format tests
│   │   └── *.txtar      # New format tests (duplicates!)
│   ├── advanced/
│   └── benchmarks/
├── testdata/            # Underutilized
├── benchmarks/
└── *.go                 # Top-level test files
```

**Problems:**
1. **Duplicate test files** (.txt and .txtar versions)
2. **No testdata organization** - test data mixed with test logic
3. **Inconsistent naming** - some numbered, some named by feature
4. **No clear test categories** - unit vs integration vs e2e mixed

## Proposed Reorganization

### 1. Refactor cmd/pe/ Structure

```
cmd/pe/
├── main.go              # Entry point only
├── commands/            # Command implementations
│   ├── run/
│   │   ├── run.go       # Command definition
│   │   ├── run_test.go  # Unit tests
│   │   └── testdata/    # Test fixtures
│   ├── prompt/
│   │   ├── init.go
│   │   ├── help.go
│   │   ├── edit.go
│   │   ├── info.go
│   │   ├── tidy.go
│   │   └── testdata/
│   ├── module/
│   │   ├── mod.go
│   │   ├── download.go
│   │   └── testdata/
│   └── ...
├── internal/            # Shared helpers (cmd-specific)
│   ├── flags/          # Flag parsing helpers
│   ├── output/         # Output formatting
│   └── validation/     # Input validation
└── doc.go              # Package documentation
```

### 2. Reorganize Test Structure

```
tests/
├── unit/               # Unit tests (if not colocated)
├── integration/        # Integration tests
│   └── testdata/
│       ├── prompts/    # Test prompt files
│       ├── modules/    # Test modules
│       └── expected/   # Expected outputs
├── e2e/                # End-to-end scripttests
│   ├── testdata/
│   │   ├── 01_init_and_setup.txtar
│   │   ├── 02_basic_operations.txtar
│   │   ├── 03_prompt_management.txtar
│   │   ├── 04_variable_flags.txtar
│   │   ├── 05_shebang_execution.txtar
│   │   ├── 06_streaming.txtar
│   │   ├── 07_modules.txtar
│   │   ├── 08_optimization.txtar
│   │   └── 09_advanced_features.txtar
│   └── runner_test.go  # Test runner
├── benchmarks/
│   ├── testdata/
│   └── *_bench_test.go
└── README.md           # Test documentation
```

### 3. Extract Shared Code to internal/

Move duplicated functionality to internal packages:

```
internal/
├── prompt/
│   ├── parser.go       # Prompt file parsing
│   ├── parser_test.go
│   ├── format.go       # Already exists
│   ├── metadata.go     # Extract defaults, description, etc.
│   └── testdata/
├── template/
│   ├── variables.go    # Variable extraction
│   ├── render.go       # Template rendering
│   └── testdata/
└── cli/
    ├── flags.go        # Custom flag parsing
    └── help.go         # Help generation
```

## Implementation Plan

### Phase 1: Extract Shared Code
1. Create `internal/prompt/metadata.go` with:
   - `ExtractDefaults()`
   - `ExtractDescription()`
   - `ExtractVariables()`
   - `ExtractSystemPrompt()`

2. Create `internal/cli/flags.go` with:
   - `ParseVariableFlags()`
   - `SetupDynamicFlags()`

3. Update all commands to use shared functions

### Phase 2: Reorganize Commands
1. Create `cmd/pe/commands/` subdirectories
2. Move each command group to its subdirectory:
   - `prompt/*` - prompt management commands
   - `run/*` - execution commands
   - `module/*` - module commands
   - `eval/*` - evaluation commands
   - `optimize/*` - optimization commands

3. Keep related commands together

### Phase 3: Consolidate Tests
1. Move to single test format (`.txtar`)
2. Organize by functionality in `testdata/`
3. Create test helpers for common operations
4. Add test categories:
   - `testdata/unit/` - Small, focused tests
   - `testdata/integration/` - Feature tests
   - `testdata/e2e/` - Full workflow tests

### Phase 4: Documentation
1. Add README.md in each test directory
2. Document test naming conventions
3. Create test writing guidelines

## Benefits

1. **Better Maintainability**
   - Smaller, focused files
   - Clear responsibilities
   - Reduced duplication

2. **Improved Testing**
   - Clear test organization
   - Reusable test data
   - Better test coverage visibility

3. **Easier Navigation**
   - Logical grouping
   - Consistent structure
   - Clear dependencies

4. **Scalability**
   - Easy to add new commands
   - Clear patterns to follow
   - Modular architecture

## Test Naming Convention

Propose standardized naming:

```
# E2E Tests (testdata/e2e/)
01_feature_name.txtar       # Numbered for execution order

# Integration Tests (testdata/integration/)
feature_name_test.txtar      # Feature-based naming

# Test Fixtures (testdata/fixtures/)
simple_prompt.prompt         # Descriptive names
complex_template.prompt
invalid_syntax.prompt
```

## Metrics to Track

After reorganization, track:
- Lines per file (target: <500)
- Test coverage per package (target: >70%)
- Duplication percentage (target: <5%)
- Build time
- Test execution time