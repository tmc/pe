# main.star - Modular configuration example

# Main configuration
configuration = config(
    provider = "openai",
    model = "gpt-4-turbo",
    prompts = [
        """You are an AI assistant with expertise in multiple domains.
        
        Context: {{context}}
        Question: {{input}}
        
        Please provide a detailed and accurate response.""",
    ],
    variables = {
        "temperature": 0.7,
        "max_tokens": 500,
    }
)

# Load test modules
math_tests = load_tests("modules/math_tests.star")
science_tests = load_tests("modules/science_tests.star") 
coding_tests = load_tests("modules/coding_tests.star")

# Combine all tests
tests = math_tests + science_tests + coding_tests

# Add some integration tests that span multiple domains
integration_tests = [
    test(
        name = "Cross-domain: Math in Physics",
        input = "Explain how calculus is used in physics to describe motion",
        variables = {
            "context": "Focus on kinematics and dynamics",
        },
        assertions = [
            contains("derivative"),
            contains("velocity"),
            contains("acceleration"),
            regex(r"(?i)newton|force"),
        ]
    ),
    
    test(
        name = "Cross-domain: Coding for Science",
        input = "Write a Python function to calculate the trajectory of a projectile",
        variables = {
            "context": "Use basic physics equations for projectile motion",
        },
        assertions = [
            contains("def"),
            contains("return"),
            regex(r"(?i)velocity|angle|time"),
            contains("math"),  # Should import math module
        ]
    ),
]

tests.extend(integration_tests)