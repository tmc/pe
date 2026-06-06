<!-- Historical draft: archived planning material, not current product documentation. Claims, metrics, and command examples in this file may be stale or aspirational. -->

# Executable Text

PE is the Go toolchain for safe prompting. Its primary artifact is executable,
templated, composable text.

Text should become executable when it declares enough structure for PE to lower
it into a bounded, testable plan. The plan is made of known PE operators, not
arbitrary model-generated code.

This is a v0.6+ design direction. It is not a v0.5 release blocker.

## Thesis

PE should make prompting feel like Go tooling: simple files, explicit inputs,
fast validation, reproducible execution, and boring artifacts.

The baseline should be as simple as a shebang line. A text file can be run
because it says how PE should run it. More metadata is optional and layered in
only when the artifact needs templates, policy, composition, or side effects.

The core pipeline scales with the file:

```text
plain text + shebang -> runnable prompt
text + inputs -> rendered prompt
text + metadata -> safe placement decision
text + imports -> composed execution graph
graph + budgets -> outputs + trace + attestable artifacts
```

PE should be the `go test` / `go run` / `go vet` experience for prompts:
small commands, explicit contracts, safe defaults, deterministic traces, and
release-gated smokes.

## Non-Goals

- No broad DSL before the schema is proven.
- No arbitrary model-generated code execution.
- No hidden shell execution.
- No provider rewrites for v0.5.
- No runtime dependency on external workflow, optimizer, schema, or recursion systems.

## Executable Text File Shape

A PE text file can be plain text. That is the baseline.

Plain text becomes useful immediately as a prompt, instruction, note, reducer,
verifier, or workflow description. It does not need front matter until it needs
inputs, outputs, budgets, imports, or side effects.

Executable text should be as simple as a shebang line:

```text
#!/usr/bin/env pe run-text
Summarize the release blockers in this repository.

Focus on commands that fail, stale docs, and examples that no longer run.
```

The shebang selects the PE runner. The body stays text.

A text file may carry metadata and opt into templating by declaring inputs.
Template variables are ordinary data bindings, not execution rights. Metadata
lets the file describe ownership, labels, data classification, allowed providers,
and placement policy before any prompt is rendered or executed.

Front matter is the equivalent of imports, build tags, and package docs: use it
when the file needs more structure, not for every prompt.

```markdown
---
kind: pe.text.v1
run: pe run-text
name: release-blocker-review
metadata:
  owner: release
  labels: [audit, docs]
  data_class: repo-internal
inputs:
  repo:
    type: path
  topic:
    type: string
safety:
  data:
    allow: [repo, docs]
    deny: [secrets, credentials]
  prompts:
    allow: [local, reviewed]
  providers:
    allow: [local, test]
    deny: [remote]
  tools:
    allow: [read, search]
    deny: [shell, network, write]
placement:
  run: local
  network: false
---

Research {{ .topic }} in {{ .repo }} and return verified findings.
```

A text file becomes an executable workflow only when it declares enough contract
for PE to validate a plan before running it.

```markdown
---
kind: pe.workflow.v1
name: release-doc-audit
inputs:
  repo:
    type: path
  topic:
    type: string
outputs:
  report:
    type: file
budget:
  max_calls: 20
  max_depth: 2
  max_tokens: 50000
imports:
  reviewer: ./prompts/reviewer.prompt
  verifier: ./workflows/verify-claims.pe.md
safety:
  data:
    allow: [repo, docs]
    deny: [secrets, credentials]
  prompts:
    allow: [local, reviewed]
  providers:
    allow: [local, test]
  tools:
    allow: [read, search, verify, write]
    deny: [network]
placement:
  run: local
  workspace: isolated
  network: false
---

# Goal

Audit {{ .topic }} in {{ .repo }} and produce a verified report.

# Steps

1. files = search(repo, topic)
2. chunks = chunk(files, max_tokens=4000)
3. findings = map(chunks, reviewer)
4. verified = map(findings, verifier)
5. report = reduce(verified, ./prompts/report-writer.prompt)

# Ensures

- write report to outputs.report
- emit trace
```

Execution is opt-in. A file is executable only if PE can validate its operators,
inputs, outputs, budgets, and side effects before running it.

## Template Model

Templating is optional. Plain text remains valid text.

When a file declares inputs, PE can render template variables from explicit
values supplied by flags, environment, stdin, parameter files, or upstream graph
outputs. Templates bind values into text contracts. They do not grant execution
rights. After rendering, PE validates the resulting contract exactly as if it
had been written by hand.

Rules:

- Use Go template semantics initially.
- Treat missing declared inputs as validation errors.
- Allow undeclared plain text files to run as static prompts or static workflow
  descriptions when no variables are present.
- Record the input values in the trace, redacting secrets.
- Preserve the rendered contract as an artifact.


## Metadata And Safety Policy

Text artifacts can declare metadata without becoming executable. Metadata is how
PE decides where text may run, which prompts may compose with it, and which data
may flow through it.

The policy is part of the artifact, so it can be reviewed, tested, cached, and
attested with the text.

Suggested fields:

- `metadata.owner`: human or team responsible for the artifact.
- `metadata.labels`: search and routing labels.
- `metadata.data_class`: public, repo-internal, private, secret-adjacent.
- `safety.data.allow` / `safety.data.deny`: classes of data that may be read or
  passed into the artifact.
- `safety.prompts.allow` / `safety.prompts.deny`: prompt provenance rules, such
  as local, reviewed, generated, external, or untrusted.
- `safety.providers.allow` / `safety.providers.deny`: provider placement rules,
  such as local, test, remote, named provider, or offline only.
- `safety.tools.allow` / `safety.tools.deny`: operator/tool capabilities.
- `placement.run`: local, isolated, remote, or dry-run only.
- `placement.network`: whether network access is allowed.
- `placement.workspace`: current, isolated, temp, or read-only.

Policy composition should be conservative. When text artifacts compose, the
combined graph gets the intersection of allowed capabilities and the union of
denials. A child artifact cannot loosen its parent's safety policy.

Examples:

- A public prompt may run on a remote provider if it allows remote providers.
- A repo-internal audit may run only locally unless it explicitly permits remote
  provider use.
- Secret-adjacent data cannot flow into a generated or unreviewed prompt.
- A workflow that denies `write` can still produce a report proposal in the
  trace, but PE must not write it to disk.

## Composition Model

Executable text composes other text artifacts:

- prompts
- workflows
- schemas
- examples
- evals
- reducers
- verifiers
- output contracts

Composition is explicit through imports and bindings. PE should avoid implicit
ambient behavior.

The minimal composition graph is:

```text
node: text artifact + operator + typed inputs + typed outputs
edge: output binding from one node to another
run: graph + budgets + workspace -> trace + artifacts
```

## Operator Standard Library

Keep the first operator set small:

- `template`: render a text contract with inputs.
- `include`: include static text.
- `use`: bind another text artifact by name.
- `search`: find files or text ranges.
- `read`: read bounded file ranges.
- `chunk`: split text into bounded spans.
- `map`: run a bounded operator over items.
- `reduce`: combine items into one result.
- `verify`: check claims against the filesystem or command output.
- `consensus`: aggregate competing results deterministically.
- `cache`: store external context or intermediate artifacts.
- `attest`: bind artifacts to local integrity manifests.
- `gate`: run a command/test and require pass/fail status.
- `write`: write declared outputs.
- `trace`: emit the run trace.

Operators must have explicit input/output contracts and budget accounting.

## Recursive Harness

Recursive long-context execution is one strategy for executable text.
It should not become a general code-generation loop.

A PE-native recursive harness should:

- Store large context outside the prompt.
- Pass cache/attest pointers into bounded workers.
- Use `chunk`, `map`, `reduce`, `search`, `recurse`, and `consensus`.
- Enforce maximum depth, calls, tokens, bytes read, and wall time.
- Emit every child call and snippet read into the trace.

Future command sketch:

```bash
pe exp recurse input.txt \
  --goal "find release blockers" \
  --max-depth 2 \
  --max-calls 20 \
  --trace trace.json
```

Future package sketch:

```text
internal/rlm
  contract.go     executable text contract structs
  trace.go        JSON trace schema
  chunk.go        bounded range model
  runner.go       RunLocal-backed execution
  aggregate.go    Majority-backed aggregation
  storage.go      cache/attest pointer helpers
```

## Trace Schema

Every executable text run should produce a trace. The trace is the receipt that
makes execution reviewable and replayable.

Minimum fields:

```json
{
  "kind": "pe.trace.v1",
  "workflow": "release-doc-audit",
  "inputs": [],
  "artifacts": [],
  "calls": [],
  "snippets_read": [],
  "children": [],
  "budgets": {},
  "costs": {},
  "aggregation_rule": "majority",
  "termination_reason": "completed"
}
```

The trace should include enough information to answer:

- What text was executed?
- Which inputs were bound?
- Which artifacts were read or written?
- Which provider calls happened?
- Which snippets were read?
- Which budgets were consumed?
- Which gates passed?
- Why did the run stop?

## Safety Rules

Executable text must be safe by construction.

- Validate before run.
- Dry-run before side effects.
- Require declared outputs for writes.
- Require explicit permission for shell commands.
- Never execute model-generated code in v1.
- Cap recursion depth and worker counts.
- Redact secrets in traces.
- Use cache/attest for external context and artifacts.
- Prefer deterministic local tests and smokes.

## Relationship To PE's Existing Pieces

Executable text should unify PE's existing strengths without making any one
subsystem dominant.

Prompt and module files provide the reusable text units. Structured outputs and
schemas provide type boundaries. Optimization work provides metric-backed
improvement loops. Distributed execution provides bounded local workers.
Consensus provides deterministic aggregation. Cache and attest provide artifact
identity. Script tests and examples prove the workflows actually run.

The design goal is to make those pieces compose through text artifacts that PE
can validate, run, trace, and replay.

## First Milestone

The first milestone should prove the toolchain feel before broad workflow code:

1. Define the shebang runner convention, starting with `pe run-text`.
2. Define `pe.text.v1`, `pe.workflow.v1`, and `pe.trace.v1`.
3. Specify how plain text graduates to templated text and then composed
   executable text.
4. Specify the operator standard library.
5. Add one realistic shebang prompt and one workflow example.
6. Sketch future CLI commands: `pe run-text`, `pe vet-text`,
   `pe exp workflow validate`, `pe exp workflow run`, and `pe exp recurse`.

Only after that should PE add runtime code.
