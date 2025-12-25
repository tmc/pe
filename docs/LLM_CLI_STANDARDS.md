# LLM CLI Standards: Token Usage Reporting

This document defines standard practices for reporting token utilization in LLM CLI tools within the PE ecosystem.

## Goals
- **Consistency**: Ensure all tools report usage in a parsable, predictable format.
- **Composability**: Enable pipelines to aggregate costs and usage tracking.
- **Visibility**: Allow users to see costs without breaking standard output streams.

## Standard Data Structure

Tools should output token usage data using the following JSON schema (compatible with Promptfoo and OpenAI formats):

```json
{
  "tokenUsage": {
    "total": 150,
    "prompt": 50,
    "completion": 100,
    "cached": 0,
    "details": {
      "reasoning": 0
    }
  },
  "cost": 0.002
}
```

### Fields
- `total` (integer): Total tokens used.
- `prompt` (integer): Tokens used in the input prompt.
- `completion` (integer): Tokens generated in the response.
- `cached` (integer, optional): Tokens retrieved from cache.
- `details` (object, optional): Breakdown of specific token types (e.g., reasoning tokens).
- `cost` (float, optional): Estimated cost in USD.

## Integration Patterns

### 1. JSON Output Requirement
When a tool is invoked with a `--json` flag, the output **MUST** be a valid JSON object containing the response. Token usage data should be included in a top-level `usage` or `tokenUsage` field, or wrapped in a metadata object.

**Recommended Wrapper Format:**
```json
{
  "output": "The response string...",
  "tokenUsage": { ... },
  "cost": 0.001
}
```

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
- Providers (like `cgpt`, `llm` wrapper) are expected to map their native usage data to this standard structure.
