# Contributing

We welcome contributions to PE!

## Workflow

1.  **Fork** the repository.
2.  **Clone** your fork.
3.  **Branch** for your feature (`git checkout -b feature/amazing-feature`).
4.  **Implement** changes.
5.  **Test**: Run `go test ./...`.
6.  **Commit** and push.
7.  **PR**: Open a Pull Request against `tmc/pe`.

## Development Setup

```bash
git clone https://github.com/tmc/pe
cd pe
go install ./cmd/pe
```

## Standards

*   Follow Go idioms (Effective Go).
*   Ensure all new features have test coverage.
*   Update documentation if CLI changes.
