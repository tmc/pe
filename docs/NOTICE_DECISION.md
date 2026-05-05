# NOTICE Decision

No separate `NOTICE` file is required for v0.5.0 based on the current dependency
and repository state.

Rationale:

- The repository now includes the MIT `LICENSE` file claimed by README.md.
- `go list -m all` is the dependency inventory source for release review.
- No tracked third-party source bundle in the release path has been identified
  that requires carrying an additional NOTICE file.

Before tagging a later release, rerun dependency and embedded-asset inventory if
new generated code, vendored sources, binary assets, or copied reference material
is added to the release artifacts.
