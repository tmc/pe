# Migration Guide

## v0.5.0

No required migration steps are known for v0.5.0.

PE is still pre-1.0, so command and provider interfaces may continue to change
in later minor releases. For v0.5.0, the release work keeps existing prompt,
evaluation, module, provider, and experimental command behavior compatible with
the current `exp` branch.

Recommended validation before upgrading a local workflow:

```bash
pe version
pe --help
pe run-text --help
pe mod vet --help
GOTOOLCHAIN=go1.25.9 go test ./...
```

If a workflow depends on live providers, rerun one representative provider call
with local credentials before adopting the release in automation.
