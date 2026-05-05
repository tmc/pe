# Release Process

This process applies to the current `exp` release branch. Do not tag from another
branch unless the maintainer explicitly changes the release branch.

## Version Planning

- Pick the target version and record the release branch.
- Confirm `ROADMAP.md` has no unchecked release-blocking items for the target.
- Confirm known limitations are listed in `RELEASE_NOTES.md`.
- Confirm compatibility and migration notes are current in `docs/MIGRATION.md`.

## Release Testing

Run these gates before tagging:

```sh
GOTOOLCHAIN=go1.25.9 go test ./...
GOTOOLCHAIN=go1.25.9 go vet ./...
make coverage-check
make security
make bench
```

Also run targeted smoke tests for changed command areas. For provider changes, use
mock-provider tests first and live-provider tests only when credentials are
available.

## Release Notes

- Update `RELEASE_NOTES.md` with features, fixes, known limitations, and validation
  commands.
- Keep release notes factual and tied to committed behavior.
- Do not claim roadmap items are complete unless `ROADMAP.md` and tests agree.

## Deployment

- Create the version tag from the release branch.
- Let `.github/workflows/release.yml` build archives and checksums.
- Verify artifact names, checksums, and the embedded version.
- Publish only after the dry-run artifacts match the expected release matrix.

## Rollback

- If a release artifact is bad, delete or supersede the GitHub release before
  announcing it.
- If a tag is wrong and has not been consumed, delete and recreate it only with
  maintainer approval.
- If users may have consumed the tag, publish a patch release instead of rewriting
  public history.

## Post-Release Monitoring

- Watch CI on the release tag and the next branch update.
- Watch issue reports for install, provider, module, and evaluation regressions.
- Check the observability runbooks for provider, evaluation, optimization, memory,
  and release-gate incidents.
- Record follow-up work in `ROADMAP.md`; do not create Beads issues.
