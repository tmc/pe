# PE Module System

PE includes a module system for sharing and reusing prompts, inspired by Go modules. Modules are stored as GitHub gists, providing a decentralized registry.

## Overview

The module system allows you to:
- Create reusable prompt modules
- Share prompts with others
- Version your prompts
- Run prompts directly from the registry

## Quick Start

### Using a Module

```bash
# Run a module from the registry
pe run tmc/hello@latest

# Run a specific version
pe run tmc/translate@v1.0.0

# With variables
pe run tmc/summarize@latest --var text="Long text here..."
```

### Creating a Module

```bash
# Initialize a new module
pe mod init myorg/hello --prompt='Say hello in {{.language}}'

# Edit the prompt
vi .pe/modules/myorg/hello/prompt.txt

# Test it
pe vet .pe/modules/myorg/hello/prompt.txt

# Push to registry (requires GITHUB_TOKEN)
export GITHUB_TOKEN=your-token
pe push myorg/hello
```

## Module Structure

A module consists of:
- `module.json` - Module metadata
- `prompt.txt` - The main prompt file
- Additional files (optional)

Example `module.json`:
```json
{
  "name": "tmc/hello",
  "version": "0.1.0",
  "description": "A friendly greeting prompt",
  "author": "tmc",
  "created": "2024-01-01T00:00:00Z",
  "updated": "2024-01-01T00:00:00Z",
  "gist_id": "abc123"
}
```

Example `prompt.txt`:
```
Say hello to {{.name}} in {{.language}}

-- defaults --
name=World&language=English

-- evals --
$ name="Alice" language="French"
Bonjour Alice!

$ name="Bob"
Hello Bob!
```

## Registry Architecture

PE uses GitHub gists as a decentralized registry:

1. **Root Gist** - Contains the registry index
2. **Module Gists** - Each module is stored as a gist
3. **Forks** - Users can fork and modify modules

### How It Works

1. `pe run org/module@version` checks local cache
2. If not cached, fetches from the registry
3. Registry lookup finds the module's gist ID
4. Module gist is fetched and cached
5. Prompt is executed

## Commands

### pe mod init

Initialize a new module:
```bash
pe mod init org/name --prompt='Your prompt here'
```

Options:
- `--force` - Overwrite existing module

### pe mod list

List available modules from the registry:
```bash
pe mod list
```

### pe mod get

Get information about a module:
```bash
pe mod get org/name
```

### pe push

Push a module to the registry:
```bash
pe push org/name
```

Options:
- `--public` - Make the gist public
- `--update` - Update existing gist

## Best Practices

1. **Versioning** - Use semantic versioning (v1.0.0)
2. **Testing** - Include comprehensive evals
3. **Defaults** - Provide sensible defaults
4. **Documentation** - Document your prompts well
5. **Examples** - Include usage examples

## Environment Variables

- `GITHUB_TOKEN` - Required for pushing modules
- `PE_ROOT_GIST_ID` - Override default registry location
- `PE_MODULE_CACHE` - Module cache directory (default: `.pe/cache`)

## Local Development

Test modules locally before pushing:
```bash
# Create module
pe mod init test/module --prompt='Test prompt'

# Run locally
pe run .pe/modules/test/module/prompt.txt

# Test with vet
pe vet .pe/modules/test/module/prompt.txt
```

## Security

- Modules are fetched over HTTPS
- Gist IDs are verified
- Local cache prevents repeated fetches
- Users control which modules they run

## Possible Enhancements

- Module dependencies
- Private registries
- Module signing
- Automated testing on push
- Module search and discovery
