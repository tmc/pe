# Release Build Matrix

Last run: 2026-05-05
Branch: `exp`
Commit: `29da4a9`

Command shape:

```bash
GOTOOLCHAIN=go1.25.9 CGO_ENABLED=0 GOOS=<goos> GOARCH=<goarch> \
  go build -trimpath \
  -ldflags "-X main.Version=v0.5.0-rc -X main.Commit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o /tmp/pe-build-<goos>-<goarch> ./cmd/pe
```

Results:

| Target | Status | Bytes | Artifact |
| --- | --- | ---: | --- |
| darwin/arm64 | PASS | 17,221,810 | `/tmp/pe-build-darwin-arm64` |
| darwin/amd64 | PASS | 18,324,928 | `/tmp/pe-build-darwin-amd64` |
| linux/amd64 | PASS | 17,952,315 | `/tmp/pe-build-linux-amd64` |
| linux/arm64 | PASS | 16,685,544 | `/tmp/pe-build-linux-arm64` |
| windows/amd64 | PASS | 18,356,224 | `/tmp/pe-build-windows-amd64.exe` |

Notes:

- Builds used `CGO_ENABLED=0` for release portability.
- Artifacts were written to `/tmp` and are not tracked.
- This is a local matrix run, not a GitHub Actions dry run.
