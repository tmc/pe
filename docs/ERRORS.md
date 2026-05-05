# Error Handling

PE uses structured errors from `internal/errors`. Every PE error carries:

- a stable error code
- a message
- severity
- retryability
- an optional component
- optional context metadata
- an optional wrapped cause

Use `errors.GetCode`, `errors.GetSeverity`, `errors.IsRetryable`,
`errors.GetComponent`, and `errors.GetContext` when reporting failures.
`errors.Suggestion` returns a short user-facing recovery hint.

## Common Codes

| Code | Meaning | Typical next step |
| --- | --- | --- |
| `INVALID_INPUT` | Bad command or API input | Check flags and command help. |
| `INVALID_CONFIG` | Invalid configuration | Run `pe config validate`. |
| `PROVIDER_AUTH` | Provider authentication failed | Check provider credentials and environment variables. |
| `PROVIDER_RATE_LIMIT` | Provider rate limit hit | Retry later or lower concurrency. |
| `NETWORK_TIMEOUT` | Network operation timed out | Check connectivity and retry. |
| `FILE_NOT_FOUND` | Missing file | Check the path. |
| `MODULE_NOT_FOUND` | Missing module | Check module name, version, and registry config. |
| `SECURITY_POLICY` | Safety policy denied an operation | Inspect the file or module policy. |

## Troubleshooting

- For provider failures, check API keys, provider selection, and model name.
- For config failures, run `pe config validate` and inspect the reported key.
- For module failures, run `pe mod verify` and check registry environment values.
- For retryable network or rate-limit errors, retry with lower concurrency.
- For security failures, inspect the declared policy before bypassing it.
