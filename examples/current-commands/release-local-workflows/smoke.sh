#!/usr/bin/env bash
set -eu

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=${TMPDIR:-/tmp}/pe-release-local-workflows-$$

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

if "$pe" diff --fail-on-regression "$dir/baseline.json" "$dir/current.jsonl" >"$tmp/diff-fail.txt" 2>&1; then
	echo "expected regression gate to fail" >&2
	exit 1
fi
grep -q "Gate: fail" "$tmp/diff-fail.txt"

"$pe" diff --fail-on-regression --max-pass-rate-drop 60 --max-score-drop 0.6 --max-latency-increase-ms 100 --max-failure-increase 1 "$dir/baseline.json" "$dir/current.jsonl" >"$tmp/diff-pass.txt"
grep -q "Gate: pass" "$tmp/diff-pass.txt"

mkdir -p "$tmp/mod"
cp "$dir/pe.mod" "$dir/keep.prompt" "$dir/add.prompt" "$tmp/mod/"
(
	cd "$tmp/mod"
	"$pe" mod tidy --json --write >"$tmp/mod-tidy.json"
)
grep -q '"updated": true' "$tmp/mod-tidy.json"
grep -q '"module": "github.com/example/add"' "$tmp/mod-tidy.json"
grep -q '"version": "v1.2.3"' "$tmp/mod-tidy.json"
grep -q '"github.com/example/remove"' "$tmp/mod-tidy.json"
grep -q 'github.com/example/add v1.2.3' "$tmp/mod/pe.mod"
! grep -q 'github.com/example/remove' "$tmp/mod/pe.mod"

mkdir -p "$tmp/attest"
cp -R "$dir/attest/." "$tmp/attest/"
"$pe" exp attest manifest --root "$tmp/attest" input.txt >"$tmp/manifest.json"
grep -q '"type": "pe.unsigned_file_manifest.v1"' "$tmp/manifest.json"
grep -q '"path": "input.txt"' "$tmp/manifest.json"
"$pe" exp attest verify --root "$tmp/attest" "$tmp/manifest.json" >"$tmp/attest-verify.txt" 2>&1
grep -q "unsigned manifest verified" "$tmp/attest-verify.txt"

key=$("$pe" exp cache manifest put --cache-dir "$tmp/cache" "$tmp/manifest.json" 2>&1)
case "$key" in
	????????????????????????????????????????????????????????????????) ;;
	*)
		echo "unexpected cache key: $key" >&2
		exit 1
		;;
esac
"$pe" exp cache manifest verify --cache-dir "$tmp/cache" --root "$tmp/attest" "$key" >"$tmp/cache-verify.txt" 2>&1
grep -q "cached manifest verified" "$tmp/cache-verify.txt"

echo "release local workflows example passed"
