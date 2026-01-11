# Prompt Modules

PE includes a module system inspired by Go modules (`go mod`), allowing you to share, version, and reuse prompts easily.

## Overview

*   **Registry**: Decentralized (based on GitHub Gists).
*   **Versioning**: Semantic versioning (e.g., `v1.0.0`).
*   **Dependency Management**: `pe.mod` file tracks dependencies.

## Key Commands

### Initialize a Module
```bash
pe mod init github.com/myorg/my-prompts
```

### Install a Module
```bash
pe mod get github.com/user/repo@v1.0.0
```

### Run from Registry
You can run prompts directly without installing:
```bash
pe run github.com/user/repo/translate.prompt@latest --var text="Hello"
```

## Module Structure

A module is simply a repository (or Gist) containing:
1.  `pe.mod`: Metadata and dependencies.
2.  `*.prompt`: Prompt files.
3.  `pe-config.yaml` (optional): Default configurations.

## Publishing

Currently, publishing is manual via Git or by pushing to Gists.
```bash
git tag v0.1.0
git push origin v0.1.0
```
