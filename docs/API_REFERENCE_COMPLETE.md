# PE API Reference: Complete Technical Documentation

Comprehensive API reference for PE, the world's most advanced prompt engineering toolkit. This documentation covers all commands, options, configurations, and programming interfaces.

## 📋 Table of Contents

- [Command Line Interface](#command-line-interface)
- [Configuration File Reference](#configuration-file-reference)
- [Go API Reference](#go-api-reference)
- [REST API Reference](#rest-api-reference)
- [Pipeline Processing](#pipeline-processing)
- [Advanced Usage Patterns](#advanced-usage-patterns)

## 🖥️ Command Line Interface

### Core Commands

#### `pe eval` - Evaluation Engine

Evaluate prompt configurations against LLM providers with advanced metrics.

```bash
pe eval [config.yaml] [flags]
```

**Flags:**
```bash
  -c, --config string           Configuration file path
  -o, --output string          Output file for results (default: stdout)
      --format string          Output format: json|yaml|csv|table (default: json)
      --metrics strings        Evaluation metrics: bleu,rouge,meteor,bertscore,g-eval,unieval
      --providers strings      LLM providers to test: openai,anthropic,google,custom
      --parallel               Enable parallel evaluation
      --max-workers int        Maximum concurrent workers (default: 5)
      --stream                 Stream results in real-time
      --save-db               Save results to database
      --cache                 Enable result caching
      --cache-ttl duration    Cache time-to-live (default: 24h)
      --timeout duration      Request timeout (default: 30s)
      --retry int             Retry attempts (default: 3)
      --verbose               Verbose output
      --quiet                 Suppress non-essential output
```

**Examples:**
```bash
# Basic evaluation
pe eval config.yaml

# Advanced evaluation with all metrics
pe eval config.yaml --metrics bleu,rouge,bertscore,g-eval --parallel --save-db

# Streaming evaluation with filtering
pe eval config.yaml --stream --parallel | pe filter --success --min-score 0.8

# Multi-provider comparison
pe eval config.yaml --providers openai:gpt-4,anthropic:claude-3,google:gemini-pro
```

#### `pe optimize` - Prompt Optimization

Optimize prompts using cutting-edge 2024-2025 research methods.

```bash
pe optimize [flags]
```

**Flags:**
```bash
  -p, --prompt string          Prompt text to optimize
  -f, --prompt-file string     File containing prompt to optimize
  -m, --method string          Optimization method: pe2,apex,textgrad,semantic,evolve,fusion
  -i, --iterations int         Number of optimization iterations (default: 5)
  -o, --output string          Output file for optimized prompt
      --temperature float      LLM temperature for optimization (default: 0.7)
      --max-tokens int         Maximum tokens for optimization (default: 2000)
      --provider string        LLM provider for optimization (default: openai:gpt-4)
      --beam-width int         Beam width for APEX method (default: 3)
      --population int         Population size for evolutionary method (default: 20)
      --generations int        Generations for evolutionary method (default: 10)
      --models strings         Models for fusion method
      --consensus string       Consensus strategy: weighted,majority,reflection (default: weighted)
      --multi-objective        Enable multi-objective optimization
      --objectives strings     Optimization objectives: accuracy,latency,cost,robustness
      --pareto-analysis        Perform Pareto frontier analysis
      --trace                  Enable optimization tracing
      --save-history           Save optimization history
      --statistical-validation Enable statistical validation
```

**Examples:**
```bash
# PE2 meta-prompt optimization
pe optimize --prompt "Analyze sentiment" --method pe2 --iterations 5

# APEX long prompt optimization
pe optimize --prompt-file system-prompt.txt --method apex --beam-width 5 --iterations 8

# TextGrad semantic optimization
pe optimize --prompt "Solve problems" --method textgrad --iterations 6 --trace

# Evolutionary multi-objective optimization
pe evolve baseline.txt --generations 25 --population 20 --multi-objective --objectives accuracy,latency,cost

# Multi-model consensus optimization
pe fusion prompt.txt --models gpt-4,claude-3,gemini-pro --consensus reflection
```

#### `pe semantic` - Semantic Backpropagation (2025 Research)

Advanced semantic gradient descent optimization using 2025 KAUST/IDSIA research.

```bash
pe semantic [command] [flags]
```

**Subcommands:**
- `backprop` - Semantic backpropagation optimization
- `descent` - Semantic gradient descent
- `gaso` - Graph-based Agentic System Optimization

**Flags:**
```bash
      --prompt string          Target prompt for optimization
      --objective string       Optimization objective
      --iterations int         Number of iterations (default: 5)
      --learning-rate float    Learning rate for gradient descent (default: 0.1)
      --adaptive              Enable adaptive learning rate
      --convergence float     Convergence threshold (default: 0.001)
      --system string         System definition file for GASO
      --multi-objective       Enable multi-objective optimization
      --gradient-strength     Analyze gradient strength
      --attention-flow        Enable attention flow mapping
      --semantic-drift        Monitor semantic drift
```

**Examples:**
```bash
# Semantic backpropagation
pe semantic backprop --prompt "prompt" --objective "accuracy" --iterations 5

# Semantic gradient descent with adaptive learning
pe semantic descent --objective "goal" --learning-rate 0.1 --adaptive

# GASO for system optimization
pe semantic gaso --system definition.json --multi-objective --objectives performance,cost
```

#### `pe compose` - Component-Based Composition

DSPy-style prompt composition with advanced features.

```bash
pe compose [components...] [flags]
```

**Flags:**
```bash
      --style string          Composition style: default,cot,few-shot,structured,conversational,dspy
      --coherence             Enable coherence analysis
      --validation-gate       Enable validation gates
      --optimize              Optimize after composition
      --library-init          Initialize component library
      --add-component string  Add component to library
      --category string       Component category
      --verify                Verify component compatibility
      --version string        Component version
      --dependency-check      Check component dependencies
      --semantic-analysis     Enable semantic coherence analysis
      --type-safety           Enable type-safe composition
      --template-engine       Use template engine for composition
```

**Examples:**
```bash
# Initialize component library
pe compose --library-init

# Style-specific composition
pe compose context.txt instruction.txt examples.txt --style cot --optimize

# Component library management
pe compose --add-component context-banking.txt --category context --verify

# Advanced composition with validation
pe compose components/ --style few-shot --coherence --validation-gate
```

#### `pe security` - Security Testing

Comprehensive security testing including OWASP LLM Top 10.

```bash
pe security [command] [flags]
```

**Subcommands:**
- `test` - Run security tests
- `monitor` - Real-time security monitoring
- `report` - Generate security reports

**Flags:**
```bash
      --target string         Target prompt or system to test
      --owasp-complete        Run complete OWASP LLM Top 10 assessment
      --categories strings    Security categories: prompt_injection,data_leakage,bias,toxicity
      --severity string       Test severity: basic,comprehensive,deep (default: comprehensive)
      --adversarial           Enable adversarial testing
      --custom-tests string   Custom test vectors file
      --adaptive              Adaptive attack strategies
      --realtime              Real-time monitoring mode
      --alerts string         Alert levels: low,medium,high,critical
      --webhook string        Webhook URL for alerts
      --compliance strings    Compliance frameworks: owasp,nist,iso27001
      --format string         Report format: json,pdf,html (default: json)
```

**Examples:**
```bash
# Complete OWASP assessment
pe security test --target system_prompt.txt --owasp-complete --severity comprehensive

# Focused security testing
pe security test --target prompt.txt --categories prompt_injection,data_leakage --adversarial

# Real-time monitoring
pe security monitor --realtime --alerts high --webhook slack://security-alerts

# Compliance reporting
pe security test --target system.txt --compliance owasp,nist --format pdf
```

#### `pe metrics` - Advanced Evaluation Metrics

State-of-the-art evaluation metrics implementation.

```bash
pe metrics [flags]
```

**Flags:**
```bash
      --type string           Metric type: bleu,rouge,meteor,bertscore,g-eval,unieval,custom
      --generated string      Generated text file
      --reference string      Reference text file
      --criteria strings      Evaluation criteria for G-Eval
      --model string          Model for BERTScore (default: bert-base-uncased)
      --language string       Language for metrics (default: en)
      --output string         Output file for metrics
      --all                   Compute all available metrics
      --confidence-intervals  Compute confidence intervals
      --bootstrap int         Bootstrap samples for CI (default: 1000)
      --significance-test     Perform significance testing
      --effect-size           Compute effect size measures
```

**Examples:**
```bash
# BLEU score evaluation
pe metrics --type bleu --generated response.txt --reference expected.txt

# All metrics with confidence intervals
pe metrics --all --generated responses/ --reference references/ --confidence-intervals

# LLM-based evaluation
pe metrics --type g-eval --criteria "accuracy,clarity,completeness" --generated outputs.txt
```

### Pipeline Commands

#### `pe ask` - Single Query

Pipeline-friendly single query to LLM providers.

```bash
echo "What is AI?" | pe ask --provider openai:gpt-4
pe ask --prompt "Explain quantum computing" --provider anthropic:claude-3
```

#### `pe stream` - Stream Processing

Process evaluation results as a stream.

```bash
pe eval config.yaml | pe stream --select response,latency,cost
pe eval config.yaml | pe stream --format csv --output streaming-results.csv
```

#### `pe filter` - Result Filtering

Filter evaluation results based on conditions.

```bash
pe eval config.yaml | pe filter --success --min-score 0.8 --max-cost 0.05
pe stream results.json | pe filter --provider openai --contains "accurate"
```

#### `pe analyze` - Statistical Analysis

Analyze evaluation results with advanced statistics.

```bash
pe eval config.yaml | pe analyze --metric latency --distribution --outliers
pe filter results.json | pe analyze --metric accuracy --group-by provider --effect-size
```

#### `pe stats` - Quick Statistics

Show quick statistics from evaluation results.

```bash
pe eval config.yaml | pe stats
pe filter results.json | pe stats --format table --export-csv summary.csv
```

### Utility Commands

#### `pe view` - Interactive Results Viewer

View evaluation results in browser UI.

```bash
pe view                    # View latest results
pe view results.json       # View specific results file
pe view --port 8080        # Custom port
```

#### `pe benchmark` - Performance Benchmarking

Compare performance metrics of prompts and providers.

```bash
pe benchmark config.yaml --iterations 10 --concurrency 5 --format table
pe benchmark --providers openai,anthropic --prompts prompts/ --statistical-analysis
```

#### `pe test` - Advanced Testing

Run property-based and regression testing.

```bash
pe test property config.yaml --properties consistency,monotonicity
pe test regression baseline.json current.json --significance-level 0.05
pe test ab-test --group-a control.json --group-b treatment.json --bayesian
```

#### `pe profile` - Performance Profiling

Real-time observability and performance profiling.

```bash
pe profile optimize --method pe2 --cpu-profile --memory-profile
pe profile eval config.yaml --trace --export-metrics prometheus
```

#### `pe template` - Template Management

Manage prompt templates and libraries.

```bash
pe template list                           # List available templates
pe template search "sentiment analysis"    # Search templates
pe template apply template-name prompts/   # Apply template
pe template create custom-template.yaml    # Create new template
```

## 📝 Configuration File Reference

### Evaluation Configuration

```yaml
# Basic evaluation configuration
description: "Sentiment analysis evaluation"

# Prompts to evaluate
prompts:
  - "Analyze the sentiment: {{text}}"
  - "What is the emotional tone of: {{text}}"
  - file: "prompts/sentiment.txt"

# Test variables
vars:
  text:
    - "I love this product!"
    - "This is terrible."
    - "It's okay, nothing special."

# Or load from file
# vars:
#   file: "data/test-cases.csv"

# LLM providers
providers:
  - openai:gpt-4
  - openai:gpt-3.5-turbo
  - anthropic:claude-3-sonnet
  - google:gemini-pro
  - custom:
      name: "local-model"
      endpoint: "http://localhost:8000/v1"
      headers:
        Authorization: "Bearer {{CUSTOM_API_KEY}}"

# Evaluation assertions
assertions:
  - type: contains
    value: "positive"
    weight: 0.3
  - type: not-contains
    value: "negative"
    weight: 0.3
  - type: regex
    value: "\\b(good|great|excellent)\\b"
    weight: 0.2
  - type: length
    min: 10
    max: 200
    weight: 0.1
  - type: llm-judge
    criteria: "Accuracy of sentiment analysis"
    model: "gpt-4"
    weight: 0.1

# Advanced evaluation settings
evaluation:
  metrics:
    - bleu
    - rouge
    - bertscore
    - g-eval:
        criteria: ["accuracy", "clarity", "completeness"]
        model: "gpt-4"
  
  statistical:
    significance_test: true
    confidence_level: 0.95
    effect_size: true
    bootstrap_samples: 1000
  
  parallel:
    enabled: true
    max_workers: 10
    timeout: 30s
    retry_attempts: 3

# Output configuration
output:
  format: json
  file: "results/evaluation-{{timestamp}}.json"
  save_to_db: true
  stream: false

# Security testing
security:
  enabled: true
  categories:
    - prompt_injection
    - data_leakage
    - bias_detection
  severity: comprehensive
  
# Cost tracking
cost:
  budget_limit: 100.0
  alert_threshold: 80.0
  track_per_provider: true
```

### Optimization Configuration

```yaml
# Optimization configuration
optimization:
  method: pe2  # pe2, apex, textgrad, semantic, evolve, fusion
  
  # Method-specific configurations
  pe2:
    iterations: 5
    temperature: 0.7
    focus_areas: ["clarity", "specificity", "examples"]
  
  apex:
    beam_width: 5
    iterations: 8
    mutation_operators: ["rephrase", "expand", "prune", "reorder"]
    search_history: true
    length_optimization: true
  
  textgrad:
    iterations: 6
    attention_flow: true
    semantic_drift_detection: true
    gradient_accumulation: true
  
  semantic:
    method: backprop  # backprop, descent, gaso
    learning_rate: 0.1
    adaptive: true
    convergence_threshold: 0.001
  
  evolutionary:
    population_size: 20
    generations: 25
    mutation_rate: 0.1
    crossover_rate: 0.8
    selection_strategy: tournament
    multi_objective: true
    objectives: ["accuracy", "latency", "cost"]
  
  fusion:
    models: ["gpt-4", "claude-3", "gemini-pro"]
    consensus_strategy: reflection  # weighted, majority, reflection
    cross_validation: true
    robustness_testing: true

# Quality gates
quality:
  gates:
    - name: "coherence"
      threshold: 0.8
      required: true
    - name: "clarity"
      threshold: 0.75
      required: true
  
  validation:
    statistical: true
    cross_validation_folds: 5
    significance_level: 0.05

# Component composition
composition:
  style: cot  # default, cot, few-shot, structured, conversational, dspy
  coherence_analysis: true
  validation_gates: true
  type_safety: true
  dependency_resolution: true
```

### Security Configuration

```yaml
# Security testing configuration
security:
  # OWASP LLM Top 10 categories
  owasp:
    prompt_injection:
      enabled: true
      test_vectors: "vectors/prompt-injection.yaml"
      severity: comprehensive
    
    insecure_output_handling:
      enabled: true
      validation_rules: "rules/output-validation.yaml"
    
    training_poisoning:
      enabled: true
      detection_methods: ["statistical", "semantic"]
    
    model_denial_of_service:
      enabled: true
      resource_limits:
        max_tokens: 4000
        timeout: 30s
    
    supply_chain_vulnerabilities:
      enabled: true
      dependency_scanning: true
    
    sensitive_information_disclosure:
      enabled: true
      pii_detection: true
      training_data_extraction: true
    
    insecure_plugin_design:
      enabled: true
      plugin_validation: true
    
    excessive_agency:
      enabled: true
      permission_boundaries: true
    
    overreliance:
      enabled: true
      confidence_calibration: true
    
    model_theft:
      enabled: true
      extraction_detection: true

  # Custom security tests
  custom_tests:
    - name: "domain_specific_injection"
      file: "tests/domain-injection.yaml"
    - name: "bias_detection"
      criteria: ["gender", "race", "age", "religion"]

  # Red team configuration
  red_team:
    enabled: true
    intensity: comprehensive  # basic, standard, comprehensive
    duration: 24h
    attack_strategies:
      - jailbreaking
      - prompt_injection
      - social_engineering
      - adversarial_examples

  # Monitoring and alerting
  monitoring:
    realtime: true
    alert_levels: ["medium", "high", "critical"]
    webhooks:
      - url: "https://hooks.slack.com/security-alerts"
        events: ["high", "critical"]
    
  # Compliance frameworks
  compliance:
    frameworks: ["owasp", "nist", "iso27001"]
    reporting:
      format: pdf
      schedule: weekly
      recipients: ["security@company.com"]
```

## 🔧 Go API Reference

### Core Types

```go
package metaprompt

import (
    "context"
    "time"
    "github.com/tmc/pe/internal/llm"
)

// OptimizationResult contains the results of prompt optimization
type OptimizationResult struct {
    OriginalPrompt    string             `json:"original_prompt"`
    OptimizedPrompt   string             `json:"optimized_prompt"`
    Iterations        []IterationResult  `json:"iterations"`
    ImprovementScore  float64            `json:"improvement_score"`
    TotalDuration     time.Duration      `json:"total_duration"`
    CreatedAt         time.Time          `json:"created_at"`
    Method            string             `json:"method"`
    Metadata          map[string]interface{} `json:"metadata"`
}

// IterationResult contains the result of a single optimization iteration
type IterationResult struct {
    Iteration     int           `json:"iteration"`
    Prompt        string        `json:"prompt"`
    Score         float64       `json:"score"`
    Feedback      string        `json:"feedback"`
    Suggestions   []string      `json:"suggestions"`
    Duration      time.Duration `json:"duration"`
    Metadata      map[string]interface{} `json:"metadata"`
}

// Config contains configuration for prompt optimization
type Config struct {
    InitialPrompt string
    Iterations    int
    Temperature   float64
    MaxTokens     int
    Method        string // "pe2", "apex", "textgrad", "semantic", "evolve", "fusion"
    Provider      string
    Metadata      map[string]interface{}
}
```

### Optimizer Interface

```go
// Optimizer implements prompt optimization using various methods
type Optimizer struct {
    llm      llm.Provider
    textGrad *TextGradOptimizer
    pe2      *PE2Optimizer
    apex     *APEXOptimizer
    semantic *SemanticOptimizer
    evolutionary *EvolutionaryOptimizer
    fusion   *FusionOptimizer
}

// NewOptimizer creates a new prompt optimizer
func NewOptimizer(llmProvider llm.Provider) *Optimizer

// Optimize runs the prompt optimization process
func (o *Optimizer) Optimize(ctx context.Context, cfg Config) (*OptimizationResult, error)

// OptimizeWithMethod optimizes using a specific method
func (o *Optimizer) OptimizeWithMethod(ctx context.Context, method string, cfg Config) (*OptimizationResult, error)

// GetSupportedMethods returns list of supported optimization methods
func (o *Optimizer) GetSupportedMethods() []string
```

### PE2 Optimizer

```go
// PE2Optimizer implements PE2 (Prompt Engineering a Prompt Engineer)
type PE2Optimizer struct {
    llm llm.Provider
}

// NewPE2Optimizer creates a new PE2 optimizer
func NewPE2Optimizer(llmProvider llm.Provider) *PE2Optimizer

// OptimizeWithPE2 runs PE2-style optimization
func (o *PE2Optimizer) OptimizeWithPE2(ctx context.Context, cfg Config) (*OptimizationResult, error)

// PE2Config contains PE2-specific configuration
type PE2Config struct {
    FocusAreas    []string  // Areas to focus optimization on
    ExpertPersona string    // Expert persona to use
    Reasoning     bool      // Enable reasoning templates
    Examples      bool      // Include examples in optimization
    Constraints   []string  // Optimization constraints
}
```

### APEX Optimizer

```go
// APEXOptimizer implements APEX (Automated Prompt Engineering Xpert)
type APEXOptimizer struct {
    llm llm.Provider
}

// NewAPEXOptimizer creates a new APEX optimizer
func NewAPEXOptimizer(llmProvider llm.Provider) *APEXOptimizer

// OptimizeWithAPEX runs APEX-style long prompt optimization
func (o *APEXOptimizer) OptimizeWithAPEX(ctx context.Context, cfg Config) (*OptimizationResult, error)

// APEXConfig contains APEX-specific configuration
type APEXConfig struct {
    BeamWidth               int      // Beam search width
    MutationOperators       []string // Operators for prompt mutation
    UseSearchHistory        bool     // Use search history for mutations
    LengthOptimization      bool     // Optimize for prompt length
    MaxPromptLength         int      // Maximum optimized prompt length
    MinImprovementThreshold float64  // Minimum improvement to continue
}
```

### TextGrad Optimizer

```go
// TextGradOptimizer implements TextGrad 2.0 with attention flow
type TextGradOptimizer struct {
    llm llm.Provider
}

// NewTextGradOptimizer creates a new TextGrad optimizer
func NewTextGradOptimizer(llmProvider llm.Provider) *TextGradOptimizer

// OptimizeWithTextGrad runs TextGrad-style optimization
func (o *TextGradOptimizer) OptimizeWithTextGrad(ctx context.Context, cfg Config) (*OptimizationResult, error)

// TextGradConfig contains TextGrad-specific configuration
type TextGradConfig struct {
    AttentionFlow           bool    // Enable attention flow mapping
    SemanticDriftDetection  bool    // Monitor semantic drift
    GradientAccumulation    bool    // Use gradient accumulation
    BackwardPropagation     bool    // Enable backward propagation
    CrossModalGradients     bool    // Support cross-modal gradients
}
```

### Semantic Optimizer (2025 Research)

```go
// SemanticOptimizer implements semantic backpropagation and GASO
type SemanticOptimizer struct {
    llm llm.Provider
}

// NewSemanticOptimizer creates a new semantic optimizer
func NewSemanticOptimizer(llmProvider llm.Provider) *SemanticOptimizer

// SemanticBackprop performs semantic backpropagation
func (o *SemanticOptimizer) SemanticBackprop(ctx context.Context, cfg SemanticConfig) (*OptimizationResult, error)

// SemanticGradientDescent performs semantic gradient descent
func (o *SemanticOptimizer) SemanticGradientDescent(ctx context.Context, cfg SemanticConfig) (*OptimizationResult, error)

// GASO performs Graph-based Agentic System Optimization
func (o *SemanticOptimizer) GASO(ctx context.Context, cfg GASOConfig) (*OptimizationResult, error)

// SemanticConfig contains semantic optimization configuration
type SemanticConfig struct {
    Objective           string    // Optimization objective
    LearningRate        float64   // Learning rate for gradient descent
    Adaptive            bool      // Adaptive learning rate
    ConvergenceThreshold float64  // Convergence threshold
    GradientStrength    bool      // Analyze gradient strength
    MultiObjective      bool      // Multi-objective optimization
}

// GASOConfig contains GASO-specific configuration
type GASOConfig struct {
    SystemDefinition string                 // System definition file
    Objectives       []string               // Optimization objectives
    ParetoAnalysis   bool                  // Pareto frontier analysis
    DependencyGraph  map[string][]string   // Component dependencies
    QualityGates     []QualityGate         // Quality gates for optimization
}
```

### Evaluation API

```go
package evaluator

// EvaluationResult contains evaluation results
type EvaluationResult struct {
    Provider    string                 `json:"provider"`
    Prompt      string                 `json:"prompt"`
    Response    string                 `json:"response"`
    Score       float64                `json:"score"`
    Latency     time.Duration          `json:"latency"`
    Cost        float64                `json:"cost"`
    Assertions  []AssertionResult      `json:"assertions"`
    Metrics     map[string]float64     `json:"metrics"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Evaluator interface for prompt evaluation
type Evaluator interface {
    Evaluate(ctx context.Context, config EvaluationConfig) ([]EvaluationResult, error)
    EvaluateStream(ctx context.Context, config EvaluationConfig) (<-chan EvaluationResult, error)
    GetSupportedMetrics() []string
    GetSupportedProviders() []string
}

// NewEvaluator creates a new evaluator
func NewEvaluator() Evaluator

// EvaluationConfig contains evaluation configuration
type EvaluationConfig struct {
    Prompts     []string               `json:"prompts"`
    Providers   []string               `json:"providers"`
    Variables   map[string]interface{} `json:"variables"`
    Assertions  []Assertion            `json:"assertions"`
    Metrics     []string               `json:"metrics"`
    Parallel    bool                   `json:"parallel"`
    MaxWorkers  int                    `json:"max_workers"`
    Timeout     time.Duration          `json:"timeout"`
    SaveResults bool                   `json:"save_results"`
}
```

### Security API

```go
package security

// SecurityTester implements comprehensive security testing
type SecurityTester struct {
    llm llm.Provider
}

// NewSecurityTester creates a new security tester
func NewSecurityTester(llmProvider llm.Provider) *SecurityTester

// TestOWASP runs OWASP LLM Top 10 security tests
func (st *SecurityTester) TestOWASP(ctx context.Context, config OWASPConfig) (*SecurityReport, error)

// TestCustom runs custom security tests
func (st *SecurityTester) TestCustom(ctx context.Context, config CustomSecurityConfig) (*SecurityReport, error)

// Monitor provides real-time security monitoring
func (st *SecurityTester) Monitor(ctx context.Context, config MonitorConfig) (<-chan SecurityAlert, error)

// SecurityReport contains security test results
type SecurityReport struct {
    Target          string                    `json:"target"`
    Timestamp       time.Time                 `json:"timestamp"`
    OverallScore    float64                   `json:"overall_score"`
    Categories      map[string]CategoryResult `json:"categories"`
    Vulnerabilities []Vulnerability           `json:"vulnerabilities"`
    Recommendations []string                  `json:"recommendations"`
    Compliance      map[string]bool          `json:"compliance"`
}

// CategoryResult contains results for a security category
type CategoryResult struct {
    Name        string        `json:"name"`
    Score       float64       `json:"score"`
    Passed      bool          `json:"passed"`
    Tests       []TestResult  `json:"tests"`
    Severity    string        `json:"severity"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
    ID          string    `json:"id"`
    Category    string    `json:"category"`
    Severity    string    `json:"severity"`
    Description string    `json:"description"`
    Evidence    string    `json:"evidence"`
    Mitigation  string    `json:"mitigation"`
    CVSS        float64   `json:"cvss"`
}
```

## 🌐 REST API Reference

PE provides a REST API for integration with web applications and services.

### Authentication

```bash
# API key authentication
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.pe.dev/v1/optimize
```

### Endpoints

#### POST /v1/optimize

Optimize a prompt using specified method.

**Request:**
```json
{
  "prompt": "Analyze sentiment",
  "method": "pe2",
  "iterations": 5,
  "provider": "openai:gpt-4",
  "config": {
    "temperature": 0.7,
    "max_tokens": 2000
  }
}
```

**Response:**
```json
{
  "id": "opt_123456",
  "status": "completed",
  "original_prompt": "Analyze sentiment",
  "optimized_prompt": "You are an expert sentiment analysis specialist...",
  "improvement_score": 0.85,
  "iterations": [...],
  "metadata": {...}
}
```

#### POST /v1/evaluate

Evaluate prompts against providers.

**Request:**
```json
{
  "prompts": ["Analyze sentiment: {{text}}"],
  "providers": ["openai:gpt-4", "anthropic:claude-3"],
  "variables": {
    "text": ["I love this!", "This is bad."]
  },
  "metrics": ["bleu", "rouge", "bertscore"],
  "assertions": [
    {
      "type": "contains",
      "value": "positive"
    }
  ]
}
```

**Response:**
```json
{
  "id": "eval_789012",
  "status": "completed",
  "results": [
    {
      "provider": "openai:gpt-4",
      "prompt": "Analyze sentiment: I love this!",
      "response": "This text expresses positive sentiment...",
      "score": 0.92,
      "latency": 1.2,
      "cost": 0.003,
      "metrics": {
        "bleu": 0.87,
        "rouge": 0.89,
        "bertscore": 0.91
      }
    }
  ],
  "summary": {...}
}
```

#### POST /v1/security/test

Run security tests on prompts.

**Request:**
```json
{
  "target": "You are a helpful assistant...",
  "categories": ["prompt_injection", "data_leakage"],
  "severity": "comprehensive",
  "compliance": ["owasp"]
}
```

**Response:**
```json
{
  "id": "sec_345678",
  "status": "completed",
  "overall_score": 0.78,
  "categories": {
    "prompt_injection": {
      "score": 0.75,
      "passed": true,
      "tests": [...]
    }
  },
  "vulnerabilities": [...],
  "compliance": {
    "owasp": true
  }
}
```

#### GET /v1/jobs/{id}

Get status of a job (optimization, evaluation, or security test).

**Response:**
```json
{
  "id": "opt_123456",
  "type": "optimization",
  "status": "running",
  "progress": 60,
  "estimated_completion": "2024-01-15T10:30:00Z",
  "result": null
}
```

#### GET /v1/metrics

Get available evaluation metrics.

**Response:**
```json
{
  "metrics": [
    {
      "name": "bleu",
      "description": "BLEU score for text generation quality",
      "type": "reference_based",
      "parameters": ["n_gram", "smoothing"]
    },
    {
      "name": "g-eval",
      "description": "LLM-based evaluation with custom criteria",
      "type": "llm_based",
      "parameters": ["criteria", "model"]
    }
  ]
}
```

#### GET /v1/providers

Get available LLM providers.

**Response:**
```json
{
  "providers": [
    {
      "name": "openai",
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "capabilities": ["text", "chat", "completion"]
    },
    {
      "name": "anthropic",
      "models": ["claude-3-sonnet", "claude-3-haiku"],
      "capabilities": ["text", "chat"]
    }
  ]
}
```

## 🔄 Pipeline Processing

PE's revolutionary Unix pipeline processing enables complex workflows.

### Basic Pipelines

```bash
# Evaluation → Filtering → Analysis
pe eval config.yaml | pe filter --success | pe analyze --metric accuracy

# Streaming evaluation with real-time filtering
pe eval config.yaml --stream | pe filter --min-score 0.8 | pe stats

# Multi-stage optimization pipeline
pe compose context.txt instruction.txt | pe optimize --method pe2 | pe fusion --models gpt-4,claude-3
```

### Advanced Pipeline Patterns

```bash
# Quality monitoring pipeline
pe eval production.yaml --stream | \
  pe filter --success --min-score 0.8 | \
  pe analyze --metric quality-degradation | \
  pe alert --webhook slack://quality-alerts

# Cost optimization pipeline
pe eval config.yaml --stream | \
  pe filter --max-cost 0.05 --min-quality 0.75 | \
  pe analyze --metric cost-effectiveness | \
  pe stats --export-csv cost-analysis.csv

# Research pipeline with full traceability
pe compose --research-mode components/ | \
  pe evolve --generations 15 --trace-genealogy | \
  pe fusion --consensus-analysis --cross-validation | \
  pe test property --comprehensive --significance-testing
```

### Pipeline Data Format

PE pipelines use JSON streaming format:

```json
{"type": "evaluation_result", "provider": "openai:gpt-4", "score": 0.92, "latency": 1.2}
{"type": "evaluation_result", "provider": "anthropic:claude-3", "score": 0.89, "latency": 0.8}
{"type": "summary", "total_evaluations": 100, "average_score": 0.87}
```

## 🚀 Advanced Usage Patterns

### Production Monitoring

```bash
# Real-time quality monitoring
pe eval production-config.yaml --stream --realtime | \
  pe filter --quality-threshold 0.8 | \
  pe alert --critical-threshold 0.6 --webhook production-alerts

# Cost tracking and optimization
pe monitor costs --providers all --budget-limit 1000 --alerts high-usage

# Performance profiling
pe profile eval config.yaml --cpu-profile --memory-profile --export-metrics prometheus
```

### Research Workflows

```bash
# Academic research pipeline
pe compose --research-mode --experiment-id exp001 | \
  pe evolve --trace-genealogy --statistical-analysis | \
  pe fusion --consensus-analysis --reproducibility-package | \
  pe publish --format paper --venue neurips

# Cross-validation experiments
pe test cross-validate --methods pe2,apex,textgrad --folds 5 --statistical-significance

# Meta-analysis across optimization sessions
pe reflect --session-data sessions/ --extract-patterns --knowledge-base
```

### Enterprise Integration

```bash
# CI/CD integration with quality gates
pe optimize prompts/ --method pe2 --quality-gates | \
  pe test comprehensive --statistical-validation | \
  pe security test --owasp-complete | \
  pe deploy --production-ready

# A/B testing pipeline
pe test ab-test --baseline current.json --treatment optimized.json --bayesian | \
  pe analyze --effect-size --confidence-intervals | \
  pe decision --significance-threshold 0.05
```

---

This comprehensive API reference covers all aspects of PE's functionality. For additional examples and tutorials, see the [examples/](../example/) directory and [documentation](.) folder.

**PE: The definitive API for the future of prompt engineering.**