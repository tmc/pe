# Experimental Features

PE includes a suite of experimental features accessible via the `pe exp` subcommand. These are prototypes that may change or graduate to top-level commands in future releases.

## Usage

```bash
pe exp [command]
```

## Available Prototypes

### Core Functionality
*   **`pe exp distributed`**: Distributed execution prototype.
*   **`pe exp attest`**: Cryptographic attestation prototype.
*   **`pe exp cache`**: Content-addressed caching prototype.
*   **`pe exp compose`**: Prototype compose entrypoint in the `exp` group.
*   **`pe experimental compose`**: Research/experimental compose command with advanced flags.

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

> **Note**: Prototype commands are intentionally unstable. Use `pe exp [command] --help`
> to inspect the currently exposed interface for your build.
