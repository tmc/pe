# Experimental Features

PE includes a suite of experimental features accessible via the `pe exp` subcommand. These are prototypes that may change or graduate to top-level commands in future releases.

## Usage

```bash
pe exp [command]
```

## Available Prototypes

### Core Functionality
*   **`pe exp distributed`**: Distributed execution engine for running massive evaluations across multiple machines.
*   **`pe exp attest`**: Cryptographic attestation for verifying the provenance of prompt results.
*   **`pe exp cache`**: Advanced content-addressed caching system.
*   **`pe exp compose`**: Type-safe prompt composition (formerly `pe compose`).

### Pipeline & Data Tools
*   **`pe exp transform`**: Transform prompts between formats.
*   **`pe exp merge`**: Intelligently merge prompt components.
*   **`pe exp batch`**: efficient batch processing of prompts.
*   **`pe exp sweep`**: Hyperparameter sweeping for prompt optimization.
*   **`pe exp export` / `import`**: Export/Import prompts to external formats.
*   **`pe exp sync`**: Synchronize with external systems.

### Analysis & Debugging
*   **`pe exp trace`**: Detailed execution tracing.
*   **`pe exp explain`**: Explain prompt behavior and provider responses.
*   **`pe exp lint`**: Lint prompts for best practices.
*   **`pe exp report`**: Generate comprehensive reports.

### Orchestration
*   **`pe exp schedule`**: Schedule prompt execution.
*   **`pe exp workflow`**: Define and execute complex workflows.
*   **`pe exp hook`**: Manage lifecycle hooks.

> **Note**: As prototypes, these commands may have limited functionality or stability.
