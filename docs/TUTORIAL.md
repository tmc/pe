# PE Hands-On Tutorial

This tutorial shows the current release-facing PE workflow. It uses commands
that are present in `pe --help` and avoids future configuration shapes.

For installation, see [INSTALLATION.md](INSTALLATION.md). For exact flags, see
[CLI_REFERENCE.md](CLI_REFERENCE.md).

## 1. Start With Plain Text

A PE prompt can be just text:

```bash
cat > review.prompt <<'PROMPT'
Review this change for release risk.
PROMPT

pe run-text review.prompt
```

`pe run-text` renders executable text locally. It does not call providers,
tools, shells, or the network.

## 2. Add Typed Inputs

Add front matter when the text needs a contract:

```bash
cat > release-review.prompt <<'PROMPT'
---
kind: pe.text.v1
inputs:
  topic:
    type: string
metadata:
  owner: release
safety:
  providers:
    allow: [local]
---
Review {{ .topic }} for release risk.
PROMPT

pe run-text release-review.prompt --var topic=v0.5.0
pe run-text release-review.prompt --check
```

Missing inputs are rejected:

```bash
pe run-text release-review.prompt
```

## 3. Compose Local Text

Executable text can import local text explicitly:

```bash
cat > checklist.prompt <<'PROMPT'
Check tests, docs, examples, and security notes.
PROMPT

cat > composed.prompt <<'PROMPT'
---
kind: pe.text.v1
imports:
  checklist: checklist.prompt
---
{{ import "checklist" }}
PROMPT

pe run-text composed.prompt
```

Imports are local files resolved relative to the importing file.

## 4. Add Module Policy

`pe.mod` can declare what executable text may use:

```bash
cat > pe.mod <<'MOD'
module example.com/prompts

pe 1

capability {
    providers deny remote
    tools deny shell network
}

policy {
    composition strict
    require-typed-io true
}
MOD

pe mod vet release-review.prompt
```

`pe mod vet` checks the module policy against executable-text front matter.

## 5. Run Current Offline Examples

The release-facing examples are under `examples/current-commands`:

```bash
go build -o /tmp/pe ./cmd/pe
PE_BIN=/tmp/pe ./examples/current-commands/smoke.sh
PE_BIN=/tmp/pe ./examples/current-commands/release-local-workflows/smoke.sh
```

These smokes cover current local workflows, including `pe diff`, `pe mod tidy`,
`pe exp attest`, and `pe exp cache`.

## 6. Use Live Providers Deliberately

Provider-backed commands require credentials and may send prompt data to the
configured provider:

```bash
export OPENAI_API_KEY=...
pe run release-review.prompt --var topic=v0.5.0 --provider openai
```

For release validation without credentials, prefer `pe run-text`, mock-provider
smokes, and the script tests in `tests/testdata/script`.

## Legacy Tutorial

The previous long tutorial is archived at
[archive/TUTORIAL_LEGACY.md](archive/TUTORIAL_LEGACY.md). It contains historical
and aspirational material and is not release-facing command documentation.
