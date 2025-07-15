# PE Cat Command Demo

The `pe cat` command provides comprehensive prompt file inspection and variable substitution capabilities.

## Features

- **Multiple Format Support**: Plain text (.prompt), YAML (.yaml), JSON (.json)
- **Variable Substitution**: Interactive or command-line variable assignment
- **Component Inspection**: View system prompts, user prompts, messages, metadata
- **Multiple Output Formats**: Text, YAML, JSON
- **Raw Display**: Show original file without processing

## Example Files

### Simple Prompt with Variables
```prompt
# analyze.prompt
Analyze the following {{topic}} from a {{perspective}} perspective.

Consider these aspects:
- Technical implementation
- Benefits and limitations  
- Real-world applications
- Future implications

Please provide {{detail_level}} analysis.
```

### Structured YAML Prompt
```yaml
# consultant.yaml
system: |
  You are an expert {{domain}} consultant with deep knowledge in {{specialty}}.
  Your task is to provide detailed, actionable advice.
  
  Always structure your response with:
  1. Executive Summary
  2. Detailed Analysis
  3. Recommendations
  4. Next Steps

user: |
  I need help with {{problem_description}}.
  
  Context: {{context}}
  Budget: {{budget}}
  Timeline: {{timeline}}
  
  Please provide your expert analysis.

variables:
  domain: "technology"
  specialty: "cloud architecture"
  problem_description: ""
  context: ""
  budget: "moderate"
  timeline: "3 months"

config:
  temperature: 0.7
  max_tokens: 1500
  provider: "openai"
  model: "gpt-4"

metadata:
  version: "1.0"
  author: "PE Team"
  description: "Structured consulting prompt"
  tags: ["consulting", "analysis"]
```

### Conversation Format
```yaml
# conversation.yaml
messages:
  - role: "system"
    content: "You are a helpful assistant specializing in {{field}}."
  - role: "user"
    content: "What are the best practices for {{task}}?"
  - role: "assistant"
    content: "Here are the key practices for {{task}}:"
  - role: "user"
    content: "Can you explain {{aspect}} in detail?"

variables:
  field: "software development"
  task: "code review"
  aspect: "performance optimization"
```

## Usage Examples

### Basic Display
```bash
# Display prompt with empty variables
pe cat analyze.prompt

# Show available variables
pe cat --variables analyze.prompt

# Display raw file content
pe cat --raw analyze.prompt
```

### Variable Substitution
```bash
# Set variables via command line
pe cat --set topic="Machine Learning" --set perspective="practical" --set detail_level="comprehensive" analyze.prompt

# Interactive variable input
pe cat --interactive analyze.prompt
```

### Component Inspection
```bash
# Show all components
pe cat --components consultant.yaml

# Show only system prompt
pe cat --system consultant.yaml

# Show only metadata
pe cat --metadata consultant.yaml
```

### Output Formats
```bash
# YAML format
pe cat --components --format yaml consultant.yaml

# JSON format
pe cat --components --format json consultant.yaml
```

### Advanced Examples
```bash
# Complex variable substitution
pe cat --set problem_description="migrating to microservices" \
       --set context="legacy monolith application" \
       --set budget="high" \
       --set timeline="6 months" \
       consultant.yaml

# View conversation messages
pe cat --components conversation.yaml

# Set conversation variables
pe cat --set field="data science" \
       --set task="machine learning model validation" \
       --set aspect="cross-validation techniques" \
       conversation.yaml
```

## Output Examples

### Default Output (with variables)
```
Analyze the following Machine Learning from a practical perspective.

Consider these aspects:
- Technical implementation
- Benefits and limitations
- Real-world applications
- Future implications

Please provide comprehensive analysis.
```

### Components View
```
=== System Prompt ===
You are an expert technology consultant with deep knowledge in cloud architecture.
Your task is to provide detailed, actionable advice.

Always structure your response with:
1. Executive Summary
2. Detailed Analysis
3. Recommendations
4. Next Steps

=== User Prompt ===
I need help with migrating to microservices.

Context: legacy monolith application
Budget: high
Timeline: 6 months

Please provide your expert analysis.

=== Variables ===
Variables:
  budget: high
  context: legacy monolith application
  domain: technology
  problem_description: migrating to microservices
  specialty: cloud architecture
  timeline: 6 months

=== Config ===
  max_tokens: 1500
  model: gpt-4
  provider: openai
  temperature: 0.7

=== Metadata ===
  author: PE Team
  description: Structured consulting prompt
  tags: [consulting analysis]
  version: 1.0
```

### Variables View
```
Variables:
  topic: (not set)
  perspective: (not set)
  detail_level: (not set)
```

## Integration with PE Workflow

The `pe cat` command integrates seamlessly with other PE tools:

```bash
# Preview prompt before evaluation
pe cat consultant.yaml

# Use in evaluation workflow
pe cat --set problem_description="API design" consultant.yaml | pe run --provider openai:gpt-4

# Debug prompt templates
pe cat --components --format yaml template.yaml > debug-output.yaml
```

## Best Practices

1. **Use Variables**: Structure prompts with variables for reusability
2. **Component Inspection**: Use `--components` to understand complex prompts
3. **Interactive Mode**: Use `--interactive` for guided variable input
4. **Format Selection**: Choose appropriate output format for your workflow
5. **Testing**: Always preview prompts with `pe cat` before evaluation

## Advanced Features

### Variable Patterns
Variables use `{{variable_name}}` syntax and support:
- Simple substitution: `{{name}}`
- Spaces are handled: `{{ name }}` and `{{name}}` work the same
- Missing variables remain as-is: `{{undefined}}` stays `{{undefined}}`

### File Format Detection
- `.prompt`, `.txt`: Plain text with variable extraction
- `.yaml`, `.yml`: Structured YAML with full component support
- `.json`: JSON format (converted to YAML internally)

### Error Handling
- Missing files: Clear error messages
- Invalid YAML/JSON: Parsing error details
- Invalid variable assignments: Format validation

This command makes prompt development and debugging much more efficient!