# PE Installation Guide

This guide covers installing PE (Go for Prompts) on various platforms and configuring it for your environment.

## Prerequisites

- **Go 1.21 or later** (required for building from source)
- **macOS, Linux, or Windows** (via WSL2)
- **Git** (optional, for version control features)

## Installation Methods

### 1. Install from Source (Recommended for Development)

Since PE is currently in active development, installing from source ensures you have the latest features:

```bash
# Clone the repository
git clone https://github.com/tmc/pe.git
cd pe

# Install directly to GOPATH/bin
go install ./cmd/pe

# Or build to a specific location
go build -o pe cmd/pe/main.go
sudo mv pe /usr/local/bin/
```

### 2. Install with Go (Once Published)

When PE is published to a module proxy, you'll be able to use:

```bash
# This will work once PE is properly published
go install github.com/tmc/pe/cmd/pe@latest
```

**Note**: Currently this may not work if the module is not published to a proxy.

### 3. Download Binary (Future)

Pre-built binaries will be available from the [releases page](https://github.com/tmc/pe/releases) once releases are created.

```bash
# macOS (Apple Silicon)
curl -L https://github.com/tmc/pe/releases/latest/download/pe-darwin-arm64 -o pe
chmod +x pe
sudo mv pe /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/tmc/pe/releases/latest/download/pe-darwin-amd64 -o pe
chmod +x pe
sudo mv pe /usr/local/bin/

# Linux
curl -L https://github.com/tmc/pe/releases/latest/download/pe-linux-amd64 -o pe
chmod +x pe
sudo mv pe /usr/local/bin/
```

## Initial Setup

### 1. Verify Installation

Check that PE is installed and accessible:

```bash
# Check PE is in your PATH
which pe
# Output: /Users/yourusername/go/bin/pe (or /usr/local/bin/pe)

# View available commands
pe --help

# Test with a simple prompt (requires API key)
pe run "What is 2+2?" --provider openai
```

### 2. Configure API Keys

PE supports multiple LLM providers. Set API keys as environment variables:

```bash
# OpenAI (for GPT models)
export OPENAI_API_KEY="sk-..."

# Anthropic (for Claude models)
export ANTHROPIC_API_KEY="sk-ant-..."

# Add to your shell profile for persistence
echo 'export OPENAI_API_KEY="sk-..."' >> ~/.zshrc
echo 'export ANTHROPIC_API_KEY="sk-ant-..."' >> ~/.zshrc

# Reload your shell
source ~/.zshrc
```

**Supported Providers:**
- `openai` - OpenAI models (GPT-4, GPT-3.5, etc.)
- `anthropic` - Anthropic models (Claude 3, Claude 2, etc.)
- `cgpt` - cgpt command-line tool (if installed)

### 3. Initialize a PE Project (Optional)

For project-specific prompts and configuration:

```bash
# In your project directory
cd your-project
pe init

# This creates:
# .pe/
#   ├── prompts/        # Project prompts
#   ├── evaluations/    # Evaluation configs
#   └── cache/          # Project-specific cache
```

## Quick Start

Once installed and configured, try these examples:

```bash
# Simple prompt execution
pe run "Explain what PE is in one sentence" --provider openai

# Using a prompt file with variables
echo "Translate {{.text}} to {{.language}}" > translate.prompt
pe run translate.prompt --var text="Hello" --var language="Spanish" --provider anthropic

# Evaluate a prompt against test cases
pe eval your-config.yaml

# Run optimization
pe optimize --prompt "Summarize text concisely" --method textgrad --iterations 3

# View all commands
pe --help
```

## Platform Notes

### macOS

No special setup required. PE works on both Intel and Apple Silicon Macs.

### Linux

PE works on most Linux distributions. Ensure Go is installed and `$GOPATH/bin` is in your PATH.

### Windows

Use WSL2 for the best experience:

```bash
# In WSL2 Ubuntu/Debian
sudo apt update
sudo apt install golang-go
go install github.com/tmc/pe/cmd/pe@latest
```

## Environment Variables

PE recognizes these environment variables:

```bash
# API Keys (Required for providers)
OPENAI_API_KEY          # OpenAI API key
ANTHROPIC_API_KEY       # Anthropic API key

# Optional Configuration
PE_TEST_MODE           # Enable test mode (true/false)
PE_DEBUG               # Enable debug output (true/false)
```

## Shell Completion (Optional)

PE supports shell completion for common shells:

### Zsh

```bash
# Generate and install completion
pe completion zsh > "${fpath[1]}/_pe"

# Or add to your .zshrc
echo 'eval "$(pe completion zsh)"' >> ~/.zshrc
source ~/.zshrc
```

### Bash

```bash
# Generate and source completion
pe completion bash > ~/.pe-completion.bash
echo 'source ~/.pe-completion.bash' >> ~/.bashrc
source ~/.bashrc
```

### Fish

```bash
# Generate completion
pe completion fish > ~/.config/fish/completions/pe.fish
```

## Troubleshooting

### Command Not Found

If `pe` is not found after installation:

```bash
# Check if pe is installed
which pe

# Check Go bin is in PATH
echo $PATH | grep "$(go env GOPATH)/bin"

# If not in PATH, add it
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# Verify Go installation
go version
```

### Permission Denied

If you get permission errors:

```bash
# Fix binary permissions
chmod +x $(which pe)

# Or if pe is in /usr/local/bin
sudo chmod +x /usr/local/bin/pe
```

### API Key Issues

Test that your API keys are set correctly:

```bash
# Check environment variables
echo $OPENAI_API_KEY
echo $ANTHROPIC_API_KEY

# Test with a simple prompt
pe run "Say hello" --provider openai

# Enable debug output if there are issues
PE_DEBUG=true pe run "test" --provider openai
```

### Build Errors

If you encounter build errors:

```bash
# Ensure you have the correct Go version
go version  # Should be 1.21 or later

# Clean and rebuild
cd path/to/pe
go clean
go build -o pe cmd/pe/main.go

# Check for dependency issues
go mod tidy
go mod verify
```

## Next Steps

- Read the [Getting Started Guide](GETTING_STARTED.md)
- Explore the [Command Reference](CLI_REFERENCE.md)
- Check out [Example Prompts](../example/)
- Review the [Documentation Overview](README.md)
