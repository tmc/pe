# Starlark Extensions

PE supports [Starlark](https://github.com/google/starlark-go) (a Python dialect) for defining complex configuration logic, assertions, and rubrics.

## Configuration Tests (`config.star`)

Instead of static YAML, you can programmatically generate test cases:

```python
# config.star

def generate_tests():
    languages = ["Python", "JavaScript", "Go"]
    tests = []
    
    for lang in languages:
        tests.append({
            "vars": {"topic": "loops", "lang": lang},
            "assert": [contains("for"), min_length(50)]
        })
            
    return tests

config = {
    "prompts": ["tutorial.prompt"],
    "tests": generate_tests()
}
```

## Custom Assertions

Starlark is powerful for custom validation logic:

```python
def check_code_quality(output):
    if "TODO" in output:
        fail("Output contains TODOs")
    if len(output.splitlines()) > 100:
        fail("Output too long")
    return True
```

## Usage

Use `.star` files exactly like `.yaml` config files:

```bash
pe eval config.star
```
