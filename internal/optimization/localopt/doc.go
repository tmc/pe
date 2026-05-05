// Package localopt provides deterministic local prompt optimization helpers.
//
// The package is intentionally provider-free: callers supply candidate
// refinements and a scorer, and the optimizer keeps the best-scoring variant.
// It is useful for testing semantic backpropagation and GASO-style refinement
// flows without making model calls.
package localopt
