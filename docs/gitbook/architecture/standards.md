# LLM CLI Standards

PE adheres to strict standards to ensure composability and predictability.

## 1. Unix Philosophy
*   **Stdin/Stdout**: All commands must support piping.
*   **Text-Based**: Interfaces should be text-first / JSON-optional.
*   **One Thing Well**: `pe run` executes, `pe eval` evaluates.

## 2. API Key Handling
*   All providers MUST support environment variables for secrets (e.g., `OPENAI_API_KEY`).
*   Keys should NEVER be stored in plain text configuration files committed to Git.

## 3. Output Formats
*   **Human-Readable**: Default output should be friendly (tables, colors).
*   **Machine-Readable**: Every command MUST support `--format json` for integration.

## 4. Error Handling
*   Exit codes must be meaningful (0 for success, non-zero for failure).
*   Errors should be printed to `stderr`, keeping `stdout` clean for pipeline data.
