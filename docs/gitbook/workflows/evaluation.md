# Evaluation

Evaluation is the heart of reliable prompt engineering. PE provides a robust framework to test your prompts against expected behaviors.

## The `pe eval` Command

The `eval` command takes a configuration file (YAML/JSON) defining prompts, test cases, and assertions.

```bash
pe eval pe-config.yaml
```

## Configuration Structure

```yaml
prompts: [prompts/*.prompt]
providers: [openai:gpt-4]
tests:
  - vars:
      input: "test case 1"
    assert:
      - type: contains
        value: "expected output"
```

## Assertion Types

PE supports 15+ assertion types:

*   `contains`: Output contains substring.
*   `not-contains`: Output does not contain substring.
*   `equals`: Exact match.
*   `max-tokens`: Length constraint.
*   `latency`: Performance constraint.
*   `llm-rubric`: AI-graded assertion (uses a judge model).

## Viewing Results

To see a graphical report of your evaluations:

```bash
pe view
```

This launches a local web server displaying a matrix of prompts vs tests.
