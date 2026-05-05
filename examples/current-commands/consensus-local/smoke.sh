#!/usr/bin/env bash
set -eu

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=${TMPDIR:-/tmp}/pe-consensus-local-$$

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

"$pe" exp consensus --input "$dir/votes.json" --output "$tmp/consensus.json"
grep -q '"Output": "south"' "$tmp/consensus.json"
grep -q '"Weight": 3' "$tmp/consensus.json"
grep -q '"Total": 5' "$tmp/consensus.json"
grep -q '"provider": "timeout"' "$tmp/consensus.json"
grep -q '"used": false' "$tmp/consensus.json"

"$pe" exp consensus --input - <"$dir/votes.json" >"$tmp/stdout.json"
grep -q '"Output": "south"' "$tmp/stdout.json"

echo "consensus local example passed"
