# PE: Architecture

This document provides an overview of the architecture of the Prompt Engineering toolkit.

## Component Overview

```
+---------------------+
|                     |
|     PE CLI Tool     |
|     (cmd/pe)        |
|                     |
+----------+----------+
           |
           |
           v
+---------------------+      
|                     |      
|  Template Engine    |      
|                     |      
+----------+----------+      
           |
           |
+----------v----------+      +---------------------+
|                     |      |                     |
|  LLM Abstraction    |<---->|  Provider           |
|                     |      |  Implementations    |
|                     |      |                     |
+----------+----------+      +---------------------+
           |
           |
+----------v----------+      +---------------------+
|                     |      |                     |
|  Evaluator          |<---->|  Assertion          |
|                     |      |  Utilities          |
|                     |      |                     |
+---------------------+      +---------------------+
```

## Core Components

### Command-Line Tools

1. **PE CLI (cmd/pe)**
   - Main command-line interface for evaluation and testing
   - Commands: eval, view, csv, benchmark

### Internal Packages

1. **Template Engine**
   - Processes templates with variable substitution
   - Core templating functionality

2. **LLM Abstraction**
   - Provider-agnostic interface for LLM services
   - Registry of available providers

3. **Provider Implementations**
   - Individual implementations for each LLM service
   - OpenAI, Anthropic, Google AI, etc.

4. **CGPT Integration**
   - Integration with the CGPT command-line tool
   - Allows using CGPT as an LLM provider

5. **Evaluator**
   - Executes evaluations of prompts against providers
   - Validates configuration files
   - Processes test results

6. **Assertion Utilities**
   - Tools for validating LLM responses
   - Verification of test assertions

## Data Flow

1. **Configuration**
   - User provides YAML/JSON configuration
   - Specifies prompts, variables, providers, and assertions

2. **Template Processing**
   - Templates are rendered with variables

3. **Provider Selection**
   - Appropriate LLM provider is selected
   - API keys and configuration applied

4. **Evaluation**
   - Prompts are sent to LLM providers
   - Responses are collected and processed

5. **Assertion**
   - Responses are validated against assertions
   - Results are aggregated and reported

6. **Reporting**
   - Results are formatted and displayed
   - Can be viewed in various formats (JSON, YAML, text)

## Extension Points

The toolkit is designed to be extensible in several key areas:

1. **LLM Providers**
   - Add new providers in `internal/llm/providers/`
   - Register in the provider registry

2. **Assertion Types**
   - Extend assertion capabilities
   - Create new assertion types for specific needs

## Configuration Format

The PE toolkit uses a structured configuration format:

```yaml
prompts:
  - "Answer the following question: {{question}}"

providers:
  - openai:gpt-4
  - anthropic:claude-3-haiku

tests:
  - vars:
      question: "What is the capital of France?"
    assert:
      - type: "contains"
        value: "Paris"
```

This format is designed to be:
- Human-readable and editable
- Version-control friendly
- Flexible for different use cases