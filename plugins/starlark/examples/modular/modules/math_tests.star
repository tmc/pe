# math_tests.star - Mathematics test module

def create_arithmetic_test(expression, result):
    """Helper to create arithmetic tests"""
    return test(
        name = "Arithmetic: " + expression,
        input = "Calculate: " + expression,
        assertions = [
            contains(str(result)),
        ]
    )

def create_word_problem(problem, answer_keywords):
    """Helper to create word problem tests"""
    return test(
        name = "Word problem: " + problem[:30] + "...",
        input = problem,
        assertions = [contains(kw) for kw in answer_keywords]
    )

# Basic arithmetic tests
arithmetic_tests = [
    create_arithmetic_test("25 * 4", 100),
    create_arithmetic_test("1000 / 25", 40),
    create_arithmetic_test("(10 + 5) * 3", 45),
    create_arithmetic_test("2^10", 1024),
]

# Word problems
word_problems = [
    create_word_problem(
        "If a rectangle has length 12 and width 8, what is its area?",
        ["96", "square"]
    ),
    create_word_problem(
        "A train travels at 60 mph for 3.5 hours. How far does it go?",
        ["210", "miles"]
    ),
    create_word_problem(
        "If 30% of 150 students passed the exam, how many students passed?",
        ["45", "students"]
    ),
]

# Advanced math tests
advanced_tests = [
    test(
        name = "Quadratic formula",
        input = "What is the quadratic formula?",
        assertions = [
            contains("x ="),
            contains("-b"),
            contains("±"),
            contains("√"),
            contains("2a"),
        ]
    ),
    
    test(
        name = "Prime factorization",
        input = "What is the prime factorization of 60?",
        assertions = [
            contains("2"),
            contains("3"),
            contains("5"),
            regex(r"2\^2|2²"),  # Should show 2 squared
        ]
    ),
    
    test(
        name = "Fibonacci sequence",
        input = "List the first 10 numbers in the Fibonacci sequence",
        assertions = [
            contains("0"),
            contains("1"),
            contains("1"),
            contains("2"),
            contains("3"),
            contains("5"),
            contains("8"),
            contains("13"),
            contains("21"),
            contains("34"),
        ]
    ),
]

# Export all tests
tests = arithmetic_tests + word_problems + advanced_tests