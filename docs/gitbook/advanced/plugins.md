# Plugin System

PE plugins are external executables named `pe-*`.

Set `PE_PLUGIN_PATH` to one or more directories containing plugin binaries.
During discovery, PE scans those directories and exposes each executable plugin
as both `pe plugin run <name>` and, when there is no built-in command conflict,
`pe <name>`.

```bash
go build -o "$HOME/bin/pe-promptfoo" ./plugins/promptfoo
export PE_PLUGIN_PATH="$HOME/bin"

pe plugin list
pe promptfoo --help
```

Plugins may implement `--pe-plugin-info` to return JSON metadata. If metadata is
not available, PE still exposes the executable with a generic description.

The current plugin system does not load Go `.so` files and does not expose a
public Go SDK package. Treat plugin boundaries as command-line process
boundaries.
