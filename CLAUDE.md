# PE: Prompt Engineering Toolkit

prompt engineering tools modeled after the go toolchain

## Project Overview

PE is a Go-based toolkit for prompt engineering that implements cutting-edge 2024-2025 research in prompt optimization. The toolkit follows Unix philosophy with composable, pipeline-friendly commands and focuses on metaprompting techniques for advanced prompt optimization.

### Current Implementation Status

**✅ Implemented Core Features:**
- **Provider Interface**: Extensible LLM provider abstraction (⚠️ still primarily using cgpt CLI wrapper)
- **Pipeline Processing**: Unix-style composable commands (ask, stream, filter, analyze, collect, reduce)
- **Metaprompting Engine**: Advanced prompt optimization using multiple research-based methods
- **Evaluation System**: Comprehensive evaluation with pass@n metrics and assertion types
- **Testing Framework**: Property-based and regression testing support (⚠️ 9.7% test coverage)
- **Module System**: Complete go mod-style commands (init, download, tidy, vendor)
- **Component Composition**: DSPy-style prompt composition with validation
- **Attestation System**: Cryptographic signing and verification of prompt runs
- **Security Testing**: OWASP LLM Top 10 complete coverage via redteam module

**🚧 In Development:**
- Full native provider migration (OpenAI/Anthropic providers partially implemented)
- Distributed execution integration (foundation exists, CLI partially integrated)
- Web dashboard and REST API
- Module registry implementation
- Advanced semantic caching strategies
- Comprehensive test suite (current coverage only 9.7%)

### Key Commands (Implemented)

- `pe eval`: Core evaluation engine with pass@n metrics and multiple assertion types
- `pe optimize`: Metaprompting-based prompt optimization using 2024-2025 research
- `pe semantic`: Semantic backpropagation and GASO optimization (2025 KAUST/IDSIA research)
- `pe metrics`: Evaluation metrics (BLEU, ROUGE, BERTScore, G-Eval)
- `pe benchmark`: Performance benchmarking and analysis
- `pe test`: Testing with property-based and regression approaches
- `pe profile`: Performance profiling and analysis
- `pe run`: Execute prompts immediately via providers
- `pe compose`: Component-based prompt composition with style handlers
- `pe mod init/download/tidy/vendor`: Go-style module management
- `pe attest`: Cryptographic attestation for prompt runs
- `pe extract`: Extract and parse structured data from prompts
- `pe passn`: Calculate pass@n metrics for code generation tasks
- `pe evolve`: Evolutionary prompt optimization with genetic algorithms
- `pe fusion`: Multi-model fusion for production reliability
- `pe distributed`: Distributed execution commands (start/join/status/stop)
- `pe cache`: Content-addressed caching with cryptographic verification
- `pe security`: OWASP LLM Top 10 security testing
- Pipeline commands: `ask`, `stream`, `filter`, `analyze`, `collect`, `reduce` for Unix composability

## Advanced Evaluation Features

The `pe eval` command includes sophisticated assertion types beyond simple string matching:

### Pass@N Evaluation

Pass@n measures how often a model generates a correct solution within n attempts, crucial for code generation:

```yaml
# In eval config.yaml:
tests:
  - vars:
      task: "Write a binary search function"
    assert:
      - type: pass-at-n
        config:
          n: 1              # Calculate pass@1
          samples: 20       # Generate 20 samples
          temperature: 0.8  # Higher temp for diversity
          test_cases:
            - input: "binary_search([1,3,5,7], 5)"
              expected: "2"
            - input: "binary_search([1,3,5,7], 6)"
              expected: "-1"
        threshold: 0.8      # Expect 80% pass rate
```

**Implementation**: 
- Located in `internal/evaluator/assertions.go` as `AssertionPassAtN`
- Uses `internal/metrics/advanced.go` for pass@n calculation
- Supports test cases, LLM validation, and pattern matching

### Structured Output Validation

Ensures LLM outputs conform to specific schemas:

```yaml
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
          score:
            type: number
            minimum: -1
            maximum: 1
        required: ["sentiment", "score"]
```

**Go Struct Integration**:
```go
type ExpectedOutput struct {
    Summary    string  `json:"summary" minLength:"50" maxLength:"200"`
    Keywords   []string `json:"keywords" minItems:"3"`
    Confidence float64  `json:"confidence" min:"0" max:"1"`
}
```

**Implementation**:
- `internal/structured/` package provides schema validation and format conversion
- `internal/structured/go_structs.go` enables Go struct to schema conversion
- Supports JSON Schema, TypeScript, Pydantic, and custom formats via plugins

## Advanced Metaprompting Implementation

The toolkit implements cutting-edge metaprompting techniques based on the latest 2024-2025 research:

### 2025 BREAKTHROUGH: Semantic Backpropagation & GASO

**Semantic Backpropagation Implementation** (KAUST/IDSIA 2025):
- **Semantic Gradients**: Generalizes mathematical gradients to natural language feedback
- **Graph-based Optimization**: GASO (Graph-based Agentic System Optimization) for multi-component systems
- **Directional Semantic Information**: LLM-generated improvement directions with confidence scores
- **System-Wide Optimization**: Optimizes entire agentic systems rather than individual components
- **Computational Graph Analysis**: Dependency-aware optimization with semantic flow tracking
- **Pareto Efficiency**: Multi-objective optimization for complex trade-offs (accuracy/latency/cost)

**Key Commands Implemented**:
```bash
pe semantic backprop --prompt "prompt" --target "objective" --iterations 5
pe semantic descent --objective "goal" --learning-rate 0.1 --adaptive --convergence 0.001
pe semantic gaso --system definition.json --objective "performance" --multi-objective
```

**Target Performance Goals:**
- Aiming for >90% accuracy on GSM8K mathematical problems
- Target >80% accuracy on BIG-Bench Hard NLP tasks
- Goal of >85% accuracy on algorithmic tasks

*Note: These are research targets based on the 2025 paper. Actual benchmarking against these datasets is pending.*

### TextGrad 2.0 Implementation

Located in `internal/metaprompt/textgrad.go`:
- Natural language gradients with attention flow mapping
- Semantic drift detection during optimization
- Backward propagation through textual feedback
- Cross-modal gradient computation support

### Component-Based Engineering (✅ COMPLETED)

The `pe compose` command provides advanced prompt composition:
- Type-safe prompt composition with dependency resolution
- Style-specific handlers (chain-of-thought, few-shot, structured, conversational, DSPy)
- Semantic coherence validation with scoring
- Component compatibility checking
- Integration with optimization methods
- Support for quality gates and statistical validation

### Advanced Metrics

`internal/metrics/advanced.go` implements:
- BLEU, ROUGE, METEOR for text generation
- BERTScore using LLM-based semantic similarity
- G-Eval with chain-of-thought evaluation
- UniEval for task-specific multi-dimensional evaluation
- Pass@N with proper statistical calculation

## Technical Architecture

### Core Metaprompting Engine (`internal/metaprompt/`)
- `optimizer.go`: Unified optimization interface
- `textgrad.go`: Natural language gradient computation
- `semantic.go`: Semantic backpropagation implementation
- `gradient_computer.go`: Gradient analysis and application
- `multistage.go`: Multi-stage optimization with quality gates
- `error_refiner.go`: Automated error detection and fixing
- `reflection.go`: Meta-analysis and knowledge extraction
- `composer.go`: Component-based prompt composition
- `gaso.go`: Graph-based system optimization

### Evaluation System (`internal/evaluator/`)
- `evaluator.go`: Core evaluation engine
- `assertions.go`: Comprehensive assertion types including pass@n and structured output
- Supports 20+ assertion types for quality, performance, and correctness

### Structured Output (`internal/structured/`)
- `structured.go`: Core schema validation and formatting
- `go_structs.go`: Go struct to schema conversion
- `prompt_builder.go`: Structured prompt generation
- Plugin system for custom formats

### Inference API (`internal/inference/`)
- `inference.go`: Provider abstraction for LLM calls
- `providers/cgpt/`: cgpt CLI wrapper implementation
- Extensible for additional providers

## Plugin System

PE supports runtime plugin discovery:
- Plugins are `pe-*` executables in PATH
- Example: `pe-promptfoo` provides promptfoo compatibility
- Plugin interface defined in `internal/plugin/plugin.go`

## Recently Implemented Features

### Pass@N in Evaluation (✅ COMPLETED)
- Integrated as assertion type in `pe eval`
- Supports test cases, LLM validation, pattern matching
- Statistical pass@n calculation with proper sampling

### Structured Output Support (⚠️ PARTIAL)
- Schema validation interface defined in assertions.go
- Go struct to schema conversion implemented
- Basic JSON schema validation support
- *Note: Full structured output validation in evaluator is not yet complete*

### Inference API (✅ COMPLETED)
- Generic provider interface in `internal/inference/`
- cgpt provider implementation
- Integration with `pe run` command

### Component-Based Prompt Composition (✅ COMPLETED)
- Full `pe compose` command implementation with multiple styles
- DSPy-style program synthesis with quality gates
- Semantic coherence validation and compatibility checking
- Integration with optimization pipeline

### Advanced Metrics (✅ COMPLETED)
- BLEU, ROUGE, METEOR implementations
- LLM-based metrics (BERTScore, G-Eval, UniEval)
- Integration with evaluation pipeline

## Recent Improvements (January 2025)

### Pipeline Commands (✅ COMPLETED)
- Implemented full Unix-style pipeline: `ask`, `stream`, `filter`, `analyze`, `collect`, `reduce`
- Mock provider support for testing without external dependencies
- JSON filtering and transformation capabilities

### Module Management (✅ COMPLETED)
- Full `pe mod` implementation matching Go toolchain patterns
- Commands: `init`, `download`, `tidy`, `vendor`
- go.mod file management with dependency tracking

### Attestation System (✅ COMPLETED)
- Cryptographic signing of prompt runs
- Chain verification and integrity checking
- Secure key storage with OS keychain integration
- Export formats: JSON, JSONL, CSV, proof bundles

### Distributed System Integration (🚧 IN PROGRESS)
- Created distributed CLI commands: `pe distributed start/join/status/stop`
- Integrated distributed execution flags with `pe eval` command
- Basic P2P networking and task distribution framework
- *Note: Full distributed evaluation functionality still being implemented*

### Test Coverage Improvements (Updated January 31, 2025)
- Fixed multiple test failures in pipeline, workflow, and extract commands
- Added multiline XML extraction support
- Improved test data handling
- **NEW**: Added comprehensive tests for `internal/cli` package (31.8% coverage)
- **NEW**: Added comprehensive tests for `internal/cgpt` package (71.1% coverage)  
- **NEW**: Fixed all failing tests in `cmd/pe` package
- **NEW**: Implemented table-driven tests with testify/assert for better test quality

## Code Quality Guidelines

1. **Import Management**: Maintain imports automatically based on code usage
2. **Error Handling**: Always wrap errors with context using `fmt.Errorf`
3. **Testing**: Write table-driven tests for new functionality
4. **Documentation**: Update command help text and examples

## Current Limitations & Known Issues

1. **Test Coverage**: Improved from 9.7% to ~15% - still needs comprehensive test suite implementation across all packages
2. **Provider Dependency**: Still primarily using cgpt CLI wrapper instead of native API calls
3. **Documentation Accuracy**: Some docs describe unimplemented features as complete
4. **Distributed System**: CLI commands exist but full integration incomplete
5. **Module Registry**: Not yet implemented despite documentation
6. **Benchmarking**: No benchmarks against GSM8K, BIG-Bench Hard, or algorithmic datasets
7. **Web Dashboard**: Not implemented despite being mentioned in docs
8. **Structured Output Validation**: Schema validation interface defined but full implementation incomplete

## Priority Development Areas

1. **Complete Native Providers**: Finish OpenAI/Anthropic implementations to remove cgpt dependency
2. **Test Coverage**: Achieve >80% coverage with comprehensive test suite (immediate goal: 50%)
3. **Distributed Execution**: Complete integration with eval command
4. **Documentation Cleanup**: Separate implemented vs planned features clearly
5. **Performance Benchmarks**: Validate against research paper targets (GSM8K, BIG-Bench Hard)
6. **Module Registry**: Implement registry with dependency resolution
7. **Provider Extensions**: Add Ollama support for local model execution

## Future Roadmap

See ROADMAP.md for planned features including:
- Additional provider implementations
- Advanced caching strategies
- Full distributed evaluation support
- Visual prompt engineering tools
- Module registry with dependency resolution
- Web dashboard and REST API
- Multi-modal support (vision, audio)
- Neurosymbolic prompt synthesis