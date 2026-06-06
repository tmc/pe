<!-- Historical draft: archived planning material, not current product documentation. Claims, metrics, and command examples in this file may be stale or aspirational. -->

# Command Implementation Plan

This document provides detailed technical implementation plans for the planned commands, including architecture decisions, interfaces, and development guidelines.

## Architecture Considerations

### Command Structure Pattern
All new commands should follow the established pattern:

```go
// cmd/pe/newcommand.go
package main

import (
    "github.com/spf13/cobra"
    "github.com/tmc/pe/internal/promptfoo/..."
)

func init() {
    rootCmd.AddCommand(newCommandCmd)
}

var newCommandCmd = &cobra.Command{
    Use:   "newcommand",
    Short: "Brief description",
    Long:  "Detailed description with examples",
    RunE:  runNewCommand,
}

func runNewCommand(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

### Shared Infrastructure Needs

#### 1. Enhanced File Processing Pipeline
Location: `internal/promptfoo/processing/`

```go
package processing

type Pipeline struct {
    Stages []Stage
    Config PipelineConfig
}

type Stage interface {
    Process(ctx context.Context, input Input) (Output, error)
    Name() string
}

// Stages for transformation pipeline
type TransformStage struct {
    Transform TransformFunc
}

type ValidateStage struct {
    Rules []ValidationRule
}

type OutputStage struct {
    Format OutputFormat
    Writer io.Writer
}
```

#### 2. Prompt Comparison Engine
Location: `internal/promptfoo/comparison/`

```go
package comparison

type ComparisonEngine struct {
    SemanticModel EmbeddingModel
    DiffAlgorithm DiffAlgorithm
}

type ComparisonResult struct {
    Similarity    float64
    Differences   []Difference
    Semantic      SemanticAnalysis
    Structural    StructuralDiff
}

type DiffAlgorithm interface {
    Compare(a, b Prompt) ComparisonResult
}
```

#### 3. Workflow Engine
Location: `internal/promptfoo/workflow/`

```go
package workflow

type Workflow struct {
    Name    string
    Steps   []Step
    Config  WorkflowConfig
}

type Step struct {
    Name         string
    Command      string
    Args         []string
    Dependencies []string
    Condition    Condition
    Retry        RetryConfig
}

type Executor struct {
    MaxWorkers int
    Scheduler  Scheduler
}
```

## Phase 1 Implementation Details

### 1. `pe diff` Command

**Files to Create:**
- `cmd/pe/diff.go` - Main command implementation
- `internal/promptfoo/comparison/differ.go` - Core diffing logic
- `internal/promptfoo/comparison/semantic.go` - Semantic comparison
- `internal/promptfoo/comparison/types.go` - Type definitions

**Key Features:**
```go
type DiffOptions struct {
    Semantic      bool
    Context       int
    Format        string // unified, side-by-side, json
    IgnoreSpacing bool
    CustomRules   []DiffRule
}

type DiffResult struct {
    Hunks       []DiffHunk
    Similarity  float64
    Summary     DiffSummary
}
```

**Implementation Strategy:**
1. Start with text-based diffing using existing Go libraries
2. Add semantic comparison using embedding models
3. Support multiple output formats
4. Add custom diff rules for prompt-specific comparisons

**Testing:**
- Unit tests for different diff algorithms
- Integration tests with various prompt formats
- Performance tests for large prompts

### 2. `pe lint` Command

**Files to Create:**
- `cmd/pe/lint.go` - Main command
- `internal/promptfoo/linting/rules.go` - Linting rules
- `internal/promptfoo/linting/engine.go` - Linting engine
- `internal/promptfoo/linting/fixes.go` - Auto-fix implementations

**Rule Categories:**
```go
type RuleCategory string

const (
    RuleSecurity     RuleCategory = "security"
    RulePerformance  RuleCategory = "performance"
    RuleStyle        RuleCategory = "style"
    RuleCompatibility RuleCategory = "compatibility"
)

type Rule interface {
    Check(prompt Prompt) []Issue
    Category() RuleCategory
    Severity() Severity
    AutoFix() bool
}
```

**Built-in Rules:**
- Variable injection vulnerabilities
- Excessive token usage patterns
- Inconsistent formatting
- Missing error handling
- Provider compatibility issues

### 3. `pe batch` Command

**Files to Create:**
- `cmd/pe/batch.go` - Main command
- `internal/promptfoo/batch/executor.go` - Batch execution engine
- `internal/promptfoo/batch/scheduler.go` - Job scheduling
- `internal/promptfoo/batch/results.go` - Result aggregation

**Batch Processing Architecture:**
```go
type BatchJob struct {
    ID       string
    Prompts  []PromptInput
    Config   BatchConfig
    Status   JobStatus
}

type BatchConfig struct {
    MaxWorkers    int
    Timeout       time.Duration
    RetryPolicy   RetryPolicy
    FailureMode   FailureMode // continue, stop, retry
    OutputFormat  string
}

type BatchExecutor struct {
    WorkerPool chan *Worker
    JobQueue   chan BatchJob
    Results    chan BatchResult
}
```

**Features:**
- Parallel execution with configurable worker pools
- Progress tracking and cancellation
- Multiple output formats (JSON, CSV, YAML)
- Error handling and retry logic
- Resource usage monitoring

### 4. `pe watch` Command

**Files to Create:**
- `cmd/pe/watch.go` - Main command
- `internal/promptfoo/watch/watcher.go` - File system watcher
- `internal/promptfoo/watch/debouncer.go` - Event debouncing
- `internal/promptfoo/watch/executor.go` - Command execution

**Watch Implementation:**
```go
type Watcher struct {
    Patterns    []string
    Excludes    []string
    Debounce    time.Duration
    Commands    []Command
    Hooks       []Hook
}

type WatchEvent struct {
    Type     EventType // create, modify, delete
    Path     string
    Time     time.Time
}

type Command struct {
    Cmd     string
    Args    []string
    WorkDir string
    Env     []string
}
```

**Features:**
- File system event monitoring
- Pattern-based filtering
- Debounced execution
- Command chaining
- Error handling and logging

## Phase 2 Implementation Details

### 1. `pe transform` Command

**Transformation Pipeline:**
```go
type Transformer interface {
    Transform(ctx context.Context, prompt Prompt) (Prompt, error)
    Name() string
    Config() TransformConfig
}

// Specific transformers
type StyleTransformer struct {
    TargetStyle Style
    Rules       []StyleRule
}

type FormatTransformer struct {
    SourceFormat Format
    TargetFormat Format
    Options      FormatOptions
}

type LanguageTransformer struct {
    SourceLang string
    TargetLang string
    Model      TranslationModel
}
```

**Transform Types:**
- Format conversion (text ↔ YAML ↔ JSON)
- Style transformation (formal, casual, technical)
- Language translation
- Compression/expansion
- Template variable extraction/substitution

### 2. `pe import/export` Commands

**Import Sources:**
```go
type ImportSource interface {
    Import(ctx context.Context, source string) ([]Prompt, error)
    Supports(source string) bool
}

// Implementations
type LangChainImporter struct{}
type OpenAIImporter struct{}
type HuggingFaceImporter struct{}
type PromptFlowImporter struct{}
```

**Export Targets:**
```go
type ExportTarget interface {
    Export(ctx context.Context, prompts []Prompt, dest string) error
    Supports(dest string) bool
}

// Implementations
type APIExporter struct{}
type KubernetesExporter struct{}
type DockerExporter struct{}
type CloudFunctionExporter struct{}
```

### 3. `pe sweep` Command

**Parameter Sweeping:**
```go
type ParameterSweep struct {
    Parameters []Parameter
    Strategy   SweepStrategy // grid, random, bayesian
    Constraints []Constraint
    Objective  Objective
}

type Parameter struct {
    Name   string
    Type   ParameterType // float, int, string, enum
    Range  Range
    Step   interface{}
}

type SweepResult struct {
    Combinations []ParameterCombination
    Results     []EvaluationResult
    BestParams  ParameterCombination
    Analysis    SweepAnalysis
}
```

## Phase 3 Implementation Details

### 1. `pe workflow` Command

**Workflow Definition:**
```yaml
name: "complex-evaluation"
version: "1.0"

variables:
  dataset: "data/test-cases.json"
  models: ["gpt-4", "claude-3", "gemini"]

steps:
  - name: "prepare-data"
    command: "pe import --format json ${dataset}"
    outputs: ["prepared-data.yaml"]
    
  - name: "run-evaluations"
    command: "pe batch prepared-data.yaml --providers ${models}"
    depends: ["prepare-data"]
    parallel: true
    outputs: ["eval-results.json"]
    
  - name: "analyze-results"
    command: "pe analyze --performance eval-results.json"
    depends: ["run-evaluations"]
    
  - name: "generate-report"
    command: "pe report --template summary eval-results.json"
    depends: ["analyze-results"]
```

**Workflow Engine:**
```go
type WorkflowEngine struct {
    Executor     Executor
    State        StateManager
    EventBus     EventBus
    Scheduler    StepScheduler
}

type StateManager interface {
    SaveState(workflow string, state WorkflowState) error
    LoadState(workflow string) (WorkflowState, error)
    Checkpoint(workflow string, step string) error
}
```

### 2. `pe trace` Command

**Execution Tracing:**
```go
type Tracer struct {
    Enabled     bool
    Output      io.Writer
    Granularity TracingLevel
    Filters     []TraceFilter
}

type TraceEvent struct {
    Timestamp   time.Time
    Type        EventType
    Component   string
    Data        map[string]interface{}
    Duration    time.Duration
    Parent      *TraceEvent
}

type TracingLevel int

const (
    TraceLevelBasic TracingLevel = iota
    TraceLevelDetailed
    TraceLevelVerbose
)
```

## Integration Points

### 1. Module System Integration
All new commands should integrate with the existing module system:

```go
// Use module-aware file resolution
resolver := module.NewResolver(moduleConfig)
prompts, err := resolver.ResolvePrompts(patterns)

// Support module dependencies
deps, err := module.LoadDependencies(workDir)
```

### 2. Provider System Integration
Commands should work with all supported providers:

```go
// Provider-agnostic execution
provider, err := providers.GetProvider(providerName)
result, err := provider.Execute(ctx, prompt, options)
```

### 3. Cache Integration
Leverage existing cache system:

```go
// Use content-addressed caching
cache := cache.NewContentAddressedCache()
if cached, ok := cache.Get(promptHash); ok {
    return cached
}
```

### 4. Security Integration
Follow security best practices:

```go
// Use existing attestation system
attestor := attestation.New(keychain)
signature, err := attestor.Sign(promptContent)

// Input validation
if err := validation.ValidatePrompt(prompt); err != nil {
    return fmt.Errorf("invalid prompt: %w", err)
}
```

## Testing Strategy

### Unit Testing
- Each command should have comprehensive unit tests
- Mock external dependencies (providers, file system)
- Test error conditions and edge cases

### Integration Testing
- Test command interactions with existing systems
- Verify module system integration
- Test with real providers (when appropriate)

### Performance Testing
- Benchmark batch operations
- Test with large prompt sets
- Memory usage profiling

### End-to-End Testing
- Full workflow testing
- Multi-command pipeline testing
- Real-world scenario simulation

## Documentation Requirements

For each new command:
1. Add to `docs/CLI_REFERENCE.md` with examples
2. Create man page style documentation
3. Add to README.md command list
4. Include in CLAUDE.md development guidance
5. Create example files in `example/` directory

## Rollout Strategy

### Alpha Release
- Implement Phase 1 commands
- Limited testing with core users
- Documentation and examples

### Beta Release  
- Add Phase 2 commands
- Community feedback integration
- Performance optimization

### Stable Release
- Complete Phase 3 commands
- Full test coverage
- Production deployment guides

This implementation plan ensures systematic development while maintaining code quality and architectural consistency.
