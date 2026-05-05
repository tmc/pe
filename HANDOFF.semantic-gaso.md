# Semantic/GASO Handoff

This branch adds a provider-free local optimization primitive under
`internal/optimization/localopt`.

The package is intentionally small:

- callers provide prompt variants through a `Refiner`
- callers score variants through a deterministic `Scorer`
- the optimizer records the trajectory and accepts only improving variants

This gives semantic backpropagation and GASO work a local test harness for
candidate refinement without model calls. A next slice can adapt existing
semantic or GASO gradients into `localopt.Refiner` implementations, then expose
that path through experimental commands once the behavior is useful.
