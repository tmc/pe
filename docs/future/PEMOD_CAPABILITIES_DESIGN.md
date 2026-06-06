<!-- Historical draft: archived planning material, not current product documentation. Claims, metrics, and command examples in this file may be stale or aspirational. -->

# pe.mod Capabilities

`pe.mod` should become the module-level capability and placement contract for
safe prompting.

This is a v0.6+ design. It is not a v0.5 release blocker.

## Current State

`pe.mod` already covers module identity, PE version, dependencies, replacements,
exclusions, retractions, trust, signing, registry configuration, and a basic
security policy model.

That is not yet a capabilities design.

The missing layer is a module-wide contract for executable text:

- which data classes may flow through the module,
- which prompt provenance classes may compose with it,
- which providers and tools are allowed or denied,
- where execution may happen,
- whether typed input/output metadata is required,
- how parent and dependency policies compose.

## Role Of pe.mod

`pe.mod` is the module boundary. It should describe what the module is allowed
to do and what it requires from dependencies.

It should not describe every prompt invocation. Per-file front matter owns local
prompt metadata, template inputs, prompt-specific output schemas, and shebang
runner choices. Trace artifacts own runtime facts.

`pe.mod` answers questions like:

- Can any prompt in this module use a remote provider?
- Can prompts read repo-internal files?
- Can generated prompts compose with reviewed prompts?
- Can dependencies request network access?
- Must executable text declare typed inputs and outputs?
- Where may this module run: local, isolated, remote, or dry-run only?

## Design Principles

Keep the syntax Go-like and boring.

Prefer allow/deny lists over programmable policy.

Prefer static validation before runtime enforcement.

Policy composition is conservative: a dependency cannot expand the parent's
capabilities. Effective permissions are the intersection of allows and the union
of denials.

## Proposed Syntax

```go
module github.com/acme/release-prompts

pe 1

require (
    github.com/tmc/pe-stdlib v0.1.0
)

capability {
    data allow repo docs public
    data deny secrets credentials

    prompts allow local reviewed
    prompts deny generated untrusted

    providers allow local test
    providers deny remote

    tools allow read search verify write
    tools deny shell network
}

placement {
    run local
    workspace isolated
    network false

    data-class public => providers local test remote
    data-class repo-internal => providers local test
    data-class secret-adjacent => providers local
}

policy {
    composition strict
    require-typed-io true
    require-reviewed-imports true
}
```

## Capability Block

The `capability` block declares module-wide allowed and denied classes.

Suggested dimensions:

- `data`: public, repo, docs, repo-internal, private, secret-adjacent, secrets,
  credentials.
- `prompts`: local, reviewed, generated, external, untrusted.
- `providers`: local, test, remote, named provider aliases.
- `tools`: read, search, verify, write, shell, network, cache, attest.

Deny wins over allow.

## Placement Block

The `placement` block declares where the module may run and where data classes
may be routed.

Suggested fields:

- `run`: local, isolated, remote, dry-run.
- `workspace`: current, isolated, temp, read-only.
- `network`: true or false.
- `data-class X => providers ...`: provider routing by data class.

Placement should be checked before provider execution, shell execution, file
writes, network access, and dependency execution.

## Policy Block

The `policy` block declares validation behavior.

Suggested fields:

- `composition strict`: dependencies must not require capabilities denied by
  the parent.
- `require-typed-io true`: executable text must declare input/output metadata.
- `require-reviewed-imports true`: composed prompt imports must be reviewed or
  locally trusted.

## Conservative Composition

When module A depends on module B:

- Effective allows are the intersection of A and B.
- Effective denials are the union of A and B.
- B cannot use a provider, tool, prompt provenance, or data class that A denies.
- If B requires a denied capability and A uses `composition strict`, validation
  fails.

This makes dependency composition safe without needing a policy language.

## Boundaries

Keep these in `pe.mod`:

- module-wide capability classes,
- placement policy,
- dependency policy,
- trust/signing requirements,
- typed-IO requirements.

Keep these in per-file front matter:

- prompt-specific template inputs,
- prompt-specific output schemas,
- shebang/runner choice,
- local metadata and labels,
- prompt-specific budgets.

Keep these in trace artifacts:

- actual providers used,
- actual tokens and latency,
- cache hits and misses,
- snippets read,
- child calls,
- dynamic consensus outcomes,
- termination reason.

## Versioned Roadmap

### v0.5

Documentation only. Do not block release.

- Document `pe.mod` capabilities as a future design.
- Keep current parser/runtime behavior unchanged.

### v0.6 Parser And Schema

- Add `Capability`, `Placement`, and `Policy` structs to `internal/pemod`.
- Parse and format the new blocks with round-trip tests.
- Add examples under docs and testdata.

### v0.6 Static Validation

- Add `pe mod vet` or extend `pe mod tidy --json` to report capability issues.
- Validate dependency policy composition.
- Validate typed-IO requirements against executable text metadata.

### v0.7 Runtime Enforcement

- Gate provider selection through effective module capabilities.
- Gate tool/network/file-write access through effective placement policy.
- Record effective policy in trace artifacts.

## First Implementation Slice

The first code slice should be parser-only:

1. Add AST structs.
2. Parse single-line and block forms.
3. Format back to Go-like syntax.
4. Add table-driven tests.
5. Do not enforce at runtime yet.
