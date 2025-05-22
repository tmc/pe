# API Reference

Complete reference for PE's command-line interface, configuration options, and programmatic APIs.

## 📋 Table of Contents

- [Command Line Interface](#command-line-interface)
- [Configuration Reference](#configuration-reference)
- [Provider API](#provider-api)
- [Assertion Types](#assertion-types)
- [Optimization Methods](#optimization-methods)
- [Go API](#go-api)
- [REST API](#rest-api)
- [Environment Variables](#environment-variables)

---

## 🖥️ Command Line Interface

### Global Options

All commands support these global flags:

```bash
pe [global-options] <command> [command-options] [arguments]

Global Options:
  --config string      Configuration file path
  --log-level string   Log level: debug, info, warn, error (default "info")
  --log-file string    Log output file (default: stderr)
  --verbose, -v        Enable verbose output
  --quiet, -q          Suppress non-essential output
  --json               Output results in JSON format
  --no-color           Disable colored output
  --help, -h           Show help
  --version            Show version information
```

### Core Commands

#### `pe eval`
Evaluate prompts against test cases.

```bash
pe eval [options] <config-file>

Options:
  -o, --output string          Output file path
  --save-db                   Save results to database
  --dry-run                   Show what would be executed without running
  --max-concurrency int       Maximum concurrent evaluations (default 4)
  --timeout duration          Request timeout (default 30s)
  --retry-attempts int        Number of retry attempts (default 3)
  --retry-delay duration      Delay between retries (default 1s)
  --stream                    Stream results as they complete
  --cache                     Enable response caching
  --cache-ttl duration        Cache time-to-live (default 1h)
  --provider strings          Override providers for this run
  --include string            Include test pattern (glob)
  --exclude string            Exclude test pattern (glob)
  --fail-fast                 Stop on first failure
  --progress                  Show progress bar
  --metrics                   Collect detailed metrics

Examples:
  pe eval config.yaml
  pe eval config.yaml --output results.json
  pe eval config.yaml --max-concurrency 8 --timeout 60s
  pe eval config.yaml --provider "openai:gpt-4" --dry-run
```

#### `pe optimize`
Optimize prompts using advanced methods.

```bash
pe optimize [options] --prompt <prompt> | --config <config-file>

Options:
  --prompt string             Base prompt to optimize
  --config string             Configuration file with optimization settings
  --method string             Optimization method: textgrad, standard, hybrid (default "textgrad")
  --iterations int            Number of optimization iterations (default 5)
  --provider string           Provider for optimization (default "openai:gpt-4")
  --objective string          Optimization objective
  --output string             Output file for optimized prompt
  --save-trajectory           Save optimization trajectory
  --convergence-threshold float  Stop when improvement < threshold (default 0.02)
  --temperature float         Optimization temperature (default 0.1)
  --exploration-factor float  Exploration vs exploitation (default 0.2)
  --gradient-accumulation     Enable gradient accumulation
  --reflection-depth int      Depth of self-reflection (default 2)

Examples:
  pe optimize --prompt "Summarize: {{text}}" --method textgrad --iterations 5
  pe optimize --config optimization.yaml --method hybrid
  pe optimize --prompt "Generate code" --objective "accuracy and readability"
```

#### `pe view`
Interactive result viewer.

```bash
pe view [options] [result-id]

Options:
  -f, --file string           Load results from file
  --port int                  Web server port (default 8080)
  --host string               Web server host (default "localhost")
  --export-report string      Export HTML report
  --export-csv string         Export CSV data
  --filter string             Filter results
  --group-by string           Group results by field

Examples:
  pe view                     # View latest results
  pe view eval-123           # View specific evaluation
  pe view --file results.json # View from file
  pe view --export-report report.html
```

#### `pe interactive`
Interactive REPL mode.

```bash
pe interactive [options]

Options:
  --provider string           Default provider
  --config string             Load configuration
  --temperature float         Default temperature (default 0.7)
  --max-tokens int           Default max tokens (default 1000)
  --save-session string       Save session file

REPL Commands:
  /help                       Show available commands
  /providers                  List available providers
  /use <provider>            Switch provider
  /config <key> <value>      Set configuration
  /optimize                   Optimize current prompt
  /save <file>               Save current session
  /load <file>               Load session
  /history                    Show command history
  /clear                      Clear screen
  /exit                       Exit REPL

Examples:
  pe interactive --provider anthropic:claude-3-sonnet
  pe interactive --config my-config.yaml
```

#### `pe benchmark`
Performance benchmarking.

```bash
pe benchmark [options] <config-file>

Options:
  --iterations int            Number of benchmark iterations (default 10)
  --concurrency int           Concurrent requests (default 4)
  --warmup int                Warmup iterations (default 2)
  --format string             Output format: table, json, csv (default "table")
  --output string             Output file
  --metric string             Primary metric to optimize
  --percentiles ints          Percentiles to calculate (default [50,90,95,99])
  --compare-providers         Compare all providers
  --statistical-analysis      Include statistical analysis
  --confidence float          Confidence level (default 0.95)

Examples:
  pe benchmark config.yaml --iterations 100
  pe benchmark config.yaml --format json --output benchmark.json
  pe benchmark config.yaml --compare-providers --statistical-analysis
```

### Pipeline Commands

#### `pe ask`
Single-shot question answering (pipeline-friendly).

```bash
pe ask [options] [prompt]

Options:
  --provider string           LLM provider (default "openai:gpt-3.5-turbo")
  --temperature float         Temperature (default 0.7)
  --max-tokens int           Maximum tokens (default 1000)
  --system string             System message
  --format string             Output format: text, json (default "text")

Examples:
  echo "What is AI?" | pe ask
  pe ask --provider anthropic:claude-3-haiku "Explain quantum computing"
  echo "Code review this function" | pe ask --system "You are a senior developer"
```

#### `pe stream`
Process evaluation results as streams.

```bash
pe stream [options]

Options:
  --select strings            Select fields to output
  --format string             Output format: json, csv, tsv (default "json")
  --buffer-size int           Buffer size for streaming (default 1000)

Examples:
  pe eval config.yaml | pe stream --select response,score,latency
  pe eval config.yaml | pe stream --format csv > results.csv
```

#### `pe filter`
Filter evaluation results.

```bash
pe filter [options]

Options:
  --success                   Include only successful results
  --failure                   Include only failed results
  --min-score float           Minimum score threshold
  --max-score float           Maximum score threshold
  --min-latency duration      Minimum latency
  --max-latency duration      Maximum latency
  --min-cost float            Minimum cost
  --max-cost float            Maximum cost
  --provider string           Filter by provider
  --prompt string             Filter by prompt
  --contains string           Filter responses containing text
  --regex string              Filter responses matching regex

Examples:
  pe eval config.yaml | pe filter --success --min-score 0.8
  pe eval config.yaml | pe filter --max-latency 5s --max-cost 0.02
  pe eval config.yaml | pe filter --contains "Paris" --provider "gpt-4"
```

#### `pe analyze`
Statistical analysis of results.

```bash
pe analyze [options]

Options:
  --metric string             Metric to analyze: score, latency, cost, tokens
  --group-by string           Group results by field
  --percentiles ints          Percentiles to calculate
  --confidence float          Confidence level (default 0.95)
  --trend-analysis            Perform trend analysis
  --correlation               Calculate correlation matrix
  --outlier-detection         Detect and flag outliers
  --export string             Export analysis results

Examples:
  pe eval config.yaml | pe analyze --metric latency --percentiles 50,90,95,99
  pe eval config.yaml | pe analyze --group-by provider --correlation
  pe eval config.yaml | pe analyze --trend-analysis --export analysis.json
```

#### `pe stats`
Quick statistics summary.

```bash
pe stats [options]

Options:
  --detailed                  Show detailed statistics
  --group-by string           Group statistics by field
  --format string             Output format: table, json, yaml
  --export string             Export to file

Examples:
  pe eval config.yaml | pe stats
  pe eval config.yaml | pe stats --detailed --group-by provider
  pe eval config.yaml | pe stats --format json --export stats.json
```

#### `pe diff`
Compare evaluation results.

```bash
pe diff [options] <baseline> <current>

Options:
  --threshold float           Significance threshold (default 0.05)
  --metric string             Primary comparison metric
  --confidence float          Confidence level (default 0.95)
  --statistical-test string   Statistical test: ttest, mannwhitney, ks
  --effect-size               Calculate effect sizes
  --visualization             Generate comparison charts
  --report string             Generate comparison report

Examples:
  pe diff baseline.json current.json
  pe diff eval-123 eval-456 --metric score --threshold 0.02
  pe diff results1.json results2.json --statistical-test ttest --report diff.html
```

### Utility Commands

#### `pe fmt`
Format configuration files.

```bash
pe fmt [options] <files...>

Options:
  --write                     Write changes to files
  --check                     Check if files are formatted
  --output string             Output format: yaml, json
  --indent int                Indentation spaces (default 2)
  --sort-keys                 Sort object keys

Examples:
  pe fmt config.yaml
  pe fmt config.yaml --write
  pe fmt *.yaml --check
```

#### `pe vet`
Validate configuration files.

```bash
pe vet [options] <files...>

Options:
  --strict                    Strict validation mode
  --schema string             Custom schema file
  --recursive                 Validate recursively
  --fix                       Auto-fix common issues

Examples:
  pe vet config.yaml
  pe vet *.yaml --strict
  pe vet configs/ --recursive
```

#### `pe convert`
Convert between configuration formats.

```bash
pe convert [options] <input> <output>

Options:
  --input-format string       Input format: yaml, json
  --output-format string      Output format: yaml, json
  --pretty                    Pretty-print output

Examples:
  pe convert config.yaml config.json
  pe convert config.json config.yaml --pretty
```

---

## ⚙️ Configuration Reference

### Top-Level Configuration

```yaml
# config.yaml
description: "Optional description of this configuration"
version: "1.0"  # Configuration schema version

# Core configuration sections
prompts: []      # Prompt definitions
providers: []    # Provider configurations  
tests: []        # Test cases
metrics: []      # Custom metrics (optional)
optimization: {} # Optimization settings (optional)
output: {}       # Output configuration (optional)
```

### Prompts Configuration

```yaml
prompts:
  # Simple string prompt
  - "What is the capital of {{country}}?"
  
  # Detailed prompt object
  - id: "detailed-prompt"
    content: |
      You are an expert in {{domain}}.
      Please answer this question: {{question}}
      
      Provide a comprehensive response including:
      1. Direct answer
      2. Supporting evidence
      3. Relevant context
    
    # Optional metadata
    description: "Detailed expert response format"
    tags: ["expert", "comprehensive"]
    version: "1.2"
    
  # Template-based prompt
  - id: "template-prompt"
    template: "expert-response"  # References templates.expert-response
    variables:
      domain: "{{subject_area}}"
      context: "academic setting"
    
  # Multi-modal prompt
  - id: "vision-prompt"
    type: "multimodal"
    content: "Analyze this image: {{image_url}}"
    modalities:
      - type: "image"
        source: "{{image_url}}"
      - type: "text"
        source: "{{question}}"

# Template library (optional)
templates:
  expert-response: |
    You are an expert in {{domain}}.
    Context: {{context}}
    
    Question: {{question}}
    
    Please provide an expert-level response.
```

### Providers Configuration

```yaml
providers:
  # Simple provider reference
  - "openai:gpt-4"
  - "anthropic:claude-3-sonnet"
  
  # Detailed provider configuration
  - id: "custom-gpt4"
    type: "openai"
    model: "gpt-4"
    config:
      temperature: 0.1
      max_tokens: 1000
      top_p: 0.9
      frequency_penalty: 0.0
      presence_penalty: 0.0
      stop: ["###", "END"]
    
    # Optional provider metadata
    description: "Conservative GPT-4 for factual responses"
    cost_per_token: 0.00003  # For cost tracking
    rate_limit: 100          # Requests per minute
    
  # Custom/local provider
  - id: "local-llm"
    type: "custom"
    endpoint: "http://localhost:8000/v1/chat/completions"
    config:
      api_key: "${LOCAL_API_KEY}"
      model: "llama-2-7b"
      temperature: 0.7
    
    # Headers for custom providers
    headers:
      "Custom-Header": "value"
      "Authorization": "Bearer ${TOKEN}"

# Provider defaults (optional)
provider_defaults:
  temperature: 0.7
  max_tokens: 500
  timeout: "30s"
  retry_attempts: 3
```

### Tests Configuration

```yaml
tests:
  # Basic test case
  - description: "Geography question"
    vars:
      country: "France"
    assert:
      - type: "contains"
        value: "Paris"
  
  # Complex test case
  - description: "Technical explanation test"
    vars:
      topic: "machine learning"
      audience: "beginners"
      length: 200
    
    # Multiple assertions
    assert:
      - type: "contains"
        value: ["machine learning", "algorithm"]
        description: "Must mention key terms"
      
      - type: "length"
        min: 150
        max: 250
        description: "Appropriate length"
      
      - type: "readability"
        min_grade_level: 8
        max_grade_level: 12
        description: "Accessible to target audience"
      
      - type: "llm-judge"
        value: |
          Rate this explanation for beginners (1-10):
          - Clarity (30%)
          - Accuracy (40%) 
          - Engagement (30%)
        threshold: 0.7
        judge: "gpt-4"
      
      - type: "latency"
        max: "5s"
        description: "Response time requirement"
      
      - type: "cost"
        max: 0.02
        description: "Cost constraint"
    
    # Test-specific configuration
    options:
      retries: 2
      timeout: "10s"
      
  # Conversation test
  - description: "Multi-turn conversation"
    conversation:
      - user: "Hello, I'm learning about AI"
        assistant_assert:
          - type: "contains"
            value: ["AI", "artificial intelligence"]
      - user: "Can you explain neural networks?"
        assistant_assert:
          - type: "contains"
            value: ["neural", "network", "nodes"]
          - type: "educational_quality"
            threshold: 0.8

# Test defaults (optional)
test_defaults:
  timeout: "30s"
  retries: 1
  fail_fast: false
```

### Assertions Reference

#### Text Content Assertions

```yaml
# Contains text
- type: "contains"
  value: "expected text"
  # or array of strings (any match passes)
  value: ["option1", "option2"]
  case_sensitive: false  # default: false
  
# Does not contain text
- type: "not-contains"
  value: "unwanted text"
  
# Regex match
- type: "regex"
  pattern: "\\d{4}-\\d{2}-\\d{2}"  # Date pattern
  flags: "i"  # Case insensitive
  
# Exact match
- type: "equals"
  value: "exact text"
  case_sensitive: true
  
# Starts/ends with
- type: "starts-with"
  value: "Hello"
- type: "ends-with"
  value: "goodbye"
```

#### Length and Structure Assertions

```yaml
# Text length
- type: "length"
  min: 100
  max: 500
  
# Word count
- type: "word-count"
  min: 50
  max: 100
  
# Sentence count
- type: "sentence-count"
  value: 3  # exact count
  # or range
  min: 2
  max: 5
  
# Line count
- type: "line-count"
  min: 5
  max: 20
```

#### Quality Assertions

```yaml
# Readability
- type: "readability"
  metric: "flesch_reading_ease"  # default
  min: 0.6  # 0-1 scale
  # Alternative metrics: flesch_kincaid, gunning_fog, coleman_liau
  
# Sentiment analysis
- type: "sentiment"
  value: "positive"  # positive, negative, neutral
  confidence: 0.7    # minimum confidence
  
# Language detection
- type: "language"
  value: "en"        # ISO 639-1 code
  confidence: 0.9
  
# Toxicity detection
- type: "toxicity"
  max: 0.1           # 0-1 scale, lower is better
  
# Coherence analysis
- type: "coherence"
  threshold: 0.8     # 0-1 scale
  
# Factuality check
- type: "factuality"
  threshold: 0.9
  knowledge_cutoff: "2024-01-01"
```

#### Performance Assertions

```yaml
# Response latency
- type: "latency"
  max: "5s"          # duration string
  
# Cost constraints
- type: "cost"
  max: 0.05          # dollars
  
# Token usage
- type: "tokens"
  max: 1000
  # or separate input/output limits
  max_input: 500
  max_output: 500
```

#### Structured Data Assertions

```yaml
# Valid JSON
- type: "json"
  # Optional schema validation
  schema: |
    {
      "type": "object",
      "required": ["name", "age"],
      "properties": {
        "name": {"type": "string"},
        "age": {"type": "integer", "minimum": 0}
      }
    }
  
# Valid XML
- type: "xml"
  schema: "schema.xsd"  # Optional XSD schema
  
# Valid YAML
- type: "yaml"
  
# CSV format
- type: "csv"
  delimiter: ","
  headers: ["name", "age", "city"]
```

#### Code Assertions

```yaml
# Code syntax validation
- type: "code-syntax"
  language: "python"
  
# Code security scan
- type: "code-security"
  language: "python"
  rules: ["no-eval", "no-exec", "no-dangerous-imports"]
  
# Code quality metrics
- type: "code-quality"
  language: "python"
  metrics: ["complexity", "maintainability", "readability"]
  threshold: 0.8
  
# Code execution test
- type: "code-execution"
  language: "python"
  test_cases:
    - input: "factorial(5)"
      expected: 120
    - input: "factorial(0)"
      expected: 1
```

#### LLM-Based Assertions

```yaml
# LLM as judge
- type: "llm-judge"
  value: "Rate the helpfulness of this response (1-10)"
  threshold: 0.7
  judge: "gpt-4"           # Provider for judging
  temperature: 0.1         # Judge temperature
  
# Custom evaluation criteria
- type: "llm-judge"
  value: |
    Evaluate this response on multiple criteria:
    1. Accuracy (40% weight)
    2. Clarity (30% weight)  
    3. Completeness (30% weight)
    
    Provide a score from 1-10 and brief explanation.
  threshold: 0.8
  judge: "anthropic:claude-3-opus"
  
# Factual accuracy check
- type: "factuality"
  threshold: 0.9
  judge: "gpt-4"
  knowledge_cutoff: "2024-01-01"
  
# Translation quality
- type: "translation-quality"
  source_language: "en"
  target_language: "es"
  reference: "reference translation"  # optional
  metrics: ["accuracy", "fluency", "adequacy"]
  threshold: 0.8
```

#### Custom Assertions

```yaml
# Python script assertion
- type: "python"
  script: |
    def evaluate(output, vars):
        # Custom evaluation logic
        score = len(output) / 100  # Example scoring
        passed = score > 0.5
        return {"score": score, "passed": passed, "message": f"Length score: {score}"}
  
# External script assertion
- type: "script"
  command: "./custom-evaluator.py"
  args: ["{{output}}", "{{expected}}"]
  
# HTTP API assertion
- type: "api"
  endpoint: "https://api.example.com/evaluate"
  method: "POST"
  payload:
    text: "{{output}}"
    criteria: "quality"
  expect:
    status: 200
    body.score: ">0.8"
```

### Optimization Configuration

```yaml
optimization:
  # Optimization method
  method: "textgrad"  # textgrad, standard, hybrid
  
  # Basic settings
  iterations: 5
  convergence_threshold: 0.02  # Stop when improvement < threshold
  
  # TextGrad specific settings
  textgrad:
    temperature: 0.1
    gradient_accumulation: true
    reflection_depth: 3
    exploration_factor: 0.2
    
  # Standard optimization settings
  standard:
    mutation_rate: 0.1
    selection_pressure: 0.8
    
  # Hybrid optimization
  hybrid:
    initial_method: "standard"
    initial_iterations: 3
    final_method: "textgrad"
    final_iterations: 5
  
  # Multi-objective optimization
  objectives:
    - name: "quality"
      weight: 0.6
      metric: "llm_judge_score"
      direction: "maximize"
    - name: "cost"
      weight: 0.2
      metric: "cost_per_response"
      direction: "minimize"
    - name: "speed"
      weight: 0.2
      metric: "latency"
      direction: "minimize"
  
  # Constraints
  constraints:
    max_cost: 0.10
    max_latency: "10s"
    min_quality: 0.7
  
  # Advanced settings
  advanced:
    save_trajectory: true
    gradient_clipping: 1.0
    learning_rate_schedule: "cosine"
    early_stopping: true
    patience: 3
```

### Output Configuration

```yaml
output:
  # Output format
  format: "json"  # json, yaml, csv, html
  
  # Include options
  include:
    - "raw_responses"
    - "metadata"
    - "costs"
    - "timing"
    - "assertions"
  
  # Exclude sensitive data
  exclude:
    - "api_keys"
    - "internal_metadata"
  
  # File output
  file: "results.json"
  overwrite: true
  
  # Database storage
  database:
    enabled: true
    retention: "30d"
    
  # Real-time streaming
  streaming:
    enabled: true
    buffer_size: 100
    
  # Report generation
  reports:
    html:
      enabled: true
      template: "detailed"
      file: "report.html"
    
    csv:
      enabled: true
      file: "results.csv"
      
  # External integrations
  webhooks:
    - url: "https://api.slack.com/hooks/..."
      events: ["completion", "failure"]
    - url: "https://monitoring.example.com/webhook"
      events: ["completion"]
      payload:
        custom_field: "value"
```

### Advanced Configuration

```yaml
# Advanced features
advanced:
  # Caching
  cache:
    enabled: true
    ttl: "1h"
    storage: "disk"  # disk, memory, redis
    redis_url: "redis://localhost:6379/0"
  
  # Rate limiting
  rate_limiting:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
  
  # Retry configuration
  retry:
    max_attempts: 3
    initial_delay: "1s"
    max_delay: "30s"
    exponential_backoff: true
    jitter: true
  
  # Monitoring and observability
  monitoring:
    enabled: true
    metrics_port: 9090
    traces_endpoint: "http://jaeger:14268"
    logs_level: "info"
  
  # Security
  security:
    api_key_rotation: true
    request_signing: true
    tls_verify: true
    
  # Performance tuning
  performance:
    worker_pool_size: 10
    request_timeout: "30s"
    connection_pool_size: 100
```

---

## 🔌 Provider API

### Built-in Providers

#### OpenAI Provider

```yaml
- type: "openai"
  model: "gpt-4"
  config:
    api_key: "${OPENAI_API_KEY}"  # Environment variable
    organization: "org-123"        # Optional
    temperature: 0.7
    max_tokens: 1000
    top_p: 1.0
    frequency_penalty: 0.0
    presence_penalty: 0.0
    stop: ["###"]                  # Stop sequences
    logit_bias: {}                 # Token bias
    user: "user-123"               # User identifier
```

**Supported Models**:
- `gpt-4`
- `gpt-4-32k`
- `gpt-4-turbo`
- `gpt-4-vision-preview`
- `gpt-3.5-turbo`
- `gpt-3.5-turbo-16k`

#### Anthropic Provider

```yaml
- type: "anthropic"
  model: "claude-3-opus-20240229"
  config:
    api_key: "${ANTHROPIC_API_KEY}"
    temperature: 0.7
    max_tokens: 1000
    top_p: 1.0
    top_k: 50
    stop_sequences: ["###"]
```

**Supported Models**:
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240229`
- `claude-2.1`
- `claude-2.0`
- `claude-instant-1.2`

### Custom Provider Implementation

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/tmc/pe/internal/providers"
)

// CustomProvider implements the Provider interface
type CustomProvider struct {
    endpoint string
    apiKey   string
    model    string
}

// Generate implements the core generation method
func (p *CustomProvider) Generate(ctx context.Context, prompt string, opts providers.GenerateOptions) (*providers.GenerateResponse, error) {
    // Implementation details
    request := CustomRequest{
        Model:       p.model,
        Prompt:      prompt,
        Temperature: opts.Temperature,
        MaxTokens:   opts.MaxTokens,
    }
    
    response, err := p.callAPI(ctx, request)
    if err != nil {
        return nil, err
    }
    
    return &providers.GenerateResponse{
        Text:         response.Text,
        TokensUsed:   response.Usage.TotalTokens,
        Cost:         calculateCost(response.Usage),
        FinishReason: response.FinishReason,
        Metadata: map[string]interface{}{
            "model": p.model,
            "custom_field": response.CustomData,
        },
    }, nil
}

// GetCost returns cost per token for this provider
func (p *CustomProvider) GetCost() providers.CostInfo {
    return providers.CostInfo{
        InputTokens:  0.00001,  // $0.00001 per input token
        OutputTokens: 0.00002,  // $0.00002 per output token
    }
}

// Register the provider
func init() {
    providers.Register("custom", func(config map[string]interface{}) (providers.Provider, error) {
        return &CustomProvider{
            endpoint: config["endpoint"].(string),
            apiKey:   config["api_key"].(string),
            model:    config["model"].(string),
        }, nil
    })
}
```

---

## 🧠 Go API

### Core Evaluation API

```go
package main

import (
    "context"
    "time"
    
    "github.com/tmc/pe/internal/evaluator"
    "github.com/tmc/pe/internal/promptfoo"
)

func main() {
    // Create configuration
    config := promptfoo.Config{
        Prompts: []string{
            "What is the capital of {{country}}?",
        },
        Providers: []string{
            "openai:gpt-4",
        },
        Tests: []promptfoo.TestCase{
            {
                Vars: map[string]interface{}{
                    "country": "France",
                },
                Assert: []promptfoo.Assertion{
                    {
                        Type:  "contains",
                        Value: "Paris",
                    },
                },
            },
        },
    }
    
    // Run evaluation
    ctx := context.Background()
    timeout := time.Minute * 5
    results, err := evaluator.Evaluate(ctx, config, timeout, false, 4, true)
    if err != nil {
        panic(err)
    }
    
    // Process results
    for _, result := range results.Results {
        fmt.Printf("Prompt: %s\n", result.Prompt)
        fmt.Printf("Response: %s\n", result.Response)
        fmt.Printf("Score: %.2f\n", result.Score)
        fmt.Printf("Pass: %v\n", result.Pass)
    }
}
```

### Optimization API

```go
package main

import (
    "context"
    
    "github.com/tmc/pe/internal/metaprompt"
)

func main() {
    // Create optimizer
    optimizer := metaprompt.NewOptimizer(metaprompt.OptimizerConfig{
        Method:     metaprompt.TextGrad,
        Iterations: 5,
        Provider:   "openai:gpt-4",
        Temperature: 0.1,
    })
    
    // Define optimization objective
    objective := metaprompt.Objective{
        Description: "Generate clear, accurate summaries",
        TestCases: []metaprompt.TestCase{
            {
                Input: map[string]interface{}{
                    "text": "Long article content...",
                },
                ExpectedQualities: []string{
                    "concise",
                    "accurate",
                    "well-structured",
                },
            },
        },
    }
    
    // Optimize prompt
    ctx := context.Background()
    basePrompt := "Summarize this text: {{text}}"
    
    result, err := optimizer.Optimize(ctx, basePrompt, objective)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Original prompt: %s\n", basePrompt)
    fmt.Printf("Optimized prompt: %s\n", result.OptimizedPrompt)
    fmt.Printf("Improvement: %.2f%%\n", result.Improvement*100)
}
```

### Custom Assertion API

```go
package main

import (
    "context"
    
    "github.com/tmc/pe/internal/assertions"
)

// Custom assertion implementation
type SentimentAssertion struct {
    ExpectedSentiment string  `json:"expected_sentiment"`
    Confidence        float64 `json:"confidence"`
}

func (a *SentimentAssertion) Evaluate(ctx context.Context, output string, vars map[string]interface{}) (*assertions.Result, error) {
    // Use sentiment analysis library or API
    sentiment, confidence := analyzeSentiment(output)
    
    passed := sentiment == a.ExpectedSentiment && confidence >= a.Confidence
    
    return &assertions.Result{
        Passed: passed,
        Score:  confidence,
        Message: fmt.Sprintf("Sentiment: %s (confidence: %.2f)", sentiment, confidence),
        Metadata: map[string]interface{}{
            "detected_sentiment": sentiment,
            "confidence": confidence,
        },
    }, nil
}

// Register custom assertion
func init() {
    assertions.Register("sentiment", func() assertions.Assertion {
        return &SentimentAssertion{}
    })
}
```

---

## 🌐 REST API

When running PE as a service (`pe serve`), it exposes a REST API:

### Start API Server

```bash
pe serve --port 8080 --host 0.0.0.0
```

### Endpoints

#### `POST /v1/evaluate`
Evaluate prompts against test cases.

```bash
curl -X POST http://localhost:8080/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "prompts": ["What is the capital of {{country}}?"],
    "providers": ["openai:gpt-4"],
    "tests": [
      {
        "vars": {"country": "France"},
        "assert": [{"type": "contains", "value": "Paris"}]
      }
    ]
  }'
```

Response:
```json
{
  "id": "eval-123",
  "status": "completed",
  "results": [
    {
      "prompt": "What is the capital of {{country}}?",
      "response": "The capital of France is Paris.",
      "score": 1.0,
      "pass": true,
      "latency": 1200,
      "tokens": 15,
      "cost": 0.0003
    }
  ],
  "summary": {
    "total_tests": 1,
    "passed": 1,
    "failed": 0,
    "success_rate": 1.0
  }
}
```

#### `POST /v1/optimize`
Optimize a prompt using specified method.

```bash
curl -X POST http://localhost:8080/v1/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Summarize: {{text}}",
    "method": "textgrad",
    "iterations": 5,
    "provider": "openai:gpt-4",
    "objective": "Create clear, concise summaries"
  }'
```

#### `GET /v1/evaluations/{id}`
Get evaluation results by ID.

```bash
curl http://localhost:8080/v1/evaluations/eval-123
```

#### `GET /v1/health`
Health check endpoint.

```bash
curl http://localhost:8080/v1/health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h34m12s"
}
```

#### `GET /v1/metrics`
Prometheus-compatible metrics.

```bash
curl http://localhost:8080/v1/metrics
```

### API Authentication

```bash
# Set API key
export PE_API_KEY="your-api-key"

# Use in requests
curl -H "Authorization: Bearer your-api-key" \
  http://localhost:8080/v1/evaluate
```

---

## 🌍 Environment Variables

### Core Configuration

```bash
# API Keys
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_AI_API_KEY="..."

# PE Configuration
export PE_CONFIG_FILE="~/.pe/config.yaml"
export PE_LOG_LEVEL="info"
export PE_LOG_FILE="~/.pe/pe.log"
export PE_DATA_DIR="~/.pe"
export PE_CACHE_DIR="~/.pe/cache"

# Performance
export PE_MAX_CONCURRENCY="4"
export PE_DEFAULT_TIMEOUT="30s"
export PE_CACHE_TTL="1h"

# Database
export PE_DATABASE_URL="sqlite://~/.pe/pe.db"
export PE_DATABASE_POOL_SIZE="10"

# Monitoring
export PE_METRICS_ENABLED="true"
export PE_METRICS_PORT="9090"
export PE_TRACES_ENDPOINT="http://jaeger:14268"

# Security
export PE_API_KEY="your-api-key"
export PE_TLS_CERT_FILE="/path/to/cert.pem"
export PE_TLS_KEY_FILE="/path/to/key.pem"

# Development
export PE_DEBUG="false"
export PE_PROFILE="false"
export PE_EXPERIMENTAL_FEATURES="false"
```

### Provider-Specific Variables

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."
export OPENAI_ORG_ID="org-..."
export OPENAI_API_BASE="https://api.openai.com/v1"

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
export ANTHROPIC_API_URL="https://api.anthropic.com"

# Custom providers
export CUSTOM_PROVIDER_ENDPOINT="http://localhost:8000"
export CUSTOM_PROVIDER_API_KEY="..."
```

---

## 📊 Error Codes and Status

### Exit Codes

```
0   - Success
1   - General error
2   - Configuration error
3   - Authentication error
4   - Network error
5   - Timeout error
10  - Evaluation failed
11  - Assertion failed
20  - Optimization failed
30  - Database error
```

### HTTP Status Codes (API)

```
200 - Success
201 - Created
400 - Bad Request
401 - Unauthorized
403 - Forbidden
404 - Not Found
429 - Rate Limited
500 - Internal Server Error
503 - Service Unavailable
```

---

This API reference provides comprehensive documentation for all PE capabilities. For more examples and tutorials, see the [Examples Library](EXAMPLES_LIBRARY.md) and [Tutorials](TUTORIALS.md).