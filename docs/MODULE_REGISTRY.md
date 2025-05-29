# PE Module Registry

The PE Module Registry provides a decentralized way to share and reuse prompt templates using GitHub gists as storage. This enables prompt sharing, versioning, and discovery within the PE ecosystem.

## Overview

The module registry allows you to:
- Package prompts as reusable modules
- Publish modules to a decentralized GitHub gist registry
- Version and update modules
- Run modules directly by reference (e.g., `pe run org/name@version`)
- Cryptographically attest module executions

## Architecture

The registry uses a two-tier gist structure:

1. **Root Registry Gist**: Contains an index of all registered modules
2. **Module Gists**: Individual gists containing module content

```
Root Registry Gist (PE_REGISTRY_GIST_ID)
├── index.json      # Module name → version → gist ID mapping
└── README.md       # Registry documentation

Module Gist
├── module.json     # Module metadata
├── prompt.txt      # The actual prompt template
└── [other files]   # Additional resources
```

## Getting Started

### Prerequisites

1. **GitHub Token**: Set the `GITHUB_TOKEN` environment variable with a token that has gist permissions:
   ```bash
   export GITHUB_TOKEN=your_github_token_here
   ```

2. **Registry Gist ID** (Optional): Set `PE_REGISTRY_GIST_ID` to use a specific registry:
   ```bash
   export PE_REGISTRY_GIST_ID=your_registry_gist_id
   ```

### Creating a Module

1. Initialize a new module:
   ```bash
   pe mod init org/modulename --prompt='Your prompt template here'
   ```

   Example with template variables:
   ```bash
   pe mod init tmc/translator --prompt='Translate {{.Text}} to {{.Language}}'
   ```

2. The module is created locally in `.pe/modules/org/modulename/` with:
   - `module.json`: Module metadata
   - `prompt.txt`: The prompt template

### Publishing a Module

Push your module to the registry:
```bash
pe push org/modulename
```

Options:
- `--public`: Make the gist public (default is private)
- `--update`: Update an existing gist

The push command will:
1. Create a GitHub gist with your module content
2. Register the module in the root registry (if configured)
3. Return the gist URL

### Using Modules

Run a module directly:
```bash
pe run org/modulename@latest
```

With template variables:
```bash
pe run tmc/translator@latest --var Text="Hello world" --var Language=Spanish
```

Version specification:
- `@latest`: Use the most recent version
- `@1.0.0`: Use a specific version
- `@v1`: Use version 1.x.x (if supported)

### Module Caching

Modules are cached locally in `.pe/cache/modules/` to avoid repeated downloads. Clear the cache with:
```bash
rm -rf .pe/cache
```

## Module Structure

### module.json

```json
{
  "name": "org/modulename",
  "version": "1.0.0",
  "description": "Brief description of what the prompt does",
  "author": "username",
  "gist_id": "abc123...",
  "created": "2024-01-01T00:00:00Z",
  "updated": "2024-01-01T00:00:00Z",
  "prompt_file": "prompt.txt"
}
```

### prompt.txt

The prompt file can contain:
- Plain text prompts
- Template variables using Go template syntax: `{{.VariableName}}`
- Multi-line prompts with complex instructions

## Registry Management

### Creating a Registry

To create your own registry:

1. Create a new gist with an empty `index.json`:
   ```json
   {}
   ```

2. Set the gist ID as your registry:
   ```bash
   export PE_REGISTRY_GIST_ID=your_gist_id
   ```

### Registry Index Format

The `index.json` file maps module names to versions and gist IDs:

```json
{
  "org/module1": {
    "1.0.0": "gist_id_1",
    "1.1.0": "gist_id_2",
    "latest": "gist_id_2"
  },
  "org/module2": {
    "0.1.0": "gist_id_3",
    "latest": "gist_id_3"
  }
}
```

## Integration with Attestation

Combine modules with cryptographic attestation:

```bash
# Run with attestation
pe run org/module@latest --attest

# Verify attestations before running
pe run org/module@latest --verify --attest
```

This creates a tamper-evident record of:
- The module that was run
- Input parameters
- Output produced
- Execution metadata

## Best Practices

1. **Versioning**: Use semantic versioning (major.minor.patch)
2. **Documentation**: Include clear descriptions and examples
3. **Templates**: Use template variables for reusable prompts
4. **Testing**: Test modules locally before publishing
5. **Security**: Review modules before running them

## Examples

### Example 1: Hello World Module

```bash
# Create
pe mod init myname/hello --prompt='Say hello in a creative way'

# Push
pe push myname/hello --public

# Use
pe run myname/hello@latest
```

### Example 2: Code Review Module

```bash
# Create with template
pe mod init dev/reviewer --prompt='Review this code for bugs and improvements: {{.Code}}'

# Push
pe push dev/reviewer

# Use with variable
pe run dev/reviewer@latest --var Code="$(cat main.go)"
```

### Example 3: Multi-Language Translator

```bash
# Create
pe mod init tools/translate --prompt='Translate "{{.Text}}" from {{.From}} to {{.To}}'

# Use
pe run tools/translate@latest --var Text="Hello" --var From=English --var To=Japanese
```

## Troubleshooting

### Module Not Found

If you get "module not found" errors:
1. Check that `GITHUB_TOKEN` is set
2. Verify the module name format (org/name)
3. Ensure the module has been pushed
4. Check network connectivity

### Registry Access Issues

If registry updates fail:
1. Verify `PE_REGISTRY_GIST_ID` is correct
2. Ensure your token has gist write permissions
3. Check that you own or have access to the registry gist

### Cache Issues

If modules aren't updating:
1. Clear the cache: `rm -rf .pe/cache`
2. Use `--update` flag when pushing updates
3. Check the module version in module.json

## Future Enhancements

Planned features for the module registry:
- Module search and discovery
- Module dependencies
- Private registries with access control
- Module signing and verification
- Automated testing of modules
- Module composition and inheritance
- Registry federation and mirroring