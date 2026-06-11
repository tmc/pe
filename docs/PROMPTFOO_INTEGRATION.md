# Promptfoo Integration Guide

PE exposes Promptfoo compatibility through the `pe-promptfoo` plugin. Once the plugin binary is on your `PATH`, PE discovers it automatically and makes it available as `pe promptfoo`.

## Installation

Build the plugin from this repository and place it on your `PATH`:

```bash
go build -o ~/bin/pe-promptfoo ./plugins/promptfoo
pe plugin list
pe promptfoo --help
```

## Supported Scope

The current plugin supports configuration-file workflows for the common evaluation schema:

- `description`
- `prompts`
- `providers`
- `tests`
- `defaultTest`

When exporting PE configs, the plugin normalizes common PE key variants:

- `variables` -> `vars`
- `assertions` -> `assert`
- assertion types like `not_contains` -> `not-contains`

When importing Promptfoo configs, provider objects are reshaped into PE-friendly entries using:

- `apiProvider` + `model` -> `id` (for example `openai:gpt-4o-mini`)
- remaining provider fields -> `config`

## Commands

Import a Promptfoo-compatible config into PE's compatibility shape:

```bash
pe promptfoo import promptfooconfig.yaml -o pe-config.yaml
pe promptfoo import promptfooconfig.yaml -o pe-config.json -f json
```

Export a PE config to Promptfoo-friendly keys:

```bash
pe promptfoo export pe-config.yaml -o promptfooconfig.yaml
pe promptfoo export pe-config.yaml -o promptfooconfig.json -f json
```

Convert in either direction:

```bash
pe promptfoo convert promptfooconfig.yaml pe-config.yaml
pe promptfoo convert pe-config.yaml promptfooconfig.json --direction pe-to-promptfoo
```

Formats default from file extensions. You can override them explicitly:

```bash
pe promptfoo convert input.yaml output.json --direction pe-to-promptfoo --output-format json
pe promptfoo convert input.json output.yaml --input-format json --output-format yaml
```

## Example

Promptfoo-style input:

```yaml
prompts:
  - "Answer {{question}}"
providers:
  - id: gpt4-mini
    apiProvider: openai
    model: gpt-4o-mini
tests:
  - vars:
      question: "What is 2+2?"
    assert:
      - type: llm-rubric
        value: "Says 4"
```

Imported PE-compatible output:

```yaml
version: "1.0"
prompts:
  - "Answer {{question}}"
providers:
  - id: openai:gpt-4o-mini
tests:
  - vars:
      question: "What is 2+2?"
    assert:
      - type: llm-rubric
        value: "Says 4"
```

## Assertion Parity

PE's evaluator accepts Promptfoo's assertion `type` ids directly: an unmodified
`promptfooconfig.yaml` runs against PE without rewriting its `assert` blocks.
Promptfoo and PE chose different ids for some checks (`regex`/`matches`,
`llm-rubric`/`llm-judge`, `similar`/`similarity`, `is-json`/`json`); the alias
table in `internal/promptfoo/evaluation/evaluator/aliases.go` resolves them, and
the `not-` prefix inverts any check. The source of truth is that alias map plus
`isModelGraded`; the tables below mirror it.

### Supported (deterministic — no judge provider required)

`contains`, `not-contains`, `contains-any`, `contains-all`, `icontains`,
`icontains-any`, `icontains-all`, `equals`, `regex`/`matches`, `length`,
`starts-with`, `is-json`/`json`/`contains-json`, `is-sql`/`sql`, `contains-sql`,
`is-html`, `contains-html`, `is-xml`, `contains-xml`, `is-refusal`,
`is-valid-openai-function-call` (and the non-`openai` spelling
`is-valid-function-call`), `is-valid-openai-tools-call`, `word-count`,
`levenshtein`, `rouge-n`, `rouge-l`, `rouge-s`, `bleu`, `gleu`, `latency`,
`cost`, `tokens`, `finish-reason`, `pass-at-n`, `code`, `structure`,
`structured-output`.

Text-metric ids (`rouge-*`, `bleu`, `gleu`, `levenshtein`) are numerically
cross-checked against the libraries Promptfoo calls (js-rouge 3.2.0, etc.):
case-sensitive, beta=1 F1 — see the per-assertion doc comments.

### Supported (model-graded — require a judge provider)

`llm-rubric`/`llm-judge`, `g-eval`, `model-graded-factuality`/`factuality`,
`model-graded-closedqa`, `classifier`/`classify`, `similar`/`similarity`
(and `similar:cosine`, `similar:dot`, `similar:euclidean`), `answer-relevance`,
`context-faithfulness`, `context-recall`, `context-relevance`,
`conversation-relevance`, `agent-rubric`, `search-rubric`, `skill-used`,
`trajectory:goal-success`, plus the native `readability`, `sentiment`,
`toxicity`, `coherence`. These error (rather than silently pass) when no judge
provider is configured.

Two documented degradations: `agent-rubric`/`search-rubric` run as a plain
local rubric — PE does not enforce Promptfoo's agentic-tool / web-search grader
requirement (analogous to `similar:cosine` ignoring the metric refinement).

### Implemented but inert until a provider supplies the data

These compute faithfully the moment the response carries the required signal;
PE has no provider that surfaces it yet, so today they fail with a clear message
rather than passing vacuously.

| Assertion | Needs |
|-----------|-------|
| `perplexity`, `perplexity-score` | `metadata.logprobs` (per-token natural-log probabilities) |
| `tool-call-f1`, `trajectory:tool-used`, `trajectory:tool-sequence`, `trajectory:tool-args-match`, `trajectory:step-count` | a recorded `_trajectory` var (falls back to parsing OpenAI `tool_calls` from the output) |
| `trace-span-count`, `trace-span-duration`, `trace-error-spans` | a recorded `_trace` var (PE has no in-process OTEL collector) |

### Comparative (row-level, applied as a post-pass)

`select-best` and `max-score` compare a test row's outputs across providers, so
they are handled after per-output assertions in `comparative.go` rather than via
the alias map.

### Deliberately not implemented (policy)

| Assertion | Reason |
|-----------|--------|
| `javascript`, `python`, `ruby`, `webhook` | no-external-execution policy (see `docs/EXTERNAL_EXECUTION_POLICY.md`) |
| `moderation`, `guardrails`, `pi` | call hosted scoring APIs (OpenAI/Azure moderation, AWS/Azure guardrails, Pi Labs); PE ships no client for these |
| `meteor` | needs the WordNet dataset; violates PE's no-large-data-dependency policy |
| `human` | manual/interactive grading; no automated analogue |

An assertion id outside every list above resolves `ok=false` in
`normalizeAssertionType` and falls back to the string-match path or surfaces an
unsupported-assertion warning, rather than silently passing.

## Current Limits

The plugin does not currently implement full Promptfoo results import/export, file include expansion, or deep provider-schema translation beyond the common config shapes above.
