# PE Markers Reference

This document describes all markers and directives used in PE prompt files. Markers use the txtar-inspired `-- name --` syntax to define sections, configuration, and structure within prompt files.

## Core Section Markers

### Basic Syntax

Sections are defined using the `-- section-name --` marker format:

```
-- section-name --
Content for this section
```

### Main Prompt Section

The first content before any markers is the main prompt:

```
This is the main prompt that will be executed by default.

-- system-prompt --
You are a helpful assistant.
```

## Special Built-in Markers

### `-- system-prompt --`

Defines a system prompt that will be sent to the LLM as part of the system context.

```
-- system-prompt --
You are an expert programmer with 20 years of experience.
Always provide well-tested code examples.
```

### `-- prompt-summary --`

Provides a brief description of what this prompt does. Used for documentation and discovery.

```
-- prompt-summary --
Analyzes code quality and provides actionable refactoring suggestions.
```

### `-- defaults --`

Specifies default values for prompt variables. Supports both single-line and multi-line formats.

**Single-line format (key=value pairs separated by &):**

```
-- defaults --
Language=Python&Framework=Django
```

**Multi-line format (key: value pairs):**

```
-- defaults --
Language: Python
Framework: Django
Model: gpt-4
Temperature: 0.7
```

### `-- config --`

Configuration directives for prompt execution behavior.

#### Supported Config Directives:

**prefill**
```
-- config --
prefill: "The analysis shows that"
```

Or reference content from another section:
```
-- config --
prefill: content/prefill-template
```

**stop-sequence**
```
-- config --
stop-sequence: "---"
stop-sequence: "END"
```

## Variants System

### `-- variants/variant-name --`

Defines variant modifications that can be applied to the base prompt using `pe run --variant variant-name`.

```
-- variants/formal --
set-system-prompt "You are a formal business consultant"
extend-prompt " Please provide your response in formal business language."

-- variants/casual --
set-system-prompt "You are a friendly coding buddy"
extend-prompt " Keep it casual and use humor where appropriate."
```

#### Variant Commands:

**`extend-system-prompt "text"`**
Appends text to the existing system prompt:
```
extend-system-prompt "Always be concise in your responses."
```

**`prepend-system-prompt "text"`**
Prepends text to the existing system prompt:
```
prepend-system-prompt "You are an expert. "
```

**`set-system-prompt "text"`**
Replaces the entire system prompt:
```
set-system-prompt "You are a creative writer specializing in science fiction."
```

**`extend-prompt "text"`**
Appends text to the main prompt:
```
extend-prompt " Focus on performance implications."
```

**`prepend-prompt "text"`**
Prepends text to the main prompt:
```
prepend-prompt "Before you start, consider: "
```

**`set-flag key value`**
Sets or overrides command-line flags:
```
set-flag model gpt-4-turbo
set-flag temperature 0.5
```

## Examples System

### `-- examples/example-name/VARIABLE --`

Defines example inputs for testing and few-shot learning.

```
-- examples/example-1/input --
What is the capital of France?

-- examples/example-1/ideal-output --
The capital of France is Paris.

-- examples/example-2/input --
What is 2 + 2?

-- examples/example-2/ideal-output --
2 + 2 = 4.

-- examples/example-2/context --
Math questions should be answered with the calculation shown.
```

### Usage Pattern

- Create named example sections (e.g., `example-1`, `example-2`)
- Use meaningful variable names (`input`, `output`, `context`, etc.)
- Special variable: `ideal-output` is used as the expected output for testing
- Used by `pe test` for validation and by `pe run --few-shot` for demonstrations

## Variable Documentation

### `-- variable-description/VARIABLE --`

Documents what a template variable represents:

```
-- variable-description/text --
The input text to be analyzed. Can be multiple sentences or paragraphs.

-- variable-description/language --
Programming language for the code snippet. Examples: Python, Go, JavaScript.

-- variable-description/tone --
The desired tone for the response: formal, casual, technical, or friendly.
```

## Tests System

### `-- tests/test-name --`

Defines test cases for the prompt:

```
-- tests/accuracy-test --
input: "What is 2+2?"
assert-contains: "4"
max-tokens: 50

-- tests/code-quality-test --
input: "Review this function"
assert-not-contains: "syntax error"
min-score: 0.8
```

## Advanced Patterns

### Multiple Variants with Shared Base

```
You are a helpful assistant that analyzes {{type}} content.

-- variants/code-review --
set-system-prompt "You are a senior code reviewer"
extend-prompt "Focus on: performance, security, and maintainability."

-- variants/documentation --
set-system-prompt "You are a technical writer"
extend-prompt "Create clear, well-structured documentation."

-- variants/testing --
set-system-prompt "You are a QA engineer"
extend-prompt "Identify edge cases and test scenarios."
```

### Complete Prompt Structure Example

```
Translate {{text}} to {{target_language}}.

-- system-prompt --
You are a professional translator with expertise in all languages.
Maintain tone and style from the original.

-- prompt-summary --
Translates text between any two languages while preserving tone and style.

-- defaults --
target_language: Spanish

-- config --
stop-sequence: "\n---"

-- examples/example-1/text --
Hello, how are you?

-- examples/example-1/target_language --
Spanish

-- examples/example-1/ideal-output --
Hola, ¿cómo estás?

-- variable-description/text --
The text to be translated. Can be a word, phrase, sentence, or multiple paragraphs.

-- variable-description/target_language --
The target language for translation (e.g., Spanish, French, German, Mandarin).

-- variants/formal --
extend-system-prompt "Use formal language and terminology."

-- variants/casual --
extend-system-prompt "Use casual, colloquial language."

-- tests/basic-translation --
input_text: "Good morning"
assert-contains: "Buenos"
```

## Usage in Commands

### Running with Variants

```bash
pe run prompt.txt --variant formal
pe run prompt.txt --variant casual
```

### Using Defaults

```bash
# Uses defaults from defaults marker
pe run prompt.txt

# Override defaults
pe run prompt.txt --var Language=German
```

### Testing

```bash
# Runs all tests defined in tests markers
pe test prompt.txt

# Validates against examples
pe run prompt.txt --few-shot
```

### Validation

```bash
# Validates prompt structure and all markers
pe validate prompt.txt --strict

# Shows all defined sections, variants, and tests
pe inspect prompt.txt
```

## Best Practices

1. **Always define system-prompt** for consistent behavior across variants
2. **Use descriptive section names** - prefer `examples/sentiment-analysis/input` over `examples/ex1/x`
3. **Document variables** - every template variable should have a `variable-description` marker
4. **Organize by purpose** - group related variants together
5. **Test thoroughly** - define multiple test cases in `tests` markers
6. **Use defaults wisely** - set reasonable defaults for common use cases
7. **Keep markers organized** - order: main → system-prompt → prompt-summary → defaults → config → examples → variants → tests

## Escaping and Special Cases

- Use quotes to preserve whitespace: `extend-prompt "  indented text"`
- Escape quotes in content using `\"`
- Multi-line content is preserved as-is within markers
- Empty lines within a marker are preserved

## Compatibility

- PE markers follow the txtar format conventions used by Go toolchain testing
- Markers are compatible with the DSPy-inspired prompt composition system
- All markers support template variables: `{{variable_name}}`
