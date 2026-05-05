#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
OUT=${1:-"$ROOT/completions"}
TMP=$(mktemp -d "${TMPDIR:-/tmp}/pe-completion.XXXXXX")
BIN="$TMP/pe"
trap 'rm -rf "$TMP"' EXIT

mkdir -p "$OUT"
GOTOOLCHAIN=${GOTOOLCHAIN:-go1.25.9} go build -o "$BIN" "$ROOT/cmd/pe"

"$BIN" completion bash >"$OUT/pe.bash"
"$BIN" completion zsh >"$OUT/_pe"
"$BIN" completion fish >"$OUT/pe.fish"
"$BIN" completion powershell >"$OUT/pe.ps1"

printf 'wrote completions to %s\n' "$OUT"
