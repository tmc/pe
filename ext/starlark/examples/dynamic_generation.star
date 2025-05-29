# Dynamic test generation example for PE Starlark extension
# This shows how to programmatically create test suites

def generate_length_tests():
    """Generate a suite of length-based tests"""
    
    test_configs = [
        {"min": 25, "name": "very_short"},
        {"min": 50, "name": "short"}, 
        {"min": 100, "name": "medium"},
        {"min": 200, "name": "long"},
        {"min": 500, "name": "very_long"}
    ]
    
    tests = {}
    
    for config in test_configs:
        min_len = config["min"]
        test_name = f"test_min_length_{config['name']}"
        
        # Create test function dynamically
        def length_test(response, minimum=min_len):
            passed = min_length(response, minimum)
            word_count_val = word_count(response)
            
            return {
                "pass": passed,
                "reason": f"Length check (min {minimum} chars): {'PASS' if passed else 'FAIL'}",
                "details": {
                    "minimum_required": minimum,
                    "actual_length": len(response),
                    "word_count": word_count_val
                }
            }
        
        tests[test_name] = length_test
    
    return tests

def generate_content_tests():
    """Generate tests for required content elements"""
    
    required_terms = [
        {"term": "analysis", "weight": 30, "description": "analytical thinking"},
        {"term": "example", "weight": 25, "description": "concrete examples"},
        {"term": "conclusion", "weight": 20, "description": "clear conclusion"},
        {"term": "evidence", "weight": 15, "description": "supporting evidence"},
        {"term": "reasoning", "weight": 10, "description": "logical reasoning"}
    ]
    
    tests = {}
    
    for term_config in required_terms:
        term = term_config["term"]
        weight = term_config["weight"]
        desc = term_config["description"]
        test_name = f"test_contains_{term}"
        
        def content_test(response, required_term=term, points=weight, description=desc):
            passed = contains(response, required_term)
            
            return {
                "pass": passed,
                "score": points / 100.0 if passed else 0.0,
                "reason": f"Content check for {description}: {'FOUND' if passed else 'MISSING'}",
                "details": {
                    "required_term": required_term,
                    "points_available": points,
                    "description": description
                }
            }
        
        tests[test_name] = content_test
    
    return tests

def generate_complexity_tests():
    """Generate tests for response complexity and sophistication"""
    
    complexity_indicators = [
        {"patterns": ["however", "furthermore", "nevertheless"], "name": "transitions", "points": 20},
        {"patterns": ["analyze", "synthesis", "evaluate"], "name": "analytical_terms", "points": 25},
        {"patterns": ["first", "second", "finally"], "name": "structure", "points": 15},
        {"patterns": ["research", "study", "evidence"], "name": "academic_terms", "points": 20},
        {"patterns": ["therefore", "thus", "consequently"], "name": "logical_connectors", "points": 20}
    ]
    
    tests = {}
    
    for indicator in complexity_indicators:
        patterns = indicator["patterns"]
        name = indicator["name"]
        points = indicator["points"]
        test_name = f"test_complexity_{name}"
        
        def complexity_test(response, pattern_list=patterns, category=name, score_points=points):
            found_patterns = []
            for pattern in pattern_list:
                if contains(response, pattern):
                    found_patterns.append(pattern)
            
            pattern_count = len(found_patterns)
            max_patterns = len(pattern_list)
            score = (pattern_count / max_patterns) * (score_points / 100.0)
            
            return {
                "pass": pattern_count > 0,
                "score": score,
                "reason": f"Complexity check ({category}): {pattern_count}/{max_patterns} patterns found",
                "details": {
                    "category": category,
                    "found_patterns": found_patterns,
                    "available_patterns": pattern_list,
                    "pattern_count": pattern_count,
                    "max_patterns": max_patterns,
                    "points": score_points
                }
            }
        
        tests[test_name] = complexity_test
    
    return tests

def create_test_matrix():
    """Create a comprehensive test matrix combining all generated tests"""
    
    # Collect all generated tests
    all_tests = {}
    
    # Add length tests
    length_tests = generate_length_tests()
    for name, test_func in length_tests.items():
        all_tests[name] = test_func
    
    # Add content tests  
    content_tests = generate_content_tests()
    for name, test_func in content_tests.items():
        all_tests[name] = test_func
    
    # Add complexity tests
    complexity_tests = generate_complexity_tests()
    for name, test_func in complexity_tests.items():
        all_tests[name] = test_func
    
    return all_tests

def run_test_suite(response):
    """Run the complete generated test suite on a response"""
    
    test_matrix = create_test_matrix()
    results = {}
    
    total_score = 0.0
    passed_tests = 0
    total_tests = len(test_matrix)
    
    # Run each test
    for test_name, test_func in test_matrix.items():
        try:
            result = test_func(response)
            results[test_name] = result
            
            # Accumulate scores
            if result.get("pass", False):
                passed_tests += 1
            
            if "score" in result:
                total_score += result["score"]
                
        except Exception as e:
            results[test_name] = {
                "pass": False,
                "reason": f"Test execution error: {str(e)}",
                "error": True
            }
    
    # Calculate overall metrics
    pass_rate = passed_tests / total_tests
    avg_score = total_score / total_tests
    
    return {
        "pass": pass_rate >= 0.7,
        "score": avg_score,
        "reason": f"Test suite: {passed_tests}/{total_tests} passed ({pass_rate:.1%}), avg score: {avg_score:.2f}",
        "details": {
            "individual_results": results,
            "summary": {
                "total_tests": total_tests,
                "passed_tests": passed_tests,
                "pass_rate": pass_rate,
                "average_score": avg_score,
                "total_score": total_score
            }
        }
    }

# Example usage functions that can be called from PE
def test_comprehensive_generated(response):
    """Main entry point for running the complete generated test suite"""
    return run_test_suite(response)

def test_length_suite_only(response):
    """Run only the length-based tests"""
    length_tests = generate_length_tests()
    results = {}
    
    for test_name, test_func in length_tests.items():
        results[test_name] = test_func(response)
    
    return {
        "pass": all(r.get("pass", False) for r in results.values()),
        "reason": "Length test suite completed",
        "details": results
    }

def test_content_suite_only(response):
    """Run only the content-based tests"""
    content_tests = generate_content_tests()
    results = {}
    total_score = 0.0
    
    for test_name, test_func in content_tests.items():
        result = test_func(response)
        results[test_name] = result
        total_score += result.get("score", 0.0)
    
    return {
        "pass": total_score >= 0.5,
        "score": total_score,
        "reason": f"Content test suite: {total_score:.2f} total score",
        "details": results
    }