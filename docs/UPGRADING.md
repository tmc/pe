# Upgrading PE

PE keeps compatibility shims for old command names and configuration shapes where
the migration cost is low. `ROADMAP.md` tracks any planned breaking changes.

## Compatibility Layer

Compatibility aliases should be represented as `internal/compat.Alias` values.
Callers can resolve old names through `compat.Layer` and show users
`compat.DeprecatedWarning` when an alias is used.

## Migration Tools

Migration tools should produce a deterministic `compat.MigrationStep` list:

- `from`: the old name or shape.
- `to`: the new name or shape.
- `note`: the warning or manual action.

Config-file migration is exposed by `pe config migrate`.

## Breaking Changes

For the current release branch, no user-visible breaking command or config changes
are approved. Planned breaking changes must be documented here before release and
cross-linked from `RELEASE_NOTES.md`.

## Compatibility Testing

Compatibility paths need tests that prove:

- Old names resolve to new names.
- Deprecation warnings are emitted or available.
- Migration plans are deterministic.
- Existing promptfoo-compatible and legacy prompt-template paths keep working.
