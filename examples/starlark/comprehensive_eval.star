# comprehensive_eval.star - Advanced evaluation configuration using Starlark
# This example demonstrates PE's Starlark integration for complex test generation and evaluation

# ==============================================================================
# PROMPT DEFINITION
# ==============================================================================

base_prompt = """You are an expert {role} assistant with deep knowledge in {domain}.

When answering questions:
1. Provide clear, structured responses
2. Include relevant examples when helpful  
3. Consider edge cases and limitations
4. Cite sources or reasoning when applicable

Current task: {task}
"""

# ==============================================================================
# DYNAMIC TEST GENERATION
# ==============================================================================

def generate_test_matrix():
    """Generate comprehensive test cases across multiple dimensions"""
    
    # Define test dimensions
    roles = ["software engineering", "data science", "system design"]
    domains = ["Python", "distributed systems", "machine learning"]
    complexity_levels = ["beginner", "intermediate", "advanced"]
    
    # Test case templates by complexity
    test_templates = {
        "beginner": [
            "What is {concept}?",
            "Explain {concept} in simple terms",
            "Give an example of {concept}"
        ],
        "intermediate": [
            "Compare {concept} with {alternative}",
            "What are the trade-offs of using {concept}?",
            "How would you implement {concept}?"
        ],
        "advanced": [
            "Design a system that uses {concept} at scale",
            "What are the theoretical limitations of {concept}?",
            "Optimize {concept} for {constraint}"
        ]
    }
    
    # Key concepts to test
    concepts = {
        "Python": ["decorators", "generators", "metaclasses"],
        "distributed systems": ["consensus", "CAP theorem", "sharding"],
        "machine learning": ["backpropagation", "regularization", "transformers"]
    }
    
    tests = []
    
    # Generate test cases
    for role in roles:
        for domain in domains:
            for level in complexity_levels:
                # Get relevant concepts
                domain_concepts = concepts.get(domain, ["general concepts"])
                
                for concept in domain_concepts:
                    for template in test_templates[level]:
                        # Create test case
                        test = {
                            "name": f"{role}_{domain}_{level}_{concept}",
                            "vars": {
                                "role": role,
                                "domain": domain,
                                "task": template.format(
                                    concept=concept,
                                    alternative=get_alternative(concept),
                                    constraint=get_constraint(level)
                                )
                            },
                            "assertions": generate_assertions(level, domain),
                            "metadata": {
                                "complexity": level,
                                "expected_tokens": get_expected_tokens(level),
                                "timeout": get_timeout(level)
                            }
                        }
                        tests.append(test)
    
    return tests

def get_alternative(concept):
    """Get alternative concept for comparison"""
    alternatives = {
        "decorators": "higher-order functions",
        "generators": "iterators",
        "consensus": "eventual consistency",
        "backpropagation": "genetic algorithms"
    }
    return alternatives.get(concept, "alternative approach")

def get_constraint(level):
    """Get optimization constraint based on level"""
    constraints = {
        "beginner": "clarity",
        "intermediate": "performance",
        "advanced": "distributed scale"
    }
    return constraints[level]

def get_expected_tokens(level):
    """Expected response length by complexity"""
    return {
        "beginner": 150,
        "intermediate": 300,
        "advanced": 500
    }[level]

def get_timeout(level):
    """Timeout in seconds by complexity"""
    return {
        "beginner": 10,
        "intermediate": 20,
        "advanced": 30
    }[level]

# ==============================================================================
# ASSERTION GENERATION
# ==============================================================================

def generate_assertions(level, domain):
    """Generate level-appropriate assertions"""
    
    # Base assertions for all levels
    assertions = [
        min_length(50),
        max_length(1000),
        no_contradictions(),
        proper_grammar()
    ]
    
    # Level-specific assertions
    if level == "beginner":
        assertions.extend([
            contains_simple_explanation(),
            avoids_jargon(),
            includes_analogy_or_example()
        ])
    elif level == "intermediate":
        assertions.extend([
            contains_technical_details(),
            mentions_trade_offs(),
            includes_practical_example()
        ])
    else:  # advanced
        assertions.extend([
            demonstrates_deep_understanding(),
            considers_edge_cases(),
            includes_references_or_citations(),
            discusses_scalability()
        ])
    
    # Domain-specific assertions
    if domain == "Python":
        assertions.append(python_code_is_valid())
    elif domain == "distributed systems":
        assertions.append(mentions_distributed_concepts())
    elif domain == "machine learning":
        assertions.append(includes_mathematical_notation())
    
    return assertions

# ==============================================================================
# CUSTOM EVALUATION RUBRICS
# ==============================================================================

def comprehensive_quality_rubric(output, context):
    """Multi-dimensional quality assessment"""
    
    scores = {}
    feedback = []
    
    # 1. Accuracy (0-1.0)
    accuracy_score = evaluate_accuracy(output, context)
    scores["accuracy"] = accuracy_score
    if accuracy_score < 0.7:
        feedback.append("⚠️ Some technical inaccuracies detected")
    else:
        feedback.append("✓ Technically accurate")
    
    # 2. Completeness (0-1.0)
    completeness = evaluate_completeness(output, context)
    scores["completeness"] = completeness
    if completeness < 0.8:
        feedback.append("⚠️ Response could be more comprehensive")
    else:
        feedback.append("✓ Comprehensive response")
    
    # 3. Clarity (0-1.0)
    clarity = evaluate_clarity(output, context["complexity"])
    scores["clarity"] = clarity
    if clarity > 0.8:
        feedback.append("✓ Clear and well-structured")
    
    # 4. Examples (0-1.0)
    example_quality = evaluate_examples(output)
    scores["examples"] = example_quality
    if example_quality < 0.5:
        feedback.append("⚠️ Could benefit from more/better examples")
    
    # 5. Engagement (0-1.0)
    engagement = evaluate_engagement(output)
    scores["engagement"] = engagement
    
    # Calculate weighted final score
    weights = {
        "accuracy": 0.3,
        "completeness": 0.25,
        "clarity": 0.2,
        "examples": 0.15,
        "engagement": 0.1
    }
    
    final_score = sum(scores[k] * weights[k] for k in weights)
    
    return {
        "score": final_score,
        "pass": final_score >= 0.7,
        "scores": scores,
        "feedback": "\n".join(feedback),
        "recommendation": get_recommendation(final_score, scores)
    }

def evaluate_accuracy(output, context):
    """Evaluate technical accuracy based on domain"""
    # Simplified - in reality would use domain-specific checks
    indicators = {
        "correct": ["accurate", "correct", "precisely"],
        "incorrect": ["wrong", "incorrect", "mistaken"]
    }
    
    correct_count = sum(1 for word in indicators["correct"] if word in output.lower())
    incorrect_count = sum(1 for word in indicators["incorrect"] if word in output.lower())
    
    if has_factual_errors(output, context["domain"]):
        return 0.3
    elif incorrect_count > 0:
        return 0.6
    else:
        return min(1.0, 0.7 + correct_count * 0.1)

def evaluate_completeness(output, context):
    """Check if all aspects of the question are addressed"""
    task = context["task"].lower()
    output_lower = output.lower()
    
    # Extract key terms from task
    key_terms = extract_key_terms(task)
    addressed = sum(1 for term in key_terms if term in output_lower)
    
    completeness = addressed / max(1, len(key_terms))
    
    # Bonus for structured response
    if has_clear_sections(output):
        completeness = min(1.0, completeness + 0.1)
    
    return completeness

def evaluate_clarity(output, complexity_level):
    """Evaluate clarity based on target complexity"""
    # Readability metrics
    fk_score = flesch_kincaid_score(output)
    
    # Target readability by level
    targets = {
        "beginner": (60, 80),    # Very easy to easy
        "intermediate": (30, 60), # Difficult to fairly easy  
        "advanced": (0, 50)       # Very difficult to difficult
    }
    
    min_score, max_score = targets.get(complexity_level, (0, 100))
    
    if min_score <= fk_score <= max_score:
        return 1.0
    elif fk_score > max_score:
        # Too simple for level
        return 0.7
    else:
        # Too complex for level
        return 0.5

def evaluate_examples(output):
    """Evaluate quality and relevance of examples"""
    example_indicators = ["for example", "for instance", "such as", "e.g.", "```"]
    example_count = sum(1 for indicator in example_indicators if indicator in output.lower())
    
    if example_count == 0:
        return 0.0
    elif example_count == 1:
        return 0.6
    elif example_count == 2:
        return 0.9
    else:
        return 1.0

def evaluate_engagement(output):
    """Evaluate how engaging the response is"""
    # Simple heuristics
    engagement_factors = {
        "questions": output.count("?"),
        "variety": len(set(output.split())) / max(1, len(output.split())),
        "paragraphs": output.count("\n\n") + 1,
        "formatting": 1 if any(marker in output for marker in ["#", "**", "- ", "1."]) else 0
    }
    
    score = 0.0
    if engagement_factors["questions"] > 0:
        score += 0.2
    if engagement_factors["variety"] > 0.3:
        score += 0.3
    if engagement_factors["paragraphs"] > 2:
        score += 0.3
    if engagement_factors["formatting"]:
        score += 0.2
        
    return min(1.0, score)

def get_recommendation(score, subscores):
    """Generate improvement recommendation"""
    if score >= 0.9:
        return "Excellent response! Minor refinements could focus on " + \
               min(subscores.items(), key=lambda x: x[1])[0]
    elif score >= 0.7:
        weak_areas = [k for k, v in subscores.items() if v < 0.7]
        return f"Good response. Consider improving: {', '.join(weak_areas)}"
    else:
        return "Response needs significant improvement in multiple areas"

# ==============================================================================
# COMPARATIVE EVALUATION
# ==============================================================================

def comparative_evaluator(outputs, context):
    """Compare multiple outputs for the same prompt"""
    
    evaluations = []
    
    for i, output in enumerate(outputs):
        eval_result = comprehensive_quality_rubric(output, context)
        eval_result["output_id"] = i
        evaluations.append(eval_result)
    
    # Rank outputs
    ranked = sorted(evaluations, key=lambda x: x["score"], reverse=True)
    
    # Statistical analysis
    scores = [e["score"] for e in evaluations]
    mean_score = sum(scores) / len(scores)
    variance = sum((s - mean_score) ** 2 for s in scores) / len(scores)
    
    return {
        "rankings": ranked,
        "best": ranked[0],
        "worst": ranked[-1],
        "statistics": {
            "mean": mean_score,
            "variance": variance,
            "std_dev": variance ** 0.5,
            "range": max(scores) - min(scores)
        },
        "consensus": calculate_consensus(evaluations)
    }

def calculate_consensus(evaluations):
    """Calculate consensus among evaluations"""
    # Check if all pass/fail agree
    pass_fail_consensus = len(set(e["pass"] for e in evaluations)) == 1
    
    # Check score variance
    scores = [e["score"] for e in evaluations]
    score_variance = sum((s - sum(scores)/len(scores)) ** 2 for s in scores) / len(scores)
    
    if pass_fail_consensus and score_variance < 0.01:
        return "strong"
    elif pass_fail_consensus or score_variance < 0.05:
        return "moderate"
    else:
        return "weak"

# ==============================================================================
# EVALUATION PIPELINE
# ==============================================================================

def create_evaluation_pipeline():
    """Create multi-stage evaluation pipeline"""
    
    return {
        "stages": [
            # Stage 1: Basic validation
            {
                "name": "validation",
                "evaluators": [
                    basic_format_check,
                    language_detection,
                    safety_check
                ],
                "stop_on_fail": True,
                "weight": 0.1
            },
            
            # Stage 2: Quality assessment
            {
                "name": "quality",
                "evaluators": [
                    comprehensive_quality_rubric,
                    domain_specific_check
                ],
                "weight": 0.6
            },
            
            # Stage 3: Comparative analysis
            {
                "name": "comparison",
                "evaluators": [
                    comparative_evaluator
                ],
                "requires_multiple": True,
                "weight": 0.3
            }
        ],
        
        "aggregation": "weighted_average",
        "pass_threshold": 0.7,
        "report_format": "detailed"
    }

# ==============================================================================
# CONFIGURATION EXPORT
# ==============================================================================

# Generate all test cases
all_tests = generate_test_matrix()

# Create the final configuration
config = {
    "name": "Comprehensive Starlark Evaluation",
    "description": "Advanced evaluation using dynamic test generation and custom rubrics",
    
    "prompt": base_prompt,
    
    # Select subset for demonstration (full matrix would be large)
    "tests": all_tests[:20],  # First 20 tests
    
    "evaluation_pipeline": create_evaluation_pipeline(),
    
    "providers": ["gpt-4", "claude-3", "gemini-pro"],
    
    "parameters": {
        "temperature": 0.7,
        "max_tokens": 1000,
        "top_p": 0.9
    },
    
    "output": {
        "format": "json",
        "include_explanations": True,
        "save_outputs": True
    },
    
    "metadata": {
        "version": "1.0",
        "author": "PE Starlark Example",
        "total_tests": len(all_tests),
        "estimated_cost": len(all_tests) * 0.02  # Rough estimate
    }
}