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

## Metaprompting Implementation

The toolkit implements cutting-edge metaprompting techniques:

### 2024 Research Integration
- **DSPy-Style Optimization**: Structured prompt generation with feedback loops
- **Textual Gradients**: Natural language feedback for iterative improvement
- **Reflection Mechanisms**: Self-critique and analysis capabilities
- **Meta-LLM Evaluation**: Using higher-capability models to optimize production prompts

### Usage Patterns
```bash
# Basic optimization
pe optimize --prompt "Summarize this text" --iterations 3

# Advanced optimization with evaluation
pe optimize --prompt "Complex task" --iterations 5 --provider anthropic --output results.json

# Pipeline integration
pe optimize --prompt "Initial" | pe eval --config test.yaml | pe analyze
```

### Technical Implementation
- `internal/metaprompt`: Core optimization engine
- `cmd/pe/optimize.go`: Command-line interface
- Integration with existing `internal/llm` provider interface
- JSON output for automation and CI/CD integration

## Important Code Editing Guidelines

### Go Code Editing Rules

1. **Import Management**:
   - Always maintain import statements automatically to reflect code changes
