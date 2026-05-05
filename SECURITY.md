# Security Policy

## Reporting a Vulnerability

Report security issues privately by opening a GitHub security advisory or by
emailing the repository owner. Do not file public issues for vulnerabilities.

Include:

- Affected version or commit.
- Steps to reproduce.
- Expected and actual impact.
- Any logs or proof of concept needed to confirm the issue.

The project will acknowledge reproducible reports, coordinate fixes privately
when practical, and publish release notes once a fix is available.

## Supported Versions

Security fixes target the current release line and the default branch. Older
development snapshots may receive fixes when the change is low risk.

## Scope

PE executes local tools, provider CLIs, plugins, and evaluation scripts when
configured to do so. Treat prompt configs, modules, plugins, and eval files from
untrusted sources as executable or file-reading input unless the command
explicitly documents a narrower sandbox.

Experimental `pe exp attest` and `pe exp cache` commands are unsigned and
local-only. They can detect local content, manifest, and cache-object tamper,
but they do not prove identity, origin, or freshness.
