# Quick Start: PE Modules & Attestation

This guide walks through using PE's module registry and cryptographic attestation features.

## Prerequisites

1. Build PE:
   ```bash
   go build ./cmd/pe
   export PATH=$PWD:$PATH
   ```

2. Set up GitHub token (for module registry):
   ```bash
   export GITHUB_TOKEN=your_github_token_with_gist_permissions
   ```

## Module Registry Quick Start

### 1. Create a Module

```bash
# Create a simple hello module
pe mod init myname/hello --prompt='Say hello in a creative and fun way!'

# Create a template-based module
pe mod init myname/translator --prompt='Translate "{{.Text}}" to {{.Language}}'
```

### 2. Push to Registry

```bash
# Push as public gist
pe push myname/hello --public

# Push as private gist (default)
pe push myname/translator
```

### 3. Run Modules

```bash
# Run the latest version
pe run myname/hello@latest

# Run with template variables
pe run myname/translator@latest --var Text="Hello world" --var Language=French
```

### 4. Module Discovery

```bash
# List local modules
pe mod list

# Get module info
pe mod get myname/hello
```

## Cryptographic Attestation Quick Start

### 1. Run with Attestation

```bash
# Create attestation for a run
pe run "What is 2+2?" --attest

# Run module with attestation
pe run myname/hello@latest --attest
```

### 2. View Attestations

```bash
# List all attestations
pe attest list

# Show specific attestation details
pe attest show <attestation-id>

# Show last attestation
pe attest list --limit 1 | tail -1 | cut -f1 | xargs pe attest show
```

### 3. Verify Attestations

```bash
# Verify entire chain
pe attest verify

# Verify specific attestation
pe attest verify <attestation-id>

# Run with chain verification
pe run "Important calculation" --verify --attest
```

### 4. Export Attestations

```bash
# Export as JSON
pe attest export --format json > attestations.json

# Export as JSONL (one per line)
pe attest export --format jsonl > attestations.jsonl

# Export specific date range
pe attest export --after 2024-01-01 --before 2024-02-01
```

### 5. Key Management

```bash
# Show current public key
pe attest key

# Generate new key pair
pe attest key --generate
```

## Combined Workflow Example

Here's a complete example combining modules and attestation:

```bash
# 1. Create a code review module
pe mod init dev/reviewer --prompt='Review this code for potential issues: {{.Code}}'

# 2. Push to registry
pe push dev/reviewer --public

# 3. Use with attestation
pe run dev/reviewer@latest --var Code="$(cat main.go)" --attest

# 4. Verify the attestation chain
pe attest verify

# 5. Export proof of review
pe attest export --format proof --output review-proof.json
```

## Setting Up a Registry

To create your own module registry:

```bash
# 1. Create registry gist (one-time setup)
curl -X POST https://api.github.com/gists \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "My PE Module Registry",
    "public": true,
    "files": {
      "index.json": {"content": "{}"},
      "README.md": {"content": "# My PE Module Registry\n"}
    }
  }'

# 2. Set registry ID (from response)
export PE_REGISTRY_GIST_ID=<gist_id_from_response>

# 3. Now modules will be registered when pushed
pe push myname/module
```

## Tips & Best Practices

1. **Module Naming**: Use `org/name` format for modules
2. **Versioning**: Update version in module.json before pushing updates
3. **Templates**: Use `{{.Variable}}` syntax for reusable prompts
4. **Attestation**: Use `--attest` for important or compliance-required runs
5. **Verification**: Use `--verify` before critical operations

## Troubleshooting

### Module Not Found
- Check `GITHUB_TOKEN` is set
- Verify module name format
- Clear cache: `rm -rf .pe/cache`

### Attestation Errors
- Ensure `.pe/attestations/` directory exists
- Check key file permissions
- Verify chain integrity: `pe attest verify`

### Registry Issues
- Verify `PE_REGISTRY_GIST_ID` if using custom registry
- Check GitHub token permissions
- Ensure gist is accessible

## Next Steps

- Read [MODULE_REGISTRY.md](MODULE_REGISTRY.md) for detailed module documentation
- Read [ATTESTATION.md](ATTESTATION.md) for cryptographic details
- Explore example modules in `example/`
- Create and share your own modules!