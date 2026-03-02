#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
	echo "usage: $0 <eval-results.json>" >&2
	exit 2
fi

file="$1"

jq -r '
  .results.results
  | group_by(.provider.id)
  | map({
      provider: .[0].provider.id,
      total: length,
      passed: map(select(.success == true)) | length,
      failed: map(select(.success != true)) | length,
      pass_rate: ((map(select(.success == true)) | length) / length * 100),
      avg_latency_ms: (map(.latencyMs) | add / length),
      p95_latency_ms: (map(.latencyMs) | sort | .[(length * 95 / 100 | floor)])
    })
  | .[]
  | [
      .provider,
      ("total=" + (.total|tostring)),
      ("passed=" + (.passed|tostring)),
      ("failed=" + (.failed|tostring)),
      ("pass_rate=" + (.pass_rate|tostring)),
      ("avg_latency_ms=" + (.avg_latency_ms|tostring)),
      ("p95_latency_ms=" + (.p95_latency_ms|tostring))
    ]
  | @tsv
' "$file"
