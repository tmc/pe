# Simple PE Example

This example demonstrates the basic functionality of the `pe` tool with a minimal configuration.

## Configuration

The `config.yaml` file contains:

- 2 prompts with a template variable
- 3 providers (OpenAI, Anthropic, and Google AI)
- 3 test cases with different values for the template variable

## Running the Example

### Dry Run Mode

To see the commands that would be executed without actually running them:

```bash
pe eval config.yaml --dry-run > commands.sh
```

This will generate a script that you can execute:

```bash
chmod +x commands.sh
./commands.sh
```

### Standard Mode

To evaluate the prompts with all providers:

```bash
pe eval config.yaml -o results.json
```

This will run all prompts against all providers with all test cases and save the results to `results.json`.

## Expected Output

In dry run mode, you'll get a shell script with commands like:

```bash
cgpt -b openai -m gpt-4 'What is the capital of France?'
cgpt -b openai -m gpt-4 'What is the capital of Japan?'
...
```

In standard mode, you'll get a JSON file with the results of each evaluation, including whether the assertions passed or failed.