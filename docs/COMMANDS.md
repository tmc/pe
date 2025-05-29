# PE Command Reference

Complete reference for all PE commands, organized by category.

## Core Commands

### pe run

Execute a prompt immediately (like `go run`).

```bash
# Run inline prompt
pe run "Explain quantum computing in simple terms"

# Run from file
pe run prompt.txt

# Run from txtar
pe run project.txtar

# Run from gist
pe run gist:username/prompt-id

# With variables
pe run "Translate {{.Text}} to {{.Language}}" \
  --var Text="Hello" \
  --var Language="Spanish"

# With specific provider
pe run prompt.txt --provider gpt-4 --temperature 0.7

# With style guide
pe run prompt.txt --style formal-technical

# Stream output
pe run prompt.txt --stream
```

**Flags:**
- `--provider, -p`: LLM provider (gpt-4, claude-3, etc.)
- `--model, -m`: Specific model version
- `--temperature, -t`: Temperature (0.0-1.0)
- `--max-tokens`: Maximum response tokens
- `--var`: Template variables (repeatable)
- `--style, -s`: Apply style guide
- `--stream`: Stream response
- `--sandbox`: Sandbox mode (strict/relaxed/disabled)

### pe test

Test prompts with assertions (like `go test`).

```bash
# Test all files in directory
pe test prompts/

# Test specific file
pe test config.yaml

# With coverage
pe test prompts/ --coverage

# Verbose output
pe test -v prompts/

# Run specific test
pe test prompts/ --run "TestSentiment"

# With timeout
pe test prompts/ --timeout 30s
```

**Test Configuration Format:**
```yaml
provider: gpt-4
tests:
  - name: "Test positive sentiment"
    prompt: "Analyze sentiment: I love this!"
    assert:
      - type: contains
        value: "positive"
      - type: not_contains
        value: "negative"
```

**Assertion Types:**
- `contains`: Response contains string
- `not_contains`: Response doesn't contain string
- `equals`: Exact match
- `regex`: Regex match
- `length_less_than`: Token count limit
- `json_valid`: Valid JSON response
- `custom`: Custom script validation

### pe build

Build optimized prompts for production (like `go build`).

```bash
# Build single prompt
pe build prompt.txt --output dist/

# Build project
pe build project.txtar --output dist/

# With optimization
pe build prompt.txt --optimize --method pe2

# Minify for production
pe build prompt.txt --minify

# Include metadata
pe build prompt.txt --metadata
```

**Build Output:**
```
dist/
├── prompt.json      # Optimized prompt
├── metadata.json    # Build information
├── tests.yaml       # Validation tests
└── README.md        # Documentation
```

### pe install

Install prompt libraries and dependencies (like `go get`).

```bash
# Install from repository
pe install github.com/tmc/prompts/analyzers

# Install from gist
pe install gist:username/library

# Install specific version
pe install github.com/tmc/prompts/analyzers@v1.2.0

# Install to local project
pe install ./vendor/prompts

# List installed
pe install list

# Update all
pe install update
```

### pe mod

Manage prompt dependencies (like `go mod`).

```bash
# Initialize module
pe mod init myproject

# Add dependency
pe mod get github.com/tmc/prompts/utils

# Update dependencies
pe mod tidy

# Download dependencies
pe mod download

# Vendor dependencies
pe mod vendor

# Show dependency graph
pe mod graph
```

**pe.mod Format:**
```
module myproject

go 1.21

require (
    github.com/tmc/prompts/utils v1.0.0
    gist:user/helpers v0.1.0
)
```

## Version Control Commands

### pe init

Initialize a PE repository.

```bash
# Initialize in current directory
pe init

# Initialize with name
pe init my-assistant

# Initialize from template
pe init --template chatbot
```

### pe branch

Manage branches for prompt development.

```bash
# List branches
pe branch

# Create branch
pe branch create feature/new-style

# Delete branch
pe branch delete old-feature

# Rename branch
pe branch rename old new
```

### pe checkout

Switch between branches or commits.

```bash
# Switch branch
pe checkout main
pe checkout feature/new-prompt

# Create and switch
pe checkout -b feature/experiment

# Checkout specific commit
pe checkout abc1234

# Checkout tag
pe checkout v1.0
```

### pe commit

Save changes to version control.

```bash
# Commit all changes
pe commit -m "Add customer service prompt"

# Commit specific files
pe commit prompt.txt tests.yaml -m "Update prompt and tests"

# Amend last commit
pe commit --amend

# With detailed message
pe commit -m "Title" -m "Detailed description..."
```

### pe merge

Merge branches.

```bash
# Merge branch into current
pe merge feature/new-style

# Merge with strategy
pe merge feature/experiment --strategy ours

# Abort merge
pe merge --abort

# Continue after conflicts
pe merge --continue
```

### pe diff

Show differences between versions.

```bash
# Diff working directory
pe diff

# Diff branches
pe diff main feature/new

# Diff specific files
pe diff main feature/new -- prompt.txt

# Diff commits
pe diff abc1234 def5678
```

## Optimization Commands

### pe optimize

Optimize prompts using various methods.

```bash
# PE2 optimization
pe optimize prompt.txt --method pe2 --iterations 5

# TextGrad optimization
pe optimize prompt.txt --method textgrad --learning-rate 0.1

# APEX for long prompts
pe optimize long-prompt.txt --method apex --beam-width 3

# Evolutionary optimization
pe optimize prompt.txt --method evolve --generations 10 --population 20

# Multi-stage optimization
pe optimize prompt.txt --method multistage --stages "pe2,textgrad,evolve"

# With target metric
pe optimize prompt.txt --target accuracy --threshold 0.9
```

**Optimization Methods:**
- `pe2`: Meta-prompt engineering
- `textgrad`: Semantic gradient descent
- `apex`: Long prompt optimization
- `evolve`: Genetic algorithms
- `fusion`: Multi-model consensus
- `multistage`: Combined methods

### pe evolve

Evolutionary prompt optimization.

```bash
# Basic evolution
pe evolve prompt.txt --generations 20

# Multi-objective
pe evolve prompt.txt --objectives accuracy,brevity,cost

# With constraints
pe evolve prompt.txt --max-tokens 100 --min-accuracy 0.8

# Save population
pe evolve prompt.txt --save-population population/
```

### pe fusion

Multi-model consensus optimization.

```bash
# Consensus across models
pe fusion prompt.txt --models gpt-4,claude-3,gemini

# With weights
pe fusion prompt.txt --models gpt-4:0.5,claude-3:0.3,gemini:0.2

# Voting strategy
pe fusion prompt.txt --strategy majority

# Save all variants
pe fusion prompt.txt --save-variants
```

## Style & Composition Commands

### pe compose

Compose prompts from components.

```bash
# Compose from files
pe compose context.txt instruction.txt examples.txt

# With style
pe compose components/ --style chain-of-thought

# From library
pe compose --library formal --components intro,body,conclusion

# With optimization
pe compose components/ --optimize
```

### pe style

Manage style guides.

```bash
# Create style
pe style create formal --rules "Use formal language"

# Import style
pe style import github.com/org/house-style

# Apply style
pe style apply prompt.txt --style formal

# List styles
pe style list

# Export style
pe style export formal > formal-style.txtar
```

### pe validate

Validate prompts against style guides.

```bash
# Validate single file
pe validate prompt.txt --style formal

# Validate directory
pe validate prompts/ --style company-guide

# With auto-fix
pe validate prompt.txt --style formal --fix

# Explain issues
pe validate prompt.txt --style formal --explain
```

### pe fmt

Format prompt files (like `go fmt`).

```bash
# Format file
pe fmt prompt.txt

# Format directory
pe fmt prompts/

# Format txtar
pe fmt project.txtar

# Check only (no changes)
pe fmt -n prompts/

# With style
pe fmt prompts/ --style formal
```

## Testing & Analysis Commands

### pe benchmark

Run performance benchmarks.

```bash
# Run benchmark file
pe benchmark config.yaml

# Quick benchmark
pe benchmark quick --prompt "Analyze this"

# Compare methods
pe benchmark compare --methods pe2,textgrad,apex

# Cost analysis
pe benchmark cost --scenarios production.yaml

# With report
pe benchmark config.yaml --report html
```

### pe analyze

Analyze prompts and results.

```bash
# Analyze prompt
pe analyze prompt.txt

# Analyze results
pe analyze results.json --metrics all

# Statistical analysis
pe analyze experiment-data/ --statistical

# Token analysis
pe analyze prompt.txt --tokens

# Complexity analysis
pe analyze prompt.txt --complexity
```

### pe compare

Compare prompts or results.

```bash
# Compare prompts
pe compare prompt1.txt prompt2.txt

# Compare results
pe compare results1.json results2.json

# Side-by-side diff
pe compare prompt1.txt prompt2.txt --side-by-side

# With metrics
pe compare results1.json results2.json --metrics accuracy,cost
```

### pe metrics

Calculate evaluation metrics.

```bash
# BLEU score
pe metrics --type bleu --generated output.txt --reference expected.txt

# All metrics
pe metrics --all --generated output.txt --reference expected.txt

# G-Eval
pe metrics --type g-eval --criteria "accuracy,helpfulness"

# Custom metrics
pe metrics --custom metrics.js --data results.json
```

## Cache & Security Commands

### pe cache

Manage response cache.

```bash
# Show status
pe cache status

# Clear cache
pe cache clear

# Export cache
pe cache export --output cache-backup.tar

# Import cache
pe cache import cache-backup.tar

# Verify cache
pe cache verify

# Sync with team
pe cache sync team-cache --verify
```

### pe trust

Manage trusted binaries.

```bash
# Add trusted binary
pe trust add /usr/local/bin/tool

# List trusted
pe trust list

# Remove trust
pe trust remove /usr/local/bin/tool

# Import trust list
pe trust import company-trust.yaml

# Verify binary
pe trust verify /usr/local/bin/tool
```

### pe sandbox

Manage sandbox settings.

```bash
# Show status
pe sandbox status

# Grant permission
pe sandbox grant ~/Documents

# Revoke permission
pe sandbox revoke ~/Documents

# Reset all
pe sandbox reset

# Test sandbox
pe sandbox test
```

## Utility Commands

### pe history

View command history.

```bash
# Show history
pe history

# Search history
pe history search "optimize"

# Replay command
pe history replay 42

# Clear history
pe history clear

# Export history
pe history export > history.txt
```

### pe session

Manage work sessions.

```bash
# New session
pe session new project-x

# List sessions
pe session list

# Resume session
pe session resume project-x

# Delete session
pe session delete old-project

# Export session
pe session export project-x > session.txtar
```

### pe config

Manage configuration.

```bash
# Set value
pe config set default.provider gpt-4

# Get value
pe config get default.provider

# List all
pe config list

# Edit config
pe config edit

# Reset to defaults
pe config reset
```

### pe plugin

Manage plugins.

```bash
# Install plugin
pe plugin install ollama

# List plugins
pe plugin list

# Update plugin
pe plugin update ollama

# Remove plugin
pe plugin remove ollama

# Create plugin
pe plugin create my-provider --template provider
```

### pe doctor

Check system health.

```bash
# Full check
pe doctor

# Quick check
pe doctor --quick

# Fix issues
pe doctor --fix

# Verbose output
pe doctor -v
```

### pe version

Show version information.

```bash
# Basic version
pe version

# Detailed version
pe version --verbose

# Check for updates
pe version --check-update
```

## Pipeline Commands

PE commands can be combined using Unix pipes:

```bash
# Optimization pipeline
pe compose components/ | pe optimize --method pe2 | pe test

# Analysis pipeline
pe run prompt.txt | pe analyze --metrics | pe benchmark

# Multi-stage pipeline
cat prompts.txt | \
  xargs -I {} pe optimize {} --method textgrad | \
  pe test --parallel | \
  pe metrics --aggregate
```

## Environment Variables

Control PE behavior with environment variables:

```bash
# Set provider
PE_PROVIDER=claude-3 pe run "Hello"

# Enable debug
PE_DEBUG=true pe optimize prompt.txt

# Mock mode
PE_MOCK_MODE=true pe test
```

## Exit Codes

PE uses standard exit codes:

- `0`: Success
- `1`: General error
- `2`: Misuse of command
- `3`: Configuration error
- `4`: API error
- `5`: Validation error

## Getting Help

```bash
# General help
pe help

# Command help
pe help run
pe run --help

# List all commands
pe help commands
```