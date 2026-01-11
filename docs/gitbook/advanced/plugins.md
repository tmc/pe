# Plugin System

Extend PE with custom providers, tools, and optimizers using Go plugins.

## Architecture

PE uses Go's `plugin` package to load shared libraries (`.so` files) at runtime. Plugins can implement interfaces for:
*   **Providers**: Add support for new APIs.
*   **Optimizers**: Implement custom prompt improvement algorithms.
*   **Validators**: Custom style checks.

## Creating a Plugin

1.  **Define your implementation**:

```go
package main

import "github.com/tmc/pe/pkg/plugin"

type MyProvider struct{}

func (p *MyProvider) Complete(...) (...) {
    // implementation
}

var Plugin = &plugin.ProviderPlugin{
    Name: "my-provider",
    Provider: &MyProvider{},
}
```

2.  **Build**:

```bash
go build -buildmode=plugin -o my-plugin.so
```

3.  **Install**: Place in `~/.pe/plugins/`.

## Configuration

Enable plugins in `~/.pe/config.yaml`:

```yaml
plugins:
  my-provider:
    endpoint: "https://api.custom.com"
```
