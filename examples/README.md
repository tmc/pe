# PE Examples

This directory contains examples demonstrating various features and use cases of PE (Prompt Engineering toolkit).

## Quick Start Examples

### Basic Usage
- [Simple Prompt](basic/simple-prompt/) - Running a basic prompt
- [Template Variables](basic/template-vars/) - Using template variables
- [Multiple Providers](basic/multi-provider/) - Testing across providers

### Evaluation
- [Basic Evaluation](evaluation/basic/) - Simple evaluation setup
- [Assertions](evaluation/assertions/) - Using different assertion types
- [Pass@N Metrics](evaluation/pass-at-n/) - Testing with multiple attempts

### Optimization
- [PE2 Optimization](optimization/pe2/) - Prompt Engineer 2 method
- [Semantic Gradient](optimization/semantic/) - Semantic backpropagation
- [Evolutionary](optimization/evolve/) - Evolutionary optimization

## Advanced Examples

### Module System
- [Creating Modules](modules/create/) - Building PE modules
- [Using Modules](modules/use/) - Importing and using modules
- [Module Registry](modules/registry/) - Publishing to registry

### Pipeline Processing
- [Unix Pipes](pipeline/unix/) - Composing PE commands
- [Stream Processing](pipeline/stream/) - Real-time streaming
- [Batch Processing](pipeline/batch/) - Processing multiple inputs

### Integration Examples
- [CI/CD Pipeline](integration/cicd/) - GitHub Actions integration
- [Testing Framework](integration/testing/) - Unit testing prompts
- [Monitoring](integration/monitoring/) - Tracking costs and performance

## API Examples

### Go API
- [Inference API](inference/) - Using PE's inference API programmatically
- [Evaluation API](api/evaluation/) - Programmatic evaluation
- [Custom Providers](api/providers/) - Implementing custom providers

### Starlark Scripting
- [Basic Scripts](starlark/) - Starlark configuration examples
- [Advanced Logic](starlark/advanced/) - Complex evaluation logic

## Running Examples

Most examples can be run directly:

```bash
# Run a simple example
cd basic/simple-prompt
pe run prompt.txt --provider openai

# Run an evaluation
cd evaluation/basic
pe eval config.yaml

# Run optimization
cd optimization/pe2
pe optimize prompt.txt --method pe2
```

## Prerequisites

1. **Install PE**:
   ```bash
   go install github.com/tmc/pe/cmd/pe@latest
   ```

2. **Set API Keys**:
   ```bash
   export OPENAI_API_KEY="your-key"
   export ANTHROPIC_API_KEY="your-key"
   ```

3. **Verify Installation**:
   ```bash
   pe --help
   ```

## Directory Structure

```
examples/
├── basic/              # Getting started examples
├── evaluation/         # Evaluation configurations
├── optimization/       # Prompt optimization examples
├── modules/           # Module system examples
├── pipeline/          # Unix-style pipeline examples
├── integration/       # Integration with other tools
├── api/              # Go API usage
├── starlark/         # Starlark scripting
└── prompts/          # Sample prompt files
```

## Contributing

To add new examples:
1. Create a directory for your example
2. Include a README.md explaining the example
3. Add working configuration files
4. Test the example thoroughly
5. Submit a pull request

## Resources

- [PE Documentation](../docs/)
- [Template Syntax Guide](../docs/TEMPLATE_SYNTAX.md)
- [Command Reference](../docs/COMMANDS.md)
- [GitHub Repository](https://github.com/tmc/pe)