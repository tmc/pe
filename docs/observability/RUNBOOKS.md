# Operations Runbooks

These runbooks map PE's current metrics, logs, traces, and profile tools to
operator actions. They are intentionally short so they can be followed during a
release or incident.

## Provider Outage

Signals:

- `provider_available` is `0`.
- `provider_error_count` increases for one provider.
- Error dashboard shows `PROVIDER_UNAVAILABLE`, `PROVIDER_AUTH`, or
  `PROVIDER_RATE_LIMIT`.

Actions:

- Check provider credentials and model name in `pe config list`.
- Run `pe config validate`.
- Retry with another configured provider.
- If the failing provider is external, verify its status page before changing PE.

## Evaluation Regression

Signals:

- `evaluation_pass_rate` drops below the release baseline.
- `evaluation_score` distribution shifts down.
- Error dashboard shows `EVALUATION_FAILED` or `ASSERTION_FAILED`.

Actions:

- Re-run the failing suite with `GOTOOLCHAIN=go1.25.9 go test ./tests/integration -count=1`.
- Compare prompts and provider configuration against the last passing commit.
- If the regression depends on a live provider, reproduce with a mock provider before
  changing assertions.

## Optimization Regression

Signals:

- `optimization_converged` is `0` for a previously stable method.
- `optimization_cancellation_count` increases.
- `optimization_duration_ms` exceeds the release baseline.

Actions:

- Re-run the optimizer package tests.
- Capture a profile with `pe profile start --type cpu,memory` before changing hot
  loops.
- Check cancellation paths before increasing iteration limits.

## High Memory Or Goroutine Count

Signals:

- Profile analysis reports `high heap object count`.
- Profile analysis reports `high goroutine count`.
- GC CPU fraction stays above 10%.

Actions:

- Run `pe profile status` and stop any stale profiles.
- Capture heap and goroutine profiles.
- Inspect recent changes that added goroutines, caches, or retained prompt data.

## Release Gate Failure

Signals:

- `make coverage-check` fails.
- `make security` fails.
- `make bench` fails or shows a large regression.

Actions:

- Treat the failing gate as authoritative until reproduced locally.
- Do not lower `COVERAGE_MIN` to ship a release; either add tests or document an
  explicit maintainer exception.
- For benchmark regressions, compare against the previous release commit with the
  same Go version and machine class.
