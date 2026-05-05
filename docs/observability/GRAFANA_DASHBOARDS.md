# Grafana Dashboards

PE exposes metrics names that can be scraped from the Prometheus text exporter in
`internal/observability`. This file defines the release dashboard set. Keep panel
queries aligned with the metric names in code.

## Provider Health

Panels:

- `provider_request_count`: request volume by `provider` and `model`.
- `provider_error_count`: provider failures by `provider` and `model`.
- `provider_request_latency_ms`: request latency distribution.
- `provider_token_usage`: token volume by `provider` and `model`.
- `provider_available`: provider availability gauge.

Alerts:

- Provider error rate above 10% for 5 minutes.
- Provider availability below 1 for 2 minutes.
- p95 provider latency above 30 seconds for 5 minutes.

## Optimization Runs

Panels:

- `optimization_duration_ms`: optimization duration distribution.
- `optimization_iteration_count`: iteration count by method.
- `optimization_converged`: convergence status by method.
- `optimization_improvement_score`: final improvement score.
- `optimization_resource_usage`: resource usage gauge.
- `optimization_cancellation_count`: cancellations by method.

Alerts:

- Cancellation count increases for 10 minutes.
- Median optimization duration doubles relative to the previous release baseline.

## Evaluation Quality

Panels:

- `evaluation_latency_ms`: evaluation latency distribution.
- `evaluation_pass_rate`: pass rate by suite.
- `evaluation_assertion_count`: assertions executed by suite.
- `evaluation_parallelism`: configured parallelism by suite.
- `evaluation_score`: score distribution.

Alerts:

- Evaluation pass rate below 95% on release suites.
- Evaluation latency p95 above 30 seconds.

## Runtime And Errors

Panels:

- Error totals by code and severity from the error dashboard output.
- Heap allocation, heap objects, goroutine count, and GC CPU fraction from profile
  analysis snapshots.
- Log volume by level from the log aggregator.

Alerts:

- Any critical error is reported.
- Goroutine count exceeds the previous baseline by 2x.
- GC CPU fraction stays above 10% for 10 minutes.
