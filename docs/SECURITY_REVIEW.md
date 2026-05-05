# Security Review

Date: 2026-05-05 PDT

Scope: release-prep security review for secrets, command execution, file/path use,
network/provider boundaries, dependency vulnerabilities, and security documentation.
This document was refreshed on 2026-05-05 after the GitHub HTTP timeout fix and
Go 1.25.9 vulnerability gate.

## Commands

```sh
date '+%Y-%m-%d %Z'
sed -n '138,162p' ROADMAP.md
find . \( -path './.git' -o -path './.beads' \) -prune -o -iname 'security.md' -print -o -iname 'README_SECURITY.md' -print | sort
sed -n '1,220p' README_SECURITY.md
rg -n --hidden --glob '!.git/**' --glob '!.beads/**' --glob '!go.sum' --glob '!plugins/**/go.sum' --glob '!*.png' --glob '!*.jpg' --glob '!*.jpeg' --glob '!*.gif' --glob '!*.pdf' --glob '!*.zip' --glob '!*.tar' --glob '!*.gz' --glob '!*.woff*' --glob '!*.ttf' --glob '!node_modules/**' -i '(api[_-]?key|secret|token|password|passwd|credential|authorization: bearer|private[_-]?key|client[_-]?secret)\s*[:=]\s*["'\'']?[^"'\''[:space:]]+' .
rg -n --hidden --glob '!.git/**' --glob '!.beads/**' --glob '!go.sum' --glob '!plugins/**/go.sum' --glob '!*.png' --glob '!*.jpg' --glob '!*.jpeg' --glob '!*.gif' --glob '!*.pdf' --glob '!*.zip' --glob '!*.tar' --glob '!*.gz' --glob '!*.woff*' --glob '!*.ttf' --glob '!node_modules/**' '(sk-[A-Za-z0-9_-]{20,}|AIza[0-9A-Za-z_-]{35}|AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9_]{20,}|-----BEGIN [A-Z ]*PRIVATE KEY-----)' .
rg -n -g '*.go' 'os/exec|exec\.Command|exec\.CommandContext|syscall\.Exec|CombinedOutput\(|\.Output\(|\.Run\(' cmd internal plugins tests example examples
rg -n -g '*.go' 'net/http|http\.Client|http\.DefaultClient|http\.(Get|Post|NewRequest|NewRequestWithContext)|NewRequestWithContext|url\.Parse|ListenAndServe|Authorization|Bearer' cmd internal plugins tests example examples
rg -n -g '*.go' 'filepath\.(Clean|Join|Abs|Rel|Base|Dir)|os\.(Open|OpenFile|Create|ReadFile|WriteFile|Mkdir|MkdirAll|Remove|RemoveAll|Rename|Stat)|fs\.ValidPath|http\.Dir' cmd internal plugins tests example examples
govulncheck ./...
govulncheck -show verbose ./...
GOTOOLCHAIN=go1.25.9 govulncheck ./...
gosec ./...
gosec -fmt=json -no-fail ./... >/tmp/pe-gosec.json 2>/tmp/pe-gosec.err
jq -r '.Stats | to_entries[] | "\(.key)=\(.value)"' /tmp/pe-gosec.json
jq -r '.Issues | group_by(.rule_id)[] | "\(.[0].rule_id)\t\(length)\t\(.[0].details)"' /tmp/pe-gosec.json | sort
GOTOOLCHAIN=go1.25.9 go test ./cmd/pe ./internal/module ./tests -run 'TestSafeConfigPathRejectsTraversal|TestModuleFilePathRejectsTraversal|TestLocalRegistryRejectsTraversal|TestGitHubRegistryDownloadRejectsTraversal|TestCache_RejectsTraversal|TestUnsignedManifestRejectsEscapesAndSymlinks|TestCacheRejectsUnsafeKeysAndSymlinks|TestCacheManifestRejectsUnsafePaths|TestScripts/security_untrusted' -count=1
GOTOOLCHAIN=go1.25.9 go test ./internal/providers -run 'TestGenericCLIProvider' -count=1
```

## Findings

- Secrets: broad secret-name scan returned 119 matches and common live-key
  pattern scan returned 5 matches. Reviewed matches were placeholders, CI secret
  references, env var reads, or test fixtures. No live-looking hardcoded secret
  was identified.
- Local key material: `.pe/keys` had no files, and no `.pe` key files were
  tracked by git in this checkout.
- Security docs: `SECURITY.md` is now the current vulnerability disclosure
  policy.
- Vulnerabilities: `govulncheck ./...` found called standard-library
  vulnerabilities when run with the default Go 1.24.13 toolchain:
  GO-2026-4947, GO-2026-4946, GO-2026-4870, GO-2026-4602, and GO-2026-4601.
  The listed fixes are in Go 1.25.8 or Go 1.25.9. Rerunning with
  `GOTOOLCHAIN=go1.25.9 govulncheck ./...` reported no vulnerabilities.
- Non-called vulnerabilities: verbose govulncheck also reported non-called
  package/module findings GO-2026-4869, GO-2026-4864, GO-2026-4865, and
  GO-2026-4603, all in the standard library.
- Static analysis: `gosec` scanned 135 files and 55,041 lines, reporting 324
  issues. Rule counts: G304 88, G306 86, G104 54, G115 29, G404 26, G301 24,
  G204 13, G302 2, G112 1, G114 1. `gosec` also reported SSA/build errors for
  `plugins/starlark/main.go`, so coverage is not complete.
- Command execution: command execution is concentrated in CLI providers, cgpt
  adapters, plugin execution, metrics helpers, and scripttests. Most calls use
  `exec.Command` or `exec.CommandContext` with argv rather than an explicit
  shell, which limits shell injection. Generic CLI command templates now quote
  prompt data before shell-style splitting, so prompt text cannot add argv
  entries. Risk remains where trusted local config controls executable names,
  command templates, plugin discovery paths, or cgpt options.
- File/path use: many commands intentionally read and write user-supplied paths.
  Config expansion, module registry/cache paths, module publish paths,
  unsigned manifests, and local cache objects have containment checks and
  focused traversal tests. Broad CLI file reads remain intended behavior when
  users explicitly pass local paths.
- Network/providers: OpenAI and Anthropic providers use HTTPS defaults and
  client timeouts, but `baseURL` is configurable. GitHub gist calls in `cmd/pe`
  use an explicit 30-second timeout client. The `pe serve` local HTTP server
  sets `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`.
- Attest/cache: experimental `pe exp attest` and `pe exp cache` workflows are
  unsigned and local-only. They detect local content, manifest, and cache-object
  tamper, but they do not prove identity, origin, or freshness.
- Error disclosure: provider and command wrappers sometimes return remote API
  bodies, subprocess stderr, or generated output in errors. This is useful for
  debugging but can leak prompts, API response details, or secrets into logs.

## Follow-up risks

- Release vulnerability checks should use Go 1.25.9 or newer. The default
  Go 1.24.13 toolchain still reports called standard-library vulnerabilities.
- G204 triage: plugin execution is now limited to `PE_PLUGIN_PATH` discovery and
  explicit plugin runs instead of PATH-wide startup execution. Generic CLI,
  cgpt, custom metric, and scripttest subprocesses remain intended behavior for
  trusted local configuration or test fixtures. Generic CLI prompt interpolation
  is quoted before argv splitting and covered by
  `TestGenericCLIProvider_CommandTemplateQuotesPrompt`.
- G304 triage: config expansion, Starlark `load_tests`, module cache paths, and
  module publish prompt paths now have containment checks. Unsigned manifests
  and cache objects reject path escapes, unsafe keys, and symlinks. Broad CLI
  file reads remain intended behavior when users pass local paths.
- Representative untrusted-input CLI checks live in
  `tests/testdata/script/security_untrusted.txt`.
- GitHub API calls now use an explicit 30-second timeout client instead of
  `http.DefaultClient`; `TestGitHubHTTPClientHasTimeout` covers the setting.
  Keep `pe serve` timeout coverage covered by command-level tests or review
  checks.
- Provider API/body/stderr error propagation now redacts common API keys, bearer
  tokens, GitHub tokens, AWS access keys, and Google API keys before returning
  diagnostics.
- `SECURITY.md` is now the current vulnerability disclosure policy.
- Keep attest/cache docs and help text clear that unsigned local manifests and
  cache entries detect tamper only; they are not identity, origin, or freshness
  proofs.
