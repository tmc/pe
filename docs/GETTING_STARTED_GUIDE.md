# Getting Started with PE: A Practical Guide

Welcome to PE (Prompt Engineering), the comprehensive toolkit for building, testing, and optimizing AI prompts. This guide will walk you through everything you need to know to get productive quickly.

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

Expected output:
```
AI is revolutionizing global industries via automation, data analysis, and machine learning.
```

### 3. Set Up Your Provider

PE works with multiple AI providers. Choose one:

#### Option A: OpenAI (Recommended for beginners)
```bash
export OPENAI_API_KEY="your-api-key-here"
pe run summarize.prompt --provider openai --model gpt-4 --var text="Your text here"
```

#### Option B: Anthropic Claude
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
pe run summarize.prompt --provider anthropic --model claude-3-sonnet-20240229 --var text="Your text here"
```

#### Option C: Local Models with Ollama
```bash
# First install and start Ollama
ollama pull llama2
export OLLAMA_HOST="http://localhost:11434"
pe run summarize.prompt --provider ollama --model llama2 --var text="Your text here"
```

Congratulations! 🎉 You've just run your first prompt with PE.

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
pe compose --base base.prompt sentiment.prompt --var text="I love this product!"
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
pe benchmark translate.prompt --iterations 10 --providers openai,anthropic,ollama

# Optimize with metaprompting
pe optimize translate.prompt --target "accuracy and conciseness" --iterations 5

# Test optimized version
pe eval eval-config.yaml --prompt optimized-translate.prompt
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
pe eval code-eval.yaml --provider openai --model gpt-4
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

Scale your testing across multiple nodes:

```bash
# Start distributed nodes
pe distributed start --capacity 10 &
pe distributed start --capacity 5 &

# Run evaluation across nodes
pe eval large-eval.yaml --distributed --max-concurrent 20

# Check status
pe distributed status
```

### 4. Cryptographic Attestation

Ensure prompt integrity and auditability:

```bash
# Enable attestation
pe attest run translate.prompt --var language=Spanish --var text="Hello"

# Verify results
pe attest verify --chain latest

# Export audit trail
pe attest export --format csv --output audit.csv
```

## 📊 Working with Metrics

### Built-in Metrics

PE provides comprehensive evaluation metrics:

```bash
# Text generation metrics
pe metrics --prompt summarize.prompt --metric bleu,rouge,bertscore

# Custom G-Eval scoring
pe metrics --prompt creative-writing.prompt --metric g-eval --criteria "creativity,coherence,relevance"

# Performance metrics
pe metrics --prompt fast-qa.prompt --metric latency,tokens-per-second
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
  pe filter --condition "confidence > 0.8" | \
  pe collect --group-by product | \
  pe reduce --operation summarize
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
  pe ask --prompt create-outline.prompt | \
  pe ask --prompt write-content.prompt

# Or use composition for reusable workflows
pe compose generate-topics.prompt create-outline.prompt write-content.prompt \
  --var subject="sustainable technology" \
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

# Add dependencies
pe mod download github.com/pe-community/base-prompts@v1.2.0

# Keep dependencies clean
pe mod tidy

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
# Cache results for repeated testing
pe run --cache summarize.prompt --var text="Same text"

# Use parallel evaluation
pe eval large-test-suite.yaml --parallel 10

# Optimize for speed vs quality
pe run --temperature 0 --max-tokens 100 quick-response.prompt

# Profile performance
pe profile --enable cpu,memory my-complex-prompt.prompt
```

## 🆘 Troubleshooting

### Common Issues

**"Provider not found"**
```bash
# Check available providers
pe providers list

# Register providers manually
pe providers register anthropic openai ollama
```

**"Template variable not found"**
```bash
# Debug template variables
pe run --debug my-prompt.prompt --var known_var="value"

# List required variables
pe analyze my-prompt.prompt --show-variables
```

**"Evaluation failed"**
```bash
# Run with verbose output
pe eval --verbose my-eval.yaml

# Test individual assertions
pe eval --test-only "test_name" my-eval.yaml
```

### Getting Help

```bash
# Command-specific help
pe run --help
pe eval --help

# Show examples
pe examples

# Check system status
pe status

# Enable debug logging
PE_DEBUG=1 pe run my-prompt.prompt
```

## 🎓 Next Steps

Now that you've mastered the basics:

1. **Explore Advanced Features**: Try semantic optimization, distributed evaluation, and custom metrics
2. **Join the Community**: Share prompts and best practices 
3. **Build Real Projects**: Apply PE to your specific use cases
4. **Contribute**: Help improve PE with feedback and contributions

### Learning Resources

- [API Reference](API_REFERENCE.md) - Complete command documentation
- [Advanced Features](ADVANCED_FEATURES.md) - Deep dive into power features  
- [Examples Library](../example/) - Real-world prompt examples
- [Architecture Guide](ARCHITECTURE.md) - Understanding PE internals

### Community

- GitHub Issues: Report bugs and request features
- Discussions: Share prompts and get help
- Examples: Contribute your best prompts

Happy prompting! 🚀