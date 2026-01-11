# System Overview

PE is designed with modularity, security, and performance in mind.

## High-Level Layers

1.  **CLI Layer** (`cmd/pe`): User-facing command parsing and output formatting.
2.  **Core Engine**:
    *   **Provider Abstraction**: A unified interface (`internal/providers`) for all LLMs.
    *   **Evaluation Engine**: Logic for running tests and verifying assertions.
    *   **Optimization Engine**: implementations of PE2, APEX, etc.
3.  **Storage Layer**:
    *   **Git-like VCS**: Internal versioning for prompt history.
    *   **Content-Addressable Cache**: Efficient caching of LLM responses.

## Key Design Principles

*   **Modularity**: Clear separation of concerns.
*   **Security**: Sandboxing and cryptographic verification are built-in.
*   **Simplicity**: Go-native implementation without heavy runtime dependencies.

See `docs/ARCHITECTURE.md` in the source repository for detailed diagrams.
