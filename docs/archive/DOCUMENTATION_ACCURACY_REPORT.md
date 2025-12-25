# Documentation Accuracy Report

## Summary

After reviewing all documentation and scripts, here are the main accuracy issues that need to be addressed:

## 1. README.md Issues

### Unimplemented Commands Mentioned:
- `pe install github.com/user/prompts` - NOT IMPLEMENTED
- `pe build config.yaml` - NOT IMPLEMENTED  
- `pe tool create` - NOT IMPLEMENTED
- `pe mod init/tidy` - NOT IMPLEMENTED
- Many "Go toolchain" parallels that don't exist

### Features Not Implemented:
- **Gist support**: `pe run gist:user/id` - NOT IMPLEMENTED
- **txtar format support** - NOT IMPLEMENTED
- **macOS sandboxing** - NOT IMPLEMENTED
- **Trust management** (`pe trust add/list/remove`) - NOT IMPLEMENTED
- **History management** (`pe history`, `pe session`) - NOT IMPLEMENTED
- **Version control** (`pe branch`, `pe commit`, `pe merge`, etc.) - NOT IMPLEMENTED
- **Style guides and mixins** (`pe import style`, `pe behavior`) - NOT IMPLEMENTED
- **Shared caching** (`pe cache`) - NOT IMPLEMENTED
- **Remote operations** (`pe push/pull/clone`) - NOT IMPLEMENTED

### Inaccurate Examples:
- Most examples in the "Core Commands" section reference unimplemented features
- The "Plugin System for Custom Inference" section describes features that don't exist

## 2. CLAUDE.md Issues

Similar to README.md, CLAUDE.md references many unimplemented features:
- txtar format support
- Security features (sandboxing, trust)
- History and session management
- Version control features
- Lineage and dependency tracking

## 3. Documentation Files Accuracy

### Accurate Documentation:
- `docs/CLI_REFERENCE.md` - Mostly accurate for implemented commands
- `docs/ADVANCED_OPTIMIZATION_GUIDE.md` - Accurately describes optimization methods
- `docs/API_REFERENCE.md` - Accurate for existing APIs

### Needs Updates:
- `docs/QUICK_START.md` - References some unimplemented features
- `docs/GETTING_STARTED.md` - May reference unimplemented commands
- `docs/WORLD_CLASS_*.md` - May overstate current capabilities

## 4. Example Scripts

### Accurate:
- `examples/run-command.sh` - Accurately shows `pe run` usage
- `examples/inference/basic.go` - Correct inference API usage

### Needs Review:
- Any examples showing unimplemented commands

## 5. Test Documentation

The test documentation (`tests/README.md`) appears to describe a test structure that exists, though the actual test implementation uses mocks.

## Recommendations

1. **Update README.md** and **CLAUDE.md** to reflect only implemented features
2. **Create a ROADMAP.md** file to document planned but unimplemented features
3. **Update examples** to only show working commands
4. **Add clear "NOT YET IMPLEMENTED" markers** for planned features
5. **Focus documentation on actual capabilities**:
   - Core evaluation (`pe eval`)
   - Optimization (`pe optimize`, `pe semantic`)
   - Metrics and analysis
   - Plugin system
   - Inference API with cgpt provider

## Implemented Commands (Accurate List)

```bash
# Core Commands
pe run                 # Execute prompts with inference API
pe eval                # Evaluate configurations
pe test                # Run comprehensive tests
pe init                # Initialize configuration

# Optimization
pe optimize            # Various optimization methods
pe semantic            # Semantic backpropagation/GASO
pe compose             # Component composition

# Analysis
pe metrics             # Advanced metrics calculation
pe benchmark           # Performance benchmarking
pe analyze             # Statistical analysis
pe stats               # Quick statistics

# Utilities
pe fmt                 # Format configurations
pe vet                 # Validate configurations
pe convert             # Convert between formats
pe watch               # Watch and re-run
pe view                # View results in browser

# Pipeline Commands
pe ask                 # Interactive prompts
pe stream              # Stream processing
pe filter              # Filter results
pe diff                # Compare results
pe interactive         # REPL mode

# Other
pe playground          # Interactive playground
pe template            # Template management
pe profile             # Performance profiling
pe security            # Security testing

# Plugin System
pe plugin list         # List plugins
pe plugin run          # Run plugin commands
pe promptfoo           # Promptfoo compatibility (via plugin)
```