# PE Documentation

PE is a prompt engineering toolkit inspired by the Go toolchain. This index
points to the current documentation set as of May 2026.

[../ROADMAP.md](../ROADMAP.md) is the source of truth for planned work and
release blockers. [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md) is the
source of truth for coverage numbers.

## Start Here

- [GETTING_STARTED.md](GETTING_STARTED.md) - first install, first prompt, and
  core workflow.
- [TUTORIAL.md](TUTORIAL.md) - longer hands-on walkthrough. Advanced chapters
  include experimental command groups and should be checked against `pe --help`
  before automation.
- [CLI_REFERENCE.md](CLI_REFERENCE.md) - current command reference.
- [COMMAND_EXAMPLES_GUIDE.md](COMMAND_EXAMPLES_GUIDE.md) - command examples and
  workflow snippets.
- [INSTALLATION.md](INSTALLATION.md) - installation and build notes.
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - common errors and fixes.

## Core Topics

- [TEMPLATE_SYNTAX.md](TEMPLATE_SYNTAX.md) - prompt template syntax.
- [MODULES.md](MODULES.md) and [MODULE_REGISTRY.md](MODULE_REGISTRY.md) -
  module workflows.
- [PLUGINS.md](PLUGINS.md) - plugin system.
- [PROMPTFOO_INTEGRATION.md](PROMPTFOO_INTEGRATION.md) - promptfoo-compatible
  evaluation configuration.
- [SECURITY_REVIEW.md](SECURITY_REVIEW.md) - release security review notes.
- [ATTESTATION.md](ATTESTATION.md) - attestation design and prototype status.

## Architecture And APIs

- [ARCHITECTURE.md](ARCHITECTURE.md) - system architecture.
- [API_REFERENCE.md](API_REFERENCE.md) - package and provider API notes.
- [LOCAL_RUNTIME_PROVIDER_DESIGN.md](LOCAL_RUNTIME_PROVIDER_DESIGN.md) -
  local-runtime provider design.
- [LLM_CLI_STANDARDS.md](LLM_CLI_STANDARDS.md) - CLI provider conventions.
- [STARLARK_EXTENSION.md](STARLARK_EXTENSION.md) - Starlark extension support.

## Status And Planning

- [CURRENT_STATUS.md](CURRENT_STATUS.md) - current implementation status.
- [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md) - measured coverage
  baseline and low-coverage packages.
- [IMPLEMENTATION_TODOS.md](IMPLEMENTATION_TODOS.md) - tombstone pointing to
  [../ROADMAP.md](../ROADMAP.md).
- [PLANNED_COMMANDS.md](PLANNED_COMMANDS.md) - aspirational command ideas, not
  current command status.

## Examples

- [../examples/](../examples/) - current runnable examples.
- [../example/](../example/) - legacy example package.

## Future And Archive

- [future/](future/) contains aspirational designs and future-facing guides.
  These documents are not current implementation references unless promoted
  into the main docs.
- [archive/](archive/) contains historical reports, older roadmaps, and
  superseded documentation.
