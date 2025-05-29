# Starlark Extension for PE Prompt Files

## Overview

PE supports using Starlark (a Python-like configuration language) for defining tests, evaluations, and rubrics in prompt files. This provides a more expressive and programmable way to define complex evaluation logic while maintaining readability.

## Why Starlark?

Starlark offers several advantages for prompt engineering:

1. **Programmable Configuration**: Write loops, conditions, and functions
2. **Type Safety**: Catch errors early with strong typing
3. **Hermetic Execution**: Safe, deterministic evaluation
4. **Familiar Syntax**: Python-like syntax that's easy to learn
5. **Rich Standard Library**: Built-in functions for common operations

## File Format

PE prompt files with Starlark use the `.star` extension or can embed Starlark in YAML/JSON configs:

### Pure Starlark Format (`config.star`)

```python
# Define the prompt
prompt = """
You are a helpful AI assistant specializing in {domain}.
Please provide clear, accurate, and concise responses.
"""

# Define variables
variables = {
    "domain": ["software engineering", "data science", "machine learning"],
}

# Define test cases programmatically
def generate_tests():
    tests = []
    
    # Generate tests for each domain
    for domain in variables["domain"]:
        tests.append({
            "vars": {"domain": domain, "question": "What is a variable?"},
            "assert": [
                contains("definition"),
                min_length(50),
                max_length(200),
            ]
        })
    
    # Add edge cases
    tests.extend([
        {
            "vars": {"domain": "software engineering", "question": ""},
            "assert": [is_error("Empty question should fail")]
        },
        {
            "vars": {"domain": "unknown", "question": "Test robustness"},
            "assert": [contains("I can help"), not_contains("error")]
        }
    ])
    
    return tests

# Define custom rubrics
def quality_rubric(output, context):
    """Evaluate output quality based on multiple criteria"""
    score = 0.0
    feedback = []
    
    # Check clarity (0-0.3 points)
    if has_clear_structure(output):
        score += 0.3
        feedback.append("✓ Clear structure")
    else:
        feedback.append("✗ Improve structure")
    
    # Check accuracy (0-0.4 points)
    if is_technically_accurate(output, context["domain"]):
        score += 0.4
        feedback.append("✓ Technically accurate")
    else:
        feedback.append("✗ Technical issues found")
    
    # Check relevance (0-0.3 points)
    if addresses_question(output, context["question"]):
        score += 0.3
        feedback.append("✓ Directly addresses question")
    else:
        feedback.append("✗ Partially off-topic")
    
    return {
        "score": score,
        "feedback": "\n".join(feedback),
        "pass": score >= 0.7
    }

# Define evaluation pipeline
evaluations = [
    # Basic assertions
    {
        "name": "format_check",
        "type": "assertion",
        "rules": [
            no_markdown_errors(),
            proper_capitalization(),
            no_trailing_whitespace(),
        ]
    },
    
    # Custom rubric
    {
        "name": "quality_assessment", 
        "type": "rubric",
        "evaluator": quality_rubric,
        "weight": 0.6
    },
    
    # LLM-based evaluation
    {
        "name": "helpfulness",
        "type": "llm_judge",
        "prompt": llm_judge_prompt("helpfulness", scale=5),
        "weight": 0.4
    }
]

# Configuration
config = {
    "prompt": prompt,
    "tests": generate_tests(),
    "evaluations": evaluations,
    "providers": ["gpt-4", "claude-3"],
    "parameters": {
        "temperature": 0.7,
        "max_tokens": 500,
    }
}
```

### Embedded Starlark in YAML

```yaml
# config.yaml
prompt: |
  You are a helpful AI assistant.

# Embed Starlark for complex logic
tests:
  starlark: |
    def generate_tests():
        base_questions = [
            "What is Python?",
            "Explain recursion",
            "What are design patterns?"
        ]
        
        return [
            {
                "vars": {"question": q},
                "assert": [
                    min_length(100),
                    contains_any(["example", "for instance", "such as"]),
                    custom_check(lambda out: "```" in out if "code" in q else True)
                ]
            }
            for q in base_questions
        ]
    
    tests = generate_tests()

evaluations:
  - name: "response_quality"
    starlark: |
      def evaluate(output, context):
          metrics = analyze_response(output)
          return {
              "readability": flesch_kincaid_score(output),
              "completeness": has_all_sections(output, context),
              "examples": count_examples(output),
              "score": weighted_average(metrics)
          }
```

## Built-in Functions

PE provides these built-in functions for Starlark scripts:

### Assertion Helpers

```python
# Content checks
contains(substring)                    # Check if output contains substring
contains_any(substrings)              # Check if output contains any substring
contains_all(substrings)              # Check if output contains all substrings
not_contains(substring)               # Check if output doesn't contain substring
regex_match(pattern)                  # Check regex match
equals(expected)                      # Exact match

# Length checks  
min_length(n)                         # Minimum character count
max_length(n)                         # Maximum character count
word_count_range(min, max)            # Word count in range

# Format checks
is_json()                            # Valid JSON
is_markdown()                        # Valid Markdown
has_code_block(lang=None)            # Contains code block
no_markdown_errors()                 # No markdown syntax errors

# Quality checks
flesch_kincaid_score()               # Readability score
sentiment_score()                    # Sentiment analysis
toxicity_check()                     # Toxicity detection
```

### Evaluation Helpers

```python
# Response analysis
analyze_response(output)              # Comprehensive analysis
extract_sections(output)              # Extract markdown sections
count_examples(output)                # Count examples in response
has_clear_structure(output)           # Check response structure

# Domain-specific checks
is_technically_accurate(output, domain)    # Technical accuracy
addresses_question(output, question)       # Relevance check
has_all_sections(output, context)         # Completeness check

# Scoring
weighted_average(metrics, weights=None)    # Calculate weighted score
normalize_score(score, min, max)          # Normalize to 0-1 range
combine_scores(scores, method="average")   # Combine multiple scores
```

### LLM Judge Templates

```python
# Pre-defined judge prompts
llm_judge_prompt(criterion, scale=5)      # Generate judge prompt
custom_judge(instruction, rubric)         # Custom judge configuration

# Common criteria
judge_helpfulness(scale=5)               # Helpfulness evaluation
judge_accuracy(scale=5)                  # Accuracy evaluation  
judge_clarity(scale=5)                   # Clarity evaluation
judge_completeness(scale=5)              # Completeness evaluation
```

## Advanced Examples

### Dynamic Test Generation

```python
# Generate tests based on categories and difficulty
def generate_comprehensive_tests():
    categories = ["basic", "intermediate", "advanced"]
    topics = load_topics("topics.json")
    
    tests = []
    for category in categories:
        for topic in topics[category]:
            # Generate multiple test variations
            for variation in range(3):
                test = create_test(
                    category=category,
                    topic=topic,
                    variation=variation,
                    assertions=get_assertions_for_level(category)
                )
                tests.append(test)
    
    return tests

def get_assertions_for_level(level):
    base = [contains("explanation"), no_contradictions()]
    
    if level == "basic":
        return base + [min_length(50), max_length(200)]
    elif level == "intermediate":
        return base + [min_length(200), contains_example()]
    else:  # advanced
        return base + [
            min_length(300),
            contains_multiple_perspectives(),
            includes_references(),
            critical_analysis()
        ]
```

### Conditional Evaluation

```python
# Different evaluation criteria based on prompt type
def adaptive_evaluation(output, context):
    prompt_type = detect_prompt_type(context["prompt"])
    
    if prompt_type == "creative":
        return creative_rubric(output, context)
    elif prompt_type == "analytical":
        return analytical_rubric(output, context)
    elif prompt_type == "code":
        return code_quality_rubric(output, context)
    else:
        return general_rubric(output, context)

def code_quality_rubric(output, context):
    checks = {
        "has_code": contains_code_block(output),
        "syntax_valid": validate_syntax(output, context.get("language", "python")),
        "has_explanation": contains_explanation(output),
        "follows_practices": follows_best_practices(output),
        "handles_errors": includes_error_handling(output)
    }
    
    score = sum(1 for check in checks.values() if check) / len(checks)
    
    return {
        "score": score,
        "pass": score >= 0.8,
        "checks": checks,
        "feedback": generate_feedback(checks)
    }
```

### Multi-Stage Evaluation

```python
# Define a multi-stage evaluation pipeline
def create_evaluation_pipeline():
    return [
        # Stage 1: Basic validation
        {
            "stage": "validation",
            "evaluations": [
                length_check(min=50, max=5000),
                format_check(),
                language_check()
            ],
            "stop_on_fail": True
        },
        
        # Stage 2: Content quality
        {
            "stage": "quality",
            "evaluations": [
                accuracy_check(threshold=0.8),
                completeness_check(required_sections),
                coherence_check()
            ],
            "weight": 0.6
        },
        
        # Stage 3: Advanced analysis
        {
            "stage": "analysis",
            "evaluations": [
                originality_check(),
                depth_analysis(),
                comparative_evaluation(baseline_outputs)
            ],
            "weight": 0.4
        }
    ]
```

### Statistical Validation

```python
# A/B testing with statistical significance
def statistical_test_suite():
    control_prompt = load_prompt("control.txt")
    variant_prompt = load_prompt("variant.txt")
    
    # Generate test cases
    test_cases = generate_diverse_test_set(n=100)
    
    # Define metrics
    metrics = [
        {
            "name": "helpfulness",
            "evaluator": judge_helpfulness,
            "higher_is_better": True
        },
        {
            "name": "response_time",
            "evaluator": measure_latency,
            "higher_is_better": False
        },
        {
            "name": "token_efficiency",
            "evaluator": lambda o, c: len(c["output"]) / c["tokens_used"],
            "higher_is_better": True
        }
    ]
    
    # Configure statistical test
    return {
        "type": "ab_test",
        "control": control_prompt,
        "variant": variant_prompt,
        "test_cases": test_cases,
        "metrics": metrics,
        "statistical_config": {
            "confidence_level": 0.95,
            "minimum_effect_size": 0.1,
            "test_type": "two_tailed"
        }
    }
```

## Integration with PE

### Running Starlark-based Tests

```bash
# Run Starlark configuration
pe test config.star

# Run with specific functions
pe test config.star --function generate_advanced_tests

# Export Starlark config to YAML
pe convert config.star --output config.yaml

# Validate Starlark syntax
pe validate config.star
```

### Combining with PE Commands

```bash
# Use Starlark for test generation in optimization
pe optimize prompt.txt --test-generator tests.star

# Apply Starlark rubrics to existing results
pe evaluate results.json --rubric quality.star

# Generate test cases and run them
pe test generate --generator cases.star | pe test run -
```

## Best Practices

1. **Keep Logic Simple**: Use Starlark for configuration logic, not complex algorithms
2. **Modularize**: Split complex evaluations into reusable functions
3. **Type Hints**: Use clear variable names and comments for types
4. **Test Generators**: Test your test generators with small datasets first
5. **Version Control**: Track Starlark files like code, with proper diffs
6. **Performance**: Be mindful of computational complexity in loops

## Debugging

```bash
# Debug Starlark execution
pe test config.star --debug

# Print intermediate values
pe test config.star --trace

# Validate without running
pe validate config.star --strict
```

## Examples Repository

Find more examples at:
- `examples/starlark/` - Complete Starlark configurations
- `examples/rubrics/` - Reusable evaluation rubrics
- `examples/generators/` - Test case generators