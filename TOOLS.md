# PE Toolkit Tools

This document describes the comprehensive set of tools available in the PE (Prompt Engineering) toolkit, organized by category and use case.

## Core Evaluation Tools

### `pe eval`
Evaluate prompt configurations against LLM providers with comprehensive metrics and analysis.

```bash
# Basic evaluation
pe eval config.yaml

# With output file
pe eval config.yaml -o results.json

# With specific provider
pe eval config.yaml --provider openai:gpt-4
```

### `pe view`
Launch browser-based UI for viewing and analyzing evaluation results interactively.

```bash
pe view                    # View latest results
pe view results.json       # View specific results file
```

### `pe vet`
Validate promptfoo configuration files for syntax and semantic correctness.

```bash
pe vet config.yaml         # Validate single file
pe vet configs/*.yaml      # Validate multiple files
```

## Configuration Management

### `pe fmt`
Format promptfoo configuration files with consistent styling and structure.

```bash
pe fmt config.yaml                    # Format to stdout
pe fmt config.yaml --output yaml      # Specify output format
pe fmt config.yaml --write           # Write in-place
```

### `pe convert`
Convert promptfoo configuration files between different formats (YAML, JSON).

```bash
pe convert config.yaml config.json --output json
pe convert config.json config.yaml --output yaml
```

## Performance and Analysis

### `pe benchmark`
Compare performance metrics of prompts and providers with detailed statistical analysis.

```bash
# Basic benchmarking
pe benchmark benchmark-config.yaml

# With custom iterations and concurrency
pe benchmark config.yaml --iterations 5 --concurrency 2

# Different output formats
pe benchmark config.yaml --format text
pe benchmark config.yaml --format json
pe benchmark config.yaml --format csv
```

Key features:
- Latency percentiles (p50, p95, p99)
- Token usage analysis
- Cost calculations
- Error rate tracking
- Statistical significance testing

### `pe analyze`
Analyze evaluation results with advanced statistics and insights.

```bash
# Basic analysis
pe eval config.yaml | pe analyze

# Focus on specific metrics
pe eval config.yaml | pe analyze --metric latency
pe eval config.yaml | pe analyze --metric accuracy

# Generate detailed reports
pe analyze results.json --report --output analysis.html
```

### `pe stats`
Show quick statistics from evaluation results for rapid insights.

```bash
pe eval config.yaml | pe stats
pe stats results.json
```

## Pipeline-Friendly Commands

### `pe ask`
Ask a single question to an LLM provider (optimized for Unix pipelines).

```bash
# Direct usage
echo "What is AI?" | pe ask --provider openai:gpt-4

# With specific parameters
pe ask --prompt "Explain quantum computing" --provider anthropic:claude-3-sonnet

# In pipelines
cat questions.txt | pe ask --provider openai:gpt-4 > answers.txt
```

### `pe stream`
Process evaluation results as a stream for real-time analysis.

```bash
# Stream specific fields
pe eval config.yaml | pe stream --select response,latency

# Stream with filters
pe eval config.yaml | pe stream --select response --format json

# Continuous monitoring
pe eval config.yaml | pe stream --watch --interval 5s
```

### `pe filter`
Filter evaluation results based on conditions and criteria.

```bash
# Filter successful results
pe eval config.yaml | pe filter --success

# Filter by score
pe eval config.yaml | pe filter --score ">0.8"

# Complex filters
pe eval config.yaml | pe filter --latency "<500ms" --success
```

### `pe diff`
Compare two evaluation results to identify improvements or regressions.

```bash
pe diff baseline.json current.json
pe diff baseline.json current.json --format table
pe diff baseline.json current.json --metric accuracy
```

## Advanced Testing

### `pe test`
Run advanced testing including property-based and regression testing.

```bash
# Property-based testing
pe test property --config property-tests.yaml

# Regression testing
pe test regression --baseline baseline.json --current current.json

# A/B testing
pe test ab --config ab-test.yaml --metrics accuracy,latency
```

Features:
- Property-based testing with random input generation
- Regression detection with statistical significance
- A/B testing with confidence intervals
- Custom test property definitions

## Prompt Optimization & Metaprompting ⭐ NEW

### `pe optimize` 
Optimize prompts using advanced metaprompting techniques based on 2024 research.

```bash
# Basic optimization
pe optimize --prompt "Summarize this text" --iterations 3

# TextGrad optimization
pe optimize --prompt "Classify sentiment" --method textgrad --iterations 5

# Hybrid approach (standard + TextGrad)
pe optimize --prompt "Generate code" --method hybrid --iterations 6

# Multi-stage optimization
pe optimize --prompt "Complex task" --method multistage

# Save optimization results
pe optimize --prompt "Analyze data" --provider anthropic --output optimized.json
```

**Optimization Methods**:
- **standard**: Traditional iterative refinement with LLM feedback
- **textgrad**: TextGrad-style optimization using natural language gradients  
- **hybrid**: Combines standard and TextGrad approaches
- **multistage**: Sequential optimization through Analysis → Refinement → Validation → Polishing

### `pe analyze`
Perform TextGrad-style analysis of prompt-response gradients.

```bash
# Analyze prompt effectiveness
pe analyze --prompt "Your prompt" --response "Model response"

# Detailed gradient analysis
pe analyze --prompt "Test prompt" --response "Test response" --detailed

# Output analysis to file
pe analyze --prompt "Prompt" --response "Response" --output analysis.json
```

**Analysis Features**:
- **Attention Flow Mapping**: Visualizes token relationships
- **Semantic Drift Detection**: Identifies concept preservation issues
- **Coherence Scoring**: Evaluates response quality metrics  
- **Optimization Hints**: Generates specific improvement suggestions

### `pe refine`
Error-driven prompt refinement based on failure analysis.

```bash
# Refine problematic prompt
pe refine --prompt "Problematic prompt" --errors "error1,error2,error3"

# Automated error detection and fixing
pe refine --prompt "Your prompt" --auto-detect

# Generate regression tests
pe refine --prompt "Prompt" --generate-tests --output refined.json
```

**Refinement Features**:
- **Error Pattern Detection**: Identifies semantic, format, logic issues
- **Automated Fix Generation**: Context-aware repair suggestions
- **Regression Testing**: Comprehensive test suite generation
- **Quality Metrics**: Tracks error reduction and robustness improvements

### `pe gradients`
Compute optimization trajectories using gradient-based methods.

```bash
# Compute gradients for optimization
pe gradients --prompt "Current prompt" --objective "Desired outcome"

# Include optimization history
pe gradients --prompt "Prompt" --objective "Goal" --history history.json

# Get trajectory recommendations
pe gradients --prompt "Prompt" --objective "Goal" --recommendations
```

**Gradient Features**:
- **Loss Function Calculation**: Semantic, structural, task-specific loss
- **Step Size Optimization**: Controlled gradient descent
- **Convergence Tracking**: Real-time optimization progress
- **Trajectory Planning**: Multi-step optimization path prediction

### `pe reflect`
Meta-analysis of prompt engineering sessions for learning extraction.

```bash
# Reflect on optimization sessions
pe reflect --session-data sessions.json

# Generate strategy recommendations  
pe reflect --session-data sessions.json --strategies

# Extract knowledge base
pe reflect --session-data sessions.json --knowledge --output kb.json
```

**Reflection Features**:
- **Success Pattern Mining**: Identifies effective techniques
- **Strategy Recommendations**: Guides tool selection
- **Knowledge Distillation**: Extracts reusable principles
- **Workflow Optimization**: Suggests process improvements

## Advanced Metaprompting Techniques

### TextGrad Integration
Implements the latest 2024 TextGrad research:

```bash
# Pure TextGrad optimization
pe optimize --method textgrad --iterations 5

# Analyze textual gradients
pe analyze --prompt "..." --response "..." --gradients
```

Key innovations:
- Natural language gradients for optimization
- Backward propagation through text
- Iterative refinement with LLM feedback
- Attention pattern analysis

### DSPy-Style Optimization
Structured prompt generation with feedback loops:

```bash
# DSPy-inspired optimization
pe optimize --method hybrid --structured

# Multi-stage with validation gates
pe optimize --method multistage --gates
```

Features:
- Structured prompt generation
- Automatic optimization algorithms  
- Multi-stage refinement processes
- Cross-validation between methods

### Error-Driven Refinement
Systematic failure mode elimination:

```bash
# Comprehensive error analysis
pe refine --prompt "..." --comprehensive

# Root cause analysis
pe refine --prompt "..." --root-cause --detailed
```

Capabilities:
- Automated failure mode detection
- Root cause analysis with triggers
- Systematic error elimination
- Regression prevention testing

### Reflection-Based Learning
Meta-analysis for continuous improvement:

```bash
# Extract team knowledge
pe reflect --sessions team-sessions.json --team-knowledge

# Workflow optimization
pe reflect --sessions sessions.json --workflow-insights
```

Benefits:
- Team knowledge building and sharing
- Process optimization recommendations
- Best practice extraction
- Continuous improvement cycles

## Template Management

### `pe template`
Manage prompt templates with built-in library and custom templates.

```bash
# List available templates
pe template list

# Search templates by category
pe template search --category classification

# Apply template to prompt
pe template apply summarization --input text.txt

# Create custom template
pe template create --name custom-qa --file template.yaml
```

Built-in templates:
- Text summarization
- Sentiment classification
- Code generation
- Q&A systems
- Creative writing
- Technical documentation

## Observability and Profiling

### `pe profile`
Profiling and observability tools for performance monitoring and optimization.

```bash
# Start profiling session
pe profile start --cpu --memory

# Generate profile report
pe profile report --type cpu --output profile.html

# Real-time metrics
pe profile metrics --live --interval 1s

# Distributed tracing
pe profile trace start --service pe-eval
```

Features:
- CPU and memory profiling
- Request tracing and latency analysis
- Real-time metrics collection
- Performance bottleneck identification
- Resource usage optimization

## Interactive Development

### `pe interactive`
Start interactive REPL mode for prompt development and testing.

```bash
# Basic REPL
pe interactive

# With specific provider
pe interactive --provider anthropic:claude-3-haiku

# With custom configuration
pe interactive --config repl-config.yaml
```

REPL features:
- Live prompt testing
- Multi-provider support
- History and session management
- Template integration
- Real-time evaluation

## Project Management

### `pe init`
Initialize new prompt engineering projects with templates and structure.

```bash
pe init my-project                    # Basic project
pe init my-project --template advanced # Advanced template
pe init my-project --provider openai   # Provider-specific setup
```

### `pe watch`
Watch configuration files for changes and auto-reload evaluations.

```bash
pe watch config.yaml                  # Watch single file
pe watch configs/                     # Watch directory
pe watch config.yaml --auto-eval     # Auto-evaluate on changes
```

## Usage Patterns

### Basic Workflow
```bash
# 1. Initialize project
pe init my-prompt-project

# 2. Create and validate configuration
pe vet config.yaml

# 3. Run evaluation
pe eval config.yaml -o results.json

# 4. Analyze results
pe view results.json
pe stats results.json
```

### Pipeline Workflow
```bash
# Stream processing pipeline
pe eval config.yaml | pe filter --success | pe stream --select response | pe analyze --metric quality
```

### Optimization Workflow
```bash
# 1. Optimize prompt
pe optimize --prompt "Initial prompt" --iterations 5 --output optimized.json

# 2. Test optimized prompt
pe eval optimized-config.yaml

# 3. Compare with baseline
pe diff baseline.json optimized.json
```

### Continuous Improvement
```bash
# 1. Run regression tests
pe test regression --baseline baseline.json --current current.json

# 2. Benchmark performance
pe benchmark config.yaml --iterations 10

# 3. Profile for bottlenecks
pe profile start --cpu
pe eval config.yaml
pe profile report --type cpu
```

## Integration and Automation

### CI/CD Integration
The PE toolkit integrates seamlessly with CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Validate prompts
  run: pe vet configs/*.yaml

- name: Run prompt tests
  run: pe test property --config tests.yaml

- name: Benchmark performance
  run: pe benchmark config.yaml --format json > benchmark.json
```

### API Integration
Many commands support JSON output for programmatic use:

```bash
pe eval config.yaml --format json | jq '.results[].score'
pe benchmark config.yaml --format json | jq '.summary.avg_latency'
```

## Advanced Features

### Custom Metrics
Define custom evaluation metrics:

```yaml
# In config.yaml
metrics:
  - type: custom
    name: coherence
    evaluator: llm_judge
    criteria: "Rate coherence from 1-10"
```

### Provider Extensions
Extend with custom providers:

```go
// Register custom provider
provider.Register("custom", &CustomProvider{})
```

### Plugin System
Extend functionality with plugins:

```bash
pe plugin install optimization-suite
pe plugin list
pe plugin enable advanced-metrics
```

## Next-Generation Metaprompting Tools ⭐ 2024-2025 RESEARCH

### `pe compose`
Component-based prompt engineering with verified libraries and TextGrad optimization.

```bash
# Initialize component library
pe compose --library-init

# Add verified components to library
pe compose --add-component context-banking.txt --category context
pe compose --add-component few-shot-examples.txt --category examples

# Compose prompt from components
pe compose context.txt instruction.txt examples.txt --style cot

# Advanced composition with optimization
pe compose components/ --style few-shot --optimize --coherence
pe compose base.txt --target gpt-4 --constraints accuracy.yaml
```

**2024-2025 Features**:
- **Verified Component Libraries**: Pre-tested, reusable prompt building blocks
- **TextGrad Flow Optimization**: Coherence-based component arrangement using natural language gradients
- **Style Templates**: Chain-of-thought, few-shot, zero-shot, constitutional AI patterns
- **Model-Specific Formatting**: Architecture-optimized prompts for GPT-4, Claude-3, Gemini, etc.
- **DSPy-Inspired Composition**: Signature-based prompt construction with type safety

### `pe evolve`
Evolutionary prompt optimization using genetic algorithms and multi-objective optimization.

```bash
# Population-based evolution with NSGA-II
pe evolve baseline.txt --generations 25 --population 20 --metric accuracy,latency,cost

# Adaptive mutation operators
pe evolve prompt.txt --operators rephrase,expand,prune --adaptive-rates

# Pareto frontier exploration
pe evolve prompt.txt --multi-objective --extract-pareto-front

# Genealogy tracking for research
pe evolve prompt.txt --trace-genealogy --research-mode
```

**2024-2025 Features**:
- **NSGA-II Multi-Objective Optimization**: Pareto-optimal prompt discovery
- **Adaptive Mutation Operators**: Dynamic strategy selection based on prompt structure
- **Population Diversity Preservation**: Novel distance metrics for exploration
- **Genealogy Tracking**: Full evolutionary lineage with mutation history
- **Research-Grade Traceability**: Comprehensive optimization audit trails

### `pe fusion`
Multi-model consensus engineering for robust, ensemble-optimized prompts.

```bash
# Multi-model consensus optimization
pe fusion prompt.txt --models gpt-4,claude-3,gemini-pro --consensus weighted

# Reflection-based multi-model analysis
pe fusion prompt.txt --analyze-consensus --reflection-depth 3

# Dynamic model weighting with learning
pe fusion prompt.txt --adaptive-weights --learning-rate 0.1

# Consensus strategy comparison
pe fusion prompt.txt --strategies voting,reflection,adaptive --compare
```

**2024-2025 Features**:
- **Ensemble Learning**: Model-specific adaptation with transfer learning
- **Dynamic Model Selection**: Task-characteristic-based provider weighting
- **Reflection-Based Consensus**: Deep pattern analysis across model responses
- **Adaptive Weighting**: Performance-based model importance adjustment
- **Cross-Model Validation**: Robustness testing across LLM architectures

## Advanced Metaprompting Workflows

### Component-Based Development
```bash
# 1. Build component library
pe compose --library-init

# 2. Add verified components
pe compose --add-component context-banking.txt --category context
pe compose --add-component few-shot-classification.txt --category examples

# 3. Compose optimized prompt
pe compose --category context,instruction,examples --style cot --optimize
```

### Evolutionary Optimization Pipeline
```bash
# 1. Initialize population from base prompt
pe evolve baseline.txt --init-population 20

# 2. Run evolution with multiple objectives
pe evolve baseline.txt --generations 25 --multi-objective accuracy,latency,cost

# 3. Extract best variants for testing
pe evolve baseline.txt --extract-pareto-front --output variants.json
```

### Multi-Model Consensus Pipeline
```bash
# 1. Gather responses from model ensemble
pe fusion prompt.txt --models gpt-4,claude-3,gemini --gather-only

# 2. Analyze response patterns
pe fusion prompt.txt --analyze-consensus --output analysis.json

# 3. Generate optimized prompt
pe fusion prompt.txt --optimize-consensus --strategy weighted
```

### Hybrid Optimization Workflow
```bash
# 1. Compose initial prompt from components
pe compose context.txt instruction.txt --style cot > initial.txt

# 2. Evolve for performance optimization
pe evolve initial.txt --generations 15 --metric accuracy > evolved.txt

# 3. Fuse with multi-model consensus
pe fusion evolved.txt --models gpt-4,claude-3 --optimize > final.txt

# 4. Validate with comprehensive testing
pe test property final.txt --comprehensive
```

## Research Integration & Technical Architecture

### 2024-2025 Research Implementation

**TextGrad 2.0 Integration**:
- Natural language gradients with attention flow mapping
- Semantic drift detection during optimization trajectories  
- Backward propagation through textual feedback loops
- Cross-modal gradient computation for multimodal prompts
- Gradient strength analysis for convergence optimization

**DSPy-Inspired Component Architecture**:
- Signature-based prompt composition with type safety
- Program synthesis for automated prompt construction
- Multi-stage optimization with validation checkpoints
- Algorithmic parameter optimization using meta-learning
- Component interface definitions for reusability

**Evolutionary Metaprompting (Novel 2024)**:
- Population-based optimization with genetic algorithms
- Multi-objective optimization using NSGA-II variants
- Adaptive mutation operators based on prompt structure analysis
- Diversity preservation through semantic distance metrics
- Convergence detection with plateau identification

**Consensus-Based Multi-Model Optimization (2025)**:
- Ensemble learning for prompt robustness across models
- Model-specific adaptation with transfer learning techniques
- Consensus strategies: voting, averaging, reflection-based synthesis
- Dynamic model selection based on task characteristics
- Cross-validation with statistical significance testing

**Reflection-Based Meta-Learning**:
- Meta-analysis of optimization session patterns and outcomes
- Success pattern mining for automated strategy recommendation
- Knowledge distillation for prompt engineering principle extraction
- Continuous improvement through reinforcement learning on sessions
- Team knowledge aggregation and best practice synthesis

### Advanced Quality Assurance

**Automated Validation**:
```bash
# Property-based testing with evolved prompts
pe test property evolved-prompts/ --generate-properties

# Cross-validation between optimization methods
pe test cross-validate --methods textgrad,evolve,fusion

# Regression testing with statistical significance
pe test regression baseline.json optimized.json --significance 0.05
```

**Performance Profiling**:
```bash
# Profile optimization algorithms
pe profile optimize --method evolve --trace-performance

# Analyze convergence patterns
pe profile convergence --sessions optimization-logs/ --visualize

# Resource usage optimization
pe profile resources --optimize-memory --parallel-efficiency
```

## Best Practices

### 2025 Optimization Strategy Selection
1. **Simple Tasks**: Use `pe optimize --method textgrad` for gradient-based refinement
2. **Complex Tasks**: Start with `pe compose` for component-based building
3. **Performance Critical**: Apply `pe evolve` for multi-objective optimization with Pareto frontiers
4. **High Reliability**: Use `pe fusion` for multi-model consensus robustness
5. **Research/Exploration**: Combine all methods in hybrid workflows with full traceability
6. **Production Systems**: Use `pe fusion --adaptive-weights` for self-improving prompts
7. **Team Environments**: Leverage `pe reflect` for knowledge sharing and best practice extraction

### Quality Assurance Workflow
1. **Component Validation**: Test building blocks before composition
2. **Iterative Refinement**: Use small evolution generations for rapid iteration
3. **Cross-Model Testing**: Validate prompts across multiple LLM providers
4. **Regression Prevention**: Maintain test suites for prompt changes
5. **Performance Monitoring**: Profile optimization processes for efficiency

### Documentation and Reproducibility
1. **Version Control**: Track optimization sessions and prompt genealogy
2. **Configuration Management**: Store optimization parameters and results
3. **Knowledge Sharing**: Use reflection tools for team learning
4. **Best Practice Extraction**: Build organizational prompt engineering knowledge
5. **Continuous Improvement**: Iterate on optimization strategies based on outcomes

## Getting Help

Each command supports `--help` for detailed usage information:

```bash
pe --help                    # General help
pe eval --help              # Command-specific help
pe optimize --help          # Optimization help
```

For more information, see:
- [Getting Started Guide](docs/GETTING_STARTED.md)
- [CLI Reference](docs/CLI_REFERENCE.md)
- [Contributing Guide](CONTRIBUTING.md)