# PE: Prompt Engineering Toolkit

prompt engineering tools modeled after the go toolchain

## Project Overview

PE is a comprehensive toolkit for prompt engineering that combines traditional evaluation capabilities with cutting-edge metaprompting techniques. The toolkit follows Unix philosophy with composable, pipeline-friendly commands and implements the latest 2024 research in prompt optimization.

### Core Architecture

- **Provider Interface**: Unified LLM provider abstraction supporting OpenAI, Anthropic, and other providers
- **Pipeline Processing**: Unix-style composable commands for streaming evaluation and analysis  
- **Metaprompting Engine**: Advanced prompt optimization using meta-LLMs and iterative refinement
- **Observability Suite**: Comprehensive profiling, metrics, and tracing capabilities
- **Testing Framework**: Property-based and regression testing for prompt reliability

### Key Commands

- `pe eval`: Core evaluation engine with multi-provider support
- `pe optimize`: **NEW** Metaprompting-based prompt optimization using 2024 research
- `pe benchmark`: Performance analysis with statistical significance testing
- `pe test`: Advanced testing (property-based, regression, A/B)
- `pe profile`: Real-time observability and performance profiling
- Pipeline commands: `ask`, `stream`, `filter`, `analyze` for Unix composability

## Advanced Metaprompting Implementation

The toolkit implements cutting-edge metaprompting techniques based on the latest 2024 research:

### 2024 Research Integration

**TextGrad-Style Optimization**:
- **Natural Language Gradients**: Uses LLM feedback as textual gradients for optimization
- **Attention Flow Analysis**: Maps token relationships using transformer attention patterns
- **Semantic Drift Detection**: Identifies concept preservation issues during optimization
- **Backward Propagation Through Text**: Applies gradient descent concepts to natural language

**DSPy-Inspired Techniques**:
- **Structured Prompt Generation**: Systematic approach to prompt construction
- **Multi-Stage Optimization**: Progressive refinement through Analysis → Refinement → Validation → Polishing
- **Cross-Validation**: Performance gating between optimization stages
- **Automatic Algorithm Selection**: AI-driven choice of optimization methods

**Error-Driven Refinement**:
- **Failure Mode Detection**: Automated identification of prompt weaknesses
- **Root Cause Analysis**: Deep investigation of optimization bottlenecks
- **Regression Testing**: Comprehensive test suite generation for validation
- **Quality Gate Enforcement**: Systematic quality control throughout optimization

**Reflection-Based Learning**:
- **Success Pattern Mining**: Extracts effective techniques from optimization history
- **Meta-Analysis**: Studies optimization processes to improve workflows
- **Knowledge Distillation**: Builds reusable prompt engineering principles
- **Strategy Recommendation**: Guides tool selection based on accumulated wisdom

### Advanced Usage Patterns

```bash
# TextGrad optimization with gradient analysis
pe optimize --prompt "Summarize this text" --method textgrad --iterations 5
pe analyze --prompt "Your prompt" --response "Response" --gradients

# Multi-stage optimization with quality gates
pe optimize --prompt "Complex task" --method multistage --gates

# Error-driven refinement with automated testing
pe refine --prompt "Problematic prompt" --auto-detect --generate-tests

# Reflection-based learning from optimization sessions
pe reflect --session-data sessions.json --strategies --knowledge

# Gradient computation for trajectory planning
pe gradients --prompt "Current prompt" --objective "Goal" --recommendations

# Hybrid approach combining multiple methods
pe optimize --prompt "Advanced task" --method hybrid --iterations 6 --provider anthropic
```

### Technical Architecture

**Core Metaprompting Engine** (`internal/metaprompt/`):
- `analyzer.go`: TextGrad-style gradient analysis and attention mapping
- `gradient_computer.go`: Numerical optimization with textual gradients
- `multistage.go`: Sequential optimization orchestration with quality gates
- `error_refiner.go`: Automated error detection and systematic fixing
- `reflection.go`: Meta-analysis and knowledge extraction
- `textgrad.go`: Natural language gradient computation and application
- `optimizer.go`: Unified optimization interface with method selection

**Command Integration** (`cmd/pe/`):
- Enhanced `optimize.go` with multi-method support
- New commands: `analyze`, `refine`, `gradients`, `reflect`
- Pipeline-friendly JSON output for automation
- Integration with existing evaluation and profiling infrastructure

**Provider Interface Extensions**:
- Enhanced `internal/llm` provider abstraction for meta-operations
- Support for meta-LLM evaluation and optimization
- Cross-provider optimization strategies
- Advanced prompt template management

### Quality Assurance & Validation

**Automated Testing Framework**:
- Property-based testing for prompt reliability
- Regression testing with statistical significance
- A/B testing with confidence intervals
- Continuous integration with quality gates

**Performance Monitoring**:
- Real-time optimization metrics
- Gradient strength and convergence tracking
- Error pattern frequency analysis
- Effectiveness measurement across optimization methods

**Observability Integration**:
- Distributed tracing for optimization workflows
- Performance profiling with bottleneck identification
- Resource usage optimization
- Quality metrics dashboards

## Important Code Editing Guidelines

### Go Code Editing Rules

1. **Import Management**:
   - Always maintain import statements automatically to reflect code changes
