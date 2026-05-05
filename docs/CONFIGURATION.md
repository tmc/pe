# Configuration

PE loads configuration in this order, from lowest to highest priority:

1. Built-in defaults from `internal/config.DefaultConfig`.
2. Configuration files discovered by `internal/config.ConfigManager`.
3. Environment variables using the configured prefix, `PE` by default.
4. CLI overrides passed to the manager.

The supported file formats are YAML and JSON. Unknown extensions are parsed as
YAML. TOML is reserved in the schema as an output/config format name, but the
manager does not parse TOML files yet.

Default discovery paths include project-local files such as `pe.yaml`,
`pe.config.yaml`, `.pe/config.yaml`, user config under `~/.pe/config.yaml`, and
system config under `/etc/pe/config.yaml`.

The schema is defined in `internal/config/schema.go`. It covers application,
provider, evaluation, module, optimization, observability, security, plugin,
template, and output settings. `ValidateConfig` checks required values, known
enums, positive sizes, valid timeouts, and cache settings.

`ConfigManager` supports:

- explicit search paths with `WithConfigPaths`
- environment prefix changes with `WithEnvPrefix`
- CLI override maps with `WithCLIOverrides`
- file watching and hot reload callbacks with `WithWatcher`
- saving YAML or JSON with `SaveToFile`

## Environment Variables

The manager derives environment variable names from exported schema field names.
For nested fields it joins struct names with underscores and prepends the manager
prefix. With the default `PE` prefix, examples are:

| Config field | Environment variable |
| --- | --- |
| `App.Verbose` | `PE_APP_VERBOSE` |
| `App.Debug` | `PE_APP_DEBUG` |
| `App.LogLevel` | `PE_APP_LOGLEVEL` |
| `Providers.Default` | `PE_PROVIDERS_DEFAULT` |
| `Eval.MaxConcurrency` | `PE_EVAL_MAXCONCURRENCY` |
| `Modules.Registry` | `PE_MODULES_REGISTRY` |
| `Optimization.DefaultMethod` | `PE_OPTIMIZATION_DEFAULTMETHOD` |
| `Output.Format` | `PE_OUTPUT_FORMAT` |

Values are parsed by destination type. Booleans use Go boolean syntax, integers
and floats use base-10 text, durations use Go duration strings such as `30s`,
string slices use comma-separated values, and string-keyed maps use JSON.

Precedence is fixed: CLI overrides win over environment variables, environment
variables win over config files, and config files win over built-in defaults.
Invalid environment values fail manager loading instead of being ignored.
