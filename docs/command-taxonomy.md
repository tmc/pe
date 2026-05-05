# Command Taxonomy

This inventory was generated from `GOTOOLCHAIN=go1.25.9 go run ./cmd/pe
--help` on 2026-05-05.

## Current Commands

```text
analyze
ask
benchmark
build
cat
collect
compose
config
completion
convert
diff
doc
edit
eval
eval-prompt
evolve
exp
expand
experimental
extract
filter
fmt
fusion
gaso
get
help
init
interactive
mod
optimize
pe2
plugin
profile
prompt
push
reduce
run
run-text
security
semantic
serve
stats
stream
template
test
textgrad
version
vet
view
watch
work
```

## Proposed Groups

Core commands are the go-toolchain-shaped surface:

```text
run
run-text
build
test
ask
doc
init
prompt
edit
work
serve
```

Evaluation commands inspect prompt behavior and result files:

```text
eval
eval-prompt
benchmark
diff
stats
view
```

Optimization commands change prompts or search prompt space:

```text
optimize
evolve
semantic
textgrad
pe2
gaso
fusion
```

Module commands manage prompt packages and registries:

```text
mod
push
get
```

Pipeline commands support Unix-style text flow:

```text
cat
stream
filter
analyze
extract
expand
compose
collect
reduce
```

Utility commands format, validate, convert, and template files:

```text
fmt
vet
convert
template
interactive
watch
config
version
```

Experimental commands carry higher-change-risk features:

```text
exp
experimental
security
profile
```

Plugin commands manage extension discovery and lifecycle:

```text
plugin
```

## Relationships

`run`, `run-text`, `cat`, and `ask` form the executable-text path: render or
execute text, optionally with variables. `mod`, `push`, and `get` form the
module path. `eval`, `benchmark`, `diff`, `stats`, and `view` form the result
analysis path. `stream`, `filter`, `analyze`, `extract`, `compose`, `collect`,
and `reduce` form the pipeline path. `optimize`, `evolve`, `semantic`,
`textgrad`, `pe2`, `gaso`, and `fusion` form the optimization path.

The reorganization should preserve existing command names as compatibility
shims. Grouping should improve help and discovery without making existing
scripts change command names.
