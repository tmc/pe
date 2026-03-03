# API Reference

PE is primarily a CLI tool, but its core packages are available for Go developers.

> **Note**: The Go API is currently unstable and subject to change.

## Key Packages

*   `github.com/tmc/pe/internal/providers`: LLM provider interfaces.
*   `github.com/tmc/pe/internal/promptfoo/evaluation/evaluator`: Evaluation engine.
*   `github.com/tmc/pe/internal/metaprompt`: Optimization and metaprompting algorithms.
*   `github.com/tmc/pe/internal/module`: Module registry and resolution.

Documentation is available via `go doc`:

```bash
go doc github.com/tmc/pe/internal/providers
```
