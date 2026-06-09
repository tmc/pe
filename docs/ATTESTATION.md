# Cryptographic Attestation

Attestation is currently exposed as a prototype command group:

```bash
pe exp attest --help
```

## Current Status

- `pe exp attest` exists in the CLI.
- It is marked prototype and exposes unsigned manifest and signed-envelope
  subcommands.
- Command shape and behavior may change as the feature is finalized.
- Unsigned manifests detect local content changes.
- Signed manifests wrap the existing unsigned payload in an Ed25519 envelope.
  They bind the manifest to a public key, but do not by themselves prove key
  ownership, freshness, or remote origin.

## Recommended Usage Today

Use the prototype command help to inspect the current interface:

```bash
pe exp --help
pe exp attest --help
```

Create and verify an unsigned local manifest:

```bash
pe exp attest manifest --root prompts review.prompt > manifest.json
pe exp attest verify --root prompts manifest.json
```

Create and verify a signed envelope:

```bash
pe exp attest keygen > attest-key.json
pe exp attest sign --private-key "$(jq -r .private_key attest-key.json)" manifest.json > signed-manifest.json
pe exp attest verify-signed --root prompts --public-key "$(jq -r .public_key attest-key.json)" signed-manifest.json
```

For production workflows, rely on stable commands (`pe run`, `pe eval`, `pe test`, `pe security`) until attestation exits prototype status.
