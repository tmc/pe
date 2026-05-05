# Legacy `example/` Directory

The sibling [`../example/`](../example/) directory is legacy development
material. It is useful for historical reference, but it is not part of the
release-facing examples set.

Use this `examples/` tree for examples that are expected to match the current
CLI. Current examples should be small, runnable, and validated with local
commands where possible.

Known legacy characteristics:

- Some demos document prototype command surfaces that moved under
  `pe experimental` or `pe exp`.
- Some promptfoo-style configs use shapes that are broader than the current
  evaluator accepts.
- Some security and attestation demos describe future capabilities.

Do not treat `../example/` as release validation evidence until a specific demo
has been refreshed and moved or linked from this `examples/` tree.
