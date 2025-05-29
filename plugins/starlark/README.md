# Starlark Plugin for PE

The Starlark plugin enables writing test configurations in [Starlark](https://github.com/bazelbuild/starlark), a Python-like configuration language. This provides a more expressive and programmable way to define prompt evaluations.

## Features

- **Python-like Syntax**: Write tests using familiar Python-like syntax
- **Built-in Assertions**: Pre-defined assertion functions for common patterns
- **Modular Tests**: Load and compose tests from multiple files
- **Variables & Templates**: Support for global and test-specific variables
- **Type Safety**: Starlark's type system prevents common configuration errors
- **Programmable**: Use loops, conditions, and functions to generate tests

## Installation

```bash
cd plugins/starlark
go build -o starlark-eval
```

## Usage

```bash
# Run evaluation with Starlark config
./starlark-eval config.star

# Output results to JSON file
./starlark-eval config.star results.json
```

## Configuration Format

### Basic Example

```python
# config.star

# Define configuration
configuration = config(
    provider = "openai",
    model = "gpt-4",
    prompts = [
        "You are a helpful assistant. Answer the question: {{input}}",
    ],
    variables = {
        "temperature": 0.7,
    }
)

# Define tests
tests = [
    test(
        name = "Basic math test",
        input = "What is 2 + 2?",
        assertions = [
            contains("4"),
            contains("four"),
        ],
        description = "Test basic arithmetic understanding"
    ),
    
    test(
        name = "Code generation",
        input = "Write a Python function to calculate factorial",
        assertions = [
            contains("def factorial"),
            regex(r"factorial\(\d+\)"),
            contains("return"),
        ]
    ),
]
```

### Advanced Example with Functions

```python
# test_suite.star

def make_math_test(a, b, operation):
    """Generate a math test case"""
    operations = {
        "+": a + b,
        "-": a - b,
        "*": a * b,
        "/": a / b if b != 0 else "undefined",
    }
    
    return test(
        name = "Math: {} {} {}".format(a, operation, b),
        input = "What is {} {} {}?".format(a, operation, b),
        assertions = [
            contains(str(operations[operation])),
        ]
    )

# Generate multiple test cases
tests = [
    make_math_test(2, 2, "+"),
    make_math_test(10, 5, "-"),
    make_math_test(3, 4, "*"),
    make_math_test(15, 3, "/"),
]

# Add custom tests
tests.extend([
    test(
        name = "Complex reasoning",
        input = "If a train travels 60 mph for 2 hours, how far does it go?",
        assertions = [
            contains("120"),
            icontains("miles"),
        ]
    )
])
```

### Modular Configuration

```python
# main.star

# Load shared configuration
configuration = config(
    provider = "anthropic",
    model = "claude-3-opus",
    prompts = [
        "You are Claude, a helpful AI assistant. {{input}}",
    ]
)

# Load tests from other files
math_tests = load_tests("tests/math_tests.star")
reasoning_tests = load_tests("tests/reasoning_tests.star")
coding_tests = load_tests("tests/coding_tests.star")

# Combine all tests
tests = math_tests + reasoning_tests + coding_tests
```

## Built-in Functions

### Assertion Functions

- `equals(value, message="")` - Exact match
- `contains(value, message="")` - Substring match
- `icontains(value, message="")` - Case-insensitive substring match
- `regex(pattern, message="")` - Regular expression match
- `starts_with(value, message="")` - String prefix match
- `ends_with(value, message="")` - String suffix match
- `not_contains(value, message="")` - Ensure substring is not present

### Configuration Functions

- `config(provider, model, prompts, variables={})` - Define evaluation configuration
- `test(name, input, assertions, variables={}, description="")` - Define a test case
- `load_tests(path)` - Load tests from another Starlark file

## Examples

### Testing Different Response Formats

```python
# format_tests.star

json_test = test(
    name = "JSON response format",
    input = "Return a JSON object with name and age for a person named John who is 30",
    assertions = [
        contains("{"),
        contains("}"),
        contains('"name"'),
        contains('"John"'),
        contains('"age"'),
        contains("30"),
    ]
)

list_test = test(
    name = "List format",
    input = "List 3 benefits of exercise",
    assertions = [
        regex(r"[1-3]\."),  # Numbered list
        contains("health"),  # Common benefits
    ]
)

tests = [json_test, list_test]
```

### Conditional Test Generation

```python
# conditional_tests.star

def create_language_tests(languages):
    """Create translation tests for multiple languages"""
    tests = []
    
    for lang in languages:
        tests.append(test(
            name = "Translate to " + lang["name"],
            input = 'Translate "Hello, world!" to ' + lang["name"],
            assertions = [
                contains(lang["hello"]),
            ]
        ))
    
    return tests

languages = [
    {"name": "Spanish", "hello": "Hola"},
    {"name": "French", "hello": "Bonjour"},
    {"name": "German", "hello": "Hallo"},
    {"name": "Japanese", "hello": "こんにちは"},
]

tests = create_language_tests(languages)
```

### Testing with Variables

```python
# variable_tests.star

configuration = config(
    provider = "openai",
    model = "gpt-4",
    prompts = [
        """Context: {{context}}
        
        Question: {{question}}
        
        Please answer based on the given context.""",
    ]
)

tests = [
    test(
        name = "Context-based QA",
        variables = {
            "context": "The Python programming language was created by Guido van Rossum and first released in 1991.",
            "question": "Who created Python?",
        },
        assertions = [
            contains("Guido van Rossum"),
        ]
    ),
    
    test(
        name = "Date extraction",
        variables = {
            "context": "The company was founded in 2005 and went public in 2012.",
            "question": "When was the company founded?",
        },
        assertions = [
            contains("2005"),
        ]
    ),
]
```

## Integration with PE

The Starlark plugin integrates seamlessly with PE's evaluation framework:

1. **Provider Support**: Works with all PE-supported providers (OpenAI, Anthropic, etc.)
2. **Metrics**: Evaluation results include all standard PE metrics
3. **Output Formats**: Results can be exported in JSON format for further analysis
4. **Pipeline Integration**: Can be used as part of PE's pipeline commands

## Best Practices

1. **Organize Tests**: Use functions to generate similar tests and avoid repetition
2. **Modular Design**: Split tests into logical files and use `load_tests()`
3. **Clear Names**: Use descriptive test names that explain what's being tested
4. **Multiple Assertions**: Combine assertions to thoroughly validate responses
5. **Variables**: Use variables for reusable prompts and test data
6. **Comments**: Document complex test logic and assertion patterns

## Limitations

- Starlark is intentionally limited (no file I/O, network access, etc.)
- All tests run through PE's standard evaluation framework
- Real-time streaming evaluation is not supported in the initial version

## Future Enhancements

- Custom assertion types via Starlark functions
- Integration with PE's metaprompting features
- Support for async test execution
- Test result post-processing hooks
- Direct CLI integration as `pe test --format starlark`