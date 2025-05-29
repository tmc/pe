# PE Tutorial: Getting Started with Go for Prompts

This tutorial will walk you through PE from basics to advanced features. By the end, you'll be comfortable using PE as your primary prompt engineering toolkit.

## Prerequisites

- PE installed (see [Installation Guide](INSTALLATION.md))
- An API key for at least one LLM provider
- Basic command line familiarity

## Part 1: Your First PE Commands

### Step 1: Initialize a Project

Let's start by creating a new PE project:

```bash
# Create a project directory
mkdir my-assistant
cd my-assistant

# Initialize PE
pe init
```

This creates a `.pe/` directory for version control and configuration.

### Step 2: Run Your First Prompt

```bash
# Run an inline prompt
pe run "What is the capital of France?"
```

You should see:
```
Paris
```

Now let's create a prompt file:

```bash
# Create a prompt file
echo "You are a helpful coding assistant. Be concise and accurate." > assistant.txt

# Run from file
pe run assistant.txt
```

### Step 3: Using Variables

PE supports template variables for dynamic prompts:

```bash
# Create a template
cat > translator.txt << 'EOF'
Translate the following {{.Language}} phrase to English:
"{{.Text}}"

Provide only the translation, no explanations.
EOF

# Run with variables
pe run translator.txt --var Language="Spanish" --var Text="Hola, ¿cómo estás?"
```

Output:
```
Hello, how are you?
```

## Part 2: Testing and Validation

### Step 4: Create Tests

Let's add tests to ensure our translator works correctly:

```bash
# Create test configuration
cat > translator-tests.yaml << 'EOF'
provider: gpt-4
tests:
  - name: "Spanish greeting"
    prompt_file: translator.txt
    variables:
      Language: Spanish
      Text: "Hola"
    assert:
      - type: contains
        value: "Hello"
      
  - name: "French greeting"
    prompt_file: translator.txt
    variables:
      Language: French
      Text: "Bonjour"
    assert:
      - type: equals
        value: "Hello"
        
  - name: "German thanks"
    prompt_file: translator.txt
    variables:
      Language: German
      Text: "Danke"
    assert:
      - type: contains
        value: "Thank"
EOF

# Run tests
pe test translator-tests.yaml
```

### Step 5: Style Guides

Create a style guide for consistent outputs:

```bash
# Create style guide
cat > concise-style.txtar << 'EOF'
-- rules.yaml --
name: concise
rules:
  - id: brevity
    description: "Keep responses under 50 words"
    weight: 1.0
  - id: no-fluff
    description: "No unnecessary pleasantries"
    weight: 0.8
    
-- validator.sh --
#!/bin/bash
# Check word count
[ $(wc -w < "$1") -le 50 ]
EOF

# Apply style to prompt
pe run assistant.txt --style concise-style.txtar
```

## Part 3: Version Control

### Step 6: Branching and Committing

PE provides Git-like version control:

```bash
# Check status
pe status

# Commit initial version
pe commit -m "Initial assistant and translator prompts"

# Create a feature branch
pe branch create feature/friendly-tone

# Switch to branch
pe checkout feature/friendly-tone

# Modify the assistant
echo "You are a friendly and helpful coding assistant. Use a warm tone." > assistant.txt

# Commit changes
pe commit -m "Add friendly tone to assistant"

# Compare branches
pe diff main feature/friendly-tone
```

### Step 7: Merging and Tagging

```bash
# Switch back to main
pe checkout main

# Merge the feature
pe merge feature/friendly-tone

# Tag a release
pe tag v1.0 -m "First stable version"
```

## Part 4: Optimization

### Step 8: Basic Optimization

Let's optimize our assistant prompt:

```bash
# Optimize using PE2 method
pe optimize assistant.txt --method pe2 --iterations 5

# View the optimized version
cat assistant_optimized.txt
```

You'll see PE2 has enhanced your prompt with:
- Expert persona
- Clear instructions
- Output formatting
- Edge case handling

### Step 9: Comparing Optimization Methods

```bash
# Try different methods
pe optimize assistant.txt --method textgrad --output assistant_textgrad.txt
pe optimize assistant.txt --method apex --output assistant_apex.txt

# Compare results
pe compare assistant.txt assistant_optimized.txt assistant_textgrad.txt
```

## Part 5: Advanced Features

### Step 10: Using txtar Format

Create a complete project in a single txtar file:

```bash
cat > chatbot.txtar << 'EOF'
-- prompt.txt --
You are a knowledgeable chatbot specializing in {{.Topic}}.
Provide accurate, helpful responses.

-- config.yaml --
provider: gpt-4
temperature: 0.7
max_tokens: 500

-- tests.yaml --
tests:
  - name: "Basic question"
    variables:
      Topic: "astronomy"
    input: "What is a black hole?"
    assert:
      - type: contains
        value: "gravity"
      
-- style.yaml --
rules:
  - id: educational
    description: "Use educational tone"
EOF

# Run the txtar directly
pe run chatbot.txtar --var Topic="astronomy"
```

### Step 11: Working with Gists

Share your prompt via GitHub gist:

```bash
# Export to gist format
pe export chatbot.txtar --format gist > chatbot-gist.txt

# If you have a gist URL, you can run directly:
pe run gist:username/chatbot-id --var Topic="physics"

# Install from gist
pe install gist:username/prompt-library
```

### Step 12: Pipeline Processing

Combine PE commands for powerful workflows:

```bash
# Optimization pipeline
pe compose context.txt instructions.txt examples.txt | \
  pe optimize --method pe2 --iterations 3 | \
  pe test --inline --assert "quality>0.8" | \
  pe build --output production/

# Benchmark pipeline
echo "assistant.txt translator.txt" | \
  xargs -n1 pe benchmark quick | \
  pe analyze --metrics cost,quality | \
  pe report --format markdown > benchmark.md
```

## Part 6: Security and Plugins

### Step 13: Sandbox and Trust

```bash
# Check sandbox status
pe sandbox status

# Run with strict sandboxing
pe run assistant.txt --sandbox=strict

# Trust a binary (if needed for your workflow)
pe trust add /usr/local/bin/jq
pe trust list
```

### Step 14: Installing Plugins

```bash
# Install local model support
pe plugin install ollama

# Configure plugin
pe config set plugins.ollama.url "http://localhost:11434"

# Use the plugin
pe run assistant.txt --provider ollama --model llama2
```

## Part 7: Team Collaboration

### Step 15: Shared Caching

```bash
# Enable caching
pe config set cache.enabled true

# Run commands (responses are cached)
pe run "Explain Python decorators"
pe run "Explain Python decorators"  # This uses cache

# Export cache for team
pe cache export --sign > team-cache.tar

# Team member imports cache
pe cache import team-cache.tar --verify
```

### Step 16: Collaborative Development

```bash
# Create a fork for experimentation
pe fork assistant.txt --name assistant-experimental

# Work on the fork
pe checkout -b experiment
echo "You are an experimental AI assistant with creative responses." > assistant-experimental.txt
pe commit -m "Add creative mode"

# Share via gist
pe push gist:myteam/assistant-experiment
```

## Part 8: Production Workflow

### Step 17: Complete Development Cycle

Here's a real-world workflow for developing a production prompt:

```bash
# 1. Start with requirements
cat > requirements.txt << 'EOF'
- Customer service chatbot
- Professional but friendly tone  
- Handle complaints gracefully
- Multilingual support
EOF

# 2. Create initial prompt
cat > customer-service.txt << 'EOF'
You are a customer service representative.
Help customers with their inquiries.
EOF

# 3. Create test cases
cat > cs-tests.yaml << 'EOF'
provider: gpt-4
tests:
  - name: "Complaint handling"
    input: "Your product is terrible!"
    assert:
      - type: contains
        value: "sorry"
      - type: contains  
        value: "help"
      - type: not_contains
        value: "terrible"
        
  - name: "Information request"
    input: "What are your business hours?"
    assert:
      - type: contains
        value: "hours"
EOF

# 4. Iterative development
pe test cs-tests.yaml  # See what fails

# 5. Optimize the prompt
pe optimize customer-service.txt --method pe2 --target "cs-tests.yaml"

# 6. Benchmark performance
cat > cs-benchmark.yaml << 'EOF'
scenarios:
  - name: "Complaint Resolution"
    prompt_file: customer-service_optimized.txt
    test_cases: 
      - "I want a refund"
      - "This is unacceptable"
      - "I'm very disappointed"
metrics:
  - satisfaction_score
  - response_time
  - resolution_rate
EOF

pe benchmark cs-benchmark.yaml

# 7. A/B test variants
pe fork customer-service_optimized.txt --name cs-formal
pe fork customer-service_optimized.txt --name cs-casual

pe ab-test cs-formal cs-casual --duration 1h --metric satisfaction

# 8. Build for production
pe build customer-service_optimized.txt \
  --tests cs-tests.yaml \
  --style professional \
  --output dist/

# 9. Tag release
pe tag v1.0-prod -m "Production-ready customer service bot"
```

## Part 9: Advanced Optimization

### Step 18: Multi-Stage Optimization

```bash
# Create a complex prompt that needs sophisticated optimization
cat > analyzer.txt << 'EOF'
Analyze the provided text.
EOF

# Multi-stage optimization pipeline
pe optimize analyzer.txt \
  --method multistage \
  --stages "pe2:5,textgrad:8,apex:3" \
  --target "accuracy>0.9,tokens<200" \
  --output analyzer_final.txt

# View optimization report
pe analyze analyzer_final.txt --compare analyzer.txt
```

### Step 19: Evolutionary Optimization

```bash
# Use genetic algorithms for optimization
pe evolve analyzer.txt \
  --generations 20 \
  --population 50 \
  --objectives accuracy,brevity,clarity \
  --constraints "tokens<150" \
  --output evolved/

# Examine the Pareto frontier
pe analyze evolved/ --pareto --visualize
```

## Part 10: Best Practices

### 1. Project Structure

```
my-project/
├── .pe/                    # PE version control
├── prompts/
│   ├── main.txt           # Main prompt
│   ├── components/        # Reusable components
│   └── variants/          # A/B test variants
├── tests/
│   ├── unit.yaml          # Unit tests
│   └── integration.yaml   # Integration tests
├── styles/
│   └── house-style.yaml   # Organization style guide
├── benchmarks/
│   └── performance.yaml   # Performance benchmarks
└── pe.mod                 # Dependencies
```

### 2. Development Workflow

```bash
# Morning routine
pe pull origin main
pe checkout -b feature/todays-work
pe session new todays-work

# Development cycle
pe run prompt.txt --watch  # Auto-reload on changes
pe test --watch           # Continuous testing

# Before committing
pe fmt prompts/
pe test
pe validate prompts/ --style house-style

# End of day
pe commit -m "Today's improvements"
pe push origin feature/todays-work
```

### 3. Testing Strategy

- **Unit tests**: Test individual prompts
- **Integration tests**: Test prompt combinations
- **Regression tests**: Ensure no quality degradation
- **A/B tests**: Compare variants in production

### 4. Optimization Strategy

1. Start with PE2 for general improvement
2. Use TextGrad for fine-tuning
3. Apply APEX for long prompts
4. Use evolutionary methods for multi-objective optimization

## Conclusion

You've learned how to:
- ✓ Run and test prompts
- ✓ Use version control for prompt development
- ✓ Apply optimization methods
- ✓ Work with teams using shared caches
- ✓ Build production-ready prompts
- ✓ Create sophisticated optimization pipelines

PE brings the power and simplicity of Go's toolchain to prompt engineering. Continue exploring with:

- [Command Reference](COMMANDS.md) for detailed command options
- [Architecture Guide](ARCHITECTURE.md) for technical details
- [Plugin Development](PLUGINS.md) to extend PE

Happy prompting!