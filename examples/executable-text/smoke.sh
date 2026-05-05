#!/usr/bin/env bash
set -eu

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=${TMPDIR:-/tmp}/pe-executable-text-$$

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

"$pe" cat --raw "$dir/plain.prompt" >"$tmp/plain.txt"
grep -q "Summarize the incident report" "$tmp/plain.txt"

"$pe" cat "$dir/templated.prompt" --set artifact=runbook --set audience=operator >"$tmp/templated.txt"
grep -q "Review runbook for operator" "$tmp/templated.txt"
grep -q "The next safe action" "$tmp/templated.txt"

grep -q "workflow: release-note-review" "$dir/workflow.pe.yaml"
grep -q "providers allow local" "$dir/pe.mod"
grep -q "cannot loosen" "$dir/pe.mod"

echo "executable text example passed"
