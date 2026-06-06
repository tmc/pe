# Module Registry

PE modules are versioned prompt packages. A registry exposes module metadata and
files over either a local directory, GitHub releases, or a static HTTP endpoint.

## Configuration

PE chooses the registry backend from environment variables:

```text
PE_REGISTRY_TYPE=local
PE_REGISTRY_DIR=/path/to/registry

PE_REGISTRY_TYPE=http
PE_REGISTRY_URL=https://example.com/registry
PE_REGISTRY_TOKEN=token

PE_REGISTRY_TYPE=github
PE_REGISTRY_OWNER=owner
PE_REGISTRY_REPO=repo
PE_REGISTRY_TOKEN=token
```

`PE_REGISTRY_TYPE` supports `local`, `http`, and `github`. An empty or
unrecognized value uses the local registry. The default local root is
`$HOME/.pe/registry`; the default HTTP URL is `https://pe.dev/registry`; the
default GitHub repository is `pe-modules/registry`. GitHub also uses
`GITHUB_TOKEN` when `PE_REGISTRY_TOKEN` is unset.

Common failure modes are reported directly by the selected backend:

- Local registry paths must be directories; missing modules and invalid
  metadata are reported as module lookup failures.
- HTTP registries must serve `modules.json` with status `200`; other statuses,
  invalid JSON, missing modules, and missing files fail list, search, get,
  health, or download operations.
- HTTP registries are read-only; publish fails.
- GitHub registries require reachable repository metadata and release assets.
  Publishing requires a token.
- Module names and file paths must stay within the registry or download root;
  absolute paths and `..` escapes fail before file access.

## Metadata

Each module version has a `module.json` file:

```json
{
  "name": "example.com/prompts",
  "version": "v1.2.3",
  "description": "review prompts",
  "author": "example",
  "dependencies": {
    "example.com/base": "^v1.0.0"
  },
  "files": ["review.pe"],
  "checksum": "sha256-hex"
}
```

`name` is a relative module path. `version` is semantic version text, normally
with a `v` prefix. `files` are relative paths inside the module version
directory. Paths must not be absolute and must not escape the module directory.

## HTTP Protocol

A static HTTP registry serves:

```text
/modules.json
/modules/<module>/<version>/module.json
/modules/<module>/<version>/<file>
```

`modules.json` is a JSON array of module metadata. Clients use it for list,
search, upgrade, and metadata lookup. File downloads use the `files` list from
metadata.

If `PE_REGISTRY_TOKEN` is set, the HTTP client sends it as:

```text
Authorization: Bearer <token>
```

## GitHub Registry

The GitHub registry reads releases from `PE_REGISTRY_OWNER` and
`PE_REGISTRY_REPO`. Releases expose a `module.json` asset and file assets for
the module version. If `PE_REGISTRY_TOKEN` is set, it is used. Otherwise
`GITHUB_TOKEN` is used when present.

## Local Registry

A local registry stores modules as:

```text
<root>/<module>/<version>/module.json
<root>/<module>/<version>/<file>
```

The default local root is `$HOME/.pe/registry`. Set `PE_REGISTRY_DIR` to use a
different root.

## Versioning

Version constraints accept exact versions, `>`, `>=`, `<`, `<=`, caret ranges,
tilde ranges, `*`, and `latest`. Resolution compares semantic versions after
removing a leading `v`, prerelease suffix, and build metadata.

## Dependencies

`dependencies` maps module paths to version constraints. The resolver builds a
directed dependency graph, checks for cycles, detects conflicts, and returns a
topological dependency order. Pruning keeps only modules reachable from selected
roots.

## Tidy and Vendor

`pe mod tidy` scans prompt and configuration files for `pe://` module
references. Without `--write`, it reports missing and unused requirements
without changing `pe.mod`. With `--write`, it removes unused requirements and
adds missing requirements only when every reference to that module has one
explicit version. References with no version or conflicting versions are
reported but not added.

`pe mod vendor` copies required modules from `.pe/cache/modules` into
`vendor/` and writes `vendor/modules.txt`. It does not download missing
modules; run `pe mod download` first. Missing cached modules are reported as
warnings and are omitted from `vendor/modules.txt`.

## Cache

Downloaded metadata is cached under `.pe/cache`. Cache entries are addressed by
module path and version. The cache supports targeted invalidation by module
version, module-wide invalidation, and full clearing.

## Integrity

`pe mod verify` recomputes a deterministic SHA-256 digest over downloaded module
files and compares it with `module.json.checksum`. The digest excludes
`module.json`, includes relative file names and file bytes, and rejects
symlinks.

## Signatures

Module signatures use Ed25519 over:

```text
pe module signature v1
<module>
<version>
<checksum>
```

The signature record stores the module name, version, checksum, algorithm,
public-key fingerprint, and hex signature. Trust is module-scoped by public-key
fingerprint.
