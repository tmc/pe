# Introduction

Welcome to **PE (Prompt Engineering)**, the comprehensive toolkit that brings the power and simplicity of the Go toolchain to prompt engineering.

**PE is "Go for Prompts".** Just as `go run`, `go test`, and `go fmt` standardized Go development, `pe` unifies the fragmented landscape of prompt engineering into a single, cohesive binary.

## Why PE?

*   **Prompts as Code**: Treat prompts like software artifacts—versioned, tested, and modular.
*   **Unix Philosophy**: Composable commands (`pe ask`, `pe filter`, `pe reduce`) that work beautifully in pipelines.
*   **Pragmatic by Default**: Stable core commands for day-to-day work, with prototypes available behind `pe exp`.
*   **Developer Experience**: Fast, native binary with no Python dependencies for the core toolchain.

## Key Features

*   **🚀 Native Providers**: Built-in support for OpenAI and Anthropic, plus integration with Ollama, Llama.cpp, and generic CLIs.
*   **🧪 Evaluation Framework**: Comprehensive testing with `pe eval`, supporting assertions, pass@k, and structured output validation.
*   **🔬 Optimization**: State-of-the-art algorithms (TextGrad, PE2, APEX) under `pe experimental`.
*   **📦 Module System**: Manage prompt dependencies with `pe mod` (inspired by Go modules).
*   **🛡️ Security**: OWASP LLM Top 10 scanning plus prototype attestation/caching/distributed command groups.

## Getting Started

Ready to dive in?

*   [**Installation**](getting-started/installation.md): Set up PE on macOS, Linux, or Windows.
*   [**Quick Start**](getting-started/quick-start.md): Run your first prompt in under 5 minutes.
*   [**Core Workflows**](workflows/README.md): Learn how to build, test, and optimize prompts.

<br>

> **Note**: PE is currently in active development. Experimental features are available under the `pe exp` command.
