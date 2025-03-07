# PE: Prompt Engineering Toolkit Architecture

This document provides an overview of the architecture of the Prompt Engineering toolkit.

## Component Overview

```
+---------------------+      +---------------------+
|                     |      |                     |
|     PE CLI Tool     |      |  Workbench (WB)     |
|     (cmd/pe)        |<---->|  (cmd/wb)           |
|                     |      |                     |
+----------+----------+      +---------+-----------+
           |                           |
           |                           |
           v                           v
+---------------------+      +---------------------+
|                     |      |                     |
|  Template Engine    |      |  Prompt Format      |
|  (internal/template)|      |  Conversion         |
|                     |      |                     |
+----------+----------+      +---------------------+
           |
           |
+----------v----------+      +---------------------+
|                     |      |                     |
|  LLM Abstraction    |<---->|  Provider           |
|  (internal/llm)     |      |  Implementations    |
|                     |      |                     |
+----------+----------+      +---------------------+
           |
           |
+----------v----------+      +---------------------+
|                     |      |                     |
|  Evaluator          |<---->|  Assertion          |
|  (internal/evaluator|      |  Utilities          |
|                     |      |                     |
+---------------------+      +---------------------+
```

## Core Components

### Command-Line Tools

1. **PE CLI (cmd/pe)**
   - Main command-line interface for evaluation and testing
   - Commands: eval, view, vet, fmt, convert, benchmark
   - Primary user interface for testing prompts

2. **Workbench (cmd/wb)**
   - Tool for interactive prompt creation and refinement
   - Focus on the creative aspects of prompt engineering
   - Commands: create, edit, list, test

### Internal Packages

1. **Template Engine (internal/template)**
   - Processes templates with variable substitution
   - Handles conditional rendering and file inclusions
   - Core templating functionality for both tools

2. **LLM Abstraction (internal/llm)**
   - Provider-agnostic interface for LLM services
   - Registry of available providers
   - Common types and utilities

3. **Provider Implementations (internal/llm/providers)**
   - Individual implementations for each LLM service
   - OpenAI, Anthropic, Google AI, etc.
   - Provider-specific API integration

4. **CGPT Integration (internal/cgpt)**
   - Integration with the CGPT command-line tool
   - Allows using CGPT as an LLM provider

5. **Evaluator (internal/evaluator)**
   - Executes evaluations of prompts against providers
   - Validates configuration files
   - Processes test results

6. **Assertion Utilities (internal/assertutil)**
   - Tools for validating LLM responses
   - Verification of test assertions
   - Supporting the evaluation workflow

### Testing Infrastructure

1. **Unit Tests**
   - Package-level tests in each package
   - Testable interfaces and pure functions

2. **Basic Tests (tests/basic_tests)**
   - Core functionality tests
   - Implementation-independent verification

3. **Test Runner (scripts/run-all-tests.sh)**
   - Unified test execution
   - Consistent test environment

## Data Flow

1. **Configuration**
   - User provides YAML/JSON configuration
   - Specifies prompts, variables, providers, and assertions

2. **Template Processing**
   - Templates are rendered with variables
   - Conditional sections are evaluated
   - File includes are processed

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
   - Can be viewed in various formats (JSON, YAML, UI)

## Extension Points

The toolkit is designed to be extensible in several key areas:

1. **LLM Providers**
   - Add new providers in `internal/llm/providers/`
   - Register in the provider registry

2. **Assertion Types**
   - Extend assertion capabilities in `internal/assertutil/`
   - Create new assertion types for specific needs

3. **Template Functionality**
   - Enhance the template engine in `internal/template/`
   - Add new template functions or processing features

4. **Command-Line Tools**
   - Extend existing tools with new commands
   - Create complementary tools for specific workflows

## Configuration Format

The PE toolkit uses a structured configuration format:

```yaml
prompts:
  - name: "example_prompt"
    template: "Answer the following question: {{.question}}"

vars:
  - question: "What is the capital of France?"

providers:
  - openai:gpt-4
  - anthropic:claude-3-opus

tests:
  - assertions:
      - type: "contains"
        value: "Paris"
```

This format is designed to be:
- Human-readable and editable
- Version-control friendly
- Flexible for different use cases
- Extensible for future features
