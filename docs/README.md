# PE Documentation

PE is a prompt engineering toolkit inspired by the Go toolchain. This index
points to the current documentation set as of May 2026.

[../ROADMAP.md](../ROADMAP.md) is the source of truth for planned work and
release blockers. [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md) is the
source of truth for coverage numbers.

## Start Here

- [GETTING_STARTED.md](GETTING_STARTED.md) - canonical first install, first prompt, and
  core workflow.
- [gitbook/getting-started/](gitbook/getting-started/) - thin GitBook navigation to the canonical getting-started docs.
- [TUTORIAL.md](TUTORIAL.md) - longer hands-on walkthrough. Advanced chapters
  include experimental command groups and should be checked against `pe --help`
  before automation.
- [CLI_REFERENCE.md](CLI_REFERENCE.md) - current command reference.
- [CLI_HELP_AUDIT.md](CLI_HELP_AUDIT.md) - generated root-help inventory
  reconciliation for the CLI reference.
- [COMMAND_EXAMPLES_GUIDE.md](COMMAND_EXAMPLES_GUIDE.md) - command examples and
  workflow snippets.
- [../examples/README.md](../examples/README.md) - current runnable examples and
  validation commands.
- [INSTALLATION.md](INSTALLATION.md) - installation and build notes.
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - common errors and fixes.
- [../RELEASE_NOTES.md](../RELEASE_NOTES.md) - v0.5.0 release-candidate notes.
- [../CHANGELOG.md](../CHANGELOG.md) - concise release deltas.

## Core Topics

- [TEMPLATE_SYNTAX.md](TEMPLATE_SYNTAX.md) - prompt template syntax.
- [MODULES.md](MODULES.md), [MODULE_REGISTRY.md](MODULE_REGISTRY.md), and
  [module-registry.md](module-registry.md) - module workflows, current registry
  configuration, tidy/vendor behavior, and validation blockers.
- [PLUGINS.md](PLUGINS.md) - plugin system.
- [PROMPTFOO_INTEGRATION.md](PROMPTFOO_INTEGRATION.md) - promptfoo-compatible
  evaluation configuration.
- [EXTERNAL_EXECUTION_POLICY.md](EXTERNAL_EXECUTION_POLICY.md) - allowed
  subprocess boundaries and required coverage.
- [SECURITY_REVIEW.md](SECURITY_REVIEW.md) - release security review notes.
- [SECURITY_AUDIT_2026-05-05.md](SECURITY_AUDIT_2026-05-05.md) - scoped
  security audit for recent config, security, compatibility, and performance work.
- [../SECURITY.md](../SECURITY.md) - vulnerability disclosure policy.
- [observability/GRAFANA_DASHBOARDS.md](observability/GRAFANA_DASHBOARDS.md) -
  dashboard definitions for release metrics.
- [observability/RUNBOOKS.md](observability/RUNBOOKS.md) - operational runbooks
  for provider, evaluation, optimization, memory, and release-gate incidents.
- [MIGRATION.md](MIGRATION.md) - v0.5.0 migration decision.
- [UPGRADING.md](UPGRADING.md) - compatibility shims, migration tooling, and
  breaking-change policy.
- [NOTICE_DECISION.md](NOTICE_DECISION.md) - v0.5.0 NOTICE decision.
- [ATTESTATION.md](ATTESTATION.md) - attestation design and prototype status.

## Architecture And APIs

- [ARCHITECTURE.md](ARCHITECTURE.md) - system architecture.
- [API_REFERENCE.md](API_REFERENCE.md) - package and provider API notes.
- [LOCAL_RUNTIME_PROVIDER_DESIGN.md](LOCAL_RUNTIME_PROVIDER_DESIGN.md) -
  local-runtime provider design.
- [PROVIDER_INTERFACE_AUDIT.md](PROVIDER_INTERFACE_AUDIT.md) - current
  `llm.Provider` and `inference.Provider` migration map.
- [PROVIDER_MIGRATION.md](PROVIDER_MIGRATION.md) - contributor guide for
  provider adapter and command migration work.
- [LLM_CLI_STANDARDS.md](LLM_CLI_STANDARDS.md) - CLI provider conventions.
- [STARLARK_EXTENSION.md](STARLARK_EXTENSION.md) - Starlark extension support.

## Status And Planning

- [CURRENT_STATUS.md](CURRENT_STATUS.md) - current implementation status.
- [TEST_COVERAGE_REPORT.md](TEST_COVERAGE_REPORT.md) - measured coverage
  baseline and low-coverage packages.
- [PERFORMANCE_PROFILE.md](PERFORMANCE_PROFILE.md) - current benchmark profile
  notes and hot-loop follow-ups.
- [RELEASE_BUILD_MATRIX.md](RELEASE_BUILD_MATRIX.md) - latest local
  cross-compilation matrix and binary sizes.
- [RELEASE_DEPENDENCY_REVIEW.md](RELEASE_DEPENDENCY_REVIEW.md) - release
  dependency counts and install-script decision.
- [RELEASE_PROCESS.md](RELEASE_PROCESS.md) - release testing, deployment,
  rollback, and post-release monitoring process.
- [PROJECT_MANAGEMENT.md](PROJECT_MANAGEMENT.md) - project board, sprint,
  review, bug tracking, and validation process.
- [../CONTRIBUTING.md](../CONTRIBUTING.md) - contributor guide.

## Examples

- [../examples/](../examples/) - current runnable examples.
- [../example/](../example/) - legacy exploratory examples kept for
  compatibility, regression tests, and design history. Do not treat this tree
  as release-facing command documentation.

## Future And Archive

- [future/](future/) contains aspirational designs and future-facing guides.
  These documents are not current implementation references unless promoted
  into the main docs.
- [archive/](archive/) contains historical reports, older roadmaps, and
  superseded documentation.
