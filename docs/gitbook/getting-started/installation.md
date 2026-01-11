# Installation

PE (Go for Prompts) is a single binary that works on macOS, Linux, and Windows.

## Prerequisites

*   **Go 1.21+** (if building from source)
*   Access to LLM API keys (OpenAI, Anthropic) or local models (Ollama).

## Install from Source

Currently, the recommended way to install PE is from source:

```bash
go install github.com/tmc/pe/cmd/pe@latest
```

Ensure your `$(go env GOPATH)/bin` is in your system `PATH`.

## Configure API Keys

PE uses environment variables for authentication. Add these to your simplified shell profile (e.g., `.zshrc`, `.bashrc`):

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
```

## Verify Installation

Check that `pe` is accessible:

```bash
pe --version
```

You should see the version output. Try running a simple prompt:

```bash
pe run "Hello, world!" --provider openai
```
