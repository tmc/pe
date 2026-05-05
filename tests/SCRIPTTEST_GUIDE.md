# PE Scripttest Suite Guide

This directory contains PE command-line tests using `rsc.io/script/scripttest`.

## Overview

`tests/scripttest_test.go` builds the local `pe` binary and the `pe-promptfoo`
plugin, then runs each `tests/testdata/script/*.txt` file in an isolated work
directory. The script engine uses `scripttest.DefaultCmds()` with `exec`
removed and custom `pe` and `pipe` commands added.

The runner is not a shell. It does not interpret pipes, redirection, heredocs,
`&&`, `||`, or shell job control. A trailing `&` is scripttest background syntax
only for commands registered as async; the custom `pe` command is not async.

## Running Tests

Run all script tests:

```bash
go test -v ./tests/...
```

Run one script:

```bash
go test -v ./tests/... -run TestScripts/init
```

Preserve work directories for debugging:

```bash
go test -v ./tests/... -testwork
```

## Supported Commands

PE tests may use:

- `pe args...`: run the built PE binary.
- `pipe command [args...] | command [args...] ...`: run a restricted pipeline.
  Pipeline stages may use `pe`, `cat`, or `echo`.
- File and environment commands: `cat`, `cd`, `chmod`, `cmp`, `cmpenv`, `cp`, `echo`, `env`, `exists`, `grep`, `mkdir`, `mv`, `replace`, `rm`, `sleep`, `symlink`.
- Assertions and control: `stdout`, `stderr`, `!`, `?`, `[condition]`, `skip`, `stop`, `wait`, `help`.

The runner does not register `exec`, `stdin`, `contains`, `concurrent`,
`count`, `json`, or `yaml`. Use `grep` for file-content checks.

## Files and Output

Use txtar sections to create input files:

```txt
pe run prompt.txt
stdout 'Paris'

-- prompt.txt --
What is the capital of France?
```

Use `pipe` for stdin-only command composition:

```txt
pipe pe run prompt.txt --provider mock | pe extract --tag answer
stdout 'Paris'

-- prompt.txt --
Answer in <answer>...</answer> tags: capital of France?
```

Use PE flags or `cp stdout file` for intermediate files instead of shell
redirection. If a command only accepts standard input and cannot be expressed
with `pipe`, add file-input support or a dedicated script command before testing
it here; do not use `exec sh -c` or `< file`.

## Conditions and Environment

Available conditions come from `scripttest.DefaultConds()`: `GOOS:<value>`,
`GOARCH:<value>`, `compiler:<value>`, `root`, `exec:<program>`, `short`, and
`verbose`.

The runner sets `PE_TEST_MODE=true`, `PE_MOCK_PROVIDER=true`, and a `PATH` that
starts with the directory containing the built test binaries. Scripttest also
sets `$WORK` to the per-test work directory and a platform temp directory.

## Adding Tests

1. Create a focused `.txt` file in `tests/testdata/script/`.
2. Use direct `pe` commands and built-in script commands.
3. Create inputs with txtar file sections.
4. Capture outputs with PE output flags or `cp stdout file`.
5. Check output with `stdout`, `stderr`, `grep`, `exists`, or `cmp`.

Keep scripts sequential unless the command under test exposes concurrency
through PE flags or configuration.

## Executable Text Test Plan

PE is the Go toolchain for safe prompting: executable, templated, composable
text. Plain text should remain a valid artifact by default. Script tests for
that direction should start with the CLI gates that prove files are accepted,
rendered, or rejected before any provider or tool side effect runs.

Initial script-test coverage should exercise:

- plain text files without declarations,
- files with declared template inputs and deterministic rendered output,
- metadata, safety, and placement declarations that select allowed data,
  prompts, providers, and tools,
- denied capability classes that fail before execution,
- parent/child composition where a child artifact cannot loosen parent
  constraints.

Keep policy-composition assertions conservative. A composed artifact should only
gain the intersection of allowed capabilities and the union of denied
capabilities. Use unit tests for parser and policy edge cases when a script
would need shell behavior or runtime features that do not exist yet.
