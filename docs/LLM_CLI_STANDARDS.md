# LLM CLI Standards: Structured Runtime Reporting

This document defines standard practices for reporting token utilization in LLM CLI tools within the PE ecosystem.

## Goals
- **Consistency**: Ensure all tools report usage in a parsable, predictable format.
- **Composability**: Enable pipelines to aggregate costs and usage tracking.
- **Visibility**: Allow users to see costs without breaking standard output streams.

## Standard Data Structure

Tools should output runtime data using the following JSON schema:

```json
{
  "output": "The response string...",
  "prompt_tokens": 50,
  "completion_tokens": 100,
  "total_tokens": 150,
  "latency_ms": 321,
  "cost": 0.002,
  "metrics": {
    "tokens_per_second": 311.5,
    "load_duration_ms": 40
  },
  "tokenUsage": {
    "total": 150,
    "prompt": 50,
    "completion": 100
  }
}
```

### Fields
- `output` (string): Generated output text.
- `prompt_tokens` (integer, optional): Tokens used in the input prompt.
- `completion_tokens` (integer, optional): Tokens generated in the response.
- `total_tokens` (integer, optional): Total token count.
- `latency_ms` (integer, optional): End-to-end latency in milliseconds.
- `cost` (float, optional): Estimated cost in USD.
- `metrics` (object, optional): Runtime-specific metrics such as throughput or load time.
- `tokenUsage` / `token_usage` (object, optional): Alternative Promptfoo/OpenAI-compatible token usage object.

## Integration Patterns

### 1. JSON Output Requirement
When a tool is invoked with a `--json` flag, the output **MUST** be a valid JSON object containing the response and any available token or latency metrics.

### 2. Standard Error (stderr) Reporting
For tools outputting raw text to `stdout` (pipeline mode), token usage info **SHOULD** be printed to `stderr` if specifically requested or if verbose logging is enabled.

**Flag:** `--show-usage` or `--verbose` / `-v`

**Format:**
```text
Token Usage: prompt=50 completion=100 total=150 cost=$0.002
```

### 3. Environment Variables
Tools SHOULD respect `PE_PRINT_USAGE=true` to force printing usage statistics to stderr, aiding debugging in complex pipelines.

## Implementation in PE

The `pe` tool and its providers adhere to these standards:
- Internal structures match the JSON schema.
- `pe run` supports `--json` for structured output.
- CLI-backed providers are expected to map their native usage data to this standard structure.
