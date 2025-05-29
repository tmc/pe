# advanced.star - Advanced Starlark configuration showing functions and loops

# Helper function to create consistent test format
def qa_test(question, expected_keywords, name=None):
    """Create a Q&A test with standard assertions"""
    return test(
        name = name or question[:30] + "...",
        input = question,
        assertions = [contains(kw) for kw in expected_keywords],
        description = "Q&A test expecting: " + ", ".join(expected_keywords)
    )

# Function to generate math tests
def math_test_series(operations):
    """Generate a series of math tests"""
    tests = []
    for op in operations:
        a, b, symbol, result = op["a"], op["b"], op["op"], op["result"]
        tests.append(test(
            name = "Math: {} {} {}".format(a, symbol, b),
            input = "Calculate {} {} {}".format(a, symbol, b),
            assertions = [
                contains(str(result)),
            ]
        ))
    return tests

# Configuration with multiple prompts
configuration = config(
    provider = "anthropic",
    model = "claude-3-sonnet",
    prompts = [
        # Professional assistant prompt
        """You are a professional assistant with expertise in {{domain}}.
        Provide clear, accurate, and helpful responses.
        
        User: {{input}}
        Assistant:""",
        
        # Concise response prompt
        """You are a concise assistant. Answer in 1-2 sentences maximum.
        
        Question: {{input}}
        Answer:""",
    ],
    variables = {
        "domain": "general knowledge",
        "temperature": 0.5,
    }
)

# Generate math tests
math_operations = [
    {"a": 15, "b": 7, "op": "+", "result": 22},
    {"a": 100, "b": 37, "op": "-", "result": 63},
    {"a": 12, "b": 8, "op": "*", "result": 96},
    {"a": 144, "b": 12, "op": "/", "result": 12},
]

# Programming language tests
programming_langs = ["Python", "JavaScript", "Go", "Rust"]
lang_tests = []

for lang in programming_langs:
    lang_tests.append(test(
        name = "Hello World in " + lang,
        input = "Write a 'Hello, World!' program in " + lang,
        assertions = [
            icontains(lang),
            regex(r"(?i)hello,?\s*world!?"),
            # Language-specific checks
            contains("print") if lang == "Python" else 
            contains("console.log") if lang == "JavaScript" else
            contains("fmt.Println") if lang == "Go" else
            contains("println!"),
        ]
    ))

# Data structure tests with complex assertions
data_structure_tests = [
    test(
        name = "Explain linked list",
        input = "Explain what a linked list is and provide a simple example",
        assertions = [
            contains("node"),
            contains("pointer"),
            regex(r"next|link"),
            # Should mention either head or first element
            regex(r"(?i)(head|first)"),
        ],
        variables = {
            "domain": "computer science",
        }
    ),
    
    test(
        name = "Binary tree traversal",
        input = "List the three main types of binary tree traversal",
        assertions = [
            icontains("inorder"),
            icontains("preorder"),
            icontains("postorder"),
        ],
        variables = {
            "domain": "data structures and algorithms",
        }
    ),
]

# Combine all tests
tests = []
tests.extend(math_test_series(math_operations))
tests.extend(lang_tests)
tests.extend(data_structure_tests)

# Add some Q&A tests using the helper function
tests.extend([
    qa_test(
        "What is the capital of France?",
        ["Paris"],
        name="Geography: France capital"
    ),
    qa_test(
        "Who wrote '1984'?",
        ["George Orwell", "Orwell"],
        name="Literature: 1984 author"
    ),
    qa_test(
        "What is the speed of light in vacuum?",
        ["299", "792", "km/s", "kilometer"],
        name="Physics: Speed of light"
    ),
])

# Conditional test based on configuration
if configuration.provider == "anthropic":
    tests.append(test(
        name = "Claude-specific test",
        input = "What is your name and who created you?",
        assertions = [
            contains("Claude"),
            contains("Anthropic"),
        ]
    ))