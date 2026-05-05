#!/usr/bin/env bash
set -eu

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
pe=${PE_BIN:-pe}
tmp=${TMPDIR:-/tmp}/pe-current-commands-$$

# Keep plugin discovery from treating temporary pe-* binaries as plugins.
export PE_PLUGIN_PATH=

cleanup() {
	rm -rf "$tmp"
}
trap cleanup EXIT

mkdir -p "$tmp"

PE_TEST_MODE=true "$pe" ask "What is the capital of France?" --provider mock >"$tmp/ask.txt"
"$pe" template list --format json >"$tmp/templates.json"
"$pe" prompt init "$tmp/demo.prompt" --force >"$tmp/prompt-init.txt"
"$pe" prompt info "$tmp/demo.prompt" >"$tmp/prompt-info.txt"
"$pe" plugin list >"$tmp/plugins.txt"
"$pe" profile status >"$tmp/profile.txt"
"$pe" build "$dir/build.yaml" --output "$tmp/built.txt" --minify >"$tmp/build.txt"
"$pe" convert "$dir/convert.yaml" "$tmp/converted.json" >"$tmp/convert.txt"
"$pe" collect --jobs 3 >"$tmp/collect.txt"
"$pe" reduce --sum sum >"$tmp/reduce.txt"
"$pe" watch --help >"$tmp/watch-help.txt"

test -s "$tmp/ask.txt"
test -s "$tmp/templates.json"
test -s "$tmp/demo.prompt"
test -s "$tmp/built.txt"
test -s "$tmp/converted.json"
grep -q "Installed plugins" "$tmp/plugins.txt"
grep -q "Profiler Status" "$tmp/profile.txt"
grep -q "Collected 3 results" "$tmp/collect.txt"
grep -q "30" "$tmp/reduce.txt"
grep -q "Usage:" "$tmp/watch-help.txt"

echo "current command examples passed"
