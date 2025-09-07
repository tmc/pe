# PE Template Syntax Guide

PE uses Go's `text/template` package for variable substitution. This provides a powerful, standard templating system.

## Basic Syntax

### Correct: Go Template Format
```
Translate {{.text}} from {{.source}} to {{.target}}
```

### Incorrect: Without Dots
```
Translate {{text}} from {{source}} to {{target}}  # ❌ Won't work
```

## Why Dots?

The dot (`.`) is how Go templates access data. It's not optional - it's the correct syntax.

## Examples

### Simple Variables
```
# prompt.txt
Summarize this text: {{.input}}

# Usage
pe run prompt.txt --var input="Your text here"
```

### Multiple Variables  
```
# translate.prompt
Translate "{{.text}}" from {{.from}} to {{.to}}

# Usage
pe run translate.prompt --var text="Hello" --var from="English" --var to="Spanish"
```

### Complex Templates
```
# analysis.prompt
Analyze the {{.type}} data for {{.company}}:

{{.data}}

Focus on {{.focus_areas}} and provide {{.output_format}}.
```

## Advanced Features

Since PE uses standard Go templates, you get all the power:

### Conditionals
```
{{if .verbose}}
Provide a detailed explanation.
{{else}}
Be concise.
{{end}}
```

### Ranges (Lists)
```
{{range .items}}
- {{.}}
{{end}}
```

### Pipes and Functions
```
{{.message | printf "%.50s"}}  # Truncate to 50 chars
{{.name | upper}}               # Uppercase (if function registered)
```

### With Blocks
```
{{with .user}}
Name: {{.name}}
Email: {{.email}}
{{end}}
```

## Variable Naming

- Use lowercase with underscores: `{{.user_name}}`
- Case-sensitive: `{{.Name}}` ≠ `{{.name}}`
- Must be valid Go identifiers

## Using Variables

### Command Line
```bash
# Single variable
pe run prompt.txt --var text="Hello"

# Multiple variables
pe run translate.prompt \
  --var text="Hello" \
  --var from="English" \
  --var to="Spanish"
```

### In Configuration Files
```yaml
# config.yaml
prompts:
  - "Translate {{.text}} from {{.source}} to {{.target}}"

tests:
  - vars:
      text: "Hello world"
      source: "English"
      target: "French"
    assert:
      - type: contains
        value: "Bonjour"
```

## Common Mistakes

### ❌ Wrong
```
{{variable}}         # Missing dot
${variable}          # Wrong syntax entirely
{{ .variable }}      # Extra spaces (works but not idiomatic)
```

### ✅ Correct
```
{{.variable}}        # Standard format
{{.my_variable}}     # With underscore
```

## Debugging

### Check Variable Substitution
```bash
# Preview how variables are replaced
pe cat prompt.txt --set name="Alice" --set age="30"
```

### List Variables in a Prompt
```bash
# Extract variable names from template
pe cat prompt.txt --variables
```

## Best Practices

1. **Always use dots**: `{{.variable}}` is the only correct syntax
2. **Validate templates**: Use `pe cat` to test before running
3. **Document variables**: Add comments explaining what each variable is for
4. **Use meaningful names**: `{{.target_language}}` not `{{.tl}}`

## Reference

- [Go text/template documentation](https://pkg.go.dev/text/template)
- [PE Command Reference](COMMAND_REFERENCE.md)