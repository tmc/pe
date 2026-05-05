#!/usr/bin/env bash
set -eu

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=${TMPDIR:-/tmp}/pe-eval-regression-gate-$$

# Keep plugin discovery from treating temporary pe-* binaries as plugins.
export PE_PLUGIN_PATH=

cleanup() {
	rm -rf "$tmp"
}
trap cleanup EXIT

mkdir -p "$tmp"

pe=${PE_BIN:-pe}
if [ "${PE_BIN:-}" ]; then
	ln -sf "$PE_BIN" "$tmp/pe"
	pe=$tmp/pe
fi

"$pe" stats "$dir/current.jsonl" >"$tmp/stats.txt"
grep -q "Total tests: 2" "$tmp/stats.txt"
grep -q "Failures: 1" "$tmp/stats.txt"

if "$pe" diff --fail-on-regression "$dir/baseline.json" "$dir/current.jsonl" >"$tmp/fail.txt" 2>&1; then
	echo "expected regression gate to fail" >&2
	exit 1
fi
grep -q "Gate: fail" "$tmp/fail.txt"
grep -q "pass rate dropped 50.00 percentage points" "$tmp/fail.txt"

"$pe" diff --fail-on-regression --max-pass-rate-drop 60 --max-score-drop 0.6 --max-latency-increase-ms 60 --max-failure-increase 1 "$dir/baseline.json" "$dir/current.jsonl" >"$tmp/pass.txt"
grep -q "Gate: pass" "$tmp/pass.txt"

echo "eval regression gate example passed"
