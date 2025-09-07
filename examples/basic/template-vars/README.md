# Template Variables Example

This example shows how to use Go template variables in prompts.

## Files
- `translate.prompt` - Translation prompt with variables
- `config.yaml` - Evaluation config with multiple test cases

## Template Syntax

PE uses Go's template syntax. Variables must be prefixed with a dot:
- ✅ Correct: `{{.variable}}`
- ❌ Wrong: `{{variable}}`

## Running the Example

### Command Line Usage
```bash
# Use default values
pe run translate.prompt --provider openai

# Override variables
pe run translate.prompt \
  --var text="Good morning!" \
  --var source_lang="English" \
  --var target_lang="French" \
  --provider openai

# Multiple languages
pe run translate.prompt --var target_lang="Japanese" --provider anthropic
pe run translate.prompt --var target_lang="German" --provider openai
```

### Preview Variable Substitution
```bash
# See how variables are replaced
pe cat translate.prompt --var text="Test" --var target_lang="Italian"
```

## Configuration File

The included `config.yaml` shows how to test multiple variable combinations systematically.