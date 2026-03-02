# Basic Evaluation Example

This example demonstrates PE's evaluation capabilities with simple math problems.

## Files
- `config.yaml` - Evaluation configuration with multiple test cases

## Features Demonstrated
- Multiple prompts (different phrasings)
- Multiple providers (OpenAI and Anthropic)
- Various assertion types
- Result saving

## Running the Evaluation

```bash
# Run the evaluation
pe eval config.yaml

# Run with verbose output
pe eval config.yaml --verbose

# Run with specific provider only
pe eval config.yaml --provider openai

# View results in browser
pe eval config.yaml --save-db
pe view
```

## MLX Backend Comparison

Use the included `pe test` config to compare `mlx-lm` and `mlx-go` backends on the same prompts:

```bash
pe test mlx_backends_test.yaml --type property --verbose
```

## Understanding Results

The evaluation will:
1. Test each prompt template
2. With each provider
3. For each test case
4. Check all assertions

Total evaluations = prompts × providers × tests
In this example: 2 × 2 × 4 = 16 evaluations

## Assertion Types Used
- `contains` - Output must contain specific text
- `not-contains` - Output must NOT contain text
- `regex` - Output must match pattern
- `length` - Output length constraints

## Viewing Results

Results are saved to:
- `results.json` - JSON format for processing
- Database (when using `--save-db`) - For web UI viewing
