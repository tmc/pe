# PE Installation Guide

This guide covers installing PE (Go for Prompts) on various platforms and configuring it for your environment.

## Prerequisites

- Go 1.21 or later (for building from source)
- macOS 11.0+ (for full sandbox support)
- Git (for version control features)

## Installation Methods

### 1. Install with Go (Recommended)

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

This installs the latest stable version of PE to `$GOPATH/bin`.

### 2. Install from Source

```bash
# Clone the repository
git clone https://github.com/tmc/pe.git
cd pe

# Build and install
go build -o pe cmd/pe/main.go
sudo mv pe /usr/local/bin/
```

### 3. Homebrew (macOS)

```bash
brew tap tmc/pe
brew install pe
```

### 4. Download Binary

Download pre-built binaries from the [releases page](https://github.com/tmc/pe/releases).

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

```bash
pe version
# Output: PE v1.0.0 (go1.21)
```

### 2. Configure API Keys

PE needs API keys for LLM providers. Set them as environment variables:

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# Google AI
export GOOGLE_AI_API_KEY="..."

# Add to your shell profile (~/.zshrc or ~/.bashrc)
echo 'export OPENAI_API_KEY="sk-..."' >> ~/.zshrc
```

### 3. Initialize PE Home Directory

```bash
# PE creates its configuration in ~/.pe
pe init --global

# This creates:
# ~/.pe/
#   ├── config.yaml     # Global configuration
#   ├── cache/          # Shared cache
#   ├── plugins/        # Installed plugins
#   └── styles/         # Style guides
```

### 4. Configure Defaults

```bash
# Set default provider
pe config set default.provider gpt-4

# Set default model
pe config set default.model gpt-4-turbo-preview

# Enable cache
pe config set cache.enabled true

# Set cache directory
pe config set cache.dir ~/.pe/cache

# Configure security
pe config set security.sandbox strict
pe config set security.require-signatures true
```

## Platform-Specific Setup

### macOS

#### Enable Sandbox Support

PE uses macOS App Sandbox for security. Grant necessary permissions:

```bash
# First run will prompt for permissions
pe run "test" --sandbox=strict

# Grant file access if needed
pe sandbox grant ~/Documents
```

#### Code Signing (Optional)

For distribution, sign the PE binary:

```bash
codesign -s "Developer ID Application: Your Name" /usr/local/bin/pe
```

### Linux

#### AppArmor Profile (Optional)

Create an AppArmor profile for additional security:

```bash
sudo tee /etc/apparmor.d/pe > /dev/null << 'EOF'
#include <tunables/global>

/usr/local/bin/pe {
  #include <abstractions/base>
  
  # Allow reading prompts
  /home/*/.pe/** r,
  /home/*/prompts/** r,
  
  # Allow network for API calls
  network inet stream,
  network inet6 stream,
  
  # Deny everything else
  deny /** w,
}
EOF

sudo apparmor_parser -r /etc/apparmor.d/pe
```

### Windows (WSL2)

PE works best on Windows through WSL2:

```bash
# In WSL2 Ubuntu
go install github.com/tmc/pe/cmd/pe@latest
```

## Environment Variables

PE recognizes these environment variables:

```bash
# API Keys
OPENAI_API_KEY          # OpenAI API key
ANTHROPIC_API_KEY       # Anthropic API key
GOOGLE_AI_API_KEY       # Google AI API key

# Configuration
PE_HOME                 # PE home directory (default: ~/.pe)
PE_CONFIG              # Config file path (default: $PE_HOME/config.yaml)
PE_CACHE_DIR           # Cache directory (default: $PE_HOME/cache)
PE_PLUGIN_DIR          # Plugin directory (default: $PE_HOME/plugins)

# Behavior
PE_MOCK_MODE           # Enable mock mode for testing (true/false)
PE_DEBUG               # Enable debug output (true/false)
PE_TRACE               # Enable trace logging (true/false)
PE_NO_COLOR            # Disable colored output (true/false)
PE_SANDBOX             # Default sandbox mode (strict/relaxed/disabled)
```

## Shell Completion

### Zsh

```bash
# Generate completion
pe completion zsh > ~/.pe/completion.zsh

# Add to ~/.zshrc
echo 'source ~/.pe/completion.zsh' >> ~/.zshrc
source ~/.zshrc
```

### Bash

```bash
# Generate completion
pe completion bash > ~/.pe/completion.bash

# Add to ~/.bashrc
echo 'source ~/.pe/completion.bash' >> ~/.bashrc
source ~/.bashrc
```

### Fish

```bash
# Generate completion
pe completion fish > ~/.config/fish/completions/pe.fish
```

## Verify Installation

Run the installation test:

```bash
pe doctor
```

Expected output:
```
PE Installation Check
====================
✓ PE binary: /usr/local/bin/pe
✓ Version: v1.0.0
✓ Go version: go1.21
✓ Config file: ~/.pe/config.yaml
✓ Cache directory: ~/.pe/cache (2.3 MB)
✓ Plugin directory: ~/.pe/plugins (3 plugins)

API Keys:
✓ OpenAI: Configured
✓ Anthropic: Configured
✗ Google AI: Not configured

Security:
✓ Sandbox: Available (macOS)
✓ Trust store: Initialized
✓ Signatures: Enabled

All systems operational!
```

## Troubleshooting

### Command Not Found

If `pe` is not found after installation:

```bash
# Check Go bin is in PATH
echo $PATH | grep -q "$(go env GOPATH)/bin" || echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc

# Reload shell
source ~/.zshrc
```

### Permission Denied

If you get permission errors:

```bash
# Fix permissions
chmod +x $(which pe)

# Fix cache permissions
chmod -R u+rw ~/.pe/cache
```

### API Key Issues

Test API keys:

```bash
# Test OpenAI
pe run "test" --provider openai --debug

# Test Anthropic  
pe run "test" --provider anthropic --debug
```

### Sandbox Issues (macOS)

If sandbox is not working:

```bash
# Reset sandbox permissions
pe sandbox reset

# Run without sandbox (not recommended)
pe run "test" --sandbox=disabled
```

## Next Steps

- Read the [Getting Started Tutorial](TUTORIAL.md)
- Explore the [Command Reference](COMMANDS.md)
- Learn about [Plugin Development](PLUGINS.md)
- Join the community on [Discord](https://discord.gg/pe-prompts)