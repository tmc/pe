# Experimental Features

PE includes a small suite of experimental entrypoints under `pe exp`. These
commands are prototypes and may change between releases.

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

> **Note**: Prototype commands are intentionally unstable. Use `pe exp [command] --help`
> to inspect the currently exposed interface for your build.
