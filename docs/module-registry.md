# Module Registry

PE modules are versioned prompt packages. A registry exposes module metadata and
files over either a local directory, GitHub releases, or a static HTTP endpoint.

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
