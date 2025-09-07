# Creating PE Modules

This example shows how to create a reusable PE module.

## Module Structure

```
translation-prompts/
├── pe.mod              # Module definition
├── pe.sum              # Checksums (auto-generated)
├── prompts/
│   ├── translate.prompt
│   ├── detect-language.prompt
│   └── multi-translate.prompt
├── configs/
│   ├── quality-test.yaml
│   └── benchmark.yaml
└── README.md
```

## Creating a Module

### 1. Initialize Module

```bash
pe mod init github.com/example/translation-prompts
```

This creates `pe.mod`:
```
module github.com/example/translation-prompts

pe 1.0
```

### 2. Add Dependencies

```bash
# Add standard library
pe mod get github.com/tmc/pe-stdlib

# Add another prompt module
pe mod get github.com/example/common-prompts
```

### 3. Create Prompts

`prompts/translate.prompt`:
```
Translate "{{.text}}" from {{.source}} to {{.target}}.

Provide:
- Direct translation
- Pronunciation guide (if applicable)
- Common usage notes

-- defaults --
source=English
target=Spanish
text=Hello
```

### 4. Create Test Config

`configs/quality-test.yaml`:
```yaml
description: "Translation quality tests"

prompts:
  - file://prompts/translate.prompt

providers:
  - name: openai
    config:
      model: gpt-4o-mini

tests:
  - vars:
      text: "Good morning"
      source: "English"
      target: "French"
    assert:
      - type: contains
        value: "Bonjour"
```

### 5. Test Your Module

```bash
# Run tests
pe eval configs/quality-test.yaml

# Benchmark
pe benchmark configs/benchmark.yaml
```

### 6. Document Your Module

Create a comprehensive README with:
- Module purpose
- Available prompts
- Usage examples
- Configuration options
- Test results

### 7. Publish (Optional)

```bash
# Push to GitHub
git init
git add .
git commit -m "Initial translation prompts module"
git remote add origin https://github.com/example/translation-prompts
git push -u origin main

# Tag version
git tag v0.1.0
git push --tags
```

## Using in Other Projects

Others can now use your module:

```bash
pe mod get github.com/example/translation-prompts
```

Then reference in configs:
```yaml
prompts:
  - module://github.com/example/translation-prompts/prompts/translate.prompt
```