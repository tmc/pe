#!/usr/bin/env bash
set -eu

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=${TMPDIR:-/tmp}/pe-distributed-local-$$

# Keep plugin discovery from treating temporary pe-* binaries as plugins.
export PE_PLUGIN_PATH=

cleanup() {
	rm -rf "$tmp"
}
trap cleanup EXIT

mkdir -p "$tmp"

pe=${PE_BIN:-pe}
if [ "${PE_BIN:-}" ]; then
	cp "$PE_BIN" "$tmp/pe"
	chmod +x "$tmp/pe"
	pe=$tmp/pe
fi

"$pe" exp distributed --workers 2 --format json "$dir/tasks.json" >"$tmp/results.json"
grep -q '"name": "slow"' "$tmp/results.json"
grep -q '"output": "first"' "$tmp/results.json"
grep -q '"name": "fast"' "$tmp/results.json"
grep -q '"output": "second"' "$tmp/results.json"

"$pe" exp distributed --format text "$dir/tasks.json" >"$tmp/results.txt"
grep -q $'slow\tfirst' "$tmp/results.txt"
grep -q $'fast\tsecond' "$tmp/results.txt"

if "$pe" exp distributed --workers 0 "$dir/tasks.json" >"$tmp/worker.err" 2>&1; then
	echo "expected worker validation failure" >&2
	exit 1
fi
grep -q "workers must be at least 1" "$tmp/worker.err"

echo "distributed local example passed"
