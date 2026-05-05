# Project Management

`ROADMAP.md` is the source of truth for PE work tracking. Beads is deprecated for
this repository.

## Project Board

Use these columns:

- Backlog: unchecked `ROADMAP.md` items not selected for the current sprint.
- Ready: items with known files, tests, and acceptance criteria.
- In Progress: one owner actively implementing.
- Review: code is committed on a branch and awaiting review.
- Done: committed, tested, and checked off in `ROADMAP.md`.

## Sprint Planning

Sprint plans should contain:

- Goal: one concrete release or roadmap outcome.
- Scope: exact roadmap rows included.
- Non-goals: rows explicitly not included.
- Gates: commands that must pass before the sprint closes.
- Risks: migrations, provider behavior, or docs that need extra review.

## Milestones And Velocity

Milestones map to roadmap phases or release candidates. Track velocity by counting
completed roadmap rows only after the implementation and verification evidence are
committed.

## Stakeholder Updates

Stakeholder updates should include:

- Current branch and latest commit.
- Completed roadmap rows.
- Failed or skipped gates.
- Remaining blockers.
- Next proposed slice.

## Dependency Management

Dependency changes require:

- `go mod verify`.
- `govulncheck ./...`.
- Review for new transitive dependencies.
- No direct edits to `go.sum`; let Go tooling update it.

## Review Process

Review should check:

- API shape and package boundaries.
- Tests for exported behavior.
- Roadmap checkbox evidence.
- No staged binaries or ignored legacy state.
- No `.beads` changes.

## Test Plans

Every implementation slice needs a test plan with:

- Focused package tests.
- Any command tests for CLI-visible changes.
- Regression commands for touched workflows.
- Explicit skipped gates and why they were skipped.

## Bug Tracking

Bugs should be recorded in `ROADMAP.md` under the nearest owning phase unless the
maintainer asks for another tracker. Include reproduction steps, expected behavior,
actual behavior, and the intended verification command.

## Validation Reviews

Performance validation:

- Run `make bench` or a focused benchmark subset.
- Compare against the previous release baseline on the same machine class.

Security review:

- Run `make security`.
- Review new input, file path, provider, credential, and shell execution paths.

Documentation review:

- Check `docs/README.md` links.
- Keep implementation status tied to current code, not future docs.
- Move aspirational material to `docs/future/` until implemented.
