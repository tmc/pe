# Performance Profile

Generated on 2026-05-05 from `exp` on Apple M4 Max.

Critical paths profiled in this pass:

- `internal/security` input sanitization and validation.
- `cmd/pe` command execution, template substitution, and file reading.
- `internal/observability` tracing and profiling helpers.
- `internal/testing` rate limiter and retry helpers.

Focused sanitizer benchmark before the allocation change:

```text
BenchmarkSanitizePrompt-16       3  3681 ns/op  250 B/op  5 allocs/op
BenchmarkValidateFilePath-16     3   347 ns/op    0 B/op  0 allocs/op
BenchmarkValidateAPIKey-16       3   167 ns/op    0 B/op  0 allocs/op
```

The sanitizer hot loop now pre-grows its output builder to the cleaned input
length. This keeps behavior unchanged while reducing growth reallocations for
larger prompts.

Focused sanitizer benchmark after the allocation change:

```text
BenchmarkSanitizePrompt-16       3  7319 ns/op   85 B/op  1 allocs/op
BenchmarkValidateFilePath-16     3   153 ns/op    0 B/op  0 allocs/op
BenchmarkValidateAPIKey-16       3    69 ns/op    0 B/op  0 allocs/op
```

Re-run with:

```sh
GOTOOLCHAIN=go1.25.9 go test -run '^$' -bench 'Benchmark(SanitizePrompt|ValidateFilePath|ValidateAPIKey)' -benchmem ./internal/security
make bench
```
