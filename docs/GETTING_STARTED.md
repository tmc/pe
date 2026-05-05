# Getting Started with PE

This is the shortest path from installation to a working prompt. For a longer
walkthrough, see [TUTORIAL.md](TUTORIAL.md). For command flags, use
[CLI_REFERENCE.md](CLI_REFERENCE.md) or `pe help [command]`.

## 🚀 Quick Start (5 minutes)

### 1. Installation

```bash
# Install PE
go install github.com/tmc/pe/cmd/pe@latest

# Verify installation
pe --version
```

### 2. Your First Prompt

Create a simple prompt file:

```bash
# Create your first prompt
echo "Summarize this text in one sentence: {{.text}}" > summarize.prompt

# Test it immediately
pe run summarize.prompt --var text="Artificial intelligence is transforming industries worldwide through automation, data analysis, and machine learning capabilities."
```

The exact output depends on the configured provider.

### 3. Set Up Your Provider

PE works with multiple AI providers. Choose one:

#### Option A: OpenAI (Recommended for beginners)
```bash
export OPENAI_API_KEY="your-api-key-here"
pe run summarize.prompt --provider openai:gpt-4 --var text="Your text here"
```

#### Option B: Anthropic Claude
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
pe run summarize.prompt --provider anthropic:claude-3-haiku-20240307 --var text="Your text here"
```

#### Option C: Local Models with Ollama
```bash
# First install and start Ollama
ollama pull llama2
export OLLAMA_HOST="http://localhost:11434"
pe run summarize.prompt --provider ollama:llama2 --var text="Your text here"
```

You have now run a prompt with PE.

## 📖 Essential Concepts

### Prompts as Code
PE treats prompts like code - versioned, tested, and modular:

```bash
# Create a prompt directory structure
mkdir my-prompts && cd my-prompts
pe mod init github.com/myorg/prompts

# This creates a go.mod file for dependency management
```

### Template Variables
Use `{{.variable}}` syntax for dynamic content:

```
You are a {{.role}} expert. {{.task}}

Context: {{.context}}
Requirements:
{{range .requirements}}
- {{.}}
{{end}}

Please provide a detailed response.
```

### Prompt Composition
Build complex prompts from reusable components:

```bash
# Base prompt
echo "You are a helpful assistant." > base.prompt

# Specific task
echo "Analyze the sentiment of: {{.text}}" > sentiment.prompt

# Compose them
pe experimental compose base.prompt sentiment.prompt --style structured --output composed.prompt
```

## 🛠️ Core Workflows

### Workflow 1: Interactive Development

```bash
# Start with a basic prompt
echo "Translate to {{.language}}: {{.text}}" > translate.prompt

# Test it interactively
pe run translate.prompt --var language=Spanish --var text="Hello world"

# Iterate and improve
echo "Translate the following text to {{.language}}, preserving tone and context: {{.text}}" > translate.prompt

# Test again
pe run translate.prompt --var language=Spanish --var text="Hello world"
```

### Workflow 2: Evaluation and Testing

Create an evaluation configuration:

```yaml
# eval-config.yaml
prompts:
  - translate.prompt

tests:
  - vars:
      language: "Spanish" 
      text: "Hello, how are you?"
    assert:
      - type: contains
        value: "Hola"
      - type: not-contains
        value: "Hello"
        
  - vars:
      language: "French"
      text: "Good morning"
    assert:
      - type: contains
        value: "Bonjour"
```

Run evaluation:

```bash
pe eval eval-config.yaml
```

### Workflow 3: Performance Optimization

```bash
# Benchmark different approaches
pe benchmark eval-config.yaml --iterations 10 --concurrency 3

# Optimize with metaprompting
pe experimental optimize --prompt "Translate to Spanish: {{.text}}" --method textgrad --iterations 5

# Test the current config
pe eval eval-config.yaml --output eval-results.json
```

## 🧪 Advanced Features

### 1. Pass@N Evaluation for Code Generation

Perfect for testing code generation reliability:

```yaml
# code-eval.yaml
tests:
  - vars:
      task: "Write a binary search function in Python"
    assert:
      - type: pass-at-n
        config:
          n: 5
          samples: 20
          test_cases:
            - input: "binary_search([1,3,5,7], 5)"
              expected: "2"
            - input: "binary_search([1,3,5,7], 6)" 
              expected: "-1"
        threshold: 0.8  # 80% success rate required
```

```bash
pe eval code-eval.yaml --output code-results.json
```

### 2. Structured Output Validation

Ensure consistent JSON output:

```yaml
# structured-eval.yaml
tests:
  - vars:
      text: "I love this new smartphone!"
    assert:
      - type: structured-output
        config:
          format: json
          schema:
            type: object
            properties:
              sentiment:
                type: string
                enum: ["positive", "negative", "neutral"]
              confidence:
                type: number
                minimum: 0
                maximum: 1
              keywords:
                type: array
                items:
                  type: string
            required: ["sentiment", "confidence"]
```

### 3. Distributed Evaluation

Distributed execution is currently a prototype command group:

```bash
# Inspect distributed prototype interface
pe exp distributed --help

# Run evaluation with core CLI controls
pe eval large-eval.yaml --max-concurrency 20

# Re-check prototype surface
pe exp distributed --help
```

### 4. Local Manifests and Cache

The `pe exp attest` and `pe exp cache` groups are local-only prototypes. They
use SHA-256 manifests and cache keys to detect local file changes; they do not
prove identity, origin, or freshness.

```bash
# Create and verify an unsigned local manifest
pe exp attest manifest prompts/ > manifest.json
pe exp attest verify manifest.json

# Store and verify a local cache entry
pe exp cache put prompt.txt
pe exp cache verify <sha256>
```

## 📊 Working with Metrics

### Built-in Metrics

Advanced evaluation metrics are exposed under the experimental command group:

```bash
# Text generation metrics
pe experimental metrics --type bleu,rouge --generated-file output.txt --reference-file expected.txt

# Custom G-Eval scoring
pe experimental metrics --type g-eval --criteria accuracy,clarity --generated-file responses.txt
```

### Custom Metrics

Define your own evaluation criteria:

```yaml
# custom-metrics.yaml
metrics:
  - name: business_value
    type: llm-judge
    prompt: "Rate the business value of this response from 1-10: {{.response}}"
    
  - name: technical_accuracy
    type: keyword-match
    keywords: ["correct", "accurate", "precise"]
    weight: 0.3
```

## 🔄 Pipeline Workflows

Build complex processing pipelines:

```bash
# Simple pipeline
echo "Raw customer feedback data" | pe ask "Extract sentiment" | pe ask "Categorize by topic" | pe ask "Generate action items"

# Advanced pipeline with filtering
cat reviews.txt | \
  pe ask "Extract product mentions" | \
  pe filter --contains "product" | \
  pe collect --jobs 4
```

## 🎯 Real-World Examples

### Example 1: Customer Support Automation

```bash
# Create support ticket classifier
cat > classify-ticket.prompt << 'EOF'
Classify this support ticket into one of these categories:
- technical: Technical issues, bugs, or feature requests
- billing: Payment, pricing, or subscription questions  
- general: General inquiries or information requests

Ticket: {{.ticket}}

Classification: 
EOF

# Test with real data
pe run classify-ticket.prompt --var ticket="My payment failed and I can't access my account"

# Create evaluation suite
cat > support-eval.yaml << 'EOF'
tests:
  - vars:
      ticket: "App crashes when I try to upload files"
    assert:
      - type: contains
        value: "technical"
  - vars:
      ticket: "How much does the premium plan cost?"
    assert:
      - type: contains
        value: "billing"
EOF

# Validate accuracy
pe eval support-eval.yaml
```

### Example 2: Content Generation Pipeline

```bash
# Create content generation workflow
mkdir content-pipeline && cd content-pipeline

# 1. Topic generation
echo "Generate 5 blog post topics about: {{.subject}}" > generate-topics.prompt

# 2. Outline creation  
echo "Create a detailed outline for this blog post: {{.topic}}" > create-outline.prompt

# 3. Content writing
echo "Write a 500-word blog post based on this outline: {{.outline}}" > write-content.prompt

# Run the full pipeline
pe run generate-topics.prompt --var subject="AI in healthcare" | \
  pe ask "Create an outline" | \
  pe ask "Write content"

# Or use composition for reusable workflows
pe experimental compose generate-topics.prompt create-outline.prompt write-content.prompt \
  --style structured \
  --output content-pipeline.prompt
```

### Example 3: Code Review Assistant

```bash
# Create code review prompt
cat > code-review.prompt << 'EOF'
Review this code for:
1. Security vulnerabilities
2. Performance issues  
3. Code quality and maintainability
4. Best practices adherence

Code:
```{{.language}}
{{.code}}
```

Provide specific, actionable feedback:
EOF

# Test with actual code
pe run code-review.prompt \
  --var language="python" \
  --var code="def login(username, password): return username == 'admin' and password == '12345'"

# Create evaluation for code review quality
cat > code-review-eval.yaml << 'EOF'
tests:
  - vars:
      language: "python"
      code: "eval(user_input)"
    assert:
      - type: contains
        value: "security"
      - type: contains
        value: "eval"
      - type: severity
        level: "high"
EOF
```

## 🏗️ Project Organization

### Recommended Structure

```
my-ai-project/
├── go.mod                    # Module dependencies
├── prompts/
│   ├── base/                 # Reusable base prompts
│   │   ├── system.prompt
│   │   └── assistant.prompt
│   ├── tasks/                # Specific task prompts
│   │   ├── summarize.prompt
│   │   ├── translate.prompt
│   │   └── analyze.prompt
│   └── composed/             # Complex composed prompts
│       └── full-pipeline.prompt
├── evaluations/
│   ├── unit-tests.yaml       # Individual prompt tests
│   ├── integration.yaml     # End-to-end tests
│   └── performance.yaml     # Performance benchmarks
├── configs/
│   ├── providers.yaml       # Provider configurations
│   └── optimization.yaml    # Optimization settings
└── scripts/
    ├── run-tests.sh         # Automated testing
    └── deploy.sh            # Deployment automation
```

### Module Management

```bash
# Initialize a new project
pe mod init github.com/myorg/ai-prompts

# Download dependencies listed in pe.mod
pe mod download

# Keep dependencies clean
pe mod tidy

# Audit dependency changes as JSON
pe mod tidy --json

# Create local copy for offline work
pe mod vendor
```

## 🔧 Configuration and Best Practices

### Provider Configuration

```yaml
# ~/.pe/config.yaml
providers:
  openai:
    api_key_env: OPENAI_API_KEY
    default_model: gpt-4
    timeout: 30s
    
  anthropic:
    api_key_env: ANTHROPIC_API_KEY
    default_model: claude-3-sonnet-20240229
    timeout: 45s
    
  ollama:
    host_env: OLLAMA_HOST
    default_model: llama2
    timeout: 60s

defaults:
  provider: openai
  temperature: 0.1
  max_tokens: 1000
```

### Best Practices

1. **Version Your Prompts**: Use git and semantic versioning
2. **Test Everything**: Write evaluations for all prompts
3. **Use Variables**: Make prompts flexible with template variables
4. **Compose Reusably**: Build complex workflows from simple components
5. **Monitor Performance**: Track latency, cost, and quality metrics
6. **Secure Secrets**: Use environment variables for API keys
7. **Document Thoroughly**: Include examples and expected outputs

### Performance Tips

```bash
# Use parallel evaluation
pe eval large-test-suite.yaml --max-concurrency 10

# Profile performance
pe profile --help
```

## 🆘 Troubleshooting

### Common Issues

**"Provider not found"**
```bash
# Check provider-specific command examples
pe run --help
pe eval --help
```

**"Template variable not found"**
```bash
# Run with an explicit variable
pe run my-prompt.prompt --var known_var=value

# Inspect a prompt file with substitution
pe cat my-prompt.prompt --set known_var=value
```

**"Evaluation failed"**
```bash
# Validate the configuration
pe vet my-eval.yaml

# Run a dry-run evaluation
pe eval my-eval.yaml --dry-run
```

### Getting Help

```bash
# Command-specific help
pe run --help
pe eval --help

# Enable debug logging
PE_DEBUG=1 pe run my-prompt.prompt
```

## 🎓 Next Steps

Now that you've mastered the basics:

1. **Explore Advanced Features**: Try experimental semantic optimization, distributed prototypes, and advanced metrics
2. **Join the Community**: Share prompts and best practices 
3. **Build Real Projects**: Apply PE to your specific use cases
4. **Contribute**: Help improve PE with feedback and contributions

### Learning Resources

- [CLI Reference](CLI_REFERENCE.md) - Complete command documentation
- [Advanced Features](ADVANCED_FEATURES.md) - Deep dive into power features  
- [Examples Library](../examples/) - Real-world prompt examples
- [Architecture Guide](ARCHITECTURE.md) - Understanding PE internals

### Community

- GitHub Issues: Report bugs and request features
- Discussions: Share prompts and get help
- Examples: Contribute your best prompts

Happy prompting! 🚀
