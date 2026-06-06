# PE Plugins

PE discovers external plugins as executable files named `pe-*` in directories
listed by `PE_PLUGIN_PATH`. A discovered executable named `pe-promptfoo` is
available as `pe promptfoo`; it can also be run explicitly with
`pe plugin run promptfoo`.

The current plugin system is a process boundary. Plugins are separate
executables, not Go shared libraries and not imports of a public PE SDK.

## Discovery

```bash
export PE_PLUGIN_PATH="$HOME/bin"
pe plugin list
```

Discovery scans each directory in `PE_PLUGIN_PATH` for executable files with a
`pe-` prefix. PE does not scan the whole `PATH` during startup.

## Metadata

Plugins may implement `--pe-plugin-info` and print JSON metadata:

```json
{
  "description": "Promptfoo compatibility",
  "version": "0.1.0",
  "commands": [
    {
      "name": "import",
      "description": "Import promptfoo configuration",
      "usage": "pe promptfoo import <config.yaml>"
    }
  ]
}
```

If a plugin does not implement metadata, PE still exposes it by name with a
generic description.

## Running Plugins

```bash
pe plugin run promptfoo import promptfooconfig.yaml -o pe-config.yaml
pe promptfoo import promptfooconfig.yaml -o pe-config.yaml
```

The first form runs a plugin explicitly. The second form is the dynamic top-level
command added for a discovered plugin.

## Promptfoo Plugin

The repository includes a Promptfoo configuration conversion plugin:

```bash
go build -o "$HOME/bin/pe-promptfoo" ./plugins/promptfoo
export PE_PLUGIN_PATH="$HOME/bin"
pe promptfoo --help
```

See [PROMPTFOO_INTEGRATION.md](PROMPTFOO_INTEGRATION.md) for supported
Promptfoo import, export, and conversion behavior.
