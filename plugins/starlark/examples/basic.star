# basic.star - Basic example of Starlark configuration for PE

# Define the evaluation configuration
configuration = config(
    provider = "openai",
    model = "gpt-4",
    prompts = [
        "You are a helpful assistant. {{input}}",
    ],
    variables = {
        "temperature": 0.7,
        "max_tokens": 150,
    }
)

# Define test cases
tests = [
    test(
        name = "Simple greeting",
        input = "Say hello!",
        assertions = [
            icontains("hello"),
        ],
        description = "Test basic greeting response"
    ),
    
    test(
        name = "Basic math",
        input = "What is 2 + 2?",
        assertions = [
            contains("4"),
        ],
        description = "Test arithmetic capability"
    ),
    
    test(
        name = "Multiple assertions",
        input = "Write a haiku about programming",
        assertions = [
            # Haikus have 3 lines
            regex(r".*\n.*\n.*"),
            # Often mention code or programming concepts
            regex(r"(?i)(code|program|debug|function|variable)"),
        ],
        description = "Test creative writing with constraints"
    ),
    
    test(
        name = "JSON generation",
        input = "Generate a JSON object representing a user with name 'Alice' and age 25",
        assertions = [
            contains("{"),
            contains("}"),
            contains('"name"'),
            contains('"Alice"'),
            contains('"age"'),
            contains("25"),
        ],
        description = "Test structured output generation"
    ),
    
    test(
        name = "Negative assertion",
        input = "Explain quantum computing without using the word 'quantum'",
        assertions = [
            not_contains("quantum"),
            contains("computing"),
        ],
        description = "Test ability to avoid specific words"
    ),
]