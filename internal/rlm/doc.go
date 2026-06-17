// Package rlm implements a bounded, local recursive language-model runner.
//
// The runner processes large inputs through typed, bounded operations rather
// than arbitrary model-generated code. Large inputs stay out of the prompt and
// are referenced by content keys. Workers inspect bounded chunks. A run maps a
// prompt over the chunks, then folds the outputs through successive reduce
// passes while depth and budget remain, so a wide input collapses
// hierarchically toward a single answer. Child results are aggregated with
// deterministic rules. Every run emits a strict JSON trace (schema
// pe.rlm.trace.v1) that tests and release gates can parse.
//
// The package is local and provider-agnostic. It calls a [Worker], not a
// concrete provider, so runs are testable without live providers. Budgets for
// depth, workers, token use, and selected chunks are enforced before any worker
// call is made.
//
// See docs/future/RLM_DESIGN.md for the full design and safety model.
package rlm
