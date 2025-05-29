# coding_tests.star - Programming and coding test module

def language_hello_world_test(language, expected_patterns):
    """Create a Hello World test for a programming language"""
    return test(
        name = language + " Hello World",
        input = "Write a complete 'Hello, World!' program in " + language,
        assertions = [contains(p) for p in expected_patterns] + [
            regex(r"(?i)hello,?\s*world!?"),
        ]
    )

# Hello World tests for different languages
hello_world_tests = [
    language_hello_world_test("Python", ["print", "(", ")"]),
    language_hello_world_test("JavaScript", ["console.log", "(", ")"]),
    language_hello_world_test("Java", ["public", "class", "System.out.println"]),
    language_hello_world_test("Go", ["package main", "import", "fmt.Println"]),
    language_hello_world_test("Rust", ["fn main", "println!"]),
    language_hello_world_test("C", ["#include", "printf", "return 0"]),
]

# Algorithm implementation tests
algorithm_tests = [
    test(
        name = "Bubble sort implementation",
        input = "Implement bubble sort in Python",
        assertions = [
            contains("def bubble_sort"),
            contains("for"),
            contains("if"),
            contains("swap"),
            contains("return"),
            regex(r"(?i)compare|comparison"),
        ]
    ),
    
    test(
        name = "Binary search",
        input = "Write a binary search function that returns the index of a target value in a sorted array",
        assertions = [
            regex(r"(?i)binary_search|binarySearch"),
            contains("mid"),
            contains("left"),
            contains("right"),
            contains("return"),
            regex(r"(?i)while|recursive"),
        ]
    ),
    
    test(
        name = "Factorial function",
        input = "Write both iterative and recursive factorial functions",
        assertions = [
            contains("factorial"),
            contains("return"),
            regex(r"(?i)recursive|recursion"),
            regex(r"(?i)iterative|loop|for|while"),
            contains("1"),  # Base case
        ]
    ),
]

# Data structure tests
data_structure_tests = [
    test(
        name = "Stack implementation",
        input = "Implement a basic stack data structure with push, pop, and peek methods",
        assertions = [
            regex(r"(?i)class Stack|def.*stack"),
            contains("push"),
            contains("pop"), 
            contains("peek"),
            regex(r"(?i)empty|isEmpty"),
        ]
    ),
    
    test(
        name = "Linked list node",
        input = "Define a linked list node class with data and next pointer",
        assertions = [
            regex(r"(?i)class Node|struct Node"),
            contains("data"),
            contains("next"),
            contains("None"),
        ]
    ),
    
    test(
        name = "Dictionary/HashMap usage",
        input = "Show how to create, add to, and retrieve from a dictionary/hashmap in Python",
        assertions = [
            contains("{"),
            contains("}"),
            contains("["),
            contains("]"),
            regex(r"(?i)dict|dictionary"),
            contains("="),
        ]
    ),
]

# Code review and best practices tests
best_practices_tests = [
    test(
        name = "Variable naming",
        input = "What are the naming conventions for variables in Python?",
        assertions = [
            icontains("snake_case"),
            icontains("lowercase"),
            contains("_"),
            regex(r"(?i)constant|UPPER"),
        ]
    ),
    
    test(
        name = "Code comments",
        input = "Show examples of good code comments in Python",
        assertions = [
            contains("#"),
            contains('"""'),
            regex(r"(?i)docstring"),
            regex(r"(?i)explain|describe|why"),
        ]
    ),
    
    test(
        name = "Error handling",
        input = "Demonstrate proper error handling with try-except in Python",
        assertions = [
            contains("try:"),
            contains("except"),
            regex(r"(?i)error|exception"),
            contains(":"),
            regex(r"(?i)finally|else"),
        ]
    ),
]

# SQL tests
sql_tests = [
    test(
        name = "Basic SELECT",
        input = "Write a SQL query to select all users older than 18 from a 'users' table",
        assertions = [
            icontains("SELECT"),
            icontains("FROM users"),
            icontains("WHERE"),
            contains("age"),
            contains(">"),
            contains("18"),
        ]
    ),
    
    test(
        name = "JOIN query",
        input = "Write a SQL query that joins 'orders' and 'customers' tables",
        assertions = [
            icontains("SELECT"),
            icontains("JOIN"),
            icontains("orders"),
            icontains("customers"),
            contains("ON"),
            contains("="),
        ]
    ),
]

# Export all coding tests
tests = hello_world_tests + algorithm_tests + data_structure_tests + best_practices_tests + sql_tests