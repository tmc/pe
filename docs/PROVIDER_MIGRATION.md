# Provider Migration

PE is moving provider execution to `internal/inference.Provider`.
`internal/llm.Provider` remains a compatibility interface while evaluator and
optimization packages migrate.

## New Command Code

New command code should create an `inference.Client` and register providers by
spec:

```go
client := inference.NewClient()
if _, err := inference.RegisterProviderSpec(client, spec, nil); err != nil {
	return err
}
resp, err := client.CompleteWith(ctx, spec, inference.Request{Prompt: prompt})
```

Use the provider spec as the client key. A spec can be a short provider name,
such as `cgpt`, or a provider/model pair, such as `openai:gpt-4o-mini`.

## Bridging Existing Code

Use the adapters in `internal/inference/migration.go` at package boundaries:

- `inference.MigrateProvider(p)` adapts an `llm.Provider` to
  `inference.Provider`.
- `inference.GetLegacyProvider(p, model)` adapts an `inference.Provider` to
  `llm.Provider`.
- `inference.RegisterProviderSpec(client, spec, config)` creates and registers
  a provider from a provider spec.

Do not add new direct calls to `llm.GetProvider` from command code. If a package
still requires `llm.Provider`, keep the legacy type at that package boundary and
adapt at the caller.

## Option Names

Preserve these option names when crossing the interface boundary:

| Legacy option | Inference field or option |
| --- | --- |
| `temperature` | `Request.Temperature` |
| `max_tokens` | `Request.MaxTokens` |
| `model` | `Request.Model` |
| `system_prompt` | `Request.SystemPrompt` |
| `top_p` | `Request.Options["top_p"]` |
| `top_k` | `Request.Options["top_k"]` |

Unknown provider options should stay in the options map. Do not silently drop
provider-specific fields.

## Migration Order

1. Migrate command entry points to `inference.Client`.
2. Keep evaluator and optimizer internals on `llm.Provider` until their tests
   have explicit adapter coverage.
3. Pick one canonical implementation for each provider.
4. Delete duplicate legacy provider code only after script tests and live smoke
   tests pass.

## Required Tests

Each migrated path needs tests for:

- provider spec registration;
- request option mapping;
- response token, cost, cache, and metadata mapping;
- streaming behavior when the command exposes streaming;
- error propagation from provider construction and execution.
