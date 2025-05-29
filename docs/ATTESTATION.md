# Cryptographic Attestation in PE

PE provides cryptographic attestation for every prompt run, creating an immutable, verifiable audit trail of all LLM interactions. This feature is essential for compliance, security, debugging, and reproducibility.

## Overview

Every prompt execution can be cryptographically signed and chained, similar to a blockchain. Each attestation includes:

- **Input hash**: SHA-256 of the canonicalized prompt and parameters
- **Output hash**: SHA-256 of the LLM response
- **Chain hash**: Link to the previous attestation
- **Digital signature**: Ed25519 signature for authenticity
- **Metadata**: Timestamp, provider, model, tokens, latency

## Quick Start

### Enable Attestation

```bash
# Run with attestation
pe run "What is 2+2?" --attest

# Run with verification
pe run prompt.txt --attest --verify

# Always attest (environment variable)
export PE_ATTEST=true
pe run "Hello world"
```

### View Attestations

```bash
# List recent attestations
pe attest list

# Show specific attestation
pe attest show run-1234567890-abcdef

# Verify the chain
pe attest verify
```

## Attestation Structure

Each attestation contains:

```json
{
  "id": "run-1706234567-a1b2c3",
  "timestamp": "2024-01-25T12:34:56Z",
  "version": "0.1.0",
  
  "prompt": "What is {{.query}}?",
  "variables": {"query": "2+2"},
  "provider": "openai",
  "model": "gpt-4",
  "temperature": 0.7,
  "system_prompt": "",
  
  "response": "2 + 2 equals 4.",
  "tokens_used": {
    "prompt": 15,
    "completion": 8,
    "total": 23
  },
  "latency": "1.234s",
  "finish_reason": "stop",
  
  "input_hash": "3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e",
  "output_hash": "9f8e7d6c5b4a3b2c1d0e9f8a7b6c5d4e",
  "previous_hash": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d",
  "signature": "base64-encoded-signature",
  "public_key": "base64-encoded-public-key"
}
```

## Commands

### pe attest list

List attestations with filtering:

```bash
# Show last 10 (default)
pe attest list

# Show more
pe attest list --limit 50

# Filter by provider
pe attest list --provider openai

# Filter by date
pe attest list --since 2024-01-25T00:00:00Z
```

### pe attest verify

Verify attestation integrity:

```bash
# Verify entire chain
pe attest verify

# Verify specific attestation
pe attest verify run-1234567890-abcdef

# Verbose verification
pe attest verify --verbose
```

### pe attest export

Export attestations for audit:

```bash
# Export as JSON
pe attest export --format json > attestations.json

# Export as CSV summary
pe attest export --format csv > attestations.csv

# Export cryptographic proof bundle
pe attest export --format proof > proof.json

# Export as JSON Lines (streaming-friendly)
pe attest export --format jsonl > attestations.jsonl
```

### pe attest key

Manage signing keys:

```bash
# Show current public key
pe attest key

# Generate new key pair (key rotation)
pe attest key --generate
```

## Security Properties

### Tamper Evidence
- Any modification to attestations is detectable
- Chain integrity ensures chronological ordering
- Hashes prevent content tampering

### Non-Repudiation
- Ed25519 signatures prove authenticity
- Public key identifies the signer
- Timestamp proves when execution occurred

### Verifiability
- Anyone can verify attestations with the public key
- Chain can be independently validated
- No trust in intermediaries required

## Use Cases

### Compliance & Audit
- Regulatory compliance (GDPR, HIPAA, etc.)
- Security audit trails
- Incident investigation
- Usage tracking

### Debugging & Development
- Reproduce exact prompts and responses
- Track prompt evolution
- Performance analysis
- A/B testing verification

### Cost Management
- Token usage tracking
- Provider cost allocation
- Budget enforcement
- Usage analytics

### Research & Reproducibility
- Experiment tracking
- Result verification
- Dataset creation
- Benchmark validation

## Advanced Features

### Chain Properties
- Each attestation links to the previous one
- Forms an append-only log
- Enables time-travel queries
- Supports merkle proofs (future)

### Export Formats

**JSON**: Full attestation data
```bash
pe attest export --format json
```

**CSV**: Summary for spreadsheets
```bash
pe attest export --format csv
```

**Proof Bundle**: For third-party verification
```bash
pe attest export --format proof
```

### Integration

Environment variables:
- `PE_ATTEST`: Always create attestations (true/false)
- `PE_VERIFY`: Always verify before running (true/false)
- `PE_DATA_DIR`: Attestation storage location

API integration:
```go
// Using attestation service directly
service, _ := attestation.NewAttestationService(".pe")
att, _ := service.AttestRun(input, output)
```

## Best Practices

1. **Enable for Production**: Always attest production prompts
2. **Regular Verification**: Periodically verify chain integrity
3. **Backup Attestations**: Export and backup attestation data
4. **Key Security**: Protect signing keys appropriately
5. **Monitor Chain**: Set up alerts for verification failures

## Technical Details

### Cryptography
- **Signing**: Ed25519 (RFC 8032)
- **Hashing**: SHA-256
- **Encoding**: Base64 for signatures and keys
- **Chain**: Each attestation includes previous hash

### Storage
- **Location**: `.pe/attestations/`
- **Chain file**: `chain.jsonl` (append-only)
- **Key Storage**:
  - **macOS**: Keychain (secure, hardware-backed when available)
  - **Other OS**: Encrypted file with PBKDF2 + AES-GCM

### Performance
- Minimal overhead (~1-2ms per attestation)
- Async storage writes
- Efficient append-only design
- Compressed storage (future)

## Future Enhancements

- **Distributed Attestation**: Multi-party signing
- **Blockchain Integration**: Anchor to public chains
- **Zero-Knowledge Proofs**: Privacy-preserving verification
- **Remote Attestation**: Network attestation service
- **HSM Support**: Hardware security module integration
- **Time Stamping**: RFC 3161 timestamps

## Troubleshooting

### Chain Verification Failed
- Check for disk corruption
- Verify key hasn't changed
- Look for missing attestations

### Performance Impact
- Use async attestation (future)
- Batch attestations (future)
- Archive old attestations

### Key Management

#### macOS Keychain Storage
- Keys automatically stored in macOS Keychain
- Access requires user authentication
- Hardware encryption when available (T2/Apple Silicon)
- View in Keychain Access app: search for 'com.github.tmc.pe'

#### Other Platforms
- Keys encrypted with passphrase
- PBKDF2 with 100,000 iterations
- AES-256-GCM encryption
- Passphrase prompted on first use

#### Key Operations
```bash
# View current public key
pe attest key

# Generate new key pair
pe attest key --generate

# Keys are automatically migrated from old file format
```