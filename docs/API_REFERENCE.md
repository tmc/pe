# 🛠️ PE API Reference: Complete Developer Guide

The comprehensive reference for PE's command-line interface, configuration options, and programmatic APIs. Everything you need to integrate PE into your development workflow and production systems.

## 📋 Table of Contents

- [🖥️ Command Line Interface](#-command-line-interface)
- [⚙️ Configuration Reference](#-configuration-reference)
- [🔌 Provider API](#-provider-api)
- [✅ Assertion Types](#-assertion-types)
- [🧠 Optimization Methods](#-optimization-methods)
- [💻 Go Programming API](#-go-programming-api)
- [🌐 REST API](#-rest-api)
- [🌍 Environment Variables](#-environment-variables)
- [🚨 Error Handling](#-error-handling)

---

## 🖥️ Command Line Interface

### Global Options

All PE commands support these global flags for consistent behavior:

```bash
pe [global-options] <command> [command-options] [arguments]

Global Options:
  --config string         Configuration file path (default: pe.yaml)
  --log-level string      Log level: debug, info, warn, error (default "info")
  --log-file string       Log output file (default: stderr)
  --verbose, -v           Enable verbose output with detailed information
  --quiet, -q             Suppress non-essential output for automation
  --json                  Output results in JSON format for processing
  --yaml                  Output results in YAML format
  --no-color              Disable colored output for scripts
  --help, -h              Show comprehensive help information
  --version               Show version and build information
  --profile               Enable performance profiling
  --debug                 Enable debug mode with trace information
```

### Core Evaluation Commands

#### `pe eval` - Comprehensive Prompt Evaluation

The cornerstone command for systematic prompt evaluation across providers and test cases.

```bash
pe eval [options] <config-file>

Essential Options:
  -o, --output string          Output file path (supports .json, .yaml, .csv)
  --save-db                   Save results to local database for pe view
  --dry-run                   Validate configuration without executing
  --stream                    Stream results as they complete (real-time)
  --watch                     Auto-rerun when files change

Performance Options:
  --max-concurrency int       Maximum concurrent evaluations (default 4)
  --timeout duration          Request timeout per call (default 30s)
  --retry-attempts int        Number of retry attempts (default 3)
  --retry-delay duration      Delay between retries (default 1s)

Filtering Options:
  --provider strings          Override/filter providers for this run
  --include string            Include test pattern (glob syntax)
  --exclude string            Exclude test pattern (glob syntax)
  --tags strings              Run only tests with specified tags

Control Options:
  --fail-fast                 Stop evaluation on first failure
  --progress                  Show detailed progress information
  --metrics                   Collect and display detailed performance metrics
  --cache                     Enable intelligent response caching
  --cache-ttl duration        Cache time-to-live (default 1h)

Quality Options:
  --statistical-analysis      Include statistical significance testing
  --confidence float          Confidence level for analysis (default 0.95)
  --baseline string           Compare against baseline results

Examples:
  # Basic evaluation
  pe eval config.yaml

  # Production evaluation with all features
  pe eval config.yaml --save-db --stream --metrics --statistical-analysis

  # High-performance evaluation
  pe eval config.yaml --max-concurrency 16 --cache --timeout 60s

  # Filtered evaluation for development
  pe eval config.yaml --include "*smoke*" --fail-fast --dry-run

  # Continuous monitoring
  pe eval config.yaml --watch --stream --metrics
```

#### `pe optimize` - Prompt Optimization

Prompt optimization using PE2, APEX, TextGrad, hybrid, and standard iterative
methods.

```bash
pe optimize [options] --prompt <prompt> | --prompt-file <file> | --config <config>

Input Options:
  --prompt string             Base prompt to optimize
  --prompt-file string        File containing prompt to optimize
  --config string             Configuration file with optimization settings

Optimization Methods (World-First Implementations):
  --method string             Method: pe2, apex, textgrad, evolve, fusion, hybrid
  --iterations int            Number of optimization iterations (default 5)
  --beam-width int            Beam search width for APEX (default 3)
  --population int            Population size for evolution (default 20)
  --generations int           Generations for evolution (default 10)

Research Options:
  --attention-flow            Enable TextGrad attention flow mapping
  --semantic-drift            Monitor semantic drift during optimization  
  --gradient-accumulation     Enable gradient accumulation for stability
  --multi-objective           Enable multi-objective optimization
  --pareto-analysis           Generate Pareto frontier analysis

Provider and Quality:
  --provider string           Provider for optimization (default "openai:gpt-4")
  --judge string              Provider for quality evaluation
  --objective string          Natural language optimization objective
  --constraints strings       Optimization constraints (cost:<val>, latency:<val>)

Output and Tracking:
  --output string             Output file for optimized prompt
  --save-trajectory           Save complete optimization trajectory
  --trace                     Enable detailed optimization tracing
  --genealogy                 Track evolutionary lineage (for evolve method)

Control:
  --convergence-threshold float  Stop when improvement < threshold (default 0.02)
  --temperature float         Optimization exploration temperature (default 0.1)
  --exploration-factor float  Exploration vs exploitation balance (default 0.2)
  --reflection-depth int      Depth of meta-analysis (default 2)

Examples:
  # PE2 Meta-Prompting (World's First)
  pe optimize --prompt "Analyze sentiment" --method pe2 --iterations 5

  # APEX Long Prompt Optimization (Industry First)
  pe optimize --prompt-file system-prompt.txt --method apex --beam-width 5 --iterations 8

  # TextGrad with attention-flow output
  pe optimize --prompt "Solve problems" --method textgrad --attention-flow --iterations 6

  # Evolutionary Multi-Objective Optimization
  pe optimize --prompt "Generate content" --method evolve --multi-objective --generations 20

  # Multi-Model Consensus Optimization
  pe optimize --prompt "Customer service" --method fusion --models gpt-4,claude-3,gemini

  # Complete Research Workflow
  pe optimize --prompt-file complex.txt --method hybrid --trace --genealogy --pareto-analysis
```

#### `pe view` - Results Viewer

Local web interface for viewing saved evaluation results.

```bash
pe view [eval-id]
pe view --file results.json
pe view --promptfoo [eval-id]

Data Sources:
  -f, --file string           Load results from a specific file
      --promptfoo             Delegate to the promptfoo CLI viewer

Interface Options:
  --port int                  Local web server port (default 8080)
  -y, --yes                   Pass -y to promptfoo with --promptfoo

Analysis Options:
  The current local viewer serves the saved JSON result and renders summary,
  per-test, and provider data in the browser.

Examples:
  # View saved local evaluations
  pe view

  # View a local file
  pe view --file results.json

  # Use a different local port
  pe view --file results.json --port 9000

  # Explicitly use promptfoo's viewer
  pe view --promptfoo eval-123 --yes
```

#### `pe interactive` - REPL Environment

Intelligent interactive environment for rapid prompt development and testing.

```bash
pe interactive [options]

Provider Configuration:
  --provider string           Default provider (e.g., "openai:gpt-4")
  --temperature float         Default temperature (default 0.7)
  --max-tokens int           Default max tokens (default 1000)
  --system string             Default system message

Session Management:
  --config string             Load initial configuration
  --save-session string       Auto-save session to file
  --load-session string       Load previous session
  --history-size int          Command history size (default 1000)

Options:
  --auto-optimize             Enable automatic prompt optimization suggestions
  --smart-completion          Enable AI-powered command completion
  --multi-provider            Enable multi-provider comparison mode
  --streaming                 Enable streaming responses

REPL Commands:
  /help                       Show all available commands
  /providers                  List and manage available providers
  /use <provider>            Switch to different provider
  /compare <providers>        Compare across multiple providers
  /config <key> <value>      Set configuration parameter
  /optimize [method]          Optimize current prompt with specified method
  /save <file>               Save current session
  /load <file>               Load session file
  /history [n]                Show command history (last n commands)
  /clear                      Clear screen and context
  /stats                      Show session statistics
  /benchmark [iterations]     Benchmark current prompt
  /export <format>            Export session data
  /templates                  Manage prompt templates
  /vars <key> <value>        Set template variables
  /debug [on|off]            Toggle debug mode
  /stream [on|off]           Toggle streaming mode
  /exit, /quit               Exit REPL

Examples:
  # Start with GPT-4 and optimization features
  pe interactive --provider openai:gpt-4 --auto-optimize

  # Start with configuration and multi-provider mode
  pe interactive --config my-config.yaml --multi-provider

  # Development session with full features
  pe interactive --streaming --smart-completion --save-session dev-session.json
```

### Optimization Commands

#### `pe experimental compose` - Component-Based Prompt Composition

Build prompts from reusable components.

```bash
pe experimental compose [options] <component-files...>

Component Management:
  --library-init              Initialize component library in current directory
  --add-component string      Add component to library with verification
  --list-components           List available components
  --verify-components         Verify component integrity and compatibility

Composition Options:
  --style string              Composition style: cot, few-shot, analytical, creative
  --target string             Target provider for optimization
  --coherence                 Enable coherence validation between components
  --flow-optimization         Optimize information flow between components
  --template string           Use composition template

Quality Control:
  --validate                  Validate composed prompt before output
  --optimize                  Auto-optimize composed prompt
  --test-suite                Generate test suite for composed prompt

Output Options:
  --output string             Output file for composed prompt
  --format string             Output format: text, yaml, json (default "text")
  --metadata                  Include composition metadata

Examples:
  # Initialize component library
  pe experimental compose --library-init

  # Basic composition with style
  pe experimental compose context.txt instruction.txt examples.txt --style cot --output composed.txt

  # Composition with optimization
  pe experimental compose components/*.txt --style few-shot --coherence --optimize --target gpt-4

  # Component management
  pe experimental compose --add-component expert-context.txt --category context --verify
```

#### `pe evolve` - Evolutionary Prompt Optimization

Population-based prompt optimization.

```bash
pe evolve [options] <base-prompt-file>

Population Configuration:
  --generations int           Number of evolutionary generations (default 25)
  --population int            Population size (default 20)
  --elite-size int            Elite population to preserve (default 4)
  --diversity-threshold float Minimum population diversity (default 0.3)

Evolutionary Operators:
  --operators strings         Mutation operators: rephrase,expand,prune,crossover
  --adaptive-rates            Enable adaptive mutation rates
  --structure-aware           Structure-aware evolutionary operations
  --semantic-preservation     Maintain semantic meaning during evolution

Multi-Objective Optimization:
  --multi-objective strings   Objectives: accuracy,latency,cost,creativity,clarity
  --nsga-ii                   Use NSGA-II multi-objective algorithm
  --pareto-analysis           Generate Pareto frontier analysis
  --weights strings           Objective weights if not using Pareto

Options:
  --trace-genealogy           Track complete evolutionary lineage
  --statistical-validation    Include statistical significance testing
  --convergence-detection     Auto-detect convergence and early stopping
  --diversity-preservation    Maintain genetic diversity through niching

Quality and Constraints:
  --fitness-function string   Custom fitness function
  --constraints strings       Hard constraints on evolution
  --quality-gates             Enforce quality gates during evolution

Examples:
  # Basic evolutionary optimization
  pe evolve baseline.txt --generations 30 --population 25

  # Multi-objective optimization with Pareto analysis
  pe evolve prompt.txt --multi-objective accuracy,latency,cost --nsga-ii --pareto-analysis

  # Evolution with traceability
  pe evolve complex-prompt.txt --trace-genealogy --statistical-validation --adaptive-rates

  # Research-grade evolution
  pe evolve prompt.txt --structure-aware --diversity-preservation --convergence-detection
```

#### Multi-Model Fusion Status

`pe fusion` is not currently exposed as a standalone CLI command.
Use `pe optimize`/`pe evolve` for available optimization flows, and check
`pe experimental --help` for currently exposed research commands.

```bash
pe optimize --help
pe evolve --help
pe experimental --help
```

### Pipeline Processing Commands

#### `pe ask` - Pipeline-Friendly Single Queries

Optimized for Unix pipeline integration and automation workflows.

```bash
pe ask [options] [prompt]

Provider Options:
  --provider string           LLM provider (default "openai:gpt-3.5-turbo")
  --temperature float         Response temperature (default 0.7)
  --max-tokens int           Maximum response tokens (default 1000)
  --system string             System message for context

Input/Output Control:
  --format string             Output format: text, json, yaml (default "text")
  --stream                    Enable streaming output
  --no-newline                Suppress trailing newline for piping
  --input-file string         Read prompt from file instead of stdin

Options:
  --cache                     Enable response caching
  --retry-on-failure          Retry on API failures
  --timeout duration          Request timeout (default 30s)
  --metrics                   Include performance metrics in output

Examples:
  # Basic pipeline usage
  echo "What is AI?" | pe ask

  # Provider with system context
  echo "Review this code" | pe ask --provider anthropic:claude-3-opus --system "You are a senior developer"

  # JSON output for further processing
  echo "Analyze sentiment" | pe ask --format json | jq '.response'

  # File processing pipeline
  cat document.txt | pe ask --system "Summarize this document" --stream

  # Batch processing with caching
  pe ask --input-file prompts.txt --cache --format json --provider gpt-4
```

#### `pe stream` - Result Processing

Process evaluation results as Unix streams with rich filtering and transformation.

```bash
pe stream [options]

Field Selection:
  --select strings            Select specific fields: response,score,latency,cost,tokens
  --exclude strings           Exclude specific fields from output
  --transform string          Transform fields using expressions

Output Formatting:
  --format string             Output format: json,csv,tsv,table (default "json")
  --delimiter string          Delimiter for CSV/TSV output
  --headers                   Include headers for CSV/TSV
  --pretty                    Pretty-print JSON output

Performance Options:
  --buffer-size int           Stream buffer size (default 1000)
  --batch-size int            Batch processing size for efficiency
  --parallel                  Enable parallel processing

Filtering and Analysis:
  --filter string             Filter expression (e.g., "score > 0.8")
  --group-by string           Group results by field
  --aggregate string          Aggregate function: count,sum,avg,min,max

Examples:
  # Basic field selection
  pe eval config.yaml | pe stream --select response,score,latency

  # CSV output for spreadsheet analysis
  pe eval config.yaml | pe stream --format csv --headers > results.csv

  # Filtering and grouping
  pe eval config.yaml | pe stream --filter "score>0.8" --group-by provider --format table

  # Performance analysis
  pe eval config.yaml | pe stream --select latency,cost --aggregate avg --format json

  # Real-time monitoring
  pe eval config.yaml --stream | pe stream --select score --filter "score<0.7" --format table
```

#### `pe filter` - Result Filtering

Filter evaluation results by provider, score, assertion status, and text.

```bash
pe filter [options]

Success/Failure Filtering:
  --success                   Include only successful results
  --failure                   Include only failed results
  --partial                   Include partially successful results

Numeric Filtering:
  --min-score float           Minimum score threshold
  --max-score float           Maximum score threshold
  --score-range string        Score range (e.g., "0.7-0.9")

Performance Filtering:
  --min-latency duration      Minimum response latency
  --max-latency duration      Maximum response latency
  --latency-percentile int    Filter by latency percentile

Cost Filtering:
  --min-cost float            Minimum cost threshold
  --max-cost float            Maximum cost threshold
  --cost-budget float         Total cost budget constraint

Content Filtering:
  --contains string           Response must contain text
  --not-contains string       Response must not contain text
  --regex string              Response must match regex pattern
  --min-length int            Minimum response length
  --max-length int            Maximum response length

Provider/Prompt Filtering:
  --provider string           Filter by specific provider
  --prompt string             Filter by prompt pattern
  --tags strings              Filter by test tags

Filtering:
  --expression string         Custom filter expression
  --statistical-outliers      Filter statistical outliers
  --quality-threshold float   Filter by quality metrics

Examples:
  # Basic success filtering with performance constraints
  pe eval config.yaml | pe filter --success --max-latency 5s --max-cost 0.02

  # Content-based filtering
  pe eval config.yaml | pe filter --contains "Paris" --not-contains "London" --provider gpt-4

  # Statistical filtering
  pe eval config.yaml | pe filter --score-range "0.8-1.0" --statistical-outliers

  # Complex expression filtering
  pe eval config.yaml | pe filter --expression "score > 0.8 AND latency < 3000 AND cost < 0.01"

  # Quality-based filtering
  pe eval config.yaml | pe filter --quality-threshold 0.9 --tags production
```

#### `pe analyze` - Statistical Analysis

Statistical analysis for evaluation result streams.

```bash
pe analyze [options]

Core Metrics:
  --metric string             Primary metric: score,latency,cost,tokens,quality
  --metrics strings           Multiple metrics for correlation analysis
  --custom-metrics strings    Custom metric definitions

Grouping and Segmentation:
  --group-by string           Group results by: provider,prompt,test,tag
  --segment-by string         Segment analysis by categorical variables
  --cohort-analysis           Perform cohort-based analysis

Statistical Analysis:
  --percentiles ints          Calculate percentiles (default [50,90,95,99])
  --confidence float          Confidence level (default 0.95)
  --statistical-tests         Run appropriate statistical tests
  --effect-size               Calculate effect sizes for comparisons
  --power-analysis            Perform statistical power analysis

Analytics:
  --trend-analysis            Perform time-series trend analysis
  --correlation               Calculate correlation matrices
  --regression                Perform regression analysis
  --clustering                Cluster analysis of results
  --outlier-detection         Outlier detection methods
  --anomaly-detection         Detect anomalous patterns

Visualization and Export:
  --charts                    Generate statistical charts
  --export string             Export analysis results (json,csv,html)
  --report string             Generate comprehensive analysis report

Examples:
  # Basic performance analysis
  pe eval config.yaml | pe analyze --metric latency --percentiles 50,90,95,99

  # Provider comparison with statistical testing
  pe eval config.yaml | pe analyze --group-by provider --statistical-tests --effect-size

  # Comprehensive quality analysis
  pe eval config.yaml | pe analyze --metrics score,latency,cost --correlation --regression

  # Pattern analysis
  pe eval config.yaml | pe analyze --clustering --anomaly-detection --trend-analysis

  # Research-grade analysis with full reporting
  pe eval config.yaml | pe analyze --metric score --power-analysis --charts --report analysis.html
```

### Development and Utility Commands

#### `pe benchmark` - Performance Benchmarking

Comprehensive benchmarking with statistical rigor and detailed performance analysis.

```bash
pe benchmark [options] <config-file>

Benchmark Configuration:
  --iterations int            Number of benchmark iterations (default 10)
  --warmup int                Warmup iterations to exclude (default 2)
  --cooldown int              Cooldown period between iterations (default 1s)

Concurrency Testing:
  --concurrency int           Concurrent requests (default 4)
  --ramp-up duration          Gradual ramp-up period (default 0s)
  --max-concurrency int       Maximum concurrency to test
  --concurrency-steps int     Steps for concurrency testing

Statistical Analysis:
  --percentiles ints          Percentiles to calculate (default [50,90,95,99])
  --confidence float          Confidence level (default 0.95)
  --statistical-analysis      Include comprehensive statistical analysis
  --outlier-handling string   Outlier handling: remove,keep,winsorize (default "keep")

Provider Comparison:
  --compare-providers         Benchmark all configured providers
  --provider-matrix           Create provider comparison matrix
  --cost-analysis             Include detailed cost analysis
  --quality-benchmarks        Include quality metrics in benchmarks

Output and Reporting:
  --format string             Output format: table,json,csv,html (default "table")
  --output string             Save results to file
  --charts                    Generate performance charts
  --report string             Generate comprehensive benchmark report

Options:
  --load-testing              Perform load testing analysis
  --stress-testing            Stress test with increasing load
  --endurance-testing         Long-duration endurance testing
  --memory-profiling          Include memory usage profiling

Examples:
  # Basic performance benchmark
  pe benchmark config.yaml --iterations 100 --statistical-analysis

  # Comprehensive provider comparison
  pe benchmark config.yaml --compare-providers --cost-analysis --quality-benchmarks

  # Load testing
  pe benchmark config.yaml --load-testing --max-concurrency 16 --ramp-up 30s

  # Research-grade benchmarking
  pe benchmark config.yaml --stress-testing --memory-profiling --charts --report benchmark.html
```

#### `pe test` - Testing Framework

Comprehensive testing capabilities with property-based testing, regression testing, and A/B testing.

```bash
pe test [options] <command> [config-file]

Test Types:
  property                    Property-based testing for prompt reliability
  regression                  Regression testing against baseline
  cross-validation           Cross-validation testing
  comprehensive              Full test suite with all methods
  a-b                        A/B testing between prompt variants

Property-Based Testing:
  --properties strings        Properties to test: consistency,robustness,fairness
  --test-cases int            Number of generated test cases (default 100)
  --shrinking                 Enable test case shrinking for failures
  --seed int                  Random seed for reproducible testing

Regression Testing:
  --baseline string           Baseline results for comparison
  --threshold float           Regression threshold (default 0.05)
  --strict                    Strict regression testing mode

Cross-Validation:
  --folds int                 Number of cross-validation folds (default 5)
  --stratified                Use stratified cross-validation
  --shuffle                   Shuffle data before folding

Statistical Testing:
  --statistical-validation    Include statistical significance testing
  --multiple-comparisons      Adjust for multiple comparisons
  --power-analysis            Perform statistical power analysis

A/B Testing:
  --variant-a string          Variant A configuration
  --variant-b string          Variant B configuration
  --sample-size int           Required sample size per variant
  --significance-level float  Statistical significance level (default 0.05)

Examples:
  # Property-based testing for reliability
  pe test property config.yaml --properties consistency,robustness --test-cases 200

  # Regression testing against baseline
  pe test regression config.yaml --baseline baseline.json --threshold 0.03

  # Cross-validation for generalization
  pe test cross-validation config.yaml --folds 10 --stratified

  # A/B testing between variants
  pe test a-b --variant-a config-a.yaml --variant-b config-b.yaml --sample-size 1000

  # Comprehensive testing suite
  pe test comprehensive config.yaml --statistical-validation --power-analysis
```

---

## ⚙️ Configuration Reference

### Complete Configuration Schema

```yaml
# config.yaml - Complete PE Configuration
description: "Comprehensive prompt evaluation configuration"
version: "1.0"  # Configuration schema version
metadata:
  author: "Your Name"
  created: "2024-01-01"
  updated: "2024-01-15"
  tags: ["production", "quality-assurance"]

# Core configuration sections
prompts: []      # Prompt definitions and templates
providers: []    # LLM provider configurations  
tests: []        # Test cases and scenarios
metrics: []      # Custom metrics and scoring
optimization: {} # Optimization settings
output: {}       # Output and reporting configuration
advanced: {}     # Additional features and tuning
```

### Prompt Configuration

```yaml
prompts:
  # Simple string prompt with variables
  - "What is the capital of {{country}}?"
  
  # Comprehensive prompt object with metadata
  - id: "expert-analysis"
    content: |
      You are an expert in {{domain}} with {{years}} years of experience.
      
      Background Context:
      {{context}}
      
      Your task is to {{task}} with the following requirements:
      1. {{requirement_1}}
      2. {{requirement_2}}
      3. {{requirement_3}}
      
      Please provide a comprehensive analysis that demonstrates your expertise.
      
      Question: {{question}}
    
    # Rich metadata for organization and tracking
    description: "Expert-level analysis prompt template"
    category: "analysis"
    tags: ["expert", "comprehensive", "structured"]
    version: "2.1"
    author: "Prompt Engineering Team"
    complexity: "high"
    estimated_tokens: 150
    
    # Variable definitions and validation
    variables:
      domain:
        type: "string"
        required: true
        description: "Field of expertise"
        examples: ["machine learning", "economics", "biology"]
      years:
        type: "integer"
        required: true
        min: 5
        max: 50
        description: "Years of experience"
      context:
        type: "string"
        required: false
        max_length: 500
        description: "Additional background context"
      task:
        type: "string"
        required: true
        description: "Specific task to perform"
      question:
        type: "string"
        required: true
        max_length: 1000
        description: "The question to analyze"
    
    # Performance and quality hints
    optimization_hints:
      preferred_providers: ["openai:gpt-4", "anthropic:claude-3-opus"]
      avoid_providers: ["openai:gpt-3.5-turbo"]
      temperature_range: [0.1, 0.3]
      max_tokens_range: [500, 1500]
    
  # Multi-modal prompt for vision models
  - id: "image-analysis"
    type: "multimodal"
    content: "Analyze this image and describe what you see: {{image_description}}"
    modalities:
      - type: "image"
        source: "{{image_url}}"
        format: ["jpg", "png", "webp"]
        max_size: "10MB"
      - type: "text"
        source: "{{additional_context}}"
        required: false
    
    # Provider compatibility
    compatible_providers:
      - "openai:gpt-4-vision-preview"
      - "anthropic:claude-3-opus"
    
  # Dynamic prompt with conditional logic
  - id: "adaptive-response"
    content: |
      {% if audience == "technical" %}
      As a technical expert, provide a detailed technical explanation of {{topic}}.
      Include implementation details, code examples, and best practices.
      {% elif audience == "business" %}
      As a business consultant, explain {{topic}} in terms of business value,
      ROI, and strategic implications.
      {% else %}
      Provide a clear, accessible explanation of {{topic}} suitable for a general audience.
      {% endif %}
      
      Topic: {{topic}}
      Detail Level: {{detail_level}}
    
    # Template engine configuration
    template_engine: "jinja2"
    template_options:
      strict_undefined: true
      auto_escape: false

# Prompt template library for reusability
templates:
  expert-response: |
    You are an expert in {{domain}}.
    Context: {{context}}
    
    Question: {{question}}
    
    Please provide an expert-level response that includes:
    1. Direct answer
    2. Supporting evidence  
    3. Relevant examples
    4. Potential implications
  
  step-by-step: |
    Let's solve this step by step:
    
    Problem: {{problem}}
    
    Step 1: {{step_1}}
    Step 2: {{step_2}}
    Step 3: {{step_3}}
    
    Solution: {{solution}}
  
  creative-writing: |
    Write a {{genre}} story with the following elements:
    - Setting: {{setting}}
    - Main character: {{character}}
    - Conflict: {{conflict}}
    - Style: {{style}}
    
    Length: {{length}} words
    Tone: {{tone}}

# Prompt inheritance and composition
prompt_inheritance:
  base_expert:
    content: "You are an expert in {{domain}}."
    variables: ["domain"]
  
  derived_analyst:
    inherits: "base_expert"
    content: |
      {{parent_content}}
      
      Your role is to analyze {{subject}} and provide insights on {{aspects}}.
    additional_variables: ["subject", "aspects"]
```

### Comprehensive Provider Configuration

```yaml
providers:
  # Production OpenAI configuration
  - id: "production-gpt4"
    type: "openai"
    model: "gpt-4"
    description: "Production GPT-4 with conservative settings"
    
    # Core model parameters
    config:
      api_key: "${OPENAI_API_KEY}"
      organization: "${OPENAI_ORG_ID}"
      temperature: 0.1
      max_tokens: 1000
      top_p: 0.95
      frequency_penalty: 0.0
      presence_penalty: 0.0
      stop: ["###", "END", "STOP"]
      
    # Additional configuration
    advanced:
      request_timeout: 30
      retry_attempts: 3
      retry_delay: 2
      exponential_backoff: true
      jitter: true
      
    # Cost and performance tracking
    cost_tracking:
      input_cost_per_token: 0.00003
      output_cost_per_token: 0.00006
      currency: "USD"
      
    # Rate limiting and quotas
    rate_limiting:
      requests_per_minute: 100
      tokens_per_minute: 40000
      burst_allowance: 10
      
    # Monitoring and alerting
    monitoring:
      enabled: true
      latency_threshold: 5000  # milliseconds
      error_rate_threshold: 0.05
      cost_threshold: 10.0  # dollars per day
      
  # Development and experimentation provider
  - id: "dev-claude"
    type: "anthropic"
    model: "claude-3-sonnet-20240229"
    description: "Development Claude for experimentation"
    
    config:
      api_key: "${ANTHROPIC_API_KEY}"
      temperature: 0.7
      max_tokens: 1500
      top_p: 0.9
      top_k: 100
      
    # Provider-specific features
    features:
      system_message_support: true
      function_calling: false
      json_mode: false
      streaming: true
      
  # Custom local provider
  - id: "local-llama"
    type: "custom"
    description: "Local Llama model via Ollama"
    
    endpoint: "http://localhost:11434/api/generate"
    model: "llama2:13b"
    
    # Custom headers and authentication
    headers:
      "Content-Type": "application/json"
      "Authorization": "Bearer ${LOCAL_API_KEY}"
      "Custom-Header": "pe-client"
      
    # Request/response transformation
    request_transform: |
      {
        "model": "{{model}}",
        "prompt": "{{prompt}}",
        "options": {
          "temperature": {{temperature}},
          "num_predict": {{max_tokens}}
        }
      }
      
    response_transform: |
      {
        "text": "{{response}}",
        "tokens": {{eval_count}},
        "done": {{done}}
      }
      
    # Health check configuration
    health_check:
      endpoint: "/api/tags"
      interval: 30
      timeout: 5
      
  # Load-balanced provider cluster
  - id: "openai-cluster"
    type: "load_balanced"
    description: "Load-balanced OpenAI cluster"
    
    # Multiple endpoints for load balancing
    endpoints:
      - endpoint: "https://api.openai.com/v1"
        weight: 60
        api_key: "${OPENAI_API_KEY_1}"
      - endpoint: "https://api.openai.com/v1"
        weight: 40
        api_key: "${OPENAI_API_KEY_2}"
        
    # Load balancing strategy
    strategy: "weighted_round_robin"  # round_robin, weighted_round_robin, least_connections
    
    # Failover configuration
    failover:
      enabled: true
      retry_on_failure: true
      circuit_breaker: true
      health_check_interval: 60

# Global provider defaults
provider_defaults:
  timeout: 30
  retry_attempts: 3
  temperature: 0.7
  max_tokens: 1000
  
  # Global monitoring settings
  monitoring:
    enabled: true
    metrics_collection: true
    trace_requests: false
    
  # Global security settings
  security:
    tls_verify: true
    request_signing: false
    api_key_rotation: false
```

### Test Configuration

```yaml
tests:
  # Comprehensive test case with all features
  - id: "comprehensive-geography-test"
    description: "Comprehensive geography knowledge test with multiple validation layers"
    category: "knowledge"
    tags: ["geography", "factual", "production"]
    priority: "high"
    
    # Test data and variables
    vars:
      country: "France"
      continent: "Europe"
      population_range: "60-70 million"
      
    # Multiple assertion layers
    assert:
      # Content validation
      - type: "contains"
        value: "Paris"
        description: "Must mention the correct capital"
        weight: 0.4
        
      - type: "not-contains"
        value: ["Lyon", "Marseille", "Toulouse"]
        description: "Should not confuse with other major cities"
        weight: 0.2
        
      # Structure and quality validation
      - type: "length"
        min: 50
        max: 200
        description: "Appropriate response length"
        weight: 0.1
        
      - type: "readability"
        metric: "flesch_reading_ease"
        min: 60
        description: "Accessible reading level"
        weight: 0.1
        
      # Content analysis
      - type: "factuality"
        threshold: 0.9
        knowledge_cutoff: "2024-01-01"
        description: "Factual accuracy verification"
        weight: 0.2
        
      # Performance constraints
      - type: "latency"
        max: "5s"
        description: "Response time requirement"
        critical: true
        
      - type: "cost"
        max: 0.02
        description: "Cost efficiency requirement"
        
    # Test-specific configuration
    options:
      retries: 2
      timeout: "10s"
      fail_fast: false
      collect_metrics: true
      
    # Conditional execution
    conditions:
      - provider: "openai:gpt-4"
        skip_if: "cost > 0.05"
      - environment: "development"
        reduce_iterations: 0.5
        
  # Multi-turn conversation test
  - id: "customer-service-conversation"
    description: "Multi-turn customer service interaction test"
    type: "conversation"
    
    conversation:
      - turn: 1
        user: "Hi, I have a problem with my order #12345"
        assistant_assert:
          - type: "contains"
            value: ["order", "12345"]
            description: "Acknowledge order number"
          - type: "sentiment"
            value: "positive"
            min_confidence: 0.7
            description: "Maintain positive tone"
            
      - turn: 2
        user: "It was supposed to arrive yesterday but didn't"
        assistant_assert:
          - type: "contains"
            value: ["sorry", "apologize", "understand"]
            description: "Show empathy"
          - type: "contains"
            value: ["track", "investigate", "check"]
            description: "Offer to help"
            
      - turn: 3
        user: "Can you refund it?"
        assistant_assert:
          - type: "contains"
            value: ["refund", "policy", "process"]
            description: "Address refund request"
          - type: "not-contains"
            value: ["no", "cannot", "impossible"]
            description: "Avoid negative language"
            
    # Conversation-specific settings
    conversation_config:
      context_window: 3  # Remember last 3 turns
      personality_consistency: true
      tone_consistency: true
      
  # Property-based test generation
  - id: "math-problem-property-test"
    description: "Property-based testing for math problem solving"
    type: "property_based"
    
    # Property definitions
    properties:
      - name: "arithmetic_consistency"
        description: "Addition should be commutative"
        generator: |
          def generate_test_case():
              a = random.randint(1, 100)
              b = random.randint(1, 100)
              return {
                  "problem_1": f"What is {a} + {b}?",
                  "problem_2": f"What is {b} + {a}?",
                  "expected_result": a + b
              }
        validator: |
          def validate(response_1, response_2, expected):
              result_1 = extract_number(response_1)
              result_2 = extract_number(response_2)
              return result_1 == result_2 == expected
              
      - name: "scaling_consistency"
        description: "Answers should scale proportionally"
        test_cases: 50
        
    # Property test configuration
    property_config:
      max_examples: 100
      shrinking: true
      seed: 42
      timeout_per_test: 30
      
  # A/B testing configuration
  - id: "prompt-variant-test"
    description: "A/B test between prompt variants"
    type: "ab_test"
    
    variants:
      - id: "variant_a"
        prompt: "Solve this problem step by step: {{problem}}"
        description: "Direct step-by-step instruction"
        
      - id: "variant_b"
        prompt: "Let's think about this problem carefully and solve it systematically: {{problem}}"
        description: "More conversational approach"
        
    # A/B test configuration
    ab_config:
      sample_size_per_variant: 100
      significance_level: 0.05
      power: 0.8
      minimum_effect_size: 0.1
      stratification: ["difficulty_level", "topic"]
      
    # Success metrics for A/B test
    success_metrics:
      primary: "accuracy"
      secondary: ["user_satisfaction", "response_time"]

# Test suite configuration
test_suite:
  # Execution settings
  execution:
    parallel: true
    max_concurrency: 8
    timeout: "5m"
    retry_failed: true
    
  # Quality gates
  quality_gates:
    minimum_pass_rate: 0.95
    maximum_cost_per_test: 0.05
    maximum_latency_p95: 5000
    
  # Reporting and analysis
  reporting:
    generate_html_report: true
    include_statistical_analysis: true
    confidence_level: 0.95
    export_raw_data: true
    
  # Integration with external systems
  integrations:
    slack_webhook: "${SLACK_WEBHOOK_URL}"
    jira_project: "PE"
    github_pr_comments: true
```

---

This API reference covers the current PE API surface. Planned and aspirational
features belong under `docs/future/` until they are implemented and validated.
