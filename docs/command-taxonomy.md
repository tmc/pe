# Command Taxonomy

This inventory was generated from `GOTOOLCHAIN=go1.25.9 go run ./cmd/pe
--help`.

## Current Commands

```text
analyze
ask
benchmark
build
cat
collect
completion
convert
diff
doc
edit
eval
eval-prompt
exp
expand
experimental
extract
filter
fmt
get
help
init
interactive
mod
plugin
profile
prompt
push
reduce
run
run-text
security
serve
stats
stream
template
test
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
doc
init
work
serve
version
```

Evaluation commands inspect prompt behavior and result files:

```text
eval
eval-prompt
benchmark
diff
stats
view
analyze
```

Optimization commands change prompts or search prompt space:

```text
experimental optimize
experimental evolve
experimental semantic
exp optimize
```

Module commands manage prompt packages and registries:

```text
mod
push
get
edit
prompt
```

Pipeline commands support Unix-style text flow:

```text
ask
cat
stream
filter
extract
expand
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
completion
plugin
```

Experimental commands carry higher-change-risk features:

```text
exp
experimental
security
profile
```

## Relationships

`run`, `run-text`, `cat`, and `ask` form the executable-text path: render or
execute text, optionally with variables. `mod`, `push`, and `get` form the
module path. `eval`, `benchmark`, `diff`, `stats`, and `view` form the result
analysis path. `stream`, `filter`, `extract`, `collect`, and `reduce` form the
pipeline path.

The reorganization should preserve existing command names as compatibility
shims. Grouping should improve help and discovery without making existing
scripts change command names.
