# Introduction

PE is a prompt engineering toolkit organized around a single CLI.

It follows Go-tool-style workflows for running prompts, testing evaluations,
formatting inputs, and managing prompt modules.

## Why PE?

*   **Prompts as Code**: Treat prompts like software artifacts—versioned, tested, and modular.
*   **Unix Philosophy**: Composable commands (`pe ask`, `pe filter`, `pe reduce`) for pipelines.
*   **Pragmatic by Default**: Stable core commands for day-to-day work, with prototypes available behind `pe exp`.
*   **Developer Experience**: Native binary with no Python dependency for the core CLI.

## Key Features

*   **Native Providers**: Built-in support for OpenAI and Anthropic, plus integration with Ollama, Llama.cpp, and generic CLIs.
*   **Evaluation Framework**: Testing with `pe eval`, supporting assertions, pass@k, and structured output validation.
*   **Optimization**: TextGrad, PE2, and APEX commands under `pe experimental`.
*   **Module System**: Manage prompt dependencies with `pe mod`.
*   **Security**: OWASP LLM Top 10 scanning plus prototype attestation, caching, and distributed command groups.

## Getting Started

*   [**Installation**](getting-started/installation.md): Set up PE on macOS, Linux, or Windows.
*   [**Quick Start**](getting-started/quick-start.md): Run a first prompt.
*   [**Core Workflows**](workflows/README.md): Learn how to build, test, and optimize prompts.

<br>

> **Note**: PE is currently in active development. Experimental features are available under the `pe exp` command.
