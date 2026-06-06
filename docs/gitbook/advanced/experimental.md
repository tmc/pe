# Experimental Features

PE includes a suite of experimental entrypoints under `pe exp`. Many of these
commands are intentionally registered as placeholders and return a "not yet
implemented" error instead of doing work.

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

### Registered Placeholders
These commands are visible for compatibility with older plans, but return a
not-implemented error in the current build:

*   `pe exp transform`
*   `pe exp merge`
*   `pe exp batch`
*   `pe exp sweep`
*   `pe exp export`
*   `pe exp import`
*   `pe exp sync`
*   `pe exp trace`
*   `pe exp explain`
*   `pe exp lint`
*   `pe exp report`
*   `pe exp schedule`
*   `pe exp workflow`
*   `pe exp hook`

> **Note**: Prototype commands are intentionally unstable. Use `pe exp [command] --help`
> to inspect the currently exposed interface for your build.
